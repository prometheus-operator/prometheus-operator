# DaemonSet deployment for Prometheus Agent

* Owners:
  * [haanhvu](https://github.com/haanhvu)
* Related Tickets:
  * [#5495](https://github.com/prometheus-operator/prometheus-operator/issues/5495)
* Other docs:
  * n/a

This proposal is about designing and implementing the deployment of Prometheus Agent as DaemonSet. Currently, Prometheus Agent can only be deployed as StatefulSet, which could be considered as “cluster-wide” strategy, meaning one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster. Though this works well for many use cases, some use cases may indeed prefer a “node-specific” strategy (DaemonSet), where each Prometheus Agent only scrapes the metrics from the targets located on the same node.


## 1. Why

When deploying Prometheus Agent in Kubernetes, three of the biggest users’ concerns are: load distribution, scalability, and security.

DaemonSet deployment solves all these three concerns:
* Load distribution: Each Prometheus Agent pod will only scrape the targets located on the same node. Even though the targets on some nodes may produce more metrics than other nodes, the load distribution would be reliable enough.
* Automatic scalability: When new nodes are added to the cluster, new Prometheus Agent pods will be automatically added in the nodes that meet user-defined restrictions (if any).
* Security: Since the scraped targets are local to the Prometheus Agent pod (on the same node), the scope of security problems is reduced to each node.

DaemonSet deployment is especially more suitable for Prometheus Agent than Prometheus, since the storage requirement for the Agent mode is pretty light compared to a fully functional Prometheus server.

This deployment mode has been implemented in [Google Cloud Managed Service for Prometheus (GMP)](https://github.com/GoogleCloudPlatform/prometheus-engine/), so we have an implementation example to learn from. Also, since this deployment has been tested and proved with GMP’s userbase, we can count on that it can solve a large enough number of use cases.


## 2. Pitfalls of the current solution

The current (StatefulSet) deployment brings long the corresponding pitfalls:
* Load management & scalability: Since one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster, users would need to calculate/estimate the load and scalability of the whole cluster to decide on replicas and sharding strategies. Estimating cluster-wide load and scalability is a much harder task than estimating node-wide load and scalability.
* Security: Similarly, cluster-wide security is a much bigger problem than node-wide security.

At the moment, the proposed DaemonSet deployment mode doesn’t mean to be a replacement for the current StatefulSet mode. It’s actually a solution for use cases where StatefulSet may not be the best choice. We’ll keep the StatefulSet mode as long as there’s user need for it.


## 3. Audience

Users with use cases where:
* Scraped load is very large or hard to estimate.
* Scalability is hard to predict.
* Security is a big concern.
* They want to collect node system metrics (e.g. kubelet, node exporter).


## 4. Goals

Provide an MVP version of the DaemonSet deployment of Prometheus Agent to the audience.
In specific, the MVP will need to:
* Allow users to deploy one Prometheus Agent pod per node
* Allow users to restrict on which set of nodes they want to deploy Prometheus Agent
* Allow each Prometheus Agent pod to scrape from the pods on the same node (PodMonitor support)


## 5. Non-Goals

The non-goals are the features that are not easy to implement and require more investigation. We will need to investigate whether there are actual user needs for them, if yes, then how to best implement them. We’ll handle these after the MVP.
* ServiceMonitor support: There's a performance issue regarding this feature. Since each Prometheus Agent running on a node requires one watch, making all Prometheus Agent pods watch all endpoints will put a huge stress on Kubernetes API server. This is the main reason why GMP hasn’t supported this, even though there are user needs stated in some issues ([#362](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/362), [#192](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/192)). However, as discussed with Danny from GMP [here](https://github.com/GoogleCloudPlatform/prometheus-engine/issues/192#issuecomment-2028850846), ServiceMonitor support based on EndpointSlice seems like a viable approach. We’ll investigate this further after the MVP.
* Storage: We will need to spend time studying more about the WAL, different storage solutions provided by Kubernetes, and how to gracefully handle storage in different cases of crashes. For example, there’s an [issue in Prometheus](https://github.com/prometheus/prometheus/issues/8809) showing that samples may be lost if remote write didn’t flush cleanly. We’ll investigate these further after the MVP.
* Secret management: We’ll need to figure out how to make each Prometheus Agent pod only watch the secrets from the pods on the same node. This is for both enhancing security and reducing load on Kubernetes API server. However, this is not easy to do. So we’ll handle this after the MVP.


## 6. How

### 6.1. CRD:

Currently, we already have a Prometheus Agent CRD that supports StatefulSet deployment. We’ll add new field(s) in this CRD to enable DaemonSet deployment.

The reason for enhancing existing CRD (instead of introducing a new CRD) is it would take less time to finish the MVP. We’ll let users experiment with the MVP, and in case users report a separate CRD is needed, we’ll separate the logic of DaemonSet deployment into a new CRD later.

The current [Prometheus Agent CRD](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1alpha1.PrometheusAgent) already has sufficient fields for the DaemonSet deployment. The DaemonSet deployment can use all the existing fields in the CRD except the ones related to:
* Selectors for service, probe, ScrapeConfig
* Replica
* Shard
* Storage
We can add a new bool field `daemonset`, which is defaulted to `false`. If the DaemonSet deployment is activated (`daemonset: true`), all the fields except the unrelated ones listed above will be used for the DaemonSet mode.

### 6.2. Node detecting:

As pointed out in [Danny from GMP’s talk](https://www.youtube.com/watch?v=yk2aaAyxgKw), to make Prometheus Agent DaemonSet know which node it’s on, we can use [Kubernetes’ downward API](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/). In `prometheus` or `config-reloader` container, we can mount the node name as an environment variable like this:
containers:
```
 - name: prometheus
   image: prom/prometheus
   env:
   - name: NODE_NAME
     valueFrom:
     	fieldRef:
     		apiVersion: v1
     		fieldPath: spec.nodeName
```

### 6.3. Targets filtering for pods (PodMonitor support):

To filter targets, Danny’s talk has pointed out two options.
The first option is to use relabel config. For example, to filter pods we could generate:
```
relabel_configs:
 - source_labels: [__meta_kubernetes_pod_node_name]
   separator: ;
   regex: $NODE_NAME
   replacement: $1
   action: keep
```
This could be considered as “last-mile filtering”. In other words, we watch for all the changes, and then filter the changes that match.

The second option is to use field selector. For example, to select pods on a specific node, we can do:
```
kubernetes_sd_configs:
   - role: pod
 	selectors:
 	- role: pod
   	  field: spec.nodeName=$NODE_NAME
```
This option performs better, because we filter targets right at discovery time, and also because Kubernetes API server watch cache indexes pods by node name (as we can see in [Kubernetes codebase](https://github.com/kubernetes/kubernetes/blob/v1.30.0-rc.0/pkg/registry/core/pod/storage/storage.go#L91)). So we’ll go with this option.


## 7. Action Plan

For the implementation, we’ll do what we detailed in the How section. The common logics between StatefulSet and DaemonSet modes will be extracted into one place. We will have a separate `daemonset.go` for the separate logic of the DaemonSet mode.

For the test, we will have unit tests covering new logic, and integration tests covering the basic user cases, which are:
* Prometheus Agent DaemonSet is created/deleted successfully.
* Prometheus Agent DaemonSet is installed on the right nodes.
* Prometheus Agent DaemonSet scales up/down following the scale of nodes.
* Prometheus Agent DaemonSet selects correctly the pods on the same node.

We’ll also need a new user guide explaining how to use this new mode.


## 8. Follow-ups

After the goals of this proposal have been met, we’ll handle what’s stated in the Non-goals section.
