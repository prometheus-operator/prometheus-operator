FROM golang:1.10-stretch

ENV JSONNET_VERSION 0.10.0

RUN apt-get update -y && apt-get install -y g++ make git python-pip python-virtualenv
RUN cd /tmp && wget https://github.com/google/jsonnet/archive/v${JSONNET_VERSION}.tar.gz && \
    tar xvfz v${JSONNET_VERSION}.tar.gz && \
    cd jsonnet-${JSONNET_VERSION} && \
    make && mv jsonnet /usr/local/bin && \
    rm -rf /tmp/v${JSONNET_VERSION}.tar.gz /tmp/jsonnet-${JSONNET_VERSION}

RUN git clone https://github.com/ksonnet/ksonnet-lib.git /ksonnet-lib && \
    cd /ksonnet-lib && \
    git checkout 152e1979764f713239ca25bcb0f3315e69215119

RUN pip install json2yaml
RUN mkdir -p /go/src/github.com/coreos/prometheus-operator
WORKDIR /go/src/github.com/coreos/prometheus-operator
