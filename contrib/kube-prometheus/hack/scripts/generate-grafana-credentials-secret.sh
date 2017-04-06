#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 user password"
    exit 1
fi

user=$1
password=$2

cat <<-EOF
apiVersion: v1
kind: Secret
metadata:
  name: grafana-credentials
data:
  user: $(echo -n ${user} | base64 --wrap=0)
  password: $(echo -n ${password} | base64 --wrap=0)
EOF

