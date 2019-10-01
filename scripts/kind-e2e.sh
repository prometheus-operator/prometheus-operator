#!/usr/bin/env bash

# bash strict mode
set -euo pipefail

KUBERNETES_VERSION="v1.16.2"
KIND_VERSION="v0.7.0"
KIND_NODE_IMAGE="kindest/node:$KUBERNETES_VERSION"

# "unique" cluster identifier
CLUSTER="$USER-$RANDOM"

cleanup() {
	kind delete cluster --name "$CLUSTER" --kubeconfig="/tmp/kubeconfig.$CLUSTER"
	unset KUBECONFIG
}

# Install kind if not already installed
if ! command -v kind; then
    curl -Lo kind "https://github.com/kubernetes-sigs/kind/releases/download/$KIND_VERSION/kind-linux-amd64"
    chmod +x kind
    sudo mv kind /usr/local/bin/
    PATH="${PATH}:/usr/local/bin/"
fi

# Install kubectl if not already installed
if ! command -v kubectl; then
    curl -Lo kubectl "https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
    PATH="${PATH}:/usr/local/bin/"
fi

# Go to the root of the repo
cd "$(git rev-parse --show-cdup)"

# Make sure scripts cleanups after itself
trap cleanup EXIT

# Start a cluster
kind create cluster --image="$KIND_NODE_IMAGE" --name="$CLUSTER" --kubeconfig="/tmp/kubeconfig.$CLUSTER"
export KUBECONFIG="/tmp/kubeconfig.$CLUSTER"

# Create images
make image

# Load images
for n in "operator" "config-reloader"; do
	kind load docker-image --name="$CLUSTER" "quay.io/coreos/prometheus-$n:$(git rev-parse --short HEAD)"
done

# Wait for cluster to be fully ready
until [ "$(kubectl get pods -n kube-system --field-selector=status.phase==Running | wc -l )" -eq 9 ]; do
	echo "Waiting for cluster to finish bootstraping"
	sleep 5
done

# SHOW CLUSTER STATUS
kubectl get pods --all-namespaces
sleep 1m

# Run e2e tests
make test-e2e
