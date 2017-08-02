#!/usr/bin/env bash

hack/concat-kubernetes-manifests.sh example/rbac/prometheus-operator/*.yaml > bundle.yaml

