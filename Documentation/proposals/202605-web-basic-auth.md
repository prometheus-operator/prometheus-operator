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

> TL;DR: Expose Prometheus's existing `basic_auth_users` web-config setting through the `Prometheus`, `Alertmanager` and `ThanosRuler` CRDs by adding two new fields to the existing `WebConfigFileFields` struct: an optional `basicAuthUsers []BasicAuth` list for external users, and a required (when basic auth is in use) `internalUser` field that the resource owner populates from a Secret. The internal user authenticates the in-pod callers — liveness/readiness probes, `prometheus-config-reloader`, and the Thanos sidecar — that would otherwise break under basic auth.

## Why

Prometheus, Alertmanager and ThanosRuler all support HTTP basic authentication on their web servers via the `--web.config.file` flag. The operator already builds a web-config Secret for TLS settings but does not emit a `basic_auth_users:` section, so users who want to protect the web UI/API of these workloads have no operator-managed path.

Today's workarounds:

* Mutual TLS — already supported via `WebTLSConfig`, but heavyweight, requires per-client cert management.
* HTTP reverse-proxy sidecar with `spec.listenLocal: true` — running an authenticating proxy in the same Pod and exposing only it. Reported as a working approach by Hades32 in [#4652](https://github.com/prometheus-operator/prometheus-operator/pull/4652#issuecomment-1116810050). Adds an extra container per Pod to run and maintain.
* Ingress-level auth (typically ingress-nginx) — basic auth terminated at the cluster edge.

We do not have data on the relative popularity of these patterns.

### Proxy-based workarounds share a structural limitation

The reverse-proxy sidecar pattern, ingress-level auth, and similar approaches all put an authenticating proxy in front of an unauthenticated Prometheus. They differ in operational shape — sidecar vs. cluster-wide ingress vs. external load balancer — but share a common limitation: the Prometheus workload itself remains unauthenticated. Anything with network reach to the Pod or its ClusterIP — another workload, a compromised sidecar, a Pod with `hostNetwork`, a lateral attacker who has obtained any in-cluster credential — can read every metric, query the API, and (for Alertmanager) silence or fire alerts. Authenticating only at the perimeter leaves the data plane open.

Native basic auth on the workload is complementary to perimeter authentication, not redundant: an attacker who has bypassed the perimeter still has to present valid credentials to reach Prometheus's API.

One specific data point worth noting: ingress-nginx, which is widely used as the perimeter-auth component in this pattern, is itself being [retired in March 2026](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/), so users relying on it will need a different solution regardless.

Basic auth is not, on its own, a complete security solution. Operators are encouraged to use layered defenses: NetworkPolicies to narrow which Pods can reach the Prometheus service, TLS (already supported via `WebTLSConfig`) to protect credentials in flight, and so on. This proposal makes basic auth one of those layers, not the only one.

### Pitfalls of the current solutions

Two community PRs (#4652 in 2022, #4942 in 2022) tried to add native basic-auth support to the operator and were closed unfinished. Both reached the same wall, not on the API design but on the second-order effects: once basic auth is on, the kubelet probes start failing, the `prometheus-config-reloader` cannot POST to `/-/reload`, and the Thanos sidecar cannot query `/api/v1/status/buildinfo`. Neither PR shipped a complete answer for those.

A separate workaround — supplying a hand-rolled web-config Secret via `spec.secrets` and `additionalArgs: --web.config.file=…` — is rejected by the operator because `web.config.file` is a managed argument. Users sometimes try this and are surprised when it fails.

This proposal addresses the operator-owned callers head-on so that turning on basic auth does not break any operator-managed plumbing, and avoids the managed-argument trap by giving users a first-class field.

## Goals

* Add `basicAuthUsers []BasicAuth` and `internalUser` fields to `WebConfigFileFields`, applicable to `Prometheus`, `Alertmanager` and `ThanosRuler`.
* Reuse the existing `SecretKeySelector` pattern for credential references (no plaintext in the CR).
* Keep all operator-internal callers (probes, config reloader, Thanos sidecar) functional when basic auth is enabled, transparently to the user.
* Do not require the operator to generate or persist any credential material itself; credentials come from user-supplied Secrets only.

## Non-Goals

* Updates to kube-prometheus's self-scrape ServiceMonitor — separate repo, follow-up PR.
* Updates to the kube-prometheus-stack helm chart's `values.yaml` plumbing — separate repo, follow-up PR.
* Grafana datasource credential plumbing — out of scope; helm-chart concern.
* Inline plaintext credentials in the CR — explicitly not supported. All credentials live in Kubernetes Secrets and are referenced via `SecretKeySelector`.
* Operator-generated credentials — the operator never creates usernames or passwords; the resource owner supplies them.
* Operator-side bcrypt hashing of *external* user passwords — `basicAuthUsers` entries always supply bcrypt directly. The operator does bcrypt the internal user's plaintext when no `passwordBcrypt` is supplied, but the result is cached so reconciles where the plaintext is unchanged do not re-bcrypt.
* Removing the existing `spec.secrets` + `additionalArgs` workaround — the managed-argument guard continues to reject it; the new fields are the documented path.
* Waiting for [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151) (path exclusion). Useful when it lands but we are not blocking on it.

## Audience

* Operators of Prometheus/Alertmanager/ThanosRuler who need HTTP basic auth on the web server without standing up a sidecar proxy or an ingress.
* Helm chart authors (kube-prometheus, kube-prometheus-stack) who want a stable contract for passing credentials through to the workload.

## How

### API change

Add two fields to the existing `WebConfigFileFields`:

```go
// pkg/apis/monitoring/v1/types.go

type WebConfigFileFields struct {
	TLSConfig  *WebTLSConfig  `json:"tlsConfig,omitempty"`
	HTTPConfig *WebHTTPConfig `json:"httpConfig,omitempty"`

	// basicAuthUsers configures additional HTTP basic-auth users for the web
	// server, intended for external callers (humans, Grafana, other scrape
	// clients). The Password Secret value MUST be a bcrypt hash (the format
	// exporter-toolkit requires). Supported prefixes are $2a$, $2b$, $2y$.
	// Plaintext passwords are not accepted; the operator will not hash these.
	//
	// Username and password references may point at different keys in the
	// same Secret. Usernames must be unique across entries and must not
	// collide with internalUser.
	//
	// basicAuthUsers may only be set when internalUser is also set. Setting
	// basicAuthUsers without internalUser is a reconcile error, because the
	// in-pod callers (probes, config-reloader, Thanos sidecar) would have
	// nothing to authenticate with.
	// +optional
	BasicAuthUsers []BasicAuth `json:"basicAuthUsers,omitempty"`

	// internalUser configures the basic-auth credential used by the in-pod
	// callers (kubelet probes, prometheus-config-reloader, Thanos sidecar)
	// to authenticate back to the workload's web server. When set, basic
	// auth is enabled on the web server.
	//
	// The user is also added to the rendered web-config, so external clients
	// can authenticate as this user too if they know the plaintext.
	// +optional
	InternalUser *WebInternalUser `json:"internalUser,omitempty"`
}

// WebInternalUser configures the in-pod-caller basic-auth credential.
type WebInternalUser struct {
	// username is a reference to a Secret key whose value is the username
	// for the internal user.
	Username v1.SecretKeySelector `json:"username"`

	// password is a reference to a Secret key whose value is the plaintext
	// password for the internal user. The plaintext is mounted into the
	// in-pod callers' containers and used to authenticate to the web server.
	Password v1.SecretKeySelector `json:"password"`

	// passwordBcrypt is an optional reference to a Secret key whose value is
	// a bcrypt hash of password. When set, the operator passes the value
	// through verbatim into the web-config after verifying that it matches
	// the supplied plaintext. On verification mismatch, the operator falls
	// back to computing bcrypt from password and emits a Warning event.
	//
	// When not set, the operator computes bcrypt of password at reconcile
	// time and caches the result keyed on a hash of the plaintext. The
	// operator re-bcrypts only when the plaintext changes.
	// +optional
	PasswordBcrypt *v1.SecretKeySelector `json:"passwordBcrypt,omitempty"`
}
```

`SecretKeySelector` is the established pattern for credential references in the operator's existing auth fields (`ServiceMonitor.basicAuth`, `RemoteWriteSpec.basicAuth`, etc.). Keeping it here means all credentials live in Kubernetes Secrets — never in the CR — and both delivery paths users have asked for compose naturally:

* Helm-values path (e.g. kube-prometheus-stack): the chart renders one or more Secrets from `values.yaml` content and the CR references them.
* External secret path (e.g. External Secrets Operator): the user/system creates the Secret(s) directly and the CR references them.

The operator does not need to know which path produced the Secret.

### Generated web-config

`pkg/webconfig/config.go` already emits `tls_server_config:` and `http_server_config:` to the mounted web-config Secret. We extend the generator to also emit:

```yaml
basic_auth_users:
  <internal-username>: <bcrypt-hash>
  <external-username-1>: <bcrypt-hash>
  <external-username-2>: <bcrypt-hash>
```

The internal user's entry (when `internalUser` is set) is always present. External entries come from `basicAuthUsers`. Username collisions — either duplicates within `basicAuthUsers` or any `basicAuthUsers` entry whose username equals `internalUser.username` — are rejected at reconcile time with a clear error event (Prometheus would accept the YAML map but silently override).

### Modes of operation

| Inputs                                          | Result                                                                                                                                                                             |
|-------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Neither `internalUser` nor `basicAuthUsers` set | No basic auth. All components run unauthenticated, identical to today.                                                                                                             |
| `internalUser` only                             | Single-user mode. The internal user is the sole entry in the web-config. External clients can authenticate as that user; in-pod callers use that user's plaintext to authenticate. |
| `internalUser` + `basicAuthUsers`               | Multi-user mode. The web-config contains all entries. External clients can authenticate as any of them; in-pod callers use *only* `internalUser`.                                  |
| `basicAuthUsers` only, no `internalUser`        | Reconcile error. Probes, reloader, and Thanos sidecar would have nothing to authenticate with.                                                                                     |

### Internal user

The internal user is the credential the in-pod callers (kubelet probes, `prometheus-config-reloader`, Thanos sidecar) use to authenticate back to the workload's web server. It is also a regular entry in the web-config — any caller that knows its plaintext can authenticate as it.

The resource owner supplies the internal user via three Secret key references:

```yaml
spec:
  web:
    internalUser:
      username:       { name: my-secret, key: user }       # required
      password:       { name: my-secret, key: plaintext }  # required, plaintext
      passwordBcrypt: { name: my-secret, key: bcrypt }     # optional
```

`username` is the username for the internal user. `password` is the plaintext form; the operator mounts it into the in-pod callers' containers via a Secret volume so they can send it in `Authorization: Basic` headers. `passwordBcrypt`, if set, is a bcrypt hash of the plaintext.

**Default behavior** (when `passwordBcrypt` is not set): the operator computes bcrypt of the plaintext at reconcile time and renders it into the web-config. The result is cached in process memory, keyed on a hash of the plaintext, so reconciles where the plaintext is unchanged reuse the cached bcrypt and produce byte-identical web-config Secrets. The operator re-bcrypts only on plaintext rotation. The cache is rebuilt on operator restart (one bcrypt operation per workload after restart, then steady state).

**Pass-through behavior** (when `passwordBcrypt` is set): the operator verifies that the supplied bcrypt matches the plaintext using `bcrypt.CompareHashAndPassword`, cached on first verification. On match, the user-supplied bcrypt is written verbatim into the web-config; the operator performs no further cryptographic operations on this credential. On mismatch, the operator emits a Warning event on the CR ("`internalUser.passwordBcrypt` does not match `internalUser.password`; falling back to operator-derived bcrypt") and logs the same message, then falls back to default behavior — bcrypts the plaintext and uses that. The system stays functional throughout; a mismatched bcrypt is loud but non-fatal.

The pass-through behavior lets organizations that want to minimize operator-side cryptography opt out, while keeping the default ergonomic. The fallback on mismatch is a deliberate choice: a stale or wrong bcrypt in a user-managed Secret should not bring down probes or block reloads.

### Liveness/readiness/startup probes

When basic auth is on, the operator switches from `HTTPGetAction` probes to `ExecAction` probes built by `pkg/operator/prober.go`. The existing helper already shells out to `curl` (or `wget` as fallback); we extend it to add basic-auth arguments. The username and password are read from a file at probe-execution time (not from an environment variable), so that rotation of the `internalUser` Secret propagates to running pods without a restart — kubelet refreshes mounted Secret contents on its sync interval, and the probe reads on each invocation. The plaintext password never appears in the pod spec or argv.

```sh
# generated probe command (conceptual)
sh -c 'U=$(cat /etc/prometheus/internal-auth/username); P=$(cat /etc/prometheus/internal-auth/password); \
       if [ -x "$(command -v curl)" ]; then exec curl --fail -u "$U:$P" http://…/-/ready; \
       elif [ -x "$(command -v wget)" ]; then exec wget -q -O /dev/null --user="$U" --password="$P" http://…/-/ready; \
       else exit 1; fi'
```

This works on every supported Prometheus / Alertmanager / ThanosRuler version regardless of whether [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151) has merged. If/when that lands, we can simplify to path-excluded `HTTPGetAction` probes in a follow-up — the internal user is still useful for the reloader and sidecar.

#### Interaction with the strategic-merge-patch probe workaround

[`Documentation/platform/strategic-merge-patch.md`](../platform/strategic-merge-patch.md) currently documents a workaround for this exact problem: when TLS + a hand-rolled web-config requires auth, users override the operator's `httpGet` probe via a strategic merge patch (typically to `tcpSocket`, since that bypasses authentication entirely). The merge-patch behavior was hardened in v0.91.0 ([#8427](https://github.com/prometheus-operator/prometheus-operator/pull/8427)) and remains a supported escape hatch.

This proposal does not deprecate that workaround. Interaction precedence:

* If `internalUser` is set **and** the user has not supplied a probe override → operator emits an exec probe with credentials. The probe still calls `/-/ready`, so unlike the `tcpSocket` workaround it remains accurate about WAL replay and TSDB initialization. This is a strict improvement over the documented workaround.
* If `internalUser` is set **and** the user has supplied a probe override via strategic merge patch → the user's override wins, per the existing merge-patch semantics. The operator's exec probe is replaced. Users opting out this way are responsible for handling auth themselves (e.g. by sticking with `tcpSocket`).
* The note in `strategic-merge-patch.md` about exec-probes being needed "when prioritizing credential security over granular readiness checks" is no longer the only path — users who do nothing get a granular readiness check *and* credential security. The doc will be updated in PR #3 (not invalidated; cross-referenced).

### prometheus-config-reloader

The reloader binary today has no outbound basic-auth support, so it cannot reach `/-/reload` or `/api/v1/status/runtimeinfo` once auth is on. We add two flags:

```text
--reload-basic-auth-username=<value>
--reload-basic-auth-password-file=<path>
```

The password is read from a file path pointing at the mounted `internalUser` Secret volume; it never lands in argv or in `/proc/PID/environ`. The username is non-sensitive and is passed as a flag value. Both files are auto-refreshed by kubelet when the underlying Secret rotates, and the reloader re-reads them on each authenticated request, so rotation does not require a pod restart.

When the operator detects `internalUser` is set on the workload, it adds the flags and mounts the Secret volume into the reloader container.

This change also closes [#5836](https://github.com/prometheus-operator/prometheus-operator/issues/5836) for users running the reloader against a separately-protected Prometheus.

### Thanos sidecar

The sidecar communicates with Prometheus via the YAML file referenced by `--prometheus.http-client-file`, which is built by `pkg/prometheus/server/thanos_sidecar_config.go`. We extend that builder to emit:

```yaml
basic_auth:
  username: <internal-username>
  password_file: /etc/thanos/auth/password
```

…and mount the `internalUser` Secret at that path. The thanos-sidecar http-client format already supports this natively.

**Thanos version gate.** The `--prometheus.http-client-file` flag is only supported in Thanos ≥ 0.24.0 (already tracked by the operator as `thanosSupportedVersionHTTPClientFlag`). When `internalUser` is set and the configured Thanos image is older than 0.24.0, the operator must fail reconcile with a clear error rather than silently produce a broken sidecar. The same applies if the user later downgrades the Thanos image while basic auth is enabled.

### bcrypt performance

bcrypt is intentionally slow (~50 ms per verify at the default cost of 10) and Prometheus's web handler serializes bcrypt comparisons through a global mutex. There are two places it shows up in this design — once on the Prometheus side at authentication time, and once on the operator side when computing or verifying the internal user's bcrypt.

**Prometheus-side (at authentication).** exporter-toolkit caches successful authentications — see [`web/handler.go`](https://github.com/prometheus/exporter-toolkit/blob/master/web/handler.go) and `cache.go` (LRU, size 100, key = `user + hash + plaintext`). After the first request per `(user, hash, password)` tuple, subsequent verifications are map lookups (~µs). The in-pod callers (probes, reloader, sidecar) use stable credentials, so they are cache hits for the life of the process. The bounded cold-start window — distinct first-time authentications queueing on the global mutex — affects at most a handful of callers per pod restart, totaling well under a second.

**Operator-side (when generating bcrypt for the internal user).** In default mode the operator bcrypts the internal user's plaintext at reconcile time. The result is cached in-process keyed on a hash of the plaintext, so reconciles where the plaintext is unchanged reuse the cached value and produce byte-identical web-config Secrets — no Secret churn, no etcd writes, no kubelet syncs, no auth-cache invalidation on the Prometheus side. The operator pays one ~50 ms bcrypt per workload at startup and one per plaintext rotation thereafter. In pass-through mode the cost is one `bcrypt.CompareHashAndPassword` per workload at startup and per rotation, same order of magnitude.

If cold-start latency on the Prometheus side ever turns out to be a real problem in production reports, the fallback (not pursued in v1) is to switch the in-pod callers from basic auth to mTLS, reusing the existing `WebTLSConfig` infrastructure. External callers would continue to use basic auth. This approximately doubles the implementation surface and is deferred unless needed.

### Validation and error surfaces

* CEL on the CR enforces the structural rule that `basicAuthUsers` requires `internalUser`.
* The operator validates that each referenced Secret/key exists at reconcile time using the existing `assets.StoreBuilder.AddBasicAuth` path.
* Duplicate usernames within `basicAuthUsers`, or a collision between any `basicAuthUsers[].username` value and `internalUser.username`, fail reconcile with a clear error event rather than silently dropping entries.
* `basicAuthUsers[].password` values that do not look like bcrypt (no `$2[aby]$` prefix) produce a Warning event — heuristic, surfacing the common plaintext-in-Secret mistake.
* `internalUser.password` values that *do* look like bcrypt produce a Warning event — heuristic, surfacing the inverse mistake.
* `internalUser.passwordBcrypt` mismatch with `internalUser.password` is handled at runtime (not validation): the operator emits a Warning event and falls back to default-mode bcrypting. See "Internal user" above.
* When `internalUser` is set and the configured Thanos image is older than 0.24.0, reconcile fails with an explicit version-incompatibility error.

CEL cannot inspect Secret values, so all content-based validations (bcrypt format heuristics, plaintext/bcrypt match, collision detection) happen at reconcile time as Warning events rather than at admission.

### Migration / compatibility

* New field, optional, defaults to "off". Zero behavior change for existing users.
* `WebConfigFileFields` is part of the `monitoring.coreos.com/v1` API. Per `CONTRIBUTING.md`, v1 forbids breaking backward or forward compatibility. This change is purely additive (optional field, omitempty); a CR written against the old schema validates and behaves identically against the new schema, and a CR written against the new schema is silently ignored by an older operator binary (it sees an unknown field). Compatible in both directions.
* The existing `spec.secrets` + `additionalArgs: --web.config.file=…` workaround continues to be rejected by the managed-argument guard. The proposal does not change that. Users who need full hand-rolled control retain the existing workaround; the new field is the supported path for everyone else.
* The strategic-merge-patch probe override (see "Interaction with the strategic-merge-patch probe workaround" above) continues to work and is documented as the opt-out for users who want operator-managed basic auth but their own probe definition.

### Testing and verification

* Unit tests: `pkg/webconfig/config_test.go` (YAML output golden tests), reloader flag wiring, ExecAction generation.
* e2e tests: boot a Prometheus / Alertmanager / ThanosRuler with `basicAuthUsers` set, verify probes pass, reloader can trigger a reload, scrape with correct credentials returns 200, scrape without credentials returns 401.

### Documentation impact

In addition to the per-component user-guide additions in the implementation PRs, the following existing docs are affected and will be updated in-flight (not in this proposal PR):

* `Documentation/user-guides/basic-auth.md` is currently titled "Basic auth for targets" and covers `ServiceMonitor`-side basic auth. PR #3 will extend it with a clearly-titled new section "Basic auth for the Prometheus web server" (and add Alertmanager / ThanosRuler sections in PR #4). The existing scrape-target content is unchanged. The risk of name collision is acknowledged; splitting the file into two was considered but rejected because users searching for "basic auth" benefit from finding both topics in one place.
* `Documentation/platform/strategic-merge-patch.md` will get a cross-reference noting that `spec.web.basicAuthUsers` now provides an operator-managed path that obsoletes the documented workaround for most users.
* `Documentation/platform/thanos.md`, `Documentation/platform/troubleshooting.md`, `Documentation/platform/exposing-prometheus-and-alertmanager.md` — short additions referencing the new field where relevant.
* `Documentation/api-reference/api.md` — autogenerated by `make generate`.

### Open questions

1. **Path-exclusion follow-up.** If [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151) merges and exposes `/-/healthy`/`/-/ready` without auth, do we want to make the probe-via-exec behavior optional and revert to `HTTPGetAction` probes by default? The internal user would still be needed for the reloader and Thanos sidecar regardless. Recommendation: revisit when that PR lands; no opt-out machinery in v1.

## Alternatives

1. **Wait for [exporter-toolkit#151](https://github.com/prometheus/exporter-toolkit/pull/151).** Path exclusion for `/-/healthy`, `/-/ready` would eliminate the probe problem and shrink this proposal substantially. The PR has been open since 2024 and is currently in conflict; we cannot block on it. If/when it merges we can simplify probes in a follow-up; the operator-managed user is still needed for the reloader and sidecar.

2. **Mutual TLS only.** Already supported via `WebTLSConfig`. Heavyweight for the common "I just want a password on the UI" use case. Not a replacement for basic auth.

3. **Ingress-level auth, reverse-proxy sidecar, and similar proxy-based patterns.** All share the structural limitation that the Prometheus workload itself remains unauthenticated. See "Proxy-based workarounds share a structural limitation" above.

4. **Operator-side bcrypt of every user's plaintext password.** Lets users supply plaintext for all entries in `basicAuthUsers` and have the operator hash them. Rejected for the general case because:
   * Every reconcile would produce a fresh salt, causing churn unless caches are maintained per entry.
   * The operator would have to persist its own derived state and track whether each source changed — significantly more state machinery for marginal UX gain.
   * It exposes the operator to potential security issues that maintainers don't want to deal with.

   Note: this proposal does perform operator-side bcrypt for the single `internalUser` credential in default mode (see "Internal user" above). The objections above are weaker when scoped to one credential with a cache keyed on the plaintext, and the pass-through mode lets organizations that want to avoid even this opt out.

5. **Secret-as-user-list form for `basicAuthUsers`.** An alternative shape where the CR points at a Secret by name (e.g. `basicAuthUsersSecret: prom-users`) and the operator interprets every key in that Secret as a username and the corresponding value as that user's bcrypt hash. More ergonomic for organizations using GitOps with External Secrets Operator or similar tooling, where the user list naturally lives in one external store and is synced to a single Secret — with the per-entry form, the CR (or chart `values.yaml`) needs to keep a parallel list of key names in sync with the Secret's contents.

   Rejected for v1 because:
   * It diverges from the project pattern of `SecretKeySelector` for every credential reference. Adding a second pattern for one field increases API surface and maintenance burden.
   * The same use case is addressable today via per-entry references plus chart templating, just less elegantly.
   * There is no demand signal yet; choosing between possible conventions (one-key-per-user, htpasswd format, YAML list) is premature without observed usage.

   The per-entry shape is forward-compatible with this addition: a `basicAuthUsersSecret` field can be added later as a mutually-exclusive alternative to `basicAuthUsers`, with the convention chosen based on observed usage.

6. **Sidecar nginx / custom auth proxy.** What this proposal aims to make unnecessary. Mentioned only because it is what users do today.

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

* [ ] **PR #3** — Prometheus end-to-end: add `WebConfigFileFields.BasicAuthUsers` and `WebConfigFileFields.InternalUser`; extend `pkg/webconfig` to emit `basic_auth_users:` (with cached bcrypt for the internal user and pass-through-with-verify for user-supplied bcrypt); switch probes to exec reading credentials from a mounted Secret volume; add basic-auth flags to `prometheus-config-reloader` and wire them; inject `basic_auth:` into the Thanos sidecar http-client file with the version gate; CEL validation for the structural rules; user guide; e2e tests. Closes [#4200](https://github.com/prometheus-operator/prometheus-operator/issues/4200).

  <gh issue="4200">

* [ ] **PR #4** — Alertmanager and ThanosRuler: wire the same fields for the remaining two CRDs, including probes, config reloader, and the corresponding sidecar/version gates where applicable. May be split into two PRs (AM, TR) if review scope demands.

  <gh issue="4200">
