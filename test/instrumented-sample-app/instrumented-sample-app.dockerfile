FROM alpine:3.7

ARG VERSION="$VERSION"
ENV VERSION="$VERSION"

COPY instrumented-sample-app /

ENTRYPOINT ["/instrumented-sample-app"]
