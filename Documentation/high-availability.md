---
weight: 208
toc: true
title: High Availability
menu:
    docs:
        parent: operator
lead: ""
images: []
draft: false
description: High Availability is a must for the monitoring infrastructure.
---

High availability is not only important for customer facing software, but if the monitoring infrastructure is not highly available, then there is a risk that operations people are not notified for alerts of the customer facing software. Therefore high availability must be just as thought through for the monitoring stack, as for anything else.

## Prometheus

To run Prometheus in a highly available manner, two (or more) instances need to be running with the same configuration except that they will have one external label with a different value to identify them. The Prometheus instances scrape the same targets and evaluate the same rules, hence they will have the same data in memory and on disk, with a slight twist that given their different external label, the scrapes and evaluations won't happen at exactly the same time. As a consequence, query requests executed against each Prometheus instance may return slightly different results. For alert evaluation this situation does not change anything, as alerts are typically only fired when a certain query triggers for a period of time. For dashboarding, sticky sessions (using `sessionAffinity` on the Kubernetes `Service`) should be used, to get consistent graphs when refreshing or you can use something like [Thanos Querier](https://thanos.io/tip/components/query.md/) to federate the data.

Running multiple Prometheus instances avoids having a single point of failure but it doesn't help scaling out Prometheus in case a single Prometheus instance can't handle all the targets and rules. This is where Prometheus' sharding feature comes into play. Sharding aims at splitting the scrape targets into multiple groups, each assigned to one Prometheus shard and small enough that they can be handled by a single Prometheus instance. If possible, functional sharding is recommended: in this case, the Prometheus shard X scrapes all pods of Service A, B and C while shard Y scrapes pods from Service D, E and F. When functional sharding is not possible, the Prometheus Operator is also able to support automatic sharding: the targets will be assigned to Prometheus shards based on their addresses. The main drawback of this solution is the additional complexity: to query all data, query federation (e.g. Thanos Query) and distributed rule evaluation engine (e.g. Thanos Ruler) should be deployed to fan in the relevant data for queries and rule evaluations. Single shards of Prometheus can be run highly available as described before.

One of the goals with the Prometheus Operator is that we want to completely automate sharding and federation. We are currently implementing some of the groundwork to make this possible, and figuring out the best approach to do so, but it is definitely on the roadmap!

## Alertmanager

To ensure high-availability of the Alertmanager service, Prometheus instances are configured to send their alerts to all configured Alertmanager instances (as described in the [Alertmanager documentation](https://prometheus.io/docs/alerting/latest/alertmanager/#high-availability)). The Alertmanager instances creates a gossip-based cluster to replicate alert silences and notification logs.

The Prometheus Operator manages the following configuration
* Alertmanager discovery using the Kubernetes API for Prometheus.
* Highly-available cluster for Alertmanager when replicas > 1.

## Exporters

For exporters, high availability depends on the particular exporter. In the case of [`kube-state-metrics`](https://github.com/kubernetes/kube-state-metrics), because it is effectively stateless, it is the same as running any other stateless service in a highly available manner. Simply run multiple replicas that are being load balanced. Key for this is that the backing service, in this case the Kubernetes API server is highly available, ensuring that the data source of `kube-state-metrics` is not a single point of failure.
