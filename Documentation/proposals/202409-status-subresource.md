# Status subresource for Prometheus operator Workload CRDs

* **Owners:**
  * [simonpasquier](https://github.com/simonpasquier)
* **Status:**
  * `Implemented`
* **Related Tickets:**
  * [#3335](https://github.com/prometheus-operator/prometheus-operator/issues/3335)
* **Other docs:**
  * N/A

This proposal describes how we will extend the Prometheus operator workload Custom
Resource Definitions (CRDs) with a Status subresource field.

## Why

Core Kubernetes resources differentiate between the desired state of an object
(the `spec` field) and the current status of the object (the `status` field)
[details](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status).
Before this proposal, the current status of the objects was never reflected by
the Prometheus operator which makes it harder for external actors to know if
the underlying resource is available or not.

### Pitfalls of the current solution

Though some of the custom resource definitions expose already a `Status`
subresource, the Prometheus operator never updates the status subresource and
it is only possible to retrieve the information via the custom API exposed by
the operator’s web service. This makes it harder than necessary for users to
know if the declared resources are ready and to understand why if they aren’t.

## Goals

* Define the structure of the status subresource for the custom resource
  definitions that materialize as Pod objects.
  * `Alertmanager`
  * `Prometheus`
  * `PrometheusAgent`
  * `ThanosRuler`
* Define how the operator would reconcile the status subresource.

## Non-goals

* Implement the status subresource for configuration objects like
  `ServiceMonitor`, `PodMonitor`, `PrometheusRule`, `Probe` and `ScrapeConfig`.
  * The main difficulty is that a `ServiceMonitor` object for instance can be
    reconciled by different objects. It brings more complexity in terms of API
    definition as well as implementation.
  * This will be addressed in a separate proposal.
* Extend the status subresource beyond what the operator can infer from the
  core Kubernetes API.
* Emit events on resource updates.
  * Once the operator implements status subresources, it seems a natural
    evolution to generate events on status changes but this isn't in the scope
    of this proposal.

## Background

The status subresource is a well-defined concept in Kubernetes:
* [Kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#status-subresource) about custom resource definitions.
* [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).
* [kubebuilder](https://book-v1.book.kubebuilder.io/basics/status_subresource.html) documentation.
* [OperatorSDK](https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/#manage-cr-status-conditions) documentation.

As much as possible, the proposal follows the recommendations established by the Kubernetes community.

## API

### Prometheus

The Prometheus CRD has a `Status` subresource that exposes the following fields:
* `Paused`
* `Replicas`
* `UpdatedReplicas`
* `AvailableReplicas`
* `UnavailableReplicas`

We propose to add the following new fields:
* `Conditions` as recommended by the document describing the [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).
* `ShardStatuses` which is a drilled-down status for each Prometheus shard.

```golang
type PrometheusStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this Prometheus deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this Prometheus deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Prometheus deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Prometheus deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The current state of the Prometheus deployment.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// The list has one entry per shard. Each entry provides a summary of the shard status.
	// +listType=map
	// +listMapKey=shardID
	// +optional
	ShardStatuses []ShardStatus `json:"shardStatuses,omitempty"`
	// Shards is the most recently observed number of shards.
	Shards int32 `json:"shards,omitempty"`
	// The selector used to match the pods targeted by this Prometheus resource.
	Selector string `json:"selector,omitempty"`
}


// Condition represents the state of the resources associated with the
// Prometheus, Alertmanager or ThanosRuler resource.
// +k8s:deepcopy-gen=true
type Condition struct {
	// Type of the condition being reported.
	// +required
	Type ConditionType `json:"type"`
	// Status of the condition.
	// +required
	Status ConditionStatus `json:"status"`
	// lastTransitionTime is the time of the last update to the current status property.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details for the condition's last transition.
	// +optional
	Message string `json:"message,omitempty"`
	// ObservedGeneration represents the .metadata.generation that the
	// condition was set based upon. For instance, if `.metadata.generation` is
	// currently 12, but the `.status.conditions[].observedGeneration` is 9, the
	// condition is out of date with respect to the current state of the
	// instance.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type ConditionType string

const (
	// Available indicates whether enough pods are ready to provide the
	// service.
	// The possible status values for this condition type are:
	// - True: all pods are running and ready, the service is fully available.
	// - Degraded: some pods aren't ready, the service is partially available.
	// - False: no pods are running, the service is totally unavailable.
	// - Unknown: the operator couldn't determine the condition status.
	Available ConditionType = "Available"
	// Reconciled indicates whether the operator has reconciled the state of
	// the underlying resources with the object's spec.
	// The possible status values for this condition type are:
	// - True: the reconciliation was successful.
	// - False: the reconciliation failed.
	// - Unknown: the operator couldn't determine the condition status.
	Reconciled ConditionType = "Reconciled"
)

type ConditionStatus string

const (
	ConditionTrue     ConditionStatus = "True"
	ConditionDegraded ConditionStatus = "Degraded"
	ConditionFalse    ConditionStatus = "False"
	ConditionUnknown  ConditionStatus = "Unknown"
)
```

Example of a Prometheus resource's status for which all pods are up and running:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: default
spec:
  replicas: 2
  shards: 2
status:
status:
  availableReplicas: 4
  conditions:
  - lastTransitionTime: "2024-09-10T14:24:26Z"
    message: ""
    observedGeneration: 4
    reason: ""
    status: "True"
    type: Available
  - lastTransitionTime: "2024-09-10T14:24:26Z"
    message: ""
    observedGeneration: 4
    reason: ""
    status: "True"
    type: Reconciled
  paused: false
  replicas: 4
  selector: app.kubernetes.io/instance=prometheus,app.kubernetes.io/managed-by=prometheus-operator,app.kubernetes.io/name=prometheus,operator.prometheus.io/name=prometheus,prometheus=prometheus
  shardStatuses:
  - availableReplicas: 2
    replicas: 2
    shardID: "0"
    unavailableReplicas: 0
    updatedReplicas: 2
  - availableReplicas: 2
    replicas: 2
    shardID: "1"
    unavailableReplicas: 0
    updatedReplicas: 2
  shards: 2
  unavailableReplicas: 0
  updatedReplicas: 4
```

Example of a Prometheus resource's status for which some pods are missing due to scheduling issues:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: monitoring
spec:
  replicas: 2
  shards: 2
status:
  availableReplicas: 2
  conditions:
  - lastTransitionTime: "2024-09-10T14:31:29Z"
    message: |-
      shard 0: pod prometheus-prometheus-1: 0/1 nodes are available: 1 node(s) didn't match pod anti-affinity rules. preemption: 0/1 nodes are available: 1 No preemption victims found for incoming pod.
      shard 1: pod prometheus-prometheus-shard-1-1: 0/1 nodes are available: 1 node(s) didn't match pod anti-affinity rules. preemption: 0/1 nodes are available: 1 No preemption victims found for incoming pod.
    observedGeneration: 5
    reason: SomePodsNotReady
    status: Degraded
    type: Available
  - lastTransitionTime: "2024-09-10T14:31:29Z"
    message: ""
    observedGeneration: 5
    reason: ""
    status: "True"
    type: Reconciled
  paused: false
  replicas: 4
  selector: app.kubernetes.io/instance=prometheus,app.kubernetes.io/managed-by=prometheus-operator,app.kubernetes.io/name=prometheus,operator.prometheus.io/name=prometheus,prometheus=prometheus
  shardStatuses:
  - availableReplicas: 1
    replicas: 2
    shardID: "0"
    unavailableReplicas: 1
    updatedReplicas: 1
  - availableReplicas: 1
    replicas: 2
    shardID: "1"
    unavailableReplicas: 1
    updatedReplicas: 1
  shards: 2
  unavailableReplicas: 2
  updatedReplicas: 2
```

### Alertmanager

The subresource status for the Alertmanager custom resource definition should
be very similar to the API defined for the Prometheus CRD. The main difference
is that we don't need the `shardStatuses` field.

```golang
type AlertmanagerStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// object (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// object that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Alertmanager cluster.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Alertmanager object.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The selector used to match the pods targeted by this Alertmanager object.
	Selector string `json:"selector,omitempty"`
	// The current state of the Alertmanager object.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}
```

### Thanos Ruler

The subresource status for the Thanos Ruler custom resource definition is
identical to the Alertmanager CRD.

```golang
type ThanosRulerStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this ThanosRuler deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this ThanosRuler deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The current state of the ThanosRuler object.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}
```

## Alternatives

N/A

## Action Plan

N/A
