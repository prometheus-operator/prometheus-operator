main
---
title: "Operator CLI Flags"
description: "Lists of possible arguments passed to operator executable."
date: 2021-06-18T14:12:33-00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 1000
toc: false
---

Operator CLI Flags
=================
This article lists arguments of operator executable.
> Note this document is automatically generated from the `cmd/operator/main.go` file and shouldn't be edited directly.

| Argument | Description | Default Value |
| -------- | ----------- | ------------- |
| web.listen-address | Address on which to expose metrics and web interface. | :8080 |
| web.enable-tls | Activate prometheus operator web server TLS.   This is useful for example when using the rule validation webhook. | false |
| web.cert-file | Cert file to be used for operator web server endpoints. | /etc/tls/private/tls.crt |
| web.key-file | Private key matching the cert file to be used for operator web server endpoints. | /etc/tls/private/tls.key |
| web.client-ca-file | Client CA certificate file to be used for operator web server endpoints. | /etc/tls/private/tls-ca.crt |
| web.tls-reload-interval | The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s). | Minute |
| web.tls-min-version | Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants. | VersionTLS13 |
| web.tls-cipher-suites | Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).If omitted, the default Go cipher suites will be used.Note that TLS 1.3 ciphersuites are not configurable. | "" |
| apiserver | API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token. | "" |
| cert-file |  - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file. | "" |
| key-file | - NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file. | "" |
| ca-file | - NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file. | "" |
| kubelet-service | Service/Endpoints object to write kubelets into in format \"namespace/name\" | "" |
| tls-insecure | - NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate. | false |
| prometheus-config-reloader | Prometheus config reloader image | "" |
| config-reloader-cpu | Config Reloader CPU request & limit. Value \"0\" disables it and causes no request/limit to be configured. Deprecated, it will be removed in v0.49.0. | 100m |
| config-reloader-cpu-request | Config Reloader CPU request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-cpu` value for the CPU request | "" |
| config-reloader-cpu-limit | Config Reloader CPU limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-cpu` for the CPU limit | "" |
| config-reloader-memory | Config Reloader Memory request & limit. Value \"0\" disables it and causes no request/limit to be configured. Deprecated, it will be removed in v0.49.0. | 50Mi |
| config-reloader-memory-request | Config Reloader Memory request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-memory` for the memory request | "" |
| config-reloader-memory-limit | Config Reloader Memory limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-memory` for the memory limit | "" |
| alertmanager-default-base-image | Alertmanager default base image (path without tag/version) | "" |
| prometheus-default-base-image | Prometheus default base image (path without tag/version) | "" |
| thanos-default-base-image | Thanos default base image (path without tag/version) | "" |
| namespaces | Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces. | N/A |
| deny-namespaces | Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces. | N/A |
| prometheus-instance-namespaces | Namespaces where Prometheus custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources. | N/A |
| alertmanager-instance-namespaces | Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources. | N/A |
| thanos-ruler-instance-namespaces | Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources. | N/A |
| labels | Labels to be add to all resources created by the operator | N/A |
| localhost | EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly. | localhost |
| cluster-domain | The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead. | "" |
| log-level |  | info |
| log-format |  | logfmt |
| prometheus-instance-selector | Label selector to filter Prometheus Custom Resources to watch. | "" |
| alertmanager-instance-selector | Label selector to filter AlertManager Custom Resources to watch. | "" |
| thanos-ruler-instance-selector | Label selector to filter ThanosRuler Custom Resources to watch. | "" |
| secret-field-selector | Field selector to filter Secrets to watch | "" |
