---
title: "High Availability"
description: "High Availability is a must for the monitoring infrastructure."
lead: ""
date: 2021-03-08T08:49:31+00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 300
toc: true
---

High availability is not only important for customer facing software, but if the monitoring infrastructure is not highly available, then there is a risk that operations people are not notified for alerts of the customer facing software. Therefore high availability must be just as thought through for the monitoring stack, as for anything else.

## Prometheus

To run Prometheus in a highly available manner, two (or more) instances need to be running with the same configuration, that means they scrape the same targets, which in turn means they will have the same data in memory and on disk, which in turn means they are answering requests the same way. In reality this is not entirely true, as the scrape cycles can be slightly different, and therefore the recorded data can be slightly different. This means that single requests can differ slightly. For alert evaluation this situation does not change anything, as alerts are typically only fired when a certain query triggers for a period of time. For dashboarding this means sticky sessions (using `sessionAffinity` on a Kubernetes `Service`) should be used, to get consistent graphs when refreshing.

What all of the above means for Prometheus is that there is a problem when a single Prometheus instance is not able to scrape the entire infrastructure anymore. This is where Prometheus' sharding feature comes into play. It divides the targets Prometheus scrapes into multiple groups, small enough for a single Prometheus instance to scrape. If possible functional sharding is recommended. What is meant by functional sharding is that all instances of Service A are being scraped by Prometheus A. When functional sharding is not enough anymore, Prometheus is also able to perform sharding automatically which is easier but also has other effects that need to be taken into account. Single shards of Prometheus can be run highly available as described before. To be able to query all data, Prometheus federation can be used to fan in the relevant data to perform queries and alerting, which is only necessary if these queries actually need data from multiple shards.

One of the goals with the Prometheus Operator is that we want to completely automate sharding and federation. We are currently implementing some of the groundwork to make this possible, and figuring out the best approach to do so, but it is definitely on the roadmap!

## Alertmanager

The final step of the high availability scheme between Prometheus and Alertmanager is that Prometheus, when an alert triggers, actually fires alerts against *all* instances of an Alertmanager cluster. Prometheus can discover all Alertmanagers through the Kubernetes API.

The Alertmanager, starting with the `v0.5.0` release, ships with a high availability mode. It implements a gossip protocol to synchronize instances of an Alertmanager cluster regarding notifications that have been sent out, to prevent duplicate notifications. It is an AP (available and partition tolerant) system. Being an AP system, means that notifications are guaranteed to be sent at least once. 

The Prometheus Operator ensures that Alertmanager clusters are properly configured to run highly available on Kubernetes, and allows easy configuration of Alertmanagers discovery for Prometheus.

## Exporters

For exporters, high availability depends on the particular exporter. In the case of [`kube-state-metrics`](https://github.com/kubernetes/kube-state-metrics), because it is effectively stateless, it is the same as running any other stateless service in a highly available manner. Simply run multiple replicas that are being load balanced. Key for this is that the backing service, in this case the Kubernetes apiserver is highly available, ensuring that the data source of `kube-state-metrics` is not a single point of failure.
