#!/bin/bash
set -e
set +x

# Generate Alert Rules ConfigMap
hack/scripts/generate-rules-configmap.sh > manifests/prometheus/prometheus-k8s-rules.yaml

# Generate Dashboard ConfigMap
hack/scripts/generate-dashboards-configmap.sh > manifests/grafana/grafana-dashboard-definitions.yaml

# Generate Dashboard ConfigMap with configmap-generator tool
# Max Size per ConfigMap: 240000
# Input dir: assets/grafana
# output file: manifests/grafana/grafana-dashboards.yaml
# grafana deployment output file: manifests/grafana/grafana-deployment.yaml
test -f manifests/grafana/grafana-dashboard-definitions.yaml && rm -f manifests/grafana/grafana-dashboard-definitions.yaml
test -f manifests/grafana/grafana-deployment.yaml && rm -f manifests/grafana/grafana-deployment.yaml
test -f manifests/grafana/grafana-dashboards.yaml && rm -f manifests/grafana/grafana-dashboards.yaml
hack/grafana-dashboards-configmap-generator/bin/grafana_dashboards_generate.sh -s 240000 -i assets/grafana/generated -o manifests/grafana/grafana-dashboard-definitions.yaml -g manifests/grafana/grafana-deployment.yaml -d manifests/grafana/grafana-dashboards.yaml

# Generate Grafana Credentials Secret
hack/scripts/generate-grafana-credentials-secret.sh admin admin > manifests/grafana/grafana-credentials.yaml

# Generate Secret for Alertmanager config
hack/scripts/generate-alertmanager-config-secret.sh > manifests/alertmanager/alertmanager-config.yaml

