#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

export MINIKUBE_VERSION=v1.9.2
export KUBERNETES_VERSION=v1.18.1

sudo mount --make-rshared /
sudo mount --make-rshared /proc
sudo mount --make-rshared /sys

curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl && \
    chmod +x kubectl &&  \
    sudo mv kubectl /usr/local/bin/
curl -Lo minikube https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-linux-amd64 && \
    chmod +x minikube && \
    sudo mv minikube /usr/local/bin/

export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir "${HOME}"/.kube || true
touch "${HOME}"/.kube/config

export KUBECONFIG=$HOME/.kube/config

# minikube config
minikube config set WantUpdateNotification false
minikube config set WantReportErrorPrompt false
minikube config set WantNoneDriverWarning false
minikube config set vm-driver none
minikube config view

# Setup Docker according to https://kubernetes.io/docs/setup/production-environment/container-runtimes/#docker
cat > daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
sudo mkdir -p /etc/docker
sudo mv daemon.json /etc/docker/daemon.json

# Restart Docker
sudo systemctl daemon-reload
sudo systemctl restart docker
docker info

minikube version
sudo minikube start --kubernetes-version=$KUBERNETES_VERSION
sudo chown -R travis: /home/travis/.minikube/

minikube update-context

# waiting for node(s) to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done

kubectl apply -f scripts/minikube-rbac.yaml
