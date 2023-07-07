---
weight: 202
toc: true
title: Compatibility
menu:
    docs:
        parent: operator
lead: The Prometheus Operator supports a number of Kubernetes and Prometheus releases.
images: []
draft: false
description: The Prometheus Operator supports a number of Kubernetes and Prometheus releases.
---

It is recommended to use versions of the components identical or close to the versions used by the operator's end-to-end test suite (the specific version numbers are listed below).

## Kubernetes

Due to the use of apiextensions.k8s.io/v1 CustomResourceDefinitions, prometheus-operator requires Kubernetes >= v1.16.0.

The Prometheus Operator uses the official [Go client](https://github.com/kubernetes/client-go) for Kubernetes to communicate with the Kubernetes API. The compatibility matrix for client-go and Kubernetes clusters can be found [here](https://github.com/kubernetes/client-go#compatibility-matrix). All additional compatibility is only best effort, or happens to be still/already supported.

The current version of the Prometheus operator uses the following Go client version:

```$ mdox-exec="go list -m  -f '{{ .Version }}' k8s.io/client-go"
v0.27.3
```

## Prometheus

Prometheus Operator supports all Prometheus versions >= v2.0.0. The operator's end-to-end tests verify that the operator can deploy the following Prometheus versions:

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility"
* v2.37.0
* v2.37.1
* v2.37.2
* v2.37.3
* v2.37.4
* v2.37.5
* v2.37.6
* v2.37.7
* v2.37.8
* v2.38.0
* v2.39.0
* v2.39.1
* v2.39.2
* v2.40.0
* v2.40.1
* v2.40.2
* v2.40.3
* v2.40.4
* v2.40.5
* v2.40.6
* v2.40.7
* v2.41.0
* v2.42.0
* v2.43.0
* v2.43.1
* v2.44.0
```

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultPrometheusVersion"
* v2.44.0
```

## Alertmanager

The Prometheus Operator is compatible with Alertmanager v0.15 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultAlertmanagerVersion"
* v0.25.0
```

## Thanos

The Prometheus Operator is compatible with Thanos v0.10 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultThanosVersion"
* v0.31.0
```
