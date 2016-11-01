# kube-prometheus

This repository collects Kubernetes manifests, dashboards, and alerting rules
combined with documentation and scripts to deploy them to get a full cluster 
monitoring setup working.

## Prerequisites

First, you need a running Kubernetes cluster. If you don't have one, follow
the instructions of [bootkube](https://github.com/kubernetes-incubator/bootkube)
or [minikube](https://github.com/kubernetes/minikube).

etcd is an important component of a working Kubernetes cluster, but it's deployed
outside of it. The monitoring setup below assumes that it is made visible from
within the cluster through a headless Kubernetes service.
An example for bootkube's multi-vagrant setup is [here](/manifests/etcd/etcd-bootkube-vagrant-multi.yaml).

## Monitoring Kubernetes

The manifests used here use the [Prometheus controller](https://github.com/coreos/kube-prometheus-controller),
which manages Prometheus servers and their configuration in your cluster. To install the
controller, the [node_exporter](https://github.com/prometheus/node_exporter),
[Grafana](https://grafana.org) including default dashboards, and the Prometheus server, run:

```bash
export KUBECONFIG=<path>          # defaults to "~/.kube/config"
hack/cluster-monitoring/deploy
```

After all pods are ready, you can reach:

* Prometheus UI on node port `30900`
* Grafana on node port `30902`

To tear it all down again, run:

```bash
hack/cluster-monitoring/teardown
```

*All services in the manifest still contain the `prometheus.io/scrape = true` annotations. It is not
used by the Prometheus controller. They remain for convential deployments as in
[this example configuration](https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus-kubernetes.yml).*

## Monitoring custom services

The example manifests in [/manifests/examples/example-app](/manifests/examples/example-app)
deploy a fake service into the `production` and `development` namespaces and define
a Prometheus server monitoring them.

```bash
kubectl --kubeconfig="$KUBECONFIG" create namespace production
kubectl --kubeconfig="$KUBECONFIG" create namespace development
hack/example-service-monitoring/deploy
```

After all pods are ready you can reach the Prometheus server monitoring your services
on node port `30100`.

Teardown:

```bash
hack/example-service-monitoring/teardown
```

## Dashboarding

The provided manifests deploy a Grafana instance serving dashboards provided via a ConfigMap.
To modify, delete, or add dashboards, the `grafana-dashboards` ConfigMap must be modified.

Currently, Grafana does not support serving dashboards from static files. Instead, the `grafana-watcher`
sidecar container aims to emulate the behavior, by keeping the Grafana database always in sync
with the provided ConfigMap. Hence, the Grafana pod is effectively stateless.
This allows managing dashboards via `git` etc. and easily deploying them via CD pipelines.

In the future, a separate Grafana controller should support gathering dashboards from multiple
ConfigMaps, which are selected by their labels.
Prometheus servers deployed by the Prometheus controller should be automatically added as
Grafana data sources.  

## Roadmap

* Incorporate [Alertmanager controller](https://github.com/coreos/kube-alertmanager-controller)
* Grafana controller that dynamically discovers and deploys dashboards from ConfigMaps
* Collection of base alerting for cluster monitoring
* KPM/Helm packages to easily provide production-ready cluster-monitoring setup (essentially contents of `hack/cluster-monitoring`)
* Add meta-monitoring to default cluster monitoring setup


