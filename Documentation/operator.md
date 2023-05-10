---
weight: 212
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
  -alertmanager-instance-selector string
    	Label selector to filter AlertManager Custom Resources to watch.
  -apiserver string
    	API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.
  -ca-file string
    	- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.
  -cert-file string
    	 - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.
  -cluster-domain string
    	The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.
  -config-reloader-cpu-limit --config-reloader-cpu
    	Config Reloader CPU limit. Value "0" disables it and causes no limit to be configured. Flag overrides --config-reloader-cpu for the CPU limit (default "100m")
  -config-reloader-cpu-request --config-reloader-cpu
    	Config Reloader CPU request. Value "0" disables it and causes no request to be configured. Flag overrides --config-reloader-cpu value for the CPU request (default "100m")
  -config-reloader-memory-limit --config-reloader-memory
    	Config Reloader Memory limit. Value "0" disables it and causes no limit to be configured. Flag overrides --config-reloader-memory for the memory limit (default "50Mi")
  -config-reloader-memory-request --config-reloader-memory
    	Config Reloader Memory request. Value "0" disables it and causes no request to be configured. Flag overrides --config-reloader-memory for the memory request (default "50Mi")
  -deny-namespaces value
    	Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.
  -enable-config-reloader-probes
    	Enable liveness and readiness for the config-reloader container. Default: false
  -key-file string
    	- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.
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
    	Prometheus config reloader image (default "quay.io/prometheus-operator/prometheus-config-reloader:v0.65.1")
  -prometheus-default-base-image string
    	Prometheus default base image (path without tag/version) (default "quay.io/prometheus/prometheus")
  -prometheus-instance-namespaces value
    	Namespaces where Prometheus and PrometheusAgent custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.
  -prometheus-instance-selector string
    	Label selector to filter Prometheus and PrometheusAgent Custom Resources to watch.
  -secret-field-selector string
    	Field selector to filter Secrets to watch
  -short-version
    	Print just the version number.
  -thanos-default-base-image string
    	Thanos default base image (path without tag/version) (default "quay.io/thanos/thanos")
  -thanos-ruler-instance-namespaces value
    	Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.
  -thanos-ruler-instance-selector string
    	Label selector to filter ThanosRuler Custom Resources to watch.
  -tls-insecure
    	- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.
  -version
    	Prints current version.
  -web.cert-file string
    	Cert file to be used for operator web server endpoints. (default "/etc/tls/private/tls.crt")
  -web.client-ca-file string
    	Client CA certificate file to be used for operator web server endpoints. (default "/etc/tls/private/tls-ca.crt")
  -web.enable-tls
    	Activate prometheus operator web server TLS.   This is useful for example when using the rule validation webhook.
  -web.key-file string
    	Private key matching the cert file to be used for operator web server endpoints. (default "/etc/tls/private/tls.key")
  -web.listen-address string
    	Address on which to expose metrics and web interface. (default ":8080")
  -web.tls-cipher-suites string
    	Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).If omitted, the default Go cipher suites will be used.Note that TLS 1.3 ciphersuites are not configurable.
  -web.tls-min-version string
    	Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants. (default "VersionTLS13")
  -web.tls-reload-interval duration
    	The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s). (default 1m0s)
```
