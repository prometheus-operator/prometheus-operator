VERSION=5.0.3
IMAGE_TAG=$(VERSION)

container:
	docker build --build-arg GRAFANA_VERSION=$(VERSION) -t quay.io/coreos/monitoring-grafana:$(IMAGE_TAG) .
