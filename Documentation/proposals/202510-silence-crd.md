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
* Track silence identity in a reliable fashion.
* Keep credential handling at the Alertmanager level: credentials belong to the Alertmanager resource, not individual Silence CRs, and `401`/`403` responses from the sidecar must be surfaced as a distinct metric from connectivity failures.

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
      lastSyncTime: "2024-01-15T10:01:00Z"
      conditions:
        - type: Accepted
          status: "True"
          reason: SecretWritten
          lastTransitionTime: "2024-01-15T10:01:00Z"
          observedGeneration: 1
```

**Spec fields:**

| Field                  | Type               | Required                             | Max length | Description                                                                                                                       |
|------------------------|--------------------|--------------------------------------|------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `matchers`             | `[]SilenceMatcher` | Yes (min 1)                          | —          | Alert label matchers                                                                                                              |
| `matchers[].name`      | string             | Yes                                  | 256        | Label name                                                                                                                        |
| `matchers[].value`     | string             | No                                   | 1024       | Label value or regex; empty matches alerts where the label is absent or empty                                                     |
| `matchers[].matchType` | enum               | No (default `=`)                     | —          | One of `=`, `!=`, `=~`, `!~`                                                                                                      |
| `startsAt`             | RFC3339            | No                                   | —          | Start time; defaults to creation timestamp                                                                                        |
| `endsAt`               | RFC3339            | Exactly one of `endsAt` / `duration` | —          | Expiry time                                                                                                                       |
| `duration`             | `NonEmptyDuration` | Exactly one of `endsAt` / `duration` | —          | Expiry as an offset from `startsAt`                                                                                               |
| `comment`              | string             | Yes                                  | 1024       | Human-readable annotation passed through to Alertmanager's `comment` field; useful for referencing change IDs or incident tickets |
| `createdBy`            | string             | No (default `prometheus-operator`)   | 256        | Passed through to Alertmanager's `createdBy` field                                                                                |

CEL validation: exactly one of `endsAt` or `duration` must be set; when both `startsAt` and `endsAt` are present, `startsAt` must precede `endsAt`.

**Every silence has an end date.** Alertmanager requires one on every write, and a silence that never expires is a silence nobody revisits. A long bound is still a bound: `duration: 52w` covers an alerting rule that genuinely cannot be disabled, and forces a review a year later. The two earlier proposals disagreed here — [#5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485) had a silence live as long as its resource, [#7798](https://github.com/prometheus-operator/prometheus-operator/pull/7798) made expiry mandatory — and this follows #7798; see [Alternatives](#alternatives).

`SilenceMatcher` is field-for-field identical to the `Matcher` type `AlertmanagerConfig` uses, as reviewers of the earlier proposal asked. It is declared separately rather than embedded because `v1alpha1.Matcher`, still the storage version for `AlertmanagerConfig`, carries the deprecated `regex` boolean, and importing the `v1beta1` type into a `v1alpha1` resource couples the two versions. The field sets are wire-compatible, so consolidating once `AlertmanagerConfig` graduates is mechanical.

`duration` reuses `monitoringv1.NonEmptyDuration`, already used by `AlertmanagerConfig` for `groupWait`, `groupInterval` and `repeatInterval` and parsed by Prometheus' `model.ParseDuration`. `NonEmptyDuration` rather than `Duration`, because the field is an optional pointer and `Duration` permits the empty string.

**`comment` and `createdBy` are passed through verbatim**, as Alertmanager requires both on every write. `comment` is required on the custom resource — a silence should always carry a reason — and `createdBy` defaults to `prometheus-operator`. Neither plays any role in identity tracking: identity is anchored to the Silence resource's UID, stored as an Alertmanager silence annotation (see [Drift Detection](#drift-detection)).

**Status fields.** `.status` reuses the convention from the [status subresource for config-based resources](accepted/202501-configuration-object-status-subresource.md) proposal — `ConfigResourceStatus` / `WorkloadBinding` / `ConfigResourceCondition`, as implemented for `ServiceMonitor`, `PodMonitor`, `Probe`, `PrometheusRule`, `ScrapeConfig` and `AlertmanagerConfig`. `.status.bindings[]` lists the Alertmanager resources for which the operator has written silence state:

| Field          | Description                                                        |
|----------------|--------------------------------------------------------------------|
| `group`        | API group of the bound workload resource (`monitoring.coreos.com`) |
| `resource`     | Resource type of the bound workload (`alertmanagers`)              |
| `name`         | Name of the Alertmanager resource                                  |
| `namespace`    | Namespace of the Alertmanager resource                             |
| `lastSyncTime` | When the operator last successfully wrote the Secret for this AM   |
| `conditions`   | Per-binding conditions, including `observedGeneration`             |

There is no top-level `conditions` or `observedGeneration`: a Silence may be selected by several Alertmanager resources and its state can legitimately differ between them, so both belong at the binding level.

This follows the same pattern as other configuration resources: the controller writes desired state to a Secret, and a sidecar reads it and applies it to the target service. The distinction from the config-reloader is that triggering a config file reload is passive — Alertmanager reads its own configuration — whereas the silence-sidecar must make active CRUD API calls to Alertmanager to create, update, and delete silences. The binding condition therefore reflects Secret-write success, consistent with how other binding conditions reflect inclusion in the rendered configuration — the sidecar's convergence is observable through its own instrumentation.

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

The controller follows the **Secret-backed sidecar** pattern used by the existing config-reloader:

```
SilenceReconciler (operator)
  watches Silence CRs and Alertmanager CRs
  for each Alertmanager instance that selects a Silence:
    writes/updates a Secret in the Alertmanager's namespace

Secret  (one per Alertmanager instance)
  name: prometheus-operator-silences-<alertmanager-name>
  key:  silences.json

silence-sidecar  (one per Alertmanager pod, like config-reloader)
  watches /etc/alertmanager/silences/silences.json (Secret mounted as a volume by the operator)
  calls local Alertmanager v2 API to create, update, or delete silences
  uses the Silence CR UID stored in an Alertmanager annotation to track identity
  orphan pass: deletes any Alertmanager silence whose annotation UID is absent from the mounted file
```

A dedicated `SilenceReconciler`, separate from the existing Alertmanager controller:

1. Watches Silence and Alertmanager custom resources.
2. On Silence change: resolves target Alertmanager resources via `silenceSelector` / `silenceNamespaceSelector`.
3. For each target Alertmanager resource: rebuilds `silences.json` in full from the informer cache — all active, non-expired Silences that select this Alertmanager, with namespace matchers injected (see [Namespace Isolation](#namespace-isolation)) and timing resolved to absolute RFC3339 — and writes it as the Secret. A full rebuild per reconcile avoids the read-modify-write race when two Silence CRs targeting the same Alertmanager are reconciled concurrently.
4. On Silence deletion: removes the silence's entry from each Alertmanager's Secret. The sidecar's next reconcile detects the orphan and deletes it from Alertmanager. No finalizer is required.

No finalizer is needed on the Silence resource. When the Silence CR is deleted, the operator removes its entry from each Secret; when the Alertmanager CR changes or is deleted, the operator reconciles or removes the Secret accordingly. Cleanup of the actual Alertmanager API silence is delegated to the sidecar.

When the effective `endsAt` is in the past — computed locally, without an API call — the controller sets `Accepted=False` with `reason: SilenceExpired` on each binding and stops writing to the Secret. The resource persists until explicitly deleted; automatic cleanup is a [non-goal](#non-goals). Updating `spec.endsAt` or `spec.duration` to a future time re-activates it: the next reconcile clears the condition and writes the Secret again.

#### Secret Format

The Secret holds a JSON document under the key `silences.json`:

```json
{
  "silences": [
    {
      "uid": "<Silence CR UID>",
      "generation": 3,
      "name": "<Silence CR name>",
      "namespace": "<Silence CR namespace>",
      "matchers": [
        {"name": "namespace", "value": "frontend", "matchType": "="},
        {"name": "alertname", "value": "HighLatency", "matchType": "="}
      ],
      "startsAt":   "2024-01-15T10:00:00Z",
      "endsAt":     "2024-01-15T14:00:00Z",
      "comment":    "CHANGE-12345",
      "createdBy":  "platform-team"
    }
  ]
}
```

Namespace matchers are pre-injected into `matchers` by the operator before writing; the sidecar posts the matchers verbatim without applying any injection logic. Timing fields are resolved to absolute RFC3339 timestamps (see [Alertmanager API Translation](#alertmanager-api-translation)) before the Secret is written.

The operator owns the Secret. The sidecar reads it but never writes back to it.

**Capacity.** A Kubernetes Secret is capped at 1 MiB (the etcd value size limit). At typical silence entry sizes of 400–800 bytes of JSON, a single Secret supports roughly 1,000–2,500 silences per Alertmanager instance before approaching the limit. This is unlikely to be reached in practice. Field-length limits on the CRD (see [API](#api)) protect against a single large `comment` or matcher value consuming a disproportionate share of the budget. Should the limit become a constraint, the natural mitigation is sharding across multiple indexed Secrets (`prometheus-operator-silences-<am-name>-0`, etc.) without changing the sidecar's design.

#### Observability

The `SilenceReconciler` registers no metrics of its own beyond those the operator already exports for every controller via `controller="silence"`: `prometheus_operator_reconcile_operations_total`, `prometheus_operator_reconcile_errors_total`, `prometheus_operator_reconcile_duration_seconds`, and the full workqueue family. Alerting on Secret-write failures is then `prometheus_operator_reconcile_errors_total{controller="silence"}`.

The sidecar requires its own instrumentation to detect convergence failures. The operator writes only to the Kubernetes API and has no visibility into whether the sidecar has successfully applied the Secret to Alertmanager, so sidecar health must be observable independently: metrics exposed by the sidecar and alerting rules that fire when it cannot converge. Sidecar metrics and alerting rules are defined as part of the sidecar implementation task.

#### Failure handling

The operator's write path reaches only the Kubernetes API (to read and write the Secret). The sidecar handles all Alertmanager API interactions.

**Operator (Secret write) failures:**

| Situation                                           | `reason`            | Retry               |
|-----------------------------------------------------|---------------------|---------------------|
| Secret written successfully                         | `SecretWritten`     | — (`Accepted=True`) |
| Kubernetes API unreachable or returning 5xx         | `SecretWriteFailed` | backoff             |
| RBAC denies the operator write access to the Secret | `SecretWriteDenied` | none                |
| Effective `endsAt` already passed                   | `SilenceExpired`    | none                |

Retryable failures are requeued with the workqueue's exponential backoff. The rest are not: the fix is a change to the Silence or the Alertmanager resource. They are still re-attempted on the ordinary resync and whenever either resource changes, so a corrected configuration recovers without the controller tracking what it gave up on.

**Sidecar (Alertmanager API) failures** are reported through sidecar metrics rather than Silence status conditions. See [Observability](#observability).

### silence-sidecar

The `silence-sidecar` is a dedicated binary, one instance per Alertmanager pod, analogous to the existing `config-reloader`. It owns the entire Alertmanager API interaction: the operator writes desired state to a Secret; the sidecar reads it and converges the live Alertmanager instance.

#### Deployment

When the `SilenceCRD` feature gate is enabled, the operator injects the `silence-sidecar` container and its associated volume into the Alertmanager StatefulSet template — the same mechanism used by `config-reloader`. Specifically:

* A `silence-config` volume of type `Secret` is added, backed by the `prometheus-operator-silences-<am-name>` Secret.
* The volume is mounted at `/etc/alertmanager/silences` inside the `silence-sidecar` container.
* The container is added alongside the existing `alertmanager` and `config-reloader` containers.

No changes to existing containers are required.

#### Configuration

| Flag                   | Default                                    | Description                            |
|------------------------|--------------------------------------------|----------------------------------------|
| `--silence-file`       | `/etc/alertmanager/silences/silences.json` | Path to the mounted silences JSON file |
| `--alertmanager-url`   | `http://localhost:9093`                    | Local Alertmanager API endpoint        |
| `--web.listen-address` | `:9097`                                    | Address to expose the metrics endpoint |

#### Reconcile loop

The sidecar reconciles on two triggers: process start, and whenever the mounted `silences.json` file changes (via filesystem watch; polling every 30 seconds as a fallback). Each reconcile cycle:

1. Read and unmarshal `silences.json` from the mounted path. If the file is absent or empty, treat it as an empty list (all operator-managed silences should be deleted).
2. List all Alertmanager silences carrying the `monitoring.coreos.com/silence-uid` annotation.
3. For each entry in the file: look up the Alertmanager silence by annotation UID.
   - If **absent**: create a new silence, setting the annotation to the entry's `uid` and storing the entry's `generation` in the silence's comment metadata.
   - If **present but different**: check the generation. If the AM silence was created from a newer generation than the file entry, skip the update (stale Secret content). Otherwise, update the silence.
   - If **matching**: no write.
4. **Orphan pass**: delete any Alertmanager silence whose `monitoring.coreos.com/silence-uid` annotation does not appear in the file. This handles Silence CR deletion without requiring a finalizer on the Kubernetes resource.

The sidecar never writes back to the Secret. It is a consumer, not a co-owner.

#### Observability

The sidecar exposes a Prometheus metrics endpoint. Key metrics:

| Metric                                        | Labels                                                 | Description                                     |
|-----------------------------------------------|--------------------------------------------------------|-------------------------------------------------|
| `silence_sidecar_reconcile_total`             | —                                                      | Total reconcile attempts                        |
| `silence_sidecar_sync_errors_total`           | `reason="unauthorized"`, `"unreachable"`, `"rejected"` | Breakdown of Alertmanager API failures          |
| `silence_sidecar_last_sync_timestamp_seconds` | —                                                      | Unix timestamp of the last successful reconcile |

`silence_sidecar_sync_errors_total` fulfils the auth-classification goal: `reason="unauthorized"` covers `401`/`403` responses, `reason="unreachable"` covers connection failures, and `reason="rejected"` covers `400`/`422` responses where Alertmanager rejected the silence payload. Operators can alert on the `unauthorized` label independently of general connectivity failures.

`silence_sidecar_last_sync_timestamp_seconds` is a staleness gauge: an alerting rule of the form `time() - silence_sidecar_last_sync_timestamp_seconds > 300` fires when the sidecar has not successfully converged within five minutes, covering the case where the sidecar is running but unable to reach Alertmanager.

#### Shipping

The `silence-sidecar` is delivered as a separate binary within the prometheus-operator release, following the same packaging model as `config-reloader`. Whether it ships in the same container image or a dedicated one is an implementation decision made when the binary is introduced; the StatefulSet injection logic in the operator must be updated to reference whichever image is chosen.

### Drift Detection

Silence identity in Alertmanager is tracked via annotations, using the annotation support added in [alertmanager PR #4965](https://github.com/prometheus/alertmanager/pull/4965). This requires **Alertmanager >= 0.32.0**.

When the sidecar creates or updates a silence it sets:

```
monitoring.coreos.com/silence-uid: <Silence CR UID>
```

The annotation is the key used by the reconcile loop (described in [silence-sidecar § Reconcile loop](#reconcile-loop)) to create, update, and orphan-delete silences without touching anything the operator did not create.

**The sidecar only touches silences it created.** Silences without the `monitoring.coreos.com/silence-uid` annotation — created through the UI, `amtool`, or any other client — are never examined, modified or deleted. Declarative management and ad-hoc silences coexist on the same Alertmanager without interference.

Users who tamper with the annotation are on their own: removing it causes the sidecar to treat the silence as unmanaged and create a duplicate on the next reconcile; setting it to a foreign UID causes the sidecar to overwrite an unrelated silence.

**UUID replacement on matcher edits.** Alertmanager keeps the silence UUID only where the change cannot alter the historic view of silencing; otherwise it expires the existing silence and stores the replacement under a fresh UUID. The annotation-based approach handles this naturally: each POST includes the annotation in the request body, so the replacement silence carries the same annotation UID. The original expired silence also retains its annotation. The sidecar's list pass finds the active silence (by annotation) and reconciles against it; the expired one is left as history.

Alertmanager silence expiry is two-phased: once `endsAt` passes the silence enters `expired` state but remains accessible via the API; once `endsAt + spec.retention` passes (default 120h) it is purged. An expired silence reached by the list pass is treated the same as a missing one — the sidecar creates a new silence — provided the Silence CR itself has not also expired (which the operator checks before writing to the Secret).

### Alertmanager API Translation

The Alertmanager v2 API differs from the CRD spec in two ways the sidecar must bridge:

**Matchers:** the API uses `isRegex` (required bool) and `isEqual` (optional bool, default `true`) rather than a `matchType` string:

| `matchType` | `isRegex` | `isEqual` |
|-------------|-----------|-----------|
| `=`         | `false`   | `true`    |
| `!=`        | `false`   | `false`   |
| `=~`        | `true`    | `true`    |
| `!~`        | `true`    | `false`   |

This matches the convention already used by `AlertmanagerConfig` matchers.

**Timing:** the API requires `startsAt` and `endsAt` as absolute RFC3339 timestamps; there is no `duration` at the API level. The operator computes `endsAt = startsAt + duration` before writing to the Secret, defaulting `startsAt` to the resource's creation timestamp. Because exactly one of `endsAt` or `duration` is always present, a concrete `endsAt` can always be computed and no synthetic far-future expiry is ever written.

**`startsAt` normalization.** When a silence is posted with a `startsAt` in the past — which is the common case for a Silence CR that was created before its start time elapsed — Alertmanager normalizes `startsAt` to the current time. The value stored in Alertmanager will therefore differ from the `startsAt` written to the Secret. The sidecar's update-detection logic must not compare `startsAt`, or it will trigger a spurious update on every reconcile cycle. Only `endsAt`, matchers, `comment`, and `createdBy` are used to decide whether an update is needed.

### Namespace Isolation

The existing `alertmanagerConfigMatcherStrategy` field controls whether a `namespace` matcher is injected into `AlertmanagerConfig` routes and inhibit rules. The Silence controller reads the **same field**, giving operators a single knob for namespace isolation across both config-based and silence-based resources.

| Strategy value                              | Behavior for Silence resources                                                         |
|---------------------------------------------|----------------------------------------------------------------------------------------|
| `OnNamespace` (default)                     | Prepends `namespace=<resource namespace>` to matchers before Secret write              |
| `OnNamespaceExceptForAlertmanagerNamespace` | Same, except Silences in the Alertmanager's own namespace receive no namespace matcher |
| `None`                                      | No namespace matcher is injected                                                       |

When injection applies, the operator prepends `namespace=<resource namespace>` before writing to the Secret: a Silence in namespace `frontend` with matcher `severity=critical` is written as `namespace="frontend" AND severity="critical"`. The sidecar posts the matchers verbatim; it applies no strategy logic of its own.

The API documentation for the `alertmanagerConfigMatcherStrategy` field will need updating to cover the `OnNamespaceExceptForAlertmanagerNamespace` and `None` behaviors for Silence resources.

**Matcher order is significant.** Alertmanager compares matchers as an ordered list, not a set, so the controller emits them in a stable order with the injected `namespace` matcher always in the same position. Reordering alone would be read as a matcher change and replace the silence with a new UUID.

#### Namespace selection

The controller resolves `silenceNamespaceSelector` through the shared `operator.SelectNamespacesFromCache` helper, the same path the Prometheus and ThanosRuler controllers use. The helper encodes the null-selector convention directly:

```go
// If the selector is nil, return the object's namespace.
if sel == nil {
	return []string{obj.GetNamespace()}, nil
}
```

so Silence inherits the operator-wide semantics rather than reimplementing them. This requires a namespace informer and `get`, `list` and `watch` on `namespaces`, matching the RBAC the operator already needs for `AlertmanagerConfig` and ServiceMonitor namespace selection.

### Authentication

The operator's access to Alertmanager is mediated by the Secret, not by a direct API connection. The sidecar runs as a container in the Alertmanager pod and calls `localhost:9093` — in the unauthenticated default no additional credentials are needed.

**Credentials belong to the Alertmanager resource, not individual Silences.** One Alertmanager serves many Silences, so per-Silence credentials would be redundant and would push secret references into a resource application teams create for themselves. When Alertmanager has HTTP authentication configured via `spec.web.httpConfig`, the operator injects the necessary credentials into the sidecar container when it adds it to the StatefulSet — the same information already used elsewhere in the operator's StatefulSet management. This keeps credential surface area proportional to Alertmanager instances, not to the number of Silence resources.

**Auth failures must be distinguishable from connectivity failures.** The sidecar must classify Alertmanager API responses: a `401` or `403` is a credential problem and must be exposed as a distinct counter or metric label (e.g. `reason="unauthorized"`) so operators can tell it apart from a connectivity failure (`reason="unreachable"`). A refused request is not an outage.

**The sidecar reads a mounted file, not the Kubernetes API.** The operator mounts the silence Secret as a volume into the Alertmanager pod (alongside the generated config Secret, following the same pattern in `makeStatefulSetSpec`). The sidecar watches the file at `/etc/alertmanager/silences/silences.json`. The kubelet handles reading the Secret and updating the mounted file when the operator writes a new version — the Alertmanager service account requires no additional RBAC for this.

**Immediate propagation via pod nudge.** After writing the Secret, the operator updates an annotation on the Alertmanager StatefulSet pods (e.g., `operator.prometheus.io/silences-updated: <timestamp>`). This triggers the kubelet to reconcile the pod spec and remount volumes immediately, avoiding the default sync period delay (typically around one minute). The sidecar's filesystem watch detects the file change and reconciles within seconds of the Secret write.

The operator's own RBAC to write the silence Secret in each Alertmanager's namespace is shown in [Feature Gate](#feature-gate).

### Clustered Alertmanager

Alertmanager replicates silences across replicas via its gossip mesh. In the sidecar model, each pod has its own sidecar instance and each independently reconciles the same Secret against its local Alertmanager API. When a silence has already been created and propagated from another replica, the sidecar finds it in the list-by-annotation pass and either updates it (if different) or leaves it alone (if matching). If cluster replication is dysfunctional, each sidecar still converges its local instance independently — the result is consistent even if gossip is not.

The operator treats an Alertmanager *resource* as the unit of Secret reconciliation, not an individual replica. Per-replica visibility can be added later without changing the status shape defined here.

### Feature Gate

A `SilenceCRD` feature gate (default `false` at `v1alpha1`) controls whether the controller runs, following the existing pattern (e.g. `PrometheusAgentDaemonSet`).

The gate is necessary but not sufficient. Even when enabled, the controller performs the same startup pre-flight as the other controllers: it verifies that the `Silence` CRD is installed and that the operator's service account holds the verbs below. It does not start if either check fails, and a failure is logged rather than fatal. This matters on upgrade, where the operator binary may roll out before the new CRD or the updated ClusterRole.

Enabling the `SilenceCRD` gate requires **Alertmanager >= 0.32.0**, which introduced annotation support for silences ([prometheus/alertmanager#4965](https://github.com/prometheus/alertmanager/pull/4965)). This requirement is recorded in the compatibility matrix (`Documentation/getting-started/compatibility.md`).

The version check is not part of the startup pre-flight. The operator manages the Alertmanager StatefulSet and always knows the version — either from `spec.image` on the Alertmanager CR, or from the operator's built-in default (a compile-time constant). The check is performed per-Alertmanager during the reconcile loop: if the resolved image tag indicates a version below `0.32.0`, the controller sets `Accepted=False` with `reason: AlertmanagerVersionTooOld` on the binding and skips the Secret write for that instance.

| Resource          | Verbs                                                         |
|-------------------|---------------------------------------------------------------|
| `silences`        | `get`, `list`, `watch`, `update`                              |
| `silences/status` | `update`                                                      |
| `secrets`         | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

`update` on `silences` is needed to write `.status`. `secrets` access is needed to read and write the silence Secret in each Alertmanager's namespace. No finalizer is required and the `silences/finalizers` verb is not needed.

### Testing and Verification

* Unit tests for helper functions: namespace matcher injection logic across all three `alertmanagerConfigMatcherStrategy` values, timing field resolution (`startsAt` / `endsAt` / `duration`), and Secret JSON serialization.
* Integration tests for the reconcile loop: the full controller logic (informer interactions, selector resolution, Secret write, status updates) is tested at the e2e level due to the many Kubernetes component interactions involved.
* E2e tests against a real Alertmanager >= 0.32.0: create a Silence → verify the silence appears in Alertmanager with the `monitoring.coreos.com/silence-uid` annotation; delete the Silence → verify the silence is eventually deleted from Alertmanager.
* E2e test for namespace isolation: verify `namespace=<ns>` matcher is injected under `OnNamespace`, absent for a Silence in the AM namespace under `OnNamespaceExceptForAlertmanagerNamespace`, and absent under `None`.
* E2e tests for drift correction: edit a managed silence's `comment` through the Alertmanager API → verify the sidecar's next reconcile restores it; delete the silence through the UI → verify the sidecar recreates it.
* E2e test for orphan cleanup: delete a Silence CR → verify the sidecar removes the corresponding Alertmanager silence within one reconcile cycle.
* E2e test for coexistence: create a silence directly through the Alertmanager API without the annotation → verify the sidecar leaves it untouched across reconciles.
* Sidecar convergence metrics and alerting rules verified in the e2e environment.

## Alternatives

### Per-namespace or sharded Secrets

The current design writes one Secret per Alertmanager instance, holding all Silences that target it regardless of source namespace. At large namespace counts this ceiling can become a constraint: 1,000–2,500 entries fit in a 1 MiB Secret, so a cluster with many simultaneously active Silences across many namespaces could exhaust a single Secret.

Two sharding approaches were considered:

**Per-namespace Secret** (`prometheus-operator-silences-<am-name>-<silence-namespace>`): each source namespace gets its own Secret. The per-namespace entry count is naturally small. The sidecar would watch a directory and merge all files before reconciling — the reconcile loop does not care whether entries came from one file or ten. The problem is the Alertmanager StatefulSet pod spec: Kubernetes volumes are declared statically, so the operator would have to update the StatefulSet (causing a pod restart) each time a new Silence namespace appears. At 10,000 namespaces this is not practical with the volume-mount approach.

**Index-based sharding** (`prometheus-operator-silences-<am-name>-0`, `-1`, ...): the operator fills shard 0 to a configurable threshold then spills to shard 1. The sidecar processes all shards in the mounted directory. Adding a new shard still requires a StatefulSet update and pod restart, though less frequently than per-namespace.

Both approaches are compatible with the sidecar's file-watching design — the sidecar union-merges all files in its directory before each reconcile. They were not adopted for the initial implementation because the 1 MiB limit is unlikely to be reached in practice, and deferring sharding keeps the operator and sidecar simpler. The shard-aware path can be added without changing the sidecar's reconcile contract.

The alternative to volume-mounting — having the sidecar watch Secrets via the Kubernetes API — removes the StatefulSet restart constraint entirely: the sidecar could list-watch any number of Secrets dynamically. This reintroduces RBAC on the Alertmanager service account (the volume-mount approach was chosen precisely to avoid it) but may be the right trade-off for deployments operating at that scale.

### Distribute silences across Alertmanager pods

In a clustered Alertmanager (multiple replicas), the current design has every sidecar reconcile every silence. All sidecars read the same Secret, all call their local Alertmanager API to create/update the same set of silences, and Alertmanager's gossip mesh replicates them across the cluster. When a sidecar finds a silence already present (created by another replica's sidecar and propagated via gossip), it checks whether an update is needed and either writes or skips.

An alternative is to **assign each silence to a specific pod** — for example, hashing the Silence CR UID modulo the replica count, so only one sidecar is responsible for reconciling a given silence. This would eliminate the redundant API calls and the brief window where multiple sidecars race to create the same silence.

**Why distribution was not adopted:**

1. **Gossip dependency.** Distribution relies on Alertmanager gossip working correctly to propagate the silence from the assigned pod to the others. If gossip is broken or a network partition isolates a pod, only the silences assigned to reachable pods would be present. A fallback mechanism — having another pod adopt orphaned silences after a timeout — adds complexity and still leaves a gap during the timeout window.

2. **Pod churn.** When replicas scale up or down, or a pod is evicted and replaced, the hash-to-pod mapping changes. Silences would need to be reassigned, which either requires coordination (complex) or results in temporarily duplicated work during the transition.

3. **Generation tracking is sufficient.** The `generation` field added to each Secret entry prevents stale writes in edge cases: if a pod is restarting during a Secret update, or if a sidecar's filesystem watch has brief lag, generation tracking ensures the stale content doesn't overwrite a silence already updated by another pod's sidecar. With the pod nudge mechanism, all pods receive the Secret update nearly simultaneously, so routine propagation skew is not a concern.

4. **Simplicity.** The brute-force approach has no special cases. Every sidecar does the same thing, and the outcome is correct whether zero, one, or all of them succeed. Distribution introduces conditional logic (am I the owner of this silence?), failure detection (is the owner pod healthy?), and handoff (reassign on scale or failure).

The redundant API calls are cheap relative to the operational complexity of distributed ownership. Alertmanager's silence API is designed for concurrent writes — the annotation UID is the natural deduplication key, and the gossip mesh ensures eventual consistency. The current design leans into this rather than fighting it.

### Use `createdBy` as the correlation key

Writing `<namespace>/<name>` into the Alertmanager `createdBy` field, so a silence returned by the API can be matched back to its custom resource, was suggested during review of the earlier proposal. It removes the need to persist anything in `.status`.

It was not taken. `createdBy` is editable in the UI, so one edit detaches the silence and the next reconcile duplicates it; `<namespace>/<name>` is not unique where several clusters share an Alertmanager; the key is lost on rename; and it conflicts with exposing `createdBy` as a user-settable field, which operators asked for.

### Use `comment` as the primary key

The silence-operator (Giant Swarm) sets `comment` to `silence-operator-<namespace>-<name>` and does a list-and-filter on every reconcile. This makes reconciliation stateless, avoiding the optimistic-concurrency handling that writing back to `.status` requires — for example a finalizer update bumping the resource version and conflicting with a subsequent status write.

It was not taken for the same reason as `createdBy`: the field is user-editable, and comment collisions across operators let one controller adopt another's silence. This proposal keeps `comment` as a user-defined field for change IDs and incident references, and never reads it back.

### Store the Alertmanager-assigned UUID in `.status`

An earlier revision of this proposal stored the Alertmanager-assigned silence UUID in `.status.bindings[].silenceID`. This made the controller stateful in a way that adds operational surface: the UUID must survive operator restarts; status writes must be synchronized with the API calls that produce them; resource-version conflicts between other updates and a status write required care.

Annotation-based identity, using the support added in [alertmanager PR #4965](https://github.com/prometheus/alertmanager/pull/4965), resolves these issues. The sidecar discovers the UUID by querying Alertmanager directly — which is authoritative — rather than relying on a cached copy in Kubernetes. There is nothing to persist and nothing to go stale.

### Direct controller → Alertmanager API calls

An earlier revision had the `SilenceReconciler` call the Alertmanager REST API directly, without a sidecar. This raises several concerns:

* **Synchronous calls to a remote service** couple the operator's availability to Alertmanager's. A slow or flapping Alertmanager stalls the operator's workqueue.
* **Finalizers** were required to ensure deletion reached Alertmanager before the Kubernetes resource was removed. A stuck operator would block deletion indefinitely.
* **Clustered Alertmanager**: writing through the governing service rather than to each replica individually relies on gossip replication being healthy, which cannot be assumed.

The sidecar approach avoids all three: the operator writes only to the Kubernetes API; no finalizer is needed; each replica reconciles independently via its local sidecar.

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

* [ ] Add the `Silence` type to the operator's API group (`pkg/apis/monitoring/v1alpha1/`), including the CEL validation rules, and generate CRD manifests, deepcopy and clients (with unit tests)

  <gh issue="">

* [ ] Add `silenceSelector` / `silenceNamespaceSelector` to the `Alertmanager` type (with unit tests)

  <gh issue="">

* [ ] Add the Silence and Secret rules to the operator ClusterRole and the bundled RBAC manifests

  <gh issue="">

* [ ] Add the controller startup pre-flight: CRD presence and operator SA RBAC permission checks

  <gh issue="">

* [ ] Add per-Alertmanager version check to the reconcile loop: resolve version from `spec.image` or operator default; set `AlertmanagerVersionTooOld` binding condition and skip Secret write if below 0.32.0; update `Documentation/getting-started/compatibility.md` with the minimum AM version requirement (with unit tests)

  <gh issue="">

* [ ] Implement `SilenceReconciler`: reconcile Silence CRs into per-Alertmanager Secrets, update Alertmanager pod annotations to trigger immediate kubelet remount, and handle selector resolution via the namespace informer (with unit tests)

  <gh issue="">

* [ ] Implement namespace matcher injection for all three `alertmanagerConfigMatcherStrategy` values (with unit tests)

  <gh issue="">

* [ ] Implement status management: per-binding conditions and expiry handling (with unit tests)

  <gh issue="">

* [ ] Implement the `silence-sidecar` binary: Secret watcher, Alertmanager v2 API client, annotation-based identity tracking, orphan cleanup, and convergence metrics (with unit and integration tests)

  <gh issue="">

* [ ] Wire the `silence-sidecar` into the Alertmanager StatefulSet template

  <gh issue="">

* [ ] Documentation: API reference, getting started guide, and sidecar deployment notes

  <gh issue="">
