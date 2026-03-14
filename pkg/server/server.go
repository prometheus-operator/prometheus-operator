// Copyright 2020 The prometheus-operator Authors
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

package server

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	kflag "k8s.io/component-base/cli/flag"

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	defaultTLSDir     = "/etc/tls/private"
	defaultTLSVersion = "VersionTLS13"
)

func DefaultConfig(listenAddress string, enableTLS bool) Config {
	return Config{
		ListenAddress: listenAddress,

		// Mitigate CVE-2023-44487 by disabling HTTP2 by default until the Go
		// standard library and golang.org/x/net are fully fixed.
		// Right now, it is possible for authenticated and unauthenticated users to
		// hold open HTTP2 connections and consume huge amounts of memory.
		// See:
		// * https://github.com/kubernetes/kubernetes/pull/121120
		// * https://github.com/kubernetes/kubernetes/issues/121197
		// * https://github.com/golang/go/issues/63417#issuecomment-1758858612
		EnableHTTP2: false,

		TLSConfig: TLSConfig{
			Enabled:        enableTLS,
			CertFile:       filepath.Join(defaultTLSDir, "tls.crt"),
			KeyFile:        filepath.Join(defaultTLSDir, "tls.key"),
			ClientCAFile:   filepath.Join(defaultTLSDir, "tls-ca.crt"),
			MinVersion:     defaultTLSVersion,
			CipherSuites:   operator.StringSet{},
			Curves:         operator.StringSet{},
			ReloadInterval: time.Minute,
		},
	}
}

func RegisterFlags(fs *flag.FlagSet, c *Config) {
	fs.StringVar(&c.ListenAddress, "web.listen-address", c.ListenAddress, "Address on which to expose metrics and web interface.")

	fs.BoolVar(&c.EnableHTTP2, "web.enable-http2", c.EnableHTTP2, "Enable HTTP2 connections.")

	fs.BoolVar(&c.TLSConfig.Enabled, "web.enable-tls", c.TLSConfig.Enabled, "Enable TLS for the web server.")
	fs.StringVar(&c.TLSConfig.CertFile, "web.cert-file", c.TLSConfig.CertFile, "Certificate file to be used for the web server.")
	fs.StringVar(&c.TLSConfig.KeyFile, "web.key-file", c.TLSConfig.KeyFile, "Private key matching the cert file to be used for the web server.")
	fs.StringVar(&c.TLSConfig.ClientCAFile, "web.client-ca-file", c.TLSConfig.ClientCAFile, "Client CA certificate file to be used for the web server.")

	fs.DurationVar(&c.TLSConfig.ReloadInterval, "web.tls-reload-interval", c.TLSConfig.ReloadInterval, "The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s).")

	fs.StringVar(&c.TLSConfig.MinVersion, "web.tls-min-version", c.TLSConfig.MinVersion,
		"Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.")
	fs.Var(&c.TLSConfig.CipherSuites, "web.tls-cipher-suites", "Comma-separated list of cipher suites for the server."+
		" Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants)."+
		"If omitted, the default Go cipher suites will be used. "+
		"Note that TLS 1.3 ciphersuites are not configurable.")
	fs.Var(&c.TLSConfig.Curves, "web.tls-curves", "Comma-separated list of TLS curves for the server. Supported values: "+strings.Join(slices.Sorted(maps.Keys(supportedCurves)), ", ")+".")
}

var supportedCurves = map[string]tls.CurveID{}

func init() {
	// NOTE: the list can be expanded as new curves get added to the Go standard
	// library. The current list corresponds to what is supported in Go 1.25.
	for _, c := range []tls.CurveID{
		tls.CurveP256,
		tls.CurveP384,
		tls.CurveP521,
		tls.X25519,
		tls.X25519MLKEM768,
	} {
		supportedCurves[c.String()] = c
	}
}

// Config defines the web server configuration.
type Config struct {
	ListenAddress string
	EnableHTTP2   bool
	TLSConfig     TLSConfig
}

// TLSConfig defines the TLS settings of the web server.
type TLSConfig struct {
	Enabled        bool
	CertFile       string
	KeyFile        string
	ClientCAFile   string
	MinVersion     string
	CipherSuites   operator.StringSet
	Curves         operator.StringSet
	ReloadInterval time.Duration
}

// Convert returns a *tls.Config from the given TLSConfig.
// It returns nil when TLS isn't enabled/configured.
func (tc *TLSConfig) Convert(logger *slog.Logger) (*tls.Config, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	if !tc.Enabled {
		return nil, nil
	}

	if tc.CertFile == "" && tc.KeyFile == "" {
		if tc.ClientCAFile != "" {
			return nil, fmt.Errorf("server key and certificate must be provided when a client CA is configured")
		}

		// Disable TLS.
		logger.Warn("server key and certificate not provided, TLS disabled")
		return nil, nil
	}

	tlsCfg := &tls.Config{}
	version, err := kflag.TLSVersion(tc.MinVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS version: %w", err)
	}

	// Any older versions won't allow a secure connection.
	switch version {
	case tls.VersionTLS12:
	case tls.VersionTLS13:
	default:
		return nil, fmt.Errorf("TLS version %q isn't supported", tls.VersionName(version))
	}

	tlsCfg.MinVersion = version

	cipherSuiteIDs, err := kflag.TLSCipherSuites(tc.CipherSuites.Slice())
	if err != nil {
		return nil, fmt.Errorf("failed to convert TLS cipher suite name to ID: %w", err)
	}

	// A list of supported cipher suites for TLS versions up to TLS 1.2.
	// If CipherSuites is nil, a default list of secure cipher suites is used.
	// Note that TLS 1.3 ciphersuites are not configurable.
	tlsCfg.CipherSuites = cipherSuiteIDs

	// While the CurvePreferences name seems to imply that the order is
	// important, it isn't taken into account when TLS negotiation happens
	// hence there's no need to preserve the original order.
	var curves []tls.CurveID
	for _, curve := range tc.Curves.Slice() {
		c, found := supportedCurves[curve]
		if !found {
			return nil, fmt.Errorf("%q is not a supported curve value", curve)
		}
		curves = append(curves, c)
	}
	tlsCfg.CurvePreferences = curves

	if tc.ClientCAFile == "" {
		return tlsCfg, nil
	}

	info, err := os.Stat(tc.ClientCAFile)
	switch {
	case err != nil:
		logger.Warn("server TLS client verification disabled", "client_ca_file", tc.ClientCAFile, "err", err)

	case !info.Mode().IsRegular():
		logger.Warn("server TLS client verification disabled", "client_ca_file", tc.ClientCAFile, "file_mode", info.Mode().String())

	default:
		// The client CA content will be checked by the cert controller.
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		logger.Info("server TLS client verification enabled", "client_ca_file", tc.ClientCAFile)
	}

	return tlsCfg, nil
}

// Server is a web server.
type Server struct {
	logger *slog.Logger

	listener net.Listener
	srv      *http.Server
	runners  []func(context.Context)

	cfg *Config
}

// NewServer initializes a web server with the given handler (typically an http.MuxServe).
func NewServer(logger *slog.Logger, c *Config, handler http.Handler) (*Server, error) {
	listener, err := net.Listen("tcp", c.ListenAddress)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := c.TLSConfig.Convert(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
	}

	var runners []func(context.Context)
	if tlsConfig != nil {
		var (
			certController *dynamiccertificates.DynamicServingCertificateController
			clientCA       dynamiccertificates.CAContentProvider
			servingCert    dynamiccertificates.CertKeyContentProvider
		)

		if tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert {
			clientCA, err = dynamiccertificates.NewDynamicCAContentFromFile("clientCA", c.TLSConfig.ClientCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client CA certificate: %w", err)
			}

			if err = clientCA.(*dynamiccertificates.DynamicFileCAContent).RunOnce(context.Background()); err != nil {
				return nil, fmt.Errorf("failed to sync client CA certificate: %w", err)
			}

			runners = append(runners, func(ctx context.Context) {
				clientCA.(*dynamiccertificates.DynamicFileCAContent).Run(ctx, 1)
			})
		}

		if c.TLSConfig.CertFile != "" && c.TLSConfig.KeyFile != "" {
			servingCert, err = dynamiccertificates.NewDynamicServingContentFromFiles("servingCert", c.TLSConfig.CertFile, c.TLSConfig.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load serving certificate and key: %w", err)
			}

			if err = servingCert.(*dynamiccertificates.DynamicCertKeyPairContent).RunOnce(context.Background()); err != nil {
				return nil, fmt.Errorf("failed to sync serving certificate: %w", err)
			}

			runners = append(runners, func(ctx context.Context) {
				servingCert.(*dynamiccertificates.DynamicCertKeyPairContent).Run(ctx, 1)
			})
		}

		certController = dynamiccertificates.NewDynamicServingCertificateController(
			tlsConfig,
			clientCA,
			servingCert,
			nil,
			nil,
		)

		if clientCA != nil {
			clientCA.AddListener(certController)
		}

		if servingCert != nil {
			servingCert.AddListener(certController)
		}

		// Ensure that the configuration is valid.
		err = certController.RunOnce()
		if err != nil {
			return nil, fmt.Errorf("failed to sync certificates: %w", err)
		}

		runners = append(runners, func(ctx context.Context) {
			certController.Run(1, ctx.Done())
		})

		tlsConfig.GetConfigForClient = certController.GetConfigForClient

		listener = tls.NewListener(listener, tlsConfig)
	}

	srv := &http.Server{
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	if !c.EnableHTTP2 {
		srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return &Server{
		logger:   logger,
		srv:      srv,
		listener: listener,
		runners:  runners,
		cfg:      c,
	}, nil
}

// Serve starts the web server. It will block until the server is shutted down
// or an error occurs.
func (s *Server) Serve(ctx context.Context) error {
	for _, r := range s.runners {
		go r(ctx)
	}

	if s.srv.TLSConfig == nil {
		s.logger.Info("starting insecure server", "address", s.listener.Addr().String())
	} else {
		s.logger.Info("starting secure server", "address", s.listener.Addr().String(), "http2", s.cfg.EnableHTTP2)
	}

	if err := s.srv.Serve(s.listener); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown closes gracefully all active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down web server")
	return s.srv.Shutdown(ctx)
}
