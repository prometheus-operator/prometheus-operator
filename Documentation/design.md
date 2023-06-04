---
weight: 201
toc: true
title: Design
menu:
    docs:
        parent: operator
images: []
draft: false
description: This document describes the design and interaction between the custom resource definitions that the Prometheus Operator manages.
---

This document describes the design and interaction between the [custom resource definitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/) that the Prometheus Operator manages.

The custom resources managed by the Prometheus Operator are:

* [Prometheus](#prometheus)
* [Alertmanager](#alertmanager)
* [ThanosRuler](#thanosruler)
* [ServiceMonitor](#servicemonitor)
* [PodMonitor](#podmonitor)
* [Probe](#probe)
* [PrometheusRule](#prometheusrule)
* [AlertmanagerConfig](#alertmanagerconfig)
* [PrometheusAgent](#prometheusagent)

## Prometheus

The `Prometheus` custom resource definition (CRD) declaratively defines a desired [Prometheus](https://prometheus.io/docs/prometheus) setup to run in a Kubernetes cluster. It provides options to configure the number of replicas, persistent storage, and Alertmanagers to which the deployed Prometheus instances send alerts to.

For each `Prometheus` resource, the Operator deploys one or several `StatefulSet` objects in the same namespace (the number of statefulsets is equal to the number of shards but by default it is 1).

The CRD defines via label and namespace selectors which `ServiceMonitor`, `PodMonitor` and `Probe` objects should be associated to the deployed Prometheus instances. The CRD also defines which `PrometheusRules` objects should be reconciled. The operator continuously reconciles the custom resources and generates one or several `Secret` objects holding the Prometheus configuration. A `config-reloader` container running as a sidecar in the Prometheus pod detects any change to the configuration and reloads Prometheus if needed.

## Alertmanager

The `Alertmanager` custom resource definition (CRD) declaratively defines a desired [Alertmanager](https://prometheus.io/docs/alerting) setup to run in a Kubernetes cluster. It provides options to configure the number of replicas and persistent storage.

For each `Alertmanager` resource, the Operator deploys a `StatefulSet` in the same namespace. The Alertmanager pods are configured to mount a `Secret` called `alertmanager-<alertmanager-name>` which holds the Alertmanager configuration under the key `alertmanager.yaml`.

When there are two or more configured replicas, the Operator runs the Alertmanager instances in high-availability mode.

## ThanosRuler

The `ThanosRuler` custom resource definition (CRD) declaratively defines a desired [Thanos Ruler](https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md) setup to run in a Kubernetes cluster. With Thanos Ruler recording and alerting rules can be processed across multiple Prometheus instances.

A `ThanosRuler` instance requires at least one query endpoint which points to the location of Thanos Queriers or Prometheus instances.

Further information can also be found in the [Thanos section]({{< ref "thanos.md" >}}).

## ServiceMonitor

The `ServiceMonitor` custom resource definition (CRD) allows to declaratively define how a dynamic set of services should be monitored. Which services are selected to be monitored with the desired configuration is defined using label selections. This allows an organization to introduce conventions around how metrics are exposed, and then following these conventions new services are automatically discovered, without the need to reconfigure the system.

For Prometheus to monitor any application within Kubernetes an `Endpoints` object needs to exist. `Endpoints` objects are essentially lists of IP addresses. Typically an `Endpoints` object is populated by a `Service` object. A `Service` object discovers `Pod`s by a label selector and adds those to the `Endpoints` object.

A `Service` may expose one or more service ports, which are backed by a list of multiple endpoints that point to a `Pod` in the common case. This is reflected in the respective `Endpoints` object as well.

The `ServiceMonitor` object introduced by the Prometheus Operator in turn discovers those `Endpoints` objects and configures Prometheus to monitor those `Pod`s.

The `endpoints` section of the `ServiceMonitorSpec`, is used to configure which ports of these `Endpoints` are going to be scraped for metrics, and with which parameters. For advanced use cases one may want to monitor ports of backing `Pod`s, which are not directly part of the service endpoints. Therefore when specifying an endpoint in the `endpoints` section, they are strictly used.

> Note: `endpoints` (lowercase) is the field in the `ServiceMonitor` CRD, while `Endpoints` (capitalized) is the Kubernetes object kind.

Both `ServiceMonitors` as well as discovered targets may come from any namespace. This is important to allow cross-namespace monitoring use cases, e.g. for meta-monitoring. Using the `ServiceMonitorNamespaceSelector` of the `PrometheusSpec`, one can restrict the namespaces `ServiceMonitor`s are selected from by the respective Prometheus server. Using the `namespaceSelector` of the `ServiceMonitorSpec`, one can restrict the namespaces the `Endpoints` objects are allowed to be discovered from.

One can discover targets in all namespaces like this:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
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

Once can discover targets in all namespaces like this:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  podMetricsEndpoints:
  - port: web
  namespaceSelector:
    any: true
```

## Probe

The `Probe` custom resource definition (CRD) allows to declarative define how groups of ingresses and static targets should be monitored. Besides the target, the `Probe` object requires a `prober` which is the service that monitors the target and provides metrics for Prometheus to scrape. Typically, this is achieved using the [blackbox exporter](https://github.com/prometheus/blackbox_exporter).

## PrometheusRule

The `PrometheusRule` custom resource definition (CRD) declaratively defines desired Prometheus rules to be consumed by Prometheus or Thanos Ruler instances.

Alerts and recording rules are reconciled by the Operator and dynamically loaded without requiring any restart of Prometheus/Thanos Rulre.

## AlertmanagerConfig

The `AlertmanagerConfig` custom resource definition (CRD) declaratively specifies subsections of the Alertmanager configuration, allowing routing of alerts to custom receivers, and setting inhibition rules. The `AlertmanagerConfig` can be defined on a namespace level providing an aggregated configuration to Alertmanager. An example on how to use it is provided below. Please be aware that this CRD is not stable yet.

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-config-example.yaml"
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: config-example
  labels:
    alertmanagerConfig: example
spec:
  route:
    groupBy: ['job']
    groupWait: 30s
    groupInterval: 5m
    repeatInterval: 12h
    receiver: 'webhook'
  receivers:
  - name: 'webhook'
    webhookConfigs:
    - url: 'http://example.com/'
```

## PrometheusAgent

The `PrometheusAgent` custom resource definition (CRD) declaratively defines a desired [Prometheus Agent](https://prometheus.io/blog/2021/11/16/agent/) setup to run in a Kubernetes cluster.

Similar to the binaries of Prometheus Server and Prometheus Agent, the `Prometheus` and `PrometheusAgent` CRs are also similar. Inspired in the Agent binary, the Agent CR has several configuration options redacted when compared with regular Prometheus CR, e.g. alerting, PrometheusRules selectors, remote-read, storage and thanos sidecars.

A more extensive read explaining why Agent support was done with a whole new CRD can be seen [here](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/designs/prometheus-agent.md).
