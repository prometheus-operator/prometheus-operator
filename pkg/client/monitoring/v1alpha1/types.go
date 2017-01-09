// Copyright 2016 The prometheus-operator Authors
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

package v1alpha1

import (
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/util/intstr"
)

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            PrometheusSpec    `json:"spec"`
	Status          *PrometheusStatus `json:"status,omitempty"`
}

// PrometheusList is a list of Prometheuses.
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*Prometheus `json:"items"`
}

// PrometheusSpec holds specification parameters of a Prometheus deployment.
type PrometheusSpec struct {
	ServiceMonitorSelector *metav1.LabelSelector   `json:"serviceMonitorSelector"`
	Version                string                  `json:"version"`
	Paused                 bool                    `json:"paused"`
	BaseImage              string                  `json:"baseImage"`
	Replicas               int32                   `json:"replicas"`
	Retention              string                  `json:"retention"`
	ExternalURL            string                  `json:"externalUrl"`
	Storage                *StorageSpec            `json:"storage"`
	Alerting               AlertingSpec            `json:"alerting"`
	Resources              v1.ResourceRequirements `json:"resources"`
	// EvaluationInterval string                    `json:"evaluationInterval"`
	// Remote          RemoteSpec                 `json:"remote"`
	// Sharding...
}

type PrometheusStatus struct {
	// Represents whether any actions on the underlaying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`

	// Total number of non-terminated pods targeted by this Prometheus deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`

	// Total number of non-terminated pods targeted by this Prometheus deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`

	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Prometheus deployment.
	AvailableReplicas int32 `json:"availableReplicas"`

	// Total number of unavailable pods targeted by this Prometheus deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

// AlertingSpec defines paramters for alerting configuration of Prometheus servers.
type AlertingSpec struct {
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
type StorageSpec struct {
	Class     string                  `json:"class"`
	Selector  *metav1.LabelSelector   `json:"selector"`
	Resources v1.ResourceRequirements `json:"resources"`
}

// AlertmanagerEndpoints defines a selection of a single Endpoints object
// containing alertmanager IPs to fire alerts against.
type AlertmanagerEndpoints struct {
	Namespace string             `json:"namespace"`
	Name      string             `json:"name"`
	Port      intstr.IntOrString `json:"port"`
	Scheme    string             `json:"scheme"`
}

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            ServiceMonitorSpec `json:"spec"`
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
type ServiceMonitorSpec struct {
	JobLabel          string               `json:"jobLabel"`
	Endpoints         []Endpoint           `json:"endpoints"`
	Selector          metav1.LabelSelector `json:"selector"`
	NamespaceSelector Selector             `json:"namespaceSelector"`
	// AllNamespaces     bool                      `json:"allNamespaces"`
	// Namespaces        []string                  `json:"namespaces"`
	// NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`
}

// Endpoint defines a scrapeable endpoint serving Prometheus metrics.
type Endpoint struct {
	Port       string             `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort"`
	Path       string             `json:"path"`
	Scheme     string             `json:"scheme"`
	Interval   string             `json:"interval"`
}

// ServiceMonitorList is a list of ServiceMonitors.
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*ServiceMonitor `json:"items"`
}

type Alertmanager struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            AlertmanagerSpec `json:"spec"`
}

type AlertmanagerSpec struct {
	// Version the cluster should be on.
	Version string `json:"version"`
	// Base image that is used to deploy pods.
	BaseImage string `json:"baseImage"`
	// Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	Replicas int32 `json:"replicas"`
	// Storage is the definition of how storage will be used by the Alertmanager
	// instances.
	Storage *StorageSpec `json:"storage"`
	// ExternalURL is the URL under which Alertmanager is externally reachable
	// (for example, if Alertmanager is served via a reverse proxy). Used for
	// generating relative and absolute links back to Alertmanager itself. If the
	// URL has a path portion, it will be used to prefix all HTTP endpoints
	// served by Alertmanager. If omitted, relevant URL components will be
	// derived automatically.
	ExternalURL string `json:"externalUrl,omitempty"`
}

type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of third party objects
	Items []Alertmanager `json:"items"`
}

type Selector struct {
	Any        bool     `json:"any,omitempty"`
	MatchNames []string `json:"matchNames,omitempty"`

	// TODO(fabxc): this should embed metav1.LabelSelector eventually.
	// Currently the selector is only used for namespaces which require more complex
	// implementation to support label selections.
}

type ListOptions v1.ListOptions
type DeleteOptions v1.DeleteOptions
