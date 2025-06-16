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

To enable OpenTelemetry, set the appropriate exporter environment variables:

```yaml
env:
  # Enable tracing with OTLP exporter
  - name: OTEL_TRACES_EXPORTER
    value: "otlp"
  - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
    value: "http://jaeger:14268/api/traces"
  
  # Enable metrics with OTLP exporter
  - name: OTEL_METRICS_EXPORTER
    value: "otlp"
  - name: OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
    value: "http://otel-collector:4318/v1/metrics"
  
  # Service identification
  - name: OTEL_SERVICE_NAME
    value: "prometheus-operator"
  - name: OTEL_SERVICE_VERSION
    value: "v0.83.0"
```

### Supported Exporters

#### Tracing Exporters

Set `OTEL_TRACES_EXPORTER` to one of:
- `otlp` - OTLP over HTTP or gRPC
- `console` - Console output (for debugging)
- `none` - Disable tracing (default)

#### Metrics Exporters

Set `OTEL_METRICS_EXPORTER` to one of:
- `otlp` - OTLP over HTTP or gRPC
- `prometheus` - Prometheus exposition format
- `console` - Console output (for debugging)  
- `none` - Disable metrics (default)

### OTLP Configuration

When using OTLP exporters, configure the endpoints and authentication:

```yaml
env:
  # Tracing
  - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
    value: "https://otlp.example.com:4318/v1/traces"
  - name: OTEL_EXPORTER_OTLP_TRACES_HEADERS
    value: "authorization=Bearer token123"
  
  # Metrics  
  - name: OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
    value: "https://otlp.example.com:4318/v1/metrics"
  - name: OTEL_EXPORTER_OTLP_METRICS_HEADERS
    value: "authorization=Bearer token123"
  
  # Or configure both at once
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "https://otlp.example.com:4318"
  - name: OTEL_EXPORTER_OTLP_HEADERS
    value: "authorization=Bearer token123"
```

### Resource Attributes

Add additional resource attributes to identify your deployment:

```yaml
env:
  - name: OTEL_RESOURCE_ATTRIBUTES
    value: "deployment.environment=production,k8s.cluster.name=my-cluster,k8s.namespace.name=monitoring"
```

## Example Deployment Configurations

### With Jaeger

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
        - name: OTEL_TRACES_EXPORTER
          value: "otlp"
        - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
          value: "http://jaeger-collector:14268/api/traces"
        - name: OTEL_SERVICE_NAME
          value: "prometheus-operator"
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: "deployment.environment=production"
```

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
        - name: OTEL_TRACES_EXPORTER
          value: "otlp"
        - name: OTEL_METRICS_EXPORTER
          value: "otlp"
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector:4318"
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

OpenTelemetry is disabled by default. To explicitly disable it:

```yaml
env:
- name: OTEL_SDK_DISABLED
  value: "true"
```

Or ensure no exporter environment variables are set (exporters default to "none").

## Environment Variables Reference

### Core Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_SDK_DISABLED` | Disable the OpenTelemetry SDK | `false` |
| `OTEL_SERVICE_NAME` | Service name | Component name |
| `OTEL_SERVICE_VERSION` | Service version | Build version |
| `OTEL_RESOURCE_ATTRIBUTES` | Additional resource attributes | - |

### Tracing Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_TRACES_EXPORTER` | Trace exporter (`otlp`, `console`, `none`) | `none` |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | OTLP traces endpoint | - |
| `OTEL_EXPORTER_OTLP_TRACES_HEADERS` | OTLP traces headers | - |

### Metrics Variables  

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_METRICS_EXPORTER` | Metrics exporter (`otlp`, `prometheus`, `console`, `none`) | `none` |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | OTLP metrics endpoint | - |
| `OTEL_EXPORTER_OTLP_METRICS_HEADERS` | OTLP metrics headers | - |

### General OTLP Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint (for both traces and metrics) | - |
| `OTEL_EXPORTER_OTLP_HEADERS` | OTLP headers (for both traces and metrics) | - |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | OTLP protocol (`grpc`, `http/protobuf`) | `http/protobuf` |

For a complete list of environment variables, see the [OpenTelemetry specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).
