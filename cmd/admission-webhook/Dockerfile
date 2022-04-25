ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

ADD admission-webhook /bin/admission-webhook

USER nobody

ENTRYPOINT ["/bin/admission-webhook"]
