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

Everything below follows from that one property.

A first-class Kubernetes resource would give silences a declared source of truth: reviewable in pull requests, tracked in version control, authorized by Kubernetes RBAC alongside the alerting rules they suppress, and reconciled continuously rather than applied once.

### Pitfalls of the current solution

* **Durability depends on deployment topology, not on intent.** Alertmanager replicates silences across a cluster and can snapshot them to disk, so whether a silence survives depends on the replica count, whether `spec.storage` is configured, and whether a restart is rolling or simultaneous. A single-replica Alertmanager on the default `emptyDir` loses every silence when its pod is replaced; a clustered one usually does not. The silence itself expresses no opinion on this, because it was never declared anywhere — so recovering from a loss means somebody remembering what was suppressed.

* **Authorization is all-or-nothing.** Alertmanager has no concept of per-user or per-namespace authorization for silences. Anyone who can reach the API can silence any alert in the cluster, including alerts belonging to teams they have no other access to. Restricting that today means putting a proxy in front of Alertmanager and granting access at the granularity of the whole API.

* **Attribution is self-reported.** Alertmanager records `createdBy` and `comment`, but both are free text supplied by the caller and neither is tied to an authenticated identity. They describe who a client said it was, not who it was.

* **Reconciling from a repository requires building a pipeline.** Declaring silences in Git is possible today, but the machinery is external: a CI job that holds credentials for every Alertmanager, knows every endpoint, runs on a schedule, and has to be re-run to correct drift. Silences created outside that pipeline are invisible to it, and nothing converges between runs.

## Audience

Platform engineers and SREs operating Kubernetes clusters with prometheus-operator who want to manage Alertmanager silences declaratively, and teams that enforce GitOps practices for their observability configuration.

## Goals

* Enable GitOps-friendly silence management via a Kubernetes CRD.
* Integrate with Kubernetes RBAC for fine-grained authorization.
* Support deployments with multiple Alertmanager **resources**, reporting status separately for each resource that selects a Silence.
* Support multi-prometheus-operator deployments with per-operator silence scoping.
* Provide namespace isolation via automatic matcher injection.
* Track silence identity using the Alertmanager-assigned UUID, not a user-settable field.
* Do not assume the Alertmanager API is unauthenticated: confine credential handling to a single integration point, and report a refused request as an authentication failure rather than an outage.

## Non-Goals

* Automatic cleanup of expired Silence resources — users manage their own lifecycle.
* Cross-cluster silence management.
* Real-time sync guarantees — eventual consistency is sufficient.
* Silences for Alertmanager instances not managed by prometheus-operator.
* Reporting status per Alertmanager **replica**. Propagating a silence between the replicas of a single Alertmanager is Alertmanager's own responsibility (see [Clustered Alertmanager](#clustered-alertmanager)).
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

**Every silence has an end date.** Alertmanager requires one on every write, and a silence that never expires is a silence nobody revisits — the alert stops firing, the reason is forgotten, and the suppression outlives the problem. Requiring the author to say how long is the single cheapest guard against that.

This does not prevent suppressing an alerting rule that genuinely cannot be disabled — a real case on platforms that ship immutable rules. A long bound is still a bound: `duration: 52w` expresses it, and the renewal it forces a year later is a feature, not an obstacle. What the API declines to offer is a silence with no end at all.

The two earlier proposals reached opposite conclusions here — [#5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485) settled on a silence living as long as its resource, [#7798](https://github.com/prometheus-operator/prometheus-operator/pull/7798) on an expiry being mandatory — so this is a choice between them rather than a departure from a settled position. It follows #7798. See [Alternatives](#alternatives) for the indefinite-silence option and why it is not taken.

`SilenceMatcher` is field-for-field identical to the `Matcher` type that `AlertmanagerConfig` uses — `name`, `value`, `matchType`, with the same `=` / `!=` / `=~` / `!~` enum — and is deliberately so. Reviewers of the earlier proposal asked that the existing type be reused rather than a parallel matcher concept invented, and that is the right instinct.

It is nonetheless declared separately, because the existing type exists in two shapes. `v1beta1.Matcher` is exactly the three fields above. `v1alpha1.Matcher` additionally carries `regex`, a deprecated boolean superseded by `matchType` and retained only for compatibility — and `v1alpha1` is still the storage version for `AlertmanagerConfig`. Embedding that type directly would ship a brand-new API with a deprecated field on its first release and no way to remove it without a breaking change, while importing the `v1beta1` type into a `v1alpha1` resource couples the two versions awkwardly.

Declaring `SilenceMatcher` with the `v1beta1` field set avoids both problems and keeps the types wire-compatible, so consolidating onto a shared matcher once `AlertmanagerConfig` graduates is a mechanical change rather than an API break. `value` is optional for the same reason it is optional there: an empty value is a meaningful matcher, selecting alerts where the label is absent or empty.

`duration` reuses the existing `monitoringv1.NonEmptyDuration` type rather than introducing a new one. Values are parsed by Prometheus' `model.ParseDuration`, supporting the units `y`, `w`, `d`, `h`, `m`, `s`, `ms` and composite expressions such as `1h30m`. `AlertmanagerConfig` already uses this type for `groupWait`, `groupInterval` and `repeatInterval`. `NonEmptyDuration` is preferred over `Duration` because the field is an optional pointer and `Duration` permits the empty string, which would otherwise make a set-but-empty `duration` valid.

**`comment` and `createdBy` are user-defined and passed through verbatim to Alertmanager.** The Alertmanager v2 API requires both on every silence write. `comment` is therefore required on the custom resource as well — a silence should always carry a reason, and defaulting it to an empty string would only push a meaningless value into Alertmanager. `createdBy` is optional and defaults to `prometheus-operator`, so the common case needs no boilerplate while operators who want to attribute silences to a team, a pipeline or an individual can do so.

Critically, neither field plays any role in identity tracking. The controller does not encode custom resource identity into either, and editing either in the Alertmanager UI has no effect on its ability to manage the silence. That is precisely what allows `createdBy` to be user-settable: identity is anchored solely to the Alertmanager-assigned UUID in `.status.bindings[].silenceID` (see [Identity Tracking](#identity-tracking) below), so nothing depends on these fields holding a particular value. The alternative of using `createdBy` as the correlation key is discussed under [Alternatives](#alternatives).

**Status fields.** `.status` reuses the existing configuration-resource status convention described in the [status subresource for config-based resources](accepted/202501-configuration-object-status-subresource.md) proposal, as implemented by `ConfigResourceStatus` / `WorkloadBinding` / `ConfigResourceCondition` for `ServiceMonitor`, `PodMonitor`, `Probe`, `PrometheusRule`, `ScrapeConfig` and `AlertmanagerConfig`. `.status.bindings[]` lists the Alertmanager resources that selected this Silence, each with its own conditions:

| Field        | Description                                                        |
|--------------|--------------------------------------------------------------------|
| `group`      | API group of the bound workload resource (`monitoring.coreos.com`) |
| `resource`   | Resource type of the bound workload (`alertmanagers`)              |
| `name`       | Name of the Alertmanager resource                                  |
| `namespace`  | Namespace of the Alertmanager resource                             |
| `silenceID`  | Alertmanager-assigned UUID for this binding (see below)            |
| `conditions` | Per-binding conditions, including `observedGeneration`             |

There is no top-level `conditions` or `observedGeneration`: a Silence may be selected by several Alertmanager resources, and its state can legitimately differ between them, so both belong at the binding level.

**`silenceID` is the one addition to the standard binding shape**, and it is load-bearing rather than informational. The Alertmanager silence API is not declarative: creating a silence returns a server-assigned UUID, and every later update or delete must address it by that UUID. Unlike every other configuration resource — whose desired state is re-rendered into a configuration file on each reconcile — the controller has no way to re-derive the handle for a silence it created. Without persisting the UUID, an operator restart orphans every silence it has ever created, and the next reconcile creates a duplicate.

The alternative is to correlate by writing a predictable value into a user-visible field such as `createdBy` or `comment` and listing silences back on each reconcile. Both fields are editable in the Alertmanager UI, so neither can anchor identity reliably; both variants are examined under [Alternatives](#alternatives).

Deliberately **not** included, to keep the binding to the minimum that is useful at `v1alpha1`: per-replica sync counters. See [Clustered Alertmanager](#clustered-alertmanager) for why replica-level accounting is left out.

**Condition type.** Reusing `ConfigResourceCondition` constrains `type` to `Accepted`, which is what this proposal assumes. Worth flagging for review: `Accepted` is documented as the workload controller having "accepted the configuration resource and updated the configuration of the workload accordingly … written to the configuration secret". Silence would be the first configuration resource whose application path is a REST call to a running Alertmanager rather than a write to a configuration secret, so the existing wording does not quite fit. The condition's *meaning* — the workload controller accepted this resource and applied it — carries over cleanly, so the proposal reuses `Accepted` rather than introducing a new condition type, but the documented semantics may want widening.

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

The two fields follow the same semantics as the existing `alertmanagerConfigSelector` / `alertmanagerConfigNamespaceSelector` pair:

| Field                      | Value          | Meaning                                          |
|----------------------------|----------------|--------------------------------------------------|
| `silenceSelector`          | null (default) | No Silences are selected — the feature is opt-in |
| `silenceSelector`          | `{}`           | All Silences in the selected namespaces          |
| `silenceSelector`          | set            | Silences whose labels match                      |
| `silenceNamespaceSelector` | null (default) | The Alertmanager's own namespace only            |
| `silenceNamespaceSelector` | `{}`           | All namespaces the operator watches              |
| `silenceNamespaceSelector` | set            | Namespaces whose labels match                    |

Restating the convention in the form already used elsewhere in the API: an empty label selector matches everything, a null label selector matches the current namespace only. With neither field set, no Silences are selected; setting only `silenceSelector` scopes selection to the Alertmanager's own namespace.

Note that `{}` on `silenceNamespaceSelector` means every namespace the operator watches, which is bounded by `--namespaces` and the namespace allow/deny lists rather than being literally cluster-wide.

### Controller

A dedicated `SilenceReconciler` (separate from the existing Alertmanager controller):

1. Watches Silence and Alertmanager custom resources.
2. Resolves target Alertmanager resources via `silenceSelector` / `silenceNamespaceSelector`.
3. For each target Alertmanager resource:
   - Looks up the existing silence by `.status.bindings[].silenceID` if present.
   - Creates or updates the silence via Alertmanager REST API v2.
   - Writes the returned UUID to `.status.bindings[]`.
4. On deletion: removes the finalizer (`monitoring.coreos.com/silence-cleanup`) after deleting the silence from every bound Alertmanager resource by `silenceID`. A 404 from Alertmanager (e.g. silence already GC'd after retention) is treated as success so the finalizer is never stuck.
5. Periodic reconcile on the informer resync, correcting any divergence (see [Drift Detection](#drift-detection)).

Alertmanager silence expiry is two-phased: once `endsAt` passes, the silence enters `expired` state but remains accessible via the API; once `endsAt + spec.retention` passes (default 120h, configurable on the Alertmanager resource), the silence is purged and subsequent GET/DELETE return 404.

When the effective `endsAt` is in the past — computed locally from `spec.endsAt` or `spec.startsAt + spec.duration`, without an AM API call — the controller sets `Accepted=False` with `reason: SilenceExpired` on each binding and stops reconciling. No further Alertmanager API calls are made. The Silence resource persists until explicitly deleted by the user; automatic cleanup is a non-goal (see [Non-Goals](#non-goals)). The 404-on-deletion handling described in point 4 ensures the finalizer is never stuck regardless of which phase the silence is in when the resource is deleted.

To re-activate an expired silence, update `spec.endsAt` (or `spec.duration`) to a future time. The next reconcile detects that the silence is no longer expired, clears the `SilenceExpired` condition, and creates a new silence in Alertmanager with the updated timestamps.

#### Observability

The controller registers no metrics of its own. Wrapping its registerer with `controller="silence"`, as the Prometheus, Alertmanager and ThanosRuler controllers already do, gives it the set the operator exports for every controller: `prometheus_operator_reconcile_operations_total`, `prometheus_operator_reconcile_errors_total`, `prometheus_operator_reconcile_duration_seconds`, `prometheus_operator_syncs` by status, and the full workqueue family. Alerting on failed silence reconciliation — the operational need raised when this feature was first requested — is then `prometheus_operator_reconcile_errors_total{controller="silence"}`.

Two things are deliberately not exposed as metrics. *Which* Alertmanager a silence failed against is reported per binding in `.status`, which is both more precise and more useful than a counter. Alertmanager API latency is not separated from total reconcile duration, which that call dominates anyway. Either could be added if operating the controller shows a need; neither justifies new surface at `v1alpha1`.

#### Failure handling

Writes to the Alertmanager API fail in ways that differ in one respect that matters operationally: whether trying again can succeed. Retrying everything with backoff means hammering a rejected credential forever and burying the cause; retrying nothing means a routine Alertmanager restart permanently breaks silences that would have healed in seconds. The condition reason on each binding is what tells the two apart.

| Situation                                               | `reason`                     | Retry               |
|---------------------------------------------------------|------------------------------|---------------------|
| Silence written successfully                            | `SilenceApplied`             | — (`Accepted=True`) |
| Alertmanager unreachable, restarting, or returning 5xx  | `AlertmanagerUnreachable`    | backoff             |
| `spec.listenLocal` leaves no reachable address          | `AlertmanagerNotAddressable` | none                |
| 401 or 403, from Alertmanager or a proxy in front of it | `AlertmanagerUnauthorized`   | none                |
| 400 — Alertmanager rejected the silence                 | `SilenceRejected`            | none                |
| Effective `endsAt` already passed                       | `SilenceExpired`             | none                |

Retryable failures are requeued with the workqueue's exponential backoff. The rest are not: the controller records the reason and stops, because the fix is a change to the Silence or to the Alertmanager resource, not the passage of time. They are still re-attempted on the ordinary resync and whenever either resource changes, so a corrected configuration recovers within a resync period without the controller holding extra state to track what it has given up on.

Only some of these are distinguishable by type. The generated Alertmanager client returns typed responses for `400` and `404`; everything else, including `401`, `403` and connection failures, arrives through the reader's default branch as an untyped error carrying the HTTP status, so classification is by status code rather than type assertion.

`AlertmanagerUnauthorized` is listed even though authentication is deferred to [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836), because an Alertmanager behind a proxy that requires credentials is reachable, answers, and refuses — and that needs to read as a credentials problem rather than an outage.

#### Controller scoping (`--controller-id`)

prometheus-operator supports a `--controller-id` flag that restricts which custom resources the operator reconciles. When set, the operator only processes resources whose `operator.prometheus.io/controller-id` annotation matches the flag value. Silence resources must carry this annotation to be picked up:

```yaml
metadata:
  annotations:
    operator.prometheus.io/controller-id: <namespace>/<operator-name>
```

This is relevant in multi-operator environments (e.g. OpenShift, where a platform operator and a user-workload operator may both be running) where each operator should only act on the Silences assigned to it. Resources without the annotation are silently ignored when `--controller-id` is set.

### Identity Tracking

`.status.bindings[].silenceID` is the sole identity anchor. It is populated on first creation and used for all subsequent reconciles. This avoids encoding custom resource identity into a user-visible field such as `comment` or `createdBy`, either of which can be edited in the Alertmanager UI and so cannot be relied on.

### Drift Detection

A silence can diverge from its Silence resource while the controller is not looking: someone edits or expires it in the Alertmanager UI, or the Alertmanager loses it entirely — purged after retention, or restarted without persistent storage.

Drift detection is not a separate mechanism. The informer resync period re-enqueues every Silence on a timer — 5 minutes, matching the Alertmanager controller — and the ordinary reconcile path corrects whatever it finds. For each binding:

1. Fetch the silence by `.status.bindings[].silenceID`.
2. If it is **missing (404)**, the silence is gone. Recreate it and record the new UUID in the binding. This covers both retention purges and an Alertmanager that restarted without storage.
3. If it is **present but differs** from what the spec describes, re-post it with the same ID. Alertmanager updates the existing silence rather than creating a second one, so the UUID is preserved.
4. If it **matches**, nothing is sent.

Only the fields the controller sets are compared: matchers (after namespace injection), `startsAt`, `endsAt`, `comment` and `createdBy`. Server-owned fields — `status`, `updatedAt` — are ignored. Comparing them would report drift on every pass and produce a write loop.

**The controller only touches silences it created.** Reconciliation is driven from Silence resources outward: the controller looks up the specific UUIDs recorded in its own status and never lists the Alertmanager's silences to decide what to act on. A silence created through the UI, `amtool`, or any other client has no corresponding binding, is never examined, and is never modified or deleted. Declarative management and ad-hoc silences therefore coexist on the same Alertmanager, which matters because ad-hoc silencing during an incident is the workflow this feature is least likely to replace. Reaping unmanaged silences is a [non-goal](#non-goals).

### Alertmanager API Translation

The Alertmanager v2 API differs from the CRD spec in two ways the controller must bridge:

**Matchers:** The API uses `isRegex` (required bool) and `isEqual` (optional bool, default `true`) rather than a `matchType` string. The controller maps the CRD's `matchType` enum as follows:

| `matchType` | `isRegex` | `isEqual` |
|-------------|-----------|-----------|
| `=`         | `false`   | `true`    |
| `!=`        | `false`   | `false`   |
| `=~`        | `true`    | `true`    |
| `!~`        | `true`    | `false`   |

This matches the convention already used by `AlertmanagerConfig` matchers in prometheus-operator.

**Timing:** The API requires both `startsAt` and `endsAt` as absolute RFC3339 timestamps — there is no `duration` concept at the API level. When a Silence specifies `duration`, the controller parses it with `model.ParseDuration` and computes `endsAt = startsAt + duration` before posting. When `startsAt` is omitted, the controller uses the resource's creation timestamp.

Because exactly one of `endsAt` or `duration` is always present, the controller can always compute a concrete `endsAt` without inventing one, and never needs to post a silence with a synthetic far-future expiry.

### Namespace Isolation

The existing `alertmanagerConfigMatcherStrategy` field on the Alertmanager resource controls whether a `namespace` matcher is injected into `AlertmanagerConfig` routes and inhibit rules. The Silence controller reads the **same field** to apply the same policy to silences, giving operators a single knob for namespace isolation across both config-based and silence-based resources.

When the strategy is `OnNamespace` (the default), the controller prepends `namespace=<resource namespace>` to the silence matchers before posting to the Alertmanager REST API, preventing the silence from matching alerts originating in other namespaces.

Example: A Silence in namespace `frontend` with matcher `severity=critical` is posted to Alertmanager as `namespace="frontend" AND severity="critical"`.

Note: the existing enforcer code that implements this for `AlertmanagerConfig` operates on the generated Alertmanager configuration file and cannot be shared here. The Silence controller implements the same namespace-injection logic independently, applied to the API call path.

#### Namespace selection

The controller resolves `silenceNamespaceSelector` through the shared `operator.SelectNamespacesFromCache` helper, the same code path the Prometheus and ThanosRuler controllers use for their own namespace selectors. The helper encodes the null-selector convention directly:

```go
if sel == nil {
	return []string{obj.GetNamespace()}, nil
}
```

so Silence inherits the operator-wide semantics described above rather than reimplementing them, and arbitrary namespace labels work without special cases:

```yaml
silenceNamespaceSelector:
  matchLabels:
    environment: production
```

This requires the Silence controller to run a namespace informer and hold `get`, `list` and `watch` on `namespaces`, matching the RBAC the operator already needs for `AlertmanagerConfig` and ServiceMonitor namespace selection.

### Alertmanager API Endpoint

The controller reaches Alertmanager over the cluster-internal network only. It never uses `spec.externalUrl`: that field describes how the Alertmanager is reached from outside the cluster, which may be a hostname the operator cannot resolve, may terminate at a proxy enforcing end-user authentication, and may not be set at all.

The URL is derived entirely from the Alertmanager resource:

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

Two details are worth stating because they are easy to get wrong:

**`spec.serviceName` must be honoured.** When it is set the operator does not create `alertmanager-operated` at all — the user supplies the governing service — so a controller that hardcodes `alertmanager-operated` will dial a service that does not exist. The field is explicitly recommended for running several Alertmanager resources in one namespace, which is exactly the topology this proposal sets out to support, so the failure would surface in the case that matters most.

**`spec.listenLocal: true` makes the API unreachable.** It binds the web listener to `127.0.0.1`, so no other pod can connect; it affects the API and UI only, not gossip, so the Alertmanager itself remains healthy and the misconfiguration is otherwise invisible. The controller cannot manage silences for such an Alertmanager. Rather than retrying a connection that can never succeed, it sets `Accepted=False` on that binding with `reason: AlertmanagerNotAddressable` — naming the configuration, not the network, as the problem, so whoever reads the status knows that waiting will not help.

### Authentication

The controller has no way to authenticate to the Alertmanager API, and this proposal does not add one. That is a deliberate deferral, but the work it defers to is less settled than a bare cross-reference would suggest, so it is worth stating precisely.

[Issue #5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836) is a request to support basic auth in `prometheus-config-reloader`. It is open and unassigned, and does not itself contain a design for how the operator authenticates to the APIs it manages. [PR #8577](https://github.com/prometheus-operator/prometheus-operator/pull/8577) proposes one, including a `spec.web.internalUser` field on the shared web configuration, but it is unmerged, marked stale, and carries no approving review; the field was added in response to review comments and that thread is still open. None of it is decided, and this proposal does not assume any particular outcome.

What it does commit to:

* **Credentials belong to the Alertmanager resource, not to individual Silences.** One Alertmanager serves many Silences, so per-Silence credentials would be redundant, and would put secret references into a resource that application teams are expected to create for themselves. No Silence-specific authentication fields are introduced.
* **`newClientForAlertmanager` is the only place a client is constructed.** Whatever mechanism is agreed upstream is read there and nowhere else, so adopting it is a change to one function rather than to the controller.
* **A refused request is reported as a refusal.** An Alertmanager that requires credentials — directly or through a proxy in front of it — is reachable, answers, and declines. The binding condition says `AlertmanagerUnauthorized` rather than reporting an outage or retrying in a loop (see [Failure handling](#failure-handling)).

Supporting such a deployment properly depends on work outside this proposal. Failing legibly in the meantime does not.

#### TLS

When the connection is over HTTPS, the certificate is verified against the system CA pool by default. Where the platform injects a service-serving CA into pods at a well-known path — OpenShift does this at `/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt` — that bundle is used instead, so Alertmanagers presenting platform-issued service certificates verify without additional configuration. Clusters using a CA that is neither in the system pool nor injected at a well-known path are not covered here; that case belongs with the broader connection configuration tracked in #5836.

### Clustered Alertmanager

Alertmanager replicates silences across replicas via its gossip mesh. The controller therefore treats an Alertmanager *resource* as the unit of reconciliation, not an individual replica: it writes the silence once, through the governing service, and lets gossip propagate it. The UUID returned is cluster-wide, so the `silenceID` recorded in the binding stays valid regardless of which replica served the request.

Replica-level accounting is deliberately out of scope at `v1alpha1`. Reporting how many replicas had converged would mean addressing each pod individually through the headless service and listing silences on every one, for every Silence, on every resync — O(silences x replicas) API calls to surface a condition that gossip is expected to resolve on its own within seconds. Replication health is Alertmanager's concern and is already observable through its own `alertmanager_cluster_*` metrics.

The per-binding `Accepted` condition therefore reports whether the controller's write to that Alertmanager resource succeeded, not whether every replica has converged. If the write fails, the condition goes `False` and the reconciliation is retried with exponential backoff. If per-replica visibility proves necessary in practice, it can be added to the binding later without breaking the shape defined here.

### Feature Gate

A `SilenceCRD` feature gate (default `false` at `v1alpha1`) controls whether the controller runs, following the existing prometheus-operator pattern (e.g. `PrometheusAgentDaemonSet`).

The gate is necessary but not sufficient. Even when it is enabled, the controller performs the same startup pre-flight the other controllers do: it verifies that the `Silence` custom resource definition is installed and that the operator's service account actually holds the verbs it needs. It does not start if either check fails, and a failure is logged rather than fatal — the rest of the operator is unaffected. This matters for upgrades, where the operator binary may roll out before the new CRD or the updated ClusterRole has been applied.

The verbs required are:

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

A suggestion raised during review of the earlier proposal was to have the controller write `<namespace>/<name>` into the Alertmanager `createdBy` field, so that a silence returned by the API could be matched back to the custom resource that produced it. This removes the need to persist anything in `.status`, giving the same stateless reconcile loop as the `comment` approach below.

It is not used here, for four reasons:

* `createdBy` is editable in the Alertmanager UI. A single edit detaches the silence from its custom resource, and the next reconcile creates a duplicate — the same fragility as the `comment` approach, which this proposal also rejects.
* It conflicts with exposing `createdBy` as a user-settable field. Alertmanager requires the field, and operators have asked to be able to attribute silences themselves; a field users control cannot simultaneously be a key the controller trusts.
* `<namespace>/<name>` is not unique where several clusters share one Alertmanager — a common topology, and one where two clusters can legitimately hold same-named resources in same-named namespaces.
* It is lossy across renames. A Silence renamed in Git produces a new key, orphaning the silence created under the old one.

The Alertmanager-assigned UUID has none of these properties: it is unique, server-generated, not displayed as an editable field, and already returned by the create call. The cost is that it has to be persisted in `.status` — see [Identity Tracking](#identity-tracking).

### Use `comment` as the primary key

The silence-operator (Giant Swarm) uses the Alertmanager `comment` field as a surrogate identifier, setting it to `silence-operator-<namespace>-<name>` and doing a list-and-filter on every reconcile.

This approach has an appealing property: reconciliation is stateless. On each pass the controller lists all Alertmanager silences and filters by comment prefix — no UUID needs to be stored or retrieved from a status subresource. This avoids the optimistic-concurrency pitfalls that come with writing back to `.status` (for example, a finalizer update bumping the resource version and causing a subsequent status write to conflict).

The cost is fragility: any manual edit to the comment in the Alertmanager UI causes the controller to lose track of the silence and create a duplicate. Comment collisions across operators or users can also cause one controller to adopt another's silence.

Storing the UUID in `.status` is strictly more correct and makes the controller's behaviour predictable regardless of what happens in the Alertmanager UI. The status subresource complexity (optimistic concurrency, resource-version handling) is the trade-off.

This proposal requires `comment` as a user-defined field (useful for change IDs and incident references) while keeping it entirely separate from identity. The controller passes it through verbatim and never reads it back for tracking purposes.

### Extend the existing Alertmanager controller

The Alertmanager controller manages the Alertmanager deployment lifecycle; mixing silence management into it conflates two distinct concerns and complicates ownership. A dedicated controller is easier to reason about, test, and eventually extract.

### Single `expiresAt` field instead of `startsAt` / `endsAt` / `duration`

PR #7798 used a single `expiresAt` field. A three-field model is more expressive at the same validation cost: relative durations (`4h`) cover the common operational case, absolute timestamps cover scheduled maintenance windows, and `startsAt` allows a silence to be declared ahead of the window it applies to. Requiring exactly one of `endsAt` or `duration` keeps the guarantee `expiresAt` gave — every silence ends — without forcing the author to compute a timestamp for "four hours from now".

### Silences that never expire

[#5485](https://github.com/prometheus-operator/prometheus-operator/pull/5485) settled on omitting the expiry field meaning the silence lives as long as its custom resource, matching the behaviour of the silence-operator, where "there is no expire date — while the CR exists, the silence exists". The motivating case is an alerting rule that cannot be disabled, typically one shipped and reconciled by a platform the user does not control.

Alertmanager has no such concept: the API requires an end date on every write, so implementing this means choosing a far-future one. The silence-operator used a thousand years; one of its authors described that as a dirty hack in the same discussion.

It is not adopted here for two reasons:

* **It fails open.** A silence with no end outlives the thing that created it. If the operator is removed, broken, or simply has its feature gate turned off, the silences it created keep suppressing alerts, and nothing surfaces that this has happened. The failure is silent by construction, which is a poor property for a component whose job is to suppress alerts. A bounded silence degrades the other way: it lapses, the alert fires, and somebody finds out.
* **The use case survives without it.** A bound of `52w` suppresses an alert that cannot be disabled just as effectively as one of a thousand years, and expires at a point where a human is asked whether it is still wanted. That review is the part worth keeping.

A third option was raised and is worth recording: keep indefinite silences but have the controller hold them open with a short rolling expiry, renewed on each reconcile, so they lapse shortly after the operator stops running. That preserves both the use case and the fail-safe, at the cost of renewal machinery and a periodic write per silence. It is not proposed for `v1alpha1`, but it is the natural way to add indefinite silences later should the bounded model prove too restrictive.

## Action Plan

* [ ] Add the `Silence` type to `pkg/apis/monitoring/v1alpha1/`, including the CEL validation rules

  <gh issue="">

* [ ] Add `silenceSelector` / `silenceNamespaceSelector` to the `Alertmanager` type

  <gh issue="">

* [ ] Generate CRD manifests, deepcopy and clients; register the new type

  <gh issue="">

* [ ] Add the `SilenceCRD` feature gate, and the startup checks for CRD presence and RBAC permissions

  <gh issue="">

* [ ] Add the Silence rules to the operator ClusterRole and the bundled RBAC manifests

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
