#!/usr/bin/env bash

kubectl apply -n monitoring -f custom-metrics-apiserver-resource-reader-cluster-role-binding.yaml
kubectl apply -n monitoring -f custom-metrics-apiservice.yaml
kubectl apply -n monitoring -f custom-metrics-cluster-role.yaml
kubectl apply -n monitoring -f custom-metrics-configmap.yaml
kubectl apply -n monitoring -f hpa-custom-metrics-cluster-role-binding.yaml
