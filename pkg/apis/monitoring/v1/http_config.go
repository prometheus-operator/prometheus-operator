// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// HTTPConfig defines the configuration for the HTTP client.
type HTTPConfig struct {
	// authorization configures the Authorization header credentials used by
	// the client.
	//
	// Cannot be set at the same time as `basicAuth`, `bearerTokenSecret` or `oauth2`.
	//
	// +optional
	Authorization *SafeAuthorization `json:"authorization,omitempty"`

	// basicAuth defines the Basic Authentication credentials used by the
	// client.
	//
	// Cannot be set at the same time as `authorization`, `bearerTokenSecret` or `oauth2`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// oauth2 defines the OAuth2 settings used by the client.
	//
	// It requires Prometheus >= 2.27.0.
	//
	// Cannot be set at the same time as `authorization`, `basicAuth` or `bearerTokenSecret`.
	//
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`

	// bearerTokenSecret defines a key of a Secret containing the bearer token
	// used by the client for authentication. The secret needs to be in the
	// same namespace as the custom resource and readable by the Prometheus
	// Operator.
	//
	// Cannot be set at the same time as `authorization`, `basicAuth` or `oauth2`.
	//
	// +optional
	//
	// Deprecated: use `authorization` instead.
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`

	// tlsConfig defines the TLS configuration used by the client.
	//
	// +optional
	TLSConfig *SafeTLSConfig `json:"tlsConfig,omitempty"`

	ProxyConfig `json:",inline"`

	// followRedirects defines whether the client should follow HTTP 3xx
	// redirects.
	//
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`

	// enableHttp2 can be used to disable HTTP2.
	//
	// +optional
	EnableHTTP2 *bool `json:"enableHttp2,omitempty"`
}

// Validate semantically validates the given HTTPConfig.
func (hc *HTTPConfig) Validate() error {
	if hc == nil {
		return nil
	}

	if (hc.BasicAuth != nil || hc.OAuth2 != nil) && (hc.BearerTokenSecret != nil) {
		return fmt.Errorf("at most one of basicAuth, oauth2, bearerTokenSecret must be configured")
	}

	if hc.Authorization != nil {
		if hc.BearerTokenSecret != nil {
			return fmt.Errorf("authorization is not compatible with bearerTokenSecret")
		}

		if hc.BasicAuth != nil || hc.OAuth2 != nil {
			return fmt.Errorf("at most one of basicAuth, oauth2 & authorization must be configured")
		}

		if err := hc.Authorization.Validate(); err != nil {
			return err
		}
	}

	if hc.OAuth2 != nil {
		if hc.BasicAuth != nil {
			return fmt.Errorf("at most one of basicAuth, oauth2 & authorization must be configured")
		}

		if err := hc.OAuth2.Validate(); err != nil {
			return err
		}
	}

	if hc.TLSConfig != nil {
		if err := hc.TLSConfig.Validate(); err != nil {
			return err
		}
	}

	if err := hc.ProxyConfig.Validate(); err != nil {
		return err
	}

	return nil
}
