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
	// RelabelConfigs defines how to rewrite the target's labels before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	// +optional
	RelabelConfigs []*v1.RelabelConfig `json:"relabelings,omitempty"`
	// MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics).
	// +optional
	MetricsPath string `json:"metricsPath,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	// +optional
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	// +optional
	HonorLabels *bool `json:"honorLabels,omitempty"`
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
}
