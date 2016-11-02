# ServiceMonitor

The `ServiceMonitor` third party resource (TPR) allows to declaratively define
how a dynamic set of services should be monitored. Which services are selected
to be monitored with the desired configuration is defined using label selections.
This allows to dynamically express monitoring without having to update additional
configuration for services that follow known monitoring patterns.

A service may expose one or more service ports, which are backed by a list
of multiple endpoints that point to a pod in the common case.

In the `endpoints` section of the TPR, we can configure which ports of these
endpoints we want to scrape for metrics and with which paramters. For advanced use
cases one may want to monitor ports of backing pods, which are not directly part
of the service endpoints. This is also made possible by the Prometheus Operator.


## Specification

...

## Current state and roadmap

### Namespaces

While `ServiceMonitor`s must live in the same namespace as the `Prometheus` TPR,
discovered targets may come from any namespace. This is important to allow cross-namespace
monitoring use cases, e.g. for meta-monitoring.

Currently, targets are always discovered from all namespaces. In the future, the
`ServiceMonitor` should allow to restrict this to one or more namespaces.
How such a configuration would look like, i.e. explicit namespaces, selection by labels,
or both, and what the default behavior should be is still up for discussion.
