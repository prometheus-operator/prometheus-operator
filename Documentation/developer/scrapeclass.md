---
weight: 254
toc: true
title: Scrape Classes
menu:
    docs:
        parent: developer
lead: ""
images: []
draft: false
description: null
---

## Prerequisites

Before you begin, ensure that you have:

* `Prometheus-Operator > v0.73.0`
* A running `Prometheus` or `PrometheusAgent` instance.

## Introduction

`ScrapeClass` is a feature that allows you to define common configuration settings to be applied across all scrape resources( **PodMonitor**, **ServiceMonitor**, **ScrapeConfig** and **Probe** ). This feature is similar to [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) in Kubernetes and is very useful when it comes to standardising configurations such as common relabelling rules, TLS certificates and authentication.

One use-case is to configure authentication with TLS certificates when running `Prometheus` to scrape all the pods in an Istio mesh with [strict mTLS](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings). Defining the TLS certificate paths in each `PodMonitor` and `ServiceMonitor` that scrapes these pods would be repetitive and error-prone. This problem is now solved by the `ScrapeClass` feature.

## Defining a ScrapeClass in Prometheus Resource

Below is an example of defining a scrape class in `Prometheus/PrometheusAgent` resource:

```yaml mdox-exec="cat example/user-guides/scrapeclass/scrapeclass-example-definition.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  scrapeClasses:
    - name: istio-mtls
      default: true
      tlsConfig:
        caFile: "/etc/istio-certs/root-cert.pem"
        certFile: "/etc/istio-certs/cert-chain.pem"
        keyFile: "/etc/istio-certs/key.pem"
        insecureSkipVerify: true

  # mount the certs from the istio sidecar (shown here for illustration purposes)
  volumeMounts:
    - name: istio-certs
      mountPath: "/etc/istio-certs/"
  volumes:
  - emptyDir:
      medium: Memory
    name: istio-certs
```

An administrator can set the `default:true` so that the scrape applies to all scrape objects that don't configure an explicit scrape class. Only one scrape class can be set as default. If there are multiple default scrape classes, the operator will fail the reconciliation and the failure will be reported in the status conditions of the resource.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  scrapeClasses:
    - name: first-class
      default: true
    - name: second-class
      default: true
status:
  conditions:
  - message: "Failed to parse scrape classes: multiple default scrape classes defined"
    reason: "MultipleDefaultScrapeClasses"
    status: "False"
    type: Reconciled
```

## Using the ScrapeClass in Monitor Resources

Once the `ScrapeClasses` is defined in the `Prometheus` resource, the `ScrapeClass` field can be used in the scrape resource to reference the particular `ScrapeClass`.

```yaml mdox-exec="cat example/user-guides/scrapeclass/scrapeclass-example-servicemonitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: servicemonitor-example
spec:
  scrapeClass: istio-mtls
  endpoints:
    - port: http
      path: /metrics
```

If the monitor resource specifies a scrape class name that isn't defined in the `Prometheus/PrometheusAgent` object, then the operator will emit a Kubernetes event. Consider the following example where a `Prometheus` instance defines a scrapeClass, but a `ServiceMonitor` references a different one.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: my-prometheus
spec:
  replicas: 1
  scrapeClasses:
    - name: istio-mtls
      default: true
```

The `ServiceMonitor` references a scrapeClass named **istio**, which is not defined in the Prometheus object.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-service-monitor
spec:
  selector:
    matchLabels:
      app: example
  scrapeClass: istio
  endpoints:
    - port: http
      path: /metrics
```

As `ServiceMonitor` references a scrapeClass that does not exist in Prometheus, the following event is emitted:

```
0s Warning InvalidConfiguration servicemonitor/example-service-monitor 
ServiceMonitor example-service-monitor was rejected due to invalid configuration: 
scrapeClass "istio" not found in Prometheus scrapeClasses
```

Similarly, we can select the scrape class for `PodMonitor` resource.

```yaml mdox-exec="cat example/user-guides/scrapeclass/scrapeclass-example-podmonitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
spec:
  scrapeClass: istio-mtls
  podMetricsEndpoints:
  - port: http
    path: /metrics
```

If a monitor resource includes a `tlsConfig` field, the Operator applies a **field-by-field** merge with the `tlsConfig` from the scrape class. Any fields set in the monitor resource take precedence, while unset fields inherit values from the scrape class.

> Note: The configuration in scrapeClass will only be applied if the scrape resources haven't set fields defined in scrapeClass.

## What's Next

{{<
link-card title="Scrape Class" href="https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.ScrapeClass" description="Check out the specifications to learn more about scrape classes">}}
