FROM golang:1.13 as builder

ENV JSONNET_VERSION v0.14.0
# This corresponds to v2.15.2, but needs to be written as commit hash due to golang 1.13 issues
ENV PROMTOOL_VERSION d9613e5c466c6e9de548c4dae1b9aabf9aaf7c57
ENV GOLANGCILINT_VERSION v1.19.1
ENV JB_VERSION v0.2.0

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
RUN go get -u github.com/jteeuwen/go-bindata/...

# Add po-lint
WORKDIR /go/src/github.com/coreos/prometheus-operator
COPY . .
RUN GO111MODULE=on make po-lint && chmod +x po-lint && mv po-lint /go/bin/

FROM golang:1.13
RUN apt-get update -y && apt-get install -y make git jq gawk python-yaml && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/bin/jsonnetfmt /usr/local/bin/jsonnetfmt
COPY --from=builder /go/bin/* /go/bin/

RUN mkdir -p /go/src/github.com/coreos/prometheus-operator /.cache && \
	chmod -R 777 /go /.cache

WORKDIR /go/src/github.com/coreos/prometheus-operator
