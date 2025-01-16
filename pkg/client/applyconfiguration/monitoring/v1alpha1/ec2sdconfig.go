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
	applyconfigurationmonitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	v1 "k8s.io/api/core/v1"
)

// EC2SDConfigApplyConfiguration represents a declarative configuration of the EC2SDConfig type for use
// with apply.
type EC2SDConfigApplyConfiguration struct {
	Region                                                            *string                `json:"region,omitempty"`
	AccessKey                                                         *v1.SecretKeySelector  `json:"accessKey,omitempty"`
	SecretKey                                                         *v1.SecretKeySelector  `json:"secretKey,omitempty"`
	RoleARN                                                           *string                `json:"roleARN,omitempty"`
	Port                                                              *int32                 `json:"port,omitempty"`
	RefreshInterval                                                   *monitoringv1.Duration `json:"refreshInterval,omitempty"`
	Filters                                                           *v1alpha1.Filters      `json:"filters,omitempty"`
	applyconfigurationmonitoringv1.ProxyConfigApplyConfiguration      `json:",inline"`
	applyconfigurationmonitoringv1.InlineHTTPConfigApplyConfiguration `json:",inline"`
	TLSConfig                                                         *applyconfigurationmonitoringv1.SafeTLSConfigApplyConfiguration `json:"tlsConfig,omitempty"`
	FollowRedirects                                                   *bool                                                           `json:"followRedirects,omitempty"`
	EnableHTTP2                                                       *bool                                                           `json:"enableHTTP2,omitempty"`
}

// EC2SDConfigApplyConfiguration constructs a declarative configuration of the EC2SDConfig type for use with
// apply.
func EC2SDConfig() *EC2SDConfigApplyConfiguration {
	return &EC2SDConfigApplyConfiguration{}
}

// WithRegion sets the Region field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Region field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithRegion(value string) *EC2SDConfigApplyConfiguration {
	b.Region = &value
	return b
}

// WithAccessKey sets the AccessKey field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AccessKey field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithAccessKey(value v1.SecretKeySelector) *EC2SDConfigApplyConfiguration {
	b.AccessKey = &value
	return b
}

// WithSecretKey sets the SecretKey field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SecretKey field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithSecretKey(value v1.SecretKeySelector) *EC2SDConfigApplyConfiguration {
	b.SecretKey = &value
	return b
}

// WithRoleARN sets the RoleARN field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RoleARN field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithRoleARN(value string) *EC2SDConfigApplyConfiguration {
	b.RoleARN = &value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithPort(value int32) *EC2SDConfigApplyConfiguration {
	b.Port = &value
	return b
}

// WithRefreshInterval sets the RefreshInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RefreshInterval field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithRefreshInterval(value monitoringv1.Duration) *EC2SDConfigApplyConfiguration {
	b.RefreshInterval = &value
	return b
}

// WithFilters sets the Filters field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Filters field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithFilters(value v1alpha1.Filters) *EC2SDConfigApplyConfiguration {
	b.Filters = &value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithProxyURL(value string) *EC2SDConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithNoProxy(value string) *EC2SDConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *EC2SDConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *EC2SDConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]v1.SecretKeySelector) *EC2SDConfigApplyConfiguration {
	if b.ProxyConnectHeader == nil && len(entries) > 0 {
		b.ProxyConnectHeader = make(map[string][]v1.SecretKeySelector, len(entries))
	}
	for k, v := range entries {
		b.ProxyConnectHeader[k] = v
	}
	return b
}

// WithHTTPHeaders adds the given value to the HTTPHeaders field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the HTTPHeaders field.
func (b *EC2SDConfigApplyConfiguration) WithHTTPHeaders(values ...*applyconfigurationmonitoringv1.HTTPHeaderApplyConfiguration) *EC2SDConfigApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithHTTPHeaders")
		}
		b.HTTPHeaders = append(b.HTTPHeaders, *values[i])
	}
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithTLSConfig(value *applyconfigurationmonitoringv1.SafeTLSConfigApplyConfiguration) *EC2SDConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithFollowRedirects(value bool) *EC2SDConfigApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *EC2SDConfigApplyConfiguration) WithEnableHTTP2(value bool) *EC2SDConfigApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}
