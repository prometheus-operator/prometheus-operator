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

// PushoverConfigApplyConfiguration represents an declarative configuration of the PushoverConfig type for use
// with apply.
type PushoverConfigApplyConfiguration struct {
	SendResolved *bool                                `json:"sendResolved,omitempty"`
	UserKey      *SecretKeySelectorApplyConfiguration `json:"userKey,omitempty"`
	Token        *SecretKeySelectorApplyConfiguration `json:"token,omitempty"`
	Title        *string                              `json:"title,omitempty"`
	Message      *string                              `json:"message,omitempty"`
	URL          *string                              `json:"url,omitempty"`
	URLTitle     *string                              `json:"urlTitle,omitempty"`
	Device       *string                              `json:"device,omitempty"`
	Sound        *string                              `json:"sound,omitempty"`
	Priority     *string                              `json:"priority,omitempty"`
	Retry        *string                              `json:"retry,omitempty"`
	Expire       *string                              `json:"expire,omitempty"`
	HTML         *bool                                `json:"html,omitempty"`
	HTTPConfig   *HTTPConfigApplyConfiguration        `json:"httpConfig,omitempty"`
}

// PushoverConfigApplyConfiguration constructs an declarative configuration of the PushoverConfig type for use with
// apply.
func PushoverConfig() *PushoverConfigApplyConfiguration {
	return &PushoverConfigApplyConfiguration{}
}

// WithSendResolved sets the SendResolved field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SendResolved field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithSendResolved(value bool) *PushoverConfigApplyConfiguration {
	b.SendResolved = &value
	return b
}

// WithUserKey sets the UserKey field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UserKey field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithUserKey(value *SecretKeySelectorApplyConfiguration) *PushoverConfigApplyConfiguration {
	b.UserKey = value
	return b
}

// WithToken sets the Token field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Token field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithToken(value *SecretKeySelectorApplyConfiguration) *PushoverConfigApplyConfiguration {
	b.Token = value
	return b
}

// WithTitle sets the Title field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Title field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithTitle(value string) *PushoverConfigApplyConfiguration {
	b.Title = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithMessage(value string) *PushoverConfigApplyConfiguration {
	b.Message = &value
	return b
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithURL(value string) *PushoverConfigApplyConfiguration {
	b.URL = &value
	return b
}

// WithURLTitle sets the URLTitle field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URLTitle field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithURLTitle(value string) *PushoverConfigApplyConfiguration {
	b.URLTitle = &value
	return b
}

// WithDevice sets the Device field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Device field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithDevice(value string) *PushoverConfigApplyConfiguration {
	b.Device = &value
	return b
}

// WithSound sets the Sound field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Sound field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithSound(value string) *PushoverConfigApplyConfiguration {
	b.Sound = &value
	return b
}

// WithPriority sets the Priority field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Priority field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithPriority(value string) *PushoverConfigApplyConfiguration {
	b.Priority = &value
	return b
}

// WithRetry sets the Retry field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Retry field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithRetry(value string) *PushoverConfigApplyConfiguration {
	b.Retry = &value
	return b
}

// WithExpire sets the Expire field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Expire field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithExpire(value string) *PushoverConfigApplyConfiguration {
	b.Expire = &value
	return b
}

// WithHTML sets the HTML field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HTML field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithHTML(value bool) *PushoverConfigApplyConfiguration {
	b.HTML = &value
	return b
}

// WithHTTPConfig sets the HTTPConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HTTPConfig field is set to the value of the last call.
func (b *PushoverConfigApplyConfiguration) WithHTTPConfig(value *HTTPConfigApplyConfiguration) *PushoverConfigApplyConfiguration {
	b.HTTPConfig = value
	return b
}
