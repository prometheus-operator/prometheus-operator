FROM scratch

ARG ARCH="amd64"

COPY linux/${ARCH}/alertmanager-webhook /bin/alertmanager-webhook

ENTRYPOINT ["/bin/alertmanager-webhook"]
