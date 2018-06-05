<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# API Docs

This Document documents the types introduced by the Prometheus Operator to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## Table of Contents
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
* [RelabelConfig](#relabelconfig)
* [RemoteReadSpec](#remotereadspec)
* [RemoteWriteSpec](#remotewritespec)
* [Rule](#rule)
* [RuleGroup](#rulegroup)
* [ServiceMonitor](#servicemonitor)
* [ServiceMonitorList](#servicemonitorlist)
* [ServiceMonitorSpec](#servicemonitorspec)
* [StorageSpec](#storagespec)
* [TLSConfig](#tlsconfig)

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
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
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

A list of Alertmanagers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#listmeta-v1-meta) | false |
| items | List of Alertmanagers | [][Alertmanager](#alertmanager) | true |

[Back to TOC](#table-of-contents)

## AlertmanagerSpec

Specification of the desired behavior of the Alertmanager cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata Metadata Labels and Annotations gets propagated to the prometheus pods. | *[metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
| version | Version the cluster should be on. | string | false |
| baseImage | Base image that is used to deploy pods, without tag. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#localobjectreference-v1-core) | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Alertmanager object, which shall be mounted into the Alertmanager Pods. The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>. | []string | false |
| logLevel | Log level for Alertmanager to be configured with. | string | false |
| replicas | Size is the expected size of the alertmanager cluster. The controller will eventually make the size of the running cluster equal to the expected size. | *int32 | false |
| storage | Storage is the definition of how storage will be used by the Alertmanager instances. | *[StorageSpec](#storagespec) | false |
| externalUrl | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Alertmanager registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| paused | If set to true all actions on the underlaying managed objects are not goint to be performed, except for delete actions. | bool | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#resourcerequirements-v1-core) | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to non root user with uid 1000 and gid 2000. | *v1.PodSecurityContext | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| listenLocal | ListenLocal makes the Alertmanager server listen on loopback, so that it does not bind against the Pod IP. Note this is only for the Alertmanager UI, not the gossip communication. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to an Alertmanager pod. | []v1.Container | false |

[Back to TOC](#table-of-contents)

## AlertmanagerStatus

Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

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
| username | The secret that contains the username for authenticate | [v1.SecretKeySelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#secretkeyselector-v1-core) | false |
| password | The secret that contains the password for authenticate | [v1.SecretKeySelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## Endpoint

Endpoint defines a scrapeable endpoint serving Prometheus metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the service port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Name or number of the target port of the endpoint. Mutually exclusive with port. | intstr.IntOrString | false |
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
| proxyUrl | ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint. | *string | false |

[Back to TOC](#table-of-contents)

## NamespaceSelector

A selector for selecting namespaces either selecting all namespaces or a list of namespaces.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| any | Boolean describing whether all namespaces are selected in contrast to a list restricting them. | bool | false |
| matchNames | List of namespace names. | []string | false |

[Back to TOC](#table-of-contents)

## Prometheus

Prometheus defines a Prometheus deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
| spec | Specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | [PrometheusSpec](#prometheusspec) | true |
| status | Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status | *[PrometheusStatus](#prometheusstatus) | false |

[Back to TOC](#table-of-contents)

## PrometheusList

PrometheusList is a list of Prometheuses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#listmeta-v1-meta) | false |
| items | List of Prometheuses | []*[Prometheus](#prometheus) | true |

[Back to TOC](#table-of-contents)

## PrometheusRule

PrometheusRule defines alerting rules for a Prometheus instance

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
| spec | Specification of desired alerting rule definitions for Prometheus. | [PrometheusRuleSpec](#prometheusrulespec) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleList

A list of PrometheusRules.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [metav1.ListMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#listmeta-v1-meta) | false |
| items | List of Rules | []*[PrometheusRule](#prometheusrule) | true |

[Back to TOC](#table-of-contents)

## PrometheusRuleSpec

PrometheusRuleSpec contains specification parameters for a Rule.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| groups | Content of Prometheus rule file | [][RuleGroup](#rulegroup) | false |

[Back to TOC](#table-of-contents)

## PrometheusSpec

Specification of the desired behavior of the Prometheus cluster. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata Metadata Labels and Annotations gets propagated to the prometheus pods. | *[metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
| serviceMonitorSelector | ServiceMonitors to be selected for target discovery. | *[metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | false |
| serviceMonitorNamespaceSelector | Namespaces to be selected for ServiceMonitor discovery. If nil, only check own namespace. | *[metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | false |
| version | Version of Prometheus to be deployed. | string | false |
| paused | When a Prometheus deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| baseImage | Base image to use for a Prometheus deployment. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling prometheus and alertmanager images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#localobjectreference-v1-core) | false |
| replicas | Number of instances to deploy for a Prometheus deployment. | *int32 | false |
| retention | Time duration Prometheus shall retain data for. | string | false |
| logLevel | Log level for Prometheus to be configured with. | string | false |
| scrapeInterval | Interval between consecutive scrapes. | string | false |
| evaluationInterval | Interval between consecutive evaluations. | string | false |
| externalLabels | The labels to add to any time series or alerts when communicating with external systems (federation, remote storage, Alertmanager). | map[string]string | false |
| externalUrl | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Prometheus registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| ruleSelector | A selector to select which PrometheusRules to mount for loading alerting rules from. | *[metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | false |
| ruleNamespaceSelector | Namespaces to be selected for PrometheusRules discovery. If unspecified, only the same namespace as the Prometheus object is in is used. | *[metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | false |
| alerting | Define details regarding alerting. | *[AlertingSpec](#alertingspec) | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#resourcerequirements-v1-core) | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The Secrets are mounted into /etc/prometheus/secrets/<secret-name>. Secrets changes after initial creation of a Prometheus object are not reflected in the running Pods. To change the secrets mounted into the Prometheus Pods, the object must be deleted and recreated with the new list of secrets. | []string | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| remoteWrite | If specified, the remote_write spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteWriteSpec](#remotewritespec) | false |
| remoteRead | If specified, the remote_read spec. This is an experimental feature, it may change in any upcoming release in a breaking way. | [][RemoteReadSpec](#remotereadspec) | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to non root user with uid 1000 and gid 2000 for Prometheus >v2.0 and default PodSecurityContext for other versions. | *v1.PodSecurityContext | false |
| listenLocal | ListenLocal makes the Prometheus server listen on loopback, so that it does not bind against the Pod IP. | bool | false |
| containers | Containers allows injecting additional containers. This is meant to allow adding an authentication proxy to a Prometheus pod. | []v1.Container | false |
| additionalScrapeConfigs | AdditionalScrapeConfigs allows specifying a key of a Secret containing additional Prometheus scrape configurations. Scrape configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<scrape_config>. As scrape configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible scrape configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#secretkeyselector-v1-core) | false |
| additionalAlertManagerConfigs | AdditionalAlertManagerConfigs allows specifying a key of a Secret containing additional Prometheus AlertManager configurations. AlertManager configurations specified are appended to the configurations generated by the Prometheus Operator. Job configurations specified must have the form as specified in the official Prometheus documentation: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<alertmanager_config>. As AlertManager configs are appended, the user is responsible to make sure it is valid. Note that using this feature may expose the possibility to break upgrades of Prometheus. It is advised to review Prometheus release notes to ensure that no incompatible AlertManager configs are going to break Prometheus after the upgrade. | *[v1.SecretKeySelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#secretkeyselector-v1-core) | false |

[Back to TOC](#table-of-contents)

## PrometheusStatus

Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Prometheus deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Prometheus deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Prometheus deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Prometheus deployment. | int32 | true |

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

[Back to TOC](#table-of-contents)

## Rule

Rule describes an alerting or recording rule.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| record |  | string | false |
| alert |  | string | false |
| expr |  | string | true |
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

## ServiceMonitor

ServiceMonitor defines monitoring for a set of services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#objectmeta-v1-meta) | false |
| spec | Specification of desired Service selection for target discrovery by Prometheus. | [ServiceMonitorSpec](#servicemonitorspec) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorList

A list of ServiceMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ListMeta](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#listmeta-v1-meta) | false |
| items | List of ServiceMonitors | []*[ServiceMonitor](#servicemonitor) | true |

[Back to TOC](#table-of-contents)

## ServiceMonitorSpec

ServiceMonitorSpec contains specification parameters for a ServiceMonitor.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | The label to use to retrieve the job name from. | string | false |
| targetLabels | TargetLabels transfers labels on the Kubernetes Service onto the target. | []string | false |
| endpoints | A list of endpoints allowed as part of this ServiceMonitor. | [][Endpoint](#endpoint) | true |
| selector | Selector to select Endpoints objects. | [metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |

[Back to TOC](#table-of-contents)

## StorageSpec

StorageSpec defines the configured storage for a group Prometheus servers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| class | Name of the StorageClass to use when requesting storage provisioning. More info: https://kubernetes.io/docs/user-guide/persistent-volumes/#storageclasses DEPRECATED | string | false |
| emptyDir | EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir | *[v1.EmptyDirVolumeSource](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#emptydirvolumesource-v1-core) | false |
| selector | A label query over volumes to consider for binding. DEPRECATED | *[metav1.LabelSelector](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#labelselector-v1-meta) | false |
| resources | Resources represents the minimum resources the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources DEPRECATED | [v1.ResourceRequirements](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#resourcerequirements-v1-core) | false |
| volumeClaimTemplate | A PVC spec to be used by the Prometheus StatefulSets. | [v1.PersistentVolumeClaim](https://v1-6.docs.kubernetes.io/docs/api-reference/v1.6/#persistentvolumeclaim-v1-core) | false |

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
