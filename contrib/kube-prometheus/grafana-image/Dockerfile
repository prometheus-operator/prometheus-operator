FROM debian:9.3-slim

RUN apt-get update && apt-get install -qq -y wget tar sqlite && \
    wget -O /tmp/grafana.tar.gz https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana-4.6.3.linux-x64.tar.gz && \
    tar -zxvf /tmp/grafana.tar.gz -C /tmp && mv /tmp/grafana-4.6.3 /grafana && \
    rm -rf /tmp/grafana.tar.gz

ADD config.toml /grafana/conf/config.toml

USER       nobody
EXPOSE     3000
VOLUME     [ "/data" ]
WORKDIR    /grafana
ENTRYPOINT [ "/grafana/bin/grafana-server" ]
CMD        [ "-config=/grafana/conf/config.toml" ]
