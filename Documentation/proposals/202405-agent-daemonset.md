# DaemonSet deployment for Prometheus Agent

* Owners:
  * [haanhvu](https://github.com/haanhvu)
* Related Tickets:
  * [#5495](https://github.com/prometheus-operator/prometheus-operator/issues/5495)
* Other docs:
  * n/a

This proposal is about designing and implementing the deployment of Prometheus Agent as DaemonSet. Currently, Prometheus Agent can only be deployed as StatefulSet, which could be considered as “cluster-wide” strategy, meaning one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster. Though this works well for many use cases, some use cases may indeed prefer a “node-specific” strategy (DaemonSet), where each Prometheus Agent only scrapes the metrics from the targets located on the same node.


## 1. Why

When deploying Prometheus Agent in Kubernetes, two of the biggest users’ concerns are load distribution and scalability.

DaemonSet deployment is a good solution for these:
* Load distribution: Each Prometheus Agent pod will only scrape the targets located on the same node. Even though the targets on some nodes may produce more metrics than other nodes, the load distribution is reliable enough. This has been proven in [Google Cloud Managed Service for Prometheus (GMP)'s operator](https://github.com/GoogleCloudPlatform/prometheus-engine/) which follows a similar approach.
* Automatic scalability: When new nodes are added to the cluster, new Prometheus Agent pods will be automatically added in those nodes. Users can also select which set of nodes they want to deploy Prometheus Agent and the priority of Prometheus Agent pods compared to other pods on the same node.

DaemonSet deployment is especially more suitable for Prometheus Agent than Prometheus, since the Agent mode requires (currently around 20 to 30%) less memory, does not produce TSDB blocks on disks, and naturally blocks querying APIs.

This deployment mode has been implemented and proven in the Google Cloud Kubernetes Engine (GKE) with [Google Cloud Managed Service for Prometheus (GMP)'s operator](https://github.com/GoogleCloudPlatform/prometheus-engine/), so we can learn on their cases and collaborate on shared improvements together.


## 2. Pitfalls of the current solution

The key pitfall of managing load distribution and scalability with the current StatefulSet deployment is they are done on the cluster scope. In other words, since one or several high-availability Prometheus Agents are responsible for scraping metrics of the whole cluster, users need to calculate/estimate the load and scalability of the whole cluster to decide on replicas and sharding strategies. Estimating cluster-wide load and scalability is a much harder task than estimating node-wide load and scalability. Even though users can use helping tools like Horizontal Pod Autoscaler (HPA), that's still additional complexity.

We're not saying that DaemonSet is superior to StatefulSet. In fact, StatefulSet has its own advantages, such as easier storage handling. What we're trying to say is DaemonSet can solve some of the existing problems in StatefulSet, and vice versa. So DaemonSet is a good deployment option to implement besides StatefulSet.


## 3. Audience

Users with use cases where scraped load is very large or hard to estimate and/or scalability is hard to predict, so they need an easy way to manage the load distribution and scalability.


## 4. Goals

Provide an MVP version of the DaemonSet deployment of Prometheus Agent to the Audience.
In specific, the MVP will need to:
* Allow users to deploy one Prometheus Agent pod per node.
* Allow users to restrict which set of nodes they want to deploy Prometheus Agent, if desired.
* Allow users to set the priority of Prometheus Agent pod compared to other pods on the same node, if desired.
* Allow each Prometheus Agent pod to only scrape from the pods from PodMonitor that run on the same node as the Agent.


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

We can add a new `mode` field that accepts either `StatefulSet` or `DaemonSet`, with `StatefulSet` being the default. If the DaemonSet mode is activated (`mode: DaemonSet`), all the unrelated fields listed above will not be accepted. In the MVP, we will simply fail the reconciliation if any of those fields are set. Following up, we will leverage validation rules with [Kubernetes' Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/).

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
* Prometheus Agent DaemonSet selects correctly the pods from PodMonitor on the same node.
Currently we only set up a Kind cluster of one node for integration tests. Since the test cases for DaemonSet deployment requires at least two nodes, we will need to modify the Kind cluster config for that.

We’ll also need a new user guide explaining how to use this new mode.


## 8. Follow-ups

After the Goals of this proposal have been met, we’ll handle what’s stated in the Non-goals section.
We will also work on enhancements, such as leveraging validation rules with [Kubernetes' Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/) for the fields in Prometheus Agent CRD for DaemonSet mode.
