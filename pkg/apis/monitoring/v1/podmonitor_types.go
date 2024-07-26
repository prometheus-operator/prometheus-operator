// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="pmon"

// The `PodMonitor` custom resource definition (CRD) defines how `Prometheus` and `PrometheusAgent` can scrape metrics from a group of pods.
// Among other things, it allows to specify:
// * The pods to scrape via label selectors.
// * The container ports to scrape.
// * Authentication credentials to use.
// * Target and metric relabeling.
//
// `Prometheus` and `PrometheusAgent` objects select `PodMonitor` objects using label and namespace selectors.
type PodMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Pod selection for target discovery by Prometheus.
	Spec PodMonitorSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PodMonitor) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PodMonitorSpec contains specification parameters for a PodMonitor.
// +k8s:openapi-gen=true
type PodMonitorSpec struct {
	// The label to use to retrieve the job name from.
	// `jobLabel` selects the label from the associated Kubernetes `Pod`
	// object which will be used as the `job` label for all metrics.
	//
	// For example if `jobLabel` is set to `foo` and the Kubernetes `Pod`
	// object is labeled with `foo: bar`, then Prometheus adds the `job="bar"`
	// label to all ingested metrics.
	//
	// If the value of this field is empty, the `job` label of the metrics
	// defaults to the namespace and name of the PodMonitor object (e.g. `<namespace>/<name>`).
	JobLabel string `json:"jobLabel,omitempty"`

	// `podTargetLabels` defines the labels which are transferred from the
	// associated Kubernetes `Pod` object onto the ingested metrics.
	//
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`

	// Defines how to scrape metrics from the selected pods.
	//
	// +optional
	PodMetricsEndpoints []PodMetricsEndpoint `json:"podMetricsEndpoints"`

	// Label selector to select the Kubernetes `Pod` objects to scrape metrics from.
	Selector metav1.LabelSelector `json:"selector"`
	// `namespaceSelector` defines in which namespace(s) Prometheus should discover the pods.
	// By default, the pods are discovered in the same namespace as the `PodMonitor` object but it is possible to select pods across different/all namespaces.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`

	// `sampleLimit` defines a per-scrape limit on the number of scraped samples
	// that will be accepted.
	//
	// +optional
	SampleLimit *uint64 `json:"sampleLimit,omitempty"`

	// `targetLimit` defines a limit on the number of scraped targets that will
	// be accepted.
	//
	// +optional
	TargetLimit *uint64 `json:"targetLimit,omitempty"`

	// `scrapeProtocols` defines the protocols to negotiate during a scrape. It tells clients the
	// protocols supported by Prometheus in order of preference (from most to least preferred).
	//
	// If unset, Prometheus uses its default value.
	//
	// It requires Prometheus >= v2.49.0.
	//
	// +listType=set
	// +optional
	ScrapeProtocols []ScrapeProtocol `json:"scrapeProtocols,omitempty"`

	// Per-scrape limit on number of labels that will be accepted for a sample.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	LabelLimit *uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	LabelNameLengthLimit *uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	LabelValueLengthLimit *uint64 `json:"labelValueLengthLimit,omitempty"`
	// Per-scrape limit on the number of targets dropped by relabeling
	// that will be kept in memory. 0 means no limit.
	//
	// It requires Prometheus >= v2.47.0.
	//
	// +optional
	KeepDroppedTargets *uint64 `json:"keepDroppedTargets,omitempty"`

	// `attachMetadata` defines additional metadata which is added to the
	// discovered targets.
	//
	// It requires Prometheus >= v2.35.0.
	//
	// +optional
	AttachMetadata *AttachMetadata `json:"attachMetadata,omitempty"`

	// The scrape class to apply.
	// +optional
	// +kubebuilder:validation:MinLength=1
	ScrapeClassName *string `json:"scrapeClass,omitempty"`

	// When defined, bodySizeLimit specifies a job level limit on the size
	// of uncompressed response body that will be accepted by Prometheus.
	//
	// It requires Prometheus >= v2.28.0.
	//
	// +optional
	BodySizeLimit *ByteSize `json:"bodySizeLimit,omitempty"`
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
func (l *PodMonitorList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PodMetricsEndpoint defines an endpoint serving Prometheus metrics to be scraped by
// Prometheus.
//
// +k8s:openapi-gen=true
type PodMetricsEndpoint struct {
	// Name of the Pod port which this endpoint refers to.
	//
	// It takes precedence over `targetPort`.
	Port string `json:"port,omitempty"`

	// Name or number of the target port of the `Pod` object behind the Service, the
	// port must be specified with container port property.
	//
	// Deprecated: use 'port' instead.
	TargetPort *intstr.IntOrString `json:"targetPort,omitempty"`

	// HTTP path from which to scrape for metrics.
	//
	// If empty, Prometheus uses the default value (e.g. `/metrics`).
	Path string `json:"path,omitempty"`

	// HTTP scheme to use for scraping.
	//
	// `http` and `https` are the expected values unless you rewrite the
	// `__scheme__` label via relabeling.
	//
	// If empty, Prometheus uses the default value `http`.
	//
	// +kubebuilder:validation:Enum=http;https
	Scheme string `json:"scheme,omitempty"`

	// `params` define optional HTTP URL parameters.
	Params map[string][]string `json:"params,omitempty"`

	// Interval at which Prometheus scrapes the metrics from the target.
	//
	// If empty, Prometheus uses the global scrape interval.
	Interval Duration `json:"interval,omitempty"`

	// Timeout after which Prometheus considers the scrape to be failed.
	//
	// If empty, Prometheus uses the global scrape timeout unless it is less
	// than the target's scrape interval value in which the latter is used.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`

	// TLS configuration to use when scraping the target.
	//
	// +optional
	TLSConfig *SafeTLSConfig `json:"tlsConfig,omitempty"`

	// `bearerTokenSecret` specifies a key of a Secret containing the bearer
	// token for scraping targets. The secret needs to be in the same namespace
	// as the PodMonitor object and readable by the Prometheus Operator.
	//
	// +optional
	//
	// Deprecated: use `authorization` instead.
	BearerTokenSecret v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`

	// When true, `honorLabels` preserves the metric's labels when they collide
	// with the target's labels.
	HonorLabels bool `json:"honorLabels,omitempty"`

	// `honorTimestamps` controls whether Prometheus preserves the timestamps
	// when exposed by the target.
	//
	// +optional
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`

	// `trackTimestampsStaleness` defines whether Prometheus tracks staleness of
	// the metrics that have an explicit timestamp present in scraped data.
	// Has no effect if `honorTimestamps` is false.
	//
	// It requires Prometheus >= v2.48.0.
	//
	// +optional
	TrackTimestampsStaleness *bool `json:"trackTimestampsStaleness,omitempty"`

	// `basicAuth` configures the Basic Authentication credentials to use when
	// scraping the target.
	//
	// Cannot be set at the same time as `authorization`, or `oauth2`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// `oauth2` configures the OAuth2 settings to use when scraping the target.
	//
	// It requires Prometheus >= 2.27.0.
	//
	// Cannot be set at the same time as `authorization`, or `basicAuth`.
	//
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`

	// `authorization` configures the Authorization header credentials to use when
	// scraping the target.
	//
	// Cannot be set at the same time as `basicAuth`, or `oauth2`.
	//
	// +optional
	Authorization *SafeAuthorization `json:"authorization,omitempty"`

	// `metricRelabelings` configures the relabeling rules to apply to the
	// samples before ingestion.
	//
	// +optional
	MetricRelabelConfigs []RelabelConfig `json:"metricRelabelings,omitempty"`

	// `relabelings` configures the relabeling rules to apply the target's
	// metadata labels.
	//
	// The Operator automatically adds relabelings for a few standard Kubernetes fields.
	//
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	//
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	//
	// +optional
	RelabelConfigs []RelabelConfig `json:"relabelings,omitempty"`

	// `proxyURL` configures the HTTP Proxy URL (e.g.
	// "http://proxyserver:2195") to go through when scraping the target.
	//
	// +optional
	ProxyURL *string `json:"proxyUrl,omitempty"`

	// `followRedirects` defines whether the scrape requests should follow HTTP
	// 3xx redirects.
	//
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`

	// `enableHttp2` can be used to disable HTTP2 when scraping the target.
	//
	// +optional
	EnableHttp2 *bool `json:"enableHttp2,omitempty"`

	// When true, the pods which are not running (e.g. either in Failed or
	// Succeeded state) are dropped during the target discovery.
	//
	// If unset, the filtering is enabled.
	//
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
	//
	// +optional
	FilterRunning *bool `json:"filterRunning,omitempty"`
}
