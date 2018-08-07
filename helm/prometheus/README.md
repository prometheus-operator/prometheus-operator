# prometheus

Installs a [Prometheus](https://prometheus.io) instance using the CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator).

## TL;DR;

```console
$ helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
$ helm install coreos/prometheus
```

## Introduction

This chart bootstraps a [Prometheus](https://github.com/prometheus/prometheus) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
  - Kubernetes 1.4+ with Beta APIs & ThirdPartyResources enabled
  - [prometheus-operator](https://github.com/coreos/prometheus-operator/blob/master/helm/prometheus-operator/README.md).

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install coreos/prometheus --name my-release
```

The command deploys Prometheus on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the prometheus chart and their default values.

Parameter | Description | Default
--- | --- | ---
`alertingEndpoints` | Alertmanagers to which alerts will be sent | `[]`
`config` | Prometheus configuration directives | `{}`
`externalLabels` | The labels to add to any time series or alerts when communicating with external systems  | `{}`
`externalUrl` | External URL at which Prometheus will be reachable | `""`
`routePrefix` | Prefix used to register routes | `"/"`
`image.repository` | Image | `quay.io/prometheus/prometheus`
`image.tag` | Image tag | `v2.2.1`
`ingress.enabled` | If true, Prometheus Ingress will be created | `false`
`ingress.annotations` | Annotations for Prometheus Ingress` | `{}`
`ingress.labels` | Labels for Prometheus Ingress | `{}`
`ingress.hosts` | Prometheus Ingress fully-qualified domain names | `[]`
`ingress.tls` | TLS configuration for Prometheus Ingress | `[]`
`nodeSelector` | Node labels for pod assignment | `{}`
`paused` | If true, the Operator won't process any Prometheus configuration changes | `false`
`podAntiAffinity` | If "soft", the scheduler attempts to place Prometheus replicas on different nodes. If "hard" the scheduler is required to place them on different nodes. If "" (empty) then no anti-affinity rules will be configured. | `soft`
`prometheusRules` | Prometheus rules | `[templates/prometheus.rules.yaml](templates/prometheus.rules.yaml)`
`replicaCount` | Number of Prometheus replicas desired | `1`
`remoteRead` | The remote_read spec configuration for Prometheus | `{}`
`remoteWrite` | The remote_read spec configuration for Prometheus | `{}`
`resources` | Pod resource requests & limits | `{}`
`retention` | How long to retain metrics | `24h`
`routePrefix` | Prefix used to register routes, overriding externalUrl route | `/`
`rules` | Prometheus alerting & recording rules | `{}`
`ruleNamespaceSelector` | Namespaces to be selected for PrometheusRules discovery | `{}`
`rulesSelector` | Rules CRD selector | `{}`
`secrets` | List of Secrets in the same namespace as the Prometheus object, which shall be mounted into the Prometheus Pods. | `{}`
`service.annotations` | Annotations to be added to the Prometheus Service | `{}`
`service.clusterIP` | Cluster-internal IP address for Prometheus Service | `""`
`service.externalIPs` | List of external IP addresses at which the Prometheus Service will be available | `[]`
`service.loadBalancerIP` | External IP address to assign to Prometheus Service | `""`
`service.loadBalancerSourceRanges` | List of client IPs allowed to access Prometheus Service | `[]`
`service.nodePort` | Port to expose Prometheus Service on each node | `39090`
`service.type` | Prometheus Service type | `ClusterIP`
`serviceMonitors` | ServiceMonitor crd resources to create & be scraped by this Prometheus instance | `[]`
`serviceMonitorsSelector` | ServiceMonitor ConfigMap selector | `{}`
`storageSpec` | Prometheus StorageSpec for persistent data | `{}`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,
$ helm install coreos/prometheus --name my-release --set externalUrl=http://prometheus.example.com


Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example:

```console
$ helm install coreos/prometheus --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)


### Service Monitors

Custom service monitors can be added in values.yaml in the `serviceMonitors` section. Please refere to `values.yaml` for all available parameters.


#### Example service monitor


This example Service Monitor will monitor applications matching `app: nginx-ingress`. The port `metrics` will be scraped with the path `/metrics`. The endpoint will be scraped every 30 seconds.

```
serviceMonitors:
  - name: kube-prometheus-nginx-ingress
    labels:
      prometheus: kube-prometheus
    selector:
      matchLabels:
        app: nginx-ingress
    endpoints:
      - port: metrics
        interval: 30s
        path: /metrics
    namespaceSelector:
      any: true
```

### CRD Resource Documentation
- [Alertmanager](/Documentation/design.md#alertmanager)
- [Prometheus](/Documentation/design.md#prometheus)
- [ServiceMonitor](/Documentation/design.md#servicemonitor)
