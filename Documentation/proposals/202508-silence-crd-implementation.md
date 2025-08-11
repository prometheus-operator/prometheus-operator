# Silence CRD Implementation

* **Owners:**
  * [mcbenjemaa](https://github.com/mcbenjemaa)
  * [@danielmellado](https://github.com/danielmellado)
  * [@mariofer](https://github.com/mariofer)
* **Status:**
  * `Proposed`
* **Related Tickets:**
  * [#5452](https://github.com/prometheus-operator/prometheus-operator/issues/5452)
  * [#2398](https://github.com/prometheus-operator/prometheus-operator/issues/2398)
* **Other docs:**
  * [Original Silence CRD Proposal](202304-silences.md)

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
* Follow existing prometheus-operator patterns and conventions
* Support multi-Alertmanager deployments
* Maintain backward compatibility with existing Alertmanager functionality

## Non-Goals

* Automatic cleanup of expired Silence CRs (users manage lifecycle)
* Cross-cluster silence management (out of scope)
* Advanced templating features (deferred to future iterations)
* Real-time sync guarantees (minor delays acceptable)

## How

### Architecture Choice

The implementation extends the existing Alertmanager controller rather than
creating a standalone controller. This follows the AlertmanagerConfig pattern
and provides:

* Reuse of existing informer infrastructure
* Consistent operator architecture
* Reduced operational complexity
* Integration with current authentication patterns

### API Design

The Silence CRD will be defined in `monitoring.coreos.com/v1alpha1`:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Silence
metadata:
  name: maintenance-silence
  namespace: monitoring
spec:
  comment: "Scheduled maintenance window"
  expiresAt: "2024-01-15T10:00:00Z"
  matchers:
  - name: "alertname"
    value: "HighCPUUsage"
    matchType: "="
  alertmanagerRef:
    name: "main"
    namespace: "monitoring"
```

### Implementation Strategy

**Controller Integration**: The Alertmanager controller will be extended with
silence informers, event handlers, and selection logic.

**Discovery Mechanism**: Alertmanager instances discover relevant silences
through label selectors (`spec.silenceSelector`), namespace selectors
(`spec.silenceNamespaceSelector`), or explicit references
(`spec.alertmanagerRef`).

**API Interaction**: The controller interacts with Alertmanager via REST API
v2 for CRUD operations, using existing authentication patterns with
Kubernetes secrets.

**Status Management**: The Silence status tracks synchronization state across
multiple Alertmanager instances with Kubernetes conditions.

### Testing and Verification

* Unit tests for controller logic and API conversion
* Integration tests with real Alertmanager instances
* End-to-end scenarios covering multi-instance deployments
* Migration testing with existing silence configurations

### Migration Strategy

Existing silences can be migrated using discovery and conversion utilities
that extract current silences via Alertmanager API and transform them to
Silence CRDs with validation.

## Alternatives

### Standalone Controller

**Why not**: Creating a separate silence controller would duplicate
infrastructure (informers, RBAC, configuration) and diverge from the
established pattern of extending existing controllers for related resources.

### Configuration File Approach

**Why not**: Managing silences through Alertmanager configuration files would
require configuration reloads, increase complexity, and prevent real-time
updates. The API-based approach provides immediate feedback and better error
handling.

### External Operator

**Why not**: Using an external silence operator (like silence-operator) would
fragment the ecosystem and require additional deployment complexity.
Integration into prometheus-operator provides unified experience.

## Action Plan

* [ ] Create Silence CRD definition in `pkg/apis/monitoring/v1alpha1/`

* [ ] Extend Alertmanager controller with silence informers and handlers

* [ ] Implement silence manager component for API interactions

* [ ] Add authentication support following existing patterns

* [ ] Implement status management and error handling

* [ ] Create documentation

* [ ] Add comprehensive test coverage
