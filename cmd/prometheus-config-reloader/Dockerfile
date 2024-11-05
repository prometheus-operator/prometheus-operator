ARG ARCH=amd64
ARG OS=linux
ARG GOLANG_BUILDER=1.23
FROM quay.io/prometheus/golang-builder:${GOLANG_BUILDER}-base AS builder
WORKDIR /workspace

COPY . .

# Download Go dependencies to reuse the Go cache in subsequent builds.
RUN go mod download && go mod verify

# Build
RUN make prometheus-config-reloader

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=builder workspace/prometheus-config-reloader /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
