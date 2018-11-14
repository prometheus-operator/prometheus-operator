#!/usr/bin/env bash

kubectl delete -n monitoring custom-metrics-apiserver-resource-reader-cluster-role-binding.yaml
kubectl delete -n monitoring custom-metrics-apiservice.yaml
kubectl delete -n monitoring custom-metrics-cluster-role.yaml
kubectl delete -n monitoring custom-metrics-configmap.yaml
kubectl delete -n monitoring hpa-custom-metrics-cluster-role-binding.yaml
