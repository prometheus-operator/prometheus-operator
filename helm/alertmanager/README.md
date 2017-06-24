# alertmanager

Installs a [Prometheus](https://prometheus.io) Alertmanager instance using the CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator).

## TL;DR;

```console
$ helm install opsgoodness/alertmanager
```

## Introduction

This chart bootstraps an [Alertmanager](https://github.com/prometheus/alertmanager) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
  - Kubernetes 1.4+ with Beta APIs & ThirdPartyResources enabled
  - [prometheus-operator](https://github.com/coreos/prometheus-operator/blob/master/helm/prometheus-operator/README.md).

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install opsgoodness/alertmanager --name my-release
```

The command deploys Alertmanager  on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the alertmanager chart and their default values.

Parameter | Description | Default
--- | --- | ---
`config` | Alertmanager configuration directives | `{}`
`externalUrl` | External URL at which Alertmanager will be reachable | `""`
`image.repository` | Image | `quay.io/prometheus/alertmanager`
`image.tag` | Image tag | `v0.5.1`
`ingress.enabled` | If true, Alertmanager Ingress will be created | `false`
`ingress.annotations` | Annotations for Alertmanager Ingress` | `{}`
`ingress.fqdn` | Alertmanager Ingress fully-qualified domain name | `""`
`ingress.tls` | TLS configuration for Alertmanager Ingress | `[]`
`nodeSelector` | Node labels for pod assignment | `{}`
`paused` | If true, the Operator won't process any Alertmanager configuration changes | `false`
`replicaCount` | Number of Alertmanager replicas desired | `1`
`resources` | Pod resource requests & limits | `{}`
`service.annotations` | Annotations to be added to the Alertmanager Service | `{}`
`service.clusterIP` | Cluster-internal IP address for Alertmanager Service | `""`
`service.externalIPs` | List of external IP addresses at which the Alertmanager Service will be available | `[]`
`service.loadBalancerIP` | External IP address to assign to Alertmanager Service | `""`
`service.loadBalancerSourceRanges` | List of client IPs allowed to access Alertmanager Service | `[]`
`service.nodePort` | Port to expose Alertmanager Service on each node | `39093`
`service.type` | Alertmanager Service type | `ClusterIP`
`storageSpec` | Alertmanager StorageSpec for persistent data | `{}`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,
$ helm install opsgoodness/alertmanager --name my-release --set externalUrl=http://alertmanager.example.com
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install opsgoodness/alertmanager --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)

### Third-party Resource Documentation
- [Alertmanager](/Documentation/design.md#alertmanager)
- [Prometheus](/Documentation/design.md#prometheus)
- [ServiceMonitor](/Documentation/design.md#servicemonitor)
