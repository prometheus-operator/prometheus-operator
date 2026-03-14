---
weight: 103
toc: true
title: Compatibility
menu:
    docs:
        parent: prologue
lead: The Prometheus Operator supports a number of Kubernetes and Prometheus releases.
images: []
draft: false
description: The Prometheus Operator supports a number of Kubernetes and Prometheus releases.
---

It is recommended to use versions of the components identical or close to the versions used by the operator's end-to-end test suite (the specific version numbers are listed below).

## Kubernetes

Due to the use of [Common Expression Language](https://kubernetes.io/docs/reference/using-api/cel/) (CEL) in Custom Resource Definitions, prometheus-operator >= v0.84.0 requires Kubernetes >= v1.25.0 (or v1.23.0 with the `CustomResourceValidationExpressions` feature gate enabled).

Releases before v0.84.0 require Kubernetes >= v1.16.0.

The Prometheus Operator uses the official [Go client](https://github.com/kubernetes/client-go) for Kubernetes to communicate with the Kubernetes API. The compatibility matrix for client-go and Kubernetes clusters can be found [here](https://github.com/kubernetes/client-go#compatibility-matrix). All additional compatibility is only best effort, or happens to be still/already supported.

The current version of the Prometheus operator uses the following Go client version:

```$ mdox-exec="go list -m  -f '{{ .Version }}' k8s.io/client-go"
v0.34.3
```

## Prometheus

Prometheus Operator supports all Prometheus versions >= v2.0.0. The operator's end-to-end tests verify that the operator can deploy the following Prometheus versions:

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility"
* v2.45.0
* v2.46.0
* v2.47.0
* v2.47.1
* v2.47.2
* v2.48.0
* v2.48.1
* v2.49.0
* v2.49.1
* v2.50.0
* v2.50.1
* v2.51.0
* v2.51.1
* v2.51.2
* v2.52.0
* v2.53.0
* v2.53.1
* v2.53.2
* v2.53.3
* v2.54.0
* v2.54.1
* v2.55.0
* v2.55.1
* v3.0.0
* v3.0.1
* v3.1.0
* v3.2.0
* v3.2.1
* v3.3.0
* v3.3.1
* v3.4.0
* v3.4.1
* v3.4.2
* v3.5.0
* v3.6.0
* v3.7.0
* v3.7.1
* v3.7.2
* v3.7.3
* v3.8.0
* v3.8.1
* v3.9.0
* v3.9.1
* v3.10.0
```

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultPrometheusVersion"
* v3.10.0
```

## Alertmanager

The Prometheus Operator is compatible with Alertmanager v0.15 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultAlertmanagerVersion"
* v0.31.1
```

## Thanos

The Prometheus Operator is compatible with Thanos v0.10 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultThanosVersion"
* v0.41.0
```
