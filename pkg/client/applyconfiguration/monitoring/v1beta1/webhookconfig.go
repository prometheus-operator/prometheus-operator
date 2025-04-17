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

package v1beta1

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// WebhookConfigApplyConfiguration represents a declarative configuration of the WebhookConfig type for use
// with apply.
type WebhookConfigApplyConfiguration struct {
	SendResolved *bool                                `json:"sendResolved,omitempty"`
	URL          *string                              `json:"url,omitempty"`
	URLSecret    *SecretKeySelectorApplyConfiguration `json:"urlSecret,omitempty"`
	HTTPConfig   *HTTPConfigApplyConfiguration        `json:"httpConfig,omitempty"`
	MaxAlerts    *int32                               `json:"maxAlerts,omitempty"`
	Timeout      *v1.Duration                         `json:"timeout,omitempty"`
}

// WebhookConfigApplyConfiguration constructs a declarative configuration of the WebhookConfig type for use with
// apply.
func WebhookConfig() *WebhookConfigApplyConfiguration {
	return &WebhookConfigApplyConfiguration{}
}

// WithSendResolved sets the SendResolved field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SendResolved field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithSendResolved(value bool) *WebhookConfigApplyConfiguration {
	b.SendResolved = &value
	return b
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithURL(value string) *WebhookConfigApplyConfiguration {
	b.URL = &value
	return b
}

// WithURLSecret sets the URLSecret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URLSecret field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithURLSecret(value *SecretKeySelectorApplyConfiguration) *WebhookConfigApplyConfiguration {
	b.URLSecret = value
	return b
}

// WithHTTPConfig sets the HTTPConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HTTPConfig field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithHTTPConfig(value *HTTPConfigApplyConfiguration) *WebhookConfigApplyConfiguration {
	b.HTTPConfig = value
	return b
}

// WithMaxAlerts sets the MaxAlerts field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MaxAlerts field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithMaxAlerts(value int32) *WebhookConfigApplyConfiguration {
	b.MaxAlerts = &value
	return b
}

// WithTimeout sets the Timeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Timeout field is set to the value of the last call.
func (b *WebhookConfigApplyConfiguration) WithTimeout(value v1.Duration) *WebhookConfigApplyConfiguration {
	b.Timeout = &value
	return b
}
