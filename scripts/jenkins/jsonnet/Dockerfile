FROM golang:1.8-stretch

ADD https://github.com/google/jsonnet/archive/v0.9.4.tar.gz /tmp
RUN apt-get update -y && apt-get install -y g++ make git python-pip
RUN cd /tmp && tar xvfz v0.9.4.tar.gz && cd jsonnet-0.9.4 && make && mv jsonnet /usr/local/bin && cd / && rm -rf /tmp/jsonnet-0.9.4 /tmp/jsonnet-0.9.4.tar.gz

RUN git clone https://github.com/ksonnet/ksonnet-lib.git /ksonnet-lib && cd /ksonnet-lib && git checkout bd6b2d618d6963ea6a81fcc5623900d8ba110a32

RUN pip install json2yaml
RUN mkdir -p /go/src/github.com/coreos/prometheus-operator
WORKDIR /go/src/github.com/coreos/prometheus-operator
