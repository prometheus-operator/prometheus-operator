FROM golang:1.12 AS builder
WORKDIR /go/src/github.com/prometheus-operator/prometheus-operator
COPY . .
RUN cd test/instrumented-sample-app && make build

FROM alpine:3.7
COPY --from=builder /go/src/github.com/prometheus-operator/prometheus-operator/test/instrumented-sample-app/instrumented-sample-app /usr/bin/instrumented-sample-app
ENTRYPOINT ["/usr/bin/instrumented-sample-app"]
