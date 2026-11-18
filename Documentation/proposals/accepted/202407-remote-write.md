---
weight: 523
toc: true
title: RemoteWrite CRD
menu:
    docs:
        parent: proposals
lead: ""
images: []
draft: false
---

- Owners:
  - [@superbrothers](https://github.com/superbrothers)
* Status:
  * `Accepted`
- Related Tickets:
  - [#6508](https://github.com/prometheus-operator/prometheus-operator/issues/6508)
- Other docs:
  - n/a

## TL;DR

This design doc proposes RemoteWrite CRD, which enables cluster admins to delegate the ability to configure Prometheus remote_write configuration to application developers/operators.

## Why

The Prometheus remote_write configuration is defined in the Prometheus CRD. Currently, the configuration data generation is the responsibility of cluster admins.

## Goals

The main goal is to enable application developers/operators to self-service the remote write, and configure how the client sends metrics to the remote endpoint.

This means exposing a new CRD to configure Prometheus remote_write configuration.

## Non-goals

Prometheus has two APIs for integrating with remote storage: 'remote write' and 'remote read'. Since no confirmed demand exists to treat remote read as a CRD, we do not cover "RemoteRead" CRD here. We can still implement it later if needed.

Authenticating with the identity of a service account from the RemoteWrite object's namespace (e.g. AWS IAM roles for service accounts or Azure workload identity) is out of scope for the first release.

## How

### Feature gate

The feature is behind the `RemoteWriteCustomResourceDefinition` feature gate (disabled by default). It shields users from bugs while the feature is under development and allows us to implement it in stages.

While the feature gate is disabled by default, the RemoteWrite CRD isn't included in the default bundle.

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

	Spec   monitoringv1.RemoteWriteSpec      `json:"spec"`
	Status monitoringv1.ConfigResourceStatus `json:"status,omitempty"`
}
```

The status subresource reports which Prometheus, PrometheusAgent and ThanosRuler resources select the object and whether it was accepted, like the other configuration resources (requires the `StatusForConfigurationResources` feature gate).

#### Credentials and allowed fields

A RemoteWrite object is typically managed by a different team than the Prometheus resource and its URL is arbitrary. If it could use the credentials available to the Prometheus pod, they could be sent to any endpoint. Hence a RemoteWrite object must only rely on information from its own namespace: the Secret and ConfigMap references are resolved in the namespace of the RemoteWrite object.

The RemoteWrite CRD reuses `RemoteWriteSpec` from the Prometheus CRD, but the following fields aren't allowed:

* References to files in the Prometheus container (`bearerTokenFile`, `authorization.credentialsFile`, `tlsConfig.caFile`, `tlsConfig.certFile` and `tlsConfig.keyFile`).
* Credentials from the Prometheus pod's environment (`sigv4` without `accessKey` and `secretKey`, `sigv4.profile`, `azureAd.managedIdentity`, `azureAd.workloadIdentity`, `azureAd.sdk` and `proxyFromEnvironment`).
* The deprecated `bearerToken` field.

These fields are rejected when the object is created or updated. A field added to `RemoteWriteSpec` in the future is rejected until it's explicitly allowed.

#### Remote write name

The name of the remote write configuration generated from a RemoteWrite object is always `<namespace>/<spec.name>` or `<namespace>/<metadata.name>` if `spec.name` isn't set.

Prometheus fails to load its configuration when 2 remote write configurations have the same name. When the name isn't set, Prometheus identifies a remote write configuration by the hash of its configuration in which the secret values are masked, so 2 configurations which only differ by their credentials would also conflict. Prefixing the name with the namespace ensures that a RemoteWrite object can't conflict with the ones from other namespaces.

A RemoteWrite object whose name conflicts with a remote write configuration of the Prometheus resource is rejected. When 2 RemoteWrite objects from the same namespace conflict, the one which sets `spec.name` is rejected (`metadata.name` is unique in a namespace, so at least one of them sets it). If both set `spec.name`, the one which comes later in the order of the `<namespace>/<name>` key is rejected.

#### Invalid RemoteWrite objects

An invalid RemoteWrite object must not break the Prometheus configuration. The operator applies the same validation as for the remote write configurations of the Prometheus resource (e.g. mutually exclusive authentication methods and the compatibility with the Prometheus version). Invalid objects are excluded from the configuration and the reason is reported in the status subresource and in events.

For example, a RemoteWrite object using `messageVersion: V2.0` is rejected when the Prometheus version is older than v3.4.0 because it would require the operator to enable the `metadata-wal-records` feature flag of Prometheus (and restart the pods).

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

#### Respect enforcedNamespaceLabel and excludedFromEnforcement

The RemoteWrite CRD should respect the Prometheus `.spec.enforcedNamespaceLabel` field.

The object’s namespace is added as the first item in the `write_relabel_configs` of the remote_write configuration.

```yaml
write_relabel_configs:
- source_labels:
  - namespace
  regex: <namespace>
  action: keep
  ...
```

The item is inserted in the first position by the operator before any other user relabelings, so there's no way that users can have access to metrics labeled with namespaces other than the one in which the RemoteWrite object is created.

The RemoteWrite CRD should also respect the Prometheus `.spec.excludedFromEnforcement` field. It is a list of references to objects to be excluded from `enforcedNamespaceLabel`.

Without `enforcedNamespaceLabel`, a RemoteWrite object sends all the series of the Prometheus instance. It's the responsibility of the Prometheus resource's owners to select only trusted namespaces with `remoteWriteNamespaceSelector` in this case.

### Configuration generation

The Prometheus operator will generate the Prometheus configuration including remote_write configuration from the Prometheus CRD and the RemoteWrite resources matching remoteWriteSelector from the namespace(s) selected by remoteWriteNamespaceSelector for additional remote_write configuration.

The operator will respect the --namespaces and --deny-namespaces flags when looking for RemoteWrite objects.

The remote write configurations from the RemoteWrite objects are appended after the ones from the Prometheus resource, sorted by their `<namespace>/<name>` key.

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

Without limits, a RemoteWrite object can exhaust the resources of the Prometheus instance shared with other teams. Hence the limits must be implemented before enabling the feature gate by default.

## PrometheusAgent CRD

The PrometheusAgent CRD also supports the RemoteWrite CRD. Everything explained about the Prometheus CRD works in the same way for the PrometheusAgent.

## ThanosRuler CRD

The ThanosRuler CRD also supports the RemoteWrite CRD with the same `remoteWriteSelector` and `remoteWriteNamespaceSelector` fields.

Thanos Ruler runs in stateless mode when its remote write configuration isn't empty. If `remoteWriteSelector` is set, the ruler runs in stateless mode regardless of the selected RemoteWrite objects, so creating or deleting a RemoteWrite object doesn't switch the ruler between the stateful and stateless modes.

When no RemoteWrite object is selected, the ruler keeps running in stateless mode without any remote write destination and the reconciliation doesn't fail, since creating the ThanosRuler resource before the RemoteWrite objects is a normal case.

## Alternatives

### Prometheus per namespace or team

An application developer/operator can deploy Prometheus instances directly. However, Prometheus instances may be provided as managed by cluster admins.

### The team responsible for the Prometheus object configuring individual remote write destinations on behalf of each "tenant"

An application developer/operator will somehow share the remote_write configuration with the team responsible for Prometheus objects. This includes credential information for remote_write.

## Action Plan

* [X] Add the `RemoteWriteCustomResourceDefinition` feature gate.

  [#8395](https://github.com/prometheus-operator/prometheus-operator/pull/8395)

* [ ] Add the RemoteWrite CRD (not included in the default bundle) with the CEL validation rules.
* [ ] Implement the support in the Prometheus and PrometheusAgent controllers: the selector fields, the selection and validation of RemoteWrite objects, the configuration generation and the status subresource.
* [ ] Implement the support in the ThanosRuler controller.
* [ ] Add the administrative limits (`remoteWriteLimits`).
* [ ] Add end-to-end tests and the user guide.
* [ ] Enable the feature gate by default once the feature is complete and we're confident about its stability.

The graduation of the CRD to `v1beta1` will be covered by a separate proposal.
