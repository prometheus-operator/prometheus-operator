---
weight: 521
toc: true
title: DaemonSet deployment for Prometheus Agent
menu:
    docs:
        parent: proposals
lead: ""
images: []
draft: false
---

* Owners:
  * [haanhvu](https://github.com/haanhvu)
  * [slashexx](https://github.com/slashexx)
* Status:
  * `Accepted`
* Related Tickets:
  * [#5495](https://github.com/prometheus-operator/prometheus-operator/issues/5495)
* Other docs:
  * n/a

This proposal is about designing and implementing the deployment of Prometheus Agent as DaemonSet. Currently, Prometheus Agent can only be deployed as StatefulSet, which could be considered as “cluster-wide” strategy, meaning one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster. Though this works well for many use cases, some use cases may indeed prefer a “node-specific” strategy (DaemonSet), where Prometheus Agent pods scale with the nodes and only scrape the metrics from the targets located on the same node.

## 1. Why

When deploying Prometheus Agent in Kubernetes, two of the biggest users’ concerns are load distribution and scalability.

DaemonSet deployment is a good solution for these:
* Load distribution: Each Prometheus Agent pod will only scrape the targets located on the same node. Even though the targets on some nodes may produce more metrics than other nodes, the load distribution is reliable enough. This has been proven in [Google Cloud Managed Service for Prometheus (GMP)'s operator](https://github.com/GoogleCloudPlatform/prometheus-engine/) which follows a similar approach.
* Automatic scalability: When new nodes are added to the cluster, new Prometheus Agent pods will be automatically deployed to those nodes. Similarly, at some durations when some nodes are not needed and removed from the cluster, Prometheus Agent pods on those nodes will also be removed. Users can also select which set of nodes they want to deploy Prometheus Agent and the priority of Prometheus Agent pods compared to other pods on the same node.

DaemonSet deployment is especially more suitable for Prometheus Agent than Prometheus because the Agent mode is customized for collect-and-forward approach, so it's more lightweight. In specific, it requires (currently around 20 to 30%) less memory, does not produce TSDB blocks on disks, and naturally blocks querying APIs and rule distribution.

This deployment mode has been implemented and proven in the Google Cloud Kubernetes Engine (GKE) with [Google Cloud Managed Service for Prometheus (GMP)'s operator](https://github.com/GoogleCloudPlatform/prometheus-engine/), so we can learn from their cases and collaborate for shared improvements.

## 2. Pitfalls of the current solution

The key pitfalls of managing load distribution and scalability with the current StatefulSet deployment are:
* They are done on the cluster scope. In other words, since one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster, users need to calculate/estimate the load and scalability of the whole cluster to decide on replicas and sharding strategies. Estimating cluster-wide load and scalability is a much harder task than estimating node-wide load and scalability.
* StatefulSet inherently doesn't scale alongside the scale of nodes. So there may be situations where more or less existing Prometheus Agent pods are needed for the scraped load. Even though users can use helping tools like Horizontal Pod Autoscaler (HPA), that's still additional complexity.

This is not to say that DaemonSet is superior to StatefulSet. StatefulSet also has its own advantages, such as easier storage handling. This is to say that DaemonSet can solve some of the existing problems in StatefulSet, and vice versa. So DaemonSet is a good deployment option to implement besides StatefulSet. Users can choose to use one or both of them, depending on their cases.

## 3. Audience

Users with use cases where scraped load is very large or hard to estimate and/or scalability is hard to predict, so they need a simple way to manage load distribution and scalability. They may apply DaemonSet deployment on the whole cluster or only some nodes according to their needs.

An example of audience is expressed [here](https://github.com/prometheus-operator/prometheus-operator/issues/5495#issuecomment-1519812510).

## 4. Goals

Provide an MVP version of the DaemonSet deployment of Prometheus Agent to the Audience.
In specific, the MVP will need to:
* Allow users to deploy one Prometheus Agent pod per node.
* Allow users to restrict which set of nodes they want to deploy Prometheus Agent, if desired.
* Allow users to set the priority of Prometheus Agent pod compared to other pods on the same node, if desired.
* Allow each Prometheus Agent pod to only scrape from the pods from PodMonitor that run on the same node.

## 5. Non-Goals

This proposal only aims at Prometheus Agent, not Prometheus.

Other non-goals are the features that are not easy to implement and require more investigation. We will need to investigate whether there are actual user needs for them, if yes, then how to best implement them. We can also learn from similar projects such as OpenTelemetry Operator (they have DaemonSet mode) and Grafana Agent on how they approach these problems and what we can apply for our cases. We’ll handle these after the MVP. Those (currently) non-goals features are:
* ServiceMonitor support: There's a performance issue regarding this feature. Since each Prometheus Agent running on a node requires one watch, making all Prometheus Agent pods watch all endpoints will put a huge stress on Kubernetes API server. This is the main reason why GMP hasn’t supported this, even though there are user needs stated in some issues ([#362](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/362), [#192](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/192)). However, as discussed with Danny from GMP [here](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/192#issuecomment-2028850846), ServiceMonitor support based on EndpointSlice seems like a viable approach. We’ll investigate this further after the MVP.
* Storage: We will need to spend time studying more about the WAL, different storage solutions provided by Kubernetes, and how to gracefully handle storage in different cases of crashes. For example, there’s an [issue in Prometheus](https://github.com/prometheus/prometheus/issues/8809) showing that samples may be lost if remote write didn’t flush cleanly. We’ll investigate these further after the MVP.

In the MVP version, we will not allow users to directly switch from a live StatefulSet to DaemonSet deployment. Reasons are explained in the CRD subsection in the How section.

## 6. How

The MVP version of DaemonSet deployment will be put behind a feature flag.

### 6.1. CRD:

Currently, we already have a PrometheusAgent CRD that supports StatefulSet deployment. We’ll add new field(s) in this CRD to enable DaemonSet deployment.

The reason for enhancing existing CRD (instead of introducing a new CRD) is it would take less time to finish the MVP. We’ll let users experiment with the MVP, and in case users report a separate CRD is needed, we’ll separate the logic of DaemonSet deployment into a new CRD later.

The current [PrometheusAgent CRD](https://prometheus-operator.dev/docs/platform/prometheus-agent/) already has sufficient fields for the DaemonSet deployment. The DaemonSet deployment can use all the existing fields in the CRD except the ones related to:
* Selectors for service, probe, ScrapeConfig
* Replica
* Shard
* Storage

We will add a new `mode` field that accepts either `StatefulSet` or `DaemonSet`, with `StatefulSet` being the default. If the DaemonSet mode is activated (`mode: DaemonSet`), all the unrelated fields listed above will not be accepted. In the MVP, we will simply fail the reconciliation if any of those fields are set. We will prevent users to directly switch from a live StatefulSet setup to DaemonSet, because that might break their workload if they forget to unset the unsupported fields. Following up, we will leverage validation rules with [Kubernetes' Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/). Only then, we will allow switching from a live StatefulSet setup to DaemonSet. We already have an issue for CEL [here](https://github.com/prometheus-operator/prometheus-operator/issues/5079).

#### 6.1.1 CEL Validation rules

When `mode:DaemonSet`, the following CEL rules will be applied:

- `replicas`
- `storage`
- `shards`
- `persistentVolumeClaimRetentionPolicy`
- `scrapeConfigSelector`
- `scrapeConfigNamespaceSelector`
- `probeSelector`
- `probeNamespaceSelector`
- `serviceMonitorSelector`
- `serviceMonitorNamespaceSelector`
- `additionalScrapeConfigs`

This is implemented by adding `x-kubernetes-validations` like:

```yaml
x-kubernetes-validations:
  - rule: "self.mode == 'DaemonSet' ? !has(self.replicas) : true"
    message: "replicas field is not allowed when mode is 'DaemonSet'"
  - rule: "self.mode == 'DaemonSet' ? !has(self.storage) : true"
    message: "storage field is not allowed when mode is 'DaemonSet'"
  - rule: "self.mode == 'DaemonSet' ? !has(self.shards) : true"
    message: "shards field is not allowed when mode is 'DaemonSet'"
  - rule: "self.mode == 'DaemonSet' ? !has(self.persistentVolumeClaimRetentionPolicy) : true"
    message: "persistentVolumeClaimRetentionPolicy field is not allowed when mode is 'DaemonSet'"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.scrapeConfigSelector))"
    message: "scrapeConfigSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.probeSelector))"
    message: "probeSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.probeNamespaceSelector))"
    message: "probeNamespaceSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.scrapeConfigNamespaceSelector))"
    message: "scrapeConfigNamespaceSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.serviceMonitorSelector))"
    message: "serviceMonitorSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.serviceMonitorNamespaceSelector))"
    message: "serviceMonitorNamespaceSelector cannot be set when mode is DaemonSet"
  - rule: "!(has(self.mode) && self.mode == 'DaemonSet' && has(self.additionalScrapeConfigs))"
    message: "additionalScrapeConfigs cannot be set when mode is DaemonSet"
```

#### 6.1.2 Runtime Validation Logic as Fallback

CEL validation will provide immediate feedback during `kubectl apply` but we will need runtime validation logic in the controller as a fallback mechanism. This fallback will be integrated directly in the `PrometheusAgent` reconciler loop.

This is mainly because :
1. CEL validation will require Kubernetes version 1.25+ and hence not all users might have CEL supported clusters.
2. This will provide an in-depth defense mechamnism against misconfigurations.
3. More detailed error response in case the first layer of defense fails.

```go
if spec.Mode == "DaemonSet" {
	if spec.Replicas != nil {
		return fmt.Errorf("cannot configure replicas when using DaemonSet mode")
	}
	if spec.Storage != nil {
		return fmt.Errorf("cannot configure storage when using DaemonSet mode")
	}
	if spec.Shards != nil {
		return fmt.Errorf("cannot configure shards when using DaemonSet mode")
	}
	if spec.PersistentVolumeClaimRetentionPolicy != nil {
		return fmt.Errorf("cannot configure persistentVolumeClaimRetentionPolicy when using DaemonSet mode")
	}
	if spec.ScrapeConfigSelector != nil {
		return fmt.Errorf("cannot configure scrapeConfigSelector when using DaemonSet mode")
	}
	if spec.ProbeSelector != nil {
		return fmt.Errorf("cannot configure probeSelector when using DaemonSet mode")
	}
	if spec.ProbeNamespaceSelector != nil {
		return fmt.Errorf("cannot configure probeNamespaceSelector when using DaemonSet mode")
	}
	if spec.ScrapeConfigNamespaceSelector != nil {
		return fmt.Errorf("cannot configure scrapeConfigNamespaceSelector when using DaemonSet mode")
	}
	if spec.ServiceMonitorSelector != nil {
		return fmt.Errorf("cannot configure serviceMonitorSelector when using DaemonSet mode")
	}
	if spec.ServiceMonitorNamespaceSelector != nil {
		return fmt.Errorf("cannot configure serviceMonitorNamespaceSelector when using DaemonSet mode")
	}
	if spec.AdditionalScrapeConfigs != nil {
		return fmt.Errorf("cannot configure additionalScrapeConfigs when using DaemonSet mode")
	}
}
```

### 6.2. Node detecting:

As pointed out in [Danny from GMP’s talk](https://www.youtube.com/watch?v=yk2aaAyxgKw), to make Prometheus Agent DaemonSet know which node it’s on, we can use [Kubernetes’ downward API](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/). In `config-reloader` container, we can mount the node name as an environment variable like this:

```
containers:
- name: config-reloader
  env:
  - name: NODE_NAME
    valueFrom:
      fieldRef:
        apiVersion: v1
        fieldPath: spec.nodeName
```

### 6.3. Targets filtering for pods (PodMonitor support):

To filter pod targets, Danny’s talk has pointed out the best option is to use field selector:

```
kubernetes_sd_configs:
- role: pod
  selectors:
  - role: pod
    field: spec.nodeName=$NODE_NAME
```

We'll go with this option, because it filters targets right at discovery time, and also because Kubernetes API server watch cache indexes pods by node name (as we can see in [Kubernetes codebase](https://github.com/kubernetes/kubernetes/blob/v1.30.0-rc.0/pkg/registry/core/pod/storage/storage.go#L91)).

We've also considered using relabel config that filters pods by `__meta_kubernetes_pod_node_name` label. However, we didn't choose to go with this option because it filters pods only after discovering all the pods from PodMonitor, which increases load on Kubernetes API server.

## Secondary/Extended goal (new feature gate)

> **Note:** We are exploring the integration of ServiceMonitor support for DaemonSet mode using EndpointSlice as an experimental feature. This exploration will determine feasibility and performance, and if viable, it may be introduced behind a separate feature gate. This approach allows the main DaemonSet mode to reach GA independently of this feature.

### ServiceMonitor Support with EndpointSlice

To enable ServiceMonitor support for DaemonSet mode while addressing the performance concerns mentioned in section 5, we implement EndpointSlice-based service discovery:

#### EndpointSlice Discovery Implementation

The PrometheusAgent CRD already supports a `serviceDiscoveryRole` field that can be set to `EndpointSlice`:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: PrometheusAgent
spec:
  mode: DaemonSet
  serviceDiscoveryRole: EndpointSlice  # Use EndpointSlice instead of Endpoints
  serviceMonitorSelector:
    matchLabels:
      team: platform
```

When `serviceDiscoveryRole: EndpointSlice` is specified, the generated Prometheus configuration will use:

```yaml
scrape_configs:
- job_name: serviceMonitor/default/my-service/0
  kubernetes_sd_configs:
  - role: endpointslice  # Instead of "endpoints"
    namespaces:
      names: [default]
```

#### Performance Benefits

EndpointSlice provides significant performance improvements over classic Endpoints:
* **Scalability**: EndpointSlice objects are limited to 1000 endpoints each, preventing massive objects
* **Efficiency**: Multiple smaller objects reduce memory usage and network traffic
* **Parallel Processing**: Multiple EndpointSlice objects can be processed in parallel
* **Reduced API Server Load**: Less stress on Kubernetes API server with distributed endpoint information

#### Implementation Details

The implementation properly handles EndpointSlice support by checking both the user's `serviceDiscoveryRole` setting and cluster compatibility. The logic involves:

```go
// Check if THIS PrometheusAgent wants EndpointSlice discovery
cpf := p.GetCommonPrometheusFields()
if ptr.Deref(cpf.ServiceDiscoveryRole, monitoringv1.EndpointsRole) == monitoringv1.EndpointSliceRole {
	if c.endpointSliceSupported {
		opts = append(opts, prompkg.WithEndpointSliceSupport())
	} else {
		// Warn user that they want EndpointSlice but cluster doesn't support it
		c.logger.Warn("EndpointSlice requested but not supported by Kubernetes cluster")
		// Fall back to classic endpoints
	}
}
```

## 7. Action Plan

For the implementation, we’ll do what we detailed in the How section. The common logics between StatefulSet and DaemonSet modes will be extracted into one place. We will have a separate `daemonset.go` for the separate logic of the DaemonSet mode.

For the test, we will have unit tests covering new logic, and integration tests covering the basic user cases, which are:
* Users cannot switch directly from StatefulSet to DaemonSet.
* Prometheus Agent DaemonSet is created/deleted successfully.
* Prometheus Agent DaemonSet is installed on the right nodes.
* Prometheus Agent DaemonSet selects correctly the pods from PodMonitor on the same node.
  Currently we only set up a Kind cluster of one node for integration tests. Since the test cases for DaemonSet deployment requires at least two nodes, we will need to modify the Kind cluster config for that.

We’ll also need a new user guide explaining how to use this new mode.

## 8. Follow-ups

After the Goals of this proposal have been met, we’ll reevaluate the features in the Non-goals section and see if any of them should/can be addressed.

We will also work on enhancements, such as leveraging validation rules with [Kubernetes' Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/) for the fields in Prometheus Agent CRD for DaemonSet mode.
