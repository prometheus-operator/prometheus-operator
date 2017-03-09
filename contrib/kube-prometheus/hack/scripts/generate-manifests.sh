#!/bin/bash

# Generate Alert Rules ConfigMap
kubectl create configmap --dry-run=true prometheus-k8s-rules --from-file=assets/prometheus/rules/ -oyaml > manifests/prometheus/prometheus-k8s-rules.yaml

# Generate Dashboard ConfigMap
kubectl create configmap --dry-run=true grafana-dashboards --from-file=assets/grafana/ -oyaml > manifests/grafana/grafana-dashboards.yaml

# Generate Secret for Alertmanager config
kubectl create secret generic alertmanager-main --dry-run --from-file=assets/alertmanager/alertmanager.yaml -oyaml > manifests/alertmanager/alertmanager-config.yaml

