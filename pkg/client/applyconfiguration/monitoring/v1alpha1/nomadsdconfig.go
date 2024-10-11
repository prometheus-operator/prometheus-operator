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
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

// NomadSDConfigApplyConfiguration represents a declarative configuration of the NomadSDConfig type for use
// with apply.
type NomadSDConfigApplyConfiguration struct {
	AllowStale                                      *bool                                             `json:"allowStale,omitempty"`
	Namespace                                       *string                                           `json:"namespace,omitempty"`
	RefreshInterval                                 *v1.Duration                                      `json:"refreshInterval,omitempty"`
	Region                                          *string                                           `json:"region,omitempty"`
	Server                                          *string                                           `json:"server,omitempty"`
	TagSeparator                                    *string                                           `json:"tagSeparator,omitempty"`
	BasicAuth                                       *monitoringv1.BasicAuthApplyConfiguration         `json:"basicAuth,omitempty"`
	Authorization                                   *monitoringv1.SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	OAuth2                                          *monitoringv1.OAuth2ApplyConfiguration            `json:"oauth2,omitempty"`
	TLSConfig                                       *monitoringv1.SafeTLSConfigApplyConfiguration     `json:"tlsConfig,omitempty"`
	monitoringv1.ProxyConfigApplyConfiguration      `json:",inline"`
	monitoringv1.CustomHTTPConfigApplyConfiguration `json:",inline"`
	FollowRedirects                                 *bool `json:"followRedirects,omitempty"`
	EnableHTTP2                                     *bool `json:"enableHTTP2,omitempty"`
}

// NomadSDConfigApplyConfiguration constructs a declarative configuration of the NomadSDConfig type for use with
// apply.
func NomadSDConfig() *NomadSDConfigApplyConfiguration {
	return &NomadSDConfigApplyConfiguration{}
}

// WithAllowStale sets the AllowStale field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AllowStale field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithAllowStale(value bool) *NomadSDConfigApplyConfiguration {
	b.AllowStale = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithNamespace(value string) *NomadSDConfigApplyConfiguration {
	b.Namespace = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithRefreshInterval(value v1.Duration) *NomadSDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithRegion sets the Region field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Region field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithRegion(value string) *NomadSDConfigApplyConfiguration {
	b.Region = &value
	return b
}

// WithServer sets the Server field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Server field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithServer(value string) *NomadSDConfigApplyConfiguration {
	b.Server = &value
	return b
}

// WithTagSeparator sets the TagSeparator field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TagSeparator field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithTagSeparator(value string) *NomadSDConfigApplyConfiguration {
	b.TagSeparator = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithBasicAuth(value *monitoringv1.BasicAuthApplyConfiguration) *NomadSDConfigApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithAuthorization(value *monitoringv1.SafeAuthorizationApplyConfiguration) *NomadSDConfigApplyConfiguration {
	b.Authorization = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithOAuth2(value *monitoringv1.OAuth2ApplyConfiguration) *NomadSDConfigApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithTLSConfig(value *monitoringv1.SafeTLSConfigApplyConfiguration) *NomadSDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithProxyURL(value string) *NomadSDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithNoProxy(value string) *NomadSDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *NomadSDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *NomadSDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *NomadSDConfigApplyConfiguration {
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
func (b *NomadSDConfigApplyConfiguration) WithHTTPHeaders(entries map[string]monitoringv1.HTTPHeaderApplyConfiguration) *NomadSDConfigApplyConfiguration {
	if b.HTTPHeaders == nil && len(entries) > 0 {
		b.HTTPHeaders = make(map[string]monitoringv1.HTTPHeaderApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.HTTPHeaders[k] = v
	}
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithFollowRedirects(value bool) *NomadSDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *NomadSDConfigApplyConfiguration) WithEnableHTTP2(value bool) *NomadSDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
