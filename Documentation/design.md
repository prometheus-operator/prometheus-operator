<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.
</div>

# Design

This document describes the design and interaction between the custom resource definitions that the Prometheus Operator introduces.

The custom resources that the Prometheus Operator introduces are:

* `Prometheus`
* `ServiceMonitor`
* `PodMonitor` 
* `Alertmanager`
* `PrometheusRule`

## Prometheus

The `Prometheus` custom resource definition (CRD) declaratively defines a desired Prometheus setup to run in a Kubernetes cluster. It provides options to configure replication, persistent storage, and Alertmanagers to which the deployed Prometheus instances send alerts to.

For each `Prometheus` resource, the Operator deploys a properly configured `StatefulSet` in the same namespace. The Prometheus `Pod`s are configured to mount a `Secret` called `<prometheus-name>` containing the configuration for Prometheus.

The CRD specifies which `ServiceMonitor`s should be covered by the deployed Prometheus instances based on label selection. The Operator then generates a configuration based on the included `ServiceMonitor`s and updates it in the `Secret` containing the configuration. It continuously does so for all changes that are made to `ServiceMonitor`s or the `Prometheus` resource itself.

If no selection of `ServiceMonitor`s is provided, the Operator leaves management of the `Secret` to the user, which allows to provide custom configurations while still benefiting from the Operator's capabilities of managing Prometheus setups.

## ServiceMonitor

The `ServiceMonitor` custom resource definition (CRD) allows to declaratively define how a dynamic set of services should be monitored. Which services are selected to be monitored with the desired configuration is defined using label selections. This allows an organization to introduce conventions around how metrics are exposed, and then following these conventions new services are automatically discovered, without the need to reconfigure the system.

For Prometheus to monitor any application within Kubernetes an `Endpoints` object needs to exist. `Endpoints` objects are essentially lists of IP addresses. Typically an `Endpoints` object is populated by a `Service` object. A `Service` object discovers `Pod`s by a label selector and adds those to the `Endpoints` object.

A `Service` may expose one or more service ports, which are backed by a list of multiple endpoints that point to a `Pod` in the common case. This is reflected in the respective `Endpoints` object as well.

The `ServiceMonitor` object introduced by the Prometheus Operator in turn discovers those `Endpoints` objects and configures Prometheus to monitor those `Pod`s.

The `endpoints` section of the `ServiceMonitorSpec`, is used to configure which ports of these `Endpoints` are going to be scraped for metrics, and with which parameters. For advanced use cases one may want to monitor ports of backing `Pod`s, which are not directly part of the service endpoints. Therefore when specifying an endpoint in the `endpoints` section, they are strictly used.

> Note: `endpoints` (lowercase) is the field in the `ServiceMonitor` CRD, while `Endpoints` (capitalized) is the Kubernetes object kind.

Both `ServiceMonitors` as well as discovered targets may come from any namespace. This is important to allow cross-namespace monitoring use cases, e.g. for meta-monitoring. Using the `ServiceMonitorNamespaceSelector` of the `PrometheusSpec`, one can restrict the namespaces `ServiceMonitor`s are selected from by the respective Prometheus server. Using the `namespaceSelector` of the `ServiceMonitorSpec`, one can restrict the namespaces the `Endpoints` objects are allowed to be discovered from.
To discover targets in all namespaces the `namespaceSelector` has to be empty:
```yaml
spec:
  namespaceSelector:
    any: true
```

## PodMonitor

The `PodMonitor` custom resource definition (CRD) allows to declaratively define how a dynamic set of pods should be monitored.
Which pods are selected to be monitored with the desired configuration is defined using label selections.
This allows an organization to introduce conventions around how metrics are exposed, and then following these conventions new pods are automatically discovered, without the need to reconfigure the system.

A `Pod` is a collection of one or more containers which can expose Prometheus metrics on a number of ports.

The `PodMonitor` object introduced by the Prometheus Operator discovers these pods and generates the relevant configuration for the Prometheus server in order to monitor them. 

The `PodMetricsEndpoints` section of the `PodMonitorSpec`, is used to configure which ports of a pod are going to be scraped for metrics, and with which parameters.

Both `PodMonitors` as well as discovered targets may come from any namespace. This is important to allow cross-namespace monitoring use cases, e.g. for meta-monitoring.
Using the `namespaceSelector` of the `PodMonitorSpec`, one can restrict the namespaces the `Pods` are allowed to be discovered from.
To discover targets in all namespaces the `namespaceSelector` has to be empty:
```yaml
spec:
  namespaceSelector:
    any: true
```

## Alertmanager

The `Alertmanager` custom resource definition (CRD) declaratively defines a desired Alertmanager setup to run in a Kubernetes cluster. It provides options to configure replication and persistent storage.

For each `Alertmanager` resource, the Operator deploys a properly configured `StatefulSet` in the same namespace. The Alertmanager pods are configured to include a `Secret` called `<alertmanager-name>` which holds the used configuration file in the key `alertmanager.yaml`.

When there are two or more configured replicas the operator runs the Alertmanager instances in high availability mode.

## PrometheusRule

The `PrometheusRule` CRD declaratively defines a desired Prometheus rule to be consumed by one or more Prometheus instances. 

Alerts and recording rules can be saved and applied as YAML files, and dynamically loaded without requiring any restart.
