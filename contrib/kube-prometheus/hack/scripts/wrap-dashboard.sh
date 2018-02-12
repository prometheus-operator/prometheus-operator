#!/bin/bash -eu

# Intended usage:
#  * Edit dashboard in Grafana (you need to login first with admin/admin
#    login/password).
#  * Save dashboard in Grafana to check is specification is correct.
#    Looks like this is the only way to check if dashboard specification
#    has errors.
#  * Download dashboard specification as JSON file in Grafana:
#    Share -> Export -> Save to file.
#  * Drop dashboard specification in assets folder:
#      mv Nodes-1488465802729.json assets/grafana/node-dashboard.json
#  * Regenerate Grafana configmap:
#      ./hack/scripts/generate-manifests.sh
#  * Apply new configmap:
#      kubectl -n monitoring apply -f manifests/grafana/grafana-cm.yaml

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 path-to-dashboard.json grafana-prometheus-datasource-name"
    exit 1
fi

dashboardjson=$1
datasource_name=$2
inputname="DS_PROMETHEUS"

if [ "$datasource_name" = "prometheus-etcd" ]; then
  inputname="DS_PROMETHEUS-ETCD"
fi

cat <<EOF
{
  "dashboard":
EOF

cat $dashboardjson

cat <<EOF
,
  "inputs": [
    {
      "name": "$inputname",
      "pluginId": "prometheus",
      "type": "datasource",
      "value": "$datasource_name"
    }
  ],
  "overwrite": true
}
EOF

