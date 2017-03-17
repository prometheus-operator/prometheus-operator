FROM quay.io/prometheus/busybox:latest

ADD prometheus-config-reloader /bin/prometheus-config-reloader

ENTRYPOINT ["/bin/prometheus-config-reloader"]
