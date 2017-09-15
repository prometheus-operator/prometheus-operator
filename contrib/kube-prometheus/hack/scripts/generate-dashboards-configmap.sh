#!/bin/bash
set -e

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards-0
data:
EOF

virtualenv -p python3 .env
source .env/bin/activate
pip install -Ur requirements.txt
for f in assets/grafana/*.dashboard.py
do
  JSON_FILENAME="$(pwd)/${f%%.*}-dashboard.json"
  generate-dashboard $f -o $JSON_FILENAME 2>&1 > /dev/null
done

for f in assets/grafana/*-dashboard.json
do
  echo "  $(basename $f): |+"
  hack/scripts/wrap-dashboard.sh $f | sed "s/^/    /g"
done

for f in assets/grafana/*-datasource.json
do
  echo "  $(basename $f): |+"
  cat $f | sed "s/^/    /g"
done
