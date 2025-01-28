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

<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.73.0, Scrape Class feature is added to Prometheus-Operator.
</div>

## Prerequisites

Before you begin, ensure that you have:

* `Prometheus-Operator > v0.73.0`
* A running `Prometheus` instance.

## Introduction

`ScrapeClass` is a feature that allows you to define common configuration settings to be applied to all scrape resources( **PodMonitor**, **ServiceMonitor**, **ScrapeConfig** and **Probe** ) of that particular class. This feature is similar to [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/) in Kubernetes. This feature is very useful when it comes to standardizing configurations such as common relabeling rules, certificates, and other configurations.

Traditionally, the scrape configurations were defined in the `Prometheus` config file itself. This process was time consuming when user wanted to apply the same configuration to many scrape targets. One such problem was when a user needed to configure secure scraping with TLS certificates when running `Prometheus` to scrape all the pods in an Istio mesh with strict `mTLS`, described [here](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings). Defining the TLS certificate paths in each `PodMonitor` and `ServiceMonitor` that scrapes these pods would be repetitive and error-prone. This problem is now solved by `ScrapeClass` feature.

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

An administrator can set the `default:true` so that the scrape applies to all scrape objects that don't configure an explicit scrape class. Only one scrape class can be set as default. If there are multiple default scrape classes, the operator will fail the reconciliation.

## Using the ScrapeClass in Monitor Resources

Once the `ScrapeClass` is defined in the `Prometheus` resource, the `ScrapeClassName` field can be used in monitor resource to reference the particular `ScrapeClass`.

```yaml mdox-exec="cat example/user-guides/scrapeclass/scrapeclass-example-servicemonitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: servicemonitor-example
spec:
  scrapeClassName: istio-mtls
  endpoints:
    - port: http
      path: /metrics
```

If the monitor resource specifies a scrape class name that isn't defined in the `Prometheus/PrometheusAgent` object, then the scrape resource is rejected by the operator.

Similarly, we can select the scrape class for `PodMonitor` resource.

```yaml mdox-exec="cat example/user-guides/scrapeclass/scrapeclass-example-podmonitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
spec:
  scrapeClassName: istio-mtls
  podMetricsEndpoints:
  - port: http
    path: /metrics
```

If a monitor resource includes a `tlsConfig` field, the Operator will apply a merge strategy to combine the `tlsConfig` fields from the monitor resource with those defined in the scrape class. The `tlsConfig` settings in the monitor resource will take precedence.

## What's Next

{{<
link-card title="Scrape Class" href="https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.ScrapeClass" description="Check out the specifications to learn more about scrape classes">}}
