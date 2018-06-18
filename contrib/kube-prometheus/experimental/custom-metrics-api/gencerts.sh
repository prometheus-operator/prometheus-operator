#!/usr/bin/env bash

# Detect if we are on mac or should use GNU base64 options
case $(uname) in
        Darwin)
            b64_opts='-b=0'
            ;; 
        *)
            b64_opts='--wrap=0'
esac

go get -v -u github.com/cloudflare/cfssl/cmd/...

export PURPOSE=metrics
openssl req -x509 -sha256 -new -nodes -days 365 -newkey rsa:2048 -keyout ${PURPOSE}-ca.key -out ${PURPOSE}-ca.crt -subj "/CN=ca"
echo '{"signing":{"default":{"expiry":"43800h","usages":["signing","key encipherment","'${PURPOSE}'"]}}}' > "${PURPOSE}-ca-config.json"

export SERVICE_NAME=custom-metrics-apiserver
export ALT_NAMES='"custom-metrics-apiserver.monitoring","custom-metrics-apiserver.monitoring.svc"'
echo "{\"CN\":\"${SERVICE_NAME}\", \"hosts\": [${ALT_NAMES}], \"key\": {\"algo\": \"rsa\",\"size\": 2048}}" | \
       	cfssl gencert -ca=metrics-ca.crt -ca-key=metrics-ca.key -config=metrics-ca-config.json - | cfssljson -bare apiserver

cat <<-EOF > cm-adapter-serving-certs.yaml
apiVersion: v1
kind: Secret
metadata:
  name: cm-adapter-serving-certs
data:
  serving.crt: $(base64 ${b64_opts} < apiserver.pem)
  serving.key: $(base64 ${b64_opts} < apiserver-key.pem)
EOF
