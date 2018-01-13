#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o xtrace

HELM_CHARTS_PACKAGED_DIR=${1:-"/tmp/helm-packaged"}
HELM_BUCKET_NAME=${2:-"coreos-charts"}
HELM_REPO_URL=${3:-"https://s3-eu-west-1.amazonaws.com/${HELM_BUCKET_NAME}/stable/"}

#Check if the current chart has the same hash from the remote one
for tgz in $(ls ${HELM_CHARTS_PACKAGED_DIR})
do
  # if remote file doesn't exist we can skip the comparison 
  file=$(wget -O out_file ${HELM_REPO_URL}${tgz})||continue
  remote_hash=$(cat out_file | md5sum )
  cur_hash=$(cat ${HELM_CHARTS_PACKAGED_DIR}/${tgz}| md5sum)
  if [ "${tgz}" != "index.yaml" ]  && [ "$cur_hash" != "$remote_hash" ]
  then
    echo "ERROR: Current hash should be the same as the remote hash. Please bump the version of chart {$tgz}."
    exit 1
   fi
done


