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

RUN make build

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=build /go/src/github.com/prometheus-operator/prometheus-operator/operator /bin/operator

# On busybox 'nobody' has uid `65534'
USER 65534

LABEL org.opencontainers.image.source="https://github.com/prometheus-operator/prometheus-operator" \
    org.opencontainers.image.url="https://prometheus-operator.dev/" \
    org.opencontainers.image.documentation="https://prometheus-operator.dev/" \
    org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/bin/operator"]
