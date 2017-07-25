# Grafana Helm Chart

* Installs the web dashboarding system [Grafana](http://grafana.org/)

## TL;DR;

```console
$ helm install opsgoodness/grafana
```
## Introduction

This chart bootstraps an [Grafana](http://grafana.org) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager, and preinstalled some defaut dashboard for dashboarding Kubernetes metrics.

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release opsgoodness/grafana
```

## Uninstalling the Chart

To uninstall/delete the my-release deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.


## Configuration

Parameter | Description | Default
--- | --- | ---
`adminUser` | Grafana admin user name | `admin`
`adminPassword` | Grafana admin user password | `admin`
`image.repository` | Image | `grafana/grafana`
`image.tag` | Image tag | `4.4.1`
`ingress.enabled` | If true, Grafana Ingress will be created | `false`
`ingress.annotations` | Annotations for Grafana Ingress | `{}`
`ingress.fqdn` | Grafana Ingress fully-qualified domain name | `""`
`ingress.tls` | TLS configuration for Grafana Ingress | `[]`
`nodeSelector` | Node labels for pod assignment | `{}`
`resources` | Pod resource requests & limits | `{}`
`service.annotations` | Annotations to be added to the Grafana Service | `{}`
`service.clusterIP` | Cluster-internal IP address for Grafana Service | `""`
`service.externalIPs` | List of external IP addresses at which the Grafana Service will be available | `[]`
`service.loadBalancerIP` | External IP address to assign to Grafana Service | `""`
`service.loadBalancerSourceRanges` | List of client IPs allowed to access Grafana Service | `[]`
`service.nodePort` | Port to expose Grafana Service on each node | `39093`
`service.type` | Grafana Service type | `ClusterIP`
`storageSpec` | Grafana StorageSpec for persistent data | `{}`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,
$ helm install opsgoodness/grafana --name my-release --set adminUser=bob
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install opsgoodness/grafana --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)
