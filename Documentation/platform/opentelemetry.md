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

The Prometheus Operator includes comprehensive OpenTelemetry support for observability through tracing and metrics. The integration provides automatic instrumentation for HTTP servers, Kubernetes clients, and external HTTP clients using industry-standard OpenTelemetry libraries including [otelhttp](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp).

## Overview

OpenTelemetry integration is available in all Prometheus Operator components:
- `prometheus-operator` (main operator)
- `admission-webhook`
- `prometheus-config-reloader`

The integration uses [autoexport](https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport) which automatically configures exporters based on environment variables, making it easy to integrate with various observability backends.

## What is Instrumented

The Prometheus Operator includes comprehensive OpenTelemetry instrumentation out of the box:

### HTTP Server Instrumentation

All HTTP servers are automatically instrumented with [otelhttp](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp):
- Prometheus Operator main server (metrics, health endpoints)
- Admission webhook HTTP server
- Prometheus config reloader HTTP server

### Kubernetes Client Instrumentation

All Kubernetes API interactions are automatically traced:
- All controller operations (Prometheus, Alertmanager, ThanosRuler)
- API server discovery and configuration checks
- Secret, ConfigMap, and CRD operations
- RBAC verification calls

### HTTP Client Instrumentation

External HTTP clients are instrumented:
- Prometheus config reloader HTTP client (for reloading Prometheus configuration)

This provides comprehensive observability across all operator components and their interactions with Kubernetes and external services.

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
