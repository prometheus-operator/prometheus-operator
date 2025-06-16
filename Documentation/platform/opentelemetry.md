---
weight: 210
toc: true
title: OpenTelemetry Integration
menu:
    docs:
        parent: platform
lead: ""
images: []
draft: false
description: Configure OpenTelemetry tracing and metrics for Prometheus Operator
---

The Prometheus Operator supports OpenTelemetry for observability through tracing and metrics. This integration uses the OpenTelemetry autoexport functionality, which allows configuration through standard environment variables.

## Overview

OpenTelemetry integration is available in all Prometheus Operator components:
- `prometheus-operator` (main operator)
- `admission-webhook` 
- `prometheus-config-reloader`

The integration uses [autoexport](https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport) which automatically configures exporters based on environment variables, making it easy to integrate with various observability backends.

## Configuration

OpenTelemetry is configured entirely through environment variables following the [OpenTelemetry specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).

### Basic Configuration

To enable OpenTelemetry, set the appropriate exporter environment variables as suggested in the [OpenTelemetry documentation](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) and in the [OTLP Exporter Configuration documentation](https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/).  

Generally, you only need to set:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
```


We use the [autoexport](https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport) package to automatically configure exporters based on environment variables.


### Resource Attributes

Add additional resource attributes to identify your deployment:

```yaml
env:
  - name: OTEL_RESOURCE_ATTRIBUTES
    value: "deployment.environment=production,k8s.cluster.name=my-cluster,k8s.namespace.name=monitoring"
```

## Example Deployment Configurations

### With OpenTelemetry Collector

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-operator
spec:
  template:
    spec:
      containers:
      - name: prometheus-operator
        image: quay.io/prometheus-operator/prometheus-operator:latest
        env:
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector:4318"
        - name: OTEL_EXPORTER_OTLP_PROTOCOL
          value: "http/protobuf"  
        - name: OTEL_SERVICE_NAME
          value: "prometheus-operator"
```

### Console Output (Development)

For development and debugging, you can output telemetry to the console:

```yaml
env:
- name: OTEL_TRACES_EXPORTER
  value: "console"
- name: OTEL_METRICS_EXPORTER
  value: "console"
- name: OTEL_SERVICE_NAME
  value: "prometheus-operator"
```

## Disabling OpenTelemetry

To disable OpenTelemetry, you can set the `OTEL_SDK_DISABLED` environment variable to `true` in your deployment configuration:

```yaml
env:
- name: OTEL_SDK_DISABLED
  value: "true"
```
