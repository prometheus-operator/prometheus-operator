// Copyright 2018 The prometheus-operator Authors
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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PodMonitorsKind   = "PodMonitor"
	PodMonitorName    = "podmonitors"
	PodMonitorKindKey = "podmonitor"
)

// PodMonitor defines monitoring for a set of pods.
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator"
type PodMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Pod selection for target discovery by Prometheus.
	Spec PodMonitorSpec `json:"spec"`
}

// PodMonitorSpec contains specification parameters for a PodMonitor.
// +k8s:openapi-gen=true
type PodMonitorSpec struct {
	// The label to use to retrieve the job name from.
	JobLabel string `json:"jobLabel,omitempty"`
	// PodTargetLabels transfers labels on the Kubernetes Pod onto the target.
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
	// A list of endpoints allowed as part of this PodMonitor.
	PodMetricsEndpoints []PodMetricsEndpoint `json:"podMetricsEndpoints"`
	// Selector to select Pod objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the Endpoints objects are discovered from.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	SampleLimit uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	TargetLimit uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelLimit uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelNameLengthLimit uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelValueLengthLimit uint64 `json:"labelValueLengthLimit,omitempty"`
}

// PodMetricsEndpoint defines a scrapeable endpoint of a Kubernetes Pod serving Prometheus metrics.
// +k8s:openapi-gen=true
type PodMetricsEndpoint struct {
	// Name of the pod port this endpoint refers to. Mutually exclusive with targetPort.
	Port string `json:"port,omitempty"`
	// Deprecated: Use 'port' instead.
	TargetPort *intstr.IntOrString `json:"targetPort,omitempty"`
	// HTTP path to scrape for metrics.
	Path string `json:"path,omitempty"`
	// HTTP scheme to use for scraping.
	Scheme string `json:"scheme,omitempty"`
	// Optional HTTP URL parameters
	Params map[string][]string `json:"params,omitempty"`
	// Interval at which metrics should be scraped
	Interval string `json:"interval,omitempty"`
	// Timeout after which the scrape is ended
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`
	// TLS configuration to use when scraping the endpoint.
	TLSConfig *PodMetricsEndpointTLSConfig `json:"tlsConfig,omitempty"`
	// Secret to mount to read bearer token for scraping targets. The secret
	// needs to be in the same namespace as the pod monitor and accessible by
	// the Prometheus Operator.
	BearerTokenSecret v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	HonorLabels bool `json:"honorLabels,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// BasicAuth allow an endpoint to authenticate over basic authentication.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoint
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// Authorization section for this endpoint
	Authorization *SafeAuthorization `json:"authorization,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	MetricRelabelConfigs []*RelabelConfig `json:"metricRelabelings,omitempty"`
	// RelabelConfigs to apply to samples before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields
	// and replaces original scrape job name with __tmp_prometheus_job_name.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	RelabelConfigs []*RelabelConfig `json:"relabelings,omitempty"`
	// ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint.
	ProxyURL *string `json:"proxyUrl,omitempty"`
}

// PodMetricsEndpointTLSConfig specifies TLS configuration parameters.
// +k8s:openapi-gen=true
type PodMetricsEndpointTLSConfig struct {
	SafeTLSConfig `json:",inline"`
}

// PodMonitorList is a list of PodMonitors.
// +k8s:openapi-gen=true
type PodMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of PodMonitors
	Items []*PodMonitor `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PodMonitor) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PodMonitorList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
