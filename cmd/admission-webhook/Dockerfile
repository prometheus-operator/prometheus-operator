ARG ARCH=amd64
ARG OS=linux
ARG GOLANG_BUILDER=1.23
FROM quay.io/prometheus/golang-builder:${GOLANG_BUILDER}-base as builder
WORKDIR /workspace

# Copy source files
COPY . .

# Build
RUN make admission-webhook

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

COPY --from=builder workspace/admission-webhook /bin/admission-webhook

USER nobody

ENTRYPOINT ["/bin/admission-webhook"]
