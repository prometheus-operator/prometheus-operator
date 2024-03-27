ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

# Create a basic /etc/os-release file
RUN echo "NAME=\"BusyBox\"\nVERSION=\"1.0\"\nID=busybox\nID_LIKE=linux\nPRETTY_NAME=\"BusyBox\"\nVERSION_ID=\"1.0\"" > /etc/os-release

COPY operator /bin/operator

# On busybox 'nobody' has uid `65534'
USER 65534

LABEL org.opencontainers.image.source="https://github.com/prometheus-operator/prometheus-operator" \
    org.opencontainers.image.url="https://prometheus-operator.dev/" \
    org.opencontainers.image.documentation="https://prometheus-operator.dev/" \
    org.opencontainers.image.licenses="Apache-2.0"
    

ENTRYPOINT ["/bin/operator"]
