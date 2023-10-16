ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
FROM ubuntu:20.04
RUN \
    apt-get update && apt-get install -y --no-install-recommends libvpx 1.13.1
    
COPY prometheus-config-reloader /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
