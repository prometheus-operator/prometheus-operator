ARG ARCH=amd64
ARG OS=linux
ARG GOLANG_BUILDER=1.23
ARG GOARCH=
FROM quay.io/prometheus/golang-builder:${GOLANG_BUILDER}-base AS builder
WORKDIR /workspace

# Copy source files
COPY . .

# Build
ENV GOOS=${OS}
ENV GOARCH=${GOARCH}
RUN make prometheus-config-reloader

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=builder workspace/prometheus-config-reloader /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
