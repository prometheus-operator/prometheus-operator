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
	"fmt"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	Version = "v1"
)

var resourceToKind = map[string]string{
	PrometheusName:     PrometheusesKind,
	AlertmanagerName:   AlertmanagersKind,
	ServiceMonitorName: ServiceMonitorsKind,
	PodMonitorName:     PodMonitorsKind,
	PrometheusRuleName: PrometheusRuleKind,
	ProbeName:          ProbesKind,
}

// ByteSize is a valid memory size type based on powers-of-2, so 1KB is 1024B.
// Supported units: B, KB, KiB, MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB Ex: `512MB`.
// +kubebuilder:validation:Pattern:="(^0|([0-9]*[.])?[0-9]+((K|M|G|T|E|P)i?)?B)$"
type ByteSize string

// Duration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
// Supported units: y, w, d, h, m, s, ms
// Examples: `30s`, `1m`, `1h20m15s`, `15d`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
type Duration string

// NonEmptyDuration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
// Compared to Duration,  NonEmptyDuration enforces a minimum length of 1.
// Supported units: y, w, d, h, m, s, ms
// Examples: `30s`, `1m`, `1h20m15s`, `15d`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
// +kubebuilder:validation:MinLength=1
type NonEmptyDuration string

// GoDuration is a valid time duration that can be parsed by Go's time.ParseDuration() function.
// Supported units: h, m, s, ms
// Examples: `45ms`, `30s`, `1m`, `1h20m15s`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
type GoDuration string

// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
// pod's hosts file.
type HostAlias struct {
	// IP address of the host file entry.
	// +kubebuilder:validation:Required
	IP string `json:"ip"`
	// Hostnames for the above IP address.
	// +kubebuilder:validation:Required
	Hostnames []string `json:"hostnames"`
}

// PrometheusRuleExcludeConfig enables users to configure excluded
// PrometheusRule names and their namespaces to be ignored while enforcing
// namespace label for alerts and metrics.
type PrometheusRuleExcludeConfig struct {
	// Namespace of the excluded PrometheusRule object.
	RuleNamespace string `json:"ruleNamespace"`
	// Name of the excluded PrometheusRule object.
	RuleName string `json:"ruleName"`
}

// ObjectReference references a PodMonitor, ServiceMonitor, Probe or PrometheusRule object.
type ObjectReference struct {
	// Group of the referent. When not specified, it defaults to `monitoring.coreos.com`
	// +optional
	// +kubebuilder:default:="monitoring.coreos.com"
	// +kubebuilder:validation:Enum=monitoring.coreos.com
	Group string `json:"group"`
	// Resource of the referent.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=prometheusrules;servicemonitors;podmonitors;probes
	Resource string `json:"resource"`
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`
	// Name of the referent. When not set, all resources in the namespace are matched.
	// +optional
	Name string `json:"name,omitempty"`
}

func (obj *ObjectReference) GroupResource() schema.GroupResource {
	return schema.GroupResource{
		Resource: obj.Resource,
		Group:    obj.getGroup(),
	}
}

func (obj *ObjectReference) GroupKind() schema.GroupKind {
	_, found := resourceToKind[obj.Resource]
	if !found {
		panic(fmt.Sprintf("failed to map resource %q to a kind", obj.Resource))
	}
	return schema.GroupKind{
		Kind:  resourceToKind[obj.Resource],
		Group: obj.getGroup(),
	}
}

// getGroup returns the group of the object.
// It is mostly needed for tests which don't create objects through the API and don't benefit from the default value.
func (obj *ObjectReference) getGroup() string {
	if obj.Group == "" {
		return monitoring.GroupName
	}
	return obj.Group
}

// ArbitraryFSAccessThroughSMsConfig enables users to configure, whether
// a service monitor selected by the Prometheus instance is allowed to use
// arbitrary files on the file system of the Prometheus container. This is the case
// when e.g. a service monitor specifies a BearerTokenFile in an endpoint. A
// malicious user could create a service monitor selecting arbitrary secret files
// in the Prometheus container. Those secrets would then be sent with a scrape
// request by Prometheus to a malicious target. Denying the above would prevent the
// attack, users can instead use the BearerTokenSecret field.
type ArbitraryFSAccessThroughSMsConfig struct {
	Deny bool `json:"deny,omitempty"`
}

// Condition represents the state of the resources associated with the
// Prometheus, Alertmanager or ThanosRuler resource.
// +k8s:deepcopy-gen=true
type Condition struct {
	// Type of the condition being reported.
	// +required
	Type ConditionType `json:"type"`
	// Status of the condition.
	// +required
	Status ConditionStatus `json:"status"`
	// lastTransitionTime is the time of the last update to the current status property.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details for the condition's last transition.
	// +optional
	Message string `json:"message,omitempty"`
	// ObservedGeneration represents the .metadata.generation that the
	// condition was set based upon. For instance, if `.metadata.generation` is
	// currently 12, but the `.status.conditions[].observedGeneration` is 9, the
	// condition is out of date with respect to the current state of the
	// instance.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type ConditionType string

const (
	// Available indicates whether enough pods are ready to provide the
	// service.
	// The possible status values for this condition type are:
	// - True: all pods are running and ready, the service is fully available.
	// - Degraded: some pods aren't ready, the service is partially available.
	// - False: no pods are running, the service is totally unavailable.
	// - Unknown: the operator couldn't determine the condition status.
	Available ConditionType = "Available"
	// Reconciled indicates whether the operator has reconciled the state of
	// the underlying resources with the object's spec.
	// The possible status values for this condition type are:
	// - True: the reconciliation was successful.
	// - False: the reconciliation failed.
	// - Unknown: the operator couldn't determine the condition status.
	Reconciled ConditionType = "Reconciled"
)

type ConditionStatus string

const (
	ConditionTrue     ConditionStatus = "True"
	ConditionDegraded ConditionStatus = "Degraded"
	ConditionFalse    ConditionStatus = "False"
	ConditionUnknown  ConditionStatus = "Unknown"
)

// EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
// It contains TypeMeta and a reduced ObjectMeta.
type EmbeddedPersistentVolumeClaim struct {
	metav1.TypeMeta `json:",inline"`

	// EmbeddedMetadata contains metadata relevant to an EmbeddedResource.
	EmbeddedObjectMetadata `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Defines the desired characteristics of a volume requested by a pod author.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec v1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// *Deprecated: this field is never set.*
	// +optional
	Status v1.PersistentVolumeClaimStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
// Only fields which are relevant to embedded resources are included.
type EmbeddedObjectMetadata struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// WebConfigFileFields defines the file content for --web.config.file flag.
// +k8s:deepcopy-gen=true
type WebConfigFileFields struct {
	// Defines the TLS parameters for HTTPS.
	TLSConfig *WebTLSConfig `json:"tlsConfig,omitempty"`
	// Defines HTTP parameters for web server.
	HTTPConfig *WebHTTPConfig `json:"httpConfig,omitempty"`
}

// WebHTTPConfig defines HTTP parameters for web server.
// +k8s:openapi-gen=true
type WebHTTPConfig struct {
	// Enable HTTP/2 support. Note that HTTP/2 is only supported with TLS.
	// When TLSConfig is not configured, HTTP/2 will be disabled.
	// Whenever the value of the field changes, a rolling update will be triggered.
	HTTP2 *bool `json:"http2,omitempty"`
	// List of headers that can be added to HTTP responses.
	Headers *WebHTTPHeaders `json:"headers,omitempty"`
}

// WebHTTPHeaders defines the list of headers that can be added to HTTP responses.
// +k8s:openapi-gen=true
type WebHTTPHeaders struct {
	// Set the Content-Security-Policy header to HTTP responses.
	// Unset if blank.
	ContentSecurityPolicy string `json:"contentSecurityPolicy,omitempty"`
	// Set the X-Frame-Options header to HTTP responses.
	// Unset if blank. Accepted values are deny and sameorigin.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
	//+kubebuilder:validation:Enum="";Deny;SameOrigin
	XFrameOptions string `json:"xFrameOptions,omitempty"`
	// Set the X-Content-Type-Options header to HTTP responses.
	// Unset if blank. Accepted value is nosniff.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	//+kubebuilder:validation:Enum="";NoSniff
	XContentTypeOptions string `json:"xContentTypeOptions,omitempty"`
	// Set the X-XSS-Protection header to all responses.
	// Unset if blank.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection
	XXSSProtection string `json:"xXSSProtection,omitempty"`
	// Set the Strict-Transport-Security header to HTTP responses.
	// Unset if blank.
	// Please make sure that you use this with care as this header might force
	// browsers to load Prometheus and the other applications hosted on the same
	// domain and subdomains over HTTPS.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
	StrictTransportSecurity string `json:"strictTransportSecurity,omitempty"`
}

// WebTLSConfig defines the TLS parameters for HTTPS.
// +k8s:openapi-gen=true
type WebTLSConfig struct {
	// Secret containing the TLS key for the server.
	KeySecret v1.SecretKeySelector `json:"keySecret"`
	// Contains the TLS certificate for the server.
	Cert SecretOrConfigMap `json:"cert"`
	// Server policy for client authentication. Maps to ClientAuth Policies.
	// For more detail on clientAuth options:
	// https://golang.org/pkg/crypto/tls/#ClientAuthType
	ClientAuthType string `json:"clientAuthType,omitempty"`
	// Contains the CA certificate for client certificate authentication to the server.
	ClientCA SecretOrConfigMap `json:"client_ca,omitempty"`
	// Minimum TLS version that is acceptable. Defaults to TLS12.
	MinVersion string `json:"minVersion,omitempty"`
	// Maximum TLS version that is acceptable. Defaults to TLS13.
	MaxVersion string `json:"maxVersion,omitempty"`
	// List of supported cipher suites for TLS versions up to TLS 1.2. If empty,
	// Go default cipher suites are used. Available cipher suites are documented
	// in the go documentation: https://golang.org/pkg/crypto/tls/#pkg-constants
	CipherSuites []string `json:"cipherSuites,omitempty"`
	// Controls whether the server selects the
	// client's most preferred cipher suite, or the server's most preferred
	// cipher suite. If true then the server's preference, as expressed in
	// the order of elements in cipherSuites, is used.
	PreferServerCipherSuites *bool `json:"preferServerCipherSuites,omitempty"`
	// Elliptic curves that will be used in an ECDHE handshake, in preference
	// order. Available curves are documented in the go documentation:
	// https://golang.org/pkg/crypto/tls/#CurveID
	CurvePreferences []string `json:"curvePreferences,omitempty"`
}

// WebTLSConfigError is returned by WebTLSConfig.Validate() on
// semantically invalid configurations.
// +k8s:openapi-gen=false
type WebTLSConfigError struct {
	err string
}

func (e *WebTLSConfigError) Error() string {
	return e.err
}

func (c *WebTLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.ClientCA != (SecretOrConfigMap{}) {
		if err := c.ClientCA.Validate(); err != nil {
			msg := fmt.Sprintf("invalid web tls config: %s", err.Error())
			return &WebTLSConfigError{msg}
		}
	}

	if c.Cert == (SecretOrConfigMap{}) {
		return &WebTLSConfigError{"invalid web tls config: cert must be defined"}
	} else if err := c.Cert.Validate(); err != nil {
		msg := fmt.Sprintf("invalid web tls config: %s", err.Error())
		return &WebTLSConfigError{msg}
	}

	if c.KeySecret == (v1.SecretKeySelector{}) {
		return &WebTLSConfigError{"invalid web tls config: key must be defined"}
	}

	return nil
}

// ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.
// +k8s:openapi-gen=true
type ThanosSpec struct {
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Thanos is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Version describes the version of Thanos to use.
	Version *string `json:"version,omitempty"`
	// Tag of Thanos sidecar container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	// Deprecated: use 'image' instead.  The image tag can be specified
	// as part of the image URL.
	Tag *string `json:"tag,omitempty"`
	// SHA of Thanos container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	// Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image URL.
	SHA *string `json:"sha,omitempty"`
	// Thanos base image if other than default.
	// Deprecated: use 'image' instead
	BaseImage *string `json:"baseImage,omitempty"`
	// Resources defines the resource requirements for the Thanos sidecar.
	// If not provided, no requests/limits will be set
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// ObjectStorageConfig configures object storage in Thanos.
	// Alternative to ObjectStorageConfigFile, and lower order priority.
	ObjectStorageConfig *v1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
	// ObjectStorageConfigFile specifies the path of the object storage configuration file.
	// When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.
	ObjectStorageConfigFile *string `json:"objectStorageConfigFile,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the HTTP and gRPC endpoints.
	// It takes precedence over `grpcListenLocal` and `httpListenLocal`.
	// Deprecated: use `grpcListenLocal` and `httpListenLocal` instead.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the gRPC endpoints.
	// It has no effect if `listenLocal` is true.
	GRPCListenLocal bool `json:"grpcListenLocal,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the HTTP endpoints.
	// It has no effect if `listenLocal` is true.
	HTTPListenLocal bool `json:"httpListenLocal,omitempty"`
	// TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.
	TracingConfig *v1.SecretKeySelector `json:"tracingConfig,omitempty"`
	// TracingConfig specifies the path of the tracing configuration file.
	// When used alongside with TracingConfig, TracingConfigFile takes precedence.
	TracingConfigFile string `json:"tracingConfigFile,omitempty"`
	// GRPCServerTLSConfig configures the TLS parameters for the gRPC server
	// providing the StoreAPI.
	// Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
	// Maps to the '--grpc-server-tls-*' CLI args.
	GRPCServerTLSConfig *TLSConfig `json:"grpcServerTlsConfig,omitempty"`
	// LogLevel for Thanos sidecar to be configured with.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// LogFormat for Thanos sidecar to be configured with.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`
	// MinTime for Thanos sidecar to be configured with. Option can be a constant time in RFC3339 format or time duration relative to current time, such as -1d or 2h45m. Valid duration units are ms, s, m, h, d, w, y.
	MinTime string `json:"minTime,omitempty"`
	// ReadyTimeout is the maximum time Thanos sidecar will wait for Prometheus to start. Eg 10m
	ReadyTimeout Duration `json:"readyTimeout,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the thanos-sidecar container.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// AdditionalArgs allows setting additional arguments for the Thanos container.
	// The arguments are passed as-is to the Thanos container which may cause issues
	// if they are invalid or not supported the given Thanos version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument the reconciliation will
	// fail and an error will be logged.
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`
}

// RemoteWriteSpec defines the configuration to write samples from Prometheus
// to a remote endpoint.
// +k8s:openapi-gen=true
type RemoteWriteSpec struct {
	RemoteWriteSpecV2 `json:",inline"`
	// Enables sending of exemplars over remote write. Note that
	// exemplar-storage itself must be enabled using the enableFeature option
	// for exemplars to be scraped in the first place.  Only valid in
	// Prometheus versions 2.27.0 and newer.
	SendExemplars *bool `json:"sendExemplars,omitempty"`
	// Custom HTTP headers to be sent along with each remote write request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	// Only valid in Prometheus versions 2.25.0 and newer.
	Headers map[string]string `json:"headers,omitempty"`
	// The list of remote write relabel configurations.
	WriteRelabelConfigs []RelabelConfig `json:"writeRelabelConfigs,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// BasicAuth for the URL.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for remote write.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for remote write.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for remote write
	Authorization *Authorization `json:"authorization,omitempty"`
	// Sigv4 allows to configures AWS's Signature Verification 4
	Sigv4 *Sigv4 `json:"sigv4,omitempty"`
	// TLS Config to use for remote write.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// MetadataConfig configures the sending of series metadata to the remote storage.
	MetadataConfig *MetadataConfig `json:"metadataConfig,omitempty"`
}

// RemoteWriteSpecV2 thanos ruler remote write spec
type RemoteWriteSpecV2 struct {
	// The URL of the endpoint to send samples to.
	URL string `json:"url"`
	// The name of the remote write queue, it must be unique if specified. The
	// name is used in metrics and logging in order to differentiate queues.
	// Only valid in Prometheus versions 2.15.0 and newer.
	Name string `json:"name,omitempty"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	FollowRedirects bool `json:"follow_redirects,omitempty"`
	// Timeout for requests to the remote write endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`
	// QueueConfig allows tuning of the remote write queue parameters.
	QueueConfig *QueueConfig `json:"queueConfig,omitempty"`
}

// QueueConfig allows the tuning of remote write's queue_config parameters.
// This object is referenced in the RemoteWriteSpec object.
// +k8s:openapi-gen=true
type QueueConfig struct {
	// Capacity is the number of samples to buffer per shard before we start dropping them.
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
// sign requests. Cannot be set at the same time as basic_auth or authorization.
// +k8s:openapi-gen=true
type Sigv4 struct {
	// Region is the AWS region. If blank, the region from the default credentials chain used.
	Region string `json:"region,omitempty"`
	// AccessKey is the AWS API key. If blank, the environment variable `AWS_ACCESS_KEY_ID` is used.
	AccessKey *v1.SecretKeySelector `json:"accessKey,omitempty"`
	// SecretKey is the AWS API secret. If blank, the environment variable `AWS_SECRET_ACCESS_KEY` is used.
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
	// The name of the remote read queue, it must be unique if specified. The name
	// is used in metrics and logging in order to differentiate read
	// configurations.  Only valid in Prometheus versions 2.15.0 and newer.
	Name string `json:"name,omitempty"`
	// An optional list of equality matchers which have to be present
	// in a selector to query the remote read endpoint.
	RequiredMatchers map[string]string `json:"requiredMatchers,omitempty"`
	// Timeout for requests to the remote read endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`
	// Custom HTTP headers to be sent along with each remote read request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	// Only valid in Prometheus versions 2.26.0 and newer.
	Headers map[string]string `json:"headers,omitempty"`
	// Whether reads should be made for queries for time ranges that
	// the local storage should have complete data for.
	ReadRecent bool `json:"readRecent,omitempty"`
	// BasicAuth for the URL.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// Bearer token for remote read.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for remote read.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for remote read
	Authorization *Authorization `json:"authorization,omitempty"`
	// TLS Config to use for remote read.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// Whether to use the external labels as selectors for the remote read endpoint.
	// Requires Prometheus v2.34.0 and above.
	FilterExternalLabels *bool `json:"filterExternalLabels,omitempty"`
}

// LabelName is a valid Prometheus label name which may only contain ASCII letters, numbers, as well as underscores.
// +kubebuilder:validation:Pattern:="^[a-zA-Z_][a-zA-Z0-9_]*$"
type LabelName string

// Endpoint defines a scrapeable endpoint serving Prometheus metrics.
// +k8s:openapi-gen=true
type Endpoint struct {
	// Name of the service port this endpoint refers to. Mutually exclusive with targetPort.
	Port string `json:"port,omitempty"`
	// Name or number of the target port of the Pod behind the Service, the port must be specified with container port property. Mutually exclusive with port.
	TargetPort *intstr.IntOrString `json:"targetPort,omitempty"`
	// HTTP path to scrape for metrics.
	// If empty, Prometheus uses the default value (e.g. `/metrics`).
	Path string `json:"path,omitempty"`
	// HTTP scheme to use for scraping.
	// `http` and `https` are the expected values unless you rewrite the `__scheme__` label via relabeling.
	// If empty, Prometheus uses the default value `http`.
	// +kubebuilder:validation:Enum=http;https
	Scheme string `json:"scheme,omitempty"`
	// Optional HTTP URL parameters
	Params map[string][]string `json:"params,omitempty"`
	// Interval at which metrics should be scraped
	// If not specified Prometheus' global scrape interval is used.
	Interval Duration `json:"interval,omitempty"`
	// Timeout after which the scrape is ended
	// If not specified, the Prometheus global scrape timeout is used unless it is less than `Interval` in which the latter is used.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`
	// TLS configuration to use when scraping the endpoint
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// File to read bearer token for scraping targets.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Secret to mount to read bearer token for scraping targets. The secret
	// needs to be in the same namespace as the service monitor and accessible by
	// the Prometheus Operator.
	BearerTokenSecret v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// Authorization section for this endpoint
	Authorization *SafeAuthorization `json:"authorization,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	HonorLabels bool `json:"honorLabels,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	MetricRelabelConfigs []*RelabelConfig `json:"metricRelabelings,omitempty"`
	// RelabelConfigs to apply to samples before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	RelabelConfigs []*RelabelConfig `json:"relabelings,omitempty"`
	// ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint.
	ProxyURL *string `json:"proxyUrl,omitempty"`
	// FollowRedirects configures whether scrape requests follow HTTP 3xx redirects.
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	EnableHttp2 *bool `json:"enableHttp2,omitempty"`
	// Drop pods that are not running. (Failed, Succeeded). Enabled by default.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
	FilterRunning *bool `json:"filterRunning,omitempty"`
}

type AttachMetadata struct {
	// When set to true, Prometheus must have permissions to get Nodes.
	Node bool `json:"node,omitempty"`
}

// OAuth2 allows an endpoint to authenticate with OAuth2.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#oauth2
// +k8s:openapi-gen=true
type OAuth2 struct {
	// The secret or configmap containing the OAuth2 client id
	ClientID SecretOrConfigMap `json:"clientId"`
	// The secret containing the OAuth2 client secret
	ClientSecret v1.SecretKeySelector `json:"clientSecret"`
	// The URL to fetch the token from
	// +kubebuilder:validation:MinLength=1
	TokenURL string `json:"tokenUrl"`
	// OAuth2 scopes used for the token request
	Scopes []string `json:"scopes,omitempty"`
	// Parameters to append to the token URL
	EndpointParams map[string]string `json:"endpointParams,omitempty"`
}

type OAuth2ValidationError struct {
	err string
}

func (e *OAuth2ValidationError) Error() string {
	return e.err
}

func (o *OAuth2) Validate() error {
	if o.TokenURL == "" {
		return &OAuth2ValidationError{err: "OAuth2 token url must be specified"}
	}

	if o.ClientID == (SecretOrConfigMap{}) {
		return &OAuth2ValidationError{err: "OAuth2 client id must be specified"}
	}

	if err := o.ClientID.Validate(); err != nil {
		return &OAuth2ValidationError{
			err: fmt.Sprintf("invalid OAuth2 client id: %s", err.Error()),
		}
	}

	return nil
}

// BasicAuth allow an endpoint to authenticate over basic authentication
// More info: https://prometheus.io/docs/operating/configuration/#endpoints
// +k8s:openapi-gen=true
type BasicAuth struct {
	// The secret in the service monitor namespace that contains the username
	// for authentication.
	Username v1.SecretKeySelector `json:"username,omitempty"`
	// The secret in the service monitor namespace that contains the password
	// for authentication.
	Password v1.SecretKeySelector `json:"password,omitempty"`
}

// SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.
type SecretOrConfigMap struct {
	// Secret containing data to use for the targets.
	Secret *v1.SecretKeySelector `json:"secret,omitempty"`
	// ConfigMap containing data to use for the targets.
	ConfigMap *v1.ConfigMapKeySelector `json:"configMap,omitempty"`
}

// SecretOrConfigMapValidationError is returned by SecretOrConfigMap.Validate()
// on semantically invalid configurations.
// +k8s:openapi-gen=false
type SecretOrConfigMapValidationError struct {
	err string
}

func (e *SecretOrConfigMapValidationError) Error() string {
	return e.err
}

// Validate semantically validates the given TLSConfig.
func (c *SecretOrConfigMap) Validate() error {
	if c.Secret != nil && c.ConfigMap != nil {
		return &SecretOrConfigMapValidationError{"SecretOrConfigMap can not specify both Secret and ConfigMap"}
	}

	return nil
}

// SafeTLSConfig specifies safe TLS configuration parameters.
// +k8s:openapi-gen=true
type SafeTLSConfig struct {
	// Certificate authority used when verifying server certificates.
	CA SecretOrConfigMap `json:"ca,omitempty"`
	// Client certificate to present when doing client-authentication.
	Cert SecretOrConfigMap `json:"cert,omitempty"`
	// Secret containing the client key file for the targets.
	KeySecret *v1.SecretKeySelector `json:"keySecret,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `json:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// Validate semantically validates the given SafeTLSConfig.
func (c *SafeTLSConfig) Validate() error {
	if c.CA != (SecretOrConfigMap{}) {
		if err := c.CA.Validate(); err != nil {
			return err
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if err := c.Cert.Validate(); err != nil {
			return err
		}
	}

	if c.Cert != (SecretOrConfigMap{}) && c.KeySecret == nil {
		return &TLSConfigValidationError{"client cert specified without client key"}
	}

	if c.KeySecret != nil && c.Cert == (SecretOrConfigMap{}) {
		return &TLSConfigValidationError{"client key specified without client cert"}
	}

	return nil
}

// TLSConfig extends the safe TLS configuration with file parameters.
// +k8s:openapi-gen=true
type TLSConfig struct {
	SafeTLSConfig `json:",inline"`
	// Path to the CA cert in the Prometheus container to use for the targets.
	CAFile string `json:"caFile,omitempty"`
	// Path to the client cert file in the Prometheus container for the targets.
	CertFile string `json:"certFile,omitempty"`
	// Path to the client key file in the Prometheus container for the targets.
	KeyFile string `json:"keyFile,omitempty"`
}

// TLSConfigValidationError is returned by TLSConfig.Validate() on semantically
// invalid tls configurations.
// +k8s:openapi-gen=false
type TLSConfigValidationError struct {
	err string
}

func (e *TLSConfigValidationError) Error() string {
	return e.err
}

// Validate semantically validates the given TLSConfig.
func (c *TLSConfig) Validate() error {
	if c.CA != (SecretOrConfigMap{}) {
		if c.CAFile != "" {
			return &TLSConfigValidationError{"tls config can not both specify CAFile and CA"}
		}
		if err := c.CA.Validate(); err != nil {
			return &TLSConfigValidationError{"tls config CA is invalid"}
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if c.CertFile != "" {
			return &TLSConfigValidationError{"tls config can not both specify CertFile and Cert"}
		}
		if err := c.Cert.Validate(); err != nil {
			return &TLSConfigValidationError{"tls config Cert is invalid"}
		}
	}

	if c.KeyFile != "" && c.KeySecret != nil {
		return &TLSConfigValidationError{"tls config can not both specify KeyFile and KeySecret"}
	}

	hasCert := c.CertFile != "" || c.Cert != (SecretOrConfigMap{})
	hasKey := c.KeyFile != "" || c.KeySecret != nil

	if hasCert && !hasKey {
		return &TLSConfigValidationError{"tls config can not specify client cert without client key"}
	}

	if hasKey && !hasCert {
		return &TLSConfigValidationError{"tls config can not specify client key without client cert"}
	}

	return nil
}

// NamespaceSelector is a selector for selecting either all namespaces or a
// list of namespaces.
// If `any` is true, it takes precedence over `matchNames`.
// If `matchNames` is empty and `any` is false, it means that the objects are
// selected from the current namespace.
// +k8s:openapi-gen=true
type NamespaceSelector struct {
	// Boolean describing whether all namespaces are selected in contrast to a
	// list restricting them.
	Any bool `json:"any,omitempty"`
	// List of namespace names to select from.
	MatchNames []string `json:"matchNames,omitempty"`

	// TODO(fabxc): this should embed metav1.LabelSelector eventually.
	// Currently the selector is only used for namespaces which require more complex
	// implementation to support label selections.
}

// Argument as part of the AdditionalArgs list.
// +k8s:openapi-gen=true
type Argument struct {
	// Name of the argument, e.g. "scrape.discovery-reload-interval".
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Argument value, e.g. 30s. Can be empty for name-only arguments (e.g. --storage.tsdb.no-lockfile)
	Value string `json:"value,omitempty"`
}
