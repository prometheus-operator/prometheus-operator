## Your Proposal Title

* **Owners:**
  * [Pger-Y](https://github.com/Pger-Y)

* **Related Tickets:**
  * [Discussion](https://github.com/prometheus-operator/prometheus-operator/discussions/7134#discussioncomment-11569949)

* **Other docs:**
  * n/a

This document proposes the creation of a Blackbox Operator and associated CRDs, similar to Prometheus-Operator, to manage blackbox-exporter instances and introduce a Module CRD for handling dynamic configuration of blackbox modules in a Kubernetes-native way.



## Why
As the above discussion described the Probe CRD is extremely convenient for icmp prober, the Probe CRD can define anything icmp needed ,but when we want to use http Probe , we must define header/body and plenty of something else in blackbox config first,because http is much more complex than icmp.

The situation is very similar to ScrapeConfig/ServiceMonitor that prometheus-operator does,We need a k8s way for generate blackbox config dynamic by reading the Probe Module Config 


### Pitfalls of the current solution

Using HTTP Probe with custom header and body,we must modify the confgiMap of config file and reload blackbox by hand which is not that much cloud native :(,this situation is really a copy of prometheus-operator does

## Goals

- Provide a way for users to self-service adding module config for their Probe CRD
- Consolidate the module configuration generation logic in a central point for other resources to use

### Audience

* Administrators providing Prometheus monitoring who want to enable self-service probe configuration for their users
* Teams wanting to manage blackbox module configurations declaratively through Kubernetes CRDs
* Users requiring a standardized, Kubernetes-native approach to monitor external targets

## Non-Goals

* This proposal aims to be an expansion of the Probe CRD rather than replacing it.

## How

As discussed in [Discussion](https://github.com/prometheus-operator/prometheus-operator/discussions/7134#discussioncomment-11569949), we propose to create a new Blackbox CRD that will function similarly to the Prometheus CRD. This controller will be responsible for managing blackbox-exporter instances and handling dynamic configuration updates and reloading.

## Alternatives

- Update blackbox configuration whenever a probe module is modified or added

## Action Plan

1. Create the BlackboxExporter CRD which will manage blackbox-exporter instances as either a DaemonSet or Deployment
2. Implement a configuration reloader sidecar container to automatically reload blackbox-exporter when configuration changes, similar to how Prometheus Operator handles configuration updates
