ARG ARCH=amd64
ARG OS=linux
ARG GOLANG_BUILDER=1.23
FROM quay.io/prometheus/golang-builder:${GOLANG_BUILDER}-base AS builder
WORKDIR /workspace

# Copy source files
COPY . .

# Build
ARG GOARCH
ENV GOARCH=${GOARCH}
RUN make operator

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=builder workspace/operator /bin/operator

# On busybox 'nobody' has uid `65534'
USER 65534

LABEL org.opencontainers.image.source="https://github.com/prometheus-operator/prometheus-operator" \
    org.opencontainers.image.url="https://prometheus-operator.dev/" \
    org.opencontainers.image.documentation="https://prometheus-operator.dev/" \
    org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/bin/operator"]
