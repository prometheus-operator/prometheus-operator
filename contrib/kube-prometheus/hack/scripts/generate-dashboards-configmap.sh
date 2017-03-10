#!/bin/bash

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
data:
EOF

for f in assets/grafana/*
do
  echo "  $(basename $f): |+"
  cat $f | sed "s/^/    /g"
done
