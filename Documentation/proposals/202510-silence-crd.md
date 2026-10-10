## Silence CRD

* **Owners:**
  * [coffeegoesincodecomesout](https://github.com/coffeegoesincodecomesout)
* **Based on prior work by:**
  * [mcbenjemaa](https://github.com/mcbenjemaa) — original proposal ([PR #5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485))
  * [danielmellado](https://github.com/danielmellado) — follow-up proposal ([PR #7798](https://github.com/prometheus-operator/prometheus-operator/pull/7798)) and the initial API types ([PR #7806](https://github.com/prometheus-operator/prometheus-operator/pull/7806))
* **Status:**
  * `Draft`
* **Related Tickets:**
  * [#5452](https://github.com/prometheus-operator/prometheus-operator/issues/5452)
  * [#2398](https://github.com/prometheus-operator/prometheus-operator/issues/2398)
  * [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836)
* **Other docs:**
  * Original proposal: [PR #5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485)
  * Prior proposal: [PR #7798](https://github.com/prometheus-operator/prometheus-operator/pull/7798)
  * Prior API implementation: [PR #7806](https://github.com/prometheus-operator/prometheus-operator/pull/7806)

> TL;DR: This proposal defines a `Silence` CRD for prometheus-operator that enables declarative, GitOps-friendly management of Alertmanager silences as Kubernetes resources. It supersedes PR #7798, incorporating production experience from the [silence-operator](https://github.com/giantswarm/silence-operator).

## Why

Alertmanager silences are created through the Alertmanager UI or REST API. Both act directly on a running Alertmanager, so a silence exists as runtime state and nowhere else. There is no declared desired state for the cluster to converge on, and no artifact describing what should be silenced or why.

A first-class Kubernetes resource would give silences a declared source of truth: reviewable in pull requests, tracked in version control, authorized by Kubernetes RBAC alongside the alerting rules they suppress, and reconciled continuously rather than applied once.

### Pitfalls of the current solution

* **Durability depends on topology, not on intent.** Survival of a restart depends on the replica count and whether `spec.storage` is configured — a single-replica Alertmanager on the default `emptyDir` loses every silence when its pod is replaced.
* **Authorization is all-or-nothing.** Anyone who can reach the API can silence any alert in the cluster. Restricting that today means putting a proxy in front of Alertmanager.
* **Attribution is self-reported.** `createdBy` and `comment` are free text supplied by the caller, tied to no authenticated identity.
* **Reconciling from a repository requires building a pipeline.** A CI job holding credentials for every Alertmanager, run on a schedule, blind to silences created outside it, converging on nothing between runs.

## Audience

Platform engineers and SREs operating Kubernetes clusters with prometheus-operator who want to manage Alertmanager silences declaratively, and teams that enforce GitOps practices for their observability configuration.

## Goals

* Enable GitOps-friendly silence management via a Kubernetes CRD.
* Integrate with Kubernetes RBAC for fine-grained authorization.
* Support deployments with multiple Alertmanager resources, reporting status separately for each resource that selects a Silence.
* Support multi-prometheus-operator deployments with per-operator silence scoping.
* Provide namespace isolation via automatic matcher injection.
* Track silence identity using the Alertmanager-assigned UUID, not a user-settable field.
* Confine credential handling to a single integration point, and report a refused request as an authentication failure rather than an outage. Do not assume the Alertmanager API is unauthenticated.

## Non-Goals

* Automatic cleanup of expired Silence resources — users manage their own lifecycle.
* Cross-cluster silence management.
* Real-time sync guarantees — eventual consistency is sufficient.
* Silences for Alertmanager instances not managed by prometheus-operator.
* Reporting status per Alertmanager replica. Propagating a silence between the replicas of a single Alertmanager is Alertmanager's own responsibility (see [Clustered Alertmanager](#clustered-alertmanager)).
* Reaping silences the controller did not create. Silences made through the UI, `amtool` or any other client are left alone; an enforcement mode in which Silence resources are the only permitted source of silences is out of scope here (see [Drift Detection](#drift-detection)).

## How

### API

A new namespace-scoped `Silence` resource under `monitoring.coreos.com/v1alpha1`:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Silence
metadata:
  name: high-latency-silence
  namespace: frontend
spec:
  matchers:
    - name: alertname
      value: HighLatency
      matchType: "="
    - name: severity
      value: critical
      matchType: "="
  startsAt: "2024-01-15T10:00:00Z"    # optional; defaults to creation time
  duration: 4h                        # required unless endsAt is set
  comment: "CHANGE-12345"             # required; passed through to Alertmanager
  createdBy: "platform-team"          # optional; defaults to prometheus-operator
status:
  bindings:
    - group: monitoring.coreos.com
      resource: alertmanagers
      name: main
      namespace: monitoring
      silenceID: "abc123-def456-..."
      conditions:
        - type: Accepted
          status: "True"
          reason: SilenceApplied
          lastTransitionTime: "2024-01-15T10:01:00Z"
          observedGeneration: 1
```

**Spec fields:**

| Field                  | Type               | Required                             | Description                                                                                                                       |
|------------------------|--------------------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `matchers`             | `[]SilenceMatcher` | Yes (min 1)                          | Alert label matchers                                                                                                              |
| `matchers[].name`      | string             | Yes                                  | Label name                                                                                                                        |
| `matchers[].value`     | string             | No                                   | Label value or regex; empty matches alerts where the label is absent or empty                                                     |
| `matchers[].matchType` | enum               | No (default `=`)                     | One of `=`, `!=`, `=~`, `!~`                                                                                                      |
| `startsAt`             | RFC3339            | No                                   | Start time; defaults to creation timestamp                                                                                        |
| `endsAt`               | RFC3339            | Exactly one of `endsAt` / `duration` | Expiry time                                                                                                                       |
| `duration`             | `NonEmptyDuration` | Exactly one of `endsAt` / `duration` | Expiry as an offset from `startsAt`                                                                                               |
| `comment`              | string             | Yes                                  | Human-readable annotation passed through to Alertmanager's `comment` field; useful for referencing change IDs or incident tickets |
| `createdBy`            | string             | No (default `prometheus-operator`)   | Passed through to Alertmanager's `createdBy` field                                                                                |

CEL validation: exactly one of `endsAt` or `duration` must be set; when both `startsAt` and `endsAt` are present, `startsAt` must precede `endsAt`.

**Every silence has an end date.** Alertmanager requires one on every write, and a silence that never expires is a silence nobody revisits. A long bound is still a bound: `duration: 52w` covers an alerting rule that genuinely cannot be disabled, and forces a review a year later. The two earlier proposals disagreed here — [#5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485) had a silence live as long as its resource, [#7798](https://github.com/prometheus-operator/prometheus-operator/pull/7798) made expiry mandatory — and this follows #7798; see [Alternatives](#alternatives).

`SilenceMatcher` is field-for-field identical to the `Matcher` type `AlertmanagerConfig` uses, as reviewers of the earlier proposal asked. It is declared separately rather than embedded because `v1alpha1.Matcher`, still the storage version for `AlertmanagerConfig`, carries the deprecated `regex` boolean, and importing the `v1beta1` type into a `v1alpha1` resource couples the two versions. The field sets are wire-compatible, so consolidating once `AlertmanagerConfig` graduates is mechanical.

`duration` reuses `monitoringv1.NonEmptyDuration`, already used by `AlertmanagerConfig` for `groupWait`, `groupInterval` and `repeatInterval` and parsed by Prometheus' `model.ParseDuration`. `NonEmptyDuration` rather than `Duration`, because the field is an optional pointer and `Duration` permits the empty string.

**`comment` and `createdBy` are passed through verbatim**, as Alertmanager requires both on every write. `comment` is required on the custom resource — a silence should always carry a reason — and `createdBy` defaults to `prometheus-operator`. Neither plays any role in identity tracking, which is what allows `createdBy` to be user-settable: identity is anchored solely to the UUID in `.status.bindings[].silenceID`, so editing either field in the Alertmanager UI cannot detach a silence from its resource. Using `createdBy` as the correlation key is discussed under [Alternatives](#alternatives).

**Status fields.** `.status` reuses the convention from the [status subresource for config-based resources](accepted/202501-configuration-object-status-subresource.md) proposal — `ConfigResourceStatus` / `WorkloadBinding` / `ConfigResourceCondition`, as implemented for `ServiceMonitor`, `PodMonitor`, `Probe`, `PrometheusRule`, `ScrapeConfig` and `AlertmanagerConfig`. `.status.bindings[]` lists the Alertmanager resources that selected this Silence, each with its own conditions:

| Field        | Description                                                        |
|--------------|--------------------------------------------------------------------|
| `group`      | API group of the bound workload resource (`monitoring.coreos.com`) |
| `resource`   | Resource type of the bound workload (`alertmanagers`)              |
| `name`       | Name of the Alertmanager resource                                  |
| `namespace`  | Namespace of the Alertmanager resource                             |
| `silenceID`  | Alertmanager-assigned UUID for this binding (see below)            |
| `conditions` | Per-binding conditions, including `observedGeneration`             |

There is no top-level `conditions` or `observedGeneration`: a Silence may be selected by several Alertmanager resources and its state can legitimately differ between them, so both belong at the binding level.

**`silenceID` is the one addition to the standard binding shape**, and it is load-bearing. The silence API is not declarative: creation returns a server-assigned UUID, and every later update or delete must address it by that UUID. Unlike other configuration resources, whose desired state is re-rendered into a configuration file on each reconcile, the controller cannot re-derive that handle — without persisting it, an operator restart orphans every silence it created and the next reconcile duplicates them. Correlating through `createdBy` or `comment` instead is examined under [Alternatives](#alternatives).

Deliberately **not** included, to keep the binding to the minimum useful at `v1alpha1`: per-replica sync counters. See [Clustered Alertmanager](#clustered-alertmanager).

**Condition type.** `ConfigResourceCondition` constrains `type` to `Accepted`. Its documented meaning is phrased in terms of writing the configuration secret; Silence would be the first configuration resource applied by a REST call to a running Alertmanager instead, so the meaning carries over but the wording may want widening.

### Alertmanager CRD Extensions

Two new fields on the `Alertmanager` resource, mirroring the existing `alertmanagerConfigSelector` pattern:

```yaml
spec:
  silenceSelector:
    matchLabels:
      team: frontend
  silenceNamespaceSelector:
    matchLabels:
      environment: production
```

| Field                      | Value          | Meaning                                          |
|----------------------------|----------------|--------------------------------------------------|
| `silenceSelector`          | null (default) | No Silences are selected — the feature is opt-in |
| `silenceSelector`          | `{}`           | All Silences in the selected namespaces          |
| `silenceSelector`          | set            | Silences whose labels match                      |
| `silenceNamespaceSelector` | null (default) | The Alertmanager's own namespace only            |
| `silenceNamespaceSelector` | `{}`           | All namespaces the operator watches              |
| `silenceNamespaceSelector` | set            | Namespaces whose labels match                    |

This is the convention already used by `alertmanagerConfigSelector` / `alertmanagerConfigNamespaceSelector`. "All namespaces the operator watches" is bounded by `--namespaces` and the namespace allow/deny lists, not literally cluster-wide.

### Controller

A dedicated `SilenceReconciler`, separate from the existing Alertmanager controller:

1. Watches Silence and Alertmanager custom resources.
2. Resolves target Alertmanager resources via `silenceSelector` / `silenceNamespaceSelector`.
3. For each target Alertmanager resource:
   - Looks up the existing silence by `.status.bindings[].silenceID` if present.
   - Creates or updates the silence via Alertmanager REST API v2.
   - Writes the returned UUID to `.status.bindings[]`.
4. On deletion: removes the finalizer (`monitoring.coreos.com/silence-cleanup`) after deleting the silence from every bound Alertmanager resource by `silenceID`. A 404 (e.g. silence already GC'd after retention) is treated as success so the finalizer is never stuck.
5. Periodic reconcile on the informer resync, correcting any divergence (see [Drift Detection](#drift-detection)).

Alertmanager silence expiry is two-phased: once `endsAt` passes the silence enters `expired` state but remains accessible via the API; once `endsAt + spec.retention` passes (default 120h) it is purged and subsequent GET/DELETE return 404.

When the effective `endsAt` is in the past — computed locally, without an API call — the controller sets `Accepted=False` with `reason: SilenceExpired` on each binding and stops reconciling. The resource persists until explicitly deleted; automatic cleanup is a [non-goal](#non-goals). Updating `spec.endsAt` or `spec.duration` to a future time re-activates it: the next reconcile clears the condition and creates a new silence.

#### Observability

The controller registers no metrics of its own. Wrapping its registerer with `controller="silence"`, as the Prometheus, Alertmanager and ThanosRuler controllers already do, gives it the set the operator exports for every controller: `prometheus_operator_reconcile_operations_total`, `prometheus_operator_reconcile_errors_total`, `prometheus_operator_reconcile_duration_seconds`, `prometheus_operator_syncs` by status, and the full workqueue family. Alerting on failed silence reconciliation — the operational need raised when this feature was first requested — is then `prometheus_operator_reconcile_errors_total{controller="silence"}`.

#### Failure handling

Writes to the Alertmanager API differ in one respect that matters operationally: whether trying again can succeed. Retrying everything means hammering a rejected credential forever; retrying nothing means a routine restart permanently breaks silences that would have healed in seconds.

| Situation                                               | `reason`                     | Retry               |
|---------------------------------------------------------|------------------------------|---------------------|
| Silence written successfully                            | `SilenceApplied`             | — (`Accepted=True`) |
| Alertmanager unreachable, restarting, or returning 5xx  | `AlertmanagerUnreachable`    | backoff             |
| `spec.listenLocal` leaves no reachable address          | `AlertmanagerNotAddressable` | none                |
| 401 or 403, from Alertmanager or a proxy in front of it | `AlertmanagerUnauthorized`   | none                |
| 400 — Alertmanager rejected the silence                 | `SilenceRejected`            | none                |
| Effective `endsAt` already passed                       | `SilenceExpired`             | none                |

Retryable failures are requeued with the workqueue's exponential backoff. The rest are not: the fix is a change to the Silence or the Alertmanager resource, not the passage of time. They are still re-attempted on the ordinary resync and whenever either resource changes, so a corrected configuration recovers without the controller tracking what it gave up on.

Classification is by status code, not type assertion: the generated client returns typed responses for `400` and `404`, but `401`, `403` and connection failures all arrive through the reader's default branch as untyped errors carrying the HTTP status. `AlertmanagerUnauthorized` is listed even though authentication is deferred to [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836), because an Alertmanager behind a proxy requiring credentials is reachable, answers, and refuses — which needs to read as a credentials problem rather than an outage.

#### Controller scoping (`--controller-id`)

When `--controller-id` is set, the operator only processes resources whose `operator.prometheus.io/controller-id` annotation matches. Silence resources must carry it to be picked up:

```yaml
metadata:
  annotations:
    operator.prometheus.io/controller-id: <namespace>/<operator-name>
```

This matters in multi-operator environments — OpenShift, for example, where a platform operator and a user-workload operator may both be running — so that each acts only on the Silences assigned to it. Resources without the annotation are silently ignored.

### Drift Detection

A silence can diverge from its resource while the controller is not looking: someone edits or expires it in the UI, or the Alertmanager loses it entirely — purged after retention, or restarted without persistent storage.

Drift detection is not a separate mechanism. The informer resync re-enqueues every Silence on a timer (5 minutes, matching the Alertmanager controller) and the ordinary reconcile path corrects what it finds. For each binding:

1. Fetch the silence by `.status.bindings[].silenceID`.
2. If **missing (404)**, recreate it and record the new UUID. This covers both retention purges and an Alertmanager restarted without storage.
3. If **present but different**, re-post it with the same ID. Alertmanager updates the existing silence rather than creating a second one, so the UUID is preserved.
4. If it **matches**, nothing is sent.

Only the fields the controller sets are compared: matchers (after namespace injection), `startsAt`, `endsAt`, `comment` and `createdBy`. Server-owned fields — `status`, `updatedAt` — are ignored, since comparing them would report drift on every pass and produce a write loop.

**The controller only touches silences it created.** Reconciliation is driven from Silence resources outward: it looks up the UUIDs recorded in its own status and never lists the Alertmanager's silences to decide what to act on. A silence created through the UI, `amtool` or any other client has no binding and is never examined, modified or deleted. Declarative management and ad-hoc silences therefore coexist on the same Alertmanager — which matters, because ad-hoc silencing during an incident is the workflow this feature is least likely to replace. Reaping unmanaged silences is a [non-goal](#non-goals).

### Alertmanager API Translation

The Alertmanager v2 API differs from the CRD spec in two ways the controller must bridge:

**Matchers:** the API uses `isRegex` (required bool) and `isEqual` (optional bool, default `true`) rather than a `matchType` string:

| `matchType` | `isRegex` | `isEqual` |
|-------------|-----------|-----------|
| `=`         | `false`   | `true`    |
| `!=`        | `false`   | `false`   |
| `=~`        | `true`    | `true`    |
| `!~`        | `true`    | `false`   |

This matches the convention already used by `AlertmanagerConfig` matchers.

**Timing:** the API requires `startsAt` and `endsAt` as absolute RFC3339 timestamps; there is no `duration` at the API level. The controller computes `endsAt = startsAt + duration` before posting, defaulting `startsAt` to the resource's creation timestamp. Because exactly one of `endsAt` or `duration` is always present, a concrete `endsAt` can always be computed and no synthetic far-future expiry is ever posted.

### Namespace Isolation

The existing `alertmanagerConfigMatcherStrategy` field controls whether a `namespace` matcher is injected into `AlertmanagerConfig` routes and inhibit rules. The Silence controller reads the **same field**, giving operators a single knob for namespace isolation across both config-based and silence-based resources.

When the strategy is `OnNamespace` (the default), the controller prepends `namespace=<resource namespace>` to the matchers before posting: a Silence in namespace `frontend` with matcher `severity=critical` is posted as `namespace="frontend" AND severity="critical"`. The existing enforcer operates on the generated configuration file and cannot be shared, so this is implemented independently on the API call path.

#### Namespace selection

The controller resolves `silenceNamespaceSelector` through the shared `operator.SelectNamespacesFromCache` helper, the same path the Prometheus and ThanosRuler controllers use. The helper encodes the null-selector convention directly:

```go
	// If the selector is nil, return the object's namespace.
	if sel == nil {
		return []string{obj.GetNamespace()}, nil
	}
```

so Silence inherits the operator-wide semantics rather than reimplementing them. This requires a namespace informer and `get`, `list` and `watch` on `namespaces`, matching the RBAC the operator already needs for `AlertmanagerConfig` and ServiceMonitor namespace selection.

### Alertmanager API Endpoint

The controller reaches Alertmanager over the cluster-internal network only. It never uses `spec.externalUrl`: that describes access from outside the cluster, may be unresolvable from the operator, may terminate at a proxy enforcing end-user authentication, and may not be set at all. The URL is derived entirely from the Alertmanager resource:

```
<scheme>://<service>.<namespace>.svc:9093<routePrefix>/api/v2/...
```

| Component     | Source                                                                                                                               |
|---------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `scheme`      | `https` when `spec.web.tlsConfig` is set, otherwise `http`                                                                           |
| `service`     | `spec.serviceName` when set; otherwise `alertmanager-operated`, the headless service the operator creates and manages                |
| `namespace`   | the namespace of the Alertmanager resource                                                                                           |
| port          | `9093` — the port Alertmanager always listens on. `spec.portName` names that port in the Service and pod spec but does not change it |
| `routePrefix` | `spec.routePrefix` when set, so handlers registered under a prefix are addressed correctly                                           |

Two cases are easy to get wrong:

* **`spec.serviceName` must be honoured.** When set, the operator does not create `alertmanager-operated` at all, so hardcoding that name dials a service that does not exist. The field is recommended for running several Alertmanager resources in one namespace — exactly the topology this proposal supports.
* **`spec.listenLocal: true` makes the API unreachable.** It binds the web listener to `127.0.0.1`, affecting the API and UI but not gossip, so the Alertmanager stays healthy and the misconfiguration is otherwise invisible. Rather than retrying a connection that can never succeed, the controller sets `Accepted=False` with `reason: AlertmanagerNotAddressable`, naming the configuration rather than the network as the problem.

### Authentication

The controller has no way to authenticate to the Alertmanager API, and this proposal does not add one. [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836) tracks that work; no design is settled, and nothing here assumes a particular outcome. What this proposal does commit to:

* **Credentials belong to the Alertmanager resource, not to individual Silences.** One Alertmanager serves many Silences, so per-Silence credentials would be redundant and would push secret references into a resource application teams create for themselves. No Silence-specific authentication fields are introduced.
* **`newClientForAlertmanager` is the only place a client is constructed.** Adopting whatever is agreed upstream is a change to one function rather than to the controller.
* **A refused request is reported as a refusal.** The binding condition says `AlertmanagerUnauthorized` rather than reporting an outage or retrying in a loop (see [Failure handling](#failure-handling)).

Over HTTPS the certificate is verified against the system CA pool, or against a platform-injected service CA where one exists at a well-known path — OpenShift provides `/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt` — so Alertmanagers presenting platform-issued service certificates verify without extra configuration. Other custom CAs belong with the broader connection configuration in #5836.

### Clustered Alertmanager

Alertmanager replicates silences across replicas via its gossip mesh. The controller therefore treats an Alertmanager *resource* as the unit of reconciliation, not an individual replica: it writes the silence once, through the governing service, and lets gossip propagate it. The returned UUID is cluster-wide, so the `silenceID` stays valid regardless of which replica served the request.

Replica-level accounting is deliberately out of scope at `v1alpha1`. Reporting how many replicas had converged would mean addressing each pod through the headless service and listing silences on every one, for every Silence, on every resync — O(silences x replicas) API calls to surface a condition gossip is expected to resolve within seconds. Replication health is Alertmanager's concern, already observable through its `alertmanager_cluster_*` metrics.

The per-binding `Accepted` condition therefore reports whether the write to that Alertmanager resource succeeded, not whether every replica has converged. Per-replica visibility can be added to the binding later without breaking the shape defined here.

### Feature Gate

A `SilenceCRD` feature gate (default `false` at `v1alpha1`) controls whether the controller runs, following the existing pattern (e.g. `PrometheusAgentDaemonSet`).

The gate is necessary but not sufficient. Even when enabled, the controller performs the same startup pre-flight as the other controllers: it verifies that the `Silence` CRD is installed and that the operator's service account holds the verbs below. It does not start if either check fails, and a failure is logged rather than fatal. This matters on upgrade, where the operator binary may roll out before the new CRD or the updated ClusterRole.

| Resource              | Verbs                            |
|-----------------------|----------------------------------|
| `silences`            | `get`, `list`, `watch`, `update` |
| `silences/status`     | `update`                         |
| `silences/finalizers` | `update`                         |

`update` on `silences` is needed for finalizer management, and `silences/finalizers` separately for clusters running the `OwnerReferencesPermissionEnforcement` admission plugin.

### Testing and Verification

* Unit tests for the reconcile loop, status updates, and finalizer handling.
* E2e tests against a real Alertmanager: create a Silence → verify the silence appears in Alertmanager API with correct `silenceID` stored in status; delete the Silence → verify silence is removed from Alertmanager.
* E2e test for namespace isolation: verify `namespace=<ns>` matcher is injected when `alertmanagerConfigMatcherStrategy: OnNamespace`.
* E2e test for drift correction: modify a managed silence directly through the Alertmanager API → verify the next reconcile restores it under the same `silenceID`; delete it outright → verify it is recreated and the new UUID recorded.
* E2e test for coexistence: create a silence directly through the Alertmanager API → verify the controller leaves it untouched across reconciles.
* E2e test for failure reporting: point a Silence at an unreachable Alertmanager → verify `Accepted=False` with `AlertmanagerUnreachable`, and that it recovers once the Alertmanager is available.

## Alternatives

### Use `createdBy` as the correlation key

Writing `<namespace>/<name>` into the Alertmanager `createdBy` field, so a silence returned by the API can be matched back to its custom resource, was suggested during review of the earlier proposal. It removes the need to persist anything in `.status`.

It was not taken. `createdBy` is editable in the UI, so one edit detaches the silence and the next reconcile duplicates it; `<namespace>/<name>` is not unique where several clusters share an Alertmanager; the key is lost on rename; and it conflicts with exposing `createdBy` as a user-settable field, which operators asked for.

### Use `comment` as the primary key

The silence-operator (Giant Swarm) sets `comment` to `silence-operator-<namespace>-<name>` and does a list-and-filter on every reconcile. This makes reconciliation stateless, avoiding the optimistic-concurrency handling that writing back to `.status` requires — for example a finalizer update bumping the resource version and conflicting with a subsequent status write.

It was not taken for the same reason as `createdBy`: the field is user-editable, and comment collisions across operators let one controller adopt another's silence. This proposal keeps `comment` as a user-defined field for change IDs and incident references, and never reads it back.

### Extend the existing Alertmanager controller

The Alertmanager controller manages the deployment lifecycle; mixing silence management into it conflates two concerns and complicates ownership. A dedicated controller is easier to reason about, test, and eventually extract.

### Single `expiresAt` field instead of `startsAt` / `endsAt` / `duration`

PR #7798 used a single `expiresAt`. The three-field model is more expressive at the same validation cost: relative durations cover the common operational case, absolute timestamps cover scheduled maintenance windows, and `startsAt` allows a silence to be declared ahead of the window it applies to. Requiring exactly one of `endsAt` or `duration` keeps the guarantee `expiresAt` gave — every silence ends — without making the author compute a timestamp for "four hours from now".

### Silences that never expire

[#5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485) had an omitted expiry mean the silence lives as long as its custom resource, matching the silence-operator. Alertmanager has no such concept — the API requires an end date — so implementing it means choosing a far-future one; the silence-operator used a thousand years, which one of its authors called a dirty hack in the same discussion.

It was not taken because it fails open: if the operator is removed, broken, or has its feature gate turned off, the silences it created keep suppressing alerts and nothing surfaces that. A bounded silence degrades the other way — it lapses, the alert fires, and somebody finds out.

A third option was raised and is worth recording: keep indefinite silences, but have the controller hold them open with a short rolling expiry renewed on each reconcile, so they lapse shortly after the operator stops running. That preserves both the use case and the fail-safe, at the cost of renewal machinery and a periodic write per silence. It is the natural way to add indefinite silences later should the bounded model prove too restrictive.

## Action Plan

The feature gate lands first so that every subsequent step can merge incrementally without affecting users.

* [ ] Add the `SilenceCRD` feature gate, disabled by default

  <gh issue="">

* [ ] Add the `Silence` type to `pkg/apis/monitoring/v1alpha1/`, including the CEL validation rules

  <gh issue="">

* [ ] Add `silenceSelector` / `silenceNamespaceSelector` to the `Alertmanager` type

  <gh issue="">

* [ ] Generate CRD manifests, deepcopy and clients; register the new type

  <gh issue="">

* [ ] Add the Silence rules to the operator ClusterRole and the bundled RBAC manifests

  <gh issue="">

* [ ] Add the controller startup pre-flight: CRD presence and RBAC permission checks

  <gh issue="">

* [ ] Implement the Alertmanager API client: endpoint resolution, silence create/read/delete, and classification of API errors into the documented condition reasons

  <gh issue="">

* [ ] Implement `SilenceReconciler`: selector resolution via the namespace informer, namespace matcher injection, and finalizer handling

  <gh issue="">

* [ ] Implement status management: bindings, per-binding conditions, expiry handling and drift correction

  <gh issue="">

* [ ] Documentation

  <gh issue="">

* [ ] Unit and e2e test coverage

  <gh issue="">
