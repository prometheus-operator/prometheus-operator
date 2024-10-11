// Copyright The prometheus-operator Authors
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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

// DockerSDConfigApplyConfiguration represents a declarative configuration of the DockerSDConfig type for use
// with apply.
type DockerSDConfigApplyConfiguration struct {
	Host                                  *string `json:"host,omitempty"`
	v1.ProxyConfigApplyConfiguration      `json:",inline"`
	v1.CustomHTTPConfigApplyConfiguration `json:",inline"`
	TLSConfig                             *v1.SafeTLSConfigApplyConfiguration     `json:"tlsConfig,omitempty"`
	Port                                  *int                                    `json:"port,omitempty"`
	HostNetworkingHost                    *string                                 `json:"hostNetworkingHost,omitempty"`
	MatchFirstNetwork                     *bool                                   `json:"matchFirstNetwork,omitempty"`
	Filters                               *v1alpha1.Filters                       `json:"filters,omitempty"`
	RefreshInterval                       *monitoringv1.Duration                  `json:"refreshInterval,omitempty"`
	BasicAuth                             *v1.BasicAuthApplyConfiguration         `json:"basicAuth,omitempty"`
	Authorization                         *v1.SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	OAuth2                                *v1.OAuth2ApplyConfiguration            `json:"oauth2,omitempty"`
	FollowRedirects                       *bool                                   `json:"followRedirects,omitempty"`
	EnableHTTP2                           *bool                                   `json:"enableHTTP2,omitempty"`
}

// DockerSDConfigApplyConfiguration constructs a declarative configuration of the DockerSDConfig type for use with
// apply.
func DockerSDConfig() *DockerSDConfigApplyConfiguration {
	return &DockerSDConfigApplyConfiguration{}
}

// WithHost sets the Host field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Host field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithHost(value string) *DockerSDConfigApplyConfiguration {
	b.Host = &value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithProxyURL(value string) *DockerSDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithNoProxy(value string) *DockerSDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *DockerSDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *DockerSDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *DockerSDConfigApplyConfiguration {
	if b.ProxyConnectHeader == nil && len(entries) > 0 {
		b.ProxyConnectHeader = make(map[string][]corev1.SecretKeySelector, len(entries))
	}
	for k, v := range entries {
		b.ProxyConnectHeader[k] = v
	}
	return b
}

// WithHTTPHeaders puts the entries into the HTTPHeaders field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the HTTPHeaders field,
// overwriting an existing map entries in HTTPHeaders field with the same key.
func (b *DockerSDConfigApplyConfiguration) WithHTTPHeaders(entries map[string]v1.HTTPHeaderApplyConfiguration) *DockerSDConfigApplyConfiguration {
	if b.HTTPHeaders == nil && len(entries) > 0 {
		b.HTTPHeaders = make(map[string]v1.HTTPHeaderApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.HTTPHeaders[k] = v
	}
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithTLSConfig(value *v1.SafeTLSConfigApplyConfiguration) *DockerSDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithPort(value int) *DockerSDConfigApplyConfiguration {
	b.Port = &value
	return b
}

// WithHostNetworkingHost sets the HostNetworkingHost field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HostNetworkingHost field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithHostNetworkingHost(value string) *DockerSDConfigApplyConfiguration {
	b.HostNetworkingHost = &value
	return b
}

// WithMatchFirstNetwork sets the MatchFirstNetwork field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MatchFirstNetwork field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithMatchFirstNetwork(value bool) *DockerSDConfigApplyConfiguration {
	b.MatchFirstNetwork = &value
	return b
}

// WithFilters sets the Filters field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Filters field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithFilters(value v1alpha1.Filters) *DockerSDConfigApplyConfiguration {
	b.Filters = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithRefreshInterval(value monitoringv1.Duration) *DockerSDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithBasicAuth(value *v1.BasicAuthApplyConfiguration) *DockerSDConfigApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithAuthorization(value *v1.SafeAuthorizationApplyConfiguration) *DockerSDConfigApplyConfiguration {
	b.Authorization = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithOAuth2(value *v1.OAuth2ApplyConfiguration) *DockerSDConfigApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithFollowRedirects(value bool) *DockerSDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *DockerSDConfigApplyConfiguration) WithEnableHTTP2(value bool) *DockerSDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
