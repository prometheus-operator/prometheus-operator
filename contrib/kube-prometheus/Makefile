BUILDER := docker run --rm -it --workdir /data -v ${PWD}:/data debian:8 ./hack/scripts/generate-manifests.sh
generate:
	@echo ">> Compiling assets and generating Kubernetes manifests"
	$(BUILDER)
