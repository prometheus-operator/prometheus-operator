#!/bin/bash

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-k8s-rules
  labels:
    role: prometheus-rulefiles
    prometheus: k8s
data:
EOF

for f in assets/prometheus/rules/*.rules.y*ml
do
  echo "  $(basename $f | sed s/\.yml//): |+"
  cat $f | sed "s/^/    /g"
done
