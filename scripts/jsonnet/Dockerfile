FROM golang:1.12-stretch

ENV JSONNET_VERSION 0.13.0
ENV PROMTOOL_VERSION 2.12.0
ENV GOLANGCILINT_VERSION 1.17.1

RUN apt-get update -y && apt-get install -y g++ make git jq gawk python-yaml && \
    rm -rf /var/lib/apt/lists/*
RUN curl -Lso - https://github.com/google/jsonnet/archive/v${JSONNET_VERSION}.tar.gz | \
    tar xfz - -C /tmp && \
    cd /tmp/jsonnet-${JSONNET_VERSION} && \
    make && mv jsonnetfmt /usr/local/bin && \
    rm -rf /tmp/jsonnet-${JSONNET_VERSION}

RUN GO111MODULE=on go get github.com/google/go-jsonnet/cmd/jsonnet@v${JSONNET_VERSION}
RUN GO111MODULE=on go get github.com/prometheus/prometheus/cmd/promtool@v${PROMTOOL_VERSION}
RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v${GOLANGCILINT_VERSION}
RUN go get github.com/brancz/gojsontoyaml
RUN go get github.com/campoy/embedmd
RUN go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
RUN go get -u github.com/jteeuwen/go-bindata/...

FROM golang:1.12-stretch

COPY --from=0 /usr/local/bin/jsonnetfmt /bin
COPY --from=0 /go/bin/jsonnet /bin
COPY --from=0 /go/bin/promtool /bin
COPY --from=0 /go/bin/golangci-lint /bin
COPY --from=0 /go/bin/gojsontoyaml /bin
COPY --from=0 /go/bin/embedmd /bin
COPY --from=0 /go/bin/jb /bin

RUN apt-get update -y && apt-get install -y make git jq gawk && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/github.com/coreos/prometheus-operator /.cache
WORKDIR /go/src/github.com/coreos/prometheus-operator

RUN chmod -R 777 /go /.cache
