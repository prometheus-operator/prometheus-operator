# Prometheus

The `Prometheus` third party resource (TPR) declaratively defines
a desired Prometheus setup to run in a Kubernetes cluster. It provides
options to configure replication, persistent storage, and Alertmanagers to
which the deployed Prometheus instances send alerts to.

For each `Prometheus` TPR, the Operator deploys a properly configured PetSet
in the same namespace. The Prometheus pods are configured to include two
ConfigMaps, `<prometheus-name>` and `<prometheus-name>-rules`, which respectively
hold the used configuration file and multiple Prometheus rule files that may 
contain alerting and recording rules. 

The TPR allows to specify which [`ServiceMonitor`s](./service-monitor.md)
should be covered by the deployed Prometheus instances based on label selection.
The Operator then generates a configuration based on the included `ServiceMonitor`s
and updates it in the ConfigMap. It continuously does so for all changes that
are made to `ServiceMonitor`s or the `Prometheus` TPR itself.

If no selection of `ServiceMonitor`s is provided, the Operator leaves management
of the ConfigMap to the user, which allows to provide custom configurations while
still benefiting from the Operator's capabilities of managing Prometheus setups.

## Specification

### `Prometheus`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| spec | Specification of the Prometheus object | true | PrometheusSpec | |

### `PrometheusSpec`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| serviceMonitorSelector | The `ServiceMonitor` TPRs to be covered by the Prometheus instances. | false | [unversioned.LabelSelector](http://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | |
| version | Prometheus version to deploy. Must match a tag of the container image. | false | string | v1.3.0 |
| paused | If true, the operator won't process any changes affecting the Prometheus setup | false | bool | false |
| baseImage | The base container image (without tag) to use. | false | string | quay.io/prometheus/prometheus |
| replicas | Number of Prometheus instances to deploy. | false | integer (int32) | 1 |
| retention | The duration for which ingested metrics are stored. | false | duration | 24h |
| storage | Configuration of persistent storage volumes to attach to deployed Prometheus pods. | false | StorageSpec |  |
| alerting | Configuration of alerting | false | AlertingSpec |  |
| resources | Resource requirements of single Prometheus server | false | [v1.ResourceRequirements](http://kubernetes.io/docs/api-reference/v1/definitions/#_v1_resourcerequirements) |  | 
| nodeSelector | [Select nodes](https://kubernetes.io/docs/tasks/administer-cluster/assign-pods-nodes/) to be used to run the Prometheus pods on | false | [object](https://kubernetes.io/docs/user-guide/node-selection/) |  |
| externalUrl | External URL Prometheus will be reachable under. Used for generating links, and registering routes. | false | string |  |
| routePrefix | Prefix used to register routes. Overrides `externalUrl` route. Useful for proxies, that rewrite URLs. | false | string |  |

### `StorageSpec`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| class | The storage class to use. | false | string | |
| selector | Selector over candidate persistent volumes. | false | [unversioned.LabelSelector](http://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | |
| resources | Resource requirements for the created persistent volume claim. | false | [v1.ResourceRequirements](http://kubernetes.io/docs/api-reference/v1/definitions/#_v1_resourcerequirements)| |

### `AlertingSpec`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| alertmanagers | Alertmanagers alerts are sent to.  | false | AlertmanagerEndpoints array | |

### `AlertmanagerEndpoints`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| namespace | Namespace of the Alertmanager endpoints. | true | string | |
| name | Name of the Alertmanager endpoints. This equals the targeted Alertmanager service. | true | string | 
| port | Name or number of the service port to push alerts to | false | integer or string |
| scheme | HTTP scheme to use when pushing alerts | false | http |
