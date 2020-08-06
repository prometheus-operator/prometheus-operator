<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.
</div>

# API Docs

This Document documents the types introduced by the Prometheus Operator to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## Table of Contents
* [APIServerConfig](#apiserverconfig)
* [AlertingSpec](#alertingspec)
* [Alertmanager](#alertmanager)
* [AlertmanagerEndpoints](#alertmanagerendpoints)
* [AlertmanagerList](#alertmanagerlist)
* [AlertmanagerSpec](#alertmanagerspec)
* [AlertmanagerStatus](#alertmanagerstatus)
* [ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig)
* [BasicAuth](#basicauth)
* [EmbeddedObjectMetadata](#embeddedobjectmetadata)
* [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim)
* [Endpoint](#endpoint)
* [NamespaceSelector](#namespaceselector)
* [PodMetricsEndpoint](#podmetricsendpoint)
* [PodMonitor](#podmonitor)
* [PodMonitorList](#podmonitorlist)
* [PodMonitorSpec](#podmonitorspec)
* [Probe](#probe)
* [ProbeList](#probelist)
* [ProbeSpec](#probespec)
* [ProbeTargetIngress](#probetargetingress)
* [ProbeTargetStaticConfig](#probetargetstaticconfig)
* [ProbeTargets](#probetargets)
* [ProberSpec](#proberspec)
* [Prometheus](#prometheus)
* [PrometheusList](#prometheuslist)
* [PrometheusRule](#prometheusrule)
* [PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig)
* [PrometheusRuleList](#prometheusrulelist)
* [PrometheusRuleSpec](#prometheusrulespec)
* [PrometheusSpec](#prometheusspec)
* [PrometheusStatus](#prometheusstatus)
* [QuerySpec](#queryspec)
* [QueueConfig](#queueconfig)
* [RelabelConfig](#relabelconfig)
* [RemoteReadSpec](#remotereadspec)
* [RemoteWriteSpec](#remotewritespec)
* [Rule](#rule)
* [RuleGroup](#rulegroup)
* [Rules](#rules)
* [RulesAlert](#rulesalert)
* [SecretOrConfigMap](#secretorconfigmap)
* [ServiceMonitor](#servicemonitor)
* [ServiceMonitorList](#servicemonitorlist)
* [ServiceMonitorSpec](#servicemonitorspec)
* [StorageSpec](#storagespec)
* [TLSConfig](#tlsconfig)
* [ThanosSpec](#thanosspec)
* [ThanosRuler](#thanosruler)
* [ThanosRulerList](#thanosrulerlist)
* [ThanosRulerSpec](#thanosrulerspec)
* [ThanosRulerStatus](#thanosrulerstatus)

## APIServerConfig

APIServerConfig defines a host and auth methods to access apiserver. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host | Host of apiserver. A valid string consisting of a hostname or IP followed by an optional port number | string | true |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication | *[BasicAuth](#basicauth) | false |
| bearerToken | Bearer token for accessing apiserver. | string | false |
| bearerTokenFile | File to read bearer token for accessing apiserver. | string | false |
| tlsConfig | TLS Config to use for accessing apiserver. | *[TLSConfig](#tlsconfig) | false |

[Back to TOC](#table-of-contents)

## AlertingSpec

AlertingSpec defines parameters for alerting configuration of Prometheus servers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| alertmanagers | AlertmanagerEndpoints Prometheus should fire alerts against. | [][AlertmanagerEndpoints](#alertmanagerendpoints) | true |

[Back to TOC](#table-of-contents)

## Alertmanager

Alertmanager describes an Alertmanager cluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [AlertmanagerSpec](#alertmanagerspec) | true |
| status | Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[AlertmanagerStatus](#alertmanagerstatus) | false |

[Back to TOC](#table-of-contents)

## AlertmanagerEndpoints

AlertmanagerEndpoints defines a selection of a single Endpoints object containing alertmanager IPs to fire alerts against.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | Namespace of Endpoints object. | string | true |
| name | Name of Endpoints object in Namespace. | string | true |
| port | Port the Alertmanager API is exposed on. | intstr.IntOrString | true |
| scheme | Scheme to use when firing alerts. | string | false |
| pathPrefix | Prefix for the HTTP path alerts are pushed to. | string | false |
| tlsConfig | TLS Config to use for alertmanager connection. | *[TLSConfig](#tlsconfig) | false |
| bearerTokenFile | BearerTokenFile to read from filesystem to use when authenticating to Alertmanager. | string | false |
| apiVersion | Version of the Alertmanager API that Prometheus uses to send alerts. It can be \"v1\" or \"v2\". | string | false |
| timeout | Timeout is a per-target Alertmanager timeout when pushing alerts. | *string | false |

[Back to TOC](#table-of-contents)

## AlertmanagerList

AlertmanagerList is a list of Alertmanagers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of Alertmanagers | [][Alertmanager](#alertmanager) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerSpec

AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Alertmanager is being configured. | *string | false |
| version | Version the cluster should be on. | string | false |
| tag | Tag of Alertmanager container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | string | false |
| sha | SHA of Alertmanager container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | string | false |
| baseImage | Base image that is used to deploy pods, without tag. Deprecated: use 'image' instead | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>. | []string | false |
| configSecret | ConfigSecret is the name of a Kubernetes Secret in the same namespace as the Alertmanager object, which contains configuration for this Alertmanager instance. Defaults to 'alertmanager-<alertmanager-name>' The secret is mounted into /etc/alertmanager/config. | string | false |
| logLevel | Log level for Alertmanager to be configured with. | string | false |
| logFormat | Log format for Alertmanager to be configured with. | string | false |
| replicas | Size is the expected size of the alertmanager cluster. The controller will eventually make the size of the running cluster equal to the expected size. | *int32 | false |
| retention | Time duration Alertmanager shall retain data for. Default is '120h', and must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds seconds minutes hours). | string | false |
| storage | Storage is the definition of how storage will be used by the Alertmanager instances. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| volumeMounts | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container, that are generated as a result of StorageSpec objects. | []v1.VolumeMount | false |
| externalUrl | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Alertmanager registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| paused | If set to true all actions on the underlaying managed objects are not goint to be performed, except for delete actions. | bool | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| listenLocal | ListenLocal makes the Alertmanager server listen on loopback, so that it does not bind against the Pod IP. Note this is only for the Alertmanager UI, not the gossip communication. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to an Alertmanager pod. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the Alertmanager configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other then secret fetching is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| additionalPeers | AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster. | []string | false |
| clusterAdvertiseAddress | ClusterAdvertiseAddress is the explicit address to advertise in cluster. Needs to be provided for non RFC1918 [1] (public) addresses. [1] RFC1918: https://tools.ietf.org/html/rfc1918 | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| forceEnableClusterMode | ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica. Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each. | bool | false |

[Back to TOC](#table-of-contents)

## AlertmanagerStatus

AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Alertmanager cluster (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Alertmanager cluster that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Alertmanager cluster. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Alertmanager cluster. | int32 | true |

[Back to TOC](#table-of-contents)

## ArbitraryFSAccessThroughSMsConfig

ArbitraryFSAccessThroughSMsConfig enables users to configure, whether a service monitor selected by the Prometheus instance is allowed to use arbitrary files on the file system of the Prometheus container. This is the case when e.g. a service monitor specifies a BearerTokenFile in an endpoint. A malicious user could create a service monitor selecting arbitrary secret files in the Prometheus container. Those secrets would then be sent with a scrape request by Prometheus to a malicious target. Denying the above would prevent the attack, users can instead use the BearerTokenSecret field.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deny |  | bool | false |

[Back to TOC](#table-of-contents)

## BasicAuth

BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username | The secret in the service monitor namespace that contains the username for authentication. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| password | The secret in the service monitor namespace that contains the password for authentication. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## EmbeddedObjectMetadata

EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta Only fields which are relevant to embedded resources are included.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names | string | false |
| labels | Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels | map[string]string | false |
| annotations | Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations | map[string]string | false |

[Back to TOC](#table-of-contents)

## EmbeddedPersistentVolumeClaim

EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim. It contains TypeMeta and a reduced ObjectMeta.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | EmbeddedMetadata contains metadata relevant to an EmbeddedResource. | [EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| spec | Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims | v1.PersistentVolumeClaimSpec | false |
| status | Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims | v1.PersistentVolumeClaimStatus | false |

[Back to TOC](#table-of-contents)

## Endpoint

Endpoint defines a scrapeable endpoint serving Prometheus metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the service port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Name or number of the target port of the Pod behind the Service, the port must be specified with container port property. Mutually exclusive with port. | *intstr.IntOrString | false |
| path | HTTP path to scrape for metrics. | string | false |
| scheme | HTTP scheme to use for scraping. | string | false |
| params | Optional HTTP URL parameters | map[string][]string | false |
| interval | Interval at which metrics should be scraped | string | false |
| scrapeTimeout | Timeout after which the scrape is ended | string | false |
| tlsConfig | TLS configuration to use when scraping the endpoint | *[TLSConfig](#tlsconfig) | false |
| bearerTokenFile | File to read bearer token for scraping targets. | string | false |
| bearerTokenSecret | Secret to mount to read bearer token for scraping targets. The secret needs to be in the same namespace as the service monitor and accessible by the Prometheus Operator. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| honorLabels | HonorLabels chooses the metric's labels on collisions with target labels. | bool | false |
| honorTimestamps | HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data. | *bool | false |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints | *[BasicAuth](#basicauth) | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| relabelings | RelabelConfigs to apply to samples before scraping. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |

[Back to TOC](#table-of-contents)

## NamespaceSelector

NamespaceSelector is a selector for selecting either all namespaces or a list of namespaces.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| any | Boolean describing whether all namespaces are selected in contrast to a list restricting them. | bool | false |
| matchNames | List of namespace names. | []string | false |

[Back to TOC](#table-of-contents)

## PodMetricsEndpoint

PodMetricsEndpoint defines a scrapeable endpoint of a Kubernetes Pod serving Prometheus metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the pod port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Deprecated: Use 'port' instead. | *intstr.IntOrString | false |
| path | HTTP path to scrape for metrics. | string | false |
| scheme | HTTP scheme to use for scraping. | string | false |
| params | Optional HTTP URL parameters | map[string][]string | false |
| interval | Interval at which metrics should be scraped | string | false |
| scrapeTimeout | Timeout after which the scrape is ended | string | false |
| honorLabels | HonorLabels chooses the metric's labels on collisions with target labels. | bool | false |
| honorTimestamps | HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data. | *bool | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| relabelings | RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |

[Back to TOC](#table-of-contents)

## PodMonitor

PodMonitor defines monitoring for a set of pods.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of desired Pod selection for target discovery by Prometheus. | [PodMonitorSpec](#podmonitorspec) | true |

[Back to TOC](#table-of-contents)

## PodMonitorList

PodMonitorList is a list of PodMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of PodMonitors | []*[PodMonitor](#podmonitor) | true |

[Back to TOC](#table-of-contents)

## PodMonitorSpec

PodMonitorSpec contains specification parameters for a PodMonitor.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | The label to use to retrieve the job name from. | string | false |
| podTargetLabels | PodTargetLabels transfers labels on the Kubernetes Pod onto the target. | []string | false |
| podMetricsEndpoints | A list of endpoints allowed as part of this PodMonitor. | [][PodMetricsEndpoint](#podmetricsendpoint) | true |
| selector | Selector to select Pod objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |

[Back to TOC](#table-of-contents)

## Probe

Probe defines monitoring for a set of static targets or ingresses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of desired Ingress selection for target discovery by Prometheus. | [ProbeSpec](#probespec) | true |

[Back to TOC](#table-of-contents)

## ProbeList

ProbeList is a list of Probes.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of Probes | []*[Probe](#probe) | true |

[Back to TOC](#table-of-contents)

## ProbeSpec

ProbeSpec contains specification parameters for a Probe.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobName | The job name assigned to scraped metrics by default. | string | false |
| prober | Specification for the prober to use for probing targets. The prober.URL parameter is required. Targets cannot be probed if left empty. | [ProberSpec](#proberspec) | false |
| module | The module to use for probing specifying how to probe the target. Example module configuring in the blackbox exporter: https://github.com/prometheus/blackbox_exporter/blob/master/example.yml | string | false |
| targets | Targets defines a set of static and/or dynamically discovered targets to be probed using the prober. | [ProbeTargets](#probetargets) | false |
| interval | Interval at which targets are probed using the configured prober. If not specified Prometheus' global scrape interval is used. | string | false |
| scrapeTimeout | Timeout for scraping metrics from the Prometheus exporter. | string | false |

[Back to TOC](#table-of-contents)

## ProbeTargetIngress

ProbeTargetIngress defines the set of Ingress objects considered for probing.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| selector | Select Ingress objects by labels. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| namespaceSelector | Select Ingress objects by namespace. | [NamespaceSelector](#namespaceselector) | false |
| relabelingConfigs | RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |

[Back to TOC](#table-of-contents)

## ProbeTargetStaticConfig

ProbeTargetStaticConfig defines the set of static targets considered for probing.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| static | Targets is a list of URLs to probe using the configured prober. | []string | false |
| labels | Labels assigned to all metrics scraped from the targets. | map[string]string | false |

[Back to TOC](#table-of-contents)

## ProbeTargets

ProbeTargets defines a set of static and dynamically discovered targets for the prober.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| staticConfig | StaticConfig defines static targets which are considers for probing. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config. | *[ProbeTargetStaticConfig](#probetargetstaticconfig) | false |
| ingress | Ingress defines the set of dynamically discovered ingress objects which hosts are considered for probing. | *[ProbeTargetIngress](#probetargetingress) | false |

[Back to TOC](#table-of-contents)

## ProberSpec

ProberSpec contains specification parameters for the Prober used for probing.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | Mandatory URL of the prober. | string | true |
| scheme | HTTP scheme to use for scraping. Defaults to `http`. | string | false |
| path | Path to collect metrics from. Defaults to `/probe`. | string | false |

[Back to TOC](#table-of-contents)

## Prometheus

Prometheus defines a Prometheus deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [PrometheusSpec](#prometheusspec) | true |
| status | Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[PrometheusStatus](#prometheusstatus) | false |

[Back to TOC](#table-of-contents)

## PrometheusList

PrometheusList is a list of Prometheuses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[Prometheus](#prometheus) | true |

[Back to TOC](#table-of-contents)

## PrometheusRule

PrometheusRule defines alerting rules for a Prometheus instance

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of desired alerting rule definitions for Prometheus. | [PrometheusRuleSpec](#prometheusrulespec) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleExcludeConfig

PrometheusRuleExcludeConfig enables users to configure excluded PrometheusRule names and their namespaces to be ignored while enforcing namespace label for alerts and metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| ruleNamespace | RuleNamespace - namespace of excluded rule | string | true |
| ruleName | RuleNamespace - name of excluded rule | string | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleList

PrometheusRuleList is a list of PrometheusRules.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of Rules | []*[PrometheusRule](#prometheusrule) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleSpec

PrometheusRuleSpec contains specification parameters for a Rule.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| groups | Content of Prometheus rule file | [][RuleGroup](#rulegroup) | false |

[Back to TOC](#table-of-contents)

## PrometheusSpec

PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata configures Labels and Annotations which are propagated to the prometheus pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| serviceMonitorSelector | ServiceMonitors to be selected for target discovery. *Deprecated:* if neither this nor podMonitorSelector are specified, configuration is unmanaged. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| serviceMonitorNamespaceSelector | Namespaces to be selected for ServiceMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| podMonitorSelector | *Experimental* PodMonitors to be selected for target discovery. *Deprecated:* if neither this nor serviceMonitorSelector are specified, configuration is unmanaged. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| podMonitorNamespaceSelector | Namespaces to be selected for PodMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| probeSelector | *Experimental* Probes to be selected for target discovery. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| probeNamespaceSelector | *Experimental* Namespaces to be selected for Probe discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| version | Version of Prometheus to be deployed. | string | false |
| tag | Tag of Prometheus container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | string | false |
| sha | SHA of Prometheus container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | string | false |
| paused | When a Prometheus deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Prometheus is being configured. | *string | false |
| baseImage | Base image to use for a Prometheus deployment. Deprecated: use 'image' instead | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) | false |
| replicas | Number of instances to deploy for a Prometheus deployment. | *int32 | false |
| replicaExternalLabelName | Name of Prometheus external label used to denote replica name. Defaults to the value of `prometheus_replica`. External label will _not_ be added when value is set to empty string (`\"\"`). | *string | false |
| prometheusExternalLabelName | Name of Prometheus external label used to denote Prometheus instance name. Defaults to the value of `prometheus`. External label will _not_ be added when value is set to empty string (`\"\"`). | *string | false |
| retention | Time duration Prometheus shall retain data for. Default is '24h', and must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | string | false |
| retentionSize | Maximum amount of disk space used by blocks. | string | false |
| disableCompaction | Disable prometheus compaction. | bool | false |
| walCompression | Enable compression of the write-ahead log using Snappy. This flag is only available in versions of Prometheus >= 2.11.0. | *bool | false |
| logLevel | Log level for Prometheus to be configured with. | string | false |
| logFormat | Log format for Prometheus to be configured with. | string | false |
| scrapeInterval | Interval between consecutive scrapes. | string | false |
| scrapeTimeout | Number of seconds to wait for target to respond before erroring. | string | false |
| evaluationInterval | Interval between consecutive evaluations. | string | false |
| rules | /--rules.*/ command-line arguments. | [Rules](#rules) | false |
| externalLabels | The labels to add to any time series or alerts when communicating with external systems (federation, remote storage, Alertmanager). | map[string]string | false |
| enableAdminAPI | Enable access to prometheus web admin API. Defaults to the value of `false`. WARNING: Enabling the admin APIs enables mutating endpoints, to delete data, shutdown Prometheus, and more. Enabling this should be done with care and the user is advised to add additional authentication authorization via a proxy to ensure only clients authorized to perform these actions can do so. For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis | bool | false |
| externalUrl | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Prometheus registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| query | QuerySpec defines the query command line flags when starting Prometheus. | *[QuerySpec](#queryspec) | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| volumeMounts | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumeMounts specified will be appended to other VolumeMounts in the prometheus container, that are generated as a result of StorageSpec objects. | []v1.VolumeMount | false |
| ruleSelector | A selector to select which PrometheusRules to mount for loading alerting/recording rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom resources selected by RuleSelector. Make sure it does not match any config maps that you do not want to be migrated. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for PrometheusRules discovery. If unspecified, only the same namespace as the Prometheus object is in is used. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| alerting | Define details regarding alerting. | *[AlertingSpec](#alertingspec) | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The Secrets are mounted into /etc/prometheus/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>. | []string | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| remoteWrite | If specified, the remote_write spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteWriteSpec](#remotewritespec) | false |
| remoteRead | If specified, the remote_read spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteReadSpec](#remotereadspec) | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| listenLocal | ListenLocal makes the Prometheus server listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| containers | Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a Prometheus pod or to change the behavior of an operator generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `prometheus`, `prometheus-config-reloader`, `rules-configmap-reloader`, and `thanos-sidecar`. Overriding containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the Prometheus configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other then secret fetching is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| additionalScrapeConfigs | AdditionalScrapeConfigs allows specifying a key of a Secret containing additional Prometheus scrape configurations. Scrape configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config. As scrape configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible scrape configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| additionalAlertRelabelConfigs | AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing additional Prometheus alert relabel configurations. Alert relabel configurations specified are appended to the configurations generated by the Prometheus Operator. Alert relabel configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs. As alert relabel configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible alert relabel configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| additionalAlertManagerConfigs | AdditionalAlertManagerConfigs allows specifying a key of a Secret containing additional Prometheus AlertManager configurations. AlertManager configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config. As AlertManager configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible AlertManager configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| apiserverConfig | APIServerConfig allows specifying a host and auth methods to access apiserver. If left empty, Prometheus is assumed to run inside of the cluster and will discover API servers automatically and use the pod's CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. | *[APIServerConfig](#apiserverconfig) | false |
| thanos | Thanos configuration allows configuring various aspects of a Prometheus server in a Thanos environment.\n\nThis section is experimental, it may change significantly without deprecation notice in any release.\n\nThis is experimental and may change significantly without backward compatibility in any release. | *[ThanosSpec](#thanosspec) | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| arbitraryFSAccessThroughSMs | ArbitraryFSAccessThroughSMs configures whether configuration based on a service monitor can access arbitrary files on the file system of the Prometheus container e.g. bearer token files. | [ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig) | false |
| overrideHonorLabels | OverrideHonorLabels if set to true overrides all user configured honor_labels. If HonorLabels is set in ServiceMonitor or PodMonitor to true, this overrides honor_labels to false. | bool | false |
| overrideHonorTimestamps | OverrideHonorTimestamps allows to globally enforce honoring timestamps in all scrape configs. | bool | false |
| ignoreNamespaceSelectors | IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector settings from the podmonitor and servicemonitor configs, and they will only discover endpoints within their current namespace.  Defaults to false. | bool | false |
| enforcedNamespaceLabel | EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert and metric that is user created. The label value will always be the namespace of the object that is being created. | string | false |
| prometheusRulesExcludedFromEnforce | PrometheusRulesExcludedFromEnforce - list of prometheus rules to be excluded from enforcing of adding namespace labels. Works only if enforcedNamespaceLabel set to true. Make sure both ruleNamespace and ruleName are set for each pair | [][PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) | false |
| queryLogFile | QueryLogFile specifies the file to which PromQL queries are logged. Note that this location must be writable, and can be persisted using an attached volume. Alternatively, the location can be set to a stdout location such as `/dev/stdout` to log querie information to the default Prometheus log stream. This is only available in versions of Prometheus >= 2.16.0. For more details, see the Prometheus docs (https://prometheus.io/docs/guides/query-log/) | string | false |
| enforcedSampleLimit | EnforcedSampleLimit defines global limit on number of scraped samples that will be accepted. This overrides any SampleLimit set per ServiceMonitor or/and PodMonitor. It is meant to be used by admins to enforce the SampleLimit to keep overall number of samples/series under the desired limit. Note that if SampleLimit is lower that value will be taken instead. | *uint64 | false |
| allowOverlappingBlocks | AllowOverlappingBlocks enables vertical compaction and vertical query merge in Prometheus. This is still experimental in Prometheus so it may change in any upcoming release. | bool | false |

[Back to TOC](#table-of-contents)

## PrometheusStatus

PrometheusStatus is the most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Prometheus deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Prometheus deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Prometheus deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Prometheus deployment. | int32 | true |

[Back to TOC](#table-of-contents)

## QuerySpec

QuerySpec defines the query command line flags when starting Prometheus.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lookbackDelta | The delta difference allowed for retrieving metrics during expression evaluations. | *string | false |
| maxConcurrency | Number of concurrent queries that can be run at once. | *int32 | false |
| maxSamples | Maximum number of samples a single query can load into memory. Note that queries will fail if they would load more samples than this into memory, so this also limits the number of samples a query can return. | *int32 | false |
| timeout | Maximum time a query may take before being aborted. | *string | false |

[Back to TOC](#table-of-contents)

## QueueConfig

QueueConfig allows the tuning of remote_write queue_config parameters. This object is referenced in the RemoteWriteSpec object.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| capacity | Capacity is the number of samples to buffer per shard before we start dropping them. | int | false |
| minShards | MinShards is the minimum number of shards, i.e. amount of concurrency. | int | false |
| maxShards | MaxShards is the maximum number of shards, i.e. amount of concurrency. | int | false |
| maxSamplesPerSend | MaxSamplesPerSend is the maximum number of samples per send. | int | false |
| batchSendDeadline | BatchSendDeadline is the maximum time a sample will wait in buffer. | string | false |
| maxRetries | MaxRetries is the maximum number of times to retry a batch on recoverable errors. | int | false |
| minBackoff | MinBackoff is the initial retry delay. Gets doubled for every retry. | string | false |
| maxBackoff | MaxBackoff is the maximum retry delay. | string | false |

[Back to TOC](#table-of-contents)

## RelabelConfig

RelabelConfig allows dynamic rewriting of the label set, being applied to samples before ingestion. It defines `<metric_relabel_configs>`-section of Prometheus configuration. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sourceLabels | The source labels select values from existing labels. Their content is concatenated using the configured separator and matched against the configured regular expression for the replace, keep, and drop actions. | []string | false |
| separator | Separator placed between concatenated source label values. default is ';'. | string | false |
| targetLabel | Label to which the resulting value is written in a replace action. It is mandatory for replace actions. Regex capture groups are available. | string | false |
| regex | Regular expression against which the extracted value is matched. Default is '(.*)' | string | false |
| modulus | Modulus to take of the hash of the source label values. | uint64 | false |
| replacement | Replacement value against which a regex replace is performed if the regular expression matches. Regex capture groups are available. Default is '$1' | string | false |
| action | Action to perform based on regex matching. Default is 'replace' | string | false |

[Back to TOC](#table-of-contents)

## RemoteReadSpec

RemoteReadSpec defines the remote_read configuration for prometheus.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | The URL of the endpoint to send samples to. | string | true |
| name | The name of the remote read queue, must be unique if specified. The name is used in metrics and logging in order to differentiate read configurations.  Only valid in Prometheus versions 2.15.0 and newer. | string | false |
| requiredMatchers | An optional list of equality matchers which have to be present in a selector to query the remote read endpoint. | map[string]string | false |
| remoteTimeout | Timeout for requests to the remote read endpoint. | string | false |
| readRecent | Whether reads should be made for queries for time ranges that the local storage should have complete data for. | bool | false |
| basicAuth | BasicAuth for the URL. | *[BasicAuth](#basicauth) | false |
| bearerToken | bearer token for remote read. | string | false |
| bearerTokenFile | File to read bearer token for remote read. | string | false |
| tlsConfig | TLS Config to use for remote read. | *[TLSConfig](#tlsconfig) | false |
| proxyUrl | Optional ProxyURL | string | false |

[Back to TOC](#table-of-contents)

## RemoteWriteSpec

RemoteWriteSpec defines the remote_write configuration for prometheus.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | The URL of the endpoint to send samples to. | string | true |
| name | The name of the remote write queue, must be unique if specified. The name is used in metrics and logging in order to differentiate queues. Only valid in Prometheus versions 2.15.0 and newer. | string | false |
| remoteTimeout | Timeout for requests to the remote write endpoint. | string | false |
| writeRelabelConfigs | The list of remote write relabel configurations. | [][RelabelConfig](#relabelconfig) | false |
| basicAuth | BasicAuth for the URL. | *[BasicAuth](#basicauth) | false |
| bearerToken | File to read bearer token for remote write. | string | false |
| bearerTokenFile | File to read bearer token for remote write. | string | false |
| tlsConfig | TLS Config to use for remote write. | *[TLSConfig](#tlsconfig) | false |
| proxyUrl | Optional ProxyURL | string | false |
| queueConfig | QueueConfig allows tuning of the remote write queue parameters. | *[QueueConfig](#queueconfig) | false |

[Back to TOC](#table-of-contents)

## Rule

Rule describes an alerting or recording rule.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| record |  | string | false |
| alert |  | string | false |
| expr |  | intstr.IntOrString | true |
| for |  | string | false |
| labels |  | map[string]string | false |
| annotations |  | map[string]string | false |

[Back to TOC](#table-of-contents)

## RuleGroup

RuleGroup is a list of sequentially evaluated recording and alerting rules. Note: PartialResponseStrategy is only used by ThanosRuler and will be ignored by Prometheus instances.  Valid values for this field are 'warn' or 'abort'.  More info: https://github.com/thanos-io/thanos/blob/master/docs/components/rule.md#partial-response

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| interval |  | string | false |
| rules |  | [][Rule](#rule) | true |
| partial_response_strategy |  | string | false |

[Back to TOC](#table-of-contents)

## Rules

/--rules.*/ command-line arguments

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| alert |  | [RulesAlert](#rulesalert) | false |

[Back to TOC](#table-of-contents)

## RulesAlert

/--rules.alert.*/ command-line arguments

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| forOutageTolerance | Max time to tolerate prometheus outage for restoring 'for' state of alert. | string | false |
| forGracePeriod | Minimum duration between alert and restored 'for' state. This is maintained only for alerts with configured 'for' time greater than grace period. | string | false |
| resendDelay | Minimum amount of time to wait before resending an alert to Alertmanager. | string | false |

[Back to TOC](#table-of-contents)

## SecretOrConfigMap

SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| secret | Secret containing data to use for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| configMap | ConfigMap containing data to use for the targets. | *v1.ConfigMapKeySelector | false |

[Back to TOC](#table-of-contents)

## ServiceMonitor

ServiceMonitor defines monitoring for a set of services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of desired Service selection for target discovery by Prometheus. | [ServiceMonitorSpec](#servicemonitorspec) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorList

ServiceMonitorList is a list of ServiceMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of ServiceMonitors | []*[ServiceMonitor](#servicemonitor) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorSpec

ServiceMonitorSpec contains specification parameters for a ServiceMonitor.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | The label to use to retrieve the job name from. | string | false |
| targetLabels | TargetLabels transfers labels on the Kubernetes Service onto the target. | []string | false |
| podTargetLabels | PodTargetLabels transfers labels on the Kubernetes Pod onto the target. | []string | false |
| endpoints | A list of endpoints allowed as part of this ServiceMonitor. | [][Endpoint](#endpoint) | true |
| selector | Selector to select Endpoints objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |

[Back to TOC](#table-of-contents)

## StorageSpec

StorageSpec defines the configured storage for a group Prometheus servers. If neither `emptyDir` nor `volumeClaimTemplate` is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disableMountSubPath | Deprecated: subPath usage will be disabled by default in a future release, this option will become unnecessary. DisableMountSubPath allows to remove any subPath usage in volume mounts. | bool | false |
| emptyDir | EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir | *[v1.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#emptydirvolumesource-v1-core) | false |
| volumeClaimTemplate | A PVC spec to be used by the Prometheus StatefulSets. | [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim) | false |

[Back to TOC](#table-of-contents)

## TLSConfig

TLSConfig specifies TLS configuration parameters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| caFile | Path to the CA cert in the Prometheus container to use for the targets. | string | false |
| ca | Stuct containing the CA cert to use for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| certFile | Path to the client cert file in the Prometheus container for the targets. | string | false |
| cert | Struct containing the client cert file for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| keyFile | Path to the client key file in the Prometheus container for the targets. | string | false |
| keySecret | Secret containing the client key file for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |

[Back to TOC](#table-of-contents)

## ThanosSpec

ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Thanos is being configured. | *string | false |
| version | Version describes the version of Thanos to use. | *string | false |
| tag | Tag of Thanos sidecar container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | *string | false |
| sha | SHA of Thanos container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | *string | false |
| baseImage | Thanos base image if other than default. Deprecated: use 'image' instead | *string | false |
| resources | Resources defines the resource requirements for the Thanos sidecar. If not provided, no requests/limits will be set | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | false |
| objectStorageConfig | ObjectStorageConfig configures object storage in Thanos. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| listenLocal | ListenLocal makes the Thanos sidecar listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| tracingConfig | TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| grpcServerTlsConfig | GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads recorded rule data. Note: Currently only the CAFile, CertFile, and KeyFile fields are supported. Maps to the '--grpc-server-tls-*' CLI args. | *[TLSConfig](#tlsconfig) | false |
| logLevel | LogLevel for Thanos sidecar to be configured with. | string | false |
| logFormat | LogFormat for Thanos sidecar to be configured with. | string | false |
| minTime | MinTime for Thanos sidecar to be configured with. Option can be a constant time in RFC3339 format or time duration relative to current time, such as -1d or 2h45m. Valid duration units are ms, s, m, h, d, w, y. | string | false |

[Back to TOC](#table-of-contents)

## ThanosRuler

ThanosRuler defines a ThanosRuler deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the ThanosRuler cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [ThanosRulerSpec](#thanosrulerspec) | true |
| status | Most recent observed status of the ThanosRuler cluster. Read-only. Not included when requesting from the apiserver, only from the ThanosRuler Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[ThanosRulerStatus](#thanosrulerstatus) | false |

[Back to TOC](#table-of-contents)

## ThanosRulerList

ThanosRulerList is a list of ThanosRulers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[ThanosRuler](#thanosruler) | true |

[Back to TOC](#table-of-contents)

## ThanosRulerSpec

ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata contains Labels and Annotations gets propagated to the thanos ruler pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| image | Thanos container image URL. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling thanos images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) | false |
| paused | When a ThanosRuler deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| replicas | Number of thanos ruler instances to deploy. | *int32 | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Resources defines the resource requirements for single Pods. If not provided, no requests/limits will be set | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Thanos Ruler Pods. | string | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| objectStorageConfig | ObjectStorageConfig configures object storage in Thanos. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| listenLocal | ListenLocal makes the Thanos ruler listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| queryEndpoints | QueryEndpoints defines Thanos querier endpoints from which to query metrics. Maps to the --query flag of thanos ruler. | []string | false |
| queryConfig | Define configuration for connecting to thanos query instances. If this is defined, the QueryEndpoints field will be ignored. Maps to the `query.config` CLI argument. Only available with thanos v0.11.0 and higher. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| alertmanagersUrl | Define URLs to send alerts to Alertmanager.  For Thanos v0.10.0 and higher, AlertManagersConfig should be used instead.  Note: this field will be ignored if AlertManagersConfig is specified. Maps to the `alertmanagers.url` arg. | []string | false |
| alertmanagersConfig | Define configuration for connecting to alertmanager.  Only available with thanos v0.10.0 and higher.  Maps to the `alertmanagers.config` arg. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| ruleSelector | A label selector to select which PrometheusRules to mount for alerting and recording. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for Rules discovery. If unspecified, only the same namespace as the ThanosRuler object is in is used. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#labelselector-v1-meta) | false |
| enforcedNamespaceLabel | EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert and metric that is user created. The label value will always be the namespace of the object that is being created. | string | false |
| prometheusRulesExcludedFromEnforce | PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing of adding namespace labels. Works only if enforcedNamespaceLabel set to true. Make sure both ruleNamespace and ruleName are set for each pair | [][PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) | false |
| logLevel | Log level for ThanosRuler to be configured with. | string | false |
| logFormat | Log format for ThanosRuler to be configured with. | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| evaluationInterval | Interval between consecutive evaluations. | string | false |
| retention | Time duration ThanosRuler shall retain data for. Default is '24h', and must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | string | false |
| containers | Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or to change the behavior of an operator generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `thanos-ruler` and `rules-configmap-reloader`. Overriding containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the ThanosRuler configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other then secret fetching is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| tracingConfig | TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secretkeyselector-v1-core) | false |
| labels | Labels configure the external label pairs to ThanosRuler. If not provided, default replica label `thanos_ruler_replica` will be added as a label and be dropped in alerts. | map[string]string | false |
| alertDropLabels | AlertDropLabels configure the label names which should be dropped in ThanosRuler alerts. If `labels` field is not provided, `thanos_ruler_replica` will be dropped in alerts by default. | []string | false |
| externalPrefix | The external URL the Thanos Ruler instances will be available under. This is necessary to generate correct URLs. This is necessary if Thanos Ruler is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path. | string | false |
| grpcServerTlsConfig | GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads recorded rule data. Note: Currently only the CAFile, CertFile, and KeyFile fields are supported. Maps to the '--grpc-server-tls-*' CLI args. | *[TLSConfig](#tlsconfig) | false |
| alertQueryUrl | The external Query URL the Thanos Ruler will set in the 'Source' field of all alerts. Maps to the '--alert.query-url' CLI arg. | string | false |

[Back to TOC](#table-of-contents)

## ThanosRulerStatus

ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this ThanosRuler deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this ThanosRuler deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this ThanosRuler deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this ThanosRuler deployment. | int32 | true |

[Back to TOC](#table-of-contents)
