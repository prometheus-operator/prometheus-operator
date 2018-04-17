#!/usr/bin/env bash
set -e
set -x

jsonnet="${1-kube-prometheus.jsonnet}"
prefix="${2-manifests}"
json="tmp/manifests.json"

rm -rf ${prefix}
mkdir -p $(dirname "${json}")
jsonnet \
    -J $GOPATH/src/github.com/ksonnet/ksonnet-lib \
    -J $GOPATH/src/github.com/grafana/grafonnet-lib \
    -J $GOPATH/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus/jsonnet \
    -J $GOPATH/src/github.com/brancz/kubernetes-grafana/src/kubernetes-jsonnet \
    ${jsonnet} > ${json}

files=$(jq -r 'keys[]' ${json})

for file in ${files}; do
    dir=$(dirname "${file}")
    path="${prefix}/${dir}"
    mkdir -p ${path}
    jq -r ".[\"${file}\"]" ${json} | gojsontoyaml -yamltojson | gojsontoyaml > "${prefix}/${file}"
done
