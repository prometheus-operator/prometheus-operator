#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

CPU_ARCHS="amd64 arm64 arm"

export TAG="${GITHUB_REF##*/}"

# Push `-dev` images unless commit is tagged
export REPO="${REPO:-"quay.io/prometheus-operator/prometheus-operator-dev"}"
export REPO_PROMETHEUS_CONFIG_RELOADER="${REPO_PROMETHEUS_CONFIG_RELOADER:-"quay.io/prometheus-operator/prometheus-config-reloader-dev"}"

# Use main image repositories if TAG is a semver tag or it is a master branch
if [[ "$TAG" =~ ^v[0-9]+\.[0-9]+ ]] || [ "${TAG}" == "master" ]; then
	export REPO="quay.io/prometheus-operator/prometheus-operator"
	export REPO_PROMETHEUS_CONFIG_RELOADER="quay.io/prometheus-operator/prometheus-config-reloader"
fi

for arch in ${CPU_ARCHS}; do
	make --always-make image GOARCH="$arch" TAG="${TAG}-$arch"
done

export DOCKER_CLI_EXPERIMENTAL=enabled
for r in ${REPO} ${REPO_PROMETHEUS_CONFIG_RELOADER}; do
	# Images need to be on remote registry before creating manifests
	for arch in $CPU_ARCHS; do
		docker push "${r}:${TAG}-$arch"
	done

	# Create manifest to join all images under one virtual tag
	docker manifest create -a "${r}:${TAG}" \
				  "${r}:${TAG}-amd64" \
				  "${r}:${TAG}-arm64" \
				  "${r}:${TAG}-arm"

	# Annotate to set which image is build for which CPU architecture
	for arch in $CPU_ARCHS; do
		docker manifest annotate --arch "$arch" "${r}:${TAG}" "${r}:${TAG}-$arch"
	done
	docker manifest push "${r}:${TAG}"
done
