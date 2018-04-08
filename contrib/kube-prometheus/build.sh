#!/usr/bin/env bash
set -e
set -x

prefix="tmp/manifests"
json="tmp/manifests.json"

rm -rf ${prefix}
mkdir -p $(dirname "${json}")
jsonnet -J /home/brancz/.jsonnet-bundler/src/git/git@github.com-ksonnet-ksonnet-lib/master jsonnet/kube-prometheus.jsonnet > ${json}

files=$(jq -r 'keys[]' ${json})

for file in ${files}; do
    dir=$(dirname "${file}")
    path="${prefix}/${dir}"
    mkdir -p ${path}
    jq -r ".[\"${file}\"]" ${json} | yaml2json | json2yaml > "${prefix}/${file}"
done
