// Copyright 2022 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	rbacproxytls "github.com/brancz/kube-rbac-proxy/pkg/tls"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/sync/errgroup"

	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	"github.com/prometheus-operator/prometheus-operator/pkg/server"
)

const (
	defaultTLSDir  = "/etc/tls/private"
	defaultCrtFile = defaultTLSDir + "/tls.crt"
	defaultKeyFile = defaultTLSDir + "/tls.key"
	defaultCaCrt   = defaultTLSDir + "/tls-ca.crt"
)

var (
	cfg     = config{}
	flagset = flag.CommandLine

	serverTLS          bool
	rawTLSCipherSuites string
)

func main() {
	flagset.StringVar(&cfg.ListenAddress, "web.listen-address", ":8443", "Address on which the admission webhook service listens")
	flagset.BoolVar(&serverTLS, "web.enable-tls", true, "Enable TLS web server")

	flagset.StringVar(&cfg.ServerTLSConfig.CertFile, "web.cert-file", defaultCrtFile, "Certificate file to be used for the web server.")
	flagset.StringVar(&cfg.ServerTLSConfig.KeyFile, "web.key-file", defaultKeyFile, "Private key matching the certificate file to be used for the web server")
	flagset.StringVar(&cfg.ServerTLSConfig.ClientCAFile, "web.client-ca-file", defaultCaCrt, "Client CA certificate file to be used for web server.")

	flagset.DurationVar(&cfg.ServerTLSConfig.ReloadInterval, "web.tls-reload-interval", time.Minute, "The interval at which to watch for TLS certificate changes (default 1m).")
	flagset.StringVar(&cfg.ServerTLSConfig.MinVersion, "web.tls-min-version", "VersionTLS13", fmt.Sprintf("Minimum TLS version supported. One of %s", validTLSVersions()))
	flagset.StringVar(&rawTLSCipherSuites, "web.tls-cipher-suites", "", "Comma-separated list of cipher suites for the server."+
		" Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants)."+
		"If omitted, the default Go cipher suites will be used."+
		"Note that TLS 1.3 ciphersuites are not configurable.")

	flagset.StringVar(&cfg.LogLevel, "log-level", logging.LevelInfo, fmt.Sprintf("Log level to use. Possible values: %s", strings.Join(logging.AvailableLogLevels, ", ")))
	flagset.StringVar(&cfg.LogFormat, "log-format", logging.FormatLogFmt, fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(logging.AvailableLogFormats, ", ")))

	_ = flagset.Parse(os.Args[1:])

	logger, err := logging.NewLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		stdlog.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg, ctx := errgroup.WithContext(ctx)

	listener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		level.Error(logger).Log("msg", "failed to start required HTTP listener", "err", err)
		os.Exit(1)
	}

	tlsConf, err := loadTLSConfigFromFlags(ctx, logger, wg)
	if err != nil {
		level.Error(logger).Log("msg", "failed to build TLS config", "err", err)
		os.Exit(1)
	}

	server := newSrv(logger, tlsConf)
	wg.Go(func() error {
		return server.run(listener)
	})

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-term:
		level.Info(logger).Log("msg", "Received signal, exiting gracefully...", "signal", sig.String())
	case <-ctx.Done():
	}

	if err := server.shutdown(ctx); err != nil {
		level.Warn(logger).Log("msg", "Server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		level.Warn(logger).Log("msg", "Unhandled error received. Exiting...", "err", err)
		os.Exit(1)
	}
}

func (s *srv) run(listener net.Listener) error {
	log := log.With(s.logger, "address", listener.Addr().String())

	if s.s.TLSConfig != nil {
		level.Info(log).Log("msg", "Starting TLS enabled server")
		if err := s.s.ServeTLS(listener, "", ""); err != http.ErrServerClosed {
			return err
		}
		return nil
	}

	level.Info(log).Log("msg", "Starting insecure server")
	if err := s.s.Serve(listener); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *srv) shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

func newSrv(logger log.Logger, tlsConf *tls.Config) *srv {
	mux := http.NewServeMux()
	admit := admission.New(log.With(logger, "component", "admissionwebhook"))
	admit.Register(mux)

	r := prometheus.NewRegistry()
	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		version.NewCollector("prometheus_operator_admission_webhook"),
	)
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"up"}`))
		w.WriteHeader(http.StatusOK)
	})

	return &srv{
		logger: logger,
		s: &http.Server{
			Handler:           mux,
			TLSConfig:         tlsConf,
			ReadHeaderTimeout: 30 * time.Second,
			ReadTimeout:       30 * time.Second,
			// use flags on standard logger to align with base logger and get consistent parsed fields form adapter:
			// use shortfile flag to get proper 'caller' field (avoid being wrongly parsed/extracted from message)
			// and no datetime related flag to keep 'ts' field from base logger (with controlled format)
			ErrorLog: stdlog.New(log.NewStdlibAdapter(logger), "", stdlog.Lshortfile),
		},
	}
}

// loadTLSConfigFromFlags creates a tls.Config if configured and starts a watch on the dir to reload certs
func loadTLSConfigFromFlags(ctx context.Context, logger log.Logger, wg *errgroup.Group) (*tls.Config, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)
	if serverTLS {
		if _, ok := allowedTLSVersions[cfg.ServerTLSConfig.MinVersion]; !ok {
			return nil, fmt.Errorf("unsupported TLS version %s provided", cfg.ServerTLSConfig.MinVersion)
		}

		if rawTLSCipherSuites != "" {
			cfg.ServerTLSConfig.CipherSuites = strings.Split(rawTLSCipherSuites, ",")
		}
		tlsConfig, err = server.NewTLSConfig(
			logger,
			cfg.ServerTLSConfig.CertFile,
			cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ClientCAFile,
			cfg.ServerTLSConfig.MinVersion,
			cfg.ServerTLSConfig.CipherSuites,
		)
		if tlsConfig == nil || err != nil {
			return nil, fmt.Errorf("invalid TLS config: %w", err)
		}
	}

	if tlsConfig != nil {
		r, err := rbacproxytls.NewCertReloader(
			cfg.ServerTLSConfig.CertFile,
			cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ReloadInterval,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize certificate reloader: %w", err)
		}

		tlsConfig.GetCertificate = r.GetCertificate
		wg.Go(func() error {
			for {
				// r.Watch will wait ReloadInterval, so this is not
				// a hot loop
				if err := r.Watch(ctx); err != nil {
					level.Warn(logger).Log("msg", "error watching certificate reloader", "err", err)
				} else {
					return nil
				}
			}
		})
	}
	return tlsConfig, nil
}

type config struct {
	ListenAddress   string
	TLSInsecure     bool
	ServerTLSConfig server.TLSServerConfig
	LocalHost       string
	LogLevel        string
	LogFormat       string
}

type srv struct {
	logger log.Logger
	s      *http.Server
}

// any older versions won't allow a secure conn
var allowedTLSVersions = map[string]bool{"VersionTLS13": true, "VersionTLS12": true}

func validTLSVersions() string {
	var out string
	for validVersion := range allowedTLSVersions {
		out += validVersion + ","
	}
	return strings.TrimRight(out, ",")
}
