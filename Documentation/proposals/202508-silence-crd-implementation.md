# Silence CRD Implementation

* **Owners:**
  * [mcbenjemaa](https://github.com/mcbenjemaa)
  * [@danielmellado](https://github.com/danielmellado)
  * [@mariofer](https://github.com/mariofer)
* **Status:**
  * `Accepted`
* **Related Tickets:**
  * [#5452](https://github.com/prometheus-operator/prometheus-operator/issues/5452)
  * [#2398](https://github.com/prometheus-operator/prometheus-operator/issues/2398)
* **Other docs:**
  * [Original Silence CRD Proposal](https://github.com/prometheus-operator/prometheus-operator/pull/5485)

> TL;DR: This proposal outlines the implementation approach for the Silence
> CRD, enabling GitOps-friendly management of Alertmanager silences through
> Kubernetes resources by extending the existing Alertmanager controller.

## Why

Currently, managing Alertmanager silences requires manual UI interaction or
direct API calls to the Alertmanager API. This creates operational friction
and prevents GitOps integration, making it difficult to manage silences as
code alongside other monitoring configuration.

### Pitfalls of the current solution

* **Manual Operations**: Users must access Alertmanager UI or use custom
  scripts for silence management
* **No GitOps Integration**: Silences cannot be version-controlled or managed
  through standard Kubernetes workflows
* **Limited Access Control**: No fine-grained RBAC for silence operations
* **Multi-cluster Complexity**: Managing silences across multiple clusters
  requires custom tooling and processes
* **No Audit Trail**: Difficult to track who created/modified silences and
  when changes occurred

## Audience

* Platform teams managing monitoring infrastructure
* DevOps engineers implementing GitOps workflows
* SRE teams needing programmatic silence management
* Organizations requiring compliance and audit trails

## Goals

* Enable GitOps workflows for silence management
* Provide RBAC integration for access control
* Support multi-Alertmanager deployments
* Maintain backward compatibility with existing Alertmanager functionality

## Non-Goals

* Automatic cleanup of expired Silence CRs (users manage lifecycle, though the controller
  will detect when Alertmanager has deleted expired silences after the retention period
  (default: 120h, configurable via .spec.retention) and update the CRD status accordingly)
* Cross-cluster silence management (creating silences that span multiple Kubernetes
  clusters or federated Alertmanager deployments is out of scope)
* Real-time sync guarantees (eventual consistency is acceptable, minor delays between
  CRD updates and Alertmanager API changes are expected due to reconciliation loops)
* Authentication support for Alertmanager API access (the controller will only support
  unauthenticated Alertmanager instances, see [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836)
  and [#7004](https://github.com/prometheus-operator/prometheus-operator/pull/7004) for ongoing authentication work), other than that this may also be handled by a reverse proxy such as `kube-rbac-proxy`

## How

*Note: This feature will be protected by a feature gate `EnableSilenceCRD`, with graceful degradation when CRD or RBAC prerequisites are not met.*

### Architecture Choice

The implementation uses a dedicated Silence controller within prometheus-operator
rather than extending the existing Alertmanager controller. This provides:

* Clean separation of concerns between StatefulSet/config management and silences
* Independent scaling and reconciliation patterns
* Simplified testing and maintenance
* Decoupled lifecycle management

### API Design

The Silence CRD will be defined in `monitoring.coreos.com/v1alpha1`:

**Silence Resource Example:**

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Silence
metadata:
  name: maintenance-silence
  namespace: monitoring
  labels:
    app: my-app
    team: platform
spec:
  comment: "Database upgrade maintenance window - disabling alerts for API services"
  expiresAt: "2024-01-15T12:00:00Z"
  matchers:
  - name: "alertname"
    value: "ServiceUnavailable"
    matchType: "="
  - name: "service"
    value: "api-.*"
    matchType: "=~"
  - name: "severity"
    value: "warning"
    matchType: "!="
status:
  observedGeneration: 1
  bindings: # to re-use the same naming than for ServiceMonitor
    - name: "main-alertmanager"
      namespace: "monitoring"
      silenceID: "550e8400-e29b-41d4-a716-446655440000"
      lastSyncTime: "2024-01-15T08:00:15Z"
      syncedInstances: 3
      totalInstances: 3
      conditions:
        - type: Ready
          status: "True"
          lastTransitionTime: "2024-01-15T08:00:15Z"
          reason: SilenceApplied
          message: "Silence applied successfully to Alertmanager instances"
```

**Alertmanager Resource with Silence Selection:**

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main-alertmanager
  namespace: monitoring
spec:
  replicas: 3
  silenceSelector:
    matchLabels:
      team: platform
  silenceNamespaceSelector:
    matchLabels:
      name: monitoring
  # ... other Alertmanager configuration
```

### Implementation Strategy

**Controller Integration**: A dedicated Silence controller will be implemented
within prometheus-operator with its own informers, event handlers, and
reconciliation logic. The controller will:

1. **Watch Silence custom resources**: Detect creation, updates, and deletions
2. **Watch Alertmanager custom resources**: React to changes in silence selectors
3. **Reconcile on Events**: Process changes and sync with Alertmanager APIs
4. **Periodic Reconciliation**: Detect drift and ensure consistency

**Discovery Mechanism**: Alertmanager resources will define `silenceSelector`
and `silenceNamespaceSelector` fields to select relevant Silence resources,
following the same pattern as AlertmanagerConfig.

**Reconciliation Logic**: The controller determines CRUD operations based on:

* **Silence CRD Events**: Create/update/delete operations trigger API calls
* **Generation Changes**: Spec modifications are detected via `observedGeneration`
* **Orphaned Silences**: Periodic cleanup of silences no longer matching any CRD
* **Alertmanager Changes**: Selector modifications trigger re-evaluation of Silence bindings

**API Interaction**: The controller interacts with Alertmanager via REST API
v2 for CRUD operations. Authentication support will follow the patterns
established for Prometheus/Alertmanager authentication in issues
[#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836)
and [#7004](https://github.com/prometheus-operator/prometheus-operator/pull/7004).

Supported authentication mechanisms:

* **No Authentication**: Direct HTTP access to Alertmanager API
* **Reverse Proxy Authentication**: Support for proxy layers that handle
  authentication transparently (users may deploy various reverse proxies
  depending on their authentication requirements)

The operator will read authentication configuration from the Alertmanager CRD
specification, reusing existing authentication fields where possible to maintain
consistency with Prometheus authentication patterns.

### Authentication and Reverse Proxy Support

**Authentication Configuration**: The Silence controller will support authenticated
access to Alertmanager APIs. The specific authentication configuration mechanism
will align with the ongoing work in [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836)
and [#7004](https://github.com/prometheus-operator/prometheus-operator/pull/7004)
to avoid duplication and ensure consistency across the operator.

The controller will support the following authentication scenarios:

* **No Authentication**: Direct access to Alertmanager API
* **Reverse Proxy Authentication**: Access through authentication proxies
* **Native Authentication**: Once #5836 is merged, native authentication support

Authentication configuration will be handled through a dedicated field structure
(not at the main spec level) to be defined as part of the #5836 implementation.

**Reverse Proxy Support**: The controller supports Alertmanager instances exposed
through reverse proxies that handle authentication. In this setup:

1. The reverse proxy handles authentication (e.g., validating service account tokens)
2. The controller accesses the proxy endpoint using its service account credentials
3. The proxy forwards authenticated requests to the underlying Alertmanager instance

Example reverse proxy configuration:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: rbac-protected-alertmanager
  namespace: monitoring
spec:
  replicas: 3
  # Controller will access the internal service directly
  # Authentication configuration will be added once #5836 is implemented
```

**Security Considerations**: All authentication credentials are stored in Kubernetes
secrets and mounted securely. The controller follows the principle of least privilege,
only requiring permissions to read authentication secrets and access Alertmanager APIs.

**Status Management**: The Silence status tracks the observed generation,
a single Ready condition, and per-Alertmanager status including silence IDs
and sync state to handle cases where multiple Alertmanager resources select
the same Silence.

**Cluster Synchronization**: When Alertmanager runs in cluster mode, the controller
creates silences on one healthy Alertmanager instance and relies on the gossip protocol
for automatic replication to other cluster members. The controller monitors sync status
by querying all instances and updates the binding status with `syncedInstances` vs
`totalInstances`. If replication is broken/degraded, the controller will detect
inconsistencies and attempt to create the silence directly on affected instances
to ensure eventual consistency.

**Configuration Drift Detection**: The controller detects drift by periodically
comparing the Silence CRD spec with the actual silence configuration in Alertmanager
(using the `createdBy` correlation). When drift is detected (e.g., user modifies
silence matchers, comment, or expiration via Alertmanager UI), the controller
reconciles by updating the Alertmanager silence to match the CRD spec, ensuring
the CRD remains the source of truth.

**Lifecycle Management**: Use finalizers (`monitoring.coreos.com/silence-cleanup`)
to ensure proper cleanup from Alertmanager API when Silence resources are deleted.

**Coexistence and Correlation**: Only manage silences created by the operator and leave manually-created
silences untouched. Correlation between Alertmanager API silences and Silence custom
resources is achieved using the `createdBy` field in the Alertmanager API, which contains `<namespace>/<name>`
(e.g., `monitoring/maintenance-silence`) to uniquely identify the source custom resource.

**CRUD Operations**:

* **Create**: When a new Silence CRD is created, the controller calls the Alertmanager API to create the silence and stores the returned silence ID in the status
* **Update**: When a Silence CRD is modified, the controller uses the stored silence ID to update the existing silence via the Alertmanager API
* **Delete**: When a Silence CRD is deleted, finalizers ensure the controller calls the Alertmanager API to delete the silence before removing the CRD
* **Drift Detection**: The controller periodically compares Silence CRDs with actual silences in Alertmanager (matching by `createdBy` field) and reconciles any differences

### Namespace-based Tenancy and Matcher Strategy

**AlertmanagerConfig Matcher Strategy Integration**: The Silence CRD will leverage
the existing `alertmanagerConfigMatcherStrategy` field from the Alertmanager CRD
to ensure consistent behavior with AlertmanagerConfig resources. This field controls
how namespace-based matchers are handled:

* **OnNamespace** (default): Automatically inject `namespace=<resource-namespace>`
  matchers into silences, ensuring tenant isolation
* **None**: No automatic namespace injection, allowing cross-namespace silence management

**Namespace-based Tenancy**: When `alertmanagerConfigMatcherStrategy` is set to
`OnNamespace`, the controller automatically injects a `namespace=<silence-namespace>`
matcher into all silences created from Silence custom resources, ensuring tenant
isolation in multi-tenant environments.

Example configuration:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main-alertmanager
  namespace: monitoring
spec:
  replicas: 3
  namespaceInjection: true  # Enable automatic namespace matcher injection
  silenceSelector:
    matchLabels:
      team: platform
```

When a Silence is created in namespace `frontend`:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Silence
metadata:
  name: api-maintenance
  namespace: frontend
spec:
  matchers:
  - name: "service"
    value: "api"
    matchType: "="
```

The controller automatically adds the namespace matcher:

```yaml
# Actual silence sent to Alertmanager API
matchers:
- name: "service"
  value: "api"
  matchType: "="
- name: "namespace"
  value: "frontend"
  matchType: "="
```

This prevents silences in the `frontend` namespace from affecting alerts generated
by resources in other namespaces like `backend` or `database`. The injection
behavior can be disabled for cluster-wide silence management use cases by setting
`namespaceInjection: false` or omitting the field entirely.

### Error Handling and Edge Cases

**API Errors**:

* **Alertmanager Unavailable**: Controller retries with exponential backoff, marks condition as False
* **Authentication Failures**: Logs error, sets binding condition to False with appropriate reason
* **Invalid Silence Data**: Validates before API call, rejects with condition message
* **Network Timeouts**: Configurable timeout, retry logic, eventual failure reporting

**Resource Conflicts**:

* **Duplicate Silence IDs**: Use `createdBy` field to identify ownership, ignore non-operator silences
* **Selector Changes**: Cleanup old bindings, establish new ones based on current selectors
* **Namespace Deletions**: Handle gracefully via owner references and finalizers

**Controller Failures**:

* **Crash Recovery**: Resume processing based on current CRD and Alertmanager state
* **Version Skew**: Handle API version differences gracefully

### Testing and Verification

* Unit tests for controller logic and API conversion
* End-to-end tests with real Alertmanager instances covering multi-instance deployments

## Alternatives

### Standalone Operator

**Why not**: Creating a separate silence operator would fragment the ecosystem
and require additional deployment complexity. Integration into prometheus-operator
provides a unified experience and reuses existing infrastructure.

### External Operator

**Why not**: Using an external silence operator (like [silence-operator](https://github.com/giantswarm/silence-operator)) would
fragment the ecosystem and require additional deployment complexity.
Integration into prometheus-operator provides unified experience.

## Action Plan

* [ ] Create Silence CRD definition in `pkg/apis/monitoring/v1alpha1/`

* [ ] Implement dedicated Silence controller with informers and handlers

* [ ] Add silence selector fields to Alertmanager CRD for resource discovery

* [ ] Implement silence manager component for API interactions

* [ ] Implement status management and error handling with finalizers

* [ ] Add feature gate (`SilenceCRD`) and graceful degradation for controlled rollout

* [ ] Create documentation

* [ ] Add comprehensive test coverage
