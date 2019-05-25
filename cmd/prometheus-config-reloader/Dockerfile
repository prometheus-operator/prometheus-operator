FROM quay.io/prometheus/busybox:latest

ADD prometheus-config-reloader /bin/prometheus-config-reloader

RUN chown nobody:nogroup /bin/prometheus-config-reloader

USER nobody

ENTRYPOINT ["/bin/prometheus-config-reloader"]
