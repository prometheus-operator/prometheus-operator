# Status subresource for Prometheus operator CRDs

## Summary

Though some of the custom resource definitions expose already a `Status`
subresource, the Prometheus operator never updates the status subresource and
it is only possible to retrieve the information via the custom API exposed by
the operator’s web service. This makes it harder than necessary for users to
know if the declared resources are ready and to understand why if they aren’t.

## Goals

* Define the structure of the status subresource for the custom resource definitions managed by the Prometheus operator.
* Define how the operator would reconcile the status subresource.

## Non-goals

* Extend the status subresource beyond what the operator can infer from the core Kubernetes resources (e.g. we don't want to correlate the service/pod monitors with the status of the targets as reported by Prometheus, at least for now). This would require a back-channel communication from the operands to the operator which doesn't exist today.
* Emit events on resource updates. Once the operator implements status subresources, it seems a natural evolution to generate events on status changes but this isn't in the scope of this proposal.

## Background

The status subresource is a well-defined concept in Kubernetes:
* [Kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#status-subresource) about custom resource definitions.
* [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).
* [kubebuilder](https://book-v1.book.kubebuilder.io/basics/status_subresource.html) documentation.
* [OperatorSDK](https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/#manage-cr-status-conditions) documentation.

As much as possible, the proposal follows the recommendations established by the Kubernetes community.

The feature is tracked in [#3335](https://github.com/prometheus-operator/prometheus-operator/issues/3335).

## API

### Prometheus

The Prometheus CRD has a `Status` subresource that exposes the following fields:
* `Paused`
* `Replicas`
* `UpdatedReplicas`
* `AvailableReplicas`
* `UnavailableReplicas`

We propose to add new fields:
* `Conditions` as recommended by the document describing the [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).
* `ShardStatuses` which is a drilled-down status for each Prometheus shard.
* [TBD] Number of resources (service monitors, pod monitors, probes and Prometheus rules) selected/rejected by the operator.

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
	// The current state of the Prometheus service.
	// +patchMergeKey=type
	// +patchMergeStrategy=merge
	// +optional
	Conditions []PrometheusCondition `json:"conditions,omitempty"`
	// The list has one entry per shard. Each entry provides a summary of the shard status.
	// +patchMergeKey=shardID
	// +patchMergeStrategy=merge
	// +optional
	ShardStatuses []ShardStatus `json:"shardStatuses,omitempty"`
}

// PrometheusCondition represents the state of the resources associated with the Prometheus resource.
// +k8s:deepcopy-gen=true
type PrometheusCondition struct {
	// Type of the condition being reported.
	// +required
	Type PrometheusConditionType `json:"type"`
	// status of the condition.
	// +required
	Status PrometheusConditionStatus `json:"status"`
	// lastTransitionTime is the time of the last update to the current status property.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason for the condition's last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// Human-readable message indicating details for the condition's last transition.
	// +optional
	Message *string `json:"reason,omitempty"`
}

type PrometheusConditionType string

const (
	// Available indicates that enough Prometheus pods are ready to provide the service.
	PrometheusAvailable PrometheusConditionType = "Available"
	// Degraded indicates that some Prometheus pods don't run as expected.
	PrometheusDegraded PrometheusConditionType = "Degraded"
	// Reconciled indicates that the operator has reconciled the state of the underlying resources with the Prometheus object spec.
	PrometheusReconciled PrometheusConditionType = "Reconciled"
)

type PrometheusConditionStatus string

const (
	PrometheusConditionTrue    PrometheusConditionStatus = "True"
	PrometheusConditionFalse   PrometheusConditionStatus = "False"
	PrometheusConditionUnknown PrometheusConditionStatus = "Unknown"
)

type ShardStatus struct {
	// Identifier of the shard.
	// +required
	ShardID string `json:"shardID"`
	// Total number of pods targeted by this shard.
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this shard
	// that have the desired spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this shard.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this shard.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}
```

Example of a Prometheus resource's status for which all pods are up and running:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: monitoring
spec:
  paused: false
  replicas: 2
  shards: 2
status:
  paused: false
  replicas: 4
  updatedReplicas: 4
  availableReplicas: 4
  unavailableReplicas: 0
  conditions:
  - type: Available
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  - type: Degraded
    status: "False"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  - type: Reconciled
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  shardStatuses:
  - shardID: "0"
    replicas: 2
    updatedReplicas: 2
    availableReplicas: 2
    unavailableReplicas: 0
  - shardID: "1"
    replicas: 2
    updatedReplicas: 2
    availableReplicas: 2
    unavailableReplicas: 0
```

Example of a Prometheus resource's status for which some pods are missing due to scheduling issues:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: monitoring
spec:
  paused: false
  replicas: 2
  shards: 2
status:
  paused: false
  replicas: 4
  updatedReplicas: 4
  availableReplicas: 2
  unavailableReplicas: 2
  conditions:
  - type: Available
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  - type: Degraded
    status: "True"
    reason: "PodsNotReady"
    message: |-
      Some pods are not ready:
      prometheus-monitoring-shard-0-1: 0/6 nodes are available: 2 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate, 2 node(s) had volume node affinity conflict, 2 node(s) were unschedulable.
      prometheus-monitoring-shard-1-1: 0/6 nodes are available: 2 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate, 2 node(s) had volume node affinity conflict, 2 node(s) were unschedulable.
    lastTransitionTime: "2022-02-08T23:54:22Z"
  - type: Reconciled
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  shardStatuses:
  - shardID: "0"
    replicas: 2
    updatedReplicas: 2
    availableReplicas: 1
    unavailableReplicas: 1
  - shardID: "1"
    replicas: 2
    updatedReplicas: 2
    availableReplicas: 1
    unavailableReplicas: 1
```

### Alertmanager

The subresource status for the Alertmanager custom resource definition should
be very similar to the API defined for the Prometheus CRD. The main difference
is that we don't need the `shardStatuses` field.

### Thanos Ruler

The subresource status for the Thanos Ruler custom resource definition is
identical to the Alertmanager CRD.

### Service Monitor

```golang
type ServiceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Service selection for target discovery by
	// Prometheus.
	Spec ServiceMonitorSpec `json:"spec"`
	// 
	// Most recent observed status of the service monitor.
	// Populated by the Prometheus operator.
	// Read-only.
	Status ServiceMonitorStatus `json:"status"`
}

type ServiceMonitorStatus struct {
	// The current state of the service monitor.
	// +patchMergeKey=type
	// +patchMergeStrategy=merge
	// +optional
	Conditions []ServiceMonitorCondition `json:"conditions,omitempty"`

	// The list of resources that the service monitor is bound to.
	// +patchMergeKey=name
	// +patchMergeStrategy=merge
	// +optional
	Bindings []ServiceMonitorBindings `json:"bindings,omitempty"`
}

type ServiceMonitorCondition struct {
	// Type of the condition being reported.
	// +required
	Type ServiceMonitorConditionType `json:"type"`
	// status of the condition.
	// +required
	Type ServiceMonitorConditionStatus `json:"status"`
	// lastTransitionTime is the time of the last update to the current status property.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason for the condition's last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// Human-readable message indicating details for the condition's last transition.
	// +optional
	Message *string `json:"reason,omitempty"`
}

type ServiceMonitorConditionType string

const (
	// Valid indicates if the operator considers the service monitor to be valid.
	ServiceMonitorValid ServiceMonitorConditionType = "Valid"
	// Selected indicates if the service monitor is reconciled by the referenced Prometheus object.
	ServiceMonitorReconciled ServiceMonitorConditionType = "Reconciled"
)

type ServiceMonitorConditionStatus string

const (
	ServiceMonitorConditionTrue    ServiceMonitorConditionStatus = "True"
	ServiceMonitorConditionFalse   ServiceMonitorConditionStatus = "False"
	ServiceMonitorConditionUnknown ServiceMonitorConditionStatus = "Unknown"
)

type ServiceMonitorBinding struct {
	// Name of the binding.
	// +required
	Name string `json:"name"`
	// Reference to the Prometheus object that binds the service monitoring (e.g. Prometheus or PrometheusAgent).
	// +required
	PrometheusRef PrometheusReference `json:"prometheusRef"`
	// The current state of the service monitor when bound to the referenced Prometheus object.
	// +patchMergeKey=type
	// +patchMergeStrategy=merge
	// +optional
	Conditions []ServiceMonitorCondition `json:"conditions,omitempty"`
}

type PrometheusReference struct {
	// The type of resource being referenced (e.g. Prometheus or PrometheusAgent).
	// +kubebuilder:validation:Enum=prometheuses;prometheusagents
	// +required
	Resource string
	// The name of the referenced object.
	// +required
	Name string
	// The namespace of the referenced object.
	// +required
	Namespace string
}
```

Example for a valid ServiceMonitor bound to one Prometheus object:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
status:
  conditions:
  - type: Valid
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  bindings:
  - name: default/prometheus-agent-monitoring
    prometheusRef:
      resource: prometheusagents
      name: monitoring
      namespace: default
    conditions:
    - type: Reconciled
      status: "True"
      lastTransitionTime: "2022-02-08T23:54:22Z"
```

Example for an invalid ServiceMonitor (the secret reference doesn't exist):

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
    bearerTokenSecret:
      name: metric-credentials
      key: bearer-token
status:
  conditions:
  - type: Valid
    status: "False"
    lastTransitionTime: "2022-02-08T23:54:22Z"
    reason: InvalidSecretReference
    message: "failed to get bearer token: failed to get token from secret: key 'bearer-token' in secret 'metric-credentials' not found"
```

Example for a ServiceMonitor that can't be bound to a Prometheus object:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
    bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
status:
  conditions:
  - type: Valid
    status: "True"
    lastTransitionTime: "2022-02-08T23:54:22Z"
  bindings:
  - name: default/prometheus-agent-monitoring
    prometheusRef:
      resource: prometheusagents
      name: monitoring
      namespace: default
    conditions:
    - type: Reconciled
      status: "False"
      lastTransitionTime: "2022-02-08T23:54:22Z"
      reason: DirectFileSystemAccessForbidden
      message: "it accesses file system via bearer token file which Prometheus specification prohibits"
```

A given ServiceMonitor object may be selected by more than one
Prometheus/PrometheusAgent resource which explains the presence of the
`bindings` field. As explained in the [Kubernetes API
conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#object-references), this creates additional complexity that need to be clarified:
* The operator needs update permissions on the ServiceMonitor's status subresource.
* When the referenced Prometheus resource stops selecting a given Service Monitor (`spec.serviceMonitorSelector` has changed or the ServiceMonitor's labels have been updated), the ServiceMonitor's status subresource should be updated.
* When the referenced Prometheus resource is deleted, the ServiceMonitor's status subresource should be updated (this probably implies to setup a finalizer on the Prometheus resources).

### Pod Monitor

TBC

### Probe

TBC

### PrometheusRule

TBC

### AlertmanagerConfig

TBC

## Implementation details

TBC
