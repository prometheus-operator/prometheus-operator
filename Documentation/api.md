---
title: "API"
description: "Generated API docs for the Prometheus Operator"
lead: ""
date: 2021-03-08T08:49:31+00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 1000
toc: true
---

This Document documents the types introduced by the Prometheus Operator to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## Table of Contents
* [APIServerConfig](#apiserverconfig)
* [AlertingSpec](#alertingspec)
* [Alertmanager](#alertmanager)
* [AlertmanagerConfiguration](#alertmanagerconfiguration)
* [AlertmanagerEndpoints](#alertmanagerendpoints)
* [AlertmanagerList](#alertmanagerlist)
* [AlertmanagerSpec](#alertmanagerspec)
* [AlertmanagerStatus](#alertmanagerstatus)
* [ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig)
* [Authorization](#authorization)
* [BasicAuth](#basicauth)
* [EmbeddedObjectMetadata](#embeddedobjectmetadata)
* [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim)
* [Endpoint](#endpoint)
* [MetadataConfig](#metadataconfig)
* [NamespaceSelector](#namespaceselector)
* [OAuth2](#oauth2)
* [ObjectReference](#objectreference)
* [PodMetricsEndpoint](#podmetricsendpoint)
* [PodMetricsEndpointTLSConfig](#podmetricsendpointtlsconfig)
* [PodMonitor](#podmonitor)
* [PodMonitorList](#podmonitorlist)
* [PodMonitorSpec](#podmonitorspec)
* [Probe](#probe)
* [ProbeList](#probelist)
* [ProbeSpec](#probespec)
* [ProbeTLSConfig](#probetlsconfig)
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
* [SafeAuthorization](#safeauthorization)
* [SecretOrConfigMap](#secretorconfigmap)
* [ServiceMonitor](#servicemonitor)
* [ServiceMonitorList](#servicemonitorlist)
* [ServiceMonitorSpec](#servicemonitorspec)
* [Sigv4](#sigv4)
* [StorageSpec](#storagespec)
* [TLSConfig](#tlsconfig)
* [ThanosSpec](#thanosspec)
* [WebSpec](#webspec)
* [WebTLSConfig](#webtlsconfig)
* [ThanosRuler](#thanosruler)
* [ThanosRulerList](#thanosrulerlist)
* [ThanosRulerSpec](#thanosrulerspec)
* [ThanosRulerStatus](#thanosrulerstatus)
* [AlertmanagerConfig](#alertmanagerconfig)
* [AlertmanagerConfigList](#alertmanagerconfiglist)
* [AlertmanagerConfigSpec](#alertmanagerconfigspec)
* [DayOfMonthRange](#dayofmonthrange)
* [EmailConfig](#emailconfig)
* [HTTPConfig](#httpconfig)
* [InhibitRule](#inhibitrule)
* [KeyValue](#keyvalue)
* [Matcher](#matcher)
* [MuteTimeInterval](#mutetimeinterval)
* [OpsGenieConfig](#opsgenieconfig)
* [OpsGenieConfigResponder](#opsgenieconfigresponder)
* [PagerDutyConfig](#pagerdutyconfig)
* [PagerDutyImageConfig](#pagerdutyimageconfig)
* [PagerDutyLinkConfig](#pagerdutylinkconfig)
* [PushoverConfig](#pushoverconfig)
* [Receiver](#receiver)
* [Route](#route)
* [SNSConfig](#snsconfig)
* [SlackAction](#slackaction)
* [SlackConfig](#slackconfig)
* [SlackConfirmationField](#slackconfirmationfield)
* [SlackField](#slackfield)
* [TimeInterval](#timeinterval)
* [TimeRange](#timerange)
* [VictorOpsConfig](#victoropsconfig)
* [WeChatConfig](#wechatconfig)
* [WebhookConfig](#webhookconfig)

## APIServerConfig

APIServerConfig defines a host and auth methods to access apiserver. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host | Host of apiserver. A valid string consisting of a hostname or IP followed by an optional port number | string | true |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication | *[BasicAuth](#basicauth) | false |
| bearerToken | Bearer token for accessing apiserver. | string | false |
| bearerTokenFile | File to read bearer token for accessing apiserver. | string | false |
| tlsConfig | TLS Config to use for accessing apiserver. | *[TLSConfig](#tlsconfig) | false |
| authorization | Authorization section for accessing apiserver | *[Authorization](#authorization) | false |

[Back to TOC](#table-of-contents)

## AlertingSpec

AlertingSpec defines parameters for alerting configuration of Prometheus servers.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| alertmanagers | AlertmanagerEndpoints Prometheus should fire alerts against. | [][AlertmanagerEndpoints](#alertmanagerendpoints) | true |

[Back to TOC](#table-of-contents)

## Alertmanager

Alertmanager describes an Alertmanager cluster.


<em>appears in: [AlertmanagerList](#alertmanagerlist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [AlertmanagerSpec](#alertmanagerspec) | true |
| status | Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[AlertmanagerStatus](#alertmanagerstatus) | false |

[Back to TOC](#table-of-contents)

## AlertmanagerConfiguration

AlertmanagerConfiguration defines the global Alertmanager configuration.


<em>appears in: [AlertmanagerSpec](#alertmanagerspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | The name of the AlertmanagerConfig resource which is used to generate the global configuration. It must be defined in the same namespace as the Alertmanager object. The operator will not enforce a `namespace` label for routes and inhibition rules. | string | false |

[Back to TOC](#table-of-contents)

## AlertmanagerEndpoints

AlertmanagerEndpoints defines a selection of a single Endpoints object containing alertmanager IPs to fire alerts against.


<em>appears in: [AlertingSpec](#alertingspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | Namespace of Endpoints object. | string | true |
| name | Name of Endpoints object in Namespace. | string | true |
| port | Port the Alertmanager API is exposed on. | intstr.IntOrString | true |
| scheme | Scheme to use when firing alerts. | string | false |
| pathPrefix | Prefix for the HTTP path alerts are pushed to. | string | false |
| tlsConfig | TLS Config to use for alertmanager connection. | *[TLSConfig](#tlsconfig) | false |
| bearerTokenFile | BearerTokenFile to read from filesystem to use when authenticating to Alertmanager. | string | false |
| authorization | Authorization section for this alertmanager endpoint | *[SafeAuthorization](#safeauthorization) | false |
| apiVersion | Version of the Alertmanager API that Prometheus uses to send alerts. It can be \"v1\" or \"v2\". | string | false |
| timeout | Timeout is a per-target Alertmanager timeout when pushing alerts. | *string | false |

[Back to TOC](#table-of-contents)

## AlertmanagerList

AlertmanagerList is a list of Alertmanagers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of Alertmanagers | [][Alertmanager](#alertmanager) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerSpec

AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [Alertmanager](#alertmanager)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Alertmanager is being configured. | *string | false |
| version | Version the cluster should be on. | string | false |
| tag | Tag of Alertmanager container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | string | false |
| sha | SHA of Alertmanager container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | string | false |
| baseImage | Base image that is used to deploy pods, without tag. Deprecated: use 'image' instead | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core) | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>. | []string | false |
| configSecret | ConfigSecret is the name of a Kubernetes Secret in the same namespace as the Alertmanager object, which contains the configuration for this Alertmanager instance. If empty, it defaults to 'alertmanager-<alertmanager-name>'.\n\nThe Alertmanager configuration should be available under the `alertmanager.yaml` key. Additional keys from the original secret are copied to the generated secret.\n\nIf either the secret or the `alertmanager.yaml` key is missing, the operator provisions an Alertmanager configuration with one empty receiver (effectively dropping alert notifications). | string | false |
| logLevel | Log level for Alertmanager to be configured with. | string | false |
| logFormat | Log format for Alertmanager to be configured with. | string | false |
| replicas | Size is the expected size of the alertmanager cluster. The controller will eventually make the size of the running cluster equal to the expected size. | *int32 | false |
| retention | Time duration Alertmanager shall retain data for. Default is '120h', and must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds seconds minutes hours). | string | false |
| storage | Storage is the definition of how storage will be used by the Alertmanager instances. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| volumeMounts | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container, that are generated as a result of StorageSpec objects. | []v1.VolumeMount | false |
| externalUrl | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Alertmanager registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| paused | If set to true all actions on the underlying managed objects are not goint to be performed, except for delete actions. | bool | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| topologySpreadConstraints | If specified, the pod's topology spread constraints. | []v1.TopologySpreadConstraint | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| listenLocal | ListenLocal makes the Alertmanager server listen on loopback, so that it does not bind against the Pod IP. Note this is only for the Alertmanager UI, not the gossip communication. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to an Alertmanager pod. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `alertmanager` and `config-reloader`. Overriding containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the Alertmanager configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other then secret fetching is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| additionalPeers | AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster. | []string | false |
| clusterAdvertiseAddress | ClusterAdvertiseAddress is the explicit address to advertise in cluster. Needs to be provided for non RFC1918 [1] (public) addresses. [1] RFC1918: https://tools.ietf.org/html/rfc1918 | string | false |
| clusterGossipInterval | Interval between gossip attempts. | string | false |
| clusterPushpullInterval | Interval between pushpull attempts. | string | false |
| clusterPeerTimeout | Timeout for cluster peering. | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| forceEnableClusterMode | ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica. Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each. | bool | false |
| alertmanagerConfigSelector | AlertmanagerConfigs to be selected for to merge and configure Alertmanager with. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| alertmanagerConfigNamespaceSelector | Namespaces to be selected for AlertmanagerConfig discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| minReadySeconds | Minimum number of seconds for which a newly created pod should be ready without any of its container crashing for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready) This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate. | *uint32 | false |
| alertmanagerConfiguration | EXPERIMENTAL: alertmanagerConfiguration specifies the global Alertmanager configuration. If defined, it takes precedence over the `configSecret` field. This field may change in future releases. | *[AlertmanagerConfiguration](#alertmanagerconfiguration) | false |

[Back to TOC](#table-of-contents)

## AlertmanagerStatus

AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [Alertmanager](#alertmanager)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Alertmanager cluster (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Alertmanager cluster that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Alertmanager cluster. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Alertmanager cluster. | int32 | true |

[Back to TOC](#table-of-contents)

## ArbitraryFSAccessThroughSMsConfig

ArbitraryFSAccessThroughSMsConfig enables users to configure, whether a service monitor selected by the Prometheus instance is allowed to use arbitrary files on the file system of the Prometheus container. This is the case when e.g. a service monitor specifies a BearerTokenFile in an endpoint. A malicious user could create a service monitor selecting arbitrary secret files in the Prometheus container. Those secrets would then be sent with a scrape request by Prometheus to a malicious target. Denying the above would prevent the attack, users can instead use the BearerTokenSecret field.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deny |  | bool | false |

[Back to TOC](#table-of-contents)

## Authorization

Authorization contains optional `Authorization` header configuration. This section is only understood by versions of Prometheus >= 2.26.0.


<em>appears in: [APIServerConfig](#apiserverconfig), [RemoteReadSpec](#remotereadspec), [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Set the authentication type. Defaults to Bearer, Basic will cause an error | string | false |
| credentials | The secret's key that contains the credentials of the request | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| credentialsFile | File to read a secret from, mutually exclusive with Credentials (from SafeAuthorization) | string | false |

[Back to TOC](#table-of-contents)

## BasicAuth

BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints


<em>appears in: [APIServerConfig](#apiserverconfig), [Endpoint](#endpoint), [PodMetricsEndpoint](#podmetricsendpoint), [ProbeSpec](#probespec), [RemoteReadSpec](#remotereadspec), [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username | The secret in the service monitor namespace that contains the username for authentication. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| password | The secret in the service monitor namespace that contains the password for authentication. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## EmbeddedObjectMetadata

EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta Only fields which are relevant to embedded resources are included.


<em>appears in: [AlertmanagerSpec](#alertmanagerspec), [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim), [PrometheusSpec](#prometheusspec), [ThanosRulerSpec](#thanosrulerspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names | string | false |
| labels | Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels | map[string]string | false |
| annotations | Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations | map[string]string | false |

[Back to TOC](#table-of-contents)

## EmbeddedPersistentVolumeClaim

EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim. It contains TypeMeta and a reduced ObjectMeta.


<em>appears in: [StorageSpec](#storagespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | EmbeddedMetadata contains metadata relevant to an EmbeddedResource. | [EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| spec | Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims | v1.PersistentVolumeClaimSpec | false |
| status | Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims | v1.PersistentVolumeClaimStatus | false |

[Back to TOC](#table-of-contents)

## Endpoint

Endpoint defines a scrapeable endpoint serving Prometheus metrics.


<em>appears in: [ServiceMonitorSpec](#servicemonitorspec)</em>

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
| bearerTokenSecret | Secret to mount to read bearer token for scraping targets. The secret needs to be in the same namespace as the service monitor and accessible by the Prometheus Operator. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| authorization | Authorization section for this endpoint | *[SafeAuthorization](#safeauthorization) | false |
| honorLabels | HonorLabels chooses the metric's labels on collisions with target labels. | bool | false |
| honorTimestamps | HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data. | *bool | false |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints | *[BasicAuth](#basicauth) | false |
| oauth2 | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. | *[OAuth2](#oauth2) | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| relabelings | RelabelConfigs to apply to samples before scraping. Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields. The original scrape job's name is available via the `__tmp_prometheus_job_name` label. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |
| followRedirects | FollowRedirects configures whether scrape requests follow HTTP 3xx redirects. | *bool | false |

[Back to TOC](#table-of-contents)

## MetadataConfig

MetadataConfig configures the sending of series metadata to the remote storage.


<em>appears in: [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| send | Whether metric metadata is sent to the remote storage or not. | bool | false |
| sendInterval | How frequently metric metadata is sent to the remote storage. | string | false |

[Back to TOC](#table-of-contents)

## NamespaceSelector

NamespaceSelector is a selector for selecting either all namespaces or a list of namespaces. If `any` is true, it takes precedence over `matchNames`. If `matchNames` is empty and `any` is false, it means that the objects are selected from the current namespace.


<em>appears in: [PodMonitorSpec](#podmonitorspec), [ProbeTargetIngress](#probetargetingress), [ServiceMonitorSpec](#servicemonitorspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| any | Boolean describing whether all namespaces are selected in contrast to a list restricting them. | bool | false |
| matchNames | List of namespace names to select from. | []string | false |

[Back to TOC](#table-of-contents)

## OAuth2

OAuth2 allows an endpoint to authenticate with OAuth2. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#oauth2


<em>appears in: [Endpoint](#endpoint), [PodMetricsEndpoint](#podmetricsendpoint), [ProbeSpec](#probespec), [RemoteReadSpec](#remotereadspec), [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| clientId | The secret or configmap containing the OAuth2 client id | [SecretOrConfigMap](#secretorconfigmap) | true |
| clientSecret | The secret containing the OAuth2 client secret | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | true |
| tokenUrl | The URL to fetch the token from | string | true |
| scopes | OAuth2 scopes used for the token request | []string | false |
| endpointParams | Parameters to append to the token URL | map[string]string | false |

[Back to TOC](#table-of-contents)

## ObjectReference

ObjectReference references a PodMonitor, ServiceMonitor, Probe or PrometheusRule object.


<em>appears in: [PrometheusSpec](#prometheusspec), [ThanosRulerSpec](#thanosrulerspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| group | Group of the referent. When not specified, it defaults to `monitoring.coreos.com` | string | true |
| resource | Resource of the referent. | string | true |
| namespace | Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/ | string | true |
| name | Name of the referent. When not set, all resources are matched. | string | false |

[Back to TOC](#table-of-contents)

## PodMetricsEndpoint

PodMetricsEndpoint defines a scrapeable endpoint of a Kubernetes Pod serving Prometheus metrics.


<em>appears in: [PodMonitorSpec](#podmonitorspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the pod port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Deprecated: Use 'port' instead. | *intstr.IntOrString | false |
| path | HTTP path to scrape for metrics. | string | false |
| scheme | HTTP scheme to use for scraping. | string | false |
| params | Optional HTTP URL parameters | map[string][]string | false |
| interval | Interval at which metrics should be scraped | string | false |
| scrapeTimeout | Timeout after which the scrape is ended | string | false |
| tlsConfig | TLS configuration to use when scraping the endpoint. | *[PodMetricsEndpointTLSConfig](#podmetricsendpointtlsconfig) | false |
| bearerTokenSecret | Secret to mount to read bearer token for scraping targets. The secret needs to be in the same namespace as the pod monitor and accessible by the Prometheus Operator. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| honorLabels | HonorLabels chooses the metric's labels on collisions with target labels. | bool | false |
| honorTimestamps | HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data. | *bool | false |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication. More info: https://prometheus.io/docs/operating/configuration/#endpoint | *[BasicAuth](#basicauth) | false |
| oauth2 | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. | *[OAuth2](#oauth2) | false |
| authorization | Authorization section for this endpoint | *[SafeAuthorization](#safeauthorization) | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| relabelings | RelabelConfigs to apply to samples before scraping. Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields. The original scrape job's name is available via the `__tmp_prometheus_job_name` label. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |
| followRedirects | FollowRedirects configures whether scrape requests follow HTTP 3xx redirects. | *bool | false |

[Back to TOC](#table-of-contents)

## PodMetricsEndpointTLSConfig

PodMetricsEndpointTLSConfig specifies TLS configuration parameters.


<em>appears in: [PodMetricsEndpoint](#podmetricsendpoint)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| ca | Struct containing the CA cert to use for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| cert | Struct containing the client cert file for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| keySecret | Secret containing the client key file for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |

[Back to TOC](#table-of-contents)

## PodMonitor

PodMonitor defines monitoring for a set of pods.


<em>appears in: [PodMonitorList](#podmonitorlist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of desired Pod selection for target discovery by Prometheus. | [PodMonitorSpec](#podmonitorspec) | true |

[Back to TOC](#table-of-contents)

## PodMonitorList

PodMonitorList is a list of PodMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of PodMonitors | []*[PodMonitor](#podmonitor) | true |

[Back to TOC](#table-of-contents)

## PodMonitorSpec

PodMonitorSpec contains specification parameters for a PodMonitor.


<em>appears in: [PodMonitor](#podmonitor)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | The label to use to retrieve the job name from. | string | false |
| podTargetLabels | PodTargetLabels transfers labels on the Kubernetes Pod onto the target. | []string | false |
| podMetricsEndpoints | A list of endpoints allowed as part of this PodMonitor. | [][PodMetricsEndpoint](#podmetricsendpoint) | true |
| selector | Selector to select Pod objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |
| targetLimit | TargetLimit defines a limit on the number of scraped targets that will be accepted. | uint64 | false |
| labelLimit | Per-scrape limit on number of labels that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelNameLengthLimit | Per-scrape limit on length of labels name that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelValueLengthLimit | Per-scrape limit on length of labels value that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |

[Back to TOC](#table-of-contents)

## Probe

Probe defines monitoring for a set of static targets or ingresses.


<em>appears in: [ProbeList](#probelist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of desired Ingress selection for target discovery by Prometheus. | [ProbeSpec](#probespec) | true |

[Back to TOC](#table-of-contents)

## ProbeList

ProbeList is a list of Probes.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of Probes | []*[Probe](#probe) | true |

[Back to TOC](#table-of-contents)

## ProbeSpec

ProbeSpec contains specification parameters for a Probe.


<em>appears in: [Probe](#probe)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobName | The job name assigned to scraped metrics by default. | string | false |
| prober | Specification for the prober to use for probing targets. The prober.URL parameter is required. Targets cannot be probed if left empty. | [ProberSpec](#proberspec) | false |
| module | The module to use for probing specifying how to probe the target. Example module configuring in the blackbox exporter: https://github.com/prometheus/blackbox_exporter/blob/master/example.yml | string | false |
| targets | Targets defines a set of static or dynamically discovered targets to probe. | [ProbeTargets](#probetargets) | false |
| interval | Interval at which targets are probed using the configured prober. If not specified Prometheus' global scrape interval is used. | string | false |
| scrapeTimeout | Timeout for scraping metrics from the Prometheus exporter. | string | false |
| tlsConfig | TLS configuration to use when scraping the endpoint. | *[ProbeTLSConfig](#probetlsconfig) | false |
| bearerTokenSecret | Secret to mount to read bearer token for scraping targets. The secret needs to be in the same namespace as the probe and accessible by the Prometheus Operator. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication. More info: https://prometheus.io/docs/operating/configuration/#endpoint | *[BasicAuth](#basicauth) | false |
| oauth2 | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. | *[OAuth2](#oauth2) | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| authorization | Authorization section for this endpoint | *[SafeAuthorization](#safeauthorization) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |
| targetLimit | TargetLimit defines a limit on the number of scraped targets that will be accepted. | uint64 | false |
| labelLimit | Per-scrape limit on number of labels that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelNameLengthLimit | Per-scrape limit on length of labels name that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelValueLengthLimit | Per-scrape limit on length of labels value that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |

[Back to TOC](#table-of-contents)

## ProbeTLSConfig

ProbeTLSConfig specifies TLS configuration parameters.


<em>appears in: [ProbeSpec](#probespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| ca | Struct containing the CA cert to use for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| cert | Struct containing the client cert file for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| keySecret | Secret containing the client key file for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |

[Back to TOC](#table-of-contents)

## ProbeTargetIngress

ProbeTargetIngress defines the set of Ingress objects considered for probing. The operator configures a target for each host/path combination of each ingress object.


<em>appears in: [ProbeTargets](#probetargets)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| selector | Selector to select the Ingress objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| namespaceSelector | From which namespaces to select Ingress objects. | [NamespaceSelector](#namespaceselector) | false |
| relabelingConfigs | RelabelConfigs to apply to the label set of the target before it gets scraped. The original ingress address is available via the `__tmp_prometheus_ingress_address` label. It can be used to customize the probed URL. The original scrape job's name is available via the `__tmp_prometheus_job_name` label. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |

[Back to TOC](#table-of-contents)

## ProbeTargetStaticConfig

ProbeTargetStaticConfig defines the set of static targets considered for probing.


<em>appears in: [ProbeTargets](#probetargets)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| static | The list of hosts to probe. | []string | false |
| labels | Labels assigned to all metrics scraped from the targets. | map[string]string | false |
| relabelingConfigs | RelabelConfigs to apply to the label set of the targets before it gets scraped. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |

[Back to TOC](#table-of-contents)

## ProbeTargets

ProbeTargets defines how to discover the probed targets. One of the `staticConfig` or `ingress` must be defined. If both are defined, `staticConfig` takes precedence.


<em>appears in: [ProbeSpec](#probespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| staticConfig | staticConfig defines the static list of targets to probe and the relabeling configuration. If `ingress` is also defined, `staticConfig` takes precedence. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config. | *[ProbeTargetStaticConfig](#probetargetstaticconfig) | false |
| ingress | ingress defines the Ingress objects to probe and the relabeling configuration. If `staticConfig` is also defined, `staticConfig` takes precedence. | *[ProbeTargetIngress](#probetargetingress) | false |

[Back to TOC](#table-of-contents)

## ProberSpec

ProberSpec contains specification parameters for the Prober used for probing.


<em>appears in: [ProbeSpec](#probespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | Mandatory URL of the prober. | string | true |
| scheme | HTTP scheme to use for scraping. Defaults to `http`. | string | false |
| path | Path to collect metrics from. Defaults to `/probe`. | string | false |
| proxyUrl | Optional ProxyURL. | string | false |

[Back to TOC](#table-of-contents)

## Prometheus

Prometheus defines a Prometheus deployment.


<em>appears in: [PrometheusList](#prometheuslist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [PrometheusSpec](#prometheusspec) | true |
| status | Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[PrometheusStatus](#prometheusstatus) | false |

[Back to TOC](#table-of-contents)

## PrometheusList

PrometheusList is a list of Prometheuses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[Prometheus](#prometheus) | true |

[Back to TOC](#table-of-contents)

## PrometheusRule

PrometheusRule defines recording and alerting rules for a Prometheus instance


<em>appears in: [PrometheusRuleList](#prometheusrulelist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of desired alerting rule definitions for Prometheus. | [PrometheusRuleSpec](#prometheusrulespec) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleExcludeConfig

PrometheusRuleExcludeConfig enables users to configure excluded PrometheusRule names and their namespaces to be ignored while enforcing namespace label for alerts and metrics.


<em>appears in: [PrometheusSpec](#prometheusspec), [ThanosRulerSpec](#thanosrulerspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| ruleNamespace | RuleNamespace - namespace of excluded rule | string | true |
| ruleName | RuleNamespace - name of excluded rule | string | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleList

PrometheusRuleList is a list of PrometheusRules.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of Rules | []*[PrometheusRule](#prometheusrule) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleSpec

PrometheusRuleSpec contains specification parameters for a Rule.


<em>appears in: [PrometheusRule](#prometheusrule)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| groups | Content of Prometheus rule file | [][RuleGroup](#rulegroup) | false |

[Back to TOC](#table-of-contents)

## PrometheusSpec

PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [Prometheus](#prometheus)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata configures Labels and Annotations which are propagated to the prometheus pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| serviceMonitorSelector | ServiceMonitors to be selected for target discovery. *Deprecated:* if neither this nor podMonitorSelector are specified, configuration is unmanaged. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| serviceMonitorNamespaceSelector | Namespace's labels to match for ServiceMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| podMonitorSelector | *Experimental* PodMonitors to be selected for target discovery. *Deprecated:* if neither this nor serviceMonitorSelector are specified, configuration is unmanaged. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| podMonitorNamespaceSelector | Namespace's labels to match for PodMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| probeSelector | *Experimental* Probes to be selected for target discovery. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| probeNamespaceSelector | *Experimental* Namespaces to be selected for Probe discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| version | Version of Prometheus to be deployed. | string | false |
| tag | Tag of Prometheus container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | string | false |
| sha | SHA of Prometheus container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | string | false |
| paused | When a Prometheus deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Prometheus is being configured. | *string | false |
| baseImage | Base image to use for a Prometheus deployment. Deprecated: use 'image' instead | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core) | false |
| replicas | Number of replicas of each shard to deploy for a Prometheus deployment. Number of replicas multiplied by shards is the total number of Pods created. | *int32 | false |
| shards | EXPERIMENTAL: Number of shards to distribute targets onto. Number of replicas multiplied by shards is the total number of Pods created. Note that scaling down shards will not reshard data onto remaining instances, it must be manually moved. Increasing shards will not reshard data either but it will continue to be available from the same instances. To query globally use Thanos sidecar and Thanos querier or remote write data to a central location. Sharding is done on the content of the `__address__` target meta-label. | *int32 | false |
| replicaExternalLabelName | Name of Prometheus external label used to denote replica name. Defaults to the value of `prometheus_replica`. External label will _not_ be added when value is set to empty string (`\"\"`). | *string | false |
| prometheusExternalLabelName | Name of Prometheus external label used to denote Prometheus instance name. Defaults to the value of `prometheus`. External label will _not_ be added when value is set to empty string (`\"\"`). | *string | false |
| logLevel | Log level for Prometheus to be configured with. | string | false |
| logFormat | Log format for Prometheus to be configured with. | string | false |
| scrapeInterval | Interval between consecutive scrapes. Default: `1m` | string | false |
| scrapeTimeout | Number of seconds to wait for target to respond before erroring. | string | false |
| evaluationInterval | Interval between consecutive evaluations. Default: `1m` | string | false |
| externalLabels | The labels to add to any time series or alerts when communicating with external systems (federation, remote storage, Alertmanager). | map[string]string | false |
| enableAdminAPI | Enable access to prometheus web admin API. Defaults to the value of `false`. WARNING: Enabling the admin APIs enables mutating endpoints, to delete data, shutdown Prometheus, and more. Enabling this should be done with care and the user is advised to add additional authentication authorization via a proxy to ensure only clients authorized to perform these actions can do so. For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis | bool | false |
| enableRemoteWriteReceiver | Enable Prometheus to be used as a receiver for the Prometheus remote write protocol. Defaults to the value of `false`. WARNING: This is not considered an efficient way of ingesting samples. Use it with caution for specific low-volume use cases. It is not suitable for replacing the ingestion via scraping and turning Prometheus into a push-based metrics collection system. For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver Only valid in Prometheus versions 2.33.0 and newer. | bool | false |
| enableFeatures | Enable access to Prometheus disabled features. By default, no features are enabled. Enabling disabled features is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. For more information see https://prometheus.io/docs/prometheus/latest/disabled_features/ | []string | false |
| externalUrl | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Prometheus registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| volumeMounts | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumeMounts specified will be appended to other VolumeMounts in the prometheus container, that are generated as a result of StorageSpec objects. | []v1.VolumeMount | false |
| web | WebSpec defines the web command line flags when starting Prometheus. | *[WebSpec](#webspec) | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core) | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The Secrets are mounted into /etc/prometheus/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>. | []string | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| topologySpreadConstraints | If specified, the pod's topology spread constraints. | []v1.TopologySpreadConstraint | false |
| remoteWrite | remoteWrite is the list of remote write configurations. | [][RemoteWriteSpec](#remotewritespec) | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| listenLocal | ListenLocal makes the Prometheus server listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| containers | Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a Prometheus pod or to change the behavior of an operator generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `prometheus`, `config-reloader`, and `thanos-sidecar`. Overriding containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the Prometheus configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ InitContainers described here modify an operator generated init containers if they share the same name and modifications are done via a strategic merge patch. The current init container name is: `init-config-reloader`. Overriding init containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| additionalScrapeConfigs | AdditionalScrapeConfigs allows specifying a key of a Secret containing additional Prometheus scrape configurations. Scrape configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config. As scrape configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible scrape configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| apiserverConfig | APIServerConfig allows specifying a host and auth methods to access apiserver. If left empty, Prometheus is assumed to run inside of the cluster and will discover API servers automatically and use the pod's CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. | *[APIServerConfig](#apiserverconfig) | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| arbitraryFSAccessThroughSMs | ArbitraryFSAccessThroughSMs configures whether configuration based on a service monitor can access arbitrary files on the file system of the Prometheus container e.g. bearer token files. | [ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig) | false |
| overrideHonorLabels | When true, Prometheus resolves label conflicts by renaming the labels in the scraped data to \"exported_<label value>\" for all targets created from service and pod monitors. Otherwise the HonorLabels field of the service or pod monitor applies. | bool | false |
| overrideHonorTimestamps | When true, Prometheus ignores the timestamps for all the targets created from service and pod monitors. Otherwise the HonorTimestamps field of the service or pod monitor applies. | bool | false |
| ignoreNamespaceSelectors | IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector settings from all PodMonitor, ServiceMonitor and Probe objects. They will only discover endpoints within their current namespace. Defaults to false. | bool | false |
| enforcedNamespaceLabel | EnforcedNamespaceLabel If set, a label will be added to\n\n1. all user-metrics (created by `ServiceMonitor`, `PodMonitor` and `Probe` objects) and 2. in all `PrometheusRule` objects (except the ones excluded in `prometheusRulesExcludedFromEnforce`) to\n   * alerting & recording rules and\n   * the metrics used in their expressions (`expr`).\n\nLabel name is this field's value. Label value is the namespace of the created object (mentioned above). | string | false |
| enforcedSampleLimit | EnforcedSampleLimit defines global limit on number of scraped samples that will be accepted. This overrides any SampleLimit set per ServiceMonitor or/and PodMonitor. It is meant to be used by admins to enforce the SampleLimit to keep overall number of samples/series under the desired limit. Note that if SampleLimit is lower that value will be taken instead. | *uint64 | false |
| enforcedTargetLimit | EnforcedTargetLimit defines a global limit on the number of scraped targets.  This overrides any TargetLimit set per ServiceMonitor or/and PodMonitor.  It is meant to be used by admins to enforce the TargetLimit to keep the overall number of targets under the desired limit. Note that if TargetLimit is lower, that value will be taken instead, except if either value is zero, in which case the non-zero value will be used.  If both values are zero, no limit is enforced. | *uint64 | false |
| enforcedLabelLimit | Per-scrape limit on number of labels that will be accepted for a sample. If more than this number of labels are present post metric-relabeling, the entire scrape will be treated as failed. 0 means no limit. Only valid in Prometheus versions 2.27.0 and newer. | *uint64 | false |
| enforcedLabelNameLengthLimit | Per-scrape limit on length of labels name that will be accepted for a sample. If a label name is longer than this number post metric-relabeling, the entire scrape will be treated as failed. 0 means no limit. Only valid in Prometheus versions 2.27.0 and newer. | *uint64 | false |
| enforcedLabelValueLengthLimit | Per-scrape limit on length of labels value that will be accepted for a sample. If a label value is longer than this number post metric-relabeling, the entire scrape will be treated as failed. 0 means no limit. Only valid in Prometheus versions 2.27.0 and newer. | *uint64 | false |
| enforcedBodySizeLimit | EnforcedBodySizeLimit defines the maximum size of uncompressed response body that will be accepted by Prometheus. Targets responding with a body larger than this many bytes will cause the scrape to fail. Example: 100MB. If defined, the limit will apply to all service/pod monitors and probes. This is an experimental feature, this behaviour could change or be removed in the future. Only valid in Prometheus versions 2.28.0 and newer. | ByteSize | false |
| minReadySeconds | Minimum number of seconds for which a newly created pod should be ready without any of its container crashing for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready) This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate. | *uint32 | false |
| retention | Time duration Prometheus shall retain data for. Default is '24h' if retentionSize is not set, and must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | string | false |
| retentionSize | Maximum amount of disk space used by blocks. | ByteSize | false |
| disableCompaction | Disable prometheus compaction. | bool | false |
| walCompression | Enable compression of the write-ahead log using Snappy. This flag is only available in versions of Prometheus >= 2.11.0. | *bool | false |
| rules | /--rules.*/ command-line arguments. | [Rules](#rules) | false |
| excludedFromEnforcement | List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects to be excluded from enforcing a namespace label of origin. Applies only if enforcedNamespaceLabel set to true. | [][ObjectReference](#objectreference) | false |
| prometheusRulesExcludedFromEnforce | PrometheusRulesExcludedFromEnforce - list of prometheus rules to be excluded from enforcing of adding namespace labels. Works only if enforcedNamespaceLabel set to true. Make sure both ruleNamespace and ruleName are set for each pair. Deprecated: use excludedFromEnforcement instead. | [][PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) | false |
| query | QuerySpec defines the query command line flags when starting Prometheus. | *[QuerySpec](#queryspec) | false |
| ruleSelector | A selector to select which PrometheusRules to mount for loading alerting/recording rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom resources selected by RuleSelector. Make sure it does not match any config maps that you do not want to be migrated. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for PrometheusRules discovery. If unspecified, only the same namespace as the Prometheus object is in is used. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| alerting | Define details regarding alerting. | *[AlertingSpec](#alertingspec) | false |
| remoteRead | remoteRead is the list of remote read configurations. | [][RemoteReadSpec](#remotereadspec) | false |
| additionalAlertRelabelConfigs | AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing additional Prometheus alert relabel configurations. Alert relabel configurations specified are appended to the configurations generated by the Prometheus Operator. Alert relabel configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs. As alert relabel configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible alert relabel configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| additionalAlertManagerConfigs | AdditionalAlertManagerConfigs allows specifying a key of a Secret containing additional Prometheus AlertManager configurations. AlertManager configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config. As AlertManager configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible AlertManager configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| thanos | Thanos configuration allows configuring various aspects of a Prometheus server in a Thanos environment.\n\nThis section is experimental, it may change significantly without deprecation notice in any release.\n\nThis is experimental and may change significantly without backward compatibility in any release. | *[ThanosSpec](#thanosspec) | false |
| queryLogFile | QueryLogFile specifies the file to which PromQL queries are logged. If the filename has an empty path, e.g. 'query.log', prometheus-operator will mount the file into an emptyDir volume at `/var/log/prometheus`. If a full path is provided, e.g. /var/log/prometheus/query.log, you must mount a volume in the specified directory and it must be writable. This is because the prometheus container runs with a read-only root filesystem for security reasons. Alternatively, the location can be set to a stdout location such as `/dev/stdout` to log query information to the default Prometheus log stream. This is only available in versions of Prometheus >= 2.16.0. For more details, see the Prometheus docs (https://prometheus.io/docs/guides/query-log/) | string | false |
| allowOverlappingBlocks | AllowOverlappingBlocks enables vertical compaction and vertical query merge in Prometheus. This is still experimental in Prometheus so it may change in any upcoming release. | bool | false |

[Back to TOC](#table-of-contents)

## PrometheusStatus

PrometheusStatus is the most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [Prometheus](#prometheus)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Prometheus deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Prometheus deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Prometheus deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Prometheus deployment. | int32 | true |

[Back to TOC](#table-of-contents)

## QuerySpec

QuerySpec defines the query command line flags when starting Prometheus.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lookbackDelta | The delta difference allowed for retrieving metrics during expression evaluations. | *string | false |
| maxConcurrency | Number of concurrent queries that can be run at once. | *int32 | false |
| maxSamples | Maximum number of samples a single query can load into memory. Note that queries will fail if they would load more samples than this into memory, so this also limits the number of samples a query can return. | *int32 | false |
| timeout | Maximum time a query may take before being aborted. | *string | false |

[Back to TOC](#table-of-contents)

## QueueConfig

QueueConfig allows the tuning of remote write's queue_config parameters. This object is referenced in the RemoteWriteSpec object.


<em>appears in: [RemoteWriteSpec](#remotewritespec)</em>

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
| retryOnRateLimit | Retry upon receiving a 429 status code from the remote-write storage. This is experimental feature and might change in the future. | bool | false |

[Back to TOC](#table-of-contents)

## RelabelConfig

RelabelConfig allows dynamic rewriting of the label set, being applied to samples before ingestion. It defines `<metric_relabel_configs>`-section of Prometheus configuration. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs


<em>appears in: [Endpoint](#endpoint), [PodMetricsEndpoint](#podmetricsendpoint), [ProbeSpec](#probespec), [ProbeTargetIngress](#probetargetingress), [ProbeTargetStaticConfig](#probetargetstaticconfig), [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sourceLabels | The source labels select values from existing labels. Their content is concatenated using the configured separator and matched against the configured regular expression for the replace, keep, and drop actions. | []LabelName | false |
| separator | Separator placed between concatenated source label values. default is ';'. | string | false |
| targetLabel | Label to which the resulting value is written in a replace action. It is mandatory for replace actions. Regex capture groups are available. | string | false |
| regex | Regular expression against which the extracted value is matched. Default is '(.*)' | string | false |
| modulus | Modulus to take of the hash of the source label values. | uint64 | false |
| replacement | Replacement value against which a regex replace is performed if the regular expression matches. Regex capture groups are available. Default is '$1' | string | false |
| action | Action to perform based on regex matching. Default is 'replace' | string | false |

[Back to TOC](#table-of-contents)

## RemoteReadSpec

RemoteReadSpec defines the configuration for Prometheus to read back samples from a remote endpoint.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | The URL of the endpoint to query from. | string | true |
| name | The name of the remote read queue, it must be unique if specified. The name is used in metrics and logging in order to differentiate read configurations.  Only valid in Prometheus versions 2.15.0 and newer. | string | false |
| requiredMatchers | An optional list of equality matchers which have to be present in a selector to query the remote read endpoint. | map[string]string | false |
| remoteTimeout | Timeout for requests to the remote read endpoint. | string | false |
| headers | Custom HTTP headers to be sent along with each remote read request. Be aware that headers that are set by Prometheus itself can't be overwritten. Only valid in Prometheus versions 2.26.0 and newer. | map[string]string | false |
| readRecent | Whether reads should be made for queries for time ranges that the local storage should have complete data for. | bool | false |
| basicAuth | BasicAuth for the URL. | *[BasicAuth](#basicauth) | false |
| oauth2 | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. | *[OAuth2](#oauth2) | false |
| bearerToken | Bearer token for remote read. | string | false |
| bearerTokenFile | File to read bearer token for remote read. | string | false |
| authorization | Authorization section for remote read | *[Authorization](#authorization) | false |
| tlsConfig | TLS Config to use for remote read. | *[TLSConfig](#tlsconfig) | false |
| proxyUrl | Optional ProxyURL. | string | false |

[Back to TOC](#table-of-contents)

## RemoteWriteSpec

RemoteWriteSpec defines the configuration to write samples from Prometheus to a remote endpoint.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | The URL of the endpoint to send samples to. | string | true |
| name | The name of the remote write queue, it must be unique if specified. The name is used in metrics and logging in order to differentiate queues. Only valid in Prometheus versions 2.15.0 and newer. | string | false |
| sendExemplars | Enables sending of exemplars over remote write. Note that exemplar-storage itself must be enabled using the enableFeature option for exemplars to be scraped in the first place.  Only valid in Prometheus versions 2.27.0 and newer. | *bool | false |
| remoteTimeout | Timeout for requests to the remote write endpoint. | string | false |
| headers | Custom HTTP headers to be sent along with each remote write request. Be aware that headers that are set by Prometheus itself can't be overwritten. Only valid in Prometheus versions 2.25.0 and newer. | map[string]string | false |
| writeRelabelConfigs | The list of remote write relabel configurations. | [][RelabelConfig](#relabelconfig) | false |
| oauth2 | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. | *[OAuth2](#oauth2) | false |
| basicAuth | BasicAuth for the URL. | *[BasicAuth](#basicauth) | false |
| bearerToken | Bearer token for remote write. | string | false |
| bearerTokenFile | File to read bearer token for remote write. | string | false |
| authorization | Authorization section for remote write | *[Authorization](#authorization) | false |
| sigv4 | Sigv4 allows to configures AWS's Signature Verification 4 | *[Sigv4](#sigv4) | false |
| tlsConfig | TLS Config to use for remote write. | *[TLSConfig](#tlsconfig) | false |
| proxyUrl | Optional ProxyURL. | string | false |
| queueConfig | QueueConfig allows tuning of the remote write queue parameters. | *[QueueConfig](#queueconfig) | false |
| metadataConfig | MetadataConfig configures the sending of series metadata to the remote storage. | *[MetadataConfig](#metadataconfig) | false |

[Back to TOC](#table-of-contents)

## Rule

Rule describes an alerting or recording rule See Prometheus documentation: [alerting](https://www.prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) or [recording](https://www.prometheus.io/docs/prometheus/latest/configuration/recording_rules/#recording-rules) rule


<em>appears in: [RuleGroup](#rulegroup)</em>

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

RuleGroup is a list of sequentially evaluated recording and alerting rules. Note: PartialResponseStrategy is only used by ThanosRuler and will be ignored by Prometheus instances.  Valid values for this field are 'warn' or 'abort'.  More info: https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response


<em>appears in: [PrometheusRuleSpec](#prometheusrulespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| interval |  | string | false |
| rules |  | [][Rule](#rule) | true |
| partial_response_strategy |  | string | false |

[Back to TOC](#table-of-contents)

## Rules

/--rules.*/ command-line arguments


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| alert |  | [RulesAlert](#rulesalert) | false |

[Back to TOC](#table-of-contents)

## RulesAlert

/--rules.alert.*/ command-line arguments


<em>appears in: [Rules](#rules)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| forOutageTolerance | Max time to tolerate prometheus outage for restoring 'for' state of alert. | string | false |
| forGracePeriod | Minimum duration between alert and restored 'for' state. This is maintained only for alerts with configured 'for' time greater than grace period. | string | false |
| resendDelay | Minimum amount of time to wait before resending an alert to Alertmanager. | string | false |

[Back to TOC](#table-of-contents)

## SafeAuthorization

SafeAuthorization specifies a subset of the Authorization struct, that is safe for use in Endpoints (no CredentialsFile field)


<em>appears in: [AlertmanagerEndpoints](#alertmanagerendpoints), [Endpoint](#endpoint), [PodMetricsEndpoint](#podmetricsendpoint), [ProbeSpec](#probespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Set the authentication type. Defaults to Bearer, Basic will cause an error | string | false |
| credentials | The secret's key that contains the credentials of the request | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## SecretOrConfigMap

SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.


<em>appears in: [OAuth2](#oauth2), [PodMetricsEndpointTLSConfig](#podmetricsendpointtlsconfig), [ProbeTLSConfig](#probetlsconfig), [TLSConfig](#tlsconfig), [WebTLSConfig](#webtlsconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| secret | Secret containing data to use for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| configMap | ConfigMap containing data to use for the targets. | *v1.ConfigMapKeySelector | false |

[Back to TOC](#table-of-contents)

## ServiceMonitor

ServiceMonitor defines monitoring for a set of services.


<em>appears in: [ServiceMonitorList](#servicemonitorlist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of desired Service selection for target discovery by Prometheus. | [ServiceMonitorSpec](#servicemonitorspec) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorList

ServiceMonitorList is a list of ServiceMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of ServiceMonitors | []*[ServiceMonitor](#servicemonitor) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorSpec

ServiceMonitorSpec contains specification parameters for a ServiceMonitor.


<em>appears in: [ServiceMonitor](#servicemonitor)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | Chooses the label of the Kubernetes `Endpoints`. Its value will be used for the `job`-label's value of the created metrics.\n\nDefault & fallback value: the name of the respective Kubernetes `Endpoint`. | string | false |
| targetLabels | TargetLabels transfers labels from the Kubernetes `Service` onto the created metrics. | []string | false |
| podTargetLabels | PodTargetLabels transfers labels on the Kubernetes `Pod` onto the created metrics. | []string | false |
| endpoints | A list of endpoints allowed as part of this ServiceMonitor. | [][Endpoint](#endpoint) | true |
| selector | Selector to select Endpoints objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Kubernetes Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |
| targetLimit | TargetLimit defines a limit on the number of scraped targets that will be accepted. | uint64 | false |
| labelLimit | Per-scrape limit on number of labels that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelNameLengthLimit | Per-scrape limit on length of labels name that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |
| labelValueLengthLimit | Per-scrape limit on length of labels value that will be accepted for a sample. Only valid in Prometheus versions 2.27.0 and newer. | uint64 | false |

[Back to TOC](#table-of-contents)

## Sigv4

Sigv4 optionally configures AWS's Signature Verification 4 signing process to sign requests. Cannot be set at the same time as basic_auth or authorization.


<em>appears in: [RemoteWriteSpec](#remotewritespec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| region | Region is the AWS region. If blank, the region from the default credentials chain used. | string | false |
| accessKey | AccessKey is the AWS API key. If blank, the environment variable `AWS_ACCESS_KEY_ID` is used. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| secretKey | SecretKey is the AWS API secret. If blank, the environment variable `AWS_SECRET_ACCESS_KEY` is used. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| profile | Profile is the named AWS profile used to authenticate. | string | false |
| roleArn | RoleArn is the named AWS profile used to authenticate. | string | false |

[Back to TOC](#table-of-contents)

## StorageSpec

StorageSpec defines the configured storage for a group Prometheus servers. If no storage option is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used. If multiple storage options are specified, priority will be given as follows: EmptyDir, Ephemeral, and lastly VolumeClaimTemplate.


<em>appears in: [AlertmanagerSpec](#alertmanagerspec), [PrometheusSpec](#prometheusspec), [ThanosRulerSpec](#thanosrulerspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disableMountSubPath | Deprecated: subPath usage will be disabled by default in a future release, this option will become unnecessary. DisableMountSubPath allows to remove any subPath usage in volume mounts. | bool | false |
| emptyDir | EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir | *[v1.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#emptydirvolumesource-v1-core) | false |
| ephemeral | EphemeralVolumeSource to be used by the Prometheus StatefulSets. This is a beta field in k8s 1.21, for lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate. More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes | *v1.EphemeralVolumeSource | false |
| volumeClaimTemplate | A PVC spec to be used by the Prometheus StatefulSets. | [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim) | false |

[Back to TOC](#table-of-contents)

## TLSConfig

TLSConfig extends the safe TLS configuration with file parameters.


<em>appears in: [APIServerConfig](#apiserverconfig), [AlertmanagerEndpoints](#alertmanagerendpoints), [Endpoint](#endpoint), [RemoteReadSpec](#remotereadspec), [RemoteWriteSpec](#remotewritespec), [ThanosRulerSpec](#thanosrulerspec), [ThanosSpec](#thanosspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| ca | Struct containing the CA cert to use for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| cert | Struct containing the client cert file for the targets. | [SecretOrConfigMap](#secretorconfigmap) | false |
| keySecret | Secret containing the client key file for the targets. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |
| caFile | Path to the CA cert in the Prometheus container to use for the targets. | string | false |
| certFile | Path to the client cert file in the Prometheus container for the targets. | string | false |
| keyFile | Path to the client key file in the Prometheus container for the targets. | string | false |

[Back to TOC](#table-of-contents)

## ThanosSpec

ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Thanos is being configured. | *string | false |
| version | Version describes the version of Thanos to use. | *string | false |
| tag | Tag of Thanos sidecar container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. Deprecated: use 'image' instead.  The image tag can be specified as part of the image URL. | *string | false |
| sha | SHA of Thanos container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. Deprecated: use 'image' instead.  The image digest can be specified as part of the image URL. | *string | false |
| baseImage | Thanos base image if other than default. Deprecated: use 'image' instead | *string | false |
| resources | Resources defines the resource requirements for the Thanos sidecar. If not provided, no requests/limits will be set | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core) | false |
| objectStorageConfig | ObjectStorageConfig configures object storage in Thanos. Alternative to ObjectStorageConfigFile, and lower order priority. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| objectStorageConfigFile | ObjectStorageConfigFile specifies the path of the object storage configuration file. When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence. | *string | false |
| listenLocal | ListenLocal makes the Thanos sidecar listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| tracingConfig | TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| tracingConfigFile | TracingConfig specifies the path of the tracing configuration file. When used alongside with TracingConfig, TracingConfigFile takes precedence. | string | false |
| grpcServerTlsConfig | GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads recorded rule data. Note: Currently only the CAFile, CertFile, and KeyFile fields are supported. Maps to the '--grpc-server-tls-*' CLI args. | *[TLSConfig](#tlsconfig) | false |
| logLevel | LogLevel for Thanos sidecar to be configured with. | string | false |
| logFormat | LogFormat for Thanos sidecar to be configured with. | string | false |
| minTime | MinTime for Thanos sidecar to be configured with. Option can be a constant time in RFC3339 format or time duration relative to current time, such as -1d or 2h45m. Valid duration units are ms, s, m, h, d, w, y. | string | false |
| readyTimeout | ReadyTimeout is the maximum time Thanos sidecar will wait for Prometheus to start. Eg 10m | string | false |
| volumeMounts | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumeMounts specified will be appended to other VolumeMounts in the thanos-sidecar container. | []v1.VolumeMount | false |

[Back to TOC](#table-of-contents)

## WebSpec

WebSpec defines the query command line flags when starting Prometheus.


<em>appears in: [PrometheusSpec](#prometheusspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| pageTitle | The prometheus web page title | *string | false |
| tlsConfig |  | *[WebTLSConfig](#webtlsconfig) | false |

[Back to TOC](#table-of-contents)

## WebTLSConfig

WebTLSConfig defines the TLS parameters for HTTPS.


<em>appears in: [WebSpec](#webspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| keySecret | Secret containing the TLS key for the server. | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | true |
| cert | Contains the TLS certificate for the server. | [SecretOrConfigMap](#secretorconfigmap) | true |
| clientAuthType | Server policy for client authentication. Maps to ClientAuth Policies. For more detail on clientAuth options: https://golang.org/pkg/crypto/tls/#ClientAuthType | string | false |
| client_ca | Contains the CA certificate for client certificate authentication to the server. | [SecretOrConfigMap](#secretorconfigmap) | false |
| minVersion | Minimum TLS version that is acceptable. Defaults to TLS12. | string | false |
| maxVersion | Maximum TLS version that is acceptable. Defaults to TLS13. | string | false |
| cipherSuites | List of supported cipher suites for TLS versions up to TLS 1.2. If empty, Go default cipher suites are used. Available cipher suites are documented in the go documentation: https://golang.org/pkg/crypto/tls/#pkg-constants | []string | false |
| preferServerCipherSuites | Controls whether the server selects the client's most preferred cipher suite, or the server's most preferred cipher suite. If true then the server's preference, as expressed in the order of elements in cipherSuites, is used. | *bool | false |
| curvePreferences | Elliptic curves that will be used in an ECDHE handshake, in preference order. Available curves are documented in the go documentation: https://golang.org/pkg/crypto/tls/#CurveID | []string | false |

[Back to TOC](#table-of-contents)

## ThanosRuler

ThanosRuler defines a ThanosRuler deployment.


<em>appears in: [ThanosRulerList](#thanosrulerlist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the ThanosRuler cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [ThanosRulerSpec](#thanosrulerspec) | true |
| status | Most recent observed status of the ThanosRuler cluster. Read-only. Not included when requesting from the apiserver, only from the ThanosRuler Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | *[ThanosRulerStatus](#thanosrulerstatus) | false |

[Back to TOC](#table-of-contents)

## ThanosRulerList

ThanosRulerList is a list of ThanosRulers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[ThanosRuler](#thanosruler) | true |

[Back to TOC](#table-of-contents)

## ThanosRulerSpec

ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [ThanosRuler](#thanosruler)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | PodMetadata contains Labels and Annotations gets propagated to the thanos ruler pods. | *[EmbeddedObjectMetadata](#embeddedobjectmetadata) | false |
| image | Thanos container image URL. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling thanos images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core) | false |
| paused | When a ThanosRuler deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| replicas | Number of thanos ruler instances to deploy. | *int32 | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Resources defines the resource requirements for single Pods. If not provided, no requests/limits will be set | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| topologySpreadConstraints | If specified, the pod's topology spread constraints. | []v1.TopologySpreadConstraint | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to the default PodSecurityContext. | *v1.PodSecurityContext | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Thanos Ruler Pods. | string | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| volumes | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects. | []v1.Volume | false |
| objectStorageConfig | ObjectStorageConfig configures object storage in Thanos. Alternative to ObjectStorageConfigFile, and lower order priority. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| objectStorageConfigFile | ObjectStorageConfigFile specifies the path of the object storage configuration file. When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence. | *string | false |
| listenLocal | ListenLocal makes the Thanos ruler listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| queryEndpoints | QueryEndpoints defines Thanos querier endpoints from which to query metrics. Maps to the --query flag of thanos ruler. | []string | false |
| queryConfig | Define configuration for connecting to thanos query instances. If this is defined, the QueryEndpoints field will be ignored. Maps to the `query.config` CLI argument. Only available with thanos v0.11.0 and higher. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| alertmanagersUrl | Define URLs to send alerts to Alertmanager.  For Thanos v0.10.0 and higher, AlertManagersConfig should be used instead.  Note: this field will be ignored if AlertManagersConfig is specified. Maps to the `alertmanagers.url` arg. | []string | false |
| alertmanagersConfig | Define configuration for connecting to alertmanager.  Only available with thanos v0.10.0 and higher.  Maps to the `alertmanagers.config` arg. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| ruleSelector | A label selector to select which PrometheusRules to mount for alerting and recording. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for Rules discovery. If unspecified, only the same namespace as the ThanosRuler object is in is used. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#labelselector-v1-meta) | false |
| enforcedNamespaceLabel | EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert and metric that is user created. The label value will always be the namespace of the object that is being created. | string | false |
| excludedFromEnforcement | List of references to PrometheusRule objects to be excluded from enforcing a namespace label of origin. Applies only if enforcedNamespaceLabel set to true. | [][ObjectReference](#objectreference) | false |
| prometheusRulesExcludedFromEnforce | PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing of adding namespace labels. Works only if enforcedNamespaceLabel set to true. Make sure both ruleNamespace and ruleName are set for each pair Deprecated: use excludedFromEnforcement instead. | [][PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) | false |
| logLevel | Log level for ThanosRuler to be configured with. | string | false |
| logFormat | Log format for ThanosRuler to be configured with. | string | false |
| portName | Port name used for the pods and governing service. This defaults to web | string | false |
| evaluationInterval | Interval between consecutive evaluations. | string | false |
| retention | Time duration ThanosRuler shall retain data for. Default is '24h', and must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | string | false |
| containers | Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or to change the behavior of an operator generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `thanos-ruler` and `config-reloader`. Overriding containers is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| initContainers | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g. fetch secrets for injection into the ThanosRuler configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other then secret fetching is entirely outside the scope of what the maintainers will support and by doing so, you accept that this behaviour may break at any time without notice. | []v1.Container | false |
| tracingConfig | TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| labels | Labels configure the external label pairs to ThanosRuler. A default replica label `thanos_ruler_replica` will be always added  as a label with the value of the pod's name and it will be dropped in the alerts. | map[string]string | false |
| alertDropLabels | AlertDropLabels configure the label names which should be dropped in ThanosRuler alerts. The replica label `thanos_ruler_replica` will always be dropped in alerts. | []string | false |
| externalPrefix | The external URL the Thanos Ruler instances will be available under. This is necessary to generate correct URLs. This is necessary if Thanos Ruler is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path. | string | false |
| grpcServerTlsConfig | GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads recorded rule data. Note: Currently only the CAFile, CertFile, and KeyFile fields are supported. Maps to the '--grpc-server-tls-*' CLI args. | *[TLSConfig](#tlsconfig) | false |
| alertQueryUrl | The external Query URL the Thanos Ruler will set in the 'Source' field of all alerts. Maps to the '--alert.query-url' CLI arg. | string | false |
| minReadySeconds | Minimum number of seconds for which a newly created pod should be ready without any of its container crashing for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready) This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate. | *uint32 | false |
| alertRelabelConfigs | AlertRelabelConfigs configures alert relabeling in ThanosRuler. Alert relabel configurations must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs Alternative to AlertRelabelConfigFile, and lower order priority. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| alertRelabelConfigFile | AlertRelabelConfigFile specifies the path of the alert relabeling configuration file. When used alongside with AlertRelabelConfigs, alertRelabelConfigFile takes precedence. | *string | false |

[Back to TOC](#table-of-contents)

## ThanosRulerStatus

ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status


<em>appears in: [ThanosRuler](#thanosruler)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this ThanosRuler deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this ThanosRuler deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this ThanosRuler deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this ThanosRuler deployment. | int32 | true |

[Back to TOC](#table-of-contents)

## AlertmanagerConfig

AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated across multiple namespaces configuring one Alertmanager cluster.


<em>appears in: [AlertmanagerConfigList](#alertmanagerconfiglist)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta) | false |
| spec |  | [AlertmanagerConfigSpec](#alertmanagerconfigspec) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerConfigList

AlertmanagerConfigList is a list of AlertmanagerConfig.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#listmeta-v1-meta) | false |
| items | List of AlertmanagerConfig | []*[AlertmanagerConfig](#alertmanagerconfig) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerConfigSpec

AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration. By definition, the Alertmanager configuration only applies to alerts for which the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.


<em>appears in: [AlertmanagerConfig](#alertmanagerconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| route | The Alertmanager route definition for alerts matching the resource’s namespace. If present, it will be added to the generated Alertmanager configuration as a first-level route. | *[Route](#route) | true |
| receivers | List of receivers. | [][Receiver](#receiver) | true |
| inhibitRules | List of inhibition rules. The rules will only apply to alerts matching the resource’s namespace. | [][InhibitRule](#inhibitrule) | false |
| muteTimeIntervals | List of MuteTimeInterval specifying when the routes should be muted. | [][MuteTimeInterval](#mutetimeinterval) | false |

[Back to TOC](#table-of-contents)

## DayOfMonthRange

DayOfMonthRange is an inclusive range of days of the month beginning at 1


<em>appears in: [TimeInterval](#timeinterval)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| start | Start of the inclusive range | int | false |
| end | End of the inclusive range | int | false |

[Back to TOC](#table-of-contents)

## EmailConfig

EmailConfig configures notifications via Email.


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| to | The email address to send notifications to. | string | false |
| from | The sender address. | string | false |
| hello | The hostname to identify to the SMTP server. | string | false |
| smarthost | The SMTP host and port through which emails are sent. E.g. example.com:25 | string | false |
| authUsername | The username to use for authentication. | string | false |
| authPassword | The secret's key that contains the password to use for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| authSecret | The secret's key that contains the CRAM-MD5 secret. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| authIdentity | The identity to use for authentication. | string | false |
| headers | Further headers email header key/value pairs. Overrides any headers previously set by the notification implementation. | [][KeyValue](#keyvalue) | false |
| html | The HTML body of the email notification. | string | false |
| text | The text body of the email notification. | string | false |
| requireTLS | The SMTP TLS requirement. Note that Go does not support unencrypted connections to remote SMTP endpoints. | *bool | false |
| tlsConfig | TLS configuration | *monitoringv1.SafeTLSConfig | false |

[Back to TOC](#table-of-contents)

## HTTPConfig

HTTPConfig defines a client HTTP configuration. See https://prometheus.io/docs/alerting/latest/configuration/#http_config


<em>appears in: [OpsGenieConfig](#opsgenieconfig), [PagerDutyConfig](#pagerdutyconfig), [PushoverConfig](#pushoverconfig), [SNSConfig](#snsconfig), [SlackConfig](#slackconfig), [VictorOpsConfig](#victoropsconfig), [WeChatConfig](#wechatconfig), [WebhookConfig](#webhookconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| authorization | Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+. | *monitoringv1.SafeAuthorization | false |
| basicAuth | BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence. | *monitoringv1.BasicAuth | false |
| oauth2 | OAuth2 client credentials used to fetch a token for the targets. | *monitoringv1.OAuth2 | false |
| bearerTokenSecret | The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| tlsConfig | TLS configuration for the client. | *monitoringv1.SafeTLSConfig | false |
| proxyURL | Optional proxy URL. | string | false |
| followRedirects | FollowRedirects specifies whether the client should follow HTTP 3xx redirects. | *bool | false |

[Back to TOC](#table-of-contents)

## InhibitRule

InhibitRule defines an inhibition rule that allows to mute alerts when other alerts are already firing. See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule


<em>appears in: [AlertmanagerConfigSpec](#alertmanagerconfigspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| targetMatch | Matchers that have to be fulfilled in the alerts to be muted. The operator enforces that the alert matches the resource’s namespace. | [][Matcher](#matcher) | false |
| sourceMatch | Matchers for which one or more alerts have to exist for the inhibition to take effect. The operator enforces that the alert matches the resource’s namespace. | [][Matcher](#matcher) | false |
| equal | Labels that must have an equal value in the source and target alert for the inhibition to take effect. | []string | false |

[Back to TOC](#table-of-contents)

## KeyValue

KeyValue defines a (key, value) tuple.


<em>appears in: [EmailConfig](#emailconfig), [OpsGenieConfig](#opsgenieconfig), [PagerDutyConfig](#pagerdutyconfig), [VictorOpsConfig](#victoropsconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| key | Key of the tuple. | string | true |
| value | Value of the tuple. | string | true |

[Back to TOC](#table-of-contents)

## Matcher

Matcher defines how to match on alert's labels.


<em>appears in: [InhibitRule](#inhibitrule), [Route](#route)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Label to match. | string | true |
| value | Label value to match. | string | true |
| matchType | Match operation available with AlertManager >= v0.22.0 and takes precedence over Regex (deprecated) if non-empty. | MatchType | false |
| regex | Whether to match on equality (false) or regular-expression (true). Deprecated as of AlertManager >= v0.22.0 where a user should use MatchType instead. | bool | false |

[Back to TOC](#table-of-contents)

## MuteTimeInterval

MuteTimeInterval specifies the periods in time when notifications will be muted


<em>appears in: [AlertmanagerConfigSpec](#alertmanagerconfigspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the time interval | string | false |
| timeIntervals | TimeIntervals is a list of TimeInterval | [][TimeInterval](#timeinterval) | false |

[Back to TOC](#table-of-contents)

## OpsGenieConfig

OpsGenieConfig configures notifications via OpsGenie. See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| apiKey | The secret's key that contains the OpsGenie API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| apiURL | The URL to send OpsGenie API requests to. | string | false |
| message | Alert text limited to 130 characters. | string | false |
| description | Description of the incident. | string | false |
| source | Backlink to the sender of the notification. | string | false |
| tags | Comma separated list of tags attached to the notifications. | string | false |
| note | Additional alert note. | string | false |
| priority | Priority level of alert. Possible values are P1, P2, P3, P4, and P5. | string | false |
| details | A set of arbitrary key/value pairs that provide further detail about the incident. | [][KeyValue](#keyvalue) | false |
| responders | List of responders responsible for notifications. | [][OpsGenieConfigResponder](#opsgenieconfigresponder) | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |
| entity | Optional field that can be used to specify which domain alert is related to. | string | false |
| actions | Comma separated list of actions that will be available for the alert. | string | false |

[Back to TOC](#table-of-contents)

## OpsGenieConfigResponder

OpsGenieConfigResponder defines a responder to an incident. One of `id`, `name` or `username` has to be defined.


<em>appears in: [OpsGenieConfig](#opsgenieconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| id | ID of the responder. | string | false |
| name | Name of the responder. | string | false |
| username | Username of the responder. | string | false |
| type | Type of responder. | string | true |

[Back to TOC](#table-of-contents)

## PagerDutyConfig

PagerDutyConfig configures notifications via PagerDuty. See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| routingKey | The secret's key that contains the PagerDuty integration key (when using Events API v2). Either this field or `serviceKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| serviceKey | The secret's key that contains the PagerDuty service key (when using integration type \"Prometheus\"). Either this field or `routingKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| url | The URL to send requests to. | string | false |
| client | Client identification. | string | false |
| clientURL | Backlink to the sender of notification. | string | false |
| description | Description of the incident. | string | false |
| severity | Severity of the incident. | string | false |
| class | The class/type of the event. | string | false |
| group | A cluster or grouping of sources. | string | false |
| component | The part or component of the affected system that is broken. | string | false |
| details | Arbitrary key/value pairs that provide further detail about the incident. | [][KeyValue](#keyvalue) | false |
| pagerDutyImageConfigs | A list of image details to attach that provide further detail about an incident. | [][PagerDutyImageConfig](#pagerdutyimageconfig) | false |
| pagerDutyLinkConfigs | A list of link details to attach that provide further detail about an incident. | [][PagerDutyLinkConfig](#pagerdutylinkconfig) | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## PagerDutyImageConfig

PagerDutyImageConfig attaches images to an incident


<em>appears in: [PagerDutyConfig](#pagerdutyconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| src | Src of the image being attached to the incident | string | false |
| href | Optional URL; makes the image a clickable link. | string | false |
| alt | Alt is the optional alternative text for the image. | string | false |

[Back to TOC](#table-of-contents)

## PagerDutyLinkConfig

PagerDutyLinkConfig attaches text links to an incident


<em>appears in: [PagerDutyConfig](#pagerdutyconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| href | Href is the URL of the link to be attached | string | false |
| alt | Text that describes the purpose of the link, and can be used as the link's text. | string | false |

[Back to TOC](#table-of-contents)

## PushoverConfig

PushoverConfig configures notifications via Pushover. See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| userKey | The secret's key that contains the recipient user’s user key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| token | The secret's key that contains the registered application’s API token, see https://pushover.net/apps. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| title | Notification title. | string | false |
| message | Notification message. | string | false |
| url | A supplementary URL shown alongside the message. | string | false |
| urlTitle | A title for supplementary URL, otherwise just the URL is shown | string | false |
| sound | The name of one of the sounds supported by device clients to override the user's default sound choice | string | false |
| priority | Priority, see https://pushover.net/api#priority | string | false |
| retry | How often the Pushover servers will send the same notification to the user. Must be at least 30 seconds. | string | false |
| expire | How long your notification will continue to be retried for, unless the user acknowledges the notification. | string | false |
| html | Whether notification message is HTML or plain text. | bool | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## Receiver

Receiver defines one or more notification integrations.


<em>appears in: [AlertmanagerConfigSpec](#alertmanagerconfigspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the receiver. Must be unique across all items from the list. | string | true |
| opsgenieConfigs | List of OpsGenie configurations. | [][OpsGenieConfig](#opsgenieconfig) | false |
| pagerdutyConfigs | List of PagerDuty configurations. | [][PagerDutyConfig](#pagerdutyconfig) | false |
| slackConfigs | List of Slack configurations. | [][SlackConfig](#slackconfig) | false |
| webhookConfigs | List of webhook configurations. | [][WebhookConfig](#webhookconfig) | false |
| wechatConfigs | List of WeChat configurations. | [][WeChatConfig](#wechatconfig) | false |
| emailConfigs | List of Email configurations. | [][EmailConfig](#emailconfig) | false |
| victoropsConfigs | List of VictorOps configurations. | [][VictorOpsConfig](#victoropsconfig) | false |
| pushoverConfigs | List of Pushover configurations. | [][PushoverConfig](#pushoverconfig) | false |
| snsConfigs | List of SNS configurations | [][SNSConfig](#snsconfig) | false |

[Back to TOC](#table-of-contents)

## Route

Route defines a node in the routing tree.


<em>appears in: [AlertmanagerConfigSpec](#alertmanagerconfigspec)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| receiver | Name of the receiver for this route. If not empty, it should be listed in the `receivers` field. | string | true |
| groupBy | List of labels to group by. Labels must not be repeated (unique list). Special label \"...\" (aggregate by all possible labels), if provided, must be the only element in the list. | []string | false |
| groupWait | How long to wait before sending the initial notification. Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` Example: \"30s\" | string | false |
| groupInterval | How long to wait before sending an updated notification. Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` Example: \"5m\" | string | false |
| repeatInterval | How long to wait before repeating the last notification. Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` Example: \"4h\" | string | false |
| matchers | List of matchers that the alert’s labels should match. For the first level route, the operator removes any existing equality and regexp matcher on the `namespace` label and adds a `namespace: <object namespace>` matcher. | [][Matcher](#matcher) | false |
| continue | Boolean indicating whether an alert should continue matching subsequent sibling nodes. It will always be overridden to true for the first-level route by the Prometheus operator. | bool | false |
| routes | Child routes. | [][apiextensionsv1.JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#json-v1-apiextensions-k8s-io) | false |
| muteTimeIntervals | Note: this comment applies to the field definition above but appears below otherwise it gets included in the generated manifest. CRD schema doesn't support self-referential types for now (see https://github.com/kubernetes/kubernetes/issues/62872). We have to use an alternative type to circumvent the limitation. The downside is that the Kube API can't validate the data beyond the fact that it is a valid JSON representation. MuteTimeIntervals is a list of MuteTimeInterval names that will mute this route when matched, | []string | false |

[Back to TOC](#table-of-contents)

## SNSConfig

SNSConfig configures notifications via AWS SNS. See https://prometheus.io/docs/alerting/latest/configuration/#sns_configs


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| apiURL | The SNS API URL i.e. https://sns.us-east-2.amazonaws.com. If not specified, the SNS API URL from the SNS SDK will be used. | string | false |
| sigv4 | Configures AWS's Signature Verification 4 signing process to sign requests. | *monitoringv1.Sigv4 | false |
| topicARN | SNS topic ARN, i.e. arn:aws:sns:us-east-2:698519295917:My-Topic If you don't specify this value, you must specify a value for the PhoneNumber or TargetARN. | string | false |
| subject | Subject line when the message is delivered to email endpoints. | string | false |
| phoneNumber | Phone number if message is delivered via SMS in E.164 format. If you don't specify this value, you must specify a value for the TopicARN or TargetARN. | string | false |
| targetARN | The  mobile platform endpoint ARN if message is delivered via mobile notifications. If you don't specify this value, you must specify a value for the topic_arn or PhoneNumber. | string | false |
| message | The message content of the SNS notification. | string | false |
| attributes | SNS message attributes. | map[string]string | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## SlackAction

SlackAction configures a single Slack action that is sent with each notification. See https://api.slack.com/docs/message-attachments#action_fields and https://api.slack.com/docs/message-buttons for more information.


<em>appears in: [SlackConfig](#slackconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type |  | string | true |
| text |  | string | true |
| url |  | string | false |
| style |  | string | false |
| name |  | string | false |
| value |  | string | false |
| confirm |  | *[SlackConfirmationField](#slackconfirmationfield) | false |

[Back to TOC](#table-of-contents)

## SlackConfig

SlackConfig configures notifications via Slack. See https://prometheus.io/docs/alerting/latest/configuration/#slack_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| apiURL | The secret's key that contains the Slack webhook URL. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| channel | The channel or user to send notifications to. | string | false |
| username |  | string | false |
| color |  | string | false |
| title |  | string | false |
| titleLink |  | string | false |
| pretext |  | string | false |
| text |  | string | false |
| fields | A list of Slack fields that are sent with each notification. | [][SlackField](#slackfield) | false |
| shortFields |  | bool | false |
| footer |  | string | false |
| fallback |  | string | false |
| callbackId |  | string | false |
| iconEmoji |  | string | false |
| iconURL |  | string | false |
| imageURL |  | string | false |
| thumbURL |  | string | false |
| linkNames |  | bool | false |
| mrkdwnIn |  | []string | false |
| actions | A list of Slack actions that are sent with each notification. | [][SlackAction](#slackaction) | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## SlackConfirmationField

SlackConfirmationField protect users from destructive actions or particularly distinguished decisions by asking them to confirm their button click one more time. See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields for more information.


<em>appears in: [SlackAction](#slackaction)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| text |  | string | true |
| title |  | string | false |
| okText |  | string | false |
| dismissText |  | string | false |

[Back to TOC](#table-of-contents)

## SlackField

SlackField configures a single Slack field that is sent with each notification. Each field must contain a title, value, and optionally, a boolean value to indicate if the field is short enough to be displayed next to other fields designated as short. See https://api.slack.com/docs/message-attachments#fields for more information.


<em>appears in: [SlackConfig](#slackconfig)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| title |  | string | true |
| value |  | string | true |
| short |  | *bool | false |

[Back to TOC](#table-of-contents)

## TimeInterval

TimeInterval describes intervals of time


<em>appears in: [MuteTimeInterval](#mutetimeinterval)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| times | Times is a list of TimeRange | [][TimeRange](#timerange) | false |
| weekdays | Weekdays is a list of WeekdayRange | []WeekdayRange | false |
| daysOfMonth | DaysOfMonth is a list of DayOfMonthRange | [][DayOfMonthRange](#dayofmonthrange) | false |
| months | Months is a list of MonthRange | []MonthRange | false |
| years | Years is a list of YearRange | []YearRange | false |

[Back to TOC](#table-of-contents)

## TimeRange

TimeRange defines a start and end time in 24hr format


<em>appears in: [TimeInterval](#timeinterval)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| startTime | StartTime is the start time in 24hr format. | Time | false |
| endTime | EndTime is the end time in 24hr format. | Time | false |

[Back to TOC](#table-of-contents)

## VictorOpsConfig

VictorOpsConfig configures notifications via VictorOps. See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| apiKey | The secret's key that contains the API key to use when talking to the VictorOps API. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| apiUrl | The VictorOps API URL. | string | false |
| routingKey | A key used to map the alert to a team. | string | true |
| messageType | Describes the behavior of the alert (CRITICAL, WARNING, INFO). | string | false |
| entityDisplayName | Contains summary of the alerted problem. | string | false |
| stateMessage | Contains long explanation of the alerted problem. | string | false |
| monitoringTool | The monitoring tool the state message is from. | string | false |
| customFields | Additional custom fields for notification. | [][KeyValue](#keyvalue) | false |
| httpConfig | The HTTP client's configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## WeChatConfig

WeChatConfig configures notifications via WeChat. See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| apiSecret | The secret's key that contains the WeChat API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| apiURL | The WeChat API URL. | string | false |
| corpID | The corp id for authentication. | string | false |
| agentID |  | string | false |
| toUser |  | string | false |
| toParty |  | string | false |
| toTag |  | string | false |
| message | API request data as defined by the WeChat API. | string | false |
| messageType |  | string | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |

[Back to TOC](#table-of-contents)

## WebhookConfig

WebhookConfig configures notifications via a generic receiver supporting the webhook payload. See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config


<em>appears in: [Receiver](#receiver)</em>

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sendResolved | Whether or not to notify about resolved alerts. | *bool | false |
| url | The URL to send HTTP POST requests to. `urlSecret` takes precedence over `url`. One of `urlSecret` and `url` should be defined. | *string | false |
| urlSecret | The secret's key that contains the webhook URL to send HTTP requests to. `urlSecret` takes precedence over `url`. One of `urlSecret` and `url` should be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#secretkeyselector-v1-core) | false |
| httpConfig | HTTP client configuration. | *[HTTPConfig](#httpconfig) | false |
| maxAlerts | Maximum number of alerts to be sent per webhook message. When 0, all alerts are included. | int32 | false |

[Back to TOC](#table-of-contents)
