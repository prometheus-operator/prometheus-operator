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
	v1 "k8s.io/api/core/v1"
)

// RocketChatConfigApplyConfiguration represents a declarative configuration of the RocketChatConfig type for use
// with apply.
type RocketChatConfigApplyConfiguration struct {
	SendResolved *bool                                      `json:"sendResolved,omitempty"`
	APIURL       *string                                    `json:"apiURL,omitempty"`
	Channel      *string                                    `json:"channel,omitempty"`
	Token        *v1.SecretKeySelector                      `json:"token,omitempty"`
	TokenFile    *string                                    `json:"tokenFile,omitempty"`
	TokenID      *v1.SecretKeySelector                      `json:"tokenID,omitempty"`
	TokenIDFile  *string                                    `json:"tokenIDFile,omitempty"`
	Color        *string                                    `json:"color,omitempty"`
	Emoji        *string                                    `json:"emoji,omitempty"`
	IconURL      *string                                    `json:"iconURL,omitempty"`
	Text         *string                                    `json:"text,omitempty"`
	Title        *string                                    `json:"title,omitempty"`
	TitleLink    *string                                    `json:"titleLink,omitempty"`
	Fields       []RocketChatFieldConfigApplyConfiguration  `json:"fields,omitempty"`
	ShortFields  *bool                                      `json:"shortFields,omitempty"`
	ImageURL     *string                                    `json:"imageURL,omitempty"`
	ThumbURL     *string                                    `json:"thumbURL,omitempty"`
	LinkNames    *bool                                      `json:"linkNames,omitempty"`
	Actions      []RocketChatActionConfigApplyConfiguration `json:"actions,omitempty"`
	HTTPConfig   *HTTPConfigApplyConfiguration              `json:"httpConfig,omitempty"`
}

// RocketChatConfigApplyConfiguration constructs a declarative configuration of the RocketChatConfig type for use with
// apply.
func RocketChatConfig() *RocketChatConfigApplyConfiguration {
	return &RocketChatConfigApplyConfiguration{}
}

// WithSendResolved sets the SendResolved field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SendResolved field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithSendResolved(value bool) *RocketChatConfigApplyConfiguration {
	b.SendResolved = &value
	return b
}

// WithAPIURL sets the APIURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIURL field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithAPIURL(value string) *RocketChatConfigApplyConfiguration {
	b.APIURL = &value
	return b
}

// WithChannel sets the Channel field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Channel field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithChannel(value string) *RocketChatConfigApplyConfiguration {
	b.Channel = &value
	return b
}

// WithToken sets the Token field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Token field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithToken(value v1.SecretKeySelector) *RocketChatConfigApplyConfiguration {
	b.Token = &value
	return b
}

// WithTokenFile sets the TokenFile field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TokenFile field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithTokenFile(value string) *RocketChatConfigApplyConfiguration {
	b.TokenFile = &value
	return b
}

// WithTokenID sets the TokenID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TokenID field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithTokenID(value v1.SecretKeySelector) *RocketChatConfigApplyConfiguration {
	b.TokenID = &value
	return b
}

// WithTokenIDFile sets the TokenIDFile field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TokenIDFile field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithTokenIDFile(value string) *RocketChatConfigApplyConfiguration {
	b.TokenIDFile = &value
	return b
}

// WithColor sets the Color field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Color field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithColor(value string) *RocketChatConfigApplyConfiguration {
	b.Color = &value
	return b
}

// WithEmoji sets the Emoji field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Emoji field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithEmoji(value string) *RocketChatConfigApplyConfiguration {
	b.Emoji = &value
	return b
}

// WithIconURL sets the IconURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IconURL field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithIconURL(value string) *RocketChatConfigApplyConfiguration {
	b.IconURL = &value
	return b
}

// WithText sets the Text field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Text field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithText(value string) *RocketChatConfigApplyConfiguration {
	b.Text = &value
	return b
}

// WithTitle sets the Title field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Title field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithTitle(value string) *RocketChatConfigApplyConfiguration {
	b.Title = &value
	return b
}

// WithTitleLink sets the TitleLink field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TitleLink field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithTitleLink(value string) *RocketChatConfigApplyConfiguration {
	b.TitleLink = &value
	return b
}

// WithFields adds the given value to the Fields field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Fields field.
func (b *RocketChatConfigApplyConfiguration) WithFields(values ...*RocketChatFieldConfigApplyConfiguration) *RocketChatConfigApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithFields")
		}
		b.Fields = append(b.Fields, *values[i])
	}
	return b
}

// WithShortFields sets the ShortFields field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ShortFields field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithShortFields(value bool) *RocketChatConfigApplyConfiguration {
	b.ShortFields = &value
	return b
}

// WithImageURL sets the ImageURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ImageURL field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithImageURL(value string) *RocketChatConfigApplyConfiguration {
	b.ImageURL = &value
	return b
}

// WithThumbURL sets the ThumbURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ThumbURL field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithThumbURL(value string) *RocketChatConfigApplyConfiguration {
	b.ThumbURL = &value
	return b
}

// WithLinkNames sets the LinkNames field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LinkNames field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithLinkNames(value bool) *RocketChatConfigApplyConfiguration {
	b.LinkNames = &value
	return b
}

// WithActions adds the given value to the Actions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Actions field.
func (b *RocketChatConfigApplyConfiguration) WithActions(values ...*RocketChatActionConfigApplyConfiguration) *RocketChatConfigApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithActions")
		}
		b.Actions = append(b.Actions, *values[i])
	}
	return b
}

// WithHTTPConfig sets the HTTPConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HTTPConfig field is set to the value of the last call.
func (b *RocketChatConfigApplyConfiguration) WithHTTPConfig(value *HTTPConfigApplyConfiguration) *RocketChatConfigApplyConfiguration {
	b.HTTPConfig = value
	return b
}
