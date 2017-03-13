#!/bin/bash

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
data:
EOF

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
