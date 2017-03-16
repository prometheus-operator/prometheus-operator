#!/bin/bash

set -ex

# Clean up any existing build.
rm -rf assets/k8s.io
mkdir -p assets/k8s.io/kubernetes

VERSIONS=( "1.4.7" "1.5.1" )

for VERSION in ${VERSIONS[@]}; do
    if [ ! -f assets/v${VERSION}.zip ]; then
        wget https://github.com/kubernetes/kubernetes/archive/v${VERSION}.zip -O assets/v${VERSION}.zip
    fi

    # Copy source tree to assets/k8s.io/kubernetes. Newer versions overwrite existing ones.
    unzip assets/v${VERSION}.zip -d assets/
    cp -r assets/kubernetes-${VERSION}/* assets/k8s.io/kubernetes
    rm -rf assets/kubernetes-${VERSION}
done

# Remove any existing generated code.
rm -rf api apis config.go runtime util types.go watch

# Generate Go code from proto definitions.
PKG=$PWD
cd assets

protobuf=$( find k8s.io/kubernetes/pkg/{api,apis,util,runtime,watch} -name '*.proto' )
for file in $protobuf; do
    echo $file
    # Generate protoc definitions at the base of this repo.
    protoc --gofast_out=$PKG $file
done

cd -

mv k8s.io/kubernetes/pkg/* .
rm -rf k8s.io

# Copy kubeconfig structs.
client_dir="client/unversioned/clientcmd/api/v1"
cp assets/k8s.io/kubernetes/pkg/${client_dir}/types.go config.go
sed -i 's|package v1|package k8s|g' config.go

# Rewrite imports for the generated fiels.
sed -i 's|"k8s.io/kubernetes/pkg|"github.com/ericchiang/k8s|g' $( find {api,apis,config.go,util,runtime,watch} -name '*.go' )
sed -i 's|"k8s.io.kubernetes.pkg.|"github.com/ericchiang.k8s.|g' $( find {api,apis,config.go,util,runtime,watch} -name '*.go' )

# Clean up assets.
rm -rf assets/k8s.io

# Generate HTTP clients from Go structs.
go run gen.go

# Fix JSON marshaling for types need by third party resources.
cat << EOF >> api/unversioned/time.go
package unversioned

import (
    "encoding/json"
    "time"
)

// JSON marshaling logic for the Time type. Need to make
// third party resources JSON work.

func (t Time) MarshalJSON() ([]byte, error) {
    var seconds, nanos int64
    if t.Seconds != nil {
        seconds = *t.Seconds
    }
    if t.Nanos != nil {
        nanos = int64(*t.Nanos)
    }
    return json.Marshal(time.Unix(seconds, nanos))
}

func (t *Time) UnmarshalJSON(p []byte) error {
    var t1 time.Time
    if err := json.Unmarshal(p, &t1); err != nil {
        return err
    }
    seconds := t1.Unix()
    nanos := int32(t1.UnixNano())
    t.Seconds = &seconds
    t.Nanos = &nanos
    return nil
}
EOF
gofmt -w api/unversioned/time.go
