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

package v1

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// RemoteReadSpecApplyConfiguration represents an declarative configuration of the RemoteReadSpec type for use
// with apply.
type RemoteReadSpecApplyConfiguration struct {
	URL                  *string                          `json:"url,omitempty"`
	Name                 *string                          `json:"name,omitempty"`
	RequiredMatchers     map[string]string                `json:"requiredMatchers,omitempty"`
	RemoteTimeout        *v1.Duration                     `json:"remoteTimeout,omitempty"`
	Headers              map[string]string                `json:"headers,omitempty"`
	ReadRecent           *bool                            `json:"readRecent,omitempty"`
	BasicAuth            *BasicAuthApplyConfiguration     `json:"basicAuth,omitempty"`
	OAuth2               *OAuth2ApplyConfiguration        `json:"oauth2,omitempty"`
	BearerToken          *string                          `json:"bearerToken,omitempty"`
	BearerTokenFile      *string                          `json:"bearerTokenFile,omitempty"`
	Authorization        *AuthorizationApplyConfiguration `json:"authorization,omitempty"`
	TLSConfig            *TLSConfigApplyConfiguration     `json:"tlsConfig,omitempty"`
	ProxyURL             *string                          `json:"proxyUrl,omitempty"`
	FollowRedirects      *bool                            `json:"followRedirects,omitempty"`
	FilterExternalLabels *bool                            `json:"filterExternalLabels,omitempty"`
}

// RemoteReadSpecApplyConfiguration constructs an declarative configuration of the RemoteReadSpec type for use with
// apply.
func RemoteReadSpec() *RemoteReadSpecApplyConfiguration {
	return &RemoteReadSpecApplyConfiguration{}
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithURL(value string) *RemoteReadSpecApplyConfiguration {
	b.URL = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithName(value string) *RemoteReadSpecApplyConfiguration {
	b.Name = &value
	return b
}

// WithRequiredMatchers puts the entries into the RequiredMatchers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the RequiredMatchers field,
// overwriting an existing map entries in RequiredMatchers field with the same key.
func (b *RemoteReadSpecApplyConfiguration) WithRequiredMatchers(entries map[string]string) *RemoteReadSpecApplyConfiguration {
	if b.RequiredMatchers == nil && len(entries) > 0 {
		b.RequiredMatchers = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.RequiredMatchers[k] = v
	}
	return b
}

// WithRemoteTimeout sets the RemoteTimeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RemoteTimeout field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithRemoteTimeout(value v1.Duration) *RemoteReadSpecApplyConfiguration {
	b.RemoteTimeout = &value
	return b
}

// WithHeaders puts the entries into the Headers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Headers field,
// overwriting an existing map entries in Headers field with the same key.
func (b *RemoteReadSpecApplyConfiguration) WithHeaders(entries map[string]string) *RemoteReadSpecApplyConfiguration {
	if b.Headers == nil && len(entries) > 0 {
		b.Headers = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Headers[k] = v
	}
	return b
}

// WithReadRecent sets the ReadRecent field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ReadRecent field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithReadRecent(value bool) *RemoteReadSpecApplyConfiguration {
	b.ReadRecent = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithBasicAuth(value *BasicAuthApplyConfiguration) *RemoteReadSpecApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithOAuth2(value *OAuth2ApplyConfiguration) *RemoteReadSpecApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithBearerToken sets the BearerToken field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BearerToken field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithBearerToken(value string) *RemoteReadSpecApplyConfiguration {
	b.BearerToken = &value
	return b
}

// WithBearerTokenFile sets the BearerTokenFile field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BearerTokenFile field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithBearerTokenFile(value string) *RemoteReadSpecApplyConfiguration {
	b.BearerTokenFile = &value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithAuthorization(value *AuthorizationApplyConfiguration) *RemoteReadSpecApplyConfiguration {
	b.Authorization = value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithTLSConfig(value *TLSConfigApplyConfiguration) *RemoteReadSpecApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithProxyURL(value string) *RemoteReadSpecApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithFollowRedirects(value bool) *RemoteReadSpecApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithFilterExternalLabels sets the FilterExternalLabels field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FilterExternalLabels field is set to the value of the last call.
func (b *RemoteReadSpecApplyConfiguration) WithFilterExternalLabels(value bool) *RemoteReadSpecApplyConfiguration {
	b.FilterExternalLabels = &value
	return b
}
