# kube-prometheus

> Note that everything in the `contrib/kube-prometheus/` directory is experimental and may change significantly at any time.

This repository collects Kubernetes manifests, [Grafana](http://grafana.com/) dashboards, and
[Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)
combined with documentation and scripts to provide single-command deployments of end-to-end
Kubernetes cluster monitoring with [Prometheus](https://prometheus.io/) (Operator).

## Prerequisites

First, you need a running Kubernetes cluster. If you don't have one, we recommend you create one
with [Tectonic Installer](https://coreos.com/tectonic/docs/latest/). Despite the name,
Tectonic Installer gives you also the choice to create a barebones Kubernetes cluster, without
CoreOS' Tectonic technology. Otherwise, you can simply make use of
[bootkube](https://github.com/kubernetes-incubator/bootkube) or
[minikube](https://github.com/kubernetes/minikube) for local testing. Some sample contents of this
repository are adapted to work with a [multi-node setup](https://github.com/kubernetes-incubator/bootkube/tree/master/hack/multi-node)
using [bootkube](https://github.com/kubernetes-incubator/bootkube).


> We assume that the kubelet uses token authN and authZ, as otherwise
> Prometheus needs a client certificate, which gives it full access to the
> kubelet, rather than just the metrics. Token authN and authZ allows more fine
> grained and easier access control. Simply start minikube with the following
> command (you can of course adapt the version and memory to your needs):
>
> $ minikube delete && minikube start --kubernetes-version=v1.9.1 --memory=4096 --bootstrapper=kubeadm --extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook --extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0
>
> In future versions of minikube and kubeadm this will be the default, but for
> the time being, we will have to configure it ourselves.

## Monitoring Kubernetes

The manifests here use the [Prometheus Operator](https://github.com/coreos/prometheus-operator),
which manages Prometheus servers and their configuration in a cluster. With a single command we can
install

* The Operator itself
* The Prometheus [node_exporter](https://github.com/prometheus/node_exporter)
* [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)
* The [Prometheus specification](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#prometheus) based on which the Operator deploys a Prometheus setup
* A Prometheus configuration covering monitoring of all Kubernetes core components and exporters
* A default set of alerting rules on the cluster components' health
* A Grafana instance serving dashboards on cluster metrics
* A three node highly available Alertmanager cluster

Simply run:

```bash
cd contrib/kube-prometheus/
hack/cluster-monitoring/deploy
```

After all pods are ready, you can reach each of the UIs by port-forwarding:

* Prometheus UI on node port `kubectl -n monitoring port-forward prometheus-k8s-0 9090`
* Alertmanager UI on node port `kubectl -n monitoring port-forward alertmanager-main-0 9093`
* Grafana on node port `kubectl -n monitoring port-forward $(kubectl get pods -n monitoring -lapp=grafana -ojsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}') 3000`

To tear it all down again, run:

```bash
hack/cluster-monitoring/teardown
```

## Customizing

As everyone's infrastructure is slightly different, different organizations have different requirements. Thereby there may be modifications you want to do on kube-prometheus to fit your needs.

The kube-prometheus stack is intended to be a jsonnet library for organizations to consume and use in their own infrastructure repository. Below is an example how it can be used to deploy the stack properly on minikube.

The three "distribution" examples we have assembled can be found in:

* `example-dist/base`: contains the plain kube-prometheus stack for organizations to build on.
* `example-dist/kubeadm`: contains the kube-prometheus stack with slight modifications to work properly monitoring kubeadm clusters and exposes UIs on NodePorts for demonstration purposes.
* `example-dist/bootkube`: contains the kube-prometheus stack with slight modifications to work properly on clusters created with bootkube.

The examples in `example-dist/` are purely meant for demonstration purposes, the `kube-prometheus.jsonnet` file should live in your organizations infrastructure repository and use the kube-prometheus library provided here.

Examples of additoinal modifications you may want to make could be adding an `Ingress` object for each of the UIs, but the point of this is that as opposed to other solutions out there, this library does not need to yield all possible customization options, it's all up to the user to customize!

### minikube kubeadm example

See `example-dist/kubeadm` for an example for deploying on minikube, using the minikube kubeadm bootstrapper. The `example-dist/kubeadm/kube-prometheus.jsonnet` file renders the kube-prometheus manifests using jsonnet and then merges the result with kubeadm specifics, such as information on how to monitor kube-controller-manager and kube-scheduler as created by kubeadm. In addition for demonstration purposes, it converts the services selecting Prometheus, Alertmanager and Grafana to NodePort services.

Let's give that a try, and create a minikube cluster:

```
minikube delete && minikube start --kubernetes-version=v1.9.6 --memory=4096 --bootstrapper=kubeadm --extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook --extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0
```

Then we can render the manifests for kubeadm (because we are using the minikube kubeadm bootstrapper):

```
docker run --rm \
  -v `pwd`:/go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus \
  --workdir /go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus \
  po-jsonnet \
  ./hack/scripts/build-jsonnet.sh example-dist/kubeadm/kube-prometheus.jsonnet example-dist/kubeadm/manifests
```

> Note the `po-jsonnet` docker image is built using [this Dockerfile](/scripts/jsonnet/Dockerfile), you can also build it using `make image` from the `contrib/kube-prometheus` folder.

Then the stack can be deployed using

```
hack/cluster-monitoring/deploy example-dist/kubeadm
```

## Monitoring custom services

The example manifests in [examples/example-app](/contrib/kube-prometheus/examples/example-app)
deploy a fake service exposing Prometheus metrics. They additionally define a new Prometheus
server and a [`ServiceMonitor`](https://github.com/coreos/prometheus-operator/blob/master/Documentation/design.md#servicemonitor),
which specifies how the example service should be monitored.
The Prometheus Operator will deploy and configure the desired Prometheus instance and continuously
manage its life cycle.

```bash
hack/example-service-monitoring/deploy
```

After all pods are ready you can reach the Prometheus server similar to the Prometheus server above:

```bash
kubectl port-forward prometheus-frontend-0 9090
```

Then you can access Prometheus through `http://localhost:9090/`.

Teardown:

```bash
hack/example-service-monitoring/teardown
```

## Dashboarding

The provided manifests deploy a Grafana instance serving dashboards provided via ConfigMaps.
Said ConfigMaps are generated from Python scripts in assets/grafana, that all have the extension
.dashboard.py as they are loaded by the [grafanalib](https://github.com/aknuds1/grafanalib)
Grafana dashboard generator. Bear in mind that we are for now using a fork of grafanalib as
we needed to make extensive changes to it, in order to be able to generate our dashboards. We are
hoping to be able to consolidate our version with the original.

As such, in order to make changes to the dashboard bundle, you need to change the \*.dashboard.py 
files in assets/grafana, eventually add your own, and then run `make generate` in the
kube-prometheus root directory.
 
To read more in depth about developing dashboards, read the
[Developing Prometheus Rules and Grafana Dashboards](docs/developing-alerts-and-dashboards.md)
documentation.

### Reloading of dashboards

Currently, Grafana does not support serving dashboards from static files. Instead, the `grafana-watcher`
sidecar container aims to emulate the behavior, by keeping the Grafana database always in sync
with the provided ConfigMap. Hence, the Grafana pod is effectively stateless.
This allows managing dashboards via `git` etc. and easily deploying them via CD pipelines.

In the future, a separate Grafana operator will support gathering dashboards from multiple
ConfigMaps based on label selection.

WARNING: If you deploy multiple Grafana instances for HA, you must use session affinity.
Otherwise if pods restart the prometheus datasource ID can get out of sync between the pods,
breaking the UI

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
