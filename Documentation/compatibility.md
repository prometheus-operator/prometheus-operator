# Compatibility

The Prometheus Operator supports a number of Kubernetes and Prometheus releases.

## Kubernetes

The Prometheus Operator uses client-go to communicate with Kubernetes clusters. The supported Kubernetes cluster version is determined by client-go. The compatibility matrix for client-go and Kubernetes clusters can be found [here](https://github.com/kubernetes/client-go#compatibility-matrix). All additional compatibility is only best effort, or happens to still/already be supported. The currently used client-go version is "v4.0.0-beta.0".

## Prometheus

The versions of Prometheus compatible to be run with the Prometheus Operator are:

* v1.4.0
* v1.4.1
* v1.5.0
* v1.5.1
* v1.5.2
* v1.5.3
* v1.6.0
* v1.6.1
* v1.6.2
* v1.6.3
* v1.7.0
* v1.7.1
* v2.0.0-beta.0
