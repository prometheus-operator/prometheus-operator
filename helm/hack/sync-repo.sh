#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

HELM_BUCKET_NAME="coreos-charts"
HELM_CHARTS_PACKAGED_DIR=${1:-"/tmp/helm-packaged"}
aws s3 sync --acl public-read ${HELM_CHARTS_PACKAGED_DIR} s3://${HELM_BUCKET_NAME}/stable/