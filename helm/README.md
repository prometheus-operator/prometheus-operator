# prometheus-operator

Installs [prometheus-operator](https://github.com/coreos/prometheus-operator) to create/configure/manage Prometheus clusters atop Kubernetes.

## TL;DR;

```console
$ helm install opsgoodness/prometheus-operator
```

## Introduction

This chart bootstraps a [prometheus-operator](https://github.com/coreos/prometheus-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
  - Kubernetes 1.4+ with Beta APIs & ThirdPartyResources enabled

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install opsgoodness/prometheus-operator --name my-release
```

The command deploys prometheus-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the prometheus-operator chart and their default values.

Parameter | Description | Default
--- | --- | ---
`global.hyperkube.repository` | Hyperkube image | `quay.io/coreos/hyperkube`
`global.hyperkube.tag` | Hyperkube image tag | `v1.5.3_coreos.0`
`global.hyperkube.pullPolicy` | Hyperkube image pull policy | `IfNotPresent`
`image.repository` | Image | `quay.io/coreos/prometheus-operator`
`image.tag` | Image tag | `v0.6.0`
`image.pullPolicy` | Image pull policy | `IfNotPresent`
`nodeSelector` | Node labels for pod assignment | `{}`
`resources` | Pod resource requests & limits | `{}`
`sendAnalytics` | Collect & send anonymous usage statistics | `true`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install opsgoodness/prometheus-operator --name my-release --set sendAnalytics=true
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install opsgoodness/prometheus-operator --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)
