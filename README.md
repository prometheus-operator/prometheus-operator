# kube-prometheus

This repository collects Kubernetes manifests, dashboards, and alerting rules
combined with documentationa and scripts to deploy them to get full cluster 
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

```
export KUBECONFIG=<path>   # defaults to "~/.kube/config"
export KUBE_NAMESPACE=<ns> # defaults to "default"
hack/cluster-monitoring/deploy
```

To tear it all down again, run:

```
hack/cluster-monitoring/teardown
```

After all pods are ready, you can reach:

* Prometheus UI on node port `30900`
* Grafana on node port `30902`

## Monitoring custom services

TODO
