#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o xtrace

HELM_BUCKET_NAME="coreos-charts"
SYNC_TO_S3=${1:-"false"}
HELM_CHARTS_PACKAGED_DIR=${2:-"/tmp/helm-packaged"}

#Check if the current chart has the same hash from the remote one
for tgz in "${HELM_CHARTS_PACKAGED_DIR}"/*
do
  if echo "${tgz}" | grep -vq "kube-prometheus" 
  then  # if remote file doesn't exist we can skip the comparison 
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "https://s3-eu-west-1.amazonaws.com/${HELM_BUCKET_NAME}/stable/${tgz}")
    if [ "$status_code" == "200" ] 
    then
      cur_hash=$(md5sum "${HELM_CHARTS_PACKAGED_DIR}/${tgz}" | awk '{print $1}' )
      remote_hash=$(curl -s "https://s3-eu-west-1.amazonaws.com/${HELM_BUCKET_NAME}/stable/${tgz}" | md5sum | awk '{print $1}')
      if [ "${tgz}" != "index.yaml" ]  && [ "$cur_hash" != "$remote_hash" ]
      then
        echo "ERROR: Current hash should be the same as the remote hash. Please bump the version of chart {$tgz}."
        exit 1
      fi
    fi
  fi
done

# sync charts
if [ "${SYNC_TO_S3}" = true ]
then
  aws s3 sync --acl public-read "${HELM_CHARTS_PACKAGED_DIR}" "s3://${HELM_BUCKET_NAME}/stable/"
fi

exit 0
