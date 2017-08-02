# kube-prometheus

This repository collects Kubernetes manifests, dashboards, and alerting rules
combined with documentation and scripts to provide single-command deployments
of end-to-end Kubernetes cluster monitoring.

## Prerequisites

First, you need a running Kubernetes cluster. If you don't have one, follow the
instructions of [bootkube](https://github.com/kubernetes-incubator/bootkube) or
[minikube](https://github.com/kubernetes/minikube). Some sample contents of this
repository are adapted to work with a [multi-node setup](https://github.com/kubernetes-incubator/bootkube/tree/master/hack/multi-node)
using [bootkube](https://github.com/kubernetes-incubator/bootkube).

## Monitoring Kubernetes

The manifests used here use the [Prometheus Operator](https://github.com/coreos/prometheus-operator),
which manages Prometheus servers and their configuration in a cluster. With a single command we can install

* The Operator itself
* The Prometheus [node_exporter](https://github.com/prometheus/node_exporter)
* [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)
* The [Prometheus specification](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#prometheus) based on which the Operator deploys a Prometheus setup
* A Prometheus configuration covering monitoring of all Kubernetes core components and exporters
* A default set of alerting rules on the cluster component's health
* A Grafana instance serving dashboards on cluster metrics
* A three node highly available Alertmanager cluster

Simply run:

```bash
export KUBECONFIG=<path>          # defaults to "~/.kube/config"
hack/cluster-monitoring/deploy
```

After all pods are ready, you can reach:

* Prometheus UI on node port `30900`
* Alertmanager UI on node port `30903`
* Grafana on node port `30902`

To tear it all down again, run:

```bash
hack/cluster-monitoring/teardown
```

## Monitoring custom services

The example manifests in [/manifests/examples/example-app](/contrib/kube-prometheus/manifests/examples/example-app)
deploy a fake service exposing Prometheus metrics. They additionally define a new Prometheus
server and a [`ServiceMonitor`](https://github.com/coreos/prometheus-operator/blob/master/Documentation/service-monitor.md),
which specifies how the example service should be monitored.
The Prometheus Operator will deploy and configure the desired Prometheus instance and continiously
manage its life cycle.

```bash
hack/example-service-monitoring/deploy
```

After all pods are ready you can reach the Prometheus server on node port `30100` and observe
how it monitors the service as specified. Same as before, this Prometheus server automatically
discovers the Alertmanager cluster deployed in the [Monitoring Kubernetes](#Monitoring-Kubernetes)
section.

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

For information about how to update/handle the dashboards check [Developing alerts and dashboards](docs/developing-alerts-and-dashboards.md) doc.

In the future, a separate Grafana operator will support gathering dashboards from multiple
ConfigMaps based on label selection.

WARNING: If you deploy multiple Grafana instances for HA, you must use session affinity.
Otherwise if pods restart the prometheus datasource ID can get out of sync between the pods, breaking the UI

## Roadmap

* Grafana Operator that dynamically discovers and deploys dashboards from ConfigMaps
* KPM/Helm packages to easily provide production-ready cluster-monitoring setup (essentially contents of `hack/cluster-monitoring`)
* Add meta-monitoring to default cluster monitoring setup
* Build out the provided dashboards and alerts for cluster monitoring to have full coverage of all system aspects

## Monitoring other Cluster Components

Discovery of API servers and kubelets works the same across all clusters.
Depending on a cluster's setup several other core components, such as etcd or the
scheduler, may be deployed in different ways.
The easiest integration point is for the cluster operator to provide headless services
of all those components to provide a common interface of discovering them. With that
setup they will automatically be discovered by the provided Prometheus configuration.

For the `kube-scheduler` and `kube-controller-manager` there are headless
services prepared, simply add them to your running cluster:

```bash
kubectl -n kube-system create -f manifests/k8s/
```

> Hint: if you use this for a cluster not created with bootkube, make sure you
> populate an endpoints object with the address to your `kube-scheduler` and
> `kube-controller-manager`, or adapt the label selectors to match your setup.

Aside from Kubernetes specific components, etcd is an important part of a
working cluster, but is typically deployed outside of it. This monitoring
setup assumes that it is made visible from within the cluster through a headless
service as well.

> Note that minikube hides some components like etcd so to see the extend of
> this setup we recommend setting up a [local cluster using bootkube](https://github.com/kubernetes-incubator/bootkube/tree/master/hack/multi-node).

An example for bootkube's multi-node vagrant setup is [here](/contrib/kube-prometheus/manifests/etcd/etcd-bootkube-vagrant-multi.yaml).

> Hint: this is merely an example for a local setup. The addresses will have to
> be adapted for a setup, that is not a single etcd bootkube created cluster.

With that setup the headless services provide endpoint lists consumed by
Prometheus to discover the endpoints as targets:

```bash
$ kubectl get endpoints --all-namespaces
NAMESPACE     NAME                                           ENDPOINTS          AGE
default       kubernetes                                     172.17.4.101:443   2h
kube-system   kube-controller-manager-prometheus-discovery   10.2.30.2:10252    1h
kube-system   kube-scheduler-prometheus-discovery            10.2.30.4:10251    1h
monitoring    etcd-k8s                                       172.17.4.51:2379   1h
```

## Other Documentation
[Install Docs for a cluster created with KOPS on AWS](docs/KOPSonAWS.md)
