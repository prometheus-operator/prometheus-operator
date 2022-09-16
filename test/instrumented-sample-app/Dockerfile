FROM scratch

ARG VERSION
ARG ARCH="amd64"

ENV VERSION=${VERSION}
COPY linux/${ARCH}/instrumented-sample-app /bin/instrumented-sample-app

ENTRYPOINT ["/bin/instrumented-sample-app"]
