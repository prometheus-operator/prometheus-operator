#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o xtrace

HELM_BUCKET_NAME="coreos-charts"
HELM_CHARTS_PACKAGED_DIR=${1:-"/tmp/helm-packaged"}
AWS_REGION=${2:-"us-west-2"}

aws configure set region ${AWS_REGION}
aws s3 sync --acl public-read ${HELM_CHARTS_PACKAGED_DIR} s3://${HELM_BUCKET_NAME}/stable/
