# Prometheus

The `Prometheus` third party resource (TPR) declaratively defines
a desired Prometheus setup to run in a Kubernetes cluster. It provides
options to configure replication, persistent storage, and Alertmanagers to
which the deployed Prometheus instances send alerts to.

For each `Prometheus` TPR, the Operator deploys a properly configured PetSet
in the same namespace. The Prometheus pods are configured to include two
ConfigMaps, `<prometheus-name>` and `<prometheus-name>-rules`, which respectively
hold the used configuration file and multiple Prometheus rule files, which may 
contain alerting and recording rules. 

The TPR allows to specify which [`ServiceMonitor`s](./service-monitor.md)
should be covered by the deployed Prometheus instances based on label selection.
The Operator than generates a configuration based on the included `ServiceMonitor`s
and updates it in the ConfigMap. It continously does so for all changes that
are made to `ServiceMonitor`s or the `Prometheus` TPR itself.

If no selection of `ServiceMonitor`s is provided, the Operator leaves management
of the ConfigMap to the user, which allows to provide custom configurations while
still benefiting from the Operator's capabilities of managing Prometheus setups.

## Specification

...


## Current state and roadmap

### Rule files

The Operator creates an empty ConfigMap of the name `<prometheus-name>-rules` if it
doesn't exist yet. It is left to the user to populate it with the desired rules.

It is still up for discussion whether it should be possible to include rule files living
in arbitrary ConfigMaps by their labels.
Intuitively, it seems fitting to define in each `ServiceMonitor` which rule files (based 
label selections over ConfigMaps) should be deployed with it.
However, rules act upon all metrics in a Prometheus server. Hence, defining the
relationship in each`ServiceMonitor` may cause undesired interference.
 
### Alerting

The TPR allows to configure multiple namespace/name pairs of Alertmanagers
services. The Prometheus instances will send their alerts to each endpoint
of this service.

Currently Prometheus only allows to configure Alertmanager URLs via flags
on startup. Thus the Prometheus pods have to be restarted manually if the 
endpoints change.
PetSets or manually maintained headless services in Kubernetes, allow to
provide stable URLs working around this. In the future, Prometheus will allow
for dynmaic service discovery of Alertmanagers ([tracking issue](https://github.com/prometheus/prometheus/issues/2057)). 

### Cluster-wide version

Currently the controller installs a default version with optional explicit
definition of the used version in the TPR.
In the future, there should be a cluster wide version so that the controller
can orchestrate upgrades of all running Prometheus setups.

### Dashboards

In the future, the Prometheus Operator should register new Prometheus setups
it brought up as data sources in potential Grafana deployments. 

### Resource limits

Prometheus instances are deployed with default values for requested and maximum
resource usage of CPU and memory. This will be made configurable in the `Prometheus` 
TPR eventually.
Prometheus comes with a variety of configuration flags for its storage engine, that
have to be tuned for better performance in large Prometheus servers. It will be the
operators job to tune those correctly to be aligned with the experiences load
and the resource limits configured by the user.

### Horzintal sharding

Prometheus has basic capabilities to run horizontally shareded setups. This is only
necessary in the largest of clusters. The Operator is an ideal candidate to manage the
sharding process and make it appear seamless to the user.
