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

package operator

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"k8s.io/component-base/cli/flag"
)

// TLSServerConfig contains the necessary fields to configure
// web server TLS
type TLSServerConfig struct {
	CertFile       string
	KeyFile        string
	ClientCAFile   string
	MinVersion     string
	CipherSuites   []string
	ReloadInterval time.Duration
}

// NewTLSConfig provides new server TLS configuration.
func NewTLSConfig(logger log.Logger, certFile, keyFile, clientCAFile, minVersion string, cipherSuites []string) (*tls.Config, error) {
	if certFile == "" && keyFile == "" {
		if clientCAFile != "" {
			return nil, errors.New("when a client CA is used a server key and certificate must also be provided")
		}

		level.Info(logger).Log("msg", "TLS disabled key and cert must be set to enable")

		return nil, nil
	}

	level.Info(logger).Log("msg", "enabling server side TLS")

	tlsCfg := &tls.Config{}

	version, err := flag.TLSVersion(minVersion)
	if err != nil {
		return nil, fmt.Errorf("TLS version invalid: %w", err)
	}

	tlsCfg.MinVersion = version

	cipherSuiteIDs, err := flag.TLSCipherSuites(cipherSuites)
	if err != nil {
		return nil, fmt.Errorf("TLS cipher suite name to ID conversion: %v", err)
	}

	// A list of supported cipher suites for TLS versions up to TLS 1.2.
	// If CipherSuites is nil, a default list of secure cipher suites is used.
	// Note that TLS 1.3 ciphersuites are not configurable.
	tlsCfg.CipherSuites = cipherSuiteIDs

	if clientCAFile != "" {
		if info, err := os.Stat(clientCAFile); err == nil && info.Mode().IsRegular() {
			caPEM, err := ioutil.ReadFile(clientCAFile)
			if err != nil {
				return nil, fmt.Errorf("reading client CA: %w", err)
			}

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caPEM) {
				return nil, fmt.Errorf("building client CA: %w", err)
			}

			tlsCfg.ClientCAs = certPool
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert

			level.Info(logger).Log("msg", "server TLS client verification enabled")
		}
	}

	return tlsCfg, nil
}
