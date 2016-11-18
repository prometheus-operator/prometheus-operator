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

package spec

import (
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 PrometheusSpec `json:"spec"`
}

// PrometheusList is a list of Prometheuses.
type PrometheusList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*Prometheus `json:"items"`
}

// PrometheusSpec holds specification parameters of a Prometheus deployment.
type PrometheusSpec struct {
	ServiceMonitors    []ServiceMonitorSelection `json:"serviceMonitors"`
	EvaluationInterval string                    `json:"evaluationInterval"`
	Version            string                    `json:"version"`
	BaseImage          string                    `json:"baseImage"`
	Replicas           int32                     `json:"replicas"`
	Retention          string                    `json:"retention"`
	Storage            *StorageSpec              `json:"storage"`
	Alerting           AlertingSpec              `json:"alerting"`
	Resources          v1.ResourceRequirements   `json:"resources"`
	// Alerting        AlertingSpec               `json:"alerting"`
	// Remote          RemoteSpec                 `json:"remote"`
	// Persistence...
	// Sharding...
}

// AlertingSpec defines paramters for alerting configuration of Prometheus servers.
type AlertingSpec struct {
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
type StorageSpec struct {
	Class     string                     `json:"class"`
	Selector  *unversioned.LabelSelector `json:"selector"`
	Resources v1.ResourceRequirements    `json:"resources"`
}

// AlertmanagerEndpoints defines a selection of a single Endpoints object
// containing alertmanager IPs to fire alerts against.
type AlertmanagerEndpoints struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ServiceMonitorSelection selects service monitors by their labels.
type ServiceMonitorSelection struct {
	Selector unversioned.LabelSelector `json:"selector"`
}

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 ServiceMonitorSpec `json:"spec"`
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
type ServiceMonitorSpec struct {
	Endpoints []Endpoint                `json:"endpoints"`
	Selector  unversioned.LabelSelector `json:"selector"`
	// AllNamespaces     bool                      `json:"allNamespaces"`
	// Namespaces        []string                  `json:"namespaces"`
	// NamespaceSelector unversioned.LabelSelector `json:"namespaceSelector"`
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
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*ServiceMonitor `json:"items"`
}
