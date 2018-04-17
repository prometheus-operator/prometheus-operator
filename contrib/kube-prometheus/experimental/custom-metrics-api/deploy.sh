#!/usr/bin/env bash

kubectl create -f custom-metrics-apiserver-auth-delegator-cluster-role-binding.yaml
kubectl create -f custom-metrics-apiserver-auth-reader-role-binding.yaml
kubectl -n monitoring create -f cm-adapter-serving-certs.yaml
kubectl -n monitoring create -f custom-metrics-apiserver-deployment.yaml
kubectl create -f custom-metrics-apiserver-resource-reader-cluster-role-binding.yaml
kubectl -n monitoring create -f custom-metrics-apiserver-service-account.yaml
kubectl -n monitoring create -f custom-metrics-apiserver-service.yaml
kubectl create -f custom-metrics-apiservice.yaml
kubectl create -f custom-metrics-cluster-role.yaml
kubectl create -f custom-metrics-resource-reader-cluster-role.yaml
kubectl create -f hpa-custom-metrics-cluster-role-binding.yaml
