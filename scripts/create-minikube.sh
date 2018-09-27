#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

export MINIKUBE_VERSION=v0.28.0
export KUBERNETES_VERSION=v1.10.0

sudo mount --make-rshared /
sudo mount --make-rshared /proc
sudo mount --make-rshared /sys

curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl && \
    chmod +x kubectl &&  \
    sudo mv kubectl /usr/local/bin/
curl -Lo minikube https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-linux-amd64 && \
    chmod +x minikube && \
    sudo mv minikube /usr/local/bin/

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir "${HOME}"/.kube || true
touch "${HOME}"/.kube/config

export KUBECONFIG=$HOME/.kube/config
minikube version
minikube addons enable storage-provisioner
sudo minikube start --vm-driver=none --bootstrapper=localkube --kubernetes-version=$KUBERNETES_VERSION --extra-config=apiserver.Authorization.Mode=RBAC --extra-config=apiserver.feature-gates=CustomResourceSubresources=true

minikube update-context

# waiting for node(s) to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done

kubectl apply -f scripts/minikube-rbac.yaml

# waiting for kube-dns to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lk8s-app=kube-dns -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for kube-dns to be available"; kubectl get pods --all-namespaces; done
