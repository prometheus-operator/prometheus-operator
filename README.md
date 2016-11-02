# kube-prometheus

This repository collects Kubernetes manifests, dashboards, and alerting rules
combined with documentation and scripts to deploy them to get a full cluster 
monitoring setup working.

## Prerequisites

First, you need a running Kubernetes cluster. If you don't have one, follow the
instructions of [bootkube](https://github.com/kubernetes-incubator/bootkube) or
[minikube](https://github.com/kubernetes/minikube). Some sample contents of this
repository are adapted to work with a [multi-node setup](https://github.com/kubernetes-incubator/bootkube/tree/master/hack/multi-node)
using [bootkube](https://github.com/kubernetes-incubator/bootkube).

Prometheus discovers targets via Kubernetes endpoints objects, which are automatically
populated by Kubernetes services. Therefore Prometheus can
automatically find and pick up all services within a cluster. By
default there is a service for the Kubernetes API server. For other Kubernetes
core components to be monitored, headless services must be setup for them to be
discovered by Prometheus as they may be deployed differently depending
on the cluster.

For the `kube-scheduler` and `kube-controller-manager` there are headless
services prepared, simply add them to your running cluster:

```bash
kubectl -n kube-system create manifests/k8s/
```

> Hint: if you use this for a cluster not created with bootkube, make sure you
> populate an endpoints object with the address to your `kube-scheduler` and
> `kube-controller-manager`, or adapt the label selectors to match your setup.

Aside from Kubernetes specific components, etcd is an important part of a
working cluster, but is typically deployed outside of it. This monitoring
setup assumes that it is made visible from within the cluster through a headless
service as well.

An example for bootkube's multi-node vagrant setup is [here](/manifests/etcd/etcd-bootkube-vagrant-multi.yaml).

> Hint: this is merely an example for a local setup. The addresses will have to
> be adapted for a setup, that is not a single etcd bootkube created cluster.

Before you continue, you should have endpoints objects for:

* `apiserver` (called `kubernetes` here)
* `kube-controller-manager`
* `kube-scheduler`
* `etcd` (called `etcd-k8s` to make clear this is the etcd used by kubernetes)

For example:

```bash
$ kubectl get endpoints --all-namespaces
NAMESPACE     NAME                                           ENDPOINTS          AGE
default       kubernetes                                     172.17.4.101:443   2h
kube-system   kube-controller-manager-prometheus-discovery   10.2.30.2:10252    1h
kube-system   kube-scheduler-prometheus-discovery            10.2.30.4:10251    1h
monitoring    etcd-k8s                                       172.17.4.51:2379   1h
```

## Monitoring Kubernetes

The manifests used here use the [Prometheus Operator](https://github.com/coreos/prometheus-operator),
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

> All services in the manifest still contain the `prometheus.io/scrape = true`
> annotations. It is not used by the Prometheus controller. They remain for
> pre Prometheus v1.3.0 deployments as in [this example configuration](https://github.com/prometheus/prometheus/blob/6703404cb431f57ca4c5097bc2762438d3c1968e/documentation/examples/prometheus-kubernetes.yml).

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
* KPM/Helm packages to easily provide production-ready cluster-monitoring setup (essentially contents of `hack/cluster-monitoring`)
* Add meta-monitoring to default cluster monitoring setup

