// Copyright 2025 The prometheus-operator Authors
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

package v1

import (
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// +kubebuilder:validation:Enum=TLS10;TLS11;TLS12;TLS13
type TLSVersion string

const (
	TLSVersion10 TLSVersion = "TLS10"
	TLSVersion11 TLSVersion = "TLS11"
	TLSVersion12 TLSVersion = "TLS12"
	TLSVersion13 TLSVersion = "TLS13"
)

// TLSConfig defines full TLS configuration.
type TLSConfig struct {
	SafeTLSConfig  `json:",inline"`
	TLSFilesConfig `json:",inline"`
}

// Validate semantically validates the given TLSConfig.
func (c *TLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if !reflect.ValueOf(c.CA).IsZero() {
		if c.CAFile != "" {
			return fmt.Errorf("cannot specify both 'caFile' and 'ca'")
		}

		if err := c.CA.Validate(); err != nil {
			return fmt.Errorf("ca: %w", err)
		}
	}

	hasCert := !reflect.ValueOf(c.Cert).IsZero()
	if hasCert {
		if c.CertFile != "" {
			return fmt.Errorf("cannot specify both 'certFile' and 'cert'")
		}

		if err := c.Cert.Validate(); err != nil {
			return fmt.Errorf("cert: %w", err)
		}
	}

	if c.KeyFile != "" && c.KeySecret != nil {
		return fmt.Errorf("cannot specify both 'keyFile' and 'keySecret'")
	}

	hasCert = hasCert || c.CertFile != ""
	hasKey := c.KeyFile != "" || c.KeySecret != nil

	if hasCert && !hasKey {
		return fmt.Errorf("cannot specify client cert without client key")
	}

	if hasKey && !hasCert {
		return fmt.Errorf("cannot specify client key without client cert")
	}

	if c.MaxVersion != nil && c.MinVersion != nil && strings.Compare(string(*c.MaxVersion), string(*c.MinVersion)) == -1 {
		return fmt.Errorf("'maxVersion' must greater than or equal to 'minVersion'")
	}

	return nil
}

// SafeTLSConfig defines safe TLS configurations.
// +k8s:openapi-gen=true
type SafeTLSConfig struct {
	// ca defines the Certificate authority used when verifying server certificates.
	// +optional
	CA SecretOrConfigMap `json:"ca,omitempty"`

	// cert defines the Client certificate to present when doing client-authentication.
	// +optional
	Cert SecretOrConfigMap `json:"cert,omitempty"`

	// keySecret defines the Secret containing the client key file for the targets.
	// +optional
	KeySecret *v1.SecretKeySelector `json:"keySecret,omitempty"`

	// serverName is used to verify the hostname for the targets.
	// +optional
	ServerName *string `json:"serverName,omitempty"`

	// insecureSkipVerify defines how to disable target certificate validation.
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"` // nolint:kubeapilinter

	// minVersion defines the minimum acceptable TLS version.
	//
	// It requires Prometheus >= v2.35.0 or Thanos >= v0.28.0.
	// +optional
	MinVersion *TLSVersion `json:"minVersion,omitempty"`

	// maxVersion defines the maximum acceptable TLS version.
	//
	// It requires Prometheus >= v2.41.0 or Thanos >= v0.31.0.
	// +optional
	MaxVersion *TLSVersion `json:"maxVersion,omitempty"`
}

// Validate semantically validates the given SafeTLSConfig.
func (c *SafeTLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.CA != (SecretOrConfigMap{}) {
		if err := c.CA.Validate(); err != nil {
			return fmt.Errorf("ca %s: %w", c.CA.String(), err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if err := c.Cert.Validate(); err != nil {
			return fmt.Errorf("cert %s: %w", c.Cert.String(), err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) && c.KeySecret == nil {
		return fmt.Errorf("client cert specified without client key")
	}

	if c.KeySecret != nil && c.Cert == (SecretOrConfigMap{}) {
		return fmt.Errorf("client key specified without client cert")
	}

	if c.MaxVersion != nil && c.MinVersion != nil && strings.Compare(string(*c.MaxVersion), string(*c.MinVersion)) == -1 {
		return fmt.Errorf("maxVersion must more than or equal to minVersion")
	}

	return nil
}

// TLSFilesConfig extends the TLS configuration with file parameters.
// +k8s:openapi-gen=true
type TLSFilesConfig struct {
	// caFile defines the path to the CA cert in the Prometheus container to use for the targets.
	// +optional
	CAFile string `json:"caFile,omitempty"`
	// certFile defines the path to the client cert file in the Prometheus container for the targets.
	// +optional
	CertFile string `json:"certFile,omitempty"`
	// keyFile defines the path to the client key file in the Prometheus container for the targets.
	// +optional
	KeyFile string `json:"keyFile,omitempty"`
}

//
