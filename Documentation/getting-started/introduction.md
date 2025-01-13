---
weight: 101
toc: true
title: Introduction
menu:
    docs:
        parent: getting-started
lead: ""
lastmod: "2020-10-06T08:48:57+00:00"
images: []
draft: false
description: The Prometheus Operator provides Kubernetes native deployment and management of Prometheus and related monitoring components
date: "2020-10-06T08:48:57+00:00"
---

Prometheus Operator is a [Kubernetes Operator](https://github.com/cncf/tag-app-delivery/blob/main/operator-wg/whitepaper/Operator-WhitePaper_v1-0.md#foundation) that provides Kubernetes native deployment and management of [Prometheus](https://prometheus.io/) and related monitoring components.

The Prometheus operator includes, but is not limited to, the following features:

- **Kubernetes Custom Resources**: Use Kubernetes custom resources to deploy and manage Prometheus, Alertmanager, and related components.

- **Simplified Deployment Configuration**: Configure the fundamentals of Prometheus like versions, persistence, retention policies, and replicas from a native Kubernetes resource.

- **Prometheus Target Configuration**: Automatically generate monitoring target configurations based on familiar Kubernetes label queries; no need to learn a Prometheus specific configuration language.

Prometheus Operator provides a set of Custom Resource Definitions(CRDs) that allows you to configure your Prometheus and related instances. Currently, the CRDs provided by Prometheus Operator are:

- Prometheus
- Alertmanager
- ThanosRuler
- ServiceMonitor
- PodMonitor
- Probe
- PrometheusRule
- AlertmanagerConfig
- PrometheusAgent
- ScrapeConfig

> Check the [Design]({{<ref "design">}}) page for an overview of all the resources provided by Prometheus Operator.

### Goals

- To significantly reduce the effort required to configure, implement and manage all components of Prometheus based monitoring stack.

- **Automation** - Automate the management of Prometheus monitoring targets, ultimately increasing efficiency. This automation is performed by the use of Kubernetes [Custom Resource Definition](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/). The Operator introduces custom resources like `Prometheus`, `Alertmanager`, `ThanosRuler`, and others, which help automate the deployment and configuration of these resources.

- **Configuration Abstraction and Validation** - Instead of learning and manually writing Prometheus Relabeling rules (which can be time consuming), you can simply use Kubernetes [Label Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors). `ServiceMonitor`, `PodMonitor` and `Probe` custom resources provide this abstraction. The Operator also removes the complexity of validating the configuration of `AlertmanagerConfig` and `PrometheusRule` objects.

- **Scaling** - There are many scaling-related features provided by the Operator like [ThanosRuler](https://prometheus-operator.dev/docs/platform/thanos/#thanos-ruler) custom resource for rule evaluation, workload distribution across multiple Prometheus instances using scrape target sharding, and running [Thanos sidecar](https://thanos.io/v0.4/components/sidecar/) in Prometheus instance for long-term storage.

### Next Steps

By now, you have the basic idea about Prometheus Operator!!

Take a look at these guides to get into action with Prometheus Operator.

<!-- Getting-Started -->

{{<
link-card title="Getting-Started" href="https://prometheus-operator.dev/docs/getting-started/introduction/" description="Get started with Prometheus-Operator.">}}

<!-- API -->

{{<
link-card title="API Reference" href="https://prometheus-operator.dev/docs/api-reference/api/" description="Reference for different fields of Custom Resources in Prometheus-Operator.">}}

<!-- Platform Guide -->

{{<
link-card title="Platform Guide" href="https://prometheus-operator.dev/docs/platform/webhook/" description="Set up, configure and manage instances of Prometheus-Operator, Prometheus, Alertmanager and ThanosRuler resources.">}}

<!-- Developer Guide -->

{{<
link-card title="Developer Guide" href="https://prometheus-operator.dev/docs/developer/getting-started/" description="Learn how to configure scraping, alerting, and recording rules for your applications.">}}

<!-- Community -->

{{<
link-card title="Community" href="https://prometheus-operator.dev/docs/community/contributing/" description="Join and interact with Prometheus-Operator community.">}}
