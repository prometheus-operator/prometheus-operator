#!/usr/bin/env bash

kubectl apply -n monitoring custom-metrics-apiserver-resource-reader-cluster-role-binding.yaml
kubectl apply -n monitoring custom-metrics-apiservice.yaml
kubectl apply -n monitoring custom-metrics-cluster-role.yaml
kubectl apply -n monitoring custom-metrics-configmap.yaml
kubectl apply -n monitoring hpa-custom-metrics-cluster-role-binding.yaml
