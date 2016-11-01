build:
	./scripts/check_license.sh
	go build github.com/coreos/prometheus-operator/cmd/operator

container:
	docker build -t quay.io/coreos/prometheus-operator.
