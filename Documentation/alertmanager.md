# Alertmanager

The `Alertmanager` third party resource (TPR) declaratively defines a desired
Alertmanager setup to run in a Kubernetes cluster. It provides options to
configure replication and persistent storage.

For each `Alertmanager` TPR, the Operator deploys a properly configured PetSet
in the same namespace. The Alertmanager pods are configured to include a
ConfigMap called `<alertmanager-name>` which holds the used configuration file.

When the configured replicas is two or more the operator runs the Alertmanager
instances in high availability mode.

## Specification

### `Alertmanager`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| spec | Specification of the Alertmanager object | true | AlertmanagerSpec | |

### `AlertmanagerSpec`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| version | Alertmanager version to deploy. Must match a tag of the container image. | false | string | v0.5.0 |
| baseImage | The base container image (without tag) to use. | false | string | quay.io/prometheus/alertmanager |
| replicas | Number of Alertmanager instances to deploy. | false | integer (int32) | 1 |
| storage | Configuration of persistent storage volumes to attach to deployed Alertmanager pods. | false | [StorageSpec](prometheus.md#storagespec) |  |
| externalUrl | External URL Alertmanager will be reachable under. Used for registering routes. | false | string |  |
| paused | If true, the operator won't process any changes affecting the Alertmanager setup | false | bool | false |

