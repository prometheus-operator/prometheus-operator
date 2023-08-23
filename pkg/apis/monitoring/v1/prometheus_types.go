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
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PrometheusesKind  = "Prometheus"
	PrometheusName    = "prometheuses"
	PrometheusKindKey = "prometheus"
)

// PrometheusInterface is used by Prometheus and PrometheusAgent to share common methods, e.g. config generation.
// +k8s:deepcopy-gen=false
type PrometheusInterface interface {
	metav1.ObjectMetaAccessor
	GetTypeMeta() metav1.TypeMeta
	GetCommonPrometheusFields() CommonPrometheusFields
	SetCommonPrometheusFields(CommonPrometheusFields)
	GetStatus() PrometheusStatus
}

func (l *Prometheus) GetCommonPrometheusFields() CommonPrometheusFields {
	return l.Spec.CommonPrometheusFields
}

func (l *Prometheus) SetCommonPrometheusFields(f CommonPrometheusFields) {
	l.Spec.CommonPrometheusFields = f
}

func (l *Prometheus) GetTypeMeta() metav1.TypeMeta {
	return l.TypeMeta
}

func (l *Prometheus) GetStatus() PrometheusStatus {
	return l.Status
}

// CommonPrometheusFields are the options available to both the Prometheus server and agent.
// +k8s:deepcopy-gen=true
type CommonPrometheusFields struct {
	// PodMetadata configures labels and annotations which are propagated to the Prometheus pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`

	// ServiceMonitors to be selected for target discovery. An empty label
	// selector matches all objects. A null label selector matches no objects.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`
	// and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is *deprecated* and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Namespaces to match for ServicedMonitors discovery. An empty label selector
	// matches all namespaces. A null label selector matches the current
	// namespace only.
	ServiceMonitorNamespaceSelector *metav1.LabelSelector `json:"serviceMonitorNamespaceSelector,omitempty"`

	// *Experimental* PodMonitors to be selected for target discovery. An empty
	// label selector matches all objects. A null label selector matches no
	// objects.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`
	// and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is *deprecated* and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	PodMonitorSelector *metav1.LabelSelector `json:"podMonitorSelector,omitempty"`
	// Namespaces to match for PodMonitors discovery. An empty label selector
	// matches all namespaces. A null label selector matches the current
	// namespace only.
	PodMonitorNamespaceSelector *metav1.LabelSelector `json:"podMonitorNamespaceSelector,omitempty"`

	// *Experimental* Probes to be selected for target discovery. An empty
	// label selector matches all objects. A null label selector matches no
	// objects.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`
	// and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is *deprecated* and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	ProbeSelector *metav1.LabelSelector `json:"probeSelector,omitempty"`
	// *Experimental* Namespaces to match for Probe discovery. An empty label
	// selector matches all namespaces. A null label selector matches the
	// current namespace only.
	ProbeNamespaceSelector *metav1.LabelSelector `json:"probeNamespaceSelector,omitempty"`

	// *Experimental* ScrapeConfigs to be selected for target discovery. An
	// empty label selector matches all objects. A null label selector matches
	// no objects.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`
	// and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is *deprecated* and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	ScrapeConfigSelector *metav1.LabelSelector `json:"scrapeConfigSelector,omitempty"`
	// Namespaces to match for ScrapeConfig discovery. An empty label selector
	// matches all namespaces. A null label selector matches the current
	// current namespace only.
	ScrapeConfigNamespaceSelector *metav1.LabelSelector `json:"scrapeConfigNamespaceSelector,omitempty"`

	// Version of Prometheus being deployed. The operator uses this information
	// to generate the Prometheus StatefulSet + configuration files.
	//
	// If not specified, the operator assumes the latest upstream version of
	// Prometheus available at the time when the version of the operator was
	// released.
	Version string `json:"version,omitempty"`

	// When a Prometheus deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`

	// Container image name for Prometheus. If specified, it takes precedence
	// over the `spec.baseImage`, `spec.tag` and `spec.sha` fields.
	//
	// Specifying `spec.version` is still necessary to ensure the Prometheus
	// Operator knows which version of Prometheus is being configured.
	//
	// If neither `spec.image` nor `spec.baseImage` are defined, the operator
	// will use the latest upstream version of Prometheus available at the time
	// when the operator was released.
	//
	// +optional
	Image *string `json:"image,omitempty"`
	// Image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// An optional list of references to Secrets in the same namespace
	// to use for pulling images from registries.
	// See http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Number of replicas of each shard to deploy for a Prometheus deployment.
	// `spec.replicas` multiplied by `spec.shards` is the total number of Pods
	// created.
	//
	// Default: 1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// EXPERIMENTAL: Number of shards to distribute targets onto. `spec.replicas`
	// multiplied by `spec.shards` is the total number of Pods created.
	//
	// Note that scaling down shards will not reshard data onto remaining
	// instances, it must be manually moved. Increasing shards will not reshard
	// data either but it will continue to be available from the same
	// instances. To query globally, use Thanos sidecar and Thanos querier or
	// remote write data to a central location.
	//
	// Sharding is performed on the content of the `__address__` target meta-label
	// for PodMonitors and ServiceMonitors and `__param_target__` for Probes.
	//
	// Default: 1
	// +optional
	Shards *int32 `json:"shards,omitempty"`

	// Name of Prometheus external label used to denote the replica name.
	// The external label will _not_ be added when the field is set to the
	// empty string (`""`).
	//
	// Default: "prometheus_replica"
	// +optional
	ReplicaExternalLabelName *string `json:"replicaExternalLabelName,omitempty"`
	// Name of Prometheus external label used to denote the Prometheus instance
	// name. The external label will _not_ be added when the field is set to
	// the empty string (`""`).
	//
	// Default: "prometheus"
	// +optional
	PrometheusExternalLabelName *string `json:"prometheusExternalLabelName,omitempty"`

	// Log level for Prometheus and the config-reloader sidecar.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for Log level for Prometheus and the config-reloader sidecar.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`

	// Interval between consecutive scrapes.
	//
	// Default: "30s"
	// +kubebuilder:default:="30s"
	ScrapeInterval Duration `json:"scrapeInterval,omitempty"`
	// Number of seconds to wait until a scrape request times out.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`

	// The labels to add to any time series or alerts when communicating with
	// external systems (federation, remote storage, Alertmanager).
	// Labels defined by `spec.replicaExternalLabelName` and
	// `spec.prometheusExternalLabelName` take precedence over this list.
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`

	// Enable Prometheus to be used as a receiver for the Prometheus remote
	// write protocol.
	//
	// WARNING: This is not considered an efficient way of ingesting samples.
	// Use it with caution for specific low-volume use cases.
	// It is not suitable for replacing the ingestion via scraping and turning
	// Prometheus into a push-based metrics collection system.
	// For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver
	//
	// It requires Prometheus >= v2.33.0.
	EnableRemoteWriteReceiver bool `json:"enableRemoteWriteReceiver,omitempty"`

	// Enable access to Prometheus feature flags. By default, no features are enabled.
	//
	// Enabling features which are disabled by default is entirely outside the
	// scope of what the maintainers will support and by doing so, you accept
	// that this behaviour may break at any time without notice.
	//
	// For more information see https://prometheus.io/docs/prometheus/latest/feature_flags/
	EnableFeatures []string `json:"enableFeatures,omitempty"`

	// The external URL under which the Prometheus service is externally
	// available. This is necessary to generate correct URLs (for instance if
	// Prometheus is accessible behind an Ingress resource).
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Prometheus registers HTTP handlers for.
	//
	// This is useful when using `spec.externalURL`, and a proxy is rewriting
	// HTTP routes of a request, and the actual ExternalURL is still true, but
	// the server serves requests under a different route prefix. For example
	// for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`

	// Storage defines the storage used by Prometheus.
	Storage *StorageSpec `json:"storage,omitempty"`

	// Volumes allows the configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows the configuration of additional VolumeMounts.
	//
	// VolumeMounts will be appended to other VolumeMounts in the 'prometheus'
	// container, that are generated as a result of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// Defines the configuration of the Prometheus web server.
	Web *PrometheusWebSpec `json:"web,omitempty"`

	// Defines the resources requests and limits of the 'prometheus' container.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Defines on which Nodes the Pods are scheduled.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Secrets is a list of Secrets in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.
	// The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the 'prometheus' container.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.
	// The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the 'prometheus' container.
	ConfigMaps []string `json:"configMaps,omitempty"`

	// Defines the Pods' affinity scheduling rules if specified.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Defines the Pods' tolerations if specified.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// Defines the pod's topology spread constraints if specified.
	//+optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// Defines the list of remote write configurations.
	// +optional
	RemoteWrite []RemoteWriteSpec `json:"remoteWrite,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// When true, the Prometheus server listens on the loopback address
	// instead of the Pod IP's address.
	ListenLocal bool `json:"listenLocal,omitempty"`

	// Containers allows injecting additional containers or modifying operator
	// generated containers. This can be used to allow adding an authentication
	// proxy to the Pods or to change the behavior of an operator generated
	// container. Containers described here modify an operator generated
	// container if they share the same name and modifications are done via a
	// strategic merge patch.
	//
	// The names of containers managed by the operator are:
	// * `prometheus`
	// * `config-reloader`
	// * `thanos-sidecar`
	//
	// Overriding containers is entirely outside the scope of what the
	// maintainers will support and by doing so, you accept that this behaviour
	// may break at any time without notice.
	// +optional
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows injecting initContainers to the Pod definition. Those
	// can be used to e.g.  fetch secrets for injection into the Prometheus
	// configuration from external sources. Any errors during the execution of
	// an initContainer will lead to a restart of the Pod. More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// InitContainers described here modify an operator generated init
	// containers if they share the same name and modifications are done via a
	// strategic merge patch.
	//
	// The names of init container name managed by the operator are:
	// * `init-config-reloader`.
	//
	// Overriding init containers is entirely outside the scope of what the
	// maintainers will support and by doing so, you accept that this behaviour
	// may break at any time without notice.
	// +optional
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// AdditionalScrapeConfigs allows specifying a key of a Secret containing
	// additional Prometheus scrape configurations. Scrape configurations
	// specified are appended to the configurations generated by the Prometheus
	// Operator. Job configurations specified must have the form as specified
	// in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.
	// As scrape configs are appended, the user is responsible to make sure it
	// is valid. Note that using this feature may expose the possibility to
	// break upgrades of Prometheus. It is advised to review Prometheus release
	// notes to ensure that no incompatible scrape configs are going to break
	// Prometheus after the upgrade.
	// +optional
	AdditionalScrapeConfigs *v1.SecretKeySelector `json:"additionalScrapeConfigs,omitempty"`

	// APIServerConfig allows specifying a host and auth methods to access the
	// Kuberntees API server.
	// If null, Prometheus is assumed to run inside of the cluster: it will
	// discover the API servers automatically and use the Pod's CA certificate
	// and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	// +optional
	APIServerConfig *APIServerConfig `json:"apiserverConfig,omitempty"`

	// Priority class assigned to the Pods.
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// Port name used for the pods and governing service.
	// Default: "web"
	// +kubebuilder:default:="web"
	PortName string `json:"portName,omitempty"`

	// When true, ServiceMonitor, PodMonitor and Probe object are forbidden to
	// reference arbitrary files on the file system of the 'prometheus'
	// container.
	// When a ServiceMonitor's endpoint specifies a `bearerTokenFile` value
	// (e.g.  '/var/run/secrets/kubernetes.io/serviceaccount/token'), a
	// malicious target can get access to the Prometheus service account's
	// token in the Prometheus' scrape request. Setting
	// `spec.arbitraryFSAccessThroughSM` to 'true' would prevent the attack.
	// Users should instead provide the credentials using the
	// `spec.bearerTokenSecret` field.
	ArbitraryFSAccessThroughSMs ArbitraryFSAccessThroughSMsConfig `json:"arbitraryFSAccessThroughSMs,omitempty"`

	// When true, Prometheus resolves label conflicts by renaming the labels in
	// the scraped data to "exported_<label value>" for all targets created
	// from service and pod monitors.
	// Otherwise the HonorLabels field of the service or pod monitor applies.
	OverrideHonorLabels bool `json:"overrideHonorLabels,omitempty"`
	// When true, Prometheus ignores the timestamps for all the targets created
	// from service and pod monitors.
	// Otherwise the HonorTimestamps field of the service or pod monitor applies.
	OverrideHonorTimestamps bool `json:"overrideHonorTimestamps,omitempty"`

	// When true, `spec.namespaceSelector` from all PodMonitor, ServiceMonitor
	// and Probe objects will be ignored. They will only discover targets
	// within the namespace of the PodMonitor, ServiceMonitor and Probe
	// objec.
	IgnoreNamespaceSelectors bool `json:"ignoreNamespaceSelectors,omitempty"`

	// When not empty, a label will be added to
	//
	// 1. All metrics scraped from `ServiceMonitor`, `PodMonitor`, `Probe` and `ScrapeConfig` objects.
	// 2. All metrics generated from recording rules defined in `PrometheusRule` objects.
	// 3. All alerts generated from alerting rules defined in `PrometheusRule` objects.
	// 4. All vector selectors of PromQL expressions defined in `PrometheusRule` objects.
	//
	// The label will not added for objects referenced in `spec.excludedFromEnforcement`.
	//
	// The label's name is this field's value.
	// The label's value is the namespace of the `ServiceMonitor`,
	// `PodMonitor`, `Probe` or `PrometheusRule` object.
	EnforcedNamespaceLabel string `json:"enforcedNamespaceLabel,omitempty"`

	// When defined, enforcedSampleLimit specifies a global limit on the number
	// of scraped samples that will be accepted. This overrides any
	// `spec.sampleLimit` set by ServiceMonitor, PodMonitor, Probe objects
	// unless `spec.sampleLimit` is greater than zero and less than than
	// `spec.enforcedSampleLimit`.
	//
	// It is meant to be used by admins to keep the overall number of
	// samples/series under a desired limit.
	//
	// +optional
	EnforcedSampleLimit *uint64 `json:"enforcedSampleLimit,omitempty"`
	// When defined, enforcedTargetLimit specifies a global limit on the number
	// of scraped targets. The value overrides any `spec.targetLimit` set by
	// ServiceMonitor, PodMonitor, Probe objects unless `spec.targetLimit` is
	// greater than zero and less than `spec.enforcedTargetLimit`.
	//
	// It is meant to be used by admins to to keep the overall number of
	// targets under a desired limit.
	//
	// +optional
	EnforcedTargetLimit *uint64 `json:"enforcedTargetLimit,omitempty"`
	// When defined, enforcedLabelLimit specifies a global limit on the number
	// of labels per sample. The value overrides any `spec.labelLimit` set by
	// ServiceMonitor, PodMonitor, Probe objects unless `spec.labelLimit` is
	// greater than zero and less than `spec.enforcedLabelLimit`.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	EnforcedLabelLimit *uint64 `json:"enforcedLabelLimit,omitempty"`
	// When defined, enforcedLabelNameLengthLimit specifies a global limit on the length
	// of labels name per sample. The value overrides any `spec.labelNameLengthLimit` set by
	// ServiceMonitor, PodMonitor, Probe objects unless `spec.labelNameLengthLimit` is
	// greater than zero and less than `spec.enforcedLabelNameLengthLimit`.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	EnforcedLabelNameLengthLimit *uint64 `json:"enforcedLabelNameLengthLimit,omitempty"`
	// When not null, enforcedLabelValueLengthLimit defines a global limit on the length
	// of labels value per sample. The value overrides any `spec.labelValueLengthLimit` set by
	// ServiceMonitor, PodMonitor, Probe objects unless `spec.labelValueLengthLimit` is
	// greater than zero and less than `spec.enforcedLabelValueLengthLimit`.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	EnforcedLabelValueLengthLimit *uint64 `json:"enforcedLabelValueLengthLimit,omitempty"`
	// When defined, enforcedBodySizeLimit specifies a global limit on the size
	// of uncompressed response body that will be accepted by Prometheus.
	// Targets responding with a body larger than this many bytes will cause
	// the scrape to fail.
	//
	// It requires Prometheus >= v2.28.0.
	EnforcedBodySizeLimit ByteSize `json:"enforcedBodySizeLimit,omitempty"`

	// Minimum number of seconds for which a newly created Pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	//
	// This is an alpha field from kubernetes 1.22 until 1.24 which requires
	// enabling the StatefulSetMinReadySeconds feature gate.
	//
	// +optional
	MinReadySeconds *uint32 `json:"minReadySeconds,omitempty"`

	// Optional list of hosts and IPs that will be injected into the Pod's
	// hosts file if specified.
	//
	// +listType=map
	// +listMapKey=ip
	// +optional
	HostAliases []HostAlias `json:"hostAliases,omitempty"`

	// AdditionalArgs allows setting additional arguments for the 'prometheus' container.
	//
	// It is intended for e.g. activating hidden flags which are not supported by
	// the dedicated configuration options yet. The arguments are passed as-is to the
	// Prometheus container which may cause issues if they are invalid or not supported
	// by the given Prometheus version.
	//
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument, the reconciliation will
	// fail and an error will be logged.
	//
	// +optional
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`

	// Configures compression of the write-ahead log (WAL) using Snappy.
	//
	// WAL compression is enabled by default for Prometheus >= 2.20.0
	//
	// Requires Prometheus v2.11.0 and above.
	//
	// +optional
	WALCompression *bool `json:"walCompression,omitempty"`

	// List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
	// to be excluded from enforcing a namespace label of origin.
	//
	// It is only applicable if `spec.enforcedNamespaceLabel` set to true.
	//
	// +optional
	ExcludedFromEnforcement []ObjectReference `json:"excludedFromEnforcement,omitempty"`

	// Use the host's network namespace if true.
	//
	// Make sure to understand the security implications if you want to enable
	// it (https://kubernetes.io/docs/concepts/configuration/overview/).
	//
	// When hostNetwork is enabled, this will set the DNS policy to
	// `ClusterFirstWithHostNet` automatically.
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// PodTargetLabels are appended to the `spec.podTargetLabels` field of all
	// PodMonitor and ServiceMonitor objects.
	//
	// +optional
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`

	// EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
	// experimental feature, it may change in any upcoming release in a
	// breaking way.
	//
	// +optional
	TracingConfig *PrometheusTracingConfig `json:"tracingConfig,omitempty"`
	// BodySizeLimit defines per-scrape on response body size.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	BodySizeLimit *ByteSize `json:"bodySizeLimit,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	SampleLimit *uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	TargetLimit *uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	LabelLimit *uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	LabelNameLengthLimit *uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.45.0 and newer.
	//
	// +optional
	LabelValueLengthLimit *uint64 `json:"labelValueLengthLimit,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="prom"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Prometheus"
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.availableReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type == 'Reconciled')].status"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1
// +kubebuilder:subresource:status

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Prometheus cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec PrometheusSpec `json:"spec"`
	// Most recent observed status of the Prometheus cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status PrometheusStatus `json:"status,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *Prometheus) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusList is a list of Prometheuses.
// +k8s:openapi-gen=true
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []*Prometheus `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusSpec struct {
	CommonPrometheusFields `json:",inline"`

	// *Deprecated: use 'spec.image' instead.*
	BaseImage string `json:"baseImage,omitempty"`
	// *Deprecated: use 'spec.image' instead. The image's tag can be specified
	// as part of the image name.*
	Tag string `json:"tag,omitempty"`
	// *Deprecated: use 'spec.image' instead. The image's digest can be
	// specified as part of the image name.*
	SHA string `json:"sha,omitempty"`

	// How long to retain the Prometheus data.
	//
	// Default: "24h" if `spec.retention` and `spec.retentionSize` are empty.
	Retention Duration `json:"retention,omitempty"`
	// Maximum number of bytes used by the Prometheus data.
	RetentionSize ByteSize `json:"retentionSize,omitempty"`

	// When true, the Prometheus compaction is disabled.
	DisableCompaction bool `json:"disableCompaction,omitempty"`

	// Defines the configuration of the Prometheus rules' engine.
	Rules Rules `json:"rules,omitempty"`
	// Defines the list of PrometheusRule objects to which the namespace label
	// enforcement doesn't apply.
	// This is only relevant when `spec.enforcedNamespaceLabel` is set to true.
	// *Deprecated: use `spec.excludedFromEnforcement` instead.*
	// +optional
	PrometheusRulesExcludedFromEnforce []PrometheusRuleExcludeConfig `json:"prometheusRulesExcludedFromEnforce,omitempty"`
	// PrometheusRule objects to be selected for rule evaluation. An empty
	// label selector matches all objects. A null label selector matches no
	// objects.
	// +optional
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`
	// Namespaces to match for PrometheusRule discovery. An empty label selector
	// matches all namespaces. A null label selector matches the current
	// namespace only.
	// +optional
	RuleNamespaceSelector *metav1.LabelSelector `json:"ruleNamespaceSelector,omitempty"`

	// QuerySpec defines the configuration of the Promethus query service.
	// +optional
	Query *QuerySpec `json:"query,omitempty"`

	// Defines the settings related to Alertmanager.
	// +optional
	Alerting *AlertingSpec `json:"alerting,omitempty"`
	// AdditionalAlertRelabelConfigs specifies a key of a Secret containing
	// additional Prometheus alert relabel configurations. The alert relabel
	// configurations are appended to the configuration generated by the
	// Prometheus Operator. They must be formatted according to the official
	// Prometheus documentation:
	//
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs
	//
	// The user is responsible for making sure that the configurations are valid
	//
	// Note that using this feature may expose the possibility to break
	// upgrades of Prometheus. It is advised to review Prometheus release notes
	// to ensure that no incompatible alert relabel configs are going to break
	// Prometheus after the upgrade.
	// +optional
	AdditionalAlertRelabelConfigs *v1.SecretKeySelector `json:"additionalAlertRelabelConfigs,omitempty"`
	// AdditionalAlertManagerConfigs specifies a key of a Secret containing
	// additional Prometheus Alertmanager configurations. The Alertmanager
	// configurations are appended to the configuration generated by the
	// Prometheus Operator. They must be formatted according to the official
	// Prometheus documentation:
	//
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config
	//
	// The user is responsible for making sure that the configurations are valid
	//
	// Note that using this feature may expose the possibility to break
	// upgrades of Prometheus. It is advised to review Prometheus release notes
	// to ensure that no incompatible AlertManager configs are going to break
	// Prometheus after the upgrade.
	// +optional
	AdditionalAlertManagerConfigs *v1.SecretKeySelector `json:"additionalAlertManagerConfigs,omitempty"`

	// Defines the list of remote read configurations.
	// +optional
	RemoteRead []RemoteReadSpec `json:"remoteRead,omitempty"`

	// Defines the configuration of the optional Thanos sidecar.
	//
	// This section is experimental, it may change significantly without
	// deprecation notice in any release.
	// +optional
	Thanos *ThanosSpec `json:"thanos,omitempty"`

	// queryLogFile specifies where the file to which PromQL queries are logged.
	//
	// If the filename has an empty path, e.g. 'query.log', The Prometheus Pods
	// will mount the file into an emptyDir volume at `/var/log/prometheus`.
	// If a full path is provided, e.g. '/var/log/prometheus/query.log', you
	// must mount a volume in the specified directory and it must be writable.
	// This is because the prometheus container runs with a read-only root
	// filesystem for security reasons.
	// Alternatively, the location can be set to a standard I/O stream, e.g.
	// `/dev/stdout`, to log query information to the default Prometheus log
	// stream.
	QueryLogFile string `json:"queryLogFile,omitempty"`

	// AllowOverlappingBlocks enables vertical compaction and vertical query
	// merge in Prometheus.
	//
	// *Deprecated: this flag has no effect for Prometheus >= 2.39.0 where overlapping blocks are enabled by default.*
	AllowOverlappingBlocks bool `json:"allowOverlappingBlocks,omitempty"`

	// Exemplars related settings that are runtime reloadable.
	// It requires to enable the `exemplar-storage` feature flag to be effective.
	// +optional
	Exemplars *Exemplars `json:"exemplars,omitempty"`

	// Interval between rule evaluations.
	// Default: "30s"
	// +kubebuilder:default:="30s"
	EvaluationInterval Duration `json:"evaluationInterval,omitempty"`

	// Enables access to the Prometheus web admin API.
	//
	// WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
	// shutdown Prometheus, and more. Enabling this should be done with care and the
	// user is advised to add additional authentication authorization via a proxy to
	// ensure only clients authorized to perform these actions can do so.
	//
	// For more information:
	// https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis
	EnableAdminAPI bool `json:"enableAdminAPI,omitempty"`

	// Defines the runtime reloadable configuration of the timeseries database
	// (TSDB).
	TSDB TSDBSpec `json:"tsdb,omitempty"`
}

type PrometheusTracingConfig struct {
	// Client used to export the traces. Supported values are `http` or `grpc`.
	//+kubebuilder:validation:Enum=http;grpc
	// +optional
	ClientType *string `json:"clientType"`

	// Endpoint to send the traces to. Should be provided in format <host>:<port>.
	// +kubebuilder:validation:MinLength:=1
	// +required
	Endpoint string `json:"endpoint"`

	// Sets the probability a given trace will be sampled. Must be a float from 0 through 1.
	// +optional
	SamplingFraction *resource.Quantity `json:"samplingFraction"`

	// If disabled, the client will use a secure connection.
	// +optional
	Insecure *bool `json:"insecure"`

	// Key-value pairs to be used as headers associated with gRPC or HTTP requests.
	// +optional
	Headers map[string]string `json:"headers"`

	// Compression key for supported compression types. The only supported value is `gzip`.
	//+kubebuilder:validation:Enum=gzip
	// +optional
	Compression *string `json:"compression"`

	// Maximum time the exporter will wait for each batch export.
	// +optional
	Timeout *Duration `json:"timeout"`

	// TLS Config to use when sending traces.
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig"`
}

// PrometheusStatus is the most recent observed status of the Prometheus cluster.
// More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusStatus struct {
	// Represents whether any actions on the underlying managed objects are
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
	// The current state of the Prometheus deployment.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// The list has one entry per shard. Each entry provides a summary of the shard status.
	// +listType=map
	// +listMapKey=shardID
	// +optional
	ShardStatuses []ShardStatus `json:"shardStatuses,omitempty"`
}

// AlertingSpec defines parameters for alerting configuration of Prometheus servers.
// +k8s:openapi-gen=true
type AlertingSpec struct {
	// AlertmanagerEndpoints Prometheus should fire alerts against.
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
// If no storage option is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.
//
// If multiple storage options are specified, priority will be given as follows:
//  1. emptyDir
//  2. ephemeral
//  3. volumeClaimTemplate
//
// +k8s:openapi-gen=true
type StorageSpec struct {
	// *Deprecated: subPath usage will be removed in a future release.*
	DisableMountSubPath bool `json:"disableMountSubPath,omitempty"`
	// EmptyDirVolumeSource to be used by the StatefulSet.
	// If specified, it takes precedence over `ephemeral` and `volumeClaimTemplate`.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// EphemeralVolumeSource to be used by the StatefulSet.
	// This is a beta field in k8s 1.21 and GA in 1.15.
	// For lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.
	// More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes
	Ephemeral *v1.EphemeralVolumeSource `json:"ephemeral,omitempty"`
	// Defines the PVC spec to be used by the Prometheus StatefulSets.
	// The easiest way to use a volume that cannot be automatically provisioned
	// is to use a label selector alongside manually created PersistentVolumes.
	VolumeClaimTemplate EmbeddedPersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// QuerySpec defines the query command line flags when starting Prometheus.
// +k8s:openapi-gen=true
type QuerySpec struct {
	// The delta difference allowed for retrieving metrics during expression evaluations.
	// +optional
	LookbackDelta *string `json:"lookbackDelta,omitempty"`
	// Number of concurrent queries that can be run at once.
	// +kubebuilder:validation:Minimum:=1
	// +optional
	MaxConcurrency *int32 `json:"maxConcurrency,omitempty"`
	// Maximum number of samples a single query can load into memory. Note that
	// queries will fail if they would load more samples than this into memory,
	// so this also limits the number of samples a query can return.
	// +optional
	MaxSamples *int32 `json:"maxSamples,omitempty"`
	// Maximum time a query may take before being aborted.
	// +optional
	Timeout *Duration `json:"timeout,omitempty"`
}

// PrometheusWebSpec defines the configuration of the Prometheus web server.
// +k8s:openapi-gen=true
type PrometheusWebSpec struct {
	WebConfigFileFields `json:",inline"`

	// The prometheus web page title.
	// +optional
	PageTitle *string `json:"pageTitle,omitempty"`

	// Defines the maximum number of simultaneous connections
	// A zero value means that Prometheus doesn't accept any incoming connection.
	// +kubebuilder:validation:Minimum:=0
	// +optional
	MaxConnections *int32 `json:"maxConnections,omitempty"`
}

// ThanosSpec defines the configuration of the Thanos sidecar.
// +k8s:openapi-gen=true
type ThanosSpec struct {
	// Container image name for Thanos. If specified, it takes precedence over
	// the `spec.thanos.baseImage`, `spec.thanos.tag` and `spec.thanos.sha`
	// fields.
	//
	// Specifying `spec.thanos.version` is still necessary to ensure the
	// Prometheus Operator knows which version of Thanos is being configured.
	//
	// If neither `spec.thanos.image` nor `spec.thanos.baseImage` are defined,
	// the operator will use the latest upstream version of Thanos available at
	// the time when the operator was released.
	//
	// +optional
	Image *string `json:"image,omitempty"`

	// Version of Thanos being deployed. The operator uses this information
	// to generate the Prometheus StatefulSet + configuration files.
	//
	// If not specified, the operator assumes the latest upstream release of
	// Thanos available at the time when the version of the operator was
	// released.
	//
	// +optional
	Version *string `json:"version,omitempty"`

	// *Deprecated: use 'image' instead. The image's tag can be specified as
	// part of the image name.*
	// +optional
	Tag *string `json:"tag,omitempty"`
	// *Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image name.*
	// +optional
	SHA *string `json:"sha,omitempty"`
	// *Deprecated: use 'image' instead.*
	// +optional
	BaseImage *string `json:"baseImage,omitempty"`

	// Defines the resources requests and limits of the Thanos sidecar.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Defines the Thanos sidecar's configuration to upload TSDB blocks to object storage.
	//
	// More info: https://thanos.io/tip/thanos/storage.md/
	//
	// objectStorageConfigFile takes precedence over this field.
	// +optional
	ObjectStorageConfig *v1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
	// Defines the Thanos sidecar's configuration file to upload TSDB blocks to object storage.
	//
	// More info: https://thanos.io/tip/thanos/storage.md/
	//
	// This field takes precedence over objectStorageConfig.
	// +optional
	ObjectStorageConfigFile *string `json:"objectStorageConfigFile,omitempty"`

	// *Deprecated: use `grpcListenLocal` and `httpListenLocal` instead.*
	ListenLocal bool `json:"listenLocal,omitempty"`

	// When true, the Thanos sidecar listens on the loopback interface instead
	// of the Pod IP's address for the gRPC endpoints.
	//
	// It has no effect if `listenLocal` is true.
	GRPCListenLocal bool `json:"grpcListenLocal,omitempty"`

	// When true, the Thanos sidecar listens on the loopback interface instead
	// of the Pod IP's address for the HTTP endpoints.
	//
	// It has no effect if `listenLocal` is true.
	HTTPListenLocal bool `json:"httpListenLocal,omitempty"`

	// Defines the tracing configuration for the Thanos sidecar.
	//
	// More info: https://thanos.io/tip/thanos/tracing.md/
	//
	// This is an experimental feature, it may change in any upcoming release
	// in a breaking way.
	//
	// tracingConfigFile takes precedence over this field.
	// +optional
	TracingConfig *v1.SecretKeySelector `json:"tracingConfig,omitempty"`
	// Defines the tracing configuration file for the Thanos sidecar.
	//
	// More info: https://thanos.io/tip/thanos/tracing.md/
	//
	// This is an experimental feature, it may change in any upcoming release
	// in a breaking way.
	//
	// This field takes precedence over tracingConfig.
	TracingConfigFile string `json:"tracingConfigFile,omitempty"`

	// Configures the TLS parameters for the gRPC server providing the StoreAPI.
	//
	// Note: Currently only the `caFile`, `certFile`, and `keyFile` fields are supported.
	//
	// +optional
	GRPCServerTLSConfig *TLSConfig `json:"grpcServerTlsConfig,omitempty"`

	// Log level for the Thanos sidecar.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for the Thanos sidecar.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`

	// Defines the start of time range limit served by the Thanos sidecar's StoreAPI.
	// The field's value should be a constant time in RFC3339 format or a time
	// duration relative to current time, such as -1d or 2h45m. Valid duration
	// units are ms, s, m, h, d, w, y.
	MinTime string `json:"minTime,omitempty"`

	// BlockDuration controls the size of TSDB blocks produced by Prometheus.
	// The default value is 2h to match the upstream Prometheus defaults.
	//
	// WARNING: Changing the block duration can impact the performance and
	// efficiency of the entire Prometheus/Thanos stack due to how it interacts
	// with memory and Thanos compactors. It is recommended to keep this value
	// set to a multiple of 120 times your longest scrape or rule interval. For
	// example, 30s * 120 = 1h.
	//
	// +kubebuilder:default:="2h"
	BlockDuration Duration `json:"blockSize,omitempty"`

	// ReadyTimeout is the maximum time that the Thanos sidecar will wait for
	// Prometheus to start.
	ReadyTimeout Duration `json:"readyTimeout,omitempty"`
	// How often to retrieve the Prometheus configuration.
	GetConfigInterval Duration `json:"getConfigInterval,omitempty"`
	// Maximum time to wait when retrieving the Prometheus configuration.
	GetConfigTimeout Duration `json:"getConfigTimeout,omitempty"`

	// VolumeMounts allows configuration of additional VolumeMounts for Thanos.
	// VolumeMounts specified will be appended to other VolumeMounts in the
	// 'thanos-sidecar' container.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// AdditionalArgs allows setting additional arguments for the Thanos container.
	// The arguments are passed as-is to the Thanos container which may cause issues
	// if they are invalid or not supported the given Thanos version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument, the reconciliation will
	// fail and an error will be logged.
	// +optional
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`
}

// RemoteWriteSpec defines the configuration to write samples from Prometheus
// to a remote endpoint.
// +k8s:openapi-gen=true
type RemoteWriteSpec struct {
	// The URL of the endpoint to send samples to.
	URL string `json:"url"`

	// The name of the remote write queue, it must be unique if specified. The
	// name is used in metrics and logging in order to differentiate queues.
	//
	// It requires Prometheus >= v2.15.0.
	//
	Name string `json:"name,omitempty"`

	// Enables sending of exemplars over remote write. Note that
	// exemplar-storage itself must be enabled using the `spec.enableFeature`
	// option for exemplars to be scraped in the first place.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// +optional
	SendExemplars *bool `json:"sendExemplars,omitempty"`

	// Enables sending of native histograms, also known as sparse histograms
	// over remote write.
	//
	// It requires Prometheus >= v2.40.0.
	//
	// +optional
	SendNativeHistograms *bool `json:"sendNativeHistograms,omitempty"`

	// Timeout for requests to the remote write endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`

	// Custom HTTP headers to be sent along with each remote write request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	//
	// It requires Prometheus >= v2.25.0.
	//
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// The list of remote write relabel configurations.
	// +optional
	WriteRelabelConfigs []RelabelConfig `json:"writeRelabelConfigs,omitempty"`

	// OAuth2 configuration for the URL.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// Cannot be set at the same time as `sigv4`, `authorization`, or `basicAuth`.
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// BasicAuth configuration for the URL.
	//
	// Cannot be set at the same time as `sigv4`, `authorization`, or `oauth2`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// File from which to read bearer token for the URL.
	//
	// *Deprecated: this will be removed in a future release. Prefer using `authorization`.*
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for the URL.
	//
	// It requires Prometheus >= v2.26.0.
	//
	// Cannot be set at the same time as `sigv4`, `basicAuth`, or `oauth2`.
	//
	// +optional
	Authorization *Authorization `json:"authorization,omitempty"`
	// Sigv4 allows to configures AWS's Signature Verification 4 for the URL.
	//
	// It requires Prometheus >= v2.26.0.
	//
	// Cannot be set at the same time as `authorization`, `basicAuth`, or `oauth2`.
	//
	// +optional
	Sigv4 *Sigv4 `json:"sigv4,omitempty"`

	// *Warning: this field shouldn't be used because the token value appears
	// in clear-text. Prefer using `authorization`.*
	//
	// *Deprecated: this will be removed in a future release.*
	BearerToken string `json:"bearerToken,omitempty"`

	// TLS Config to use for the URL.
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`

	// QueueConfig allows tuning of the remote write queue parameters.
	// +optional
	QueueConfig *QueueConfig `json:"queueConfig,omitempty"`

	// MetadataConfig configures the sending of series metadata to the remote storage.
	// +optional
	MetadataConfig *MetadataConfig `json:"metadataConfig,omitempty"`
}

// QueueConfig allows the tuning of remote write's queue_config parameters.
// This object is referenced in the RemoteWriteSpec object.
// +k8s:openapi-gen=true
type QueueConfig struct {
	// Capacity is the number of samples to buffer per shard before we start
	// dropping them.
	Capacity int `json:"capacity,omitempty"`
	// MinShards is the minimum number of shards, i.e. amount of concurrency.
	MinShards int `json:"minShards,omitempty"`
	// MaxShards is the maximum number of shards, i.e. amount of concurrency.
	MaxShards int `json:"maxShards,omitempty"`
	// MaxSamplesPerSend is the maximum number of samples per send.
	MaxSamplesPerSend int `json:"maxSamplesPerSend,omitempty"`
	// BatchSendDeadline is the maximum time a sample will wait in buffer.
	BatchSendDeadline string `json:"batchSendDeadline,omitempty"`
	// MaxRetries is the maximum number of times to retry a batch on recoverable errors.
	MaxRetries int `json:"maxRetries,omitempty"`
	// MinBackoff is the initial retry delay. Gets doubled for every retry.
	MinBackoff string `json:"minBackoff,omitempty"`
	// MaxBackoff is the maximum retry delay.
	MaxBackoff string `json:"maxBackoff,omitempty"`
	// Retry upon receiving a 429 status code from the remote-write storage.
	// This is experimental feature and might change in the future.
	RetryOnRateLimit bool `json:"retryOnRateLimit,omitempty"`
}

// Sigv4 optionally configures AWS's Signature Verification 4 signing process to
// sign requests.
// +k8s:openapi-gen=true
type Sigv4 struct {
	// Region is the AWS region. If blank, the region from the default credentials chain used.
	Region string `json:"region,omitempty"`
	// AccessKey is the AWS API key. If not specified, the environment variable
	// `AWS_ACCESS_KEY_ID` is used.
	// +optional
	AccessKey *v1.SecretKeySelector `json:"accessKey,omitempty"`
	// SecretKey is the AWS API secret. If not specified, the environment
	// variable `AWS_SECRET_ACCESS_KEY` is used.
	// +optional
	SecretKey *v1.SecretKeySelector `json:"secretKey,omitempty"`
	// Profile is the named AWS profile used to authenticate.
	Profile string `json:"profile,omitempty"`
	// RoleArn is the named AWS profile used to authenticate.
	RoleArn string `json:"roleArn,omitempty"`
}

// RemoteReadSpec defines the configuration for Prometheus to read back samples
// from a remote endpoint.
// +k8s:openapi-gen=true
type RemoteReadSpec struct {
	// The URL of the endpoint to query from.
	URL string `json:"url"`

	// The name of the remote read queue, it must be unique if specified. The
	// name is used in metrics and logging in order to differentiate read
	// configurations.
	//
	// It requires Prometheus >= v2.15.0.
	//
	Name string `json:"name,omitempty"`

	// An optional list of equality matchers which have to be present
	// in a selector to query the remote read endpoint.
	// +optional
	RequiredMatchers map[string]string `json:"requiredMatchers,omitempty"`

	// Timeout for requests to the remote read endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`

	// Custom HTTP headers to be sent along with each remote read request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	// Only valid in Prometheus versions 2.26.0 and newer.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// Whether reads should be made for queries for time ranges that
	// the local storage should have complete data for.
	ReadRecent bool `json:"readRecent,omitempty"`

	// OAuth2 configuration for the URL.
	//
	// It requires Prometheus >= v2.27.0.
	//
	// Cannot be set at the same time as `authorization`, or `basicAuth`.
	//
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// BasicAuth configuration for the URL.
	//
	// Cannot be set at the same time as `authorization`, or `oauth2`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// File from which to read the bearer token for the URL.
	//
	// *Deprecated: this will be removed in a future release. Prefer using `authorization`.*
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for the URL.
	//
	// It requires Prometheus >= v2.26.0.
	//
	// Cannot be set at the same time as `basicAuth`, or `oauth2`.
	//
	// +optional
	Authorization *Authorization `json:"authorization,omitempty"`

	// *Warning: this field shouldn't be used because the token value appears
	// in clear-text. Prefer using `authorization`.*
	//
	// *Deprecated: this will be removed in a future release.*
	BearerToken string `json:"bearerToken,omitempty"`

	// TLS Config to use for the URL.
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`

	// Configure whether HTTP requests follow HTTP 3xx redirects.
	//
	// It requires Prometheus >= v2.26.0.
	//
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`

	// Whether to use the external labels as selectors for the remote read endpoint.
	//
	// It requires Prometheus >= v2.34.0.
	//
	// +optional
	FilterExternalLabels *bool `json:"filterExternalLabels,omitempty"`
}

// RelabelConfig allows dynamic rewriting of the label set for targets, alerts,
// scraped samples and remote write samples.
//
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
//
// +k8s:openapi-gen=true
type RelabelConfig struct {
	// The source labels select values from existing labels. Their content is
	// concatenated using the configured Separator and matched against the
	// configured regular expression.
	//
	// +optional
	SourceLabels []LabelName `json:"sourceLabels,omitempty"`

	// Separator is the string between concatenated SourceLabels.
	Separator string `json:"separator,omitempty"`

	// Label to which the resulting string is written in a replacement.
	//
	// It is mandatory for `Replace`, `HashMod`, `Lowercase`, `Uppercase`,
	// `KeepEqual` and `DropEqual` actions.
	//
	// Regex capture groups are available.
	TargetLabel string `json:"targetLabel,omitempty"`

	// Regular expression against which the extracted value is matched.
	Regex string `json:"regex,omitempty"`

	// Modulus to take of the hash of the source label values.
	//
	// Only applicable when the action is `HashMod`.
	Modulus uint64 `json:"modulus,omitempty"`

	// Replacement value against which a Replace action is performed if the
	// regular expression matches.
	//
	// Regex capture groups are available.
	Replacement string `json:"replacement,omitempty"`

	// Action to perform based on the regex matching.
	//
	// `Uppercase` and `Lowercase` actions require Prometheus >= v2.36.0.
	// `DropEqual` and `KeepEqual` actions require Prometheus >= v2.41.0.
	//
	// Default: "Replace"
	//
	// +kubebuilder:validation:Enum=replace;Replace;keep;Keep;drop;Drop;hashmod;HashMod;labelmap;LabelMap;labeldrop;LabelDrop;labelkeep;LabelKeep;lowercase;Lowercase;uppercase;Uppercase;keepequal;KeepEqual;dropequal;DropEqual
	// +kubebuilder:default=replace
	Action string `json:"action,omitempty"`
}

// APIServerConfig defines how the Prometheus server connects to the Kubernetes API server.
//
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
//
// +k8s:openapi-gen=true
type APIServerConfig struct {
	// Kubernetes API address consisting of a hostname or IP address followed
	// by an optional port number.
	Host string `json:"host"`

	// BasicAuth configuration for the API server.
	//
	// Cannot be set at the same time as `authorization`, `bearerToken`, or
	// `bearerTokenFile`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// File to read bearer token for accessing apiserver.
	//
	// Cannot be set at the same time as `basicAuth`, `authorization`, or `bearerToken`.
	//
	// *Deprecated: this will be removed in a future release. Prefer using `authorization`.*
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`

	// TLS Config to use for the API server.
	//
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// Authorization section for the API server.
	//
	// Cannot be set at the same time as `basicAuth`, `bearerToken`, or
	// `bearerTokenFile`.
	//
	// +optional
	Authorization *Authorization `json:"authorization,omitempty"`

	// *Warning: this field shouldn't be used because the token value appears
	// in clear-text. Prefer using `authorization`.*
	//
	// *Deprecated: this will be removed in a future release.*
	BearerToken string `json:"bearerToken,omitempty"`
}

// AlertmanagerEndpoints defines a selection of a single Endpoints object
// containing Alertmanager IPs to fire alerts against.
// +k8s:openapi-gen=true
type AlertmanagerEndpoints struct {
	// Namespace of the Endpoints object.
	Namespace string `json:"namespace"`
	// Name of the Endpoints object in the namespace.
	Name string `json:"name"`

	// Port on which the Alertmanager API is exposed.
	Port intstr.IntOrString `json:"port"`

	// Scheme to use when firing alerts.
	Scheme string `json:"scheme,omitempty"`

	// Prefix for the HTTP path alerts are pushed to.
	PathPrefix string `json:"pathPrefix,omitempty"`

	// TLS Config to use for Alertmanager.
	//
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// BasicAuth configuration for Alertmanager.
	//
	// Cannot be set at the same time as `bearerTokenFile`, or `authorization`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// File to read bearer token for Alertmanager.
	//
	// Cannot be set at the same time as `basicAuth`, or `authorization`.
	//
	// *Deprecated: this will be removed in a future release. Prefer using `authorization`.*
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`

	// Authorization section for Alertmanager.
	//
	// Cannot be set at the same time as `basicAuth`, or `bearerTokenFile`.
	//
	// +optional
	Authorization *SafeAuthorization `json:"authorization,omitempty"`

	// Version of the Alertmanager API that Prometheus uses to send alerts.
	// It can be "v1" or "v2".
	APIVersion string `json:"apiVersion,omitempty"`

	// Timeout is a per-target Alertmanager timeout when pushing alerts.
	//
	// +optional
	Timeout *Duration `json:"timeout,omitempty"`

	// Whether to enable HTTP2.
	//
	// +optional
	EnableHttp2 *bool `json:"enableHttp2,omitempty"`
}

// +k8s:openapi-gen=true
type Rules struct {
	// Defines the parameters of the Prometheus rules' engine.
	//
	// Any update to these parameters trigger a restart of the pods.
	Alert RulesAlert `json:"alert,omitempty"`
}

// +k8s:openapi-gen=true
type RulesAlert struct {
	// Max time to tolerate prometheus outage for restoring 'for' state of
	// alert.
	ForOutageTolerance string `json:"forOutageTolerance,omitempty"`

	// Minimum duration between alert and restored 'for' state.
	//
	// This is maintained only for alerts with a configured 'for' time greater
	// than the grace period.
	ForGracePeriod string `json:"forGracePeriod,omitempty"`

	// Minimum amount of time to wait before resending an alert to
	// Alertmanager.
	ResendDelay string `json:"resendDelay,omitempty"`
}

// MetadataConfig configures the sending of series metadata to the remote storage.
//
// +k8s:openapi-gen=true
type MetadataConfig struct {
	// Defines whether metric metadata is sent to the remote storage or not.
	Send bool `json:"send,omitempty"`

	// Defines how frequently metric metadata is sent to the remote storage.
	SendInterval Duration `json:"sendInterval,omitempty"`
}

type ShardStatus struct {
	// Identifier of the shard.
	// +required
	ShardID string `json:"shardID"`
	// Total number of pods targeted by this shard.
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this shard
	// that have the desired spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this shard.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this shard.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

type TSDBSpec struct {
	// Configures how old an out-of-order/out-of-bounds sample can be with
	// respect to the TSDB max time.
	//
	// An out-of-order/out-of-bounds sample is ingested into the TSDB as long as
	// the timestamp of the sample is >= (TSDB.MaxTime - outOfOrderTimeWindow).
	//
	// Out of order ingestion is an experimental feature.
	//
	// It requires Prometheus >= v2.39.0.
	OutOfOrderTimeWindow Duration `json:"outOfOrderTimeWindow,omitempty"`
}

type Exemplars struct {
	// Maximum number of exemplars stored in memory for all series.
	//
	// exemplar-storage itself must be enabled using the `spec.enableFeature`
	// option for exemplars to be scraped in the first place.
	//
	// If not set, Prometheus uses its default value. A value of zero or less
	// than zero disables the storage.
	//
	// +optional
	MaxSize *int64 `json:"maxSize,omitempty"`
}

// SafeAuthorization specifies a subset of the Authorization struct, that is
// safe for use because it doesn't provide access to the Prometheus container's
// filesystem.
//
// +k8s:openapi-gen=true
type SafeAuthorization struct {
	// Defines the authentication type. The value is case-insensitive.
	//
	// "Basic" is not a supported value.
	//
	// Default: "Bearer"
	Type string `json:"type,omitempty"`

	// Selects a key of a Secret in the namespace that contains the credentials for authentication.
	Credentials *v1.SecretKeySelector `json:"credentials,omitempty"`
}

// Validate semantically validates the given Authorization section.
func (c *SafeAuthorization) Validate() error {
	if c == nil {
		return nil
	}

	if strings.ToLower(strings.TrimSpace(c.Type)) == "basic" {
		return &AuthorizationValidationError{`Authorization type cannot be set to "basic", use "basic_auth" instead`}
	}
	if c.Credentials == nil {
		return &AuthorizationValidationError{"Authorization credentials are required"}
	}
	return nil
}

type Authorization struct {
	SafeAuthorization `json:",inline"`

	// File to read a secret from, mutually exclusive with `credentials`.
	CredentialsFile string `json:"credentialsFile,omitempty"`
}

// Validate semantically validates the given Authorization section.
func (c *Authorization) Validate() error {
	if c.Credentials != nil && c.CredentialsFile != "" {
		return &AuthorizationValidationError{"Authorization can not specify both Credentials and CredentialsFile"}
	}
	if strings.ToLower(strings.TrimSpace(c.Type)) == "basic" {
		return &AuthorizationValidationError{"Authorization type cannot be set to \"basic\", use \"basic_auth\" instead"}
	}
	return nil
}

// AuthorizationValidationError is returned by Authorization.Validate()
// on semantically invalid configurations.
// +k8s:openapi-gen=false
type AuthorizationValidationError struct {
	err string
}

func (e *AuthorizationValidationError) Error() string {
	return e.err
}
