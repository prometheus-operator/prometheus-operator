# API Docs

This Document documents the types introduced by the Prometheus Operator to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## AlertingSpec

AlertingSpec defines paramters for alerting configuration of Prometheus servers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| alertmanagers | AlertmanagerEndpoints Prometheus should fire alerts against. | [][AlertmanagerEndpoints](#alertmanagerendpoints) | true |

## Alertmanager

Describes an Alertmanager cluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [v1.ObjectMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_objectmeta) | false |
| spec | Specification of the desired behavior of the Alertmanager cluster. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status | [AlertmanagerSpec](#alertmanagerspec) | true |
| status | Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status | *[AlertmanagerStatus](#alertmanagerstatus) | false |

## AlertmanagerEndpoints

AlertmanagerEndpoints defines a selection of a single Endpoints object containing alertmanager IPs to fire alerts against.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | Namespace of Endpoints object. | string | true |
| name | Name of Endpoints object in Namespace. | string | true |
| port | Port the Alertmanager API is exposed on. | intstr.IntOrString | true |
| scheme | Scheme to use when firing alerts. | string | true |

## AlertmanagerList

A list of Alertmanagers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [unversioned.ListMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_listmeta) | false |
| items | List of Alertmanagers | [][Alertmanager](#alertmanager) | true |

## AlertmanagerSpec

Specification of the desired behavior of the Alertmanager cluster. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version | Version the cluster should be on. | string | false |
| baseImage | Base image that is used to deploy pods. | string | false |
| replicas | Size is the expected size of the alertmanager cluster. The controller will eventually make the size of the running cluster equal to the expected size. | *int32 | false |
| storage | Storage is the definition of how storage will be used by the Alertmanager instances. | *[StorageSpec](#storagespec) | false |
| externalUrl | ExternalURL is the URL under which Alertmanager is externally reachable (for example, if Alertmanager is served via a reverse proxy). Used for generating relative and absolute links back to Alertmanager itself. If the URL has a path portion, it will be used to prefix all HTTP endpoints served by Alertmanager. If omitted, relevant URL components will be derived automatically. | string | false |
| paused | If set to true all actions on the underlaying managed objects are not goint to be performed, except for delete actions. | bool | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |

## AlertmanagerStatus

Most recent observed status of the Alertmanager cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Alertmanager cluster (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Alertmanager cluster that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Alertmanager cluster. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Alertmanager cluster. | int32 | true |

## Endpoint

Endpoint defines a scrapeable endpoint serving Prometheus metrics.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| port | Name of the service port this endpoint refers to. Mutually exclusive with targetPort. | string | false |
| targetPort | Name or number of the target port of the endpoint. Mutually exclusive with port. | intstr.IntOrString | false |
| path | HTTP path to scrape for metrics. | string | false |
| scheme | HTTP scheme to use for scraping. | string | false |
| interval | Interval at which metrics should be scraped | string | false |
| tlsConfig | TLS configuration to use when scraping the endpoint | *[TLSConfig](#tlsconfig) | false |
| bearerTokenFile | File to read bearer token for scraping targets. | string | false |

## NamespaceSelector

A selector for selecting namespaces either selecting all namespaces or a list of namespaces.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| any | Boolean describing whether all namespaces are selected in contrast to a list restricting them. | bool | false |
| matchNames | List of namespace names. | []string | false |

## Prometheus

Prometheus defines a Prometheus deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [v1.ObjectMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_objectmeta) | false |
| spec | Specification of the desired behavior of the Prometheus cluster. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status | [PrometheusSpec](#prometheusspec) | true |
| status | Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status | *[PrometheusStatus](#prometheusstatus) | false |

## PrometheusList

PrometheusList is a list of Prometheuses.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [unversioned.ListMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_listmeta) | false |
| items | List of Prometheuses | []*[Prometheus](#prometheus) | true |

## PrometheusSpec

Specification of the desired behavior of the Prometheus cluster. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| serviceMonitorSelector | ServiceMonitors to be selected for target discovery. | *[unversioned.LabelSelector](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | false |
| version | Version of Prometheus to be deployed. | string | false |
| paused | When a Prometheus deployment is paused, no actions except for deletion will be performed on the underlying objects. | bool | false |
| baseImage | Base image to use for a Prometheus deployment. | string | false |
| replicas | Number of instances to deploy for a Prometheus deployment. | *int32 | false |
| retention | Time duration Prometheus shall retain data for. | string | false |
| externalUrl | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | string | false |
| routePrefix | The route prefix Prometheus registers HTTP handlers for. This is useful, if using ExternalURL and a proxy is rewriting HTTP routes of a request, and the actual ExternalURL is still true, but the server serves requests under a different route prefix. For example for use with `kubectl proxy`. | string | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| ruleSelector | A selector to select which ConfigMaps to mount for loading rule files from. | *[unversioned.LabelSelector](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | false |
| alerting | Define details regarding alerting. | [AlertingSpec](#alertingspec) | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_resourcerequirements) | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the Prometheus Pods. | string | false |
| secrets | Secrets is a list of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. The Secrets are mounted into /etc/prometheus/secrets/<secret-name>. Secrets changes after initial creation of a Prometheus object are not reflected in the running Pods. To change the secrets mounted into the Prometheus Pods, the object must be deleted and recreated with the new list of secrets. | []string | false |

## PrometheusStatus

Most recent observed status of the Prometheus cluster. Read-only. Not included when requesting from the apiserver, only from the Prometheus Operator API itself. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | true |
| replicas | Total number of non-terminated pods targeted by this Prometheus deployment (their labels match the selector). | int32 | true |
| updatedReplicas | Total number of non-terminated pods targeted by this Prometheus deployment that have the desired version spec. | int32 | true |
| availableReplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this Prometheus deployment. | int32 | true |
| unavailableReplicas | Total number of unavailable pods targeted by this Prometheus deployment. | int32 | true |

## ServiceMonitor

ServiceMonitor defines monitoring for a set of services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [v1.ObjectMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_objectmeta) | false |
| spec | Specification of desired Service selection for target discrovery by Prometheus. | [ServiceMonitorSpec](#servicemonitorspec) | true |

## ServiceMonitorList

A list of ServiceMonitors.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata | [unversioned.ListMeta](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_listmeta) | false |
| items | List of ServiceMonitors | []*[ServiceMonitor](#servicemonitor) | true |

## ServiceMonitorSpec

ServiceMonitorSpec contains specification parameters for a ServiceMonitor.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jobLabel | The label to use to retrieve the job name from. | string | false |
| endpoints | A list of endpoints allowed as part of this ServiceMonitor. | [][Endpoint](#endpoint) | false |
| selector | Selector to select Endpoints objects. | [unversioned.LabelSelector](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | true |
| namespaceSelector | Selector to select which namespaces the Endpoints objects are discovered from. | [NamespaceSelector](#namespaceselector) | false |

## StorageSpec

StorageSpec defines the configured storage for a group Prometheus servers.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| class | Name of the StorageClass to use when requesting storage provisioning. More info: https://kubernetes.io/docs/user-guide/persistent-volumes/#storageclasses | string | true |
| selector | A label query over volumes to consider for binding. | *[unversioned.LabelSelector](https://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | true |
| resources | Resources represents the minimum resources the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources | [v1.ResourceRequirements](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_resourcerequirements) | true |

## TLSConfig

TLSConfig specifies TLS configuration parameters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| caFile | The CA cert to use for the targets. | string | false |
| certFile | The client cert file for the targets. | string | false |
| keyFile | The client key file for the targets. | string | false |
| serverName | Used to verify the hostname for the targets. | string | false |
| insecureSkipVerify | Disable target certificate validation. | bool | false |
