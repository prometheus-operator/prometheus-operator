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
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

// ProbeSpecApplyConfiguration represents a declarative configuration of the ProbeSpec type for use
// with apply.
type ProbeSpecApplyConfiguration struct {
	JobName                                 *string                              `json:"jobName,omitempty"`
	ProberSpec                              *ProberSpecApplyConfiguration        `json:"prober,omitempty"`
	Module                                  *string                              `json:"module,omitempty"`
	Targets                                 *ProbeTargetsApplyConfiguration      `json:"targets,omitempty"`
	Interval                                *monitoringv1.Duration               `json:"interval,omitempty"`
	ScrapeTimeout                           *monitoringv1.Duration               `json:"scrapeTimeout,omitempty"`
	TLSConfig                               *SafeTLSConfigApplyConfiguration     `json:"tlsConfig,omitempty"`
	BearerTokenSecret                       *corev1.SecretKeySelector            `json:"bearerTokenSecret,omitempty"`
	BasicAuth                               *BasicAuthApplyConfiguration         `json:"basicAuth,omitempty"`
	OAuth2                                  *OAuth2ApplyConfiguration            `json:"oauth2,omitempty"`
	MetricRelabelConfigs                    []RelabelConfigApplyConfiguration    `json:"metricRelabelings,omitempty"`
	Authorization                           *SafeAuthorizationApplyConfiguration `json:"authorization,omitempty"`
	SampleLimit                             *uint64                              `json:"sampleLimit,omitempty"`
	TargetLimit                             *uint64                              `json:"targetLimit,omitempty"`
	ScrapeProtocols                         []monitoringv1.ScrapeProtocol        `json:"scrapeProtocols,omitempty"`
	LabelLimit                              *uint64                              `json:"labelLimit,omitempty"`
	LabelNameLengthLimit                    *uint64                              `json:"labelNameLengthLimit,omitempty"`
	LabelValueLengthLimit                   *uint64                              `json:"labelValueLengthLimit,omitempty"`
	NativeHistogramConfigApplyConfiguration `json:",inline"`
	KeepDroppedTargets                      *uint64 `json:"keepDroppedTargets,omitempty"`
	ScrapeClassName                         *string `json:"scrapeClass,omitempty"`
}

// ProbeSpecApplyConfiguration constructs a declarative configuration of the ProbeSpec type for use with
// apply.
func ProbeSpec() *ProbeSpecApplyConfiguration {
	return &ProbeSpecApplyConfiguration{}
}

// WithJobName sets the JobName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the JobName field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithJobName(value string) *ProbeSpecApplyConfiguration {
	b.JobName = &value
	return b
}

// WithProberSpec sets the ProberSpec field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProberSpec field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithProberSpec(value *ProberSpecApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.ProberSpec = value
	return b
}

// WithModule sets the Module field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Module field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithModule(value string) *ProbeSpecApplyConfiguration {
	b.Module = &value
	return b
}

// WithTargets sets the Targets field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Targets field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithTargets(value *ProbeTargetsApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.Targets = value
	return b
}

// WithInterval sets the Interval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Interval field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithInterval(value monitoringv1.Duration) *ProbeSpecApplyConfiguration {
	b.Interval = &value
	return b
}

// WithScrapeTimeout sets the ScrapeTimeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeTimeout field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithScrapeTimeout(value monitoringv1.Duration) *ProbeSpecApplyConfiguration {
	b.ScrapeTimeout = &value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithTLSConfig(value *SafeTLSConfigApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithBearerTokenSecret sets the BearerTokenSecret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BearerTokenSecret field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithBearerTokenSecret(value corev1.SecretKeySelector) *ProbeSpecApplyConfiguration {
	b.BearerTokenSecret = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithBasicAuth(value *BasicAuthApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithOAuth2(value *OAuth2ApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithMetricRelabelConfigs adds the given value to the MetricRelabelConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the MetricRelabelConfigs field.
func (b *ProbeSpecApplyConfiguration) WithMetricRelabelConfigs(values ...*RelabelConfigApplyConfiguration) *ProbeSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithMetricRelabelConfigs")
		}
		b.MetricRelabelConfigs = append(b.MetricRelabelConfigs, *values[i])
	}
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithAuthorization(value *SafeAuthorizationApplyConfiguration) *ProbeSpecApplyConfiguration {
	b.Authorization = value
	return b
}

// WithSampleLimit sets the SampleLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SampleLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithSampleLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.SampleLimit = &value
	return b
}

// WithTargetLimit sets the TargetLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TargetLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithTargetLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.TargetLimit = &value
	return b
}

// WithScrapeProtocols adds the given value to the ScrapeProtocols field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ScrapeProtocols field.
func (b *ProbeSpecApplyConfiguration) WithScrapeProtocols(values ...monitoringv1.ScrapeProtocol) *ProbeSpecApplyConfiguration {
	for i := range values {
		b.ScrapeProtocols = append(b.ScrapeProtocols, values[i])
	}
	return b
}

// WithLabelLimit sets the LabelLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithLabelLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.LabelLimit = &value
	return b
}

// WithLabelNameLengthLimit sets the LabelNameLengthLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelNameLengthLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithLabelNameLengthLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.LabelNameLengthLimit = &value
	return b
}

// WithLabelValueLengthLimit sets the LabelValueLengthLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelValueLengthLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithLabelValueLengthLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.LabelValueLengthLimit = &value
	return b
}

// WithScrapeClassicHistograms sets the ScrapeClassicHistograms field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeClassicHistograms field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithScrapeClassicHistograms(value bool) *ProbeSpecApplyConfiguration {
	b.ScrapeClassicHistograms = &value
	return b
}

// WithNativeHistogramBucketLimit sets the NativeHistogramBucketLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NativeHistogramBucketLimit field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithNativeHistogramBucketLimit(value uint64) *ProbeSpecApplyConfiguration {
	b.NativeHistogramBucketLimit = &value
	return b
}

// WithNativeHistogramMinBucketFactor sets the NativeHistogramMinBucketFactor field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NativeHistogramMinBucketFactor field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithNativeHistogramMinBucketFactor(value float64) *ProbeSpecApplyConfiguration {
	b.NativeHistogramMinBucketFactor = &value
	return b
}

// WithKeepDroppedTargets sets the KeepDroppedTargets field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the KeepDroppedTargets field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithKeepDroppedTargets(value uint64) *ProbeSpecApplyConfiguration {
	b.KeepDroppedTargets = &value
	return b
}

// WithScrapeClassName sets the ScrapeClassName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeClassName field is set to the value of the last call.
func (b *ProbeSpecApplyConfiguration) WithScrapeClassName(value string) *ProbeSpecApplyConfiguration {
	b.ScrapeClassName = &value
	return b
}
