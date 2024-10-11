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

// PuppetDBSDConfigApplyConfiguration represents a declarative configuration of the PuppetDBSDConfig type for use
// with apply.
type PuppetDBSDConfigApplyConfiguration struct {
	URL                                             *string                                           `json:"url,omitempty"`
	Query                                           *string                                           `json:"query,omitempty"`
	IncludeParameters                               *bool                                             `json:"includeParameters,omitempty"`
	RefreshInterval                                 *v1.Duration                                      `json:"refreshInterval,omitempty"`
	Port                                            *int32                                            `json:"port,omitempty"`
	BasicAuth                                       *monitoringv1.BasicAuthApplyConfiguration         `json:"basicAuth,omitempty"`
	Authorization                                   *monitoringv1.SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	OAuth2                                          *monitoringv1.OAuth2ApplyConfiguration            `json:"oauth2,omitempty"`
	monitoringv1.ProxyConfigApplyConfiguration      `json:",inline"`
	monitoringv1.CustomHTTPConfigApplyConfiguration `json:",inline"`
	TLSConfig                                       *monitoringv1.SafeTLSConfigApplyConfiguration `json:"tlsConfig,omitempty"`
	FollowRedirects                                 *bool                                         `json:"followRedirects,omitempty"`
	EnableHTTP2                                     *bool                                         `json:"enableHTTP2,omitempty"`
}

// PuppetDBSDConfigApplyConfiguration constructs a declarative configuration of the PuppetDBSDConfig type for use with
// apply.
func PuppetDBSDConfig() *PuppetDBSDConfigApplyConfiguration {
	return &PuppetDBSDConfigApplyConfiguration{}
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithURL(value string) *PuppetDBSDConfigApplyConfiguration {
	b.URL = &value
	return b
}

// WithQuery sets the Query field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Query field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithQuery(value string) *PuppetDBSDConfigApplyConfiguration {
	b.Query = &value
	return b
}

// WithIncludeParameters sets the IncludeParameters field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IncludeParameters field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithIncludeParameters(value bool) *PuppetDBSDConfigApplyConfiguration {
	b.IncludeParameters = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithRefreshInterval(value v1.Duration) *PuppetDBSDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithPort(value int32) *PuppetDBSDConfigApplyConfiguration {
	b.Port = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithBasicAuth(value *monitoringv1.BasicAuthApplyConfiguration) *PuppetDBSDConfigApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithAuthorization(value *monitoringv1.SafeAuthorizationApplyConfiguration) *PuppetDBSDConfigApplyConfiguration {
	b.Authorization = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithOAuth2(value *monitoringv1.OAuth2ApplyConfiguration) *PuppetDBSDConfigApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithProxyURL(value string) *PuppetDBSDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithNoProxy(value string) *PuppetDBSDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *PuppetDBSDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *PuppetDBSDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *PuppetDBSDConfigApplyConfiguration {
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
func (b *PuppetDBSDConfigApplyConfiguration) WithHTTPHeaders(entries map[string]monitoringv1.HTTPHeaderApplyConfiguration) *PuppetDBSDConfigApplyConfiguration {
	if b.HTTPHeaders == nil && len(entries) > 0 {
		b.HTTPHeaders = make(map[string]monitoringv1.HTTPHeaderApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.HTTPHeaders[k] = v
	}
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithTLSConfig(value *monitoringv1.SafeTLSConfigApplyConfiguration) *PuppetDBSDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithFollowRedirects(value bool) *PuppetDBSDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *PuppetDBSDConfigApplyConfiguration) WithEnableHTTP2(value bool) *PuppetDBSDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
