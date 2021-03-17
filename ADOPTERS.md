---
title: "Adopters"
date: 2021-03-08T23:50:39+01:00
draft: false
---

This document tracks people and use cases for the Prometheus Operator in production. By creating a list of production use cases we hope to build a community of advisors that we can reach out to with experience using various the Prometheus Operator applications, operation environments, and cluster sizes. The Prometheus Operator development team may reach out periodically to check-in on how the Prometheus Operator is working in the field and update this list.

## Giant Swarm

[giantswarm.io](https://www.giantswarm.io/)

Environments: AWS, Azure, Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional tight Giant Swarm integrations)

Details:
- One prometheus operator per management cluster and one prometheus instance per workload cluster
- Customers can also install kube-prometheus for their workload using our App Platform
- 760000 samples/s
- 35M active series

## Lunar

[lunar.app](https://lunar.app/)

Environments: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- One prometheus operator in our platform cluster and one prometheus instance per workload cluster
- 17k samples/s
- 841k active series

## OpenShift

[openshift.com](https://www.openshift.com/)

Environments: AWS, Azure, Google Cloud, Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional tight OpenShift integrations)

This is a meta user; please feel free to document specific OpenShift users!

All OpenShift clusters use the Prometheus Operator to manage the cluster monitoring stack as well as user workload monitoring. This means the Prometheus Operator's users include all OpenShift customers.

## Polar Signals

[polarsignals.com](https://polarsignals.com/)

Environment: Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- HA Pair of Prometheus
- 4000 samples/s
- 100k active series

## Skyscanner

[skyscanner.net](https://skyscanner.net/)

Environment: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- HA Pairs of Prometheus
- 25000 samples/s
- 1.2M active series

## Veepee

[veepee.com](https://www.veepee.com)

Environments: Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- HA Pair of Prometheus
- 517000 samples/s
- 10.7M active series

## VSHN AG

[vshn.ch](https://www.vshn.ch/)

Environments: AWS, Azure, Google Cloud, cloudscale.ch, Exoscale, Swisscom

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- A huge fleet of OpenShift and Kubernetes clusters, each using Prometheus Operator
- All managed by [Project Syn](https://syn.tools/), leveraging Commodore Components like [component-rancher-monitoring](https://github.com/projectsyn/component-rancher-monitoring) which re-uses Prometheus Operator

---

## <Insert Company/Organization Name>

https://our-link.com/

Environments: AWS, Azure, Google Cloud, Bare Metal, etc

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes | No

Details (optional):
- HA Pair of Prometheus
- 1000 samples/s (query: `rate(prometheus_tsdb_head_samples_appended_total[5m])`)
- 10k active series (query: `prometheus_tsdb_head_series`)
