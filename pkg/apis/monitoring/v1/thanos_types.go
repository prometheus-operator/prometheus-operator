// Copyright 2020 The prometheus-operator Authors
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
	ThanosRulerKind    = "ThanosRuler"
	ThanosRulerName    = "thanosrulers"
	ThanosRulerKindKey = "thanosrulers"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="ruler"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Thanos Ruler"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.availableReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type == 'Reconciled')].status"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1
// +kubebuilder:subresource:status

// The `ThanosRuler` custom resource definition (CRD) defines a desired [Thanos Ruler](https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md) setup to run in a Kubernetes cluster.
//
// A `ThanosRuler` instance requires at least one compatible Prometheus API endpoint (either Thanos Querier or Prometheus services).
//
// The resource defines via label and namespace selectors which `PrometheusRule` objects should be associated to the deployed Thanos Ruler instances.
type ThanosRuler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the ThanosRuler cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec ThanosRulerSpec `json:"spec"`
	// Most recent observed status of the ThanosRuler cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status ThanosRulerStatus `json:"status,omitempty"`
}

// ThanosRulerList is a list of ThanosRulers.
// +k8s:openapi-gen=true
type ThanosRulerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []ThanosRuler `json:"items"`
}

// ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type ThanosRulerSpec struct {
	// Version of Thanos to be deployed.
	// +optional
	Version *string `json:"version,omitempty"`

	// PodMetadata configures labels and annotations which are propagated to the ThanosRuler pods.
	//
	// The following items are reserved and cannot be overridden:
	// * "app.kubernetes.io/name" label, set to "thanos-ruler".
	// * "app.kubernetes.io/managed-by" label, set to "prometheus-operator".
	// * "app.kubernetes.io/instance" label, set to the name of the ThanosRuler instance.
	// * "thanos-ruler" label, set to the name of the ThanosRuler instance.
	// * "kubectl.kubernetes.io/default-container" annotation, set to "thanos-ruler".
	// +optional
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`

	// Thanos container image URL.
	Image string `json:"image,omitempty"`

	// Image pull policy for the 'thanos', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// An optional list of references to secrets in the same namespace
	// to use for pulling thanos images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// When a ThanosRuler deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`

	// Number of thanos ruler instances to deploy.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Resources defines the resource requirements for single Pods.
	// If not provided, no requests/limits will be set
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// If specified, the pod's scheduling constraints.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// If specified, the pod's topology spread constraints.
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// Defines the DNS policy for the pods.
	//
	// +optional
	DNSPolicy *DNSPolicy `json:"dnsPolicy,omitempty"`
	// Defines the DNS configuration for the pods.
	//
	// +optional
	DNSConfig *PodDNSConfig `json:"dnsConfig,omitempty"`

	// Indicates whether information about services should be injected into pod's environment variables
	// +optional
	EnableServiceLinks *bool `json:"enableServiceLinks,omitempty"`

	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// The name of the service name used by the underlying StatefulSet(s) as the governing service.
	// If defined, the Service  must be created before the ThanosRuler resource in the same namespace and it must define a selector that matches the pod labels.
	// If empty, the operator will create and manage a headless service named `thanos-ruler-operated` for ThanosRuler resources.
	// When deploying multiple ThanosRuler resources in the same namespace, it is recommended to specify a different value for each.
	// See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details.
	// +optional
	// +kubebuilder:validation:MinLength=1
	ServiceName *string `json:"serviceName,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Thanos Ruler Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Storage spec to specify how storage shall be used.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
	// be appended to other volumes that are generated as a result of StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the ruler container,
	// that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// Configures object storage.
	//
	// The configuration format is defined at https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage
	//
	// The operator performs no validation of the configuration.
	//
	// `objectStorageConfigFile` takes precedence over this field.
	//
	// +optional
	ObjectStorageConfig *v1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
	// Configures the path of the object storage configuration file.
	//
	// The configuration format is defined at https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage
	//
	// The operator performs no validation of the configuration file.
	//
	// This field takes precedence over `objectStorageConfig`.
	//
	// +optional
	ObjectStorageConfigFile *string `json:"objectStorageConfigFile,omitempty"`

	// ListenLocal makes the Thanos ruler listen on loopback, so that it
	// does not bind against the Pod IP.
	ListenLocal bool `json:"listenLocal,omitempty"`

	// Configures the list of Thanos Query endpoints from which to query metrics.
	//
	// For Thanos >= v0.11.0, it is recommended to use `queryConfig` instead.
	//
	// `queryConfig` takes precedence over this field.
	//
	// +optional
	QueryEndpoints []string `json:"queryEndpoints,omitempty"`

	// Configures the list of Thanos Query endpoints from which to query metrics.
	//
	// The configuration format is defined at https://thanos.io/tip/components/rule.md/#query-api
	//
	// It requires Thanos >= v0.11.0.
	//
	// The operator performs no validation of the configuration.
	//
	// This field takes precedence over `queryEndpoints`.
	//
	// +optional
	QueryConfig *v1.SecretKeySelector `json:"queryConfig,omitempty"`

	// Configures the list of Alertmanager endpoints to send alerts to.
	//
	// For Thanos >= v0.10.0, it is recommended to use `alertmanagersConfig` instead.
	//
	// `alertmanagersConfig` takes precedence over this field.
	//
	// +optional
	AlertManagersURL []string `json:"alertmanagersUrl,omitempty"`
	// Configures the list of Alertmanager endpoints to send alerts to.
	//
	// The configuration format is defined at https://thanos.io/tip/components/rule.md/#alertmanager.
	//
	// It requires Thanos >= v0.10.0.
	//
	// The operator performs no validation of the configuration.
	//
	// This field takes precedence over `alertmanagersUrl`.
	//
	// +optional
	AlertManagersConfig *v1.SecretKeySelector `json:"alertmanagersConfig,omitempty"`

	// PrometheusRule objects to be selected for rule evaluation. An empty
	// label selector matches all objects. A null label selector matches no
	// objects.
	//
	// +optional
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`
	// Namespaces to be selected for Rules discovery. If unspecified, only
	// the same namespace as the ThanosRuler object is in is used.
	//
	// +optional
	RuleNamespaceSelector *metav1.LabelSelector `json:"ruleNamespaceSelector,omitempty"`

	// EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert
	// and metric that is user created. The label value will always be the namespace of the object that is
	// being created.
	EnforcedNamespaceLabel string `json:"enforcedNamespaceLabel,omitempty"`
	// List of references to PrometheusRule objects
	// to be excluded from enforcing a namespace label of origin.
	// Applies only if enforcedNamespaceLabel set to true.
	// +optional
	ExcludedFromEnforcement []ObjectReference `json:"excludedFromEnforcement,omitempty"`
	// PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing
	// of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
	// Make sure both ruleNamespace and ruleName are set for each pair
	// Deprecated: use excludedFromEnforcement instead.
	// +optional
	PrometheusRulesExcludedFromEnforce []PrometheusRuleExcludeConfig `json:"prometheusRulesExcludedFromEnforce,omitempty"`

	// Log level for ThanosRuler to be configured with.
	// +kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for ThanosRuler to be configured with.
	// +kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`

	// Port name used for the pods and governing service.
	// Defaults to `web`.
	// +kubebuilder:default:="web"
	PortName string `json:"portName,omitempty"`

	// Interval between consecutive evaluations.
	// +kubebuilder:default:="15s"
	EvaluationInterval Duration `json:"evaluationInterval,omitempty"`

	// Max time to tolerate prometheus outage for restoring "for" state of alert.
	// It requires Thanos >= v0.30.0.
	// +optional
	RuleOutageTolerance *Duration `json:"ruleOutageTolerance,omitempty"`

	// The default rule group's query offset duration to use.
	// It requires Thanos >= v0.38.0.
	// +optional
	RuleQueryOffset *Duration `json:"ruleQueryOffset,omitempty"`

	// How many rules can be evaluated concurrently.
	// It requires Thanos >= v0.37.0.
	// +kubebuilder:validation:Minimum=1
	//
	// +optional
	RuleConcurrentEval *int32 `json:"ruleConcurrentEval,omitempty"`

	// Time duration ThanosRuler shall retain data for. Default is '24h', and
	// must match the regular expression `[0-9]+(ms|s|m|h|d|w|y)` (milliseconds
	// seconds minutes hours days weeks years).
	//
	// The field has no effect when remote-write is configured since the Ruler
	// operates in stateless mode.
	//
	// +kubebuilder:default:="24h"
	Retention Duration `json:"retention,omitempty"`

	// Containers allows injecting additional containers or modifying operator generated
	// containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or
	// to change the behavior of an operator generated container. Containers described here modify
	// an operator generated container if they share the same name and modifications are done via a
	// strategic merge patch. The current container names are: `thanos-ruler` and `config-reloader`.
	// Overriding containers is entirely outside the scope of what the maintainers will support and by doing
	// so, you accept that this behaviour may break at any time without notice.
	// +optional
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the ThanosRuler configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	// +optional
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// Configures tracing.
	//
	// The configuration format is defined at https://thanos.io/tip/thanos/tracing.md/#configuration
	//
	// This is an *experimental feature*, it may change in any upcoming release
	// in a breaking way.
	//
	// The operator performs no validation of the configuration.
	//
	// `tracingConfigFile` takes precedence over this field.
	//
	//+optional
	TracingConfig *v1.SecretKeySelector `json:"tracingConfig,omitempty"`
	// Configures the path of the tracing configuration file.
	//
	// The configuration format is defined at https://thanos.io/tip/thanos/tracing.md/#configuration
	//
	// This is an *experimental feature*, it may change in any upcoming release
	// in a breaking way.
	//
	// The operator performs no validation of the configuration file.
	//
	// This field takes precedence over `tracingConfig`.
	//
	//+optional
	TracingConfigFile string `json:"tracingConfigFile,omitempty"`

	// Configures the external label pairs of the ThanosRuler resource.
	//
	// A default replica label `thanos_ruler_replica` will be always added as a
	// label with the value of the pod's name.
	//
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Configures the label names which should be dropped in Thanos Ruler
	// alerts.
	//
	// The replica label `thanos_ruler_replica` will always be dropped from the alerts.
	//
	// +optional
	AlertDropLabels []string `json:"alertDropLabels,omitempty"`

	// The external URL the Thanos Ruler instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Thanos Ruler is not
	// served from root of a DNS name.
	ExternalPrefix string `json:"externalPrefix,omitempty"`
	// The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path.
	RoutePrefix string `json:"routePrefix,omitempty"`

	// GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads
	// recorded rule data.
	// Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
	// Maps to the '--grpc-server-tls-*' CLI args.
	// +optional
	GRPCServerTLSConfig *TLSConfig `json:"grpcServerTlsConfig,omitempty"`

	// The external Query URL the Thanos Ruler will set in the 'Source' field
	// of all alerts.
	// Maps to the '--alert.query-url' CLI arg.
	AlertQueryURL string `json:"alertQueryUrl,omitempty"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.
	// +optional
	MinReadySeconds *uint32 `json:"minReadySeconds,omitempty"`

	// Configures alert relabeling in Thanos Ruler.
	//
	// Alert relabel configuration must have the form as specified in the
	// official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs
	//
	// The operator performs no validation of the configuration.
	//
	// `alertRelabelConfigFile` takes precedence over this field.
	//
	// +optional
	AlertRelabelConfigs *v1.SecretKeySelector `json:"alertRelabelConfigs,omitempty"`
	// Configures the path to the alert relabeling configuration file.
	//
	// Alert relabel configuration must have the form as specified in the
	// official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs
	//
	// The operator performs no validation of the configuration file.
	//
	// This field takes precedence over `alertRelabelConfig`.
	//
	// +optional
	AlertRelabelConfigFile *string `json:"alertRelabelConfigFile,omitempty"`

	// Pods' hostAliases configuration
	// +listType=map
	// +listMapKey=ip
	HostAliases []HostAlias `json:"hostAliases,omitempty"`

	// AdditionalArgs allows setting additional arguments for the ThanosRuler container.
	// It is intended for e.g. activating hidden flags which are not supported by
	// the dedicated configuration options yet. The arguments are passed as-is to the
	// ThanosRuler container which may cause issues if they are invalid or not supported
	// by the given ThanosRuler version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument the reconciliation will
	// fail and an error will be logged.
	// +optional
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`

	// Defines the configuration of the ThanosRuler web server.
	// +optional
	Web *ThanosRulerWebSpec `json:"web,omitempty"`

	// Defines the list of remote write configurations.
	//
	// When the list isn't empty, the ruler is configured with stateless mode.
	//
	// It requires Thanos >= 0.24.0.
	//
	// +optional
	RemoteWrite []RemoteWriteSpec `json:"remoteWrite,omitempty"`

	// Optional duration in seconds the pod needs to terminate gracefully.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down) which may lead to data corruption.
	//
	// Defaults to 120 seconds.
	//
	// +kubebuilder:validation:Minimum:=0
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// ThanosRulerWebSpec defines the configuration of the ThanosRuler web server.
// +k8s:openapi-gen=true
type ThanosRulerWebSpec struct {
	WebConfigFileFields `json:",inline"`
}

// ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only.
// More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type ThanosRulerStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this ThanosRuler deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this ThanosRuler deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The current state of the ThanosRuler object.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

func (tr *ThanosRuler) ExpectedReplicas() int {
	if tr.Spec.Replicas == nil {
		return 1
	}
	return int(*tr.Spec.Replicas)
}

func (tr *ThanosRuler) SetReplicas(i int)            { tr.Status.Replicas = int32(i) }
func (tr *ThanosRuler) SetUpdatedReplicas(i int)     { tr.Status.UpdatedReplicas = int32(i) }
func (tr *ThanosRuler) SetAvailableReplicas(i int)   { tr.Status.AvailableReplicas = int32(i) }
func (tr *ThanosRuler) SetUnavailableReplicas(i int) { tr.Status.UnavailableReplicas = int32(i) }

// DeepCopyObject implements the runtime.Object interface.
func (l *ThanosRuler) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ThanosRulerList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
