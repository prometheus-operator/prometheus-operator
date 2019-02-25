<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
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
* [BasicAuth](#basicauth)
* [Endpoint](#endpoint)
* [NamespaceSelector](#namespaceselector)
* [Prometheus](#prometheus)
* [PrometheusList](#prometheuslist)
* [PrometheusRule](#prometheusrule)
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
* [ServiceMonitor](#servicemonitor)
* [ServiceMonitorList](#servicemonitorlist)
* [ServiceMonitorSpec](#servicemonitorspec)
* [StorageSpec](#storagespec)
* [TLSConfig](#tlsconfig)
* [ThanosGCSSpec](#thanosgcsspec)
* [ThanosS3Spec](#thanoss3spec)
* [ThanosSpec](#thanosspec)

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
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | [AlertmanagerSpec](#alertmanagerspec) | true |
| status | Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | *[AlertmanagerStatus](#alertmanagerstatus) | false |

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

[Back to TOC](#table-of-contents)

## AlertmanagerList

AlertmanagerList is a list of Alertmanagers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#listmeta-v1-meta) | false |
| items | List of Alertmanagers | [][Alertmanager](#alertmanager) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerSpec

AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata Metadata Labels and Annotations gets propagated to the prometheus pods. | *[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Alertmanager is being configured. | *string | false |
| version | Version the cluster should be on. | string | false |
| tag | Tag of Alertmanager container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. | string | false |
| sha | SHA of Alertmanager container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. | string | false |
| baseImage | Base image that is used to deploy pods, without tag. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#localobjectreference-v1-core) | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>. | []string | false |
| logLevel | Log level for Alertmanager to be configured with. | string | false |
| replicas | Size is the expected size of the alertmanager cluster. The controller will eventually make the size of the running cluster equal to the expected size. | *int32 | false |
| retention | Time duration Alertmanager shall retain data for. Default is '120h', and must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds seconds minutes hours). | string | false |
| storage | Storage is the definition of how storage will be used by the Alertmanager instances. | *[StorageSpec](#storagespec) | false |
| externalUrl | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Alertmanager registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| paused | If set to true all actions on the underlaying managed objects are not goint to be performed, except for delete actions. | bool | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to non root user with uid 1000 and gid 2000. | *v1.PodSecurityContext | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| listenLocal | ListenLocal makes the Alertmanager server listen on loopback, so that it does not bind against the Pod IP. Note this is only for the Alertmanager UI, not the gossip communication. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to an Alertmanager pod. | []v1.Container | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| additionalPeers | AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster. | []string | false |

[Back to TOC](#table-of-contents)

## AlertmanagerStatus

AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Alertmanager cluster (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Alertmanager cluster that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Alertmanager cluster. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Alertmanager cluster. | int32 | true |

[Back to TOC](#table-of-contents)

## BasicAuth

BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username | The secret that contains the username for authenticate | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| password | The secret that contains the password for authenticate | [v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## Endpoint

Endpoint defines a scrapeable endpoint serving Prometheus metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the service port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Name or number of the target port of the endpoint. Mutually exclusive with port. | *intstr.IntOrString | false |
| path | HTTP path to scrape for metrics. | string | false |
| scheme | HTTP scheme to use for scraping. | string | false |
| params | Optional HTTP URL parameters | map[string][]string | false |
| interval | Interval at which metrics should be scraped | string | false |
| scrapeTimeout | Timeout after which the scrape is ended | string | false |
| tlsConfig | TLS configuration to use when scraping the endpoint | *[TLSConfig](#tlsconfig) | false |
| bearerTokenFile | File to read bearer token for scraping targets. | string | false |
| honorLabels | HonorLabels chooses the metric's labels on collisions with target labels. | bool | false |
| basicAuth | BasicAuth allow an endpoint to authenticate over basic authentication More info: https://prometheus.io/docs/operating/configuration/#endpoints | *[BasicAuth](#basicauth) | false |
| metricRelabelings | MetricRelabelConfigs to apply to samples before ingestion. | []*[RelabelConfig](#relabelconfig) | false |
| relabelings | RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config | []*[RelabelConfig](#relabelconfig) | false |
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |

[Back to TOC](#table-of-contents)

## NamespaceSelector

NamespaceSelector is a selector for selecting either all namespaces or a list of namespaces.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| any | Boolean describing whether all namespaces are selected in contrast to a list restricting them. | bool | false |
| matchNames | List of namespace names. | []string | false |

[Back to TOC](#table-of-contents)

## Prometheus

Prometheus defines a Prometheus deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | [PrometheusSpec](#prometheusspec) | true |
| status | Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | *[PrometheusStatus](#prometheusstatus) | false |

[Back to TOC](#table-of-contents)

## PrometheusList

PrometheusList is a list of Prometheuses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[Prometheus](#prometheus) | true |

[Back to TOC](#table-of-contents)

## PrometheusRule

PrometheusRule defines alerting rules for a Prometheus instance

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| spec | Specification of desired alerting rule definitions for Prometheus. | [PrometheusRuleSpec](#prometheusrulespec) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleList

PrometheusRuleList is a list of PrometheusRules.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#listmeta-v1-meta) | false |
| items | List of Rules | []*[PrometheusRule](#prometheusrule) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleSpec

PrometheusRuleSpec contains specification parameters for a Rule.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| groups | Content of Prometheus rule file | [][RuleGroup](#rulegroup) | false |

[Back to TOC](#table-of-contents)

## PrometheusSpec

PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata Metadata Labels and Annotations gets propagated to the prometheus pods. | *[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| serviceMonitorSelector | ServiceMonitors to be selected for target discovery. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) | false |
| serviceMonitorNamespaceSelector | Namespaces to be selected for ServiceMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) | false |
| version | Version of Prometheus to be deployed. | string | false |
| tag | Tag of Prometheus container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. | string | false |
| sha | SHA of Prometheus container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. | string | false |
| paused | When a Prometheus deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Prometheus is being configured. | *string | false |
| baseImage | Base image to use for a Prometheus deployment. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#localobjectreference-v1-core) | false |
| replicas | Number of instances to deploy for a Prometheus deployment. | *int32 | false |
| replicaExternalLabelName | Name of Prometheus external label used to denote replica name. Defaults to the value of `prometheus_replica`. | string | false |
| prometheusExternalLabelName | Name of Prometheus external label used to denote Prometheus instance name. Defaults to the value of `prometheus`. External label will _not_ be added when value is set to empty string (`\"\"`). | *string | false |
| retention | Time duration Prometheus shall retain data for. Default is '24h', and must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | string | false |
| logLevel | Log level for Prometheus to be configured with. | string | false |
| logFormat | Log format for Prometheus to be configured with. | string | false |
| scrapeInterval | Interval between consecutive scrapes. | string | false |
| evaluationInterval | Interval between consecutive evaluations. | string | false |
| rules | /--rules.*/ command-line arguments. | [Rules](#rules) | false |
| externalLabels | The labels to add to any time series or alerts when communicating with external systems (federation, remote storage, Alertmanager). | map[string]string | false |
| enableAdminAPI | Enable access to prometheus web admin API. Defaults to the value of `false`. WARNING: Enabling the admin APIs enables mutating endpoints, to delete data, shutdown Prometheus, and more. Enabling this should be done with care and the user is advised to add additional authentication authorization via a proxy to ensure only clients authorized to perform these actions can do so. For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis | bool | false |
| externalUrl | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Prometheus registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| query | QuerySpec defines the query command line flags when starting Prometheus. | *[QuerySpec](#queryspec) | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| ruleSelector | A selector to select which PrometheusRules to mount for loading alerting rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom resources selected by RuleSelector. Make sure it does not match any config maps that you do not want to be migrated. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for PrometheusRules discovery. If unspecified, only the same namespace as the Prometheus object is in is used. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) | false |
| alerting | Define details regarding alerting. | *[AlertingSpec](#alertingspec) | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#resourcerequirements-v1-core) | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The Secrets are mounted into /etc/prometheus/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>. | []string | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| remoteWrite | If specified, the remote_write spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteWriteSpec](#remotewritespec) | false |
| remoteRead | If specified, the remote_read spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteReadSpec](#remotereadspec) | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to non root user with uid 1000 and gid 2000 for Prometheus >v2.0 and default PodSecurityContext for other versions. | *v1.PodSecurityContext | false |
| listenLocal | ListenLocal makes the Prometheus server listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to a Prometheus pod. | []v1.Container | false |
| additionalScrapeConfigs | AdditionalScrapeConfigs allows specifying a key of a Secret containing additional Prometheus scrape configurations. Scrape configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config. As scrape configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible scrape configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| additionalAlertRelabelConfigs | AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing additional Prometheus alert relabel configurations. Alert relabel configurations specified are appended to the configurations generated by the Prometheus Operator. Alert relabel configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs. As alert relabel configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible alert relabel configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| additionalAlertManagerConfigs | AdditionalAlertManagerConfigs allows specifying a key of a Secret containing additional Prometheus AlertManager configurations. AlertManager configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config. As AlertManager configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible AlertManager configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| apiserverConfig | APIServerConfig allows specifying a host and auth methods to access apiserver. If left empty, Prometheus is assumed to run inside of the cluster and will discover API servers automatically and use the pod's CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. | *[APIServerConfig](#apiserverconfig) | false |
| thanos | Thanos configuration allows configuring various aspects of a Prometheus server in a Thanos environment.\n\nThis section is experimental, it may change significantly without deprecation notice in any release.\n\nThis is experimental and may change significantly without backward compatibility in any release. | *[ThanosSpec](#thanosspec) | false |
| priorityClassName | Priority class assigned to the Pods | string | false |

[Back to TOC](#table-of-contents)

## PrometheusStatus

PrometheusStatus is the most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

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
| regex | Regular expression against which the extracted value is matched. defailt is '(.*)' | string | false |
| modulus | Modulus to take of the hash of the source label values. | uint64 | false |
| replacement | Replacement value against which a regex replace is performed if the regular expression matches. Regex capture groups are available. Default is '$1' | string | false |
| action | Action to perform based on regex matching. Default is 'replace' | string | false |

[Back to TOC](#table-of-contents)

## RemoteReadSpec

RemoteReadSpec defines the remote_read configuration for prometheus.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | The URL of the endpoint to send samples to. | string | true |
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

RuleGroup is a list of sequentially evaluated recording and alerting rules.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| interval |  | string | false |
| rules |  | [][Rule](#rule) | true |

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

## ServiceMonitor

ServiceMonitor defines monitoring for a set of services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#objectmeta-v1-meta) | false |
| spec | Specification of desired Service selection for target discrovery by Prometheus. | [ServiceMonitorSpec](#servicemonitorspec) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorList

ServiceMonitorList is a list of ServiceMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#listmeta-v1-meta) | false |
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
| selector | Selector to select Endpoints objects. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |
| sampleLimit | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. | uint64 | false |

[Back to TOC](#table-of-contents)

## StorageSpec

StorageSpec defines the configured storage for a group Prometheus servers. If neither `emptyDir` nor `volumeClaimTemplate` is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| emptyDir | EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir | *[v1.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#emptydirvolumesource-v1-core) | false |
| volumeClaimTemplate | A PVC spec to be used by the Prometheus StatefulSets. | [v1.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#persistentvolumeclaim-v1-core) | false |

[Back to TOC](#table-of-contents)

## TLSConfig

TLSConfig specifies TLS configuration parameters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| caFile | The CA cert to use for the targets. | string | false |
| certFile | The client cert file for the targets. | string | false |
| keyFile | The client key file for the targets. | string | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |

[Back to TOC](#table-of-contents)

## ThanosGCSSpec

Deprecated: ThanosGCSSpec should be configured with an ObjectStorageConfig secret starting with Thanos v0.2.0. ThanosGCSSpec will be removed.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| bucket | Google Cloud Storage bucket name for stored blocks. If empty it won't store any block inside Google Cloud Storage. | *string | false |
| credentials | Secret to access our Bucket. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## ThanosS3Spec

Deprecated: ThanosS3Spec should be configured with an ObjectStorageConfig secret starting with Thanos v0.2.0. ThanosS3Spec will be removed.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| bucket | S3-Compatible API bucket name for stored blocks. | *string | false |
| endpoint | S3-Compatible API endpoint for stored blocks. | *string | false |
| accessKey | AccessKey for an S3-Compatible API. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| secretKey | SecretKey for an S3-Compatible API. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| insecure | Whether to use an insecure connection with an S3-Compatible API. | *bool | false |
| signatureVersion2 | Whether to use S3 Signature Version 2; otherwise Signature Version 4 will be used. | *bool | false |
| encryptsse | Whether to use Server Side Encryption | *bool | false |

[Back to TOC](#table-of-contents)

## ThanosSpec

ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| peers | Peers is a DNS name for Thanos to discover peers through. | *string | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the Prometheus Operator knows what version of Thanos is being configured. | *string | false |
| version | Version describes the version of Thanos to use. | *string | false |
| tag | Tag of Thanos sidecar container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. | *string | false |
| sha | SHA of Thanos container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. | *string | false |
| baseImage | Thanos base image if other than default. | *string | false |
| resources | Resources defines the resource requirements for the Thanos sidecar. If not provided, no requests/limits will be set | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#resourcerequirements-v1-core) | false |
| gcs | Deprecated: GCS should be configured with an ObjectStorageConfig secret starting with Thanos v0.2.0. This field will be removed. | *[ThanosGCSSpec](#thanosgcsspec) | false |
| s3 | Deprecated: S3 should be configured with an ObjectStorageConfig secret starting with Thanos v0.2.0. This field will be removed. | *[ThanosS3Spec](#thanoss3spec) | false |
| objectStorageConfig | ObjectStorageConfig configures object storage in Thanos. | *[v1.SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#secretkeyselector-v1-core) | false |
| grpcAdvertiseAddress | Explicit (external) host:port address to advertise for gRPC StoreAPI in gossip cluster. If empty, 'grpc-address' will be used. | *string | false |
| clusterAdvertiseAddress | Explicit (external) ip:port address to advertise for gossip in gossip cluster. Used internally for membership only. | *string | false |

[Back to TOC](#table-of-contents)
