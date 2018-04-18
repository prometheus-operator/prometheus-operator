#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

function defer {
	docker logout quay.io
}
trap defer EXIT

if [[ "${TRAVIS_PULL_REQUEST}" != "false" ]]; then
	exit 0
fi

export REPO=quay.io/coreos/prometheus-operator
# Push to Quay '-dev' repo if not a git tag or master branch build
if [[ "${TRAVIS_TAG}" == "" ]] && [[ "${TRAVIS_BRANCH}" != master ]]; then
	export REPO="${REPO}-dev"
fi

make crossbuild

# For both git tags and git branches 'TRAVIS_BRANCH' contains the name.
export TAG="${TRAVIS_BRANCH}"
make container
echo "${QUAY_PASSWORD}" | docker login -u "${QUAY_USERNAME}" --password-stdin quay.io
docker push "${REPO}:${TRAVIS_BRANCH}"
