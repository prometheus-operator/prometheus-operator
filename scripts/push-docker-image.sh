#!/usr/bin/env bash
#
# Script is meant to build multi-arch container images and publish them to multiple container registries
# 
# Script is:
# - figuring out if an image is a development one (by default suffixed with `-dev` in image name)
# - figuring out the image tag (aka version) based on GITHUB_REF value
#
# Script is not:
# - directly executing `docker build`. This is done in Makefile
# - logging to registries 
#

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

CPU_ARCHS="amd64 arm64 arm"
REGISTRIES="${REGISTRIES:-"quay.io ghcr.io"}"

# IMAGE_OPERATOR and IMAGER_RELOADER need to be exported to be used by `make`
export IMAGE_OPERATOR="${IMAGE_OPERATOR:-"prometheus-operator/prometheus-operator"}"
export IMAGE_RELOADER="${IMAGE_RELOADER:-"prometheus-operator/prometheus-config-reloader"}"

# Figure out if current commit is tagged
export TAG="${GITHUB_REF##*/}"

# Push `-dev` images unless commit is tagged
IMAGE_SUFFIX="-dev"

# Use main image repositories if TAG is a semver tag or it is a master branch
# Prepare image tag from VERSION file and short commit SHA in other cases
if [[ "$TAG" =~ ^v[0-9]+\.[0-9]+ ]] || [ "${TAG}" == "master" ]; then
	# Reset suffixes as images are not development ones
	IMAGE_SUFFIX=""
else
	TAG="v$(cat "$(git rev-parse --show-toplevel)/VERSION")-$(git rev-parse --short HEAD)"
fi

# Compose full image names for retagging and publishing to remote container registries
OPERATORS=""
RELOADERS=""
for i in ${REGISTRIES}; do
	OPERATORS="$i/${IMAGE_OPERATOR}${IMAGE_SUFFIX} ${OPERATORS}"
	RELOADERS="$i/${IMAGE_RELOADER}${IMAGE_SUFFIX} ${RELOADERS}"
done

for img in ${OPERATORS} ${RELOADERS}; do
	echo "Building multi-arch image: $img:$TAG"
done

# Build images and rename them for each remote registry
for arch in ${CPU_ARCHS}; do
	make --always-make image GOARCH="$arch" TAG="${TAG}-$arch"
	# Retag operator image
	for i in ${OPERATORS}; do
		docker tag "${IMAGE_OPERATOR}:${TAG}-$arch" "${i}:${TAG}-$arch"
	done
	# Retag reloader image
	for i in ${RELOADERS}; do
		docker tag "${IMAGE_RELOADER}:${TAG}-$arch" "${i}:${TAG}-$arch"
	done
done

# Compose multi-arch images and push them to remote repositories
export DOCKER_CLI_EXPERIMENTAL=enabled
for r in ${OPERATORS} ${RELOADERS}; do
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
