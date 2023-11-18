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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ServiceMonitorsKind   = "ServiceMonitor"
	ServiceMonitorName    = "servicemonitors"
	ServiceMonitorKindKey = "servicemonitor"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="smon"

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Service selection for target discovery by
	// Prometheus.
	Spec ServiceMonitorSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ServiceMonitor) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ServiceMonitorSpec defines the specification parameters for a ServiceMonitor.
// +k8s:openapi-gen=true
type ServiceMonitorSpec struct {
	// `jobLabel` selects the label from the associated Kubernetes `Service`
	// object which will be used as the `job` label for all metrics.
	//
	// For example if `jobLabel` is set to `foo` and the Kubernetes `Service`
	// object is labeled with `foo: bar`, then Prometheus adds the `job="bar"`
	// label to all ingested metrics.
	//
	// If the value of this field is empty or if the label doesn't exist for
	// the given Service, the `job` label of the metrics defaults to the name
	// of the associated Kubernetes `Service`.
	JobLabel string `json:"jobLabel,omitempty"`

	// `targetLabels` defines the labels which are transferred from the
	// associated Kubernetes `Service` object onto the ingested metrics.
	//
	// +optional
	TargetLabels []string `json:"targetLabels,omitempty"`
	// `podTargetLabels` defines the labels which are transferred from the
	// associated Kubernetes `Pod` object onto the ingested metrics.
	//
	// +optional
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`

	// List of endpoints part of this ServiceMonitor.
	//
	// +optional
	Endpoints []Endpoint `json:"endpoints"`

	// Label selector to select the Kubernetes `Endpoints` objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the Kubernetes `Endpoints` objects
	// are discovered from.
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
	// It requires Prometheus >= v2.37.0.
	//
	// +optional
	AttachMetadata *AttachMetadata `json:"attachMetadata,omitempty"`
}

// ServiceMonitorList is a list of ServiceMonitors.
// +k8s:openapi-gen=true
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of ServiceMonitors
	Items []*ServiceMonitor `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ServiceMonitorList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
