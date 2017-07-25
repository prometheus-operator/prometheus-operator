all:
	CGO_ENABLED=0 go build --installsuffix cgo main.go
	docker build -t quay.io/coreos/prometheus-alertmanager-test-webhook .
