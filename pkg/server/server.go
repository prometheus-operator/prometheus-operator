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
	"crypto/x509"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	rbacproxytls "github.com/brancz/kube-rbac-proxy/pkg/tls"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
			ReloadInterval: time.Minute,
		},
	}
}

func RegisterFlags(fs *flag.FlagSet, c *Config) {
	fs.StringVar(&c.ListenAddress, "web.listen-address", c.ListenAddress, "Address on which to expose metrics and web interface.")

	fs.BoolVar(&c.EnableHTTP2, "web.enable-http2", c.EnableHTTP2, "Enable HTTP2 connections.")

	fs.BoolVar(&c.TLSConfig.Enabled, "web.enable-tls", c.TLSConfig.Enabled, "Enable TLS for the web server.")
	fs.StringVar(&c.TLSConfig.CertFile, "web.cert-file", c.TLSConfig.CertFile, "Certficate file to be used for the web server.")
	fs.StringVar(&c.TLSConfig.KeyFile, "web.key-file", c.TLSConfig.KeyFile, "Private key matching the cert file to be used for the web server.")
	fs.StringVar(&c.TLSConfig.ClientCAFile, "web.client-ca-file", c.TLSConfig.ClientCAFile, "Client CA certificate file to be used for the web server.")

	fs.DurationVar(&c.TLSConfig.ReloadInterval, "web.tls-reload-interval", c.TLSConfig.ReloadInterval, "The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s).")

	fs.StringVar(&c.TLSConfig.MinVersion, "web.tls-min-version", c.TLSConfig.MinVersion,
		"Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.")
	fs.Var(&c.TLSConfig.CipherSuites, "web.tls-cipher-suites", "Comma-separated list of cipher suites for the server."+
		" Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants)."+
		"If omitted, the default Go cipher suites will be used. "+
		"Note that TLS 1.3 ciphersuites are not configurable.")
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
	ReloadInterval time.Duration
}

// Convert returns a *tls.Config from the given TLSConfig.
func (tc *TLSConfig) Convert(logger log.Logger) (*tls.Config, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	if !tc.Enabled {
		return nil, nil
	}

	if tc.CertFile == "" && tc.KeyFile == "" {
		if tc.ClientCAFile != "" {
			return nil, fmt.Errorf("server key and certificate must be provided when a client CA is configured")
		}

		// Disable TLS.
		level.Warn(logger).Log("msg", "server key and certificate not provided, TLS disabled")
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

	if tc.ClientCAFile == "" {
		return tlsCfg, nil
	}

	info, err := os.Stat(tc.ClientCAFile)
	switch {
	case err != nil:
		level.Warn(logger).Log("msg", "server TLS client verification disabled", "err", err, "client_ca_file", tc.ClientCAFile)

	case !info.Mode().IsRegular():
		level.Warn(logger).Log("msg", "server TLS client verification disabled", "client_ca_file", tc.ClientCAFile, "file_mode", info.Mode().String())

	default:
		caPEM, err := os.ReadFile(tc.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("reading client CA %q: %w", tc.ClientCAFile, err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("client CA %q: failed to parse certificate", tc.ClientCAFile)
		}

		tlsCfg.ClientCAs = certPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		level.Info(logger).Log("msg", "server TLS client verification enabled", "client_ca_file", tc.ClientCAFile)
	}

	return tlsCfg, nil
}

// Server is a web server.
type Server struct {
	logger   log.Logger
	srv      *http.Server
	listener net.Listener
	cfg      *Config
}

// NewServer initializes a web server with the given handler (typically an http.MuxServe).
func NewServer(logger log.Logger, c *Config, handler http.Handler) (*Server, error) {
	tlsConfig, err := c.TLSConfig.Convert(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
	}

	listener, err := net.Listen("tcp", c.ListenAddress)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		// use flags on standard logger to align with base logger and get consistent parsed fields form adapter:
		// use shortfile flag to get proper 'caller' field (avoid being wrongly parsed/extracted from message)
		// and no datetime related flag to keep 'ts' field from base logger (with controlled format)
		ErrorLog: stdlog.New(log.NewStdlibAdapter(logger), "", stdlog.Lshortfile),
	}

	if !c.EnableHTTP2 {
		srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return &Server{
		logger:   logger,
		srv:      srv,
		listener: listener,
		cfg:      c,
	}, nil
}

// Serve starts the web server. It will block until the server is shutted down
// or an error occurs.
func (s *Server) Serve(ctx context.Context) error {
	if s.srv.TLSConfig == nil {
		level.Info(s.logger).Log("msg", "starting insecure server", "address", s.listener.Addr().String())
		if err := s.srv.Serve(s.listener); err != http.ErrServerClosed {
			return err
		}

		return nil
	}

	r, err := rbacproxytls.NewCertReloader(
		s.cfg.TLSConfig.CertFile,
		s.cfg.TLSConfig.KeyFile,
		s.cfg.TLSConfig.ReloadInterval,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize certificate reloader: %w", err)
	}

	s.srv.TLSConfig.GetCertificate = r.GetCertificate
	go func() {
		for {
			// r.Watch will wait ReloadInterval, so this is not
			// a hot loop
			if err := r.Watch(ctx); err != nil {
				level.Warn(s.logger).Log(
					"msg", "error watching certificate reloader",
					"err", err)
			}
		}
	}()

	level.Info(s.logger).Log("msg", "starting secure server", "address", s.listener.Addr().String(), "http2", s.cfg.EnableHTTP2)
	if err := s.srv.ServeTLS(s.listener, "", ""); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown closes gracefully all active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	level.Info(s.logger).Log("msg", "shutting down web server")
	return s.srv.Shutdown(ctx)
}
