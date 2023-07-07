---
title: Adopters
draft: false
date: "2021-03-08T23:50:39+01:00"
---

<!--

Insert your entry using this template keeping the list alphabetically sorted:

## <Company/Organization Name>

https://our-link.com/

Environments: AWS, Azure, Google Cloud, Bare Metal, etc

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes | No

Details (optional):
- HA Pair of Prometheus
- 1000 samples/s (query: `rate(prometheus_tsdb_head_samples_appended_total[5m])`)
- 10k active series (query: `prometheus_tsdb_head_series`)

-->

This document tracks people and use cases for the Prometheus Operator in production. By creating a list of production use cases we hope to build a community of advisors that we can reach out to with experience using various the Prometheus Operator applications, operation environments, and cluster sizes. The Prometheus Operator development team may reach out periodically to check-in on how the Prometheus Operator is working in the field and update this list.

Go ahead and [add your organization](https://github.com/prometheus-operator/prometheus-operator/edit/main/ADOPTERS.md) to the list.

## CERN

[European Laboratory for Particle Physics](https://home.cern/)

Environments: On-premises

Prometheus is used extensively as part of the CERN Kubernetes infrastructure,
both managed and unmanaged. Metrics deployment is managed by the community owned
__kube-prometheus-stack__ helm chart. Be sure to check our [blog](https://kubernetes.web.cern.ch/).

Details:
- 400+ Kubernetes clusters, with cluster sizes ranging from few nodes to ~100s

Significant usage also exists outside Kubernetes for generic service and infrastructure monitoring.

## Clyso

[clyso.com](https://www.clyso.com/en)

Environments: Bare Metal, Opennebula

Uses kube-prometheus: Yes

Details:
- multiple K8s cluster with prometheus deployed through prom-operator
- several own ceph cluster providing metrics via ceph mgr prometheus module
- several customer ceph clusters pushing metrics via external pushgateway to our our central monitoring instances
- thanos receiver connected to own S3 storage

## Coralogix

[coralogix.com](https://coralogix.com)

Environments: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- Operator installed on each Kubernetes cluster, with Thanos aggregating metrics from a central query endpoint
- Two Prometheus instances per cluster
- Loose coupling between Kubernetes cluster administrators who manage alerting sinks and service owners who define alerts for their services
- 800K samples/s
- 30M active series

## Deckhouse

[deckhouse.io](https://deckhouse.io/)

Environments: AWS, Azure, Google Cloud, Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Deckhouse is a Kubernetes Platform. Its clusters running on any infrastructure are provided with the monitoring system based on highly available Prometheus and Prometheus Operator. Essential metrics are preconfigured out-of-the-box to ensure monitoring of all levels, from hardware and Kubernetes internals to the platform’s modules functionality. The monitoring-custom module simplifies adding custom metrics for user applications. Deckhouse also hosts a dedicated Prometheus instance in each cluster to store downsampled metric series for longer periods.

## Giant Swarm

[giantswarm.io](https://www.giantswarm.io/)

Environments: AWS, Azure, Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional tight Giant Swarm integrations)

Details:
- One prometheus operator per management cluster and one prometheus instance per workload cluster
- Customers can also install kube-prometheus for their workload using our App Platform
- 760000 samples/s
- 35M active series

## Gitpod

[gitpod.io](https://www.gitpod.io/)

Environments: Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional Gitpod mixins)

Details:
- One prometheus instance per cluster (8 so far)
- 20000 samples/s
- 1M active series

## Innovaccer

https://innovaccer.com/

Environments: AWS, Azure

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- multiple remote K8s cluster in which we have prometheus deployed through prom-operator.
- these remote prometheus instances push cluster metrics to central Thanos receiver which is connected to S3 storage.
- on top of Thanos we have Grafana for dashboarding and visualisation.

## Kinvolk Lokomotive Kubernetes

https://kinvolk.io/lokomotive-kubernetes/

Environments: AKS, AWS, Bare Metal, Equinix Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- Self-hosted (control plane runs as pods inside the cluster)
- Deploys full K8s stack (as a distro) or managed Kubernetes (currently only AKS supported)
- Deployed by Kinvolk for its own hosted infrastructure (including Flatcar Container Linux update server), as well as by Kinvolk customers and community users

## Lunar

[lunar.app](https://www.lunar.app/)

Environments: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- One prometheus operator in our platform cluster and one prometheus instance per workload cluster
- 17k samples/s
- 841k active series

## Mattermost

[mattermost.com](https://mattermost.com)

Environments: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- All Mattermost clusters use the Prometheus Operator with Thanos sidecar for cluster monitoring and central Thanos query component to gather all data.
- 977k samples/s
- 29.4M active series

## Nozzle

[nozzle.io](https://nozzle.io)

Environment: Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- 100k samples/s
- 1M active series

## OpenShift

[openshift.com](https://www.openshift.com/)

Environments: AWS, Azure, Google Cloud, Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes (with additional tight OpenShift integrations)

This is a meta user; please feel free to document specific OpenShift users!

All OpenShift clusters use the Prometheus Operator to manage the cluster monitoring stack as well as user workload monitoring. This means the Prometheus Operator's users include all OpenShift customers.

## Opstrace

[https://opstrace.com](https://opstrace.com)

Environments: AWS, Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): No

Opstrace installations use the Prometheus Operator internally to collect metrics and to alert. Opstrace users also often use the Prometheus Operator to scrape their own aplications and remote_write those metrics to Opstrace.

## Polar Signals

[polarsignals.com](https://www.polarsignals.com/)

Environment: Google Cloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- HA Pair of Prometheus
- 4000 samples/s
- 100k active series

## Robusta

[Robusta docs](https://docs.robusta.dev/master/)

Environments: EKS, GKE, AKS, and self-hosted Kubernetes

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

We're an open source project that builds upon the awesome Prometheus Operator. We run automated playbooks in response to Prometheus alerts and other events in your cluster. For example, you can automatically fetch logs and send them to Slack when a Prometheus alert occurs. All it takes is this YAML:

```yaml
triggers:
  - on_prometheus_alert:
      alert_name: KubePodCrashLooping
actions:
  - logs_enricher: {}
sinks:
  - slack
```

## Skyscanner

[skyscanner.net](https://skyscanner.net/)

Environment: AWS

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- HA Pairs of Prometheus
- 25000 samples/s
- 1.2M active series

## SUSE Rancher

[suse.com/products/suse-rancher](https://www.suse.com/products/suse-rancher/)

Environments: RKE, RKE2, K3s, Windows, AWS, Azure, Google Cloud, Bare Metal, etc.

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Rancher Monitoring supports use cases for Prometheus Operator across various different
cluster types and setups that are managed via the Rancher product. All Rancher users that
install Monitoring V2 deploy this chart.

For more information, please see [how Rancher monitoring works](https://rancher.com/docs/rancher/v2.6/en/monitoring-alerting/how-monitoring-works/).

The open-source rancher-monitoring Helm chart (based on [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)) can be found at [rancher/charts](https://github.com/rancher/charts).

## Trendyol

[trendyol.com](https://trendyol.com)

Environments: OpenStack, VMware vCloud

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details:
- All Kubernetes clusters use one Prometheus Operator instance with remote write enabled
- Prometheus instances push metrics to central H/A VirtualMetric, which gathers all data from clusters in 3 different data centers
- Grafana is used for dashboarding and visualization
- 7.50M samples/s
- 190M active series

## Veepee

[veepee.com](https://www.veepee.com)

Environments: Bare Metal

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- HA Pair of Prometheus
- 786000 samples/s
- 23.6M active series

## VSHN AG

[vshn.ch](https://www.vshn.ch/)

Environments: AWS, Azure, Google Cloud, cloudscale.ch, Exoscale, Swisscom

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes

Details (optional):
- A huge fleet of OpenShift and Kubernetes clusters, each using Prometheus Operator
- All managed by [Project Syn](https://syn.tools/), leveraging Commodore Components like [component-rancher-monitoring](https://github.com/projectsyn/component-rancher-monitoring) which re-uses Prometheus Operator

## Wise

[wise.com](https://wise.com)

Environments: Kubernetes, AWS (via some EC2)

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): No

Details (optional):
- About 30 HA pairs of sharded Promethei across 10 environments, wired together with Thanos
- Operator also helps us seamlessly manage anywhere between 600-1500 short-lived prometheus instances for our "integration" kubernetes cluster.
- ~15mn samples/s
- ~200mn active series

## <Insert Company/Organization Name>

https://our-link.com/

Environments: AWS, Azure, Google Cloud, Bare Metal, etc

Uses [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus): Yes | No

Details (optional):
- HA Pair of Prometheus
- 1000 samples/s (query: `rate(prometheus_tsdb_head_samples_appended_total[5m])`)
- 10k active series (query: `prometheus_tsdb_head_series`)
