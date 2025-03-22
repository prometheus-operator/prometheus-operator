ARG ARCH=amd64
ARG OS=linux
ARG GOLANG_BUILDER=1.24

FROM quay.io/prometheus/golang-builder:${GOLANG_BUILDER}-base AS builder
WORKDIR /workspace

COPY . .

# Download Go dependencies to reuse the Go cache in subsequent builds.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download -x && go mod verify

# Build
ARG GOARCH
ENV GOARCH=${GOARCH}
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build make prometheus-config-reloader

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=builder workspace/prometheus-config-reloader /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
