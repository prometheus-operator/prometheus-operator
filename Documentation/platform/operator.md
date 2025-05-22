---
weight: 211
toc: false
title: CLI reference
menu:
    docs:
        parent: operator
lead: Command line arguments for the operator binary
images: []
draft: false
description: Command line arguments for the operator binary
---

> Note this document is automatically generated from the `cmd/operator/main.go` file and shouldn't be edited directly.

```console mdox-exec="./operator --help"
Usage of ./operator:
  -alertmanager-config-namespaces value
    	Namespaces where AlertmanagerConfig custom resources and corresponding Secrets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for AlertmanagerConfig custom resources.
  -alertmanager-default-base-image string
    	Alertmanager default base image (path without tag/version) (default "quay.io/prometheus/alertmanager")
  -alertmanager-instance-namespaces value
    	Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources.
  -alertmanager-instance-selector value
    	Label selector to filter Alertmanager Custom Resources to watch.
  -annotations value
    	Annotations to be add to all resources created by the operator
  -apiserver string
    	API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.
  -as string
    	Username to impersonate. User could be a regular user or a service account in a namespace.
  -auto-gomemlimit-ratio float
    	The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. The value should be greater than 0.0 and less than 1.0. Default: 0.0 (disabled).
  -ca-file string
    	- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.
  -cert-file string
    	 - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.
  -cluster-domain string
    	The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.
  -config-reloader-cpu-limit value
    	Config Reloader CPU limits. Value "0" disables it and causes no limit to be configured. (default 10m)
  -config-reloader-cpu-request value
    	Config Reloader CPU requests. Value "0" disables it and causes no request to be configured. (default 10m)
  -config-reloader-memory-limit value
    	Config Reloader memory limits. Value "0" disables it and causes no limit to be configured. (default 50Mi)
  -config-reloader-memory-request value
    	Config Reloader memory requests. Value "0" disables it and causes no request to be configured. (default 50Mi)
  -controller-id operator.prometheus.io/controller-id
    	Value used by the operator to filter Alertmanager, Prometheus, PrometheusAgent and ThanosRuler objects that it should reconcile. If the value isn't empty, the operator only reconciles objects with an operator.prometheus.io/controller-id annotation of the same value. Otherwise the operator reconciles all objects without the annotation or with an empty annotation value.
  -deny-namespaces value
    	Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.
  -disable-unmanaged-prometheus-configuration
    	Disable support for unmanaged Prometheus configuration when all resource selectors are nil. As stated in the API documentation, unmanaged Prometheus configuration is a deprecated feature which can be avoided with '.spec.additionalScrapeConfigs' or the ScrapeConfig CRD. Default: false.
  -enable-config-reloader-probes
    	Enable liveness and readiness for the config-reloader container. Default: false
  -feature-gates value
    	Feature gates are a set of key=value pairs that describe Prometheus-Operator features.
    	Available feature gates:
    	  PrometheusAgentDaemonSet: Enables the DaemonSet mode for PrometheusAgent (enabled: false)
    	  PrometheusShardRetentionPolicy: Enables shard retention policy for Prometheus (enabled: false)
    	  PrometheusTopologySharding: Enables the zone aware sharding for Prometheus (enabled: false)
    	  StatusForConfigurationResources: Updates the status subresource for configuration resources (enabled: false)
  -key-file string
    	- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.
  -kubelet-endpoints
    	Create Endpoints objects for kubelet targets. (default true)
  -kubelet-endpointslice
    	Create EndpointSlice objects for kubelet targets.
  -kubelet-node-address-priority value
    	Node address priority used by kubelet. Either 'internal' or 'external'. Default: 'internal'.
  -kubelet-selector value
    	Label selector to filter nodes.
  -kubelet-service string
    	Service/Endpoints object to write kubelets into in format "namespace/name"
  -labels value
    	Labels to be add to all resources created by the operator
  -localhost string
    	EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly. (default "localhost")
  -log-format string
    	Log format to use. Possible values: logfmt, json (default "logfmt")
  -log-level string
    	Log level to use. Possible values: all, debug, info, warn, error, none (default "info")
  -namespaces value
    	Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces.
  -prometheus-config-reloader string
    	Prometheus config reloader image (default "quay.io/prometheus-operator/prometheus-config-reloader:v0.82.2")
  -prometheus-default-base-image string
    	Prometheus default base image (path without tag/version) (default "quay.io/prometheus/prometheus")
  -prometheus-instance-namespaces value
    	Namespaces where Prometheus and PrometheusAgent custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.
  -prometheus-instance-selector value
    	Label selector to filter Prometheus and PrometheusAgent Custom Resources to watch.
  -secret-field-selector value
    	Field selector to filter Secrets to watch
  -secret-label-selector value
    	Label selector to filter Secrets to watch
  -short-version
    	Print just the version number.
  -thanos-default-base-image string
    	Thanos default base image (path without tag/version) (default "quay.io/thanos/thanos")
  -thanos-ruler-instance-namespaces value
    	Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.
  -thanos-ruler-instance-selector value
    	Label selector to filter ThanosRuler Custom Resources to watch.
  -tls-insecure
    	- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.
  -version
    	Prints current version.
  -web.cert-file string
    	Certficate file to be used for the web server. (default "/etc/tls/private/tls.crt")
  -web.client-ca-file string
    	Client CA certificate file to be used for the web server. (default "/etc/tls/private/tls-ca.crt")
  -web.enable-http2
    	Enable HTTP2 connections.
  -web.enable-tls
    	Enable TLS for the web server.
  -web.key-file string
    	Private key matching the cert file to be used for the web server. (default "/etc/tls/private/tls.key")
  -web.listen-address string
    	Address on which to expose metrics and web interface. (default ":8080")
  -web.tls-cipher-suites value
    	Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).If omitted, the default Go cipher suites will be used. Note that TLS 1.3 ciphersuites are not configurable.
  -web.tls-min-version string
    	Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants. (default "VersionTLS13")
  -web.tls-reload-interval duration
    	The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s). (default 1m0s)
```
