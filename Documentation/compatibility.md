<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# Compatibility

The Prometheus Operator supports a number of Kubernetes and Prometheus releases.

## Kubernetes

The Prometheus Operator uses client-go to communicate with Kubernetes clusters. The supported Kubernetes cluster version is determined by client-go. The compatibility matrix for client-go and Kubernetes clusters can be found [here](https://github.com/kubernetes/client-go#compatibility-matrix). All additional compatibility is only best effort, or happens to still/already be supported. The currently used client-go version is "v4.0.0-beta.0".

Due to the use of CustomResourceDefinitions Kubernetes >= v1.7.0 is required.

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
* v1.7.2
* v1.8.0
* v2.0.0
* v2.2.1
* v2.3.1
* v2.3.2
* v2.4.0
* v2.4.1
* v2.4.2
* v2.4.3
* v2.5.0
* v2.6.0
* v2.6.1
* v2.7.0
* v2.7.1
