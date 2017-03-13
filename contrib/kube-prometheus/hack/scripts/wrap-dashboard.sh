#!/bin/bash -eu

# Intended usage:
#  * Edit dashboard in Grafana (you need to login first with admin/admin
#    login/password).
#  * Save dashboard in Grafana to check is specification is correct.
#    Looks like this is the only way to check is dashboard specification
#    has error.
#  * Download dashboard specification as JSON file in Grafana:
#    Share -> Export -> Save to file.
#  * Drop dashboard specification in assets folder:
#      mv Nodes-1488465802729.json assets/grafana/node-dashboard.json
#  * Regenerate Grafana configmap:
#      ./hack/scripts/generate-manifests.sh
#  * Apply new configmap:
#      kubectl -n monitoring apply -f manifests/grafana/grafana-cm.yaml

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 path-to-dashboard.json"
    exit 1
fi

dashboardjson=$1

cat <<EOF
{
  "dashboard":
EOF

cat $dashboardjson

cat <<EOF
,
  "inputs": [
    {
      "name": "DS_PROMETHEUS",
      "pluginId": "prometheus",
      "type": "datasource",
      "value": "prometheus"
    }
  ],
  "overwrite": true
}
EOF

