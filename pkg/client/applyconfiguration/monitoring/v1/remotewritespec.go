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
	corev1 "k8s.io/api/core/v1"
)

// RemoteWriteSpecApplyConfiguration represents a declarative configuration of the RemoteWriteSpec type for use
// with apply.
type RemoteWriteSpecApplyConfiguration struct {
	URL                                *string                           `json:"url,omitempty"`
	Name                               *string                           `json:"name,omitempty"`
	MessageVersion                     *v1.RemoteWriteMessageVersion     `json:"messageVersion,omitempty"`
	SendExemplars                      *bool                             `json:"sendExemplars,omitempty"`
	SendNativeHistograms               *bool                             `json:"sendNativeHistograms,omitempty"`
	RemoteTimeout                      *v1.Duration                      `json:"remoteTimeout,omitempty"`
	Headers                            map[string]string                 `json:"headers,omitempty"`
	WriteRelabelConfigs                []RelabelConfigApplyConfiguration `json:"writeRelabelConfigs,omitempty"`
	OAuth2                             *OAuth2ApplyConfiguration         `json:"oauth2,omitempty"`
	BasicAuth                          *BasicAuthApplyConfiguration      `json:"basicAuth,omitempty"`
	BearerTokenFile                    *string                           `json:"bearerTokenFile,omitempty"`
	Authorization                      *AuthorizationApplyConfiguration  `json:"authorization,omitempty"`
	Sigv4                              *Sigv4ApplyConfiguration          `json:"sigv4,omitempty"`
	AzureAD                            *AzureADApplyConfiguration        `json:"azureAd,omitempty"`
	BearerToken                        *string                           `json:"bearerToken,omitempty"`
	TLSConfig                          *TLSConfigApplyConfiguration      `json:"tlsConfig,omitempty"`
	ProxyConfigApplyConfiguration      `json:",inline"`
	InlineHTTPConfigApplyConfiguration `json:",inline"`
	FollowRedirects                    *bool                             `json:"followRedirects,omitempty"`
	QueueConfig                        *QueueConfigApplyConfiguration    `json:"queueConfig,omitempty"`
	MetadataConfig                     *MetadataConfigApplyConfiguration `json:"metadataConfig,omitempty"`
	EnableHttp2                        *bool                             `json:"enableHTTP2,omitempty"`
}

// RemoteWriteSpecApplyConfiguration constructs a declarative configuration of the RemoteWriteSpec type for use with
// apply.
func RemoteWriteSpec() *RemoteWriteSpecApplyConfiguration {
	return &RemoteWriteSpecApplyConfiguration{}
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithURL(value string) *RemoteWriteSpecApplyConfiguration {
	b.URL = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithName(value string) *RemoteWriteSpecApplyConfiguration {
	b.Name = &value
	return b
}

// WithMessageVersion sets the MessageVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MessageVersion field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithMessageVersion(value v1.RemoteWriteMessageVersion) *RemoteWriteSpecApplyConfiguration {
	b.MessageVersion = &value
	return b
}

// WithSendExemplars sets the SendExemplars field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SendExemplars field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithSendExemplars(value bool) *RemoteWriteSpecApplyConfiguration {
	b.SendExemplars = &value
	return b
}

// WithSendNativeHistograms sets the SendNativeHistograms field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SendNativeHistograms field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithSendNativeHistograms(value bool) *RemoteWriteSpecApplyConfiguration {
	b.SendNativeHistograms = &value
	return b
}

// WithRemoteTimeout sets the RemoteTimeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RemoteTimeout field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithRemoteTimeout(value v1.Duration) *RemoteWriteSpecApplyConfiguration {
	b.RemoteTimeout = &value
	return b
}

// WithHeaders puts the entries into the Headers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Headers field,
// overwriting an existing map entries in Headers field with the same key.
func (b *RemoteWriteSpecApplyConfiguration) WithHeaders(entries map[string]string) *RemoteWriteSpecApplyConfiguration {
	if b.Headers == nil && len(entries) > 0 {
		b.Headers = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Headers[k] = v
	}
	return b
}

// WithWriteRelabelConfigs adds the given value to the WriteRelabelConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the WriteRelabelConfigs field.
func (b *RemoteWriteSpecApplyConfiguration) WithWriteRelabelConfigs(values ...*RelabelConfigApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithWriteRelabelConfigs")
		}
		b.WriteRelabelConfigs = append(b.WriteRelabelConfigs, *values[i])
	}
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithOAuth2(value *OAuth2ApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithBasicAuth(value *BasicAuthApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithBearerTokenFile sets the BearerTokenFile field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BearerTokenFile field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithBearerTokenFile(value string) *RemoteWriteSpecApplyConfiguration {
	b.BearerTokenFile = &value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithAuthorization(value *AuthorizationApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.Authorization = value
	return b
}

// WithSigv4 sets the Sigv4 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Sigv4 field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithSigv4(value *Sigv4ApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.Sigv4 = value
	return b
}

// WithAzureAD sets the AzureAD field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AzureAD field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithAzureAD(value *AzureADApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.AzureAD = value
	return b
}

// WithBearerToken sets the BearerToken field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BearerToken field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithBearerToken(value string) *RemoteWriteSpecApplyConfiguration {
	b.BearerToken = &value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithTLSConfig(value *TLSConfigApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithProxyURL(value string) *RemoteWriteSpecApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithNoProxy(value string) *RemoteWriteSpecApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithProxyFromEnvironment(value bool) *RemoteWriteSpecApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *RemoteWriteSpecApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *RemoteWriteSpecApplyConfiguration {
	if b.ProxyConnectHeader == nil && len(entries) > 0 {
		b.ProxyConnectHeader = make(map[string][]corev1.SecretKeySelector, len(entries))
	}
	for k, v := range entries {
		b.ProxyConnectHeader[k] = v
	}
	return b
}

// WithHTTPHeaders adds the given value to the HTTPHeaders field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the HTTPHeaders field.
func (b *RemoteWriteSpecApplyConfiguration) WithHTTPHeaders(values ...*HTTPHeaderApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithHTTPHeaders")
		}
		b.HTTPHeaders = append(b.HTTPHeaders, *values[i])
	}
	return b
}

// WithFollowRedirects sets the FollowRedirects field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FollowRedirects field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithFollowRedirects(value bool) *RemoteWriteSpecApplyConfiguration {
	b.FollowRedirects = &value
	return b
}

// WithQueueConfig sets the QueueConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the QueueConfig field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithQueueConfig(value *QueueConfigApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.QueueConfig = value
	return b
}

// WithMetadataConfig sets the MetadataConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MetadataConfig field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithMetadataConfig(value *MetadataConfigApplyConfiguration) *RemoteWriteSpecApplyConfiguration {
	b.MetadataConfig = value
	return b
}

// WithEnableHttp2 sets the EnableHttp2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHttp2 field is set to the value of the last call.
func (b *RemoteWriteSpecApplyConfiguration) WithEnableHttp2(value bool) *RemoteWriteSpecApplyConfiguration {
	b.EnableHttp2 = &value
	return b
}
