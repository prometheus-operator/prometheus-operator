#!/bin/bash

# Generate Alert Rules ConfigMap
kubectl create configmap --dry-run=true prometheus-k8s-alerts --from-file=assets/alerts/ -oyaml > manifests/prometheus/prometheus-k8s-alerts.yaml

# Generate Dashboard ConfigMap
kubectl create configmap --dry-run=true grafana-dashboards --from-file=assets/grafana/ -oyaml > manifests/grafana/grafana-cm.yaml
