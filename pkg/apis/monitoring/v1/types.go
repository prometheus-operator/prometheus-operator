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
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
)

const (
	Version = "v1"
)

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

type ProxyConfig struct {
	// `proxyURL` defines the HTTP proxy server to use.
	//
	// +kubebuilder:validation:Pattern:="^(http|https|socks5)://.+$"
	// +optional
	ProxyURL *string `json:"proxyUrl,omitempty"`
	// `noProxy` is a comma-separated string that can contain IPs, CIDR notation, domain names
	// that should be excluded from proxying. IP and domain names can
	// contain port numbers.
	//
	// It requires Prometheus >= v2.43.0, Alertmanager >= v0.25.0 or Thanos >= v0.32.0.
	// +optional
	NoProxy *string `json:"noProxy,omitempty"`
	// Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).
	//
	// It requires Prometheus >= v2.43.0, Alertmanager >= v0.25.0 or Thanos >= v0.32.0.
	// +optional
	ProxyFromEnvironment *bool `json:"proxyFromEnvironment,omitempty"`
	// ProxyConnectHeader optionally specifies headers to send to
	// proxies during CONNECT requests.
	//
	// It requires Prometheus >= v2.43.0, Alertmanager >= v0.25.0 or Thanos >= v0.32.0.
	// +optional
	// +mapType:=atomic
	ProxyConnectHeader map[string][]v1.SecretKeySelector `json:"proxyConnectHeader,omitempty"`
}

// Validate semantically validates the given ProxyConfig.
func (pc *ProxyConfig) Validate() error {
	if pc == nil {
		return nil
	}

	if reflect.ValueOf(pc).IsZero() {
		return nil
	}

	proxyFromEnvironmentDefined := pc.ProxyFromEnvironment != nil && *pc.ProxyFromEnvironment
	proxyURLDefined := pc.ProxyURL != nil && *pc.ProxyURL != ""
	noProxyDefined := pc.NoProxy != nil && *pc.NoProxy != ""

	if len(pc.ProxyConnectHeader) > 0 && (!proxyFromEnvironmentDefined && !proxyURLDefined) {
		return fmt.Errorf("if proxyConnectHeader is configured, proxyUrl or proxyFromEnvironment must also be configured")
	}

	if proxyFromEnvironmentDefined && proxyURLDefined {
		return fmt.Errorf("if proxyFromEnvironment is configured, proxyUrl must not be configured")
	}

	if proxyFromEnvironmentDefined && noProxyDefined {
		return fmt.Errorf("if proxyFromEnvironment is configured, noProxy must not be configured")
	}

	if !proxyURLDefined && noProxyDefined {
		return fmt.Errorf("if noProxy is configured, proxyUrl must also be configured")
	}

	for k, v := range pc.ProxyConnectHeader {
		if len(v) == 0 {
			return fmt.Errorf("proxyConnetHeader[%s]: selector must not be empty", k)
		}
		for i, sel := range v {
			if sel == (v1.SecretKeySelector{}) {
				return fmt.Errorf("proxyConnectHeader[%s][%d]: selector must be defined", k, i)
			}
		}
	}

	if pc.ProxyURL != nil {
		if _, err := url.Parse(*pc.ProxyURL); err != nil {
			return err
		}
	}
	return nil
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
	// +kubebuilder:validation:Enum=prometheusrules;servicemonitors;podmonitors;probes;scrapeconfigs
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
	return schema.GroupKind{
		Kind:  monitoring.ResourceToKind(obj.Resource),
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

// +kubebuilder:validation:MinLength=1
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

// +kubebuilder:validation:MinLength=1
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

	// +optional
	// Deprecated: this field is never set.
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
	// Secret or ConfigMap containing the TLS certificate for the web server.
	//
	// Either `keySecret` or `keyFile` must be defined.
	//
	// It is mutually exclusive with `certFile`.
	//
	// +optional
	Cert SecretOrConfigMap `json:"cert,omitempty"`
	// Path to the TLS certificate file in the container for the web server.
	//
	// Either `keySecret` or `keyFile` must be defined.
	//
	// It is mutually exclusive with `cert`.
	//
	// +optional
	CertFile *string `json:"certFile,omitempty"`

	// Secret containing the TLS private key for the web server.
	//
	// Either `cert` or `certFile` must be defined.
	//
	// It is mutually exclusive with `keyFile`.
	//
	// +optional
	KeySecret v1.SecretKeySelector `json:"keySecret,omitempty"`
	// Path to the TLS private key file in the container for the web server.
	//
	// If defined, either `cert` or `certFile` must be defined.
	//
	// It is mutually exclusive with `keySecret`.
	//
	// +optional
	KeyFile *string `json:"keyFile,omitempty"`

	// Secret or ConfigMap containing the CA certificate for client certificate
	// authentication to the server.
	//
	// It is mutually exclusive with `clientCAFile`.
	//
	// +optional
	ClientCA SecretOrConfigMap `json:"client_ca,omitempty"`
	// Path to the CA certificate file for client certificate authentication to
	// the server.
	//
	// It is mutually exclusive with `client_ca`.
	//
	// +optional
	ClientCAFile *string `json:"clientCAFile,omitempty"`
	// The server policy for client TLS authentication.
	//
	// For more detail on clientAuth options:
	// https://golang.org/pkg/crypto/tls/#ClientAuthType
	//
	// +optional
	ClientAuthType *string `json:"clientAuthType,omitempty"`

	// Minimum TLS version that is acceptable.
	//
	// +optional
	MinVersion *string `json:"minVersion,omitempty"`
	// Maximum TLS version that is acceptable.
	//
	// +optional
	MaxVersion *string `json:"maxVersion,omitempty"`

	// List of supported cipher suites for TLS versions up to TLS 1.2.
	//
	// If not defined, the Go default cipher suites are used.
	// Available cipher suites are documented in the Go documentation:
	// https://golang.org/pkg/crypto/tls/#pkg-constants
	//
	// +optional
	CipherSuites []string `json:"cipherSuites,omitempty"`

	// Controls whether the server selects the client's most preferred cipher
	// suite, or the server's most preferred cipher suite.
	//
	// If true then the server's preference, as expressed in
	// the order of elements in cipherSuites, is used.
	//
	// +optional
	PreferServerCipherSuites *bool `json:"preferServerCipherSuites,omitempty"`

	// Elliptic curves that will be used in an ECDHE handshake, in preference
	// order.
	//
	// Available curves are documented in the Go documentation:
	// https://golang.org/pkg/crypto/tls/#CurveID
	//
	// +optional
	CurvePreferences []string `json:"curvePreferences,omitempty"`
}

// Validate returns an error if one of the WebTLSConfig fields is invalid.
// A valid WebTLSConfig should have (Cert or CertFile) and (KeySecret or KeyFile) fields which are not
// zero values.
func (c *WebTLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.ClientCA != (SecretOrConfigMap{}) {
		if c.ClientCAFile != nil && *c.ClientCAFile != "" {
			return errors.New("cannot specify both clientCAFile and clientCA")
		}

		if err := c.ClientCA.Validate(); err != nil {
			return fmt.Errorf("invalid client CA: %w", err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if c.CertFile != nil && *c.CertFile != "" {
			return errors.New("cannot specify both cert and certFile")
		}
		if err := c.Cert.Validate(); err != nil {
			return fmt.Errorf("invalid TLS certificate: %w", err)
		}
	}

	if c.KeyFile != nil && *c.KeyFile != "" && c.KeySecret != (v1.SecretKeySelector{}) {
		return errors.New("cannot specify both keyFile and keySecret")
	}

	if (c.KeyFile == nil || *c.KeyFile == "") && c.KeySecret == (v1.SecretKeySelector{}) {
		return errors.New("TLS private key must be defined")
	}

	if (c.CertFile == nil || *c.CertFile == "") && c.Cert == (SecretOrConfigMap{}) {
		return errors.New("TLS certificate must be defined")
	}

	return nil
}

// LabelName is a valid Prometheus label name which may only contain ASCII
// letters, numbers, as well as underscores.
//
// +kubebuilder:validation:Pattern:="^[a-zA-Z_][a-zA-Z0-9_]*$"
type LabelName string

// Endpoint defines an endpoint serving Prometheus metrics to be scraped by
// Prometheus.
//
// +k8s:openapi-gen=true
type Endpoint struct {
	// Name of the Service port which this endpoint refers to.
	//
	// It takes precedence over `targetPort`.
	Port string `json:"port,omitempty"`

	// Name or number of the target port of the `Pod` object behind the
	// Service. The port must be specified with the container's port property.
	//
	// +optional
	TargetPort *intstr.IntOrString `json:"targetPort,omitempty"`

	// HTTP path from which to scrape for metrics.
	//
	// If empty, Prometheus uses the default value (e.g. `/metrics`).
	Path string `json:"path,omitempty"`

	// HTTP scheme to use for scraping.
	//
	// `http` and `https` are the expected values unless you rewrite the
	// `__scheme__` label via relabeling.
	//
	// If empty, Prometheus uses the default value `http`.
	//
	// +kubebuilder:validation:Enum=http;https
	Scheme string `json:"scheme,omitempty"`

	// params define optional HTTP URL parameters.
	Params map[string][]string `json:"params,omitempty"`

	// Interval at which Prometheus scrapes the metrics from the target.
	//
	// If empty, Prometheus uses the global scrape interval.
	Interval Duration `json:"interval,omitempty"`

	// Timeout after which Prometheus considers the scrape to be failed.
	//
	// If empty, Prometheus uses the global scrape timeout unless it is less
	// than the target's scrape interval value in which the latter is used.
	// The value cannot be greater than the scrape interval otherwise the operator will reject the resource.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`

	// TLS configuration to use when scraping the target.
	//
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// File to read bearer token for scraping the target.
	//
	// Deprecated: use `authorization` instead.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`

	// `bearerTokenSecret` specifies a key of a Secret containing the bearer
	// token for scraping targets. The secret needs to be in the same namespace
	// as the ServiceMonitor object and readable by the Prometheus Operator.
	//
	// +optional
	//
	// Deprecated: use `authorization` instead.
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`

	// `authorization` configures the Authorization header credentials to use when
	// scraping the target.
	//
	// Cannot be set at the same time as `basicAuth`, or `oauth2`.
	//
	// +optional
	Authorization *SafeAuthorization `json:"authorization,omitempty"`

	// When true, `honorLabels` preserves the metric's labels when they collide
	// with the target's labels.
	HonorLabels bool `json:"honorLabels,omitempty"`

	// `honorTimestamps` controls whether Prometheus preserves the timestamps
	// when exposed by the target.
	//
	// +optional
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`

	// `trackTimestampsStaleness` defines whether Prometheus tracks staleness of
	// the metrics that have an explicit timestamp present in scraped data.
	// Has no effect if `honorTimestamps` is false.
	//
	// It requires Prometheus >= v2.48.0.
	//
	// +optional
	TrackTimestampsStaleness *bool `json:"trackTimestampsStaleness,omitempty"`

	// `basicAuth` configures the Basic Authentication credentials to use when
	// scraping the target.
	//
	// Cannot be set at the same time as `authorization`, or `oauth2`.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// `oauth2` configures the OAuth2 settings to use when scraping the target.
	//
	// It requires Prometheus >= 2.27.0.
	//
	// Cannot be set at the same time as `authorization`, or `basicAuth`.
	//
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`

	// `metricRelabelings` configures the relabeling rules to apply to the
	// samples before ingestion.
	//
	// +optional
	MetricRelabelConfigs []RelabelConfig `json:"metricRelabelings,omitempty"`

	// `relabelings` configures the relabeling rules to apply the target's
	// metadata labels.
	//
	// The Operator automatically adds relabelings for a few standard Kubernetes fields.
	//
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	//
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	//
	// +optional
	RelabelConfigs []RelabelConfig `json:"relabelings,omitempty"`

	// `proxyURL` configures the HTTP Proxy URL (e.g.
	// "http://proxyserver:2195") to go through when scraping the target.
	//
	// +optional
	ProxyURL *string `json:"proxyUrl,omitempty"`

	// `followRedirects` defines whether the scrape requests should follow HTTP
	// 3xx redirects.
	//
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`

	// `enableHttp2` can be used to disable HTTP2 when scraping the target.
	//
	// +optional
	EnableHttp2 *bool `json:"enableHttp2,omitempty"`

	// When true, the pods which are not running (e.g. either in Failed or
	// Succeeded state) are dropped during the target discovery.
	//
	// If unset, the filtering is enabled.
	//
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
	//
	// +optional
	FilterRunning *bool `json:"filterRunning,omitempty"`
}

type AttachMetadata struct {
	// When set to true, Prometheus attaches node metadata to the discovered
	// targets.
	//
	// The Prometheus service account must have the `list` and `watch`
	// permissions on the `Nodes` objects.
	//
	// +optional
	Node *bool `json:"node,omitempty"`
}

// OAuth2 configures OAuth2 settings.
//
// +k8s:openapi-gen=true
type OAuth2 struct {
	// `clientId` specifies a key of a Secret or ConfigMap containing the
	// OAuth2 client's ID.
	ClientID SecretOrConfigMap `json:"clientId"`

	// `clientSecret` specifies a key of a Secret containing the OAuth2
	// client's secret.
	ClientSecret v1.SecretKeySelector `json:"clientSecret"`

	// `tokenURL` configures the URL to fetch the token from.
	//
	// +kubebuilder:validation:MinLength=1
	TokenURL string `json:"tokenUrl"`

	// `scopes` defines the OAuth2 scopes used for the token request.
	//
	// +optional.
	Scopes []string `json:"scopes,omitempty"`

	// `endpointParams` configures the HTTP parameters to append to the token
	// URL.
	//
	// +optional
	EndpointParams map[string]string `json:"endpointParams,omitempty"`

	// TLS configuration to use when connecting to the OAuth2 server.
	// It requires Prometheus >= v2.43.0.
	//
	// +optional
	TLSConfig *SafeTLSConfig `json:"tlsConfig,omitempty"`

	// Proxy configuration to use when connecting to the OAuth2 server.
	// It requires Prometheus >= v2.43.0.
	//
	// +optional
	ProxyConfig `json:",inline"`
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

	if err := o.TLSConfig.Validate(); err != nil {
		return &OAuth2ValidationError{
			err: fmt.Sprintf("invalid OAuth2 tlsConfig: %s", err.Error()),
		}
	}

	return nil
}

// BasicAuth configures HTTP Basic Authentication settings.
//
// +k8s:openapi-gen=true
type BasicAuth struct {
	// `username` specifies a key of a Secret containing the username for
	// authentication.
	Username v1.SecretKeySelector `json:"username,omitempty"`

	// `password` specifies a key of a Secret containing the password for
	// authentication.
	Password v1.SecretKeySelector `json:"password,omitempty"`
}

// SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.
type SecretOrConfigMap struct {
	// Secret containing data to use for the targets.
	Secret *v1.SecretKeySelector `json:"secret,omitempty"`
	// ConfigMap containing data to use for the targets.
	ConfigMap *v1.ConfigMapKeySelector `json:"configMap,omitempty"`
}

// Validate semantically validates the given SecretOrConfigMap.
func (c *SecretOrConfigMap) Validate() error {
	if c == nil {
		return nil
	}

	if c.Secret != nil && c.ConfigMap != nil {
		return fmt.Errorf("cannot specify both Secret and ConfigMap")
	}

	return nil
}

func (c *SecretOrConfigMap) String() string {
	if c == nil {
		return "<nil>"
	}

	switch {
	case c.Secret != nil:
		return fmt.Sprintf("<secret=%s,key=%s>", c.Secret.LocalObjectReference.Name, c.Secret.Key)
	case c.ConfigMap != nil:
		return fmt.Sprintf("<configmap=%s,key=%s>", c.ConfigMap.LocalObjectReference.Name, c.ConfigMap.Key)
	}

	return "<empty>"
}

// +kubebuilder:validation:Enum=TLS10;TLS11;TLS12;TLS13
type TLSVersion string

const (
	TLSVersion10 TLSVersion = "TLS10"
	TLSVersion11 TLSVersion = "TLS11"
	TLSVersion12 TLSVersion = "TLS12"
	TLSVersion13 TLSVersion = "TLS13"
)

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
	// +optional
	ServerName *string `json:"serverName,omitempty"`

	// Disable target certificate validation.
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`

	// Minimum acceptable TLS version.
	//
	// It requires Prometheus >= v2.35.0 or Thanos >= v0.28.0.
	// +optional
	MinVersion *TLSVersion `json:"minVersion,omitempty"`

	// Maximum acceptable TLS version.
	//
	// It requires Prometheus >= v2.41.0 or Thanos >= v0.31.0.
	// +optional
	MaxVersion *TLSVersion `json:"maxVersion,omitempty"`
}

// Validate semantically validates the given SafeTLSConfig.
func (c *SafeTLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.CA != (SecretOrConfigMap{}) {
		if err := c.CA.Validate(); err != nil {
			return fmt.Errorf("ca %s: %w", c.CA.String(), err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if err := c.Cert.Validate(); err != nil {
			return fmt.Errorf("cert %s: %w", c.Cert.String(), err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) && c.KeySecret == nil {
		return fmt.Errorf("client cert specified without client key")
	}

	if c.KeySecret != nil && c.Cert == (SecretOrConfigMap{}) {
		return fmt.Errorf("client key specified without client cert")
	}

	if c.MaxVersion != nil && c.MinVersion != nil && strings.Compare(string(*c.MaxVersion), string(*c.MinVersion)) == -1 {
		return fmt.Errorf("maxVersion must more than or equal to minVersion")
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

// Validate semantically validates the given TLSConfig.
func (c *TLSConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.CA != (SecretOrConfigMap{}) {
		if c.CAFile != "" {
			return fmt.Errorf("cannot specify both caFile and ca")
		}
		if err := c.CA.Validate(); err != nil {
			return fmt.Errorf("SecretOrConfigMap ca: %w", err)
		}
	}

	if c.Cert != (SecretOrConfigMap{}) {
		if c.CertFile != "" {
			return fmt.Errorf("cannot specify both certFile and cert")
		}
		if err := c.Cert.Validate(); err != nil {
			return fmt.Errorf("SecretOrConfigMap cert: %w", err)
		}
	}

	if c.KeyFile != "" && c.KeySecret != nil {
		return fmt.Errorf("cannot specify both keyFile and keySecret")
	}

	hasCert := c.CertFile != "" || c.Cert != (SecretOrConfigMap{})
	hasKey := c.KeyFile != "" || c.KeySecret != nil

	if hasCert && !hasKey {
		return fmt.Errorf("cannot specify client cert without client key")
	}

	if hasKey && !hasCert {
		return fmt.Errorf("cannot specify client key without client cert")
	}

	if c.MaxVersion != nil && c.MinVersion != nil && strings.Compare(string(*c.MaxVersion), string(*c.MinVersion)) == -1 {
		return fmt.Errorf("maxVersion must more than or equal to minVersion")
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

// The valid options for Role.
const (
	RoleNode          = "node"
	RolePod           = "pod"
	RoleService       = "service"
	RoleEndpoint      = "endpoints"
	RoleEndpointSlice = "endpointslice"
	RoleIngress       = "ingress"
)

// NativeHistogramConfig extends the native histogram configuration settings.
// +k8s:openapi-gen=true
type NativeHistogramConfig struct {
	// Whether to scrape a classic histogram that is also exposed as a native histogram.
	// It requires Prometheus >= v2.45.0.
	//
	// +optional
	ScrapeClassicHistograms *bool `json:"scrapeClassicHistograms,omitempty"`

	// If there are more than this many buckets in a native histogram,
	// buckets will be merged to stay within the limit.
	// It requires Prometheus >= v2.45.0.
	//
	// +optional
	NativeHistogramBucketLimit *uint64 `json:"nativeHistogramBucketLimit,omitempty"`

	// If the growth factor of one bucket to the next is smaller than this,
	// buckets will be merged to increase the factor sufficiently.
	// It requires Prometheus >= v2.50.0.
	//
	// +optional
	NativeHistogramMinBucketFactor *resource.Quantity `json:"nativeHistogramMinBucketFactor,omitempty"`

	// Whether to convert all scraped classic histograms into a native histogram with custom buckets.
	// It requires Prometheus >= v3.0.0.
	//
	// +optional
	ConvertClassicHistogramsToNHCB *bool `json:"convertClassicHistogramsToNHCB,omitempty"`
}

// +kubebuilder:validation:Enum=RelabelConfig;RoleSelector
type SelectorMechanism string

const (
	SelectorMechanismRelabel SelectorMechanism = "RelabelConfig"
	SelectorMechanismRole    SelectorMechanism = "RoleSelector"
)
