# Current state and roadmap

The following is a loose collection of potential features that fit the future
scope of the operator. Their exact implementation and viability are subject
to further discussion.

# Prometheus

### Rule files

The Operator creates an empty ConfigMap of the name `<prometheus-name>-rules` if it
doesn't exist yet. It is left to the user to populate it with the desired rules.

It is still up for discussion whether it should be possible to include rule files living
in arbitrary ConfigMaps by their labels.
Intuitively, it seems fitting to define in each `ServiceMonitor` which rule files (based 
label selections over ConfigMaps) should be deployed with it.
However, rules act upon all metrics in a Prometheus server. Hence, defining the
relationship in each `ServiceMonitor` may cause undesired interference.
 
### Dashboards

In the future, the Prometheus Operator should register new Prometheus setups
it brought up as data sources in potential Grafana deployments. 

### Horizontal sharding

Prometheus has basic capabilities to run horizontally sharded setups. This is only
necessary in the largest of clusters. The Operator is an ideal candidate to manage the
sharding process and make it appear seamless to the user.

### Federation

Prometheus supports federation patterns, which have to be setup manually. Direct support
in the operator is a desirable feature that could potentially tightly integrate with
Kubernetes cluster federation to minimize user-defined configuration.

## Alertmanager

### Configuration file

The Operator expects a `ConfigMap` of the name `<alertmanager-name>` which
contains the configuration for the Alertmanager instances to run. It is left to
the user to populate it with the desired configuration. Note, that the
Alertmanager pods will stay in a `Pending` state as long as the `ConfigMap`
does not exist.

### Deployment

The Alertmanager, in high availability mode, is a distributed system. A
desired deployment ensures no data loss and zero downtime while performing a
deployment. Zero downtime is simply done as the Alertmanager is running high
availability mode. No data loss is achieved by using PVCs and attaching the
same volumes a previous Alertmanager instance had to a new instance. The hard
part, however, is knowing whether a new instance is healthy or not.

A healthy node would be one that has joined the existing mesh network and has
been communicated the state that it missed while that particular instance was
down for the upgrade.

Currently there is no way to tell whether an Alertmanager instance is healthy
under the above conditions. There are discussions of using vector clocks to
resolve merges in the above mentioned situation, and ensure on a best effort
basis that joining the network was successful.

> Note that single instance Alertmanager setups will therefore not have zero
> downtime on deployments.

The current implementation of rolling deployments simply decides based on the
Pod state whether an instance is considered healthy. This mechanism may be part
of an implementation with the characteristics that are mentioned above.

