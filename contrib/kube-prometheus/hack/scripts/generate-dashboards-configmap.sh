#!/bin/bash
set -e
set +x

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards-0
data:
EOF

for f in assets/grafana/generated/*-dashboard.json
do
    rm -rf $f
done

virtualenv -p python3 .env 2>&1 > /dev/null
source .env/bin/activate 2>&1 > /dev/null
pip install -Ur requirements.txt 2>&1 > /dev/null
for f in assets/grafana/*.dashboard.py
do
  basefilename=$(basename $f)
  JSON_FILENAME="assets/grafana/generated/${basefilename%%.*}-dashboard.json"
  generate-dashboard $f -o $JSON_FILENAME 2>&1 > /dev/null
done

cp assets/grafana/raw-json-dashboards/*-dashboard.json assets/grafana/generated/

for f in assets/grafana/generated/*-dashboard.json
do
  basefilename=$(basename $f)
  echo "  $basefilename: |+"
  if [ "$basefilename" = "etcd-dashboard.json" ]; then
    hack/scripts/wrap-dashboard.sh $f prometheus-etcd | sed "s/^/    /g"
  else
    hack/scripts/wrap-dashboard.sh $f prometheus | sed "s/^/    /g"
  fi
done

for f in assets/grafana/*-datasource.json
do
  cp $f assets/grafana/generated/
  echo "  $(basename $f): |+"
  cat $f | sed "s/^/    /g"
done
