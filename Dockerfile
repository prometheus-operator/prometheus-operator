FROM quay.io/prometheus/busybox:latest

ADD controller /bin/controller

ENTRYPOINT ["/bin/controller"]