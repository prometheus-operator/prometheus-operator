FROM golang:1.12-stretch

ENV JSONNET_VERSION 0.12.1

RUN apt-get update -y && apt-get install -y g++ make git jq && \
    rm -rf /var/lib/apt/lists/*
RUN curl -Lso - https://github.com/google/jsonnet/archive/v${JSONNET_VERSION}.tar.gz | \
    tar xfz - -C /tmp && \
    cd /tmp/jsonnet-${JSONNET_VERSION} && \
    make && mv jsonnet /usr/local/bin && \
    rm -rf /tmp/jsonnet-${JSONNET_VERSION}
RUN go get github.com/brancz/gojsontoyaml
RUN go get github.com/campoy/embedmd
RUN go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb

RUN mkdir -p /go/src/github.com/coreos/prometheus-operator
WORKDIR /go/src/github.com/coreos/prometheus-operator

RUN chmod -R 777 /go
