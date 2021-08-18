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
)

const (
	AlertmanagersKind   = "Alertmanager"
	AlertmanagerName    = "alertmanagers"
	AlertManagerKindKey = "alertmanager"
)

// Alertmanager describes an Alertmanager cluster.
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Alertmanager"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The desired replicas number of Alertmanagers"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Alertmanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Alertmanager cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec AlertmanagerSpec `json:"spec"`
	// Most recent observed status of the Alertmanager cluster. Read-only. Not
	// included when requesting from the apiserver, only from the Prometheus
	// Operator API itself. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status *AlertmanagerStatus `json:"status,omitempty"`
}

// AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type AlertmanagerSpec struct {
	// PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Alertmanager is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Version the cluster should be on.
	Version string `json:"version,omitempty"`
	// Tag of Alertmanager container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	// Deprecated: use 'image' instead.  The image tag can be specified
	// as part of the image URL.
	Tag string `json:"tag,omitempty"`
	// SHA of Alertmanager container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	// Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image URL.
	SHA string `json:"sha,omitempty"`
	// Base image that is used to deploy pods, without tag.
	// Deprecated: use 'image' instead
	BaseImage string `json:"baseImage,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the Alertmanager
	// object, which shall be mounted into the Alertmanager Pods.
	// The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager
	// object, which shall be mounted into the Alertmanager Pods.
	// The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// ConfigSecret is the name of a Kubernetes Secret in the same namespace as the
	// Alertmanager object, which contains configuration for this Alertmanager
	// instance. Defaults to 'alertmanager-<alertmanager-name>'
	// The secret is mounted into /etc/alertmanager/config.
	ConfigSecret string `json:"configSecret,omitempty"`
	// Log level for Alertmanager to be configured with.
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for Alertmanager to be configured with.
	LogFormat string `json:"logFormat,omitempty"`
	// Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	Replicas *int32 `json:"replicas,omitempty"`
	// Time duration Alertmanager shall retain data for. Default is '120h',
	// and must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).
	Retention string `json:"retention,omitempty"`
	// Storage is the definition of how storage will be used by the Alertmanager
	// instances.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output StatefulSet definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,
	// that are generated as a result of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// The external URL the Alertmanager instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Alertmanager is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Alertmanager registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// If set to true all actions on the underlying managed objects are not
	// goint to be performed, except for delete actions.
	Paused bool `json:"paused,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// If specified, the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// If specified, the pod's topology spread constraints.
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// ListenLocal makes the Alertmanager server listen on loopback, so that it
	// does not bind against the Pod IP. Note this is only for the Alertmanager
	// UI, not the gossip communication.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// Containers allows injecting additional containers. This is meant to
	// allow adding an authentication proxy to an Alertmanager pod.
	// Containers described here modify an operator generated container if they
	// share the same name and modifications are done via a strategic merge
	// patch. The current container names are: `alertmanager` and
	// `config-reloader`. Overriding containers is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that
	// this behaviour may break at any time without notice.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the Alertmanager configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster.
	AdditionalPeers []string `json:"additionalPeers,omitempty"`
	// ClusterAdvertiseAddress is the explicit address to advertise in cluster.
	// Needs to be provided for non RFC1918 [1] (public) addresses.
	// [1] RFC1918: https://tools.ietf.org/html/rfc1918
	ClusterAdvertiseAddress string `json:"clusterAdvertiseAddress,omitempty"`
	// Interval between gossip attempts.
	ClusterGossipInterval string `json:"clusterGossipInterval,omitempty"`
	// Interval between pushpull attempts.
	ClusterPushpullInterval string `json:"clusterPushpullInterval,omitempty"`
	// Timeout for cluster peering.
	ClusterPeerTimeout string `json:"clusterPeerTimeout,omitempty"`
	// Port name used for the pods and governing service.
	// This defaults to web
	PortName string `json:"portName,omitempty"`
	// ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica.
	// Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each.
	ForceEnableClusterMode bool `json:"forceEnableClusterMode,omitempty"`
	// AlertmanagerConfigs to be selected for to merge and configure Alertmanager with.
	AlertmanagerConfigSelector *metav1.LabelSelector `json:"alertmanagerConfigSelector,omitempty"`
	// Namespaces to be selected for AlertmanagerConfig discovery. If nil, only
	// check own namespace.
	AlertmanagerConfigNamespaceSelector *metav1.LabelSelector `json:"alertmanagerConfigNamespaceSelector,omitempty"`
}

// AlertmanagerList is a list of Alertmanagers.
// +k8s:openapi-gen=true
type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Alertmanagers
	Items []Alertmanager `json:"items"`
}

// AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only. Not
// included when requesting from the apiserver, only from the Prometheus
// Operator API itself. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type AlertmanagerStatus struct {
	// Represents whether any actions on the underlying managed objects are
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

// DeepCopyObject implements the runtime.Object interface.
func (l *Alertmanager) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
