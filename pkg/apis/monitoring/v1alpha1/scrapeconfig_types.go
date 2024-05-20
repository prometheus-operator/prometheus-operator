// Copyright 2023 The prometheus-operator Authors
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

package v1alpha1

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ScrapeConfigsKind   = "ScrapeConfig"
	ScrapeConfigName    = "scrapeconfigs"
	ScrapeConfigKindKey = "scrapeconfig"
)

// Target represents a target for Prometheus to scrape
type Target string

// SDFile represents a file used for service discovery
// +kubebuilder:validation:Pattern=`^[^*]*(\*[^/]*)?\.(json|yml|yaml|JSON|YML|YAML)$`
type SDFile string

// NamespaceDiscovery is the configuration for discovering
// Kubernetes namespaces.
type NamespaceDiscovery struct {
	// Includes the namespace in which the Prometheus pod exists to the list of watched namesapces.
	// +optional
	IncludeOwnNamespace *bool `json:"ownNamespace,omitempty"`
	// List of namespaces where to watch for resources.
	// If empty and `ownNamespace` isn't true, Prometheus watches for resources in all namespaces.
	// +optional
	Names []string `json:"names,omitempty"`
}

type AttachMetadata struct {
	// Attaches node metadata to discovered targets.
	// When set to true, Prometheus must have the `get` permission on the
	// `Nodes` objects.
	// Only valid for Pod, Endpoint and Endpointslice roles.
	//
	// +optional
	Node *bool `json:"node,omitempty"`
}

// EC2Filter is the configuration for filtering EC2 instances.
type EC2Filter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// DockerFilter is the configuration to limit the discovery process to a subset of available resources.
type DockerFilter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Role is role of the service in Kubernetes.
// +kubebuilder:validation:Enum=Node;node;Service;service;Pod;pod;Endpoints;endpoints;EndpointSlice;endpointslice;Ingress;ingress
type Role string

// K8SSelectorConfig is Kubernetes Selector Config
type K8SSelectorConfig struct {
	// +kubebuilder:validation:Required
	Role  Role   `json:"role"`
	Label string `json:"label,omitempty"`
	Field string `json:"field,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="scfg"
// +kubebuilder:storageversion

// ScrapeConfig defines a namespaced Prometheus scrape_config to be aggregated across
// multiple namespaces into the Prometheus configuration.
type ScrapeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScrapeConfigSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ScrapeConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ScrapeConfigList is a list of ScrapeConfigs.
// +k8s:openapi-gen=true
type ScrapeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of ScrapeConfigs
	Items []*ScrapeConfig `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ScrapeConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ScrapeConfigSpec is a specification of the desired configuration for a scrape configuration.
// +k8s:openapi-gen=true
type ScrapeConfigSpec struct {
	// StaticConfigs defines a list of static targets with a common label set.
	// +optional
	StaticConfigs []StaticConfig `json:"staticConfigs,omitempty"`
	// FileSDConfigs defines a list of file service discovery configurations.
	// +optional
	FileSDConfigs []FileSDConfig `json:"fileSDConfigs,omitempty"`
	// HTTPSDConfigs defines a list of HTTP service discovery configurations.
	// +optional
	HTTPSDConfigs []HTTPSDConfig `json:"httpSDConfigs,omitempty"`
	// KubernetesSDConfigs defines a list of Kubernetes service discovery configurations.
	// +optional
	KubernetesSDConfigs []KubernetesSDConfig `json:"kubernetesSDConfigs,omitempty"`
	// ConsulSDConfigs defines a list of Consul service discovery configurations.
	// +optional
	ConsulSDConfigs []ConsulSDConfig `json:"consulSDConfigs,omitempty"`
	//DNSSDConfigs defines a list of DNS service discovery configurations.
	// +optional
	DNSSDConfigs []DNSSDConfig `json:"dnsSDConfigs,omitempty"`
	// EC2SDConfigs defines a list of EC2 service discovery configurations.
	// +optional
	EC2SDConfigs []EC2SDConfig `json:"ec2SDConfigs,omitempty"`
	// AzureSDConfigs defines a list of Azure service discovery configurations.
	// +optional
	AzureSDConfigs []AzureSDConfig `json:"azureSDConfigs,omitempty"`
	// GCESDConfigs defines a list of GCE service discovery configurations.
	// +optional
	GCESDConfigs []GCESDConfig `json:"gceSDConfigs,omitempty"`
	// OpenStackSDConfigs defines a list of OpenStack service discovery configurations.
	// +optional
	OpenStackSDConfigs []OpenStackSDConfig `json:"openstackSDConfigs,omitempty"`
	// DigitalOceanSDConfigs defines a list of DigitalOcean service discovery configurations.
	// +optional
	DigitalOceanSDConfigs []DigitalOceanSDConfig `json:"digitalOceanSDConfigs,omitempty"`
	// KumaSDConfigs defines a list of Kuma service discovery configurations.
	// +optional
	KumaSDConfigs []KumaSDConfig `json:"kumaSDConfigs,omitempty"`
	// EurekaSDConfigs defines a list of Eureka service discovery configurations.
	// +optional
	EurekaSDConfigs []EurekaSDConfig `json:"eurekaSDConfigs,omitempty"`
	// DockerSDConfigs defines a list of Docker service discovery configurations.
	// +optional
	DockerSDConfigs []DockerSDConfig `json:"dockerSDConfigs,omitempty"`
	// HetznerSDConfigs defines a list of Hetzner service discovery configurations.
	// +optional
	HetznerSDConfigs []HetznerSDConfig `json:"hetznerSDConfigs,omitempty"`
	// NomadSDConfigs defines a list of Nomad service discovery configurations.
	// +optional
	NomadSDConfigs []NomadSDConfig `json:"NomadSDConfigs,omitempty"`
	// RelabelConfigs defines how to rewrite the target's labels before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	// +optional
	RelabelConfigs []v1.RelabelConfig `json:"relabelings,omitempty"`
	// MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics).
	// +optional
	MetricsPath *string `json:"metricsPath,omitempty"`
	// ScrapeInterval is the interval between consecutive scrapes.
	// +optional
	ScrapeInterval *v1.Duration `json:"scrapeInterval,omitempty"`
	// ScrapeTimeout is the number of seconds to wait until a scrape request times out.
	// +optional
	ScrapeTimeout *v1.Duration `json:"scrapeTimeout,omitempty"`
	// The protocols to negotiate during a scrape. It tells clients the
	// protocols supported by Prometheus in order of preference (from most to least preferred).
	//
	// If unset, Prometheus uses its default value.
	//
	// It requires Prometheus >= v2.49.0.
	//
	// +listType=set
	// +optional
	ScrapeProtocols []v1.ScrapeProtocol `json:"scrapeProtocols,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	// +optional
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// TrackTimestampsStaleness whether Prometheus tracks staleness of
	// the metrics that have an explicit timestamp present in scraped data.
	// Has no effect if `honorTimestamps` is false.
	// It requires Prometheus >= v2.48.0.
	//
	// +optional
	TrackTimestampsStaleness *bool `json:"trackTimestampsStaleness,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	// +optional
	HonorLabels *bool `json:"honorLabels,omitempty"`
	// Optional HTTP URL parameters
	// +optional
	// +mapType:=atomic
	Params map[string][]string `json:"params,omitempty"`
	// Configures the protocol scheme used for requests.
	// If empty, Prometheus uses HTTP by default.
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	// +optional
	Scheme *string `json:"scheme,omitempty"`
	// When false, Prometheus will request uncompressed response from the scraped target.
	//
	// It requires Prometheus >= v2.49.0.
	//
	// If unset, Prometheus uses true by default.
	// +optional
	EnableCompression *bool `json:"enableCompression,omitempty"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// TLS configuration to use on every scrape request
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	// +optional
	SampleLimit *uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	// +optional
	TargetLimit *uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelLimit *uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelNameLengthLimit *uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	// +optional
	LabelValueLengthLimit *uint64 `json:"labelValueLengthLimit,omitempty"`
	// Per-scrape limit on the number of targets dropped by relabeling
	// that will be kept in memory. 0 means no limit.
	//
	// It requires Prometheus >= v2.47.0.
	//
	// +optional
	KeepDroppedTargets *uint64 `json:"keepDroppedTargets,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	// +optional
	MetricRelabelConfigs []v1.RelabelConfig `json:"metricRelabelings,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`

	// The scrape class to apply.
	// +optional
	// +kubebuilder:validation:MinLength=1
	ScrapeClassName *string `json:"scrapeClass,omitempty"`
}

// StaticConfig defines a Prometheus static configuration.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
// +k8s:openapi-gen=true
type StaticConfig struct {
	// List of targets for this static configuration.
	// +optional
	Targets []Target `json:"targets,omitempty"`
	// Labels assigned to all metrics scraped from the targets.
	// +mapType:=atomic
	// +optional
	Labels map[v1.LabelName]string `json:"labels,omitempty"`
}

// FileSDConfig defines a Prometheus file service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config
// +k8s:openapi-gen=true
type FileSDConfig struct {
	// List of files to be used for file discovery. Recommendation: use absolute paths. While relative paths work, the
	// prometheus-operator project makes no guarantees about the working directory where the configuration file is
	// stored.
	// Files must be mounted using Prometheus.ConfigMaps or Prometheus.Secrets.
	// +kubebuilder:validation:MinItems:=1
	Files []SDFile `json:"files"`
	// RefreshInterval configures the refresh interval at which Prometheus will reload the content of the files.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
}

// HTTPSDConfig defines a prometheus HTTP service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config
// +k8s:openapi-gen=true
type HTTPSDConfig struct {
	// URL from which the targets are fetched.
	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:Pattern:="^http(s)?://.+$"
	URL string `json:"url"`
	// RefreshInterval configures the refresh interval at which Prometheus will re-query the
	// endpoint to update the target list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// BasicAuth information to authenticate against the target HTTP endpoint.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration to authenticate against the target HTTP endpoint.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
}

// KubernetesSDConfig allows retrieving scrape targets from Kubernetes' REST API.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
// +k8s:openapi-gen=true
type KubernetesSDConfig struct {
	// The API server address consisting of a hostname or IP address followed
	// by an optional port number.
	// If left empty, Prometheus is assumed to run inside
	// of the cluster. It will discover API servers automatically and use the pod's
	// CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	// +optional
	APIServer *string `json:"apiServer,omitempty"`
	// Role of the Kubernetes entities that should be discovered.
	// +required
	Role Role `json:"role"`
	// BasicAuth information to use on every scrape request.
	// Cannot be set at the same time as `authorization`, or `oauth2`.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// Cannot be set at the same time as `basicAuth`, or `oauth2`.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization`, or `basicAuth`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
	// TLS configuration to use on every scrape request.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// Optional namespace discovery. If omitted, Prometheus discovers targets across all namespaces.
	// +optional
	Namespaces *NamespaceDiscovery `json:"namespaces,omitempty"`
	// Optional metadata to attach to discovered targets.
	// It requires Prometheus >= v2.35.0 for `pod` role and
	// Prometheus >= v2.37.0 for `endpoints` and `endpointslice` roles.
	// +optional
	AttachMetadata *AttachMetadata `json:"attachMetadata,omitempty"`
	// Selector to select objects.
	// +optional
	// +listType=map
	// +listMapKey=role
	Selectors []K8SSelectorConfig `json:"selectors,omitempty"`
}

// ConsulSDConfig defines a Consul service discovery configuration
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config
// +k8s:openapi-gen=true
type ConsulSDConfig struct {
	// A valid string consisting of a hostname or IP followed by an optional port number.
	// +kubebuilder:validation:MinLength=1
	// +required
	Server string `json:"server"`
	// Consul ACL TokenRef, if not provided it will use the ACL from the local Consul Agent.
	// +optional
	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
	// Consul Datacenter name, if not provided it will use the local Consul Agent Datacenter.
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`
	// Namespaces are only supported in Consul Enterprise.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Admin Partitions are only supported in Consul Enterprise.
	// +optional
	Partition *string `json:"partition,omitempty"`
	// HTTP Scheme default "http"
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	// +optional
	Scheme *string `json:"scheme,omitempty"`
	// A list of services for which targets are retrieved. If omitted, all services are scraped.
	// +listType:=atomic
	// +optional
	Services []string `json:"services,omitempty"`
	// An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list.
	//+listType:=atomic
	// +optional
	Tags []string `json:"tags,omitempty"`
	// The string by which Consul tags are joined into the tag label.
	// If unset, Prometheus uses its default value.
	// +optional
	TagSeparator *string `json:"tagSeparator,omitempty"`
	// Node metadata key/value pairs to filter nodes for a given service.
	// +mapType:=atomic
	// +optional
	NodeMeta map[string]string `json:"nodeMeta,omitempty"`
	// Allow stale Consul results (see https://www.consul.io/api/features/consistency.html). Will reduce load on Consul.
	// If unset, Prometheus uses its default value.
	// +optional
	AllowStale *bool `json:"allowStale,omitempty"`
	// The time after which the provided names are refreshed.
	// On large setup it might be a good idea to increase this value because the catalog will change all the time.
	// If unset, Prometheus uses its default value.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// BasicAuth information to authenticate against the Consul Server.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoints
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration to authenticate against the Consul Server.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// +optional
	Oauth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// If unset, Prometheus uses its default value.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// If unset, Prometheus uses its default value.
	// +optional
	EnableHttp2 *bool `json:"enableHTTP2,omitempty"`
	// TLS Config
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// DNSSDConfig allows specifying a set of DNS domain names which are periodically queried to discover a list of targets.
// The DNS servers to be contacted are read from /etc/resolv.conf.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dns_sd_config
// +k8s:openapi-gen=true
type DNSSDConfig struct {
	// A list of DNS domain names to be queried.
	// +kubebuilder:validation:MinItems:=1
	Names []string `json:"names"`
	// RefreshInterval configures the time after which the provided names are refreshed.
	// If not set, Prometheus uses its default value.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The type of DNS query to perform. One of SRV, A, AAAA, MX or NS.
	// If not set, Prometheus uses its default value.
	//
	// When set to NS, It requires Prometheus >= 2.49.0.
	//
	// +kubebuilder:validation:Enum=SRV;A;AAAA;MX;NS
	// +optional
	Type *string `json:"type"`
	// The port number used if the query type is not SRV
	// Ignored for SRV records
	// +optional
	Port *int `json:"port"`
}

// EC2SDConfig allow retrieving scrape targets from AWS EC2 instances.
// The private IP address is used by default, but may be changed to the public IP address with relabeling.
// The IAM credentials used must have the ec2:DescribeInstances permission to discover scrape targets
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config
// +k8s:openapi-gen=true
type EC2SDConfig struct {
	// The AWS region
	// +optional
	Region *string `json:"region"`
	// AccessKey is the AWS API key.
	// +optional
	AccessKey *corev1.SecretKeySelector `json:"accessKey,omitempty"`
	// SecretKey is the AWS API secret.
	// +optional
	SecretKey *corev1.SecretKeySelector `json:"secretKey,omitempty"`
	// AWS Role ARN, an alternative to using AWS API keys.
	// +optional
	RoleARN *string `json:"roleARN,omitempty"`
	// RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The port to scrape metrics from. If using the public IP address, this must
	// instead be specified in the relabeling rule.
	// +optional
	Port *int `json:"port"`
	// Filters can be used optionally to filter the instance list by other criteria.
	// Available filter criteria can be found here:
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
	// Filter API documentation: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html
	// +optional
	Filters []*EC2Filter `json:"filters"`
}

// AzureSDConfig allow retrieving scrape targets from Azure VMs.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#azure_sd_config
// +k8s:openapi-gen=true
type AzureSDConfig struct {
	// The Azure environment.
	// +optional
	Environment *string `json:"environment,omitempty"`
	// # The authentication method, either `OAuth` or `ManagedIdentity` or `SDK`.
	// See https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview
	// SDK authentication method uses environment variables by default.
	// See https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication
	// +kubebuilder:validation:Enum=OAuth;ManagedIdentity;SDK
	// +optional
	AuthenticationMethod *string `json:"authenticationMethod,omitempty"`
	// The subscription ID. Always required.
	// +kubebuilder:validation:MinLength=1
	// +required
	SubscriptionID string `json:"subscriptionID"`
	// Optional tenant ID. Only required with the OAuth authentication method.
	// +optional
	TenantID *string `json:"tenantID,omitempty"`
	// Optional client ID. Only required with the OAuth authentication method.
	// +optional
	ClientID *string `json:"clientID,omitempty"`
	// Optional client secret. Only required with the OAuth authentication method.
	// +optional
	ClientSecret *corev1.SecretKeySelector `json:"clientSecret,omitempty"`
	// Optional resource group name. Limits discovery to this resource group.
	// +optional
	ResourceGroup *string `json:"resourceGroup,omitempty"`
	// RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The port to scrape metrics from. If using the public IP address, this must
	// instead be specified in the relabeling rule.
	// +optional
	Port *int `json:"port"`
}

// GCESDConfig configures scrape targets from GCP GCE instances.
// The private IP address is used by default, but may be changed to
// the public IP address with relabeling.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config
//
// The GCE service discovery will load the Google Cloud credentials
// from the file specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable.
// See https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform
//
// A pre-requisite for using GCESDConfig is that a Secret containing valid
// Google Cloud credentials is mounted into the Prometheus or PrometheusAgent
// pod via the `.spec.secrets` field and that the GOOGLE_APPLICATION_CREDENTIALS
// environment variable is set to /etc/prometheus/secrets/<secret-name>/<credentials-filename.json>.
// +k8s:openapi-gen=true
type GCESDConfig struct {
	// The Google Cloud Project ID
	// +kubebuilder:validation:MinLength:=1
	// +required
	Project string `json:"project"`
	// The zone of the scrape targets. If you need multiple zones use multiple GCESDConfigs.
	// +kubebuilder:validation:MinLength:=1
	// +required
	Zone string `json:"zone"`
	// Filter can be used optionally to filter the instance list by other criteria
	// Syntax of this filter is described in the filter query parameter section:
	// https://cloud.google.com/compute/docs/reference/latest/instances/list
	// +optional
	Filter *string `json:"filter,omitempty"`
	// RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The port to scrape metrics from. If using the public IP address, this must
	// instead be specified in the relabeling rule.
	// +optional
	Port *int `json:"port"`
	// The tag separator is used to separate the tags on concatenation
	// +optional
	TagSeparator *string `json:"tagSeparator,omitempty"`
}

// OpenStackSDConfig allow retrieving scrape targets from OpenStack Nova instances.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#openstack_sd_config
// +k8s:openapi-gen=true
type OpenStackSDConfig struct {
	// The OpenStack role of entities that should be discovered.
	// +kubebuilder:validation:Enum=Instance;instance;Hypervisor;hypervisor
	// +required
	Role string `json:"role"`
	// The OpenStack Region.
	// +kubebuilder:validation:MinLength:=1
	// +required
	Region string `json:"region"`
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version.
	// +optional
	IdentityEndpoint *string `json:"identityEndpoint,omitempty"`
	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username.
	// In Identity V3, either userid or a combination of username
	// and domainId or domainName are needed
	// +optional
	Username *string `json:"username,omitempty"`
	// UserID
	// +optional
	UserID *string `json:"userid,omitempty"`
	// Password for the Identity V2 and V3 APIs. Consult with your provider's
	// control panel to discover your account's preferred method of authentication.
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty"`
	// At most one of domainId and domainName must be provided if using username
	// with Identity V3. Otherwise, either are optional.
	// +optional
	DomainName *string `json:"domainName,omitempty"`
	// DomainID
	// +optional
	DomainID *string `json:"domainID,omitempty"`
	// The ProjectId and ProjectName fields are optional for the Identity V2 API.
	// Some providers allow you to specify a ProjectName instead of the ProjectId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	// +optional
	ProjectName *string `json:"projectName,omitempty"`
	//  ProjectID
	// +optional
	ProjectID *string `json:"projectID,omitempty"`
	// The ApplicationCredentialID or ApplicationCredentialName fields are
	// required if using an application credential to authenticate. Some providers
	// allow you to create an application credential to authenticate rather than a
	// password.
	// +optional
	ApplicationCredentialName *string `json:"applicationCredentialName,omitempty"`
	// ApplicationCredentialID
	// +optional
	ApplicationCredentialID *string `json:"applicationCredentialId,omitempty"`
	// The applicationCredentialSecret field is required if using an application
	// credential to authenticate.
	// +optional
	ApplicationCredentialSecret *corev1.SecretKeySelector `json:"applicationCredentialSecret,omitempty"`
	// Whether the service discovery should list all instances for all projects.
	// It is only relevant for the 'instance' role and usually requires admin permissions.
	// +optional
	AllTenants *bool `json:"allTenants,omitempty"`
	// Refresh interval to re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The port to scrape metrics from. If using the public IP address, this must
	// instead be specified in the relabeling rule.
	// +optional
	Port *int `json:"port"`
	// Availability of the endpoint to connect to.
	// +kubebuilder:validation:Enum=Public;public;Admin;admin;Internal;internal
	// +optional
	Availability *string `json:"availability,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// DigitalOceanSDConfig allow retrieving scrape targets from DigitalOcean's Droplets API.
// This service discovery uses the public IPv4 address by default, by that can be changed with relabeling
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#digitalocean_sd_config
// +k8s:openapi-gen=true
type DigitalOceanSDConfig struct {
	// Authorization header configuration to authenticate against the DigitalOcean API.
	// Cannot be set at the same time as `oauth2`.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// The port to scrape metrics from.
	// +optional
	Port *int `json:"port,omitempty"`
	// Refresh interval to re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
}

// KumaSDConfig allow retrieving scrape targets from Kuma's control plane.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kuma_sd_config
// +k8s:openapi-gen=true
type KumaSDConfig struct {
	// Address of the Kuma Control Plane's MADS xDS server.
	// +kubebuilder:validation:MinLength=1
	// +required
	Server string `json:"server"`
	// Client id is used by Kuma Control Plane to compute Monitoring Assignment for specific Prometheus backend.
	// +optional
	ClientID *string `json:"clientID,omitempty"`
	// The time to wait between polling update requests.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// The time after which the monitoring assignments are refreshed.
	// +optional
	FetchTimeout *v1.Duration `json:"fetchTimeout,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// TLS configuration to use on every scrape request
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization`, or `basicAuth`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
}

// Eureka SD configurations allow retrieving scrape targets using the Eureka REST API.
// Prometheus will periodically check the REST endpoint and create a target for every app instance.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#eureka_sd_config
// +k8s:openapi-gen=true
type EurekaSDConfig struct {
	// The URL to connect to the Eureka server.
	// +kubebuilder:validation:MinLength=1
	// +required
	Server string `json:"server"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization` or `basic_auth`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
	// Refresh interval to re-read the instance list.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
}

// Docker SD configurations allow retrieving scrape targets from Docker Engine hosts.
// This SD discovers "containers" and will create a target for each network IP and
// port the container is configured to expose.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#docker_sd_config
// +k8s:openapi-gen=true
type DockerSDConfig struct {
	// Address of the docker daemon
	// +kubebuilder:validation:MinLength=1
	// +required
	Host string `json:"host"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// The port to scrape metrics from.
	// +optional
	Port *int `json:"port,omitempty"`
	// The host to use if the container is in host networking mode.
	// +optional
	HostNetworkingHost *string `json:"hostNetworkingHost,omitempty"`
	// Optional filters to limit the discovery process to a subset of the available resources.
	// +optional
	Filters *[]DockerFilter `json:"filters,omitempty"`
	// Time after which the container is refreshed.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration to authenticate against the Docker API.
	// Cannot be set at the same time as `oauth2`.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
}

// HetznerSDConfig allow retrieving scrape targets from Hetzner Cloud API and Robot API.
// This service discovery uses the public IPv4 address by default, but that can be changed with relabeling
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#hetzner_sd_config
// +k8s:openapi-gen=true
type HetznerSDConfig struct {
	// The Hetzner role of entities that should be discovered.
	// +kubebuilder:validation:Enum=hcloud;Hcloud;robot;Robot
	// +required
	Role string `json:"role"`
	// BasicAuth information to use on every scrape request, required when role is robot.
	// Role hcloud does not support basic auth.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header configuration, required when role is hcloud.
	// Role robot does not support bearer token authentication.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be used at the same time as `basic_auth` or `authorization`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
	// TLS configuration to use on every scrape request.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// The port to scrape metrics from.
	// +optional
	Port *int `json:"port,omitempty"`
	// The time after which the servers are refreshed.
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
}

// NomadSDConfig configurations allow retrieving scrape targets from Nomad's Service API.
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#nomad_sd_config
// +k8s:openapi-gen=true
type NomadSDConfig struct {
	// The information to access the Nomad API. It is to be defined
	// as the Nomad documentation requires.
	// +optional
	AllowStale *bool `json:"allowStale,omitempty"`
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// +optional
	RefreshInterval *v1.Duration `json:"refreshInterval,omitempty"`
	// +optional
	Region *string `json:"region,omitempty"`
	// +kubebuilder:validation:MinLength=1
	// +required
	Server string `json:"server"`
	// +optional
	TagSeparator *string `json:"tagSeparator,omitempty"`
	// BasicAuth information to use on every scrape request.
	// +optional
	BasicAuth *v1.BasicAuth `json:"basicAuth,omitempty"`
	// Authorization header to use on every scrape request.
	// +optional
	Authorization *v1.SafeAuthorization `json:"authorization,omitempty"`
	// Optional OAuth 2.0 configuration.
	// Cannot be set at the same time as `authorization` or `basic_auth`.
	// +optional
	OAuth2 *v1.OAuth2 `json:"oauth2,omitempty"`
	// TLS configuration applying to the target HTTP endpoint.
	// +optional
	TLSConfig *v1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// ProxyConfig allows customizing the proxy behaviour for this scrape config.
	// +optional
	v1.ProxyConfig `json:",inline"`
	// Configure whether HTTP requests follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Whether to enable HTTP2.
	// +optional
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
}
