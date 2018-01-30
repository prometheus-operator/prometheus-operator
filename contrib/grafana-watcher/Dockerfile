FROM golang:1.9.3-stretch as builder
WORKDIR /go/src/github.com/coreos/prometheus-operator/contrib/grafana-watcher
COPY . /go/src/github.com/coreos/prometheus-operator/contrib/grafana-watcher
RUN go get -d -v github.com/coreos/prometheus-operator/contrib/grafana-watcher
RUN make

FROM quay.io/prometheus/busybox:latest
COPY --from=builder /go/src/github.com/coreos/prometheus-operator/contrib/grafana-watcher/grafana-watcher /bin/grafana-watcher
ENTRYPOINT ["/bin/grafana-watcher"]
