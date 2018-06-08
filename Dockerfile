FROM quay.io/prometheus/busybox:latest

ADD operator /bin/operator

ENTRYPOINT ["/bin/operator"]
