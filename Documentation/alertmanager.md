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

## Current state and roadmap

### Config file

The Operator expects a `ConfigMap` of the name `<alertmanager-name>` which
contains the configuration for the Alertmanager instances to run. It is left to
the user to populate it with the desired configuration. Note, that the
Alertmanager pods will stay in a `Pending` state as long as the `ConfigMap`
does not exist.

### Deployment

The Alertmanager, in high availablility mode, is a distributed system. A
desired deployment ensures no data loss and zero downtime while performing a
deployment. Zero downtime is simply done as the Alertmanager is running high
availability mode. No data loss is achieved by using PVCs and attaching the
same volumes a previous Alertmanager instance had to a new instance. The hard
part, however, is knowing whether a new instance is healthy or not.

A healthy node would be one that has joined the existing mesh network and has
been communicated the state that it missed while that particular instance was
down for the upgrade.

Currently there is no way to tell whether an Alertmanager instance is healthy
under the above conditions. There are discussions of using vector clocks to
resolve merges in the above mentioned situation, and ensure on a best effort
basis that joining the network was successful.

> Note that single instance Alertmanager setups will therefore not have zero
> downtime on deployments.

The current implementation of rolling deployments simply decides based on the
Pod state whether an instance is considered healthy. This mechanism may be part
of an implementation with the characteristics that are mentioned above.

### Cluster-wide version

Currently the operator installs a default version with optional explicit
definition of the used version in the TPR.

In the future, there should be a cluster wide version so that the controller
can orchestrate upgrades of all running Alertmanager setups.

