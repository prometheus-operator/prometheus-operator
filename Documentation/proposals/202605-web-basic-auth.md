## Basic Auth for the Prometheus, Alertmanager and ThanosRuler Web Servers

* **Owners:**
  * [Leo Bicknell](https://github.com/bicknell-effodio)
* **Status:**
  * `Proposed`
* **Related Tickets:**
  * [#4200](https://github.com/prometheus-operator/prometheus-operator/issues/4200) — Support basic_auth_users in web configuration for Prometheus and Alertmanager
  * [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836) — Add basic auth to prometheus-config-reloader
  * [#4652](https://github.com/prometheus-operator/prometheus-operator/pull/4652) — prior unfinished attempt
  * [#4942](https://github.com/prometheus-operator/prometheus-operator/pull/4942) — prior unfinished attempt
* **Other docs:**
  * [Prometheus HTTPS & authentication guide](https://prometheus.io/docs/prometheus/latest/configuration/https/)
  * [exporter-toolkit web/handler.go](https://github.com/prometheus/exporter-toolkit/blob/master/web/handler.go) — implements `basic_auth_users` + auth cache

> TL;DR: Expose Prometheus's existing `basic_auth_users` web-config setting through the `Prometheus`, `Alertmanager` and `ThanosRuler` CRDs by adding a `basicAuthUsers []BasicAuth` field to the existing `WebConfigFileFields` struct. Handle the downstream effects on liveness/readiness probes, prometheus-config-reloader, and the Thanos sidecar by minting an operator-managed internal user that all in-pod callers use.

## Why

Prometheus, Alertmanager and ThanosRuler all support HTTP basic authentication on their web servers via the `--web.config.file` flag. The operator already builds a web-config Secret for TLS settings but does not emit a `basic_auth_users:` section, so users who want to protect the web UI/API of these workloads have no operator-managed path.

Today's workarounds are unsatisfying:

* Mutual TLS — heavyweight, requires per-client cert management.
* Ingress-level auth (typically ingress-nginx) — the most common workaround in the field, but it has two structural problems detailed below.
* `spec.secrets` + `additionalArgs: --web.config.file=…` — rejected by the operator because `web.config.file` is a managed argument.

### Ingress-level auth is no longer a defensible answer

The most popular existing workaround — and the one frequently recommended in [#4200](https://github.com/prometheus-operator/prometheus-operator/issues/4200)'s discussion — is to put ingress-nginx in front of Prometheus and configure basic auth at the ingress layer. As of 2026 this position is significantly weaker than it was when the issue was filed in 2021, for two independent reasons.

**1. It only protects the perimeter, not the workload.** The Prometheus Pod itself remains unauthenticated. Anything with network reach to the ClusterIP — another workload, a compromised sidecar, a pod running with `hostNetwork`, a lateral attacker who has obtained any in-cluster credential — can read every metric, query the API, and (for Alertmanager) silence or fire alerts at will. The ingress is a thin shell around an open service. Native basic auth on the workload provides defense-in-depth: even an attacker who has bypassed the perimeter must still authenticate to reach the data plane.

**2. ingress-nginx is being retired.** SIG Network and the Kubernetes Security Response Committee [announced the retirement of Ingress NGINX on 2025-11-11](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/), with the project to be retired in **March 2026**. From the announcement:

> "In March 2026, Ingress NGINX maintenance will be halted, and the project will be retired. […] After that time, there will be no further releases, no bugfixes, and no updates to resolve any security vulnerabilities that may be discovered."

A follow-up [statement from the Steering and Security Response Committees on 2026-01-29](https://kubernetes.io/blog/2026/01/29/ingress-nginx-statement/) reiterates:

> "Choosing to remain with Ingress NGINX after its retirement leaves you and your users vulnerable to attack. Existing deployments will continue to work, so unless you proactively check, you may not know you are affected until you are compromised."

The committees recommend that all Ingress NGINX users begin migration immediately. There is no drop-in replacement; users are pointed at Gateway API or third-party controllers. For users whose only reason to run an ingress in front of Prometheus is "to get basic auth," telling them to migrate to Gateway API just to keep that workaround is a poor answer — particularly when the Prometheus binary itself has supported `basic_auth_users` for years and the operator only needs to expose it.

This proposal does not deprecate ingress-level auth as a deployment pattern (organisations have many reasons to run an ingress), but it removes the situation where ingress-nginx is the only practical way to put a password on Prometheus.

### Pitfalls of the current solution

Two community PRs (#4652 in 2022, #4942 in 2022) tried to implement this and were closed unfinished. Both reached the same wall, not on the API design but on the second-order effects: once basic auth is on, the kubelet probes start failing, the `prometheus-config-reloader` cannot POST to `/-/reload`, and the Thanos sidecar cannot query `/api/v1/status/buildinfo`. Neither PR shipped a complete answer for those.

This proposal addresses the operator-owned callers head-on so that turning on `basic_auth_users` does not break any operator-managed plumbing.

## Goals

* Add a `basicAuthUsers []BasicAuth` field to `WebConfigFileFields`, applicable to `Prometheus`, `Alertmanager` and `ThanosRuler`.
* Reuse the existing `BasicAuth` type (Secret-backed; no plaintext in the CR).
* Keep all operator-internal callers (probes, config reloader, Thanos sidecar) functional when basic auth is enabled, transparently to the user.
* Provide a documented, stable name for the operator-managed user's Secret so downstream charts (kube-prometheus, kube-prometheus-stack) can reference it.

## Non-Goals

* Updates to kube-prometheus's self-scrape ServiceMonitor — separate repo, follow-up PR.
* Updates to the kube-prometheus-stack helm chart's `values.yaml` plumbing — separate repo, follow-up PR.
* Grafana datasource credential plumbing — out of scope; helm-chart concern.
* Inline plaintext credentials in the CR — explicitly not supported.
* Removing the existing `spec.secrets` + `additionalArgs` workaround — kept as-is to avoid breaking existing users; the new field is the documented path.
* Operator-side bcrypt hashing of plaintext passwords — see [Alternatives](#alternatives).
* Waiting for [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151) (path exclusion). Useful when it lands but we are not blocking on it.

## Audience

* Operators of Prometheus/Alertmanager/ThanosRuler who need HTTP basic auth on the web server without standing up a sidecar proxy or an ingress.
* Helm chart authors (kube-prometheus, kube-prometheus-stack) who want a stable contract for passing credentials through to the workload.

## How

### API change

Add one field to the existing `WebConfigFileFields`:

```go
// pkg/apis/monitoring/v1/types.go

type WebConfigFileFields struct {
    TLSConfig   *WebTLSConfig  `json:"tlsConfig,omitempty"`
    HTTPConfig  *WebHTTPConfig `json:"httpConfig,omitempty"`

    // basicAuthUsers configures HTTP basic-auth users for the web server.
    //
    // The Password Secret value MUST be a bcrypt hash (the format
    // exporter-toolkit requires). Supported prefixes are $2a$, $2b$, $2y$.
    // Plaintext passwords are not accepted; the operator will not hash them.
    //
    // Username and password references may point at different keys in the
    // same Secret. Usernames must be unique across entries.
    // +optional
    BasicAuthUsers []BasicAuth `json:"basicAuthUsers,omitempty"`
}
```

`BasicAuth` already exists and uses `SecretKeySelector` for both fields, which keeps every credential out of the CR object itself. This satisfies both delivery paths users have asked for:

* Helm-values path (e.g. kube-prometheus-stack): the chart renders a `Secret` from `values.yaml` and the CR references it.
* External secret path (e.g. External Secrets Operator): the user/system creates the `Secret` directly and the CR references it.

The operator does not need to know which path produced the Secret.

### Generated web-config

`pkg/webconfig/config.go` already emits `tls_server_config:` and `http_server_config:` to the mounted web-config Secret. We extend the generator to also emit:

```yaml
basic_auth_users:
  <username>: <bcrypt-hash>
  <username2>: <bcrypt-hash2>
```

Duplicate usernames are rejected at config-generation time (Prometheus accepts the YAML map but silently overrides).

### Operator-managed internal user

When `basicAuthUsers` is non-empty, the operator reconciles one additional Secret per workload CR. The name is a stable, documented contract:

```
<PrefixedName>-web-auth
```

`<PrefixedName>` is the existing operator convention (`prometheus-<name>` for `Prometheus`, `alertmanager-<name>` for `Alertmanager`, `thanos-ruler-<name>` for `ThanosRuler`). The name is committed-to as part of this proposal so downstream consumers (kube-prometheus, kube-prometheus-stack) can rely on it.

The Secret contains three keys:

* `username` — a fixed value (e.g. `prometheus-operator`), so the in-pod callers can read a single key without coordination
* `password` — a generated random plaintext, used by in-pod callers (probes, reloader, sidecar)
* `password-bcrypt` — the bcrypt of `password`, used in the rendered web-config

The Secret is owned by the workload CR (ownerReference) so it GCs on deletion. Its bcrypt-form is appended to the rendered `basic_auth_users:` map alongside the user-supplied entries. The password is regenerated only when the Secret is missing, not on every reconcile.

### Liveness/readiness/startup probes

When basic auth is on, the operator switches from `HTTPGetAction` probes to `ExecAction` probes built by `pkg/operator/prober.go`. The existing helper already shells out to `curl` (or `wget` as fallback); we extend it to add basic-auth arguments and inject `$U` / `$P` env vars from the operator-managed Secret via `valueFrom.secretKeyRef`. The plaintext password never appears in the pod spec or argv.

```sh
# generated probe command (conceptual)
sh -c 'if [ -x "$(command -v curl)" ]; then exec curl --fail -u "$U:$P" http://…/-/ready; \
       elif [ -x "$(command -v wget)" ]; then exec wget -q -O /dev/null --user="$U" --password="$P" http://…/-/ready; \
       else exit 1; fi'
```

This works on every supported Prometheus / Alertmanager / ThanosRuler version regardless of whether [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151) has merged. If/when that lands, we can simplify to path-excluded `HTTPGetAction` probes in a follow-up — the operator-managed user is still useful for the reloader and sidecar.

#### Interaction with the strategic-merge-patch probe workaround

[`Documentation/platform/strategic-merge-patch.md`](../platform/strategic-merge-patch.md) currently documents a workaround for this exact problem: when TLS + a hand-rolled web-config requires auth, users override the operator's `httpGet` probe via a strategic merge patch (typically to `tcpSocket`, since that bypasses authentication entirely). The merge-patch behavior was hardened in v0.91.0 ([#8427](https://github.com/prometheus-operator/prometheus-operator/pull/8427)) and remains a supported escape hatch.

This proposal does not deprecate that workaround. Interaction precedence:

* If `basicAuthUsers` is set **and** the user has not supplied a probe override → operator emits an exec probe with credentials. The probe still calls `/-/ready`, so unlike the `tcpSocket` workaround it remains accurate about WAL replay and TSDB initialization. This is a strict improvement over the documented workaround.
* If `basicAuthUsers` is set **and** the user has supplied a probe override via strategic merge patch → the user's override wins, per the existing merge-patch semantics. The operator's exec probe is replaced. Users opting out this way are responsible for handling auth themselves (e.g. by sticking with `tcpSocket`).
* The note in `strategic-merge-patch.md` about exec-probes being needed "when prioritizing credential security over granular readiness checks" is no longer the only path — users who do nothing get a granular readiness check *and* credential security. The doc will be updated in PR #3 (not invalidated; cross-referenced).

### prometheus-config-reloader

The reloader binary today has no outbound basic-auth support, so it cannot reach `/-/reload` or `/api/v1/status/runtimeinfo` once auth is on. We add two flags:

```
--reload-basic-auth-username=<value>
--reload-basic-auth-password-file=<path>
```

Both are read from a Secret-mounted env var or file so the password never lands in argv. When the operator detects basic auth on the workload, it adds the flags and a `secretKeyRef`-sourced env var to the reloader container, pointing at the operator-managed Secret.

This change also closes [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836) for users running the reloader against a separately-protected Prometheus.

### Thanos sidecar

The sidecar communicates with Prometheus via the YAML file referenced by `--prometheus.http-client-file`, which is built by `pkg/prometheus/server/thanos_sidecar_config.go`. We extend that builder to emit:

```yaml
basic_auth:
  username: <operator-managed>
  password_file: /etc/thanos/auth/password
```

…and mount the operator-managed Secret at that path. The thanos-sidecar http-client format already supports this natively.

### bcrypt performance

bcrypt is cited as a concern because it is intentionally slow (~50 ms per verify at the default cost of 10) and Prometheus's web handler serializes bcrypt comparisons through a global mutex. Two factors make this acceptable in practice:

1. **exporter-toolkit caches successful authentications.** See [`web/handler.go`](https://github.com/prometheus/exporter-toolkit/blob/master/web/handler.go) (`cache.go` LRU, size 100, key = `user + hash + plaintext`). After the first request per `(user, hash, password)` tuple, subsequent verifications are map lookups (~µs). The operator-managed callers use stable credentials, so they are cache hits for the life of the process.

2. **The cold-start window is bounded.** Distinct first-time authentications immediately after a pod restart queue on the global mutex at ~50 ms each. With ten concurrent cold scrapers that is ~500 ms of slow auth, then steady state.

If cold-start latency turns out to be a real problem in production reports, the fallback (not pursued in v1) is to switch the operator-managed internal callers from basic auth to mTLS, reusing the existing `WebTLSConfig` infrastructure. External callers would continue to use basic auth. This approximately doubles the implementation surface and is deferred unless needed.

### Validation and error surfaces

* The operator validates each referenced Secret/key exists at reconcile time using the existing `assets.StoreBuilder.AddBasicAuth` path.
* Duplicate usernames in `basicAuthUsers` fail reconcile with a clear error rather than silently dropping entries.
* Obviously malformed password values (no `$2[aby]$` prefix) produce a warning event so users do not waste hours debugging plaintext-in-Secret.

### Migration / compatibility

* New field, optional, defaults to "off". Zero behavior change for existing users.
* `WebConfigFileFields` is part of the `monitoring.coreos.com/v1` API. Per `CONTRIBUTING.md`, v1 forbids breaking backward or forward compatibility. This change is purely additive (optional field, omitempty); a CR written against the old schema validates and behaves identically against the new schema, and a CR written against the new schema is silently ignored by an older operator binary (it sees an unknown field). Compatible in both directions.
* The existing `spec.secrets` + `additionalArgs: --web.config.file=…` workaround continues to be rejected by the managed-argument guard. The proposal does not change that. Users who need full hand-rolled control retain the existing workaround; the new field is the supported path for everyone else.
* The strategic-merge-patch probe override (see "Interaction with the strategic-merge-patch probe workaround" above) continues to work and is documented as the opt-out for users who want operator-managed basic auth but their own probe definition.

### Testing and verification

* Unit tests: `pkg/webconfig/config_test.go` (YAML output golden tests), reloader flag wiring, ExecAction generation.
* e2e tests: boot a Prometheus / Alertmanager / ThanosRuler with `basicAuthUsers` set, verify probes pass, reloader can trigger a reload, scrape with correct creds returns 200, scrape without creds returns 401.

### Documentation impact

In addition to the per-component user-guide additions in the implementation PRs, the following existing docs are affected and will be updated in-flight (not in this proposal PR):

* `Documentation/user-guides/basic-auth.md` is currently titled "Basic auth for targets" and covers `ServiceMonitor`-side basic auth. PR #3 will extend it with a clearly-titled new section "Basic auth for the Prometheus web server" (and add Alertmanager / ThanosRuler sections in PR #4). The existing scrape-target content is unchanged. The risk of name collision is acknowledged; splitting the file into two was considered but rejected because users searching for "basic auth" benefit from finding both topics in one place.
* `Documentation/platform/strategic-merge-patch.md` will get a cross-reference noting that `spec.web.basicAuthUsers` now provides an operator-managed path that obsoletes the documented workaround for most users.
* `Documentation/platform/thanos.md`, `Documentation/platform/troubleshooting.md`, `Documentation/platform/exposing-prometheus-and-alertmanager.md` — short additions referencing the new field where relevant.
* `Documentation/api-reference/api.md` — autogenerated by `make generate`.

### Open questions

1. **Opt-out of the operator-managed user.** Should users be able to disable it, e.g. when they want to manage probes/reloader auth themselves via exporter-toolkit path exclusion once that lands? Recommendation: no opt-out in v1; revisit if exporter-toolkit#151 merges.

2. **Validation surface.** Recent CRD work (v0.91.0, [#8480](https://github.com/prometheus-operator/prometheus-operator/pull/8480)) has moved toward CEL/admission-time enforcement of structural rules (e.g. mutual exclusion in `ScrapeConfig`). Should obviously-malformed bcrypt values fail admission (CEL on the Secret value at resolve time) rather than emitting a runtime warning event? Recommendation: runtime warning event for v1 (the Secret value is not visible to admission), but open to maintainer preference.

3. **Status condition / metric for "basic auth active".** Cheap to add; useful for users debugging "why is my scrape 401'ing". Defer to maintainer preference.

## Alternatives

1. **Wait for [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151).** Path exclusion for `/-/healthy`, `/-/ready` would eliminate the probe problem and shrink this proposal substantially. The PR has been open since 2024 and is currently in conflict; we cannot block on it. If/when it merges we can simplify probes in a follow-up; the operator-managed user is still needed for the reloader and sidecar.

2. **Mutual TLS only.** Already supported via `WebTLSConfig`. Heavyweight for the common "I just want a password on the UI" use case. Not a replacement for basic auth.

3. **Ingress-level auth.** Only protects perimeter traffic, leaving the in-cluster ClusterIP unauthenticated; and the most-used implementation (ingress-nginx) is being retired in March 2026 with no further security patches. See "Ingress-level auth is no longer a defensible answer" above.

4. **Operator-side bcrypt of plaintext.** Lets users supply plaintext in a Secret and have the operator hash it. Rejected because:
   * Every reconcile would produce a fresh salt, causing a rolling restart on each reconcile, OR
   * The operator would have to persist its own derived Secret and track whether the source changed — significantly more state machinery for marginal UX gain.

5. **Sidecar nginx / custom auth proxy.** What this proposal aims to make unnecessary. Mentioned only because it is what users do today.

## Action Plan

Implementation is split into four upstream PRs, opened sequentially from a single fork. Each is sized to be reviewable in isolation. All PRs follow the conventions in `CONTRIBUTING.md` and `RELEASE.md`:

* Commit subjects use `<subsystem>: <what>` (e.g. `webconfig: emit basic_auth_users section`, `cmd/prometheus-config-reloader: support outbound basic auth`).
* PR description includes the `Type of change` checkbox and a `release-note` block per the PR template. PR #1 is `NONE` (doc-only, no CHANGELOG entry). PRs #2–#4 are `FEATURE`.
* CHANGELOG entries follow the documented prefix ordering (`[CHANGE]`, `[FEATURE]`, `[ENHANCEMENT]`, `[BUGFIX]`).
* Every commit is DCO-signed (`git commit -s`).
* AI use will be disclosed in PR descriptions per the project's AI use policy.

* [ ] **PR #1** — This proposal document.

  <gh issue="4200">

* [ ] **PR #2** — Foundation: add outbound basic-auth flags to `prometheus-config-reloader`, and extend `operator.ExecAction` to inject basic-auth credentials from env vars. No CRD surface change. Independently useful; closes [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836).

  <gh issue="5836">

* [ ] **PR #3** — Prometheus end-to-end: add `WebConfigFileFields.BasicAuthUsers`; extend `pkg/webconfig` to emit `basic_auth_users:`; reconcile the operator-managed Secret per Prometheus CR; switch probes to exec; wire reloader env + flags; inject `basic_auth:` into the Thanos sidecar http-client file; user guide; e2e tests. Closes [#4200](https://github.com/prometheus-operator/prometheus-operator/issues/4200).

  <gh issue="4200">

* [ ] **PR #4** — Alertmanager and ThanosRuler: wire the same field for the remaining two CRDs, with operator-managed Secrets, probes, and reloader for each. May be split into two PRs (AM, TR) if review scope demands.

  <gh issue="4200">
