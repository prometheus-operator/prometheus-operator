#!/bin/bash

# Generate Alert Rules ConfigMap
kubectl create configmap --dry-run=true prometheus-k8s-rules --from-file=assets/alerts/ -oyaml > manifests/prometheus/prometheus-k8s-rules.yaml

# Generate Dashboard ConfigMap
kubectl create configmap --dry-run=true grafana-dashboards --from-file=assets/grafana/ -oyaml > manifests/grafana/grafana-cm.yaml
