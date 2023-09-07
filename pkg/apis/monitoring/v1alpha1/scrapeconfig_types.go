// Copyright 2023 The prometheus-operator Authors
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

package v1alpha1

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ScrapeConfigsKind   = "ScrapeConfig"
	ScrapeConfigName    = "scrapeconfigs"
	ScrapeConfigKindKey = "scrapeconfig"
)

// Target represents a target for Prometheus to scrape
type Target string

// SDFile represents a file used for service discovery
// +kubebuilder:validation:Pattern=`^[^*]*(\*[^/]*)?\.(json|yml|yaml|JSON|YML|YAML)$`
type SDFile string

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="scfg"
// +kubebuilder:storageversion

// ScrapeConfig defines a namespaced Prometheus scrape_config to be aggregated across
// multiple namespaces into the Prometheus configuration.
type ScrapeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScrapeConfigSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ScrapeConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ScrapeConfigList is a list of ScrapeConfigs.
// +k8s:openapi-gen=true
type ScrapeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of ScrapeConfigs
	Items []*ScrapeConfig `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ScrapeConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ScrapeConfigSpec is a specification of the desired configuration for a scrape configuration.
// +k8s:openapi-gen=true
type ScrapeConfigSpec struct {
	// StaticConfigs defines a list of static targets with a common label set.
	// +optional
	StaticConfigs []StaticConfig `json:"staticConfigs,omitempty"`
	// FileSDConfigs defines a list of file service discovery configurations.
	// +optional
	FileSDConfigs []FileSDConfig `json:"fileSDConfigs,omitempty"`
	// HTTPSDConfigs defines a list of HTTP service discovery configurations.
	// +optional
	HTTPSDConfigs []HTTPSDConfig `json:"httpSDConfigs,omitempty"`
	// KubernetesSDConfigs defines a list of Kubernetes service discovery configurations.
	// +optional
	KubernetesSDConfigs []KubernetesSDConfig `json:"kubernetesSDConfigs,omitempty"`
	// ConsulSDConfigs defines a list of Consul service discovery configurations.
	// +optional
	ConsulSDConfigs []ConsulSDConfig `json:"consulSDConfigs,omitempty"`
	//DNSSDConfigs defines a list of DNS service discovery configurations.
	// +optional
	DNSSDConfigs []DNSSDConfig `json:"dnsSDConfigs,omitempty"`
	// RelabelConfigs defines how to rewrite the target's labels before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	// +optional
	RelabelConfigs []*v1.RelabelConfig `json:"relabelings,omitempty"`
	// MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics).
	// +optional
	MetricsPath *string `json:"metricsPath,omitempty"`
	// ScrapeInterval is the interval between consecutive scrapes.
	// +optional
	ScrapeInterval *v1.Duration `json:"scrapeInterval,omitempty"`
	// ScrapeTimeout is the number of seconds to wait until a scrape request times out.
	// +optional
	ScrapeTimeout *v1.Duration `json:"scrapeTimeout,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	// +optional
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	// +optional
	HonorLabels *bool `json:"honorLabels,omitempty"`
	// Optional HTTP URL parameters
	// +optional
	// +mapType:=atomic
	Params map[string][]string `json:"params,omitempty"`
	// Configures the protocol scheme used for requests.
	// If empty, Prometheus uses HTTP by default.
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	// +optional
	Scheme *string `json:"scheme,omitempty"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// TLS configuration to use on every scrape request
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	// +optional
	SampleLimit *uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	// +optional
	TargetLimit *uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelLimit *uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelNameLengthLimit *uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelValueLengthLimit *uint64 `json:"labelValueLengthLimit,omitempty"`
	// Per-scrape limit on the number of targets dropped by relabeling
	// that will be kept in memory. 0 means no limit.
	//
	// It requires Prometheus >= v2.47.0.
	//
	// +optional
	KeepDroppedTargets *uint64 `json:"keepDroppedTargets,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	// +optional
	MetricRelabelConfigs []*v1.RelabelConfig `json:"metricRelabelings,omitempty"`
}

// StaticConfig defines a Prometheus static configuration.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
// +k8s:openapi-gen=true
type StaticConfig struct {
	// List of targets for this static configuration.
	// +optional
	Targets []Target `json:"targets,omitempty"`
	// Labels assigned to all metrics scraped from the targets.
	// +mapType:=atomic
	// +optional
	Labels map[v1.LabelName]string `json:"labels,omitempty"`
}

// FileSDConfig defines a Prometheus file service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config
// +k8s:openapi-gen=true
type FileSDConfig struct {
	// List of files to be used for file discovery. Recommendation: use absolute paths. While relative paths work, the
	// prometheus-operator project makes no guarantees about the working directory where the configuration file is
	// stored.
	// Files must be mounted using Prometheus.ConfigMaps or Prometheus.Secrets.
	// +kubebuilder:validation:MinItems:=1
	Files []SDFile `json:"files"`
	// RefreshInterval configures the refresh interval at which Prometheus will reload the content of the files.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
}

// HTTPSDConfig defines a prometheus HTTP service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config
// +k8s:openapi-gen=true
type HTTPSDConfig struct {
	// URL from which the targets are fetched.
	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:Pattern:="^http(s)?://.+$"
	URL string `json:"url"`
	// RefreshInterval configures the refresh interval at which Prometheus will re-query the
	// endpoint to update the target list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// BasicAuth information to authenticate against the target HTTP endpoint.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration to authenticate against the target HTTP endpoint.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// KubernetesSDConfig allows retrieving scrape targets from Kubernetes' REST API.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
// +k8s:openapi-gen=true
type KubernetesSDConfig struct {
	// Role of the Kubernetes entities that should be discovered.
	// Currently the only supported role is "Node".
	// +kubebuilder:validation:Enum=Node
	// +required
	Role string `json:"role"`
}

// ConsulSDConfig defines a Consul service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config
// +k8s:openapi-gen=true
type ConsulSDConfig struct {
	// A valid string consisting of a hostname or IP followed by an optional port number.
	// +kubebuilder:validation:MinLength=1
	// +required
	Server string `json:"server"`
	// Consul ACL TokenRef, if not provided it will use the ACL from the local Consul Agent.
	// +optional
	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
	// Consul Datacenter name, if not provided it will use the local Consul Agent Datacenter.
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`
	// Namespaces are only supported in Consul Enterprise.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Admin Partitions are only supported in Consul Enterprise.
	// +optional
	Partition *string `json:"partition,omitempty"`
	// HTTP Scheme default "http"
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	// +optional
	Scheme *string `json:"scheme,omitempty"`
	// A list of services for which targets are retrieved. If omitted, all services are scraped.
	// +listType:=atomic
	// +optional
	Services []string `json:"services,omitempty"`
	// An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list.
	//+listType:=atomic
	// +optional
	Tags []string `json:"tags,omitempty"`
	// The string by which Consul tags are joined into the tag label.
	// If unset, Prometheus uses its default value.
	// +optional
	TagSeparator *string `json:"tag_separator,omitempty"`
	// Node metadata key/value pairs to filter nodes for a given service.
	// +mapType:=atomic
	// +optional
	NodeMeta map[string]string `json:"node_meta,omitempty"`
	// Allow stale Consul results (see https://www.consul.io/api/features/consistency.html). Will reduce load on Consul.
	// If unset, Prometheus uses its default value.
	// +optional
	AllowStale *bool `json:"allow_stale,omitempty"`
	// The time after which the provided names are refreshed.
	// On large setup it might be a good idea to increase this value because the catalog will change all the time.
	// If unset, Prometheus uses its default value.
	// +optional
	RefreshInterval *v1.Duration `json:"refresh_interval,omitempty"`
	// BasicAuth information to authenticate against the Consul Server.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration to authenticate against the Consul Server.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// +optional
	Oauth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// Optional proxy URL.
	// +optional
	ProxyUrl *string `json:"proxy_url,omitempty"`
	// Comma-separated string that can contain IPs, CIDR notation, domain names
	// that should be excluded from proxying. IP and domain names can
	// contain port numbers.
	// +optional
	NoProxy *string `json:"no_proxy,omitempty"`
	// Use proxy URL indicated by environment variables (HTTP_PROXY, https_proxy, HTTPs_PROXY, https_proxy, and no_proxy)
	// If unset, Prometheus uses its default value.
	// +optional
	ProxyFromEnvironment *bool `json:"proxy_from_environment,omitempty"`
	// Specifies headers to send to proxies during CONNECT requests.
	// +mapType:=atomic
	// +optional
	ProxyConnectHeader map[string]corev1.SecretKeySelector `json:"proxy_connect_header,omitempty"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// If unset, Prometheus uses its default value.
	// +optional
	FollowRedirects *bool `json:"follow_redirects,omitempty"`
	// Whether to enable HTTP2.
	// If unset, Prometheus uses its default value.
	// +optional
	EnableHttp2 *bool `json:"enable_http2,omitempty"`
	// TLS Config
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// DNSSDConfig allows specifying a set of DNS domain names which are periodically queried to discover a list of targets.
// The DNS servers to be contacted are read from /etc/resolv.conf.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dns_sd_config
// +k8s:openapi-gen=true
type DNSSDConfig struct {
	// A list of DNS domain names to be queried.
	// +kubebuilder:validation:MinItems:=1
	Names []string `json:"names"`
	// RefreshInterval configures the time after which the provided names are refreshed.
	// If not set, Prometheus uses its default value.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The type of DNS query to perform. One of SRV, A, AAAA or MX.
	// If not set, Prometheus uses its default value.
	// +kubebuilder:validation:Enum=SRV;A;AAAA;MX
	// +optional
	Type *string `json:"type"`
	// The port number used if the query type is not SRV
	// Ignored for SRV records
	// +optional
	Port *int `json:"port"`
}
