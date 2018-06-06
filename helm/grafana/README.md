# Grafana Helm Chart

* Installs the web dashboarding system [Grafana](http://grafana.org/)

## TL;DR;

```console
$ helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
$ helm install coreos/grafana
```
## Introduction

This chart bootstraps an [Grafana](http://grafana.org) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager, and preinstalled some defaut dashboard for dashboarding Kubernetes metrics.

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release coreos/grafana
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
`routePrefix` | Prefix used to register routes | `"/"`
`auth.anonymous.enabled` | If true, enable anonymous authentication | `true`
`adminUser` | Grafana admin user name | `admin`
`adminPassword` | Grafana admin user password | `admin`
`image.repository` | Image | `grafana/grafana`
`image.tag` | Image tag | `4.4.1`
`extraVars` | Pass extra environment variables to the Grafana container. | `{}`
`grafanaWatcher.repository` | Image | `quay.io/coreos/grafana-watcher`
`grafanaWatcher.tag` | Image tag | `v0.0.8`
`ingress.enabled` | If true, Grafana Ingress will be created | `false`
`ingress.annotations` | Annotations for Grafana Ingress | `{}`
`ingress.labels` | Labels for Grafana Ingress | `{}`
`ingress.hosts` | Grafana Ingress fully-qualified domain names | `[]`
`ingress.tls` | TLS configuration for Grafana Ingress | `[]`
`nodeSelector` | Node labels for pod assignment | `{}`
`resources` | Pod resource requests & limits | `{}`
`service.annotations` | Annotations to be added to the Grafana Service | `{}`
`service.clusterIP` | Cluster-internal IP address for Grafana Service | `""`
`service.externalIPs` | List of external IP addresses at which the Grafana Service will be available | `[]`
`service.labels` | Labels for Grafana Service | `{}`
`service.loadBalancerIP` | External IP address to assign to Grafana Service | `""`
`service.loadBalancerSourceRanges` | List of client IPs allowed to access Grafana Service | `[]`
`service.nodePort` | Port to expose Grafana Service on each node | `30902`
`service.type` | Grafana Service type | `ClusterIP`
`storageSpec` | Grafana StorageSpec for persistent data | `{}`
`resources` | Pod resource requests & limits | `{}`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,
$ helm install coreos/grafana --name my-release --set adminUser=bob
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install coreos/grafana --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)

> **Tip**: On GCE If you want to  use  `Ingress.enabled=true`, you must put `service.type=NodePort`

## Adding Grafana Dashboards

You can either add new dashboards via `serverDashboardConfigmaps` in `values.yaml`. These can then be
picked up by Grafana Watcher.

```yaml
serverDashboardConfigmaps:
  - example-dashboards
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-dashboards
data:
{{ (.Files.Glob "custom-dashboards/*.json").AsConfig | indent 2 }}

# Or
#
# data:
#   custom-dashboard.json: |-
# {{ (.Files.Get "custom.json") | indent 4 }}
#
# The filename (and consequently the key under data) must be in the format `xxx-dashboard.json` or `xxx-datasource.json`
# for them to be picked up.
```

Another way is to add them through `serverDashboardFiles` directly in `values.yaml`. These are then combined
into the same ConfigMap for the rest of the default dashboards.

```yaml
serverDashboardFiles:
  example-dashboard.json: |-
    {
      "dashboard": {
        "annotations:[]
        ...
      }
    }
```

In both cases, if you're exporting the jsons directly from Grafana, you'll want to wrap it in `{"dashboard": {}}`
as stated in [Grafana Watcher's README](https://github.com/coreos/prometheus-operator/tree/master/contrib/grafana-watcher).
