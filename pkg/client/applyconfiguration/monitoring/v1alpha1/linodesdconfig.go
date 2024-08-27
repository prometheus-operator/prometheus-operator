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

// LinodeSDConfigApplyConfiguration represents a declarative configuration of the LinodeSDConfig type for use
// with apply.
type LinodeSDConfigApplyConfiguration struct {
	Region                                     *string                                           `json:"region,omitempty"`
	Port                                       *int32                                            `json:"port,omitempty"`
	TagSeparator                               *string                                           `json:"tagSeparator,omitempty"`
	RefreshInterval                            *v1.Duration                                      `json:"refreshInterval,omitempty"`
	Authorization                              *monitoringv1.SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	OAuth2                                     *monitoringv1.OAuth2ApplyConfiguration            `json:"oauth2,omitempty"`
	monitoringv1.ProxyConfigApplyConfiguration `json:",inline"`
	FollowRedirects                            *bool                                         `json:"followRedirects,omitempty"`
	TLSConfig                                  *monitoringv1.SafeTLSConfigApplyConfiguration `json:"tlsConfig,omitempty"`
	EnableHTTP2                                *bool                                         `json:"enableHTTP2,omitempty"`
}

// LinodeSDConfigApplyConfiguration constructs a declarative configuration of the LinodeSDConfig type for use with
// apply.
func LinodeSDConfig() *LinodeSDConfigApplyConfiguration {
	return &LinodeSDConfigApplyConfiguration{}
}

// WithRegion sets the Region field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Region field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithRegion(value string) *LinodeSDConfigApplyConfiguration {
	b.Region = &value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithPort(value int32) *LinodeSDConfigApplyConfiguration {
	b.Port = &value
	return b
}

// WithTagSeparator sets the TagSeparator field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TagSeparator field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithTagSeparator(value string) *LinodeSDConfigApplyConfiguration {
	b.TagSeparator = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithRefreshInterval(value v1.Duration) *LinodeSDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithAuthorization(value *monitoringv1.SafeAuthorizationApplyConfiguration) *LinodeSDConfigApplyConfiguration {
	b.Authorization = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithOAuth2(value *monitoringv1.OAuth2ApplyConfiguration) *LinodeSDConfigApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithProxyURL(value string) *LinodeSDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithNoProxy(value string) *LinodeSDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *LinodeSDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *LinodeSDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *LinodeSDConfigApplyConfiguration {
	if b.ProxyConnectHeader == nil && len(entries) > 0 {
		b.ProxyConnectHeader = make(map[string][]corev1.SecretKeySelector, len(entries))
	}
	for k, v := range entries {
		b.ProxyConnectHeader[k] = v
	}
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithFollowRedirects(value bool) *LinodeSDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithTLSConfig(value *monitoringv1.SafeTLSConfigApplyConfiguration) *LinodeSDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *LinodeSDConfigApplyConfiguration) WithEnableHTTP2(value bool) *LinodeSDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
