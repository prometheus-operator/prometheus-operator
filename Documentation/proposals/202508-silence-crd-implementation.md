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
reconciliation logic.

**Discovery Mechanism**: Alertmanager resources will define `silenceSelector`
and `silenceNamespaceSelector` fields to select relevant Silence resources,
following the same pattern as AlertmanagerConfig.

**API Interaction**: The controller interacts with Alertmanager via REST API
v2 for CRUD operations. Authentication support will follow the patterns
established for Prometheus/Alertmanager authentication in issues
[#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836)
and [#7004](https://github.com/prometheus-operator/prometheus-operator/pull/7004). Until those are merged, we'll just use kube-rbac proxy for authentication.

Supported authentication mechanisms:
- **No Authentication**: Direct HTTP access to Alertmanager API
- **Basic Authentication**: Using credentials from Kubernetes secrets
- **Bearer Token**: Using service account tokens or custom tokens from secrets
- **TLS Client Certificates**: Using certificates from Kubernetes secrets
- **Reverse Proxy Authentication**: Support for proxy layers like kube-rbac-proxy
  that handle authentication transparently

The operator will read authentication configuration from the Alertmanager CRD
specification, reusing existing authentication fields where possible to maintain
consistency with Prometheus authentication patterns.

### Authentication and Reverse Proxy Support

**Authentication Configuration**: The Alertmanager CRD will be extended with
authentication fields following the same patterns used for Prometheus resources:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: secure-alertmanager
  namespace: monitoring
spec:
  replicas: 3
  # Basic Authentication
  basicAuth:
    username:
      name: alertmanager-auth
      key: username
    password:
      name: alertmanager-auth
      key: password
  
  # Bearer Token Authentication  
  authorization:
    credentials:
      name: alertmanager-token
      key: token
  
  # TLS Client Certificate Authentication
  tlsConfig:
    cert:
      name: alertmanager-client-cert
      key: tls.crt
    keySecret:
      name: alertmanager-client-cert
      key: tls.key
    ca:
      name: alertmanager-ca
      key: ca.crt
```

**Reverse Proxy Support**: The controller supports Alertmanager instances exposed
through reverse proxies like kube-rbac-proxy. In this setup:

1. The reverse proxy handles authentication (e.g., validating service account tokens)
2. The controller accesses the proxy endpoint using its service account credentials
3. The proxy forwards authenticated requests to the underlying Alertmanager instance

Example kube-rbac-proxy configuration:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: rbac-protected-alertmanager
  namespace: monitoring
spec:
  replicas: 3
  # Access via kube-rbac-proxy service
  externalURL: "https://alertmanager-proxy.monitoring.svc.cluster.local:8443"
  authorization:
    credentials:
      name: prometheus-operator-token
      key: token
```

**Security Considerations**: All authentication credentials are stored in Kubernetes
secrets and mounted securely. The controller follows the principle of least privilege,
only requiring permissions to read authentication secrets and access Alertmanager APIs.

**Status Management**: The Silence status tracks the observed generation,
a single Ready condition, and per-Alertmanager status including silence IDs
and sync state to handle cases where multiple Alertmanager resources select
the same Silence.

**Cluster Synchronization**: Create silences on one healthy Alertmanager instance
and rely on gossip protocol for replication, while monitoring sync status across
all cluster members.

**Configuration Drift Detection**: The controller detects drift by periodically
comparing the Silence CRD spec with the actual silence configuration in Alertmanager
(using the `createdBy` correlation). When drift is detected (e.g., user modifies
silence matchers, comment, or expiration via Alertmanager UI), the controller
reconciles by updating the Alertmanager silence to match the CRD spec, ensuring
the CRD remains the source of truth.

**Lifecycle Management**: Use finalizers (`monitoring.coreos.com/silence-cleanup`)
to ensure proper cleanup from Alertmanager API when Silence resources are deleted.

**Coexistence**: Only manage silences created by the operator and leave manually-created
silences untouched. Correlation between Alertmanager API silences and Silence custom
resources is achieved using the `createdBy` field, which contains `<namespace>/<name>`
(e.g., `monitoring/maintenance-silence`) to uniquely identify the source custom resource.

**Namespace-based Tenancy**: To support multi-tenant environments, the Alertmanager
CRD will include an optional `namespaceInjection` field. When enabled, the controller
automatically injects a `namespace=<silence-namespace>` matcher into all silences
created from that namespace, ensuring tenant isolation.

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

* [ ] Add authentication support (basic auth, bearer tokens, TLS client certs, reverse proxy) // This may depend on the prom-op feat being there.

* [ ] Implement status management and error handling with finalizers

* [ ] Add feature gate and graceful degradation

* [ ] Create documentation

* [ ] Add comprehensive test coverage
