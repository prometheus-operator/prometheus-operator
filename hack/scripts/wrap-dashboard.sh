#!/bin/bash -eu

# Intended usage:
#  * Edit dashboard in Grafana (you need to login first with admin/admin
#    login/password).
#  * Save dashboard in Grafana to check is specification is correct.
#    Looks like this is the only way to check is dashboard specification
#    has error.
#  * Download dashboard specification as JSON file in Grafana:
#    Share -> Export -> Save to file.
#  * Wrap dashboard specification to make it digestable by kube-prometheus:
#      ./hack/scripts/wrap-dashboard.sh Nodes-1488465802729.json
#  * Replace dashboard specification:
#      mv Nodes-1488465802729.json assets/grafana/node-dashboard.json
#  * Regenerate Grafana configmap:
#      ./hack/scripts/generate-configmaps.sh
#  * Apply new configmap:
#      kubectl -n monitoring apply -f manifests/grafana/grafana-cm.yaml

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 path-to-dashboard.json"
    exit 1
fi

json=$1
temp=$(tempfile -m 0644)

cat >> $temp <<EOF
{
  "dashboard":
EOF

cat $json >> $temp

cat >> $temp <<EOF
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

mv $temp $json

