# Adopters

This document tracks people and use cases for the Prometheus Operator in production. By creating a list of production use cases we hope to build a community of advisors that we can reach out to with experience using various the Prometheus Operator applications, operation environments, and cluster sizes. The the Prometheus Operator development team may reach out periodically to check-in on how the Prometheus Operator is working in the field and update this list.

## All OpenShift Users

https://www.openshift.com/

Environments: AWS, Azure, Google Cloud, Bare Metal, etc

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional tight OpenShift integrations)

This is a meta user; please feel free to document specific OpenShift users!

All OpenShift clusters use the Prometheus Operator to manage the cluster monitoring stack as well as user workload monitoring. This means the Prometheus Operator's users include all OpenShift customers.

## Polar Signals

https://polarsignals.com/

Environment: Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- HA Pair of Prometheus
- 4000 samples/s
- 100k active series
