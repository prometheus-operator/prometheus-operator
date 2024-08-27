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

// SlackActionApplyConfiguration represents a declarative configuration of the SlackAction type for use
// with apply.
type SlackActionApplyConfiguration struct {
	Type         *string                                   `json:"type,omitempty"`
	Text         *string                                   `json:"text,omitempty"`
	URL          *string                                   `json:"url,omitempty"`
	Style        *string                                   `json:"style,omitempty"`
	Name         *string                                   `json:"name,omitempty"`
	Value        *string                                   `json:"value,omitempty"`
	ConfirmField *SlackConfirmationFieldApplyConfiguration `json:"confirm,omitempty"`
}

// SlackActionApplyConfiguration constructs a declarative configuration of the SlackAction type for use with
// apply.
func SlackAction() *SlackActionApplyConfiguration {
	return &SlackActionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithType(value string) *SlackActionApplyConfiguration {
	b.Type = &value
	return b
}

// WithText sets the Text field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Text field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithText(value string) *SlackActionApplyConfiguration {
	b.Text = &value
	return b
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithURL(value string) *SlackActionApplyConfiguration {
	b.URL = &value
	return b
}

// WithStyle sets the Style field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Style field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithStyle(value string) *SlackActionApplyConfiguration {
	b.Style = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithName(value string) *SlackActionApplyConfiguration {
	b.Name = &value
	return b
}

// WithValue sets the Value field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Value field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithValue(value string) *SlackActionApplyConfiguration {
	b.Value = &value
	return b
}

// WithConfirmField sets the ConfirmField field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ConfirmField field is set to the value of the last call.
func (b *SlackActionApplyConfiguration) WithConfirmField(value *SlackConfirmationFieldApplyConfiguration) *SlackActionApplyConfiguration {
	b.ConfirmField = value
	return b
}
