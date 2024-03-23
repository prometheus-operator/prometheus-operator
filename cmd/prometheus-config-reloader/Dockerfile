ARG ARCH="amd64"
ARG OS="linux"

FROM golang:1.21 as build

WORKDIR /go/src/github.com/prometheus-operator/prometheus-operator

RUN apt update
RUN apt install make -y

COPY Makefile go.mod go.sum VERSION .header /go/src/github.com/prometheus-operator/prometheus-operator/
COPY cmd /go/src/github.com/prometheus-operator/prometheus-operator/cmd
COPY internal /go/src/github.com/prometheus-operator/prometheus-operator/internal
COPY pkg /go/src/github.com/prometheus-operator/prometheus-operator/pkg
COPY scripts /go/src/github.com/prometheus-operator/prometheus-operator/scripts

RUN make prometheus-config-reloader

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=build /go/src/github.com/prometheus-operator/prometheus-operator/prometheus-config-reloader /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
