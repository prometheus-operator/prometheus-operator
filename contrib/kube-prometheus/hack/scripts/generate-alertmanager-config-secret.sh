#!/bin/bash

cat <<-EOF
apiVersion: v1
kind: Secret
metadata:
  name: alertmanager-main
data:
  alertmanager.yaml: $(cat assets/alertmanager/alertmanager.yaml | base64 --wrap=0)
EOF

