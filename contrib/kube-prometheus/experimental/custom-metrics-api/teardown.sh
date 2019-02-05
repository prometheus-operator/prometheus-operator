#!/usr/bin/env bash

kubectl delete -n monitoring -f custom-metrics-apiserver-resource-reader-cluster-role-binding.yaml
kubectl delete -n monitoring -f custom-metrics-apiservice.yaml
kubectl delete -n monitoring -f custom-metrics-cluster-role.yaml
kubectl delete -n monitoring -f custom-metrics-configmap.yaml
kubectl delete -n monitoring -f hpa-custom-metrics-cluster-role-binding.yaml
