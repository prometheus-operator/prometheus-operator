#!/bin/bash

# Generate Alert Rules ConfigMap
hack/scripts/generate-rules-configmap.sh > manifests/prometheus/prometheus-k8s-rules.yaml

# Generate Dashboard ConfigMap
hack/scripts/generate-dashboards-configmap.sh > manifests/grafana/grafana-dashboards.yaml

# Generate Grafana Credentials Secret
hack/scripts/generate-grafana-credentials-secret.sh admin admin > manifests/grafana/grafana-credentials.yaml

# Generate Secret for Alertmanager config
hack/scripts/generate-alertmanager-config-secret.sh > manifests/alertmanager/alertmanager-config.yaml

