#!/usr/bin/env bash


HELM_BUCKET_NAME="coreos-charts"
HELM_CHARTS_PACKAGED_DIR=${1:-"/tmp/helm-packaged"}
AWS_REGION=${2:-"us-west-2"}

aws configure set region ${AWS_REGION}

#Check if the current chart has the same hash from the remote one
for tgz in $(ls ${HELM_CHARTS_PACKAGED_DIR})
do
  cur_hash=$(cat ${HELM_CHARTS_PACKAGED_DIR}/${tgz}|md5sum)
  remote_hash=$(aws s3api head-object --bucket ${HELM_BUCKET_NAME} --key stable/${tgz} | jq '.ETag' -r| tr -d '"')
  if [ "$cur_hash" != "$remote_hash" ]
  then
    echo "WARNING: Current hash should be the same as the remote hash. Please bump the version of chart {$tgz}."
   fi
done

# sync charts
aws s3 sync --acl public-read ${HELM_CHARTS_PACKAGED_DIR} s3://${HELM_BUCKET_NAME}/stable/
