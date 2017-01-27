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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"
)

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PrometheusSpec    `json:"spec"`
	Status            *PrometheusStatus `json:"status,omitempty"`
}

// PrometheusList is a list of Prometheuses.
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*Prometheus `json:"items"`
}

// PrometheusSpec holds specification parameters of a Prometheus deployment.
type PrometheusSpec struct {
	// ServiceMonitors to be selected for target discovery.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Version of Prometheus to be deployed.
	Version string `json:"version,omitempty"`
	// When a Prometheus deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`
	// Base image to use for a Prometheus deployment.
	BaseImage string `json:"baseImage,omitempty"`
	// Number of instances to deploy for a Prometheus deployment.
	Replicas int32 `json:"replicas,omitempty"`
	// Time duration Prometheus shall retain data for.
	Retention string `json:"retention,omitempty"`
	// The external URL the Prometheus instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Prometheus is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Prometheus registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// Storage spec to specify how storage shall be used.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Define details regarding alerting.
	Alerting AlertingSpec `json:"alerting,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
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
	// AlertmanagerEndpoints Prometheus should fire alerts against.
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
type StorageSpec struct {
	// Name of the StorageClass to use when requesting storage provisioning. More
	// info: https://kubernetes.io/docs/user-guide/persistent-volumes/#storageclasses
	Class string `json:"class"`
	// A label query over volumes to consider for binding.
	Selector *metav1.LabelSelector `json:"selector"`
	// Resources represents the minimum resources the volume should have. More
	// info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources
	Resources v1.ResourceRequirements `json:"resources"`
}

// AlertmanagerEndpoints defines a selection of a single Endpoints object
// containing alertmanager IPs to fire alerts against.
type AlertmanagerEndpoints struct {
	// Namespace of Endpoints object.
	Namespace string `json:"namespace"`
	// Name of Endpoints object in Namespace.
	Name string `json:"name"`
	// Port the Alertmanager API is exposed on.
	Port intstr.IntOrString `json:"port"`
	// Scheme to use when firing alerts.
	Scheme string `json:"scheme"`
}

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceMonitorSpec `json:"spec"`
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
type ServiceMonitorSpec struct {
	// The label to use to retrieve the job name from.
	JobLabel string `json:"jobLabel,omitempty"`
	// A list of endpoints allowed as part of this ServiceMonitor.
	Endpoints []Endpoint `json:"endpoints,omitempty"`
	// Selector to select Endpoints objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the Endpoints objects are discovered from.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
}

// Endpoint defines a scrapeable endpoint serving Prometheus metrics.
type Endpoint struct {
	// Name of the service port this endpoint refers to. Mutually exclusive with targetPort.
	Port string `json:"port,omitempty"`
	// Name or number of the target port of the endpoint. Mutually exclusive with port.
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
	// HTTP path to scrape for metrics.
	Path string `json:"path,omitempty"`
	// HTTP scheme to use for scraping.
	Scheme string `json:"scheme,omitempty"`
	// Interval at which metrics should be scraped
	Interval string `json:"interval,omitempty"`
	// TLS configuration to use when scraping the endpoint
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// File to read bearer token for scraping targets.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
}

// TLSConfig specifies TLS configuration parameters.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"caFile,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"certFile,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"keyFile,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify,omitempty"`
}

// ServiceMonitorList is a list of ServiceMonitors.
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*ServiceMonitor `json:"items"`
}

type Alertmanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AlertmanagerSpec    `json:"spec"`
	Status            *AlertmanagerStatus `json:"status,omitempty"`
}

type AlertmanagerSpec struct {
	// Version the cluster should be on.
	Version string `json:"version,omitempty"`
	// Base image that is used to deploy pods.
	BaseImage string `json:"baseImage,omitempty"`
	// Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	Replicas int32 `json:"replicas,omitempty"`
	// Storage is the definition of how storage will be used by the Alertmanager
	// instances.
	Storage *StorageSpec `json:"storage,omitempty"`
	// ExternalURL is the URL under which Alertmanager is externally reachable
	// (for example, if Alertmanager is served via a reverse proxy). Used for
	// generating relative and absolute links back to Alertmanager itself. If the
	// URL has a path portion, it will be used to prefix all HTTP endpoints
	// served by Alertmanager. If omitted, relevant URL components will be
	// derived automatically.
	ExternalURL string `json:"externalUrl,omitempty"`
	// If set to true all actions on the underlaying managed objects are not
	// goint to be performed, except for delete actions.
	Paused bool `json:"paused,omitempty"`
}

type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of third party objects
	Items []Alertmanager `json:"items"`
}

type AlertmanagerStatus struct {
	// Represents whether any actions on the underlaying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// cluster (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// cluster that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Alertmanager cluster.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Alertmanager cluster.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

type NamespaceSelector struct {
	Any        bool     `json:"any,omitempty"`
	MatchNames []string `json:"matchNames,omitempty"`

	// TODO(fabxc): this should embed metav1.LabelSelector eventually.
	// Currently the selector is only used for namespaces which require more complex
	// implementation to support label selections.
}
