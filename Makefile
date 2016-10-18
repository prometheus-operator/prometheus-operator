build:
	go build github.com/coreos/kube-prometheus-controller/cmd/controller

container:
	docker build -t quay.io/coreos/kube-prometheus-controller .
