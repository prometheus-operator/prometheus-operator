FROM quay.io/prometheus/busybox:latest

ADD po-lint /bin/po-lint

USER nobody

ENTRYPOINT ["/bin/po-lint"]
