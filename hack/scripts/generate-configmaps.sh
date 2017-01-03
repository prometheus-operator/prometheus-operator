#!/bin/bash

# Generate Prometheus configuration ConfigMap
kubectl create configmap --dry-run=true prometheus-k8s --from-file=assets/prometheus/prometheus.yaml -oyaml > manifests/prometheus/prometheus-k8s-cm.yaml

# Generate Alert Rules ConfigMap
kubectl create configmap --dry-run=true prometheus-k8s-rules --from-file=assets/prometheus/rules/ -oyaml > manifests/prometheus/prometheus-k8s-rules.yaml

# Generate Dashboard ConfigMap
kubectl create configmap --dry-run=true grafana-dashboards --from-file=assets/grafana/ -oyaml > manifests/grafana/grafana-cm.yaml

