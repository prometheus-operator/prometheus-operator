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

// IonosSDConfigApplyConfiguration represents a declarative configuration of the IonosSDConfig type for use
// with apply.
type IonosSDConfigApplyConfiguration struct {
	DataCenterID                               *string                                           `json:"datacenterID,omitempty"`
	Port                                       *int32                                            `json:"port,omitempty"`
	RefreshInterval                            *v1.Duration                                      `json:"refreshInterval,omitempty"`
	Authorization                              *monitoringv1.SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	monitoringv1.ProxyConfigApplyConfiguration `json:",inline"`
	TLSConfig                                  *monitoringv1.SafeTLSConfigApplyConfiguration `json:"tlsConfig,omitempty"`
	FollowRedirects                            *bool                                         `json:"followRedirects,omitempty"`
	EnableHTTP2                                *bool                                         `json:"enableHTTP2,omitempty"`
}

// IonosSDConfigApplyConfiguration constructs a declarative configuration of the IonosSDConfig type for use with
// apply.
func IonosSDConfig() *IonosSDConfigApplyConfiguration {
	return &IonosSDConfigApplyConfiguration{}
}

// WithDataCenterID sets the DataCenterID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DataCenterID field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithDataCenterID(value string) *IonosSDConfigApplyConfiguration {
	b.DataCenterID = &value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithPort(value int32) *IonosSDConfigApplyConfiguration {
	b.Port = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithRefreshInterval(value v1.Duration) *IonosSDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithAuthorization(value *monitoringv1.SafeAuthorizationApplyConfiguration) *IonosSDConfigApplyConfiguration {
	b.Authorization = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithProxyURL(value string) *IonosSDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithNoProxy(value string) *IonosSDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *IonosSDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *IonosSDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *IonosSDConfigApplyConfiguration {
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
func (b *IonosSDConfigApplyConfiguration) WithHTTPHeaders(entries map[string]monitoringv1.HTTPHeaderApplyConfiguration) *IonosSDConfigApplyConfiguration {
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
func (b *IonosSDConfigApplyConfiguration) WithTLSConfig(value *monitoringv1.SafeTLSConfigApplyConfiguration) *IonosSDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithFollowRedirects(value bool) *IonosSDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *IonosSDConfigApplyConfiguration) WithEnableHTTP2(value bool) *IonosSDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
