# RemoteWrite CRD

- Owners:
  - [@superbrothers](https://github.com/superbrothers)
* Status:
  * `Accepted`
- Related Tickets:
  - [#6508](https://github.com/prometheus-operator/prometheus-operator/issues/6508)
- Other docs:
  - n/a

## TL;DR

This design doc proposes RemoteWrite CRD, which enables cluster admins to delegate the ability to configure Prometheus remote_rewrite configuration to application developers/operators.

## Why

The Prometheus remote_write configuration is defined in the Prometheus CRD. Currently, the configuration data generation is the responsibility of cluster admins.

## Goals

The main goal is to enable application developers/operators to self-service the remote write, and configure how the client sends metrics to the remote endpoint.

This means exposing a new CRD to configure Prometheus remote_write configuration.

## Non-goals

Prometheus has two APIs for integrating with remote storage: 'remote write' and 'remote read'. Since no confirmed demand exists to treat remote read as a CRD, we do not cover "RemoteRead" CRD here. We can still implement it later if needed.

## How

### RemoteWrite CRD

The RemoteWrite CRD represents one of the Prometheus remote_write configuration scoped to the resource’s namespace.

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: RemoteWrite
metadata:
  name: example
  namespace: default
spec:
  url: "https://aps-workspaces.us-west-2.amazonaws.com/workspaces/<workspace id>/api/v1/remote_write"
  sigv4:
    region: us-west-2
    accessKey:
      # Secret object references will be resolved in the namespace of this remoteWrite object.
      # In this manifest it is the "default" namespace.
      name: aws_access
      key: access_key
    secretKey:
      name: aws_access
      key: secret_key
```

```go
package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RemoteWrite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec monitoringv1.RemoteWriteSpec `json:"spec"`
}
```

### Prometheus CRD

The Prometheus CRD is extended with 2 new fields (remoteWriteSelector and remoteWriteNamespaceSelector) that define which RemoteWrite resources are associated with this Prometheus instance.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: example
  namespace: default
spec:
  # RemoteWrite objects to be selected. An empty label selector matches all
  # objects. A null label selector matches no objects.
  remoteWriteSelector: {}
  # Namespaces to match for RemoteWrite discovery. An empty label selector
  # matches all namespaces. A null label selector matches the current namespace
  # only.
  remoteWriteNamespaceSelector: {}
  ...
```

#### Respect enforceNamespaceLabel and excludedFromEnforcement

The RemoteWrite CRD should respect the Prometheus `.spec.enforcedNamespaceLabel` field.

The object’s namespace is added as the first item in the `write_relabel_config` of the remote_write configuration.

```yaml
write_relabel_configs:
- sourceLabels:
  - namespace
  regex: <namespace>
  action: keep
  ...
```

The item is inserted in the first position by the operator before any other user relabelings, so there's no way that users can have access to metrics labeled with namespaces other than the one in which the RemoteWrite object is created.

The RemoteWrite CRD should also respect the Prometheus `.spec.excludedFromEnforcement` field. It is a list of references to objects to be excluded from `enforcedNamespaceLabel`.

### Configuration generation

The Prometheus operator will generate the Prometheus configuration including remote_write configuration from the Prometheus CRD and the RemoteWrite resources matching remoteWriteSelector from the namespace(s) selected by remoteWriteNamespaceSelector for additional remote_rewrite configuration.

The operator will respect the --namespaces and --deny-namespaces flags when looking for RemoteWrite objects.

### Administrative limits

Similar to what exists for scrape resources (e.g. `enforcedSampleLimit` for `ServiceMonitor`), we will allow the Prometheus resource's owners to setup upper-bound limits on the remote write options. For example, it will be possible to disable metadata sending or limit the queue capacity:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: example
  namespace: default
spec:
  remoteWriteLimits:
    metadataConfig:
      send: false
    queueConfig:
      capacity: 1024
      maxShards: 16
```

## PrometheusAgent CRD

The PrometheusAgent CRD also supports the RemoteWrite CRD. Everything explained about the Prometheus CRD works in the same way for the PrometheusAgent.

## Alternatives

### Prometheus per namespace or team

An application developer/operator can deploy Prometheus instances directly. However, Prometheus instances may be provided as managed by cluster admins.

### The team responsible for the Prometheus object configuring individual remote write destinations on behalf of each "tenant"

An application developer/operator will somehow share the remote_write configuration with the team responsible for Prometheus objects. This includes credential information for remote_write.
