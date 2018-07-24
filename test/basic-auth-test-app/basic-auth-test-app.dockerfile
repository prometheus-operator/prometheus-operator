FROM alpine:3.7

ARG VERSION="$VERSION"
ENV VERSION="$VERSION"

COPY basic-auth-test-app /

ENTRYPOINT ["/basic-auth-test-app"]
