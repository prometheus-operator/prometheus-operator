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
v0.25.2
```

## Prometheus

The versions of Prometheus compatible with the Prometheus Operator are:

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility"
* v2.0.0
* v2.2.1
* v2.3.1
* v2.3.2
* v2.4.0
* v2.4.1
* v2.4.2
* v2.4.3
* v2.5.0
* v2.6.0
* v2.6.1
* v2.7.0
* v2.7.1
* v2.7.2
* v2.8.1
* v2.9.2
* v2.10.0
* v2.11.0
* v2.14.0
* v2.15.2
* v2.16.0
* v2.17.2
* v2.18.0
* v2.18.1
* v2.18.2
* v2.19.0
* v2.19.1
* v2.19.2
* v2.19.3
* v2.20.0
* v2.20.1
* v2.21.0
* v2.22.0
* v2.22.1
* v2.22.2
* v2.23.0
* v2.24.0
* v2.24.1
* v2.25.0
* v2.25.1
* v2.25.2
* v2.26.0
* v2.26.1
* v2.27.0
* v2.27.1
* v2.28.0
* v2.28.1
* v2.29.0
* v2.29.1
* v2.30.0
* v2.30.1
* v2.30.2
* v2.30.3
* v2.31.0
* v2.31.1
* v2.32.0
* v2.32.1
* v2.33.0
* v2.33.1
* v2.33.2
* v2.33.3
* v2.33.4
* v2.33.5
* v2.34.0
* v2.35.0
* v2.36.0
* v2.37.0
* v2.37.1
* v2.38.0
* v2.39.0
```

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultPrometheusVersion"
* v2.39.0
```

## Alertmanager

The Prometheus Operator is compatible with Alertmanager v0.15 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultAlertmanagerVersion"
* v0.24.0
```

## Thanos

The Prometheus Operator is compatible with Thanos v0.10 and above.

The end-to-end tests are mostly tested against

```$ mdox-exec="go run ./cmd/po-docgen/. compatibility defaultThanosVersion"
* v0.28.0
```
