---
weight: 209
toc: true
title: Sharding
menu:
    docs:
        parent: operator
lead: ""
images: []
draft: false
description: Sharding Prometheus
---

When there are too much data to ingest and process, scaling Prometheus vertically may come to an end and it might become necessary to distribute scraped targets across multiple Prometheus shards.

## Design

The Prometheus operator will create `.spec.shards` StatefulSets multiplied by `.spec.replicas` pods.

By default, shards use the Prometheus `modulus` configuration which takes the hash of the source label values in order to split scraped
targets based on the number of shards. By default, Prometheus hashes the value of the
* `__address__` label for `ServiceMonitor` and `PodMonitor` resources
* `__param_target__` label for `Probe` resources

To query globally, deploy the Thanos querier connecting to all Thanos sidecars (in the same way, use the Thanos ruler to evaluate rules across shards). Another option is to remote write the samples to a central location.

**Limitations:**

* Scaling down the number of shards doesn't reshard existing data onto remaining instances. It must be manually moved (see also Scaling below).
* Scaling up the number of shards will not reshard existing data either. It will continue to be available from the same instances.

## Configuration

### Implementing a custom target distribution

To implement a custom distribution, set the `__tmp_hash` label during target discovery using relabeling configuration. The operator uses this label's value instead of the default labels when computing the shard assignment.

For example, to shard targets by pod namespace and name rather than by address:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
    relabelings:
    - sourceLabels: [__meta_kubernetes_pod_namespace, __meta_kubernetes_pod_name]
      separator: /
      targetLabel: __tmp_hash
```

The relabeling can also be applied at the scrape class level to affect multiple monitoring resources at once.

### Scraping a target from all the shards

By default, each target is assigned to exactly one shard. To have all shards scrape the same target — useful for singleton services such as kube-state-metrics where every shard needs the full set of metrics — set the `__tmp_disable_sharding` label to a non-empty value using relabeling configuration.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-state-metrics
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-state-metrics
  endpoints:
  - port: http
    relabelings:
    - targetLabel: __tmp_disable_sharding
      replacement: "true"
```

### Topology-aware sharding

> **Alpha:** Topology-aware sharding requires the `PrometheusTopologySharding` feature gate to be enabled on the operator.

In multi-zone clusters, the default address-based sharding distributes targets without regard for their zone, which can generate costly cross-zone traffic. Topology-aware sharding pins each Prometheus shard to a specific zone so that it only scrapes targets local to that zone.

When `mode: Topology` is set, the operator:
* Generates relabeling rules so each shard keeps only targets whose zone label matches its assigned zone.
* Automatically adds a `nodeSelector` to schedule each shard's pods in the correct zone.
* Adds a `zone` external label to each shard's Prometheus configuration (configurable via `externalLabelName`).

The number of shards must be greater than or equal to the number of topology values. When `spec.shards` is a multiple of the number of zones, the shards are evenly distributed across zones. Otherwise, some zones receive more shards than others.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  shards: 4
  replicas: 2
  shardingStrategy:
    mode: Topology
    topology:
      values:
        - europe-west4-a
        - europe-west4-b
```

With this configuration and 4 shards across 2 zones, shards 0 and 2 are scheduled in `europe-west4-a` and shards 1 and 3 in `europe-west4-b`. Each shard only scrapes targets in its zone.

### Retaining shards

> **Alpha:** Shard retention requires the `PrometheusShardRetentionPolicy` feature gate to be enabled on the operator.

When scaling down the number of shards, the pods from the removed shards are deleted by default along with access to their historical data. To preserve scaled-down shards so their data remains queryable until the retention duration expires, set `.spec.shardRetentionPolicy.whenScaled` to `Retain`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  shards: 2
  shardRetentionPolicy:
    whenScaled: Retain
```

Retained shards continue running and can be queried via the Thanos sidecar and querier alongside the active shards. By default, the operator deletes them once the Prometheus retention time has been reached. This can be overridden with the `retain.retentionPeriod` field:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  shards: 2
  retentionSize: 100Gi
  shardRetentionPolicy:
    whenScaled: Retain
    retain:
      retentionPeriod: 7d
```

> **Note:** If the Prometheus resource uses size-based retention only (no retention time configured), retained shards are kept forever by default.

## Example

The following manifest creates a Prometheus server with two replicas:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: default
spec:
  serviceAccountName: prometheus
  replicas: 2
  serviceMonitorSelector:
    matchLabels:
      team: frontend
```

This can be verified with the following command:

```bash
kubectl get pods -n default
```

The output is similar to this:

```bash
prometheus-prometheus-0                2/2     Running   1          10s
prometheus-prometheus-1                2/2     Running   1          10s
```

Deploy the example application and monitor it:

```yaml mdox-exec="cat example/shards/example-app-deployment.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      labels:
        app: example-app
    spec:
      containers:
      - name: example-app
        image: quay.io/brancz/prometheus-example-app:v0.5.0
        ports:
        - name: web
          containerPort: 8080
```

```yaml mdox-exec="cat example/shards/example-app-service.yaml"
kind: Service
apiVersion: v1
metadata:
  name: example-app
  namespace: default
  labels:
    app: example-app
spec:
  selector:
    app: example-app
  ports:
  - name: web
    port: 8080
```

```yaml mdox-exec="cat example/shards/example-app-service-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
  namespace: default
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
```

Explore one of the monitoring Prometheus instances:

```bash
kubectl port-forward pod/prometheus-prometheus-0 9090:9090
```

We find the prometheus server scrapes three targets.

Now let's expand the Prometheus resource to two shards as shown below:

```yaml mdox-exec="cat example/shards/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: shards
  name: prometheus
  namespace: default
spec:
  serviceAccountName: prometheus
  replicas: 2
  shards: 2
  serviceMonitorSelector:
    matchLabels:
      team: frontend
```

This can be verified with the following command:

```bash
kubectl get pods -n <namespace>
```

The output is similar to this:

```bash
prometheus-prometheus-0                2/2     Running   1          11m
prometheus-prometheus-1                2/2     Running   1          11m
prometheus-prometheus-shard-1-0        2/2     Running   1          12s
prometheus-prometheus-shard-1-1        2/2     Running   1          12s
```

Explore one of the monitoring Prometheus instances added for sharding:

```bash
kubectl port-forward prometheus-prometheus-shard-1-0 9091:9090
```

We should find that one or two targets are being scraped by this instance while the original Prometheus shard scrapes the remaining target(s).
