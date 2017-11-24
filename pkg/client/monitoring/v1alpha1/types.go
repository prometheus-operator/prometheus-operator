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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Prometheus cluster. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Spec PrometheusSpec `json:"spec"`
	// Most recent observed status of the Prometheus cluster. Read-only. Not
	// included when requesting from the apiserver, only from the Prometheus
	// Operator API itself. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Status *PrometheusStatus `json:"status,omitempty"`
}

// PrometheusList is a list of Prometheuses.
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []*Prometheus `json:"items"`
}

// Specification of the desired behavior of the Prometheus cluster. More info:
// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
type PrometheusSpec struct {
	// Standard object’s metadata. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	// Metadata Labels and Annotations gets propagated to the prometheus pods.
	PodMetadata *metav1.ObjectMeta `json:"podMetadata,omitempty"`
	// ServiceMonitors to be selected for target discovery.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Version of Prometheus to be deployed.
	Version string `json:"version,omitempty"`
	// When a Prometheus deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`
	// Base image to use for a Prometheus deployment.
	BaseImage string `json:"baseImage,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Number of instances to deploy for a Prometheus deployment.
	Replicas *int32 `json:"replicas,omitempty"`
	// Time duration Prometheus shall retain data for.
	Retention string `json:"retention,omitempty"`
	// Interval between consecutive evaluations.
	EvaluationInterval string `json:"evaluationInterval,omitempty"`
	// The labels to add to any time series or alerts when communicating with
	// external systems (federation, remote storage, Alertmanager).
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`
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
	// A selector to select which ConfigMaps to mount for loading rule files from.
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`
	// Define details regarding alerting.
	Alerting AlertingSpec `json:"alerting,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.
	// Secrets changes after initial creation of a Prometheus object are not
	// reflected in the running Pods. To change the secrets mounted into the
	// Prometheus Pods, the object must be deleted and recreated with the new list
	// of secrets.
	Secrets []string `json:"secrets,omitempty"`
	// EvaluationInterval string                    `json:"evaluationInterval"`
	// Remote          RemoteSpec                 `json:"remote"`
	// Sharding...
}

// Most recent observed status of the Prometheus cluster. Read-only. Not
// included when requesting from the apiserver, only from the Prometheus
// Operator API itself. More info:
// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
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

// AlertingSpec defines parameters for alerting configuration of Prometheus servers.
type AlertingSpec struct {
	// AlertmanagerEndpoints Prometheus should fire alerts against.
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
type StorageSpec struct {
	// Name of the StorageClass to use when requesting storage provisioning. More
	// info: https://kubernetes.io/docs/user-guide/persistent-volumes/#storageclasses
	// DEPRECATED
	Class string `json:"class"`
	// A label query over volumes to consider for binding.
	// DEPRECATED
	Selector *metav1.LabelSelector `json:"selector"`
	// Resources represents the minimum resources the volume should have. More
	// info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources
	// DEPRECATED
	Resources v1.ResourceRequirements `json:"resources"`
	// A PVC spec to be used by the Prometheus StatefulSets.
	VolumeClaimTemplate v1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
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
	// Prefix for the HTTP path alerts are pushed to.
	PathPrefix string `json:"pathPrefix"`
}

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Service selection for target discrovery by
	// Prometheus.
	Spec ServiceMonitorSpec `json:"spec"`
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
type ServiceMonitorSpec struct {
	// The label to use to retrieve the job name from.
	JobLabel string `json:"jobLabel,omitempty"`
	// A list of endpoints allowed as part of this ServiceMonitor.
	Endpoints []Endpoint `json:"endpoints"`
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
	// Timeout after which the scrape is ended
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`
	// TLS configuration to use when scraping the endpoint
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// File to read bearer token for scraping targets.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	HonorLabels bool `json:"honorLabels,omitempty"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
}

// BasicAuth allow an endpoint to authenticate over basic authentication
// More info: https://prometheus.io/docs/operating/configuration/#endpoints
type BasicAuth struct {
	// The secret that contains the username for authenticate
	Username v1.SecretKeySelector `json:"username,omitempty"`
	// The secret that contains the password for authenticate
	Password v1.SecretKeySelector `json:"password,omitempty"`
}

// TLSConfig specifies TLS configuration parameters.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `json:"caFile,omitempty"`
	// The client cert file for the targets.
	CertFile string `json:"certFile,omitempty"`
	// The client key file for the targets.
	KeyFile string `json:"keyFile,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `json:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// A list of ServiceMonitors.
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of ServiceMonitors
	Items []*ServiceMonitor `json:"items"`
}

// Describes an Alertmanager cluster.
type Alertmanager struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Alertmanager cluster. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Spec AlertmanagerSpec `json:"spec"`
	// Most recent observed status of the Alertmanager cluster. Read-only. Not
	// included when requesting from the apiserver, only from the Prometheus
	// Operator API itself. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Status *AlertmanagerStatus `json:"status,omitempty"`
}

// Specification of the desired behavior of the Alertmanager cluster. More info:
// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
type AlertmanagerSpec struct {
	// Standard object’s metadata. More info:
	// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	// Metadata Labels and Annotations gets propagated to the prometheus pods.
	PodMetadata *metav1.ObjectMeta `json:"podMetadata,omitempty"`
	// Version the cluster should be on.
	Version string `json:"version,omitempty"`
	// Base image that is used to deploy pods.
	BaseImage string `json:"baseImage,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	Replicas *int32 `json:"replicas,omitempty"`
	// Storage is the definition of how storage will be used by the Alertmanager
	// instances.
	Storage *StorageSpec `json:"storage,omitempty"`
	// The external URL the Alertmanager instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Alertmanager is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Alertmanager registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// If set to true all actions on the underlaying managed objects are not
	// goint to be performed, except for delete actions.
	Paused bool `json:"paused,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// A list of Alertmanagers.
type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Alertmanagers
	Items []Alertmanager `json:"items"`
}

// Most recent observed status of the Alertmanager cluster. Read-only. Not
// included when requesting from the apiserver, only from the Prometheus
// Operator API itself. More info:
// http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
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

// A selector for selecting namespaces either selecting all namespaces or a
// list of namespaces.
type NamespaceSelector struct {
	// Boolean describing whether all namespaces are selected in contrast to a
	// list restricting them.
	Any bool `json:"any,omitempty"`
	// List of namespace names.
	MatchNames []string `json:"matchNames,omitempty"`

	// TODO(fabxc): this should embed metav1.LabelSelector eventually.
	// Currently the selector is only used for namespaces which require more complex
	// implementation to support label selections.
}

func (l *Alertmanager) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for Alertmanager")
}

func (l *AlertmanagerList) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for AlertmanagerList")
}

func (l *Prometheus) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for Prometheus")
}

func (l *PrometheusList) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for PrometheusList")
}

func (l *ServiceMonitor) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for ServiceMonitor")
}

func (l *ServiceMonitorList) DeepCopyObject() runtime.Object {
	panic("DeepCopyObject not implemented for ServiceMonitorList")
}
