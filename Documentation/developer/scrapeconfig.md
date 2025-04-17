---
weight: 253
toc: true
title: ScrapeConfig CRD
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Guide to use ScrapeConfig to scrape targets external to the Kubernetes cluster
---

Starting with prometheus-operator v0.65.x, one can use the `ScrapeConfig` CRD to scrape targets external to the
Kubernetes cluster or create scrape configurations that are not possible with the higher level
`ServiceMonitor`/`Probe`/`PodMonitor` resources.

# Prerequisites
* `prometheus-operator` `>v0.65.1`
* `ScrapeConfig` CRD installed in the cluster. Make sure to (re)start the operator after the CRD has been created/updated.

# Configure Prometheus or PrometheusAgent to select ScrapeConfigs

Both the Prometheus and PrometheusAgent CRD have a `scrapeConfigSelector` field. This field needs to be set to a list of
labels to match `ScrapeConfigs`:

```yaml
spec:
  scrapeConfigSelector:
    matchLabels:
      prometheus: system-monitoring-prometheus
```

With this example, all `ScrapeConfig` having the `prometheus` label set to `system-monitoring-prometheus` will be used
to generate scrape configurations.

# Use ScrapeConfig to scrape an external target

`ScrapeConfig` currently supports a limited set of service discoveries:
* `static_config`
* `file_sd`
* `http_sd`
* `kubernetes_sd`
* `consul_sd`

The following examples are basic and don't cover all the supported service discovery mechanisms. The CRD is constantly evolving, adding new features and support for new Service Discoveries. Check the [API documentation](https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1alpha1.ScrapeConfig) to see all supported fields.

If you have an interest in another service discovery mechanism or you see something missing in the implementation, please
[open an issue](https://github.com/prometheus-operator/prometheus-operator/issues).

## `static_config`

For example, to scrape the target located at `http://prometheus.demo.do.prometheus.io:9090`, use the following:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeConfig
metadata:
  name: static-config
  namespace: my-namespace
  labels:
    prometheus: system-monitoring-prometheus
spec:
  staticConfigs:
    - labels:
        job: prometheus
      targets:
        - prometheus.demo.do.prometheus.io:9090
```

## `file_sd`

To use `file_sd`, a file has to be mounted in the Prometheus or PrometheusAgent pods. The following configmap is a service discovery file:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scrape-file-sd-targets
  namespace: monitoring
  labels:
    prometheus: system-monitoring-prometheus
data:
  targets.yaml: |
    - labels:
        job: node-demo
      targets:
      - node.demo.do.prometheus.io:9100
    - labels:
        job: prometheus
      targets:
      - prometheus.demo.do.prometheus.io:9090
```

This `ConfigMap` will then need to be mounted in the `Prometheus` spec:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: your-prometheus
  namespace: my-namespace
  labels:
    prometheus: system-monitoring-prometheus
spec:
  scrapeConfigSelector:
    matchLabels:
      prometheus: system-monitoring-prometheus
  configMaps:
    - scrape-file-sd-targets
```

You can then use ScrapeConfig to reference that file and scrape the associated targets:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeConfig
metadata:
  name: file-sd
  namespace: my-namespace
  labels:
    prometheus: system-monitoring-prometheus
    app.kubernetes.io/name: scrape-config-example
spec:
  fileSDConfigs:
    - files:
        - /etc/prometheus/configmaps/scrape-file-sd-targets/targets.yaml
```

## `http_sd`

`http_sd` uses an endpoint for data, unlike `file_sd` which uses a file, removing the need for a configmap. For instance:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeConfig
metadata:
  name: http-sd
  namespace: my-namespace
  labels:
    prometheus: system-monitoring-prometheus
    app.kubernetes.io/name: scrape-config-example
spec:
  httpSDConfigs:
    - url: http://my-external-api/discovery
      refreshInterval: 15s
```
