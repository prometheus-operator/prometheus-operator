.PHONY: image
	
IMAGE := coreos/generate-prometheus-operator-manifests

image: Dockerfile
	docker build -t $(IMAGE) .

BUILDER := docker run --rm -it --workdir /data -v ${PWD}:/data $(IMAGE) ./hack/scripts/generate-manifests.sh
generate: image
	@echo ">> Compiling assets and generating Kubernetes manifests"
	$(BUILDER)
