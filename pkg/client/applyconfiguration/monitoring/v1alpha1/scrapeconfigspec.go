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
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// ScrapeConfigSpecApplyConfiguration represents a declarative configuration of the ScrapeConfigSpec type for use
// with apply.
type ScrapeConfigSpecApplyConfiguration struct {
	JobName                                    *string                                  `json:"jobName,omitempty"`
	StaticConfigs                              []StaticConfigApplyConfiguration         `json:"staticConfigs,omitempty"`
	FileSDConfigs                              []FileSDConfigApplyConfiguration         `json:"fileSDConfigs,omitempty"`
	HTTPSDConfigs                              []HTTPSDConfigApplyConfiguration         `json:"httpSDConfigs,omitempty"`
	KubernetesSDConfigs                        []KubernetesSDConfigApplyConfiguration   `json:"kubernetesSDConfigs,omitempty"`
	ConsulSDConfigs                            []ConsulSDConfigApplyConfiguration       `json:"consulSDConfigs,omitempty"`
	DNSSDConfigs                               []DNSSDConfigApplyConfiguration          `json:"dnsSDConfigs,omitempty"`
	EC2SDConfigs                               []EC2SDConfigApplyConfiguration          `json:"ec2SDConfigs,omitempty"`
	AzureSDConfigs                             []AzureSDConfigApplyConfiguration        `json:"azureSDConfigs,omitempty"`
	GCESDConfigs                               []GCESDConfigApplyConfiguration          `json:"gceSDConfigs,omitempty"`
	OpenStackSDConfigs                         []OpenStackSDConfigApplyConfiguration    `json:"openstackSDConfigs,omitempty"`
	DigitalOceanSDConfigs                      []DigitalOceanSDConfigApplyConfiguration `json:"digitalOceanSDConfigs,omitempty"`
	KumaSDConfigs                              []KumaSDConfigApplyConfiguration         `json:"kumaSDConfigs,omitempty"`
	EurekaSDConfigs                            []EurekaSDConfigApplyConfiguration       `json:"eurekaSDConfigs,omitempty"`
	DockerSDConfigs                            []DockerSDConfigApplyConfiguration       `json:"dockerSDConfigs,omitempty"`
	LinodeSDConfigs                            []LinodeSDConfigApplyConfiguration       `json:"linodeSDConfigs,omitempty"`
	HetznerSDConfigs                           []HetznerSDConfigApplyConfiguration      `json:"hetznerSDConfigs,omitempty"`
	NomadSDConfigs                             []NomadSDConfigApplyConfiguration        `json:"nomadSDConfigs,omitempty"`
	DockerSwarmSDConfigs                       []DockerSwarmSDConfigApplyConfiguration  `json:"dockerSwarmSDConfigs,omitempty"`
	PuppetDBSDConfigs                          []PuppetDBSDConfigApplyConfiguration     `json:"puppetDBSDConfigs,omitempty"`
	LightSailSDConfigs                         []LightSailSDConfigApplyConfiguration    `json:"lightSailSDConfigs,omitempty"`
	OVHCloudSDConfigs                          []OVHCloudSDConfigApplyConfiguration     `json:"ovhcloudSDConfigs,omitempty"`
	ScalewaySDConfigs                          []ScalewaySDConfigApplyConfiguration     `json:"scalewaySDConfigs,omitempty"`
	IonosSDConfigs                             []IonosSDConfigApplyConfiguration        `json:"ionosSDConfigs,omitempty"`
	RelabelConfigs                             []v1.RelabelConfigApplyConfiguration     `json:"relabelings,omitempty"`
	MetricsPath                                *string                                  `json:"metricsPath,omitempty"`
	ScrapeInterval                             *monitoringv1.Duration                   `json:"scrapeInterval,omitempty"`
	ScrapeTimeout                              *monitoringv1.Duration                   `json:"scrapeTimeout,omitempty"`
	ScrapeProtocols                            []monitoringv1.ScrapeProtocol            `json:"scrapeProtocols,omitempty"`
	HonorTimestamps                            *bool                                    `json:"honorTimestamps,omitempty"`
	TrackTimestampsStaleness                   *bool                                    `json:"trackTimestampsStaleness,omitempty"`
	HonorLabels                                *bool                                    `json:"honorLabels,omitempty"`
	Params                                     map[string][]string                      `json:"params,omitempty"`
	Scheme                                     *string                                  `json:"scheme,omitempty"`
	EnableCompression                          *bool                                    `json:"enableCompression,omitempty"`
	EnableHTTP2                                *bool                                    `json:"enableHTTP2,omitempty"`
	BasicAuth                                  *v1.BasicAuthApplyConfiguration          `json:"basicAuth,omitempty"`
	Authorization                              *v1.SafeAuthorizationApplyConfiguration  `json:"authorization,omitempty"`
	OAuth2                                     *v1.OAuth2ApplyConfiguration             `json:"oauth2,omitempty"`
	TLSConfig                                  *v1.SafeTLSConfigApplyConfiguration      `json:"tlsConfig,omitempty"`
	SampleLimit                                *uint64                                  `json:"sampleLimit,omitempty"`
	TargetLimit                                *uint64                                  `json:"targetLimit,omitempty"`
	LabelLimit                                 *uint64                                  `json:"labelLimit,omitempty"`
	LabelNameLengthLimit                       *uint64                                  `json:"labelNameLengthLimit,omitempty"`
	LabelValueLengthLimit                      *uint64                                  `json:"labelValueLengthLimit,omitempty"`
	v1.NativeHistogramConfigApplyConfiguration `json:",inline"`
	KeepDroppedTargets                         *uint64                              `json:"keepDroppedTargets,omitempty"`
	MetricRelabelConfigs                       []v1.RelabelConfigApplyConfiguration `json:"metricRelabelings,omitempty"`
	v1.ProxyConfigApplyConfiguration           `json:",inline"`
	v1.CustomHTTPConfigApplyConfiguration      `json:",inline"`
	ScrapeClassName                            *string `json:"scrapeClass,omitempty"`
}

// ScrapeConfigSpecApplyConfiguration constructs a declarative configuration of the ScrapeConfigSpec type for use with
// apply.
func ScrapeConfigSpec() *ScrapeConfigSpecApplyConfiguration {
	return &ScrapeConfigSpecApplyConfiguration{}
}

// WithJobName sets the JobName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the JobName field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithJobName(value string) *ScrapeConfigSpecApplyConfiguration {
	b.JobName = &value
	return b
}

// WithStaticConfigs adds the given value to the StaticConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the StaticConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithStaticConfigs(values ...*StaticConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithStaticConfigs")
		}
		b.StaticConfigs = append(b.StaticConfigs, *values[i])
	}
	return b
}

// WithFileSDConfigs adds the given value to the FileSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the FileSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithFileSDConfigs(values ...*FileSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithFileSDConfigs")
		}
		b.FileSDConfigs = append(b.FileSDConfigs, *values[i])
	}
	return b
}

// WithHTTPSDConfigs adds the given value to the HTTPSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the HTTPSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithHTTPSDConfigs(values ...*HTTPSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithHTTPSDConfigs")
		}
		b.HTTPSDConfigs = append(b.HTTPSDConfigs, *values[i])
	}
	return b
}

// WithKubernetesSDConfigs adds the given value to the KubernetesSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the KubernetesSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithKubernetesSDConfigs(values ...*KubernetesSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithKubernetesSDConfigs")
		}
		b.KubernetesSDConfigs = append(b.KubernetesSDConfigs, *values[i])
	}
	return b
}

// WithConsulSDConfigs adds the given value to the ConsulSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ConsulSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithConsulSDConfigs(values ...*ConsulSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithConsulSDConfigs")
		}
		b.ConsulSDConfigs = append(b.ConsulSDConfigs, *values[i])
	}
	return b
}

// WithDNSSDConfigs adds the given value to the DNSSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DNSSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithDNSSDConfigs(values ...*DNSSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDNSSDConfigs")
		}
		b.DNSSDConfigs = append(b.DNSSDConfigs, *values[i])
	}
	return b
}

// WithEC2SDConfigs adds the given value to the EC2SDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the EC2SDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithEC2SDConfigs(values ...*EC2SDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithEC2SDConfigs")
		}
		b.EC2SDConfigs = append(b.EC2SDConfigs, *values[i])
	}
	return b
}

// WithAzureSDConfigs adds the given value to the AzureSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the AzureSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithAzureSDConfigs(values ...*AzureSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithAzureSDConfigs")
		}
		b.AzureSDConfigs = append(b.AzureSDConfigs, *values[i])
	}
	return b
}

// WithGCESDConfigs adds the given value to the GCESDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the GCESDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithGCESDConfigs(values ...*GCESDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithGCESDConfigs")
		}
		b.GCESDConfigs = append(b.GCESDConfigs, *values[i])
	}
	return b
}

// WithOpenStackSDConfigs adds the given value to the OpenStackSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the OpenStackSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithOpenStackSDConfigs(values ...*OpenStackSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithOpenStackSDConfigs")
		}
		b.OpenStackSDConfigs = append(b.OpenStackSDConfigs, *values[i])
	}
	return b
}

// WithDigitalOceanSDConfigs adds the given value to the DigitalOceanSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DigitalOceanSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithDigitalOceanSDConfigs(values ...*DigitalOceanSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDigitalOceanSDConfigs")
		}
		b.DigitalOceanSDConfigs = append(b.DigitalOceanSDConfigs, *values[i])
	}
	return b
}

// WithKumaSDConfigs adds the given value to the KumaSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the KumaSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithKumaSDConfigs(values ...*KumaSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithKumaSDConfigs")
		}
		b.KumaSDConfigs = append(b.KumaSDConfigs, *values[i])
	}
	return b
}

// WithEurekaSDConfigs adds the given value to the EurekaSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the EurekaSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithEurekaSDConfigs(values ...*EurekaSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithEurekaSDConfigs")
		}
		b.EurekaSDConfigs = append(b.EurekaSDConfigs, *values[i])
	}
	return b
}

// WithDockerSDConfigs adds the given value to the DockerSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DockerSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithDockerSDConfigs(values ...*DockerSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDockerSDConfigs")
		}
		b.DockerSDConfigs = append(b.DockerSDConfigs, *values[i])
	}
	return b
}

// WithLinodeSDConfigs adds the given value to the LinodeSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the LinodeSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithLinodeSDConfigs(values ...*LinodeSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithLinodeSDConfigs")
		}
		b.LinodeSDConfigs = append(b.LinodeSDConfigs, *values[i])
	}
	return b
}

// WithHetznerSDConfigs adds the given value to the HetznerSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the HetznerSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithHetznerSDConfigs(values ...*HetznerSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithHetznerSDConfigs")
		}
		b.HetznerSDConfigs = append(b.HetznerSDConfigs, *values[i])
	}
	return b
}

// WithNomadSDConfigs adds the given value to the NomadSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the NomadSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithNomadSDConfigs(values ...*NomadSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithNomadSDConfigs")
		}
		b.NomadSDConfigs = append(b.NomadSDConfigs, *values[i])
	}
	return b
}

// WithDockerSwarmSDConfigs adds the given value to the DockerSwarmSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DockerSwarmSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithDockerSwarmSDConfigs(values ...*DockerSwarmSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDockerSwarmSDConfigs")
		}
		b.DockerSwarmSDConfigs = append(b.DockerSwarmSDConfigs, *values[i])
	}
	return b
}

// WithPuppetDBSDConfigs adds the given value to the PuppetDBSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the PuppetDBSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithPuppetDBSDConfigs(values ...*PuppetDBSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithPuppetDBSDConfigs")
		}
		b.PuppetDBSDConfigs = append(b.PuppetDBSDConfigs, *values[i])
	}
	return b
}

// WithLightSailSDConfigs adds the given value to the LightSailSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the LightSailSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithLightSailSDConfigs(values ...*LightSailSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithLightSailSDConfigs")
		}
		b.LightSailSDConfigs = append(b.LightSailSDConfigs, *values[i])
	}
	return b
}

// WithOVHCloudSDConfigs adds the given value to the OVHCloudSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the OVHCloudSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithOVHCloudSDConfigs(values ...*OVHCloudSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithOVHCloudSDConfigs")
		}
		b.OVHCloudSDConfigs = append(b.OVHCloudSDConfigs, *values[i])
	}
	return b
}

// WithScalewaySDConfigs adds the given value to the ScalewaySDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ScalewaySDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithScalewaySDConfigs(values ...*ScalewaySDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithScalewaySDConfigs")
		}
		b.ScalewaySDConfigs = append(b.ScalewaySDConfigs, *values[i])
	}
	return b
}

// WithIonosSDConfigs adds the given value to the IonosSDConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the IonosSDConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithIonosSDConfigs(values ...*IonosSDConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithIonosSDConfigs")
		}
		b.IonosSDConfigs = append(b.IonosSDConfigs, *values[i])
	}
	return b
}

// WithRelabelConfigs adds the given value to the RelabelConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the RelabelConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithRelabelConfigs(values ...*v1.RelabelConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithRelabelConfigs")
		}
		b.RelabelConfigs = append(b.RelabelConfigs, *values[i])
	}
	return b
}

// WithMetricsPath sets the MetricsPath field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MetricsPath field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithMetricsPath(value string) *ScrapeConfigSpecApplyConfiguration {
	b.MetricsPath = &value
	return b
}

// WithScrapeInterval sets the ScrapeInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeInterval field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithScrapeInterval(value monitoringv1.Duration) *ScrapeConfigSpecApplyConfiguration {
	b.ScrapeInterval = &value
	return b
}

// WithScrapeTimeout sets the ScrapeTimeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeTimeout field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithScrapeTimeout(value monitoringv1.Duration) *ScrapeConfigSpecApplyConfiguration {
	b.ScrapeTimeout = &value
	return b
}

// WithScrapeProtocols adds the given value to the ScrapeProtocols field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ScrapeProtocols field.
func (b *ScrapeConfigSpecApplyConfiguration) WithScrapeProtocols(values ...monitoringv1.ScrapeProtocol) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		b.ScrapeProtocols = append(b.ScrapeProtocols, values[i])
	}
	return b
}

// WithHonorTimestamps sets the HonorTimestamps field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HonorTimestamps field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithHonorTimestamps(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.HonorTimestamps = &value
	return b
}

// WithTrackTimestampsStaleness sets the TrackTimestampsStaleness field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TrackTimestampsStaleness field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithTrackTimestampsStaleness(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.TrackTimestampsStaleness = &value
	return b
}

// WithHonorLabels sets the HonorLabels field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HonorLabels field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithHonorLabels(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.HonorLabels = &value
	return b
}

// WithParams puts the entries into the Params field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Params field,
// overwriting an existing map entries in Params field with the same key.
func (b *ScrapeConfigSpecApplyConfiguration) WithParams(entries map[string][]string) *ScrapeConfigSpecApplyConfiguration {
	if b.Params == nil && len(entries) > 0 {
		b.Params = make(map[string][]string, len(entries))
	}
	for k, v := range entries {
		b.Params[k] = v
	}
	return b
}

// WithScheme sets the Scheme field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Scheme field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithScheme(value string) *ScrapeConfigSpecApplyConfiguration {
	b.Scheme = &value
	return b
}

// WithEnableCompression sets the EnableCompression field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableCompression field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithEnableCompression(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.EnableCompression = &value
	return b
}

// WithEnableHTTP2 sets the EnableHTTP2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableHTTP2 field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithEnableHTTP2(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.EnableHTTP2 = &value
	return b
}

// WithBasicAuth sets the BasicAuth field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BasicAuth field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithBasicAuth(value *v1.BasicAuthApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	b.BasicAuth = value
	return b
}

// WithAuthorization sets the Authorization field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Authorization field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithAuthorization(value *v1.SafeAuthorizationApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	b.Authorization = value
	return b
}

// WithOAuth2 sets the OAuth2 field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OAuth2 field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithOAuth2(value *v1.OAuth2ApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	b.OAuth2 = value
	return b
}

// WithTLSConfig sets the TLSConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSConfig field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithTLSConfig(value *v1.SafeTLSConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	b.TLSConfig = value
	return b
}

// WithSampleLimit sets the SampleLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SampleLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithSampleLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.SampleLimit = &value
	return b
}

// WithTargetLimit sets the TargetLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TargetLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithTargetLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.TargetLimit = &value
	return b
}

// WithLabelLimit sets the LabelLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithLabelLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.LabelLimit = &value
	return b
}

// WithLabelNameLengthLimit sets the LabelNameLengthLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelNameLengthLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithLabelNameLengthLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.LabelNameLengthLimit = &value
	return b
}

// WithLabelValueLengthLimit sets the LabelValueLengthLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LabelValueLengthLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithLabelValueLengthLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.LabelValueLengthLimit = &value
	return b
}

// WithScrapeClassicHistograms sets the ScrapeClassicHistograms field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeClassicHistograms field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithScrapeClassicHistograms(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.ScrapeClassicHistograms = &value
	return b
}

// WithNativeHistogramBucketLimit sets the NativeHistogramBucketLimit field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NativeHistogramBucketLimit field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithNativeHistogramBucketLimit(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.NativeHistogramBucketLimit = &value
	return b
}

// WithNativeHistogramMinBucketFactor sets the NativeHistogramMinBucketFactor field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NativeHistogramMinBucketFactor field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithNativeHistogramMinBucketFactor(value resource.Quantity) *ScrapeConfigSpecApplyConfiguration {
	b.NativeHistogramMinBucketFactor = &value
	return b
}

// WithKeepDroppedTargets sets the KeepDroppedTargets field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the KeepDroppedTargets field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithKeepDroppedTargets(value uint64) *ScrapeConfigSpecApplyConfiguration {
	b.KeepDroppedTargets = &value
	return b
}

// WithMetricRelabelConfigs adds the given value to the MetricRelabelConfigs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the MetricRelabelConfigs field.
func (b *ScrapeConfigSpecApplyConfiguration) WithMetricRelabelConfigs(values ...*v1.RelabelConfigApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithMetricRelabelConfigs")
		}
		b.MetricRelabelConfigs = append(b.MetricRelabelConfigs, *values[i])
	}
	return b
}

// WithProxyURL sets the ProxyURL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyURL field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithProxyURL(value string) *ScrapeConfigSpecApplyConfiguration {
	b.ProxyURL = &value
	return b
}

// WithNoProxy sets the NoProxy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NoProxy field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithNoProxy(value string) *ScrapeConfigSpecApplyConfiguration {
	b.NoProxy = &value
	return b
}

// WithProxyFromEnvironment sets the ProxyFromEnvironment field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ProxyFromEnvironment field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithProxyFromEnvironment(value bool) *ScrapeConfigSpecApplyConfiguration {
	b.ProxyFromEnvironment = &value
	return b
}

// WithProxyConnectHeader puts the entries into the ProxyConnectHeader field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the ProxyConnectHeader field,
// overwriting an existing map entries in ProxyConnectHeader field with the same key.
func (b *ScrapeConfigSpecApplyConfiguration) WithProxyConnectHeader(entries map[string][]corev1.SecretKeySelector) *ScrapeConfigSpecApplyConfiguration {
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
func (b *ScrapeConfigSpecApplyConfiguration) WithHTTPHeaders(entries map[string]v1.HTTPHeaderApplyConfiguration) *ScrapeConfigSpecApplyConfiguration {
	if b.HTTPHeaders == nil && len(entries) > 0 {
		b.HTTPHeaders = make(map[string]v1.HTTPHeaderApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.HTTPHeaders[k] = v
	}
	return b
}

// WithScrapeClassName sets the ScrapeClassName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScrapeClassName field is set to the value of the last call.
func (b *ScrapeConfigSpecApplyConfiguration) WithScrapeClassName(value string) *ScrapeConfigSpecApplyConfiguration {
	b.ScrapeClassName = &value
	return b
}
