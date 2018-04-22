FROM golang:1.10-stretch

ENV JSONNET_VERSION 0.10.0

RUN apt-get update -y && apt-get install -y g++ make git jq
RUN cd /tmp && wget https://github.com/google/jsonnet/archive/v${JSONNET_VERSION}.tar.gz && \
    tar xvfz v${JSONNET_VERSION}.tar.gz && \
    cd jsonnet-${JSONNET_VERSION} && \
    make && mv jsonnet /usr/local/bin && \
    rm -rf /tmp/v${JSONNET_VERSION}.tar.gz /tmp/jsonnet-${JSONNET_VERSION}

RUN mkdir -p /go/src/github.com/ksonnet/ksonnet-lib && \
    git clone https://github.com/ksonnet/ksonnet-lib.git /go/src/github.com/ksonnet/ksonnet-lib && \
    cd /go/src/github.com/ksonnet/ksonnet-lib && \
    git checkout 152e1979764f713239ca25bcb0f3315e69215119

RUN mkdir -p /go/src/github.com/grafana/grafonnet-lib && \
    git clone https://github.com/grafana/grafonnet-lib.git /go/src/github.com/grafana/grafonnet-lib && \
    cd /go/src/github.com/grafana/grafonnet-lib && \
    git checkout 6e2a2e6d242a244c037e785b5907711a2fee0cb6

RUN mkdir -p /go/src/github.com/brancz/kubernetes-grafana && \
    git clone https://github.com/brancz/kubernetes-grafana.git /go/src/github.com/brancz/kubernetes-grafana && \
    cd /go/src/github.com/brancz/kubernetes-grafana && \
    git checkout 5df0eaa17361ce77f4a0486cc89f44f43bd51829

RUN go get github.com/brancz/gojsontoyaml
RUN go get github.com/campoy/embedmd
RUN mkdir -p /go/src/github.com/coreos/prometheus-operator
WORKDIR /go/src/github.com/coreos/prometheus-operator
