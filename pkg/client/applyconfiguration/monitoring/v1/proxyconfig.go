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
	v1 "k8s.io/api/core/v1"
)

// ProxyConfigApplyConfiguration represents a declarative configuration of the ProxyConfig type for use
// with apply.
type ProxyConfigApplyConfiguration struct {
	ProxyURL             *string                           `json:"proxyUrl,omitempty"`
	NoProxy              *string                           `json:"noProxy,omitempty"`
	ProxyFromEnvironment *bool                             `json:"proxyFromEnvironment,omitempty"`
	ProxyConnectHeader   map[string][]v1.SecretKeySelector `json:"proxyConnectHeader,omitempty"`
}

// ProxyConfigApplyConfiguration constructs a declarative configuration of the ProxyConfig type for use with
// apply.
func ProxyConfig() *ProxyConfigApplyConfiguration {
	return &ProxyConfigApplyConfiguration{}
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *ProxyConfigApplyConfiguration) WithProxyURL(value string) *ProxyConfigApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *ProxyConfigApplyConfiguration) WithNoProxy(value string) *ProxyConfigApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *ProxyConfigApplyConfiguration) WithProxyFromEnvironment(value bool) *ProxyConfigApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *ProxyConfigApplyConfiguration) WithProxyConnectHeader(entries map[string][]v1.SecretKeySelector) *ProxyConfigApplyConfiguration {
	if b.ProxyConnectHeader == nil && len(entries) > 0 {
		b.ProxyConnectHeader = make(map[string][]v1.SecretKeySelector, len(entries))
	}
	for k, v := range entries {
		b.ProxyConnectHeader[k] = v
	}
	return b
}
