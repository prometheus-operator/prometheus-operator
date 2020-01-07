FROM quay.io/prometheus/busybox:latest

ADD lint /bin/lint

USER nobody

ENTRYPOINT ["/bin/lint"]
