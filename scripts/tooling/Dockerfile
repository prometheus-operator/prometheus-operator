FROM golang:1.14 as builder

ENV JSONNET_VERSION v0.15.0
# This corresponds to v2.16.0, but needs to be written as commit hash due to golang 1.13 issues
ENV PROMTOOL_VERSION b90be6f32a33c03163d700e1452b54454ddce0ec
ENV GOLANGCILINT_VERSION v1.23.6
ENV JB_VERSION v0.3.1
ENV GO_BINDATA_VERSION v3.1.3

RUN apt-get update -y && apt-get install -y g++ make git && \
    rm -rf /var/lib/apt/lists/*
RUN curl -Lso - https://github.com/google/jsonnet/archive/${JSONNET_VERSION}.tar.gz | \
    tar xfz - -C /tmp && \
    cd /tmp/jsonnet-${JSONNET_VERSION#v} && \
    make && mv jsonnetfmt /usr/local/bin && \
    rm -rf /tmp/jsonnet-${JSONNET_VERSION#v}

RUN GO111MODULE=on go get github.com/google/go-jsonnet/cmd/jsonnet@${JSONNET_VERSION}
RUN GO111MODULE=on go get github.com/prometheus/prometheus/cmd/promtool@${PROMTOOL_VERSION}
RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCILINT_VERSION}
RUN GO111MODULE=on go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@${JB_VERSION}
RUN go get github.com/brancz/gojsontoyaml
RUN go get github.com/campoy/embedmd
RUN GO111MODULE=on go get github.com/go-bindata/go-bindata/v3/go-bindata@${GO_BINDATA_VERSION}

# Add po-lint
WORKDIR /go/src/github.com/prometheus-operator/prometheus-operator
COPY . .
RUN GO111MODULE=on make po-lint && chmod +x po-lint && mv po-lint /go/bin/

FROM golang:1.14
RUN apt-get update -y && apt-get install -y make git jq gawk python-yaml && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/bin/jsonnetfmt /usr/local/bin/jsonnetfmt
COPY --from=builder /go/bin/* /go/bin/

RUN mkdir -p /go/src/github.com/prometheus-operator/prometheus-operator /.cache && \
	chmod -R 777 /go /.cache

WORKDIR /go/src/github.com/prometheus-operator/prometheus-operator
