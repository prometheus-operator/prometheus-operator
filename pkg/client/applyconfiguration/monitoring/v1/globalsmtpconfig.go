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
	corev1 "k8s.io/api/core/v1"
)

// GlobalSMTPConfigApplyConfiguration represents a declarative configuration of the GlobalSMTPConfig type for use
// with apply.
type GlobalSMTPConfigApplyConfiguration struct {
	From         *string                          `json:"from,omitempty"`
	SmartHost    *HostPortApplyConfiguration      `json:"smartHost,omitempty"`
	Hello        *string                          `json:"hello,omitempty"`
	AuthUsername *string                          `json:"authUsername,omitempty"`
	AuthPassword *corev1.SecretKeySelector        `json:"authPassword,omitempty"`
	AuthIdentity *string                          `json:"authIdentity,omitempty"`
	AuthSecret   *corev1.SecretKeySelector        `json:"authSecret,omitempty"`
	RequireTLS   *bool                            `json:"requireTLS,omitempty"`
	TLSConfig    *SafeTLSConfigApplyConfiguration `json:"tlsConfig,omitempty"`
}

// GlobalSMTPConfigApplyConfiguration constructs a declarative configuration of the GlobalSMTPConfig type for use with
// apply.
func GlobalSMTPConfig() *GlobalSMTPConfigApplyConfiguration {
	return &GlobalSMTPConfigApplyConfiguration{}
}

// WithFrom sets the From field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the From field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithFrom(value string) *GlobalSMTPConfigApplyConfiguration {
	b.From = &value
	return b
}

// WithSmartHost sets the SmartHost field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SmartHost field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithSmartHost(value *HostPortApplyConfiguration) *GlobalSMTPConfigApplyConfiguration {
	b.SmartHost = value
	return b
}

// WithHello sets the Hello field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Hello field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithHello(value string) *GlobalSMTPConfigApplyConfiguration {
	b.Hello = &value
	return b
}

// WithAuthUsername sets the AuthUsername field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AuthUsername field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithAuthUsername(value string) *GlobalSMTPConfigApplyConfiguration {
	b.AuthUsername = &value
	return b
}

// WithAuthPassword sets the AuthPassword field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AuthPassword field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithAuthPassword(value corev1.SecretKeySelector) *GlobalSMTPConfigApplyConfiguration {
	b.AuthPassword = &value
	return b
}

// WithAuthIdentity sets the AuthIdentity field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AuthIdentity field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithAuthIdentity(value string) *GlobalSMTPConfigApplyConfiguration {
	b.AuthIdentity = &value
	return b
}

// WithAuthSecret sets the AuthSecret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AuthSecret field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithAuthSecret(value corev1.SecretKeySelector) *GlobalSMTPConfigApplyConfiguration {
	b.AuthSecret = &value
	return b
}

// WithRequireTLS sets the RequireTLS field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RequireTLS field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithRequireTLS(value bool) *GlobalSMTPConfigApplyConfiguration {
	b.RequireTLS = &value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *GlobalSMTPConfigApplyConfiguration) WithTLSConfig(value *SafeTLSConfigApplyConfiguration) *GlobalSMTPConfigApplyConfiguration {
	b.TLSConfig = value
	return b
}
