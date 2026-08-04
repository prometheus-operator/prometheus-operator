# Agents Guide for Prometheus Operator

This document captures patterns and preferences observed from maintainer reviews
of recently merged pull requests. Use it to align your contributions with what
maintainers expect.

---

## Project Overview

Prometheus Operator uses Kubernetes Custom Resource Definitions (CRDs) to
declaratively manage Prometheus, Alertmanager, Thanos Ruler and associated
monitoring infrastructure. It automates tasks such as deploying and scaling
monitoring workloads, generating scrape and alerting configurations, and
managing the lifecycle of the underlying StatefulSets and Secrets.

The project exposes CRDs across three API versions (`v1`, `v1alpha1`,
`v1beta1`) and ships three binaries:

- **`operator`** — the main controller that reconciles monitoring CRDs.
- **`prometheus-config-reloader`** — a sidecar that watches for configuration
  changes and triggers Prometheus/Alertmanager reloads.
- **`admission-webhook`** — validates and defaults monitoring CRDs on admission.

The operator manages the following CRDs (under the `monitoring.coreos.com` API group):

| CRD                  | API Version | Description                                        |
|----------------------|-------------|----------------------------------------------------|
| `Prometheus`         | `v1`        | Manages Prometheus server StatefulSets.            |
| `PrometheusAgent`    | `v1alpha1`  | Manages Prometheus in agent mode.                  |
| `Alertmanager`       | `v1`        | Manages Alertmanager clusters.                     |
| `ThanosRuler`        | `v1`        | Manages Thanos Ruler instances.                    |
| `ServiceMonitor`     | `v1`        | Defines how services should be scraped.            |
| `PodMonitor`         | `v1`        | Defines how pods should be scraped.                |
| `Probe`              | `v1`        | Defines blackbox probing targets.                  |
| `ScrapeConfig`       | `v1alpha1`  | Defines custom scrape configurations.              |
| `PrometheusRule`     | `v1`        | Defines Prometheus alerting and recording rules.   |
| `AlertmanagerConfig` | `v1alpha1`  | Defines Alertmanager routing and receiver configs. |

Many files in this repository (CRDs, client-go libraries, bundle manifests, docs)
are auto-generated — never edit files matching `zz_generated.*\.go` or generated
manifests by hand. Always run `make generate` and commit the generated changes
before submitting a pull request.

---

## Commit Message Format

Messages must follow `<subsystem>: <what changed>`, with an optional body
explaining why. The subject line should be no longer than 70 characters.
Wrap the body at 80 characters.

```text
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

Examples from merged commits:

```text
feat: migrate retention options to config file
fix: drop targets for inactive shards
operator: fix dropped gzip Close errors in GzipConfig and GunzipConfig
alertmanager: return error on invalid SMTP smarthost format
pkg/prometheus: validate Probe static target labels
docs: clarify ServiceMonitor port vs targetPort
chore: update default Alertmanager version to v0.33.0
chore(api): enable notimestamp KAL linter
ci: retrigger E2E after flaky ThanosRulerStateless timeout
test: add unit test for Alertmanager
refactor(crd): refactoring resource.Quantity validate
*: modernize Go code
```

Common subsystem prefixes: `prometheus`, `alertmanager`, `thanos`, `scrapeconfig`,
`operator`, `admission`, `reloader`, `pkg/<name>`, `docs`, `chore`, `feat`, `fix`,
`test`, `refactor`, `build(deps)`, `ci`. Use `*` when the change spans many packages.

---

## Commits

- Each commit must compile and pass tests independently.
- Keep commits small and focused. Do not bundle unrelated changes in one commit.
  If a refactor is necessary, do it in a separate commit.
- Sign off every commit with `git commit -s` to satisfy the DCO requirement.
- Commits must be verified (GPG or SSH signed). Configure commit signing with
  `git config commit.gpgsign true`. See the
  [GitHub docs on signing commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
  for setup instructions.

---

## Releases

The project follows a 6-week release cycle. Release commits use the `chore:` prefix:

```text
chore: bump go dependencies before v0.92.0
chore: cut v0.92.0
```

A release shepherd is assigned for each release and is responsible for the
entire release series including patch releases. See [RELEASE.md](RELEASE.md)
for the full schedule and process.

---

## CHANGELOG

The CHANGELOG uses the following prefixes. When adding an entry, include the
PR number at the end.

```text
[CHANGE]      breaking or behavioural change
[FEATURE]     new capability
[ENHANCEMENT] improvement to existing behaviour
[BUGFIX]      bug fix
```

Combined prefixes like `[CHANGE/BUGFIX]` are acceptable when a fix also changes
behaviour. Example:

```text
* [BUGFIX] Fix goroutine leak and data race in `pollBasedListerWatcher`. #8593
```

---

## Code Style

- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
  and the formatting/style section of
  [Go: Best Practices for Production Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).
- All Go source files must carry the Apache 2.0 license header. Run
  `make fix-license` to add missing headers.
- Import ordering is enforced by `gci` via golangci-lint with three sections:
  standard library, third-party, then project-local
  (`github.com/prometheus-operator/prometheus-operator`).
- The `importas` linter enforces canonical import aliases for Kubernetes and
  project packages. Key aliases:
  - `corev1` for `k8s.io/api/core/v1`
  - `metav1` for `k8s.io/apimachinery/pkg/apis/meta/v1`
  - `apierrors` for `k8s.io/apimachinery/pkg/api/errors`
  - `monitoringv1` for the project's `pkg/apis/monitoring/v1`
  - `monitoringv1alpha1` / `monitoringv1beta1` for alpha and beta APIs
- Use `fmt.Errorf` or standard `errors` for error handling. The `github.com/pkg/errors`
  package is forbidden by `depguard`.
- Use the `slices` package instead of `sort`.
- Run `make check` (which runs both `make check-golang` and `make check-api`)
  before submitting. The project uses `golangci-lint` with linters including
  `depguard`, `godot`, `importas`, `misspell`, `revive`, `testifylint`,
  `unconvert`, `modernize` and others.
- The `godot` linter requires comments to end with a period.
- Use `//nolint:linter1[,linter2,...]` sparingly; prefer fixing the code.

---

## CRD / API Design

When designing or extending Custom Resource Definitions:

- Follow the [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
  and [API changes guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md).
- API stability policy:
  - **Alpha** (`v1alpha1`, `v1alpha2`): breaking changes may be allowed.
  - **Beta** (`v1beta1`, `v1beta2`): backward-incompatible changes may be allowed
    but forward compatibility must be preserved.
  - **Stable** (`v1`): backward and forward compatibility must be preserved.
- The `golangci-kube-api-linter` (KAL) enforces Kubernetes API conventions on
  CRD types in `pkg/apis/monitoring/`. Run `make check-api` to validate.
  See `.golangci-kal.yml` for the full list of enabled rules.

---

## Tests

- Bug fixes require a test that reproduces the bug.
- New behaviour or API changes require unit and/or e2e tests.
- Golden files are used extensively for testing generated Prometheus configs.
  Update them with `make test-unit-update-golden` when modifying config generation.
- Golden files in `pkg/prometheus/testdata/` are validated with `promtool check config`.

### Running Tests

```bash
# Unit tests (short mode).
make test-unit

# All tests including long-running.
make test-long

# End-to-end tests (requires a KinD cluster).
make test-e2e

# Targeted e2e tests.
make test-e2e-prometheus
make test-e2e-alertmanager
make test-e2e-thanos-ruler
make test-e2e-feature-gates
```

Run specific tests:

```bash
go test -run ^TestPodLabelsAnnotations$ ./pkg/prometheus/server
TEST_RUN_ARGS="-run TestPrometheusRuleCRDValidation/valid-rule-names" make test-e2e-prometheus
```

---

## Pull Requests

Every PR description must include:

- A concise description of the change and its motivation.
- A **type of change** checkbox (`CHANGE`, `FEATURE`, `BUGFIX`, `ENHANCEMENT`
  or `NONE`).
- A `release-note` fenced code block with a one-line user-facing summary.
  If there is no user-facing change, leave the block empty.

```text
```release-note
Add metadataConfig field to the Prometheus CRD for configuring how remote-write sends metadata information.
```

```

- Use GitHub closing keywords so linked issues close automatically on merge
  (e.g. `Fixes #8243`).
- Do not include unrelated changes — make a separate PR instead.
- If a PR is large, split it into preparatory and follow-up PRs and reference
  them with "Part of #NNNN" or "Depends on #NNNN".

---

## Documentation Changes

- When changing behaviour, update any relevant text in the `Documentation/`
  directory.
- API reference docs are auto-generated from CRD types — update the type
  comments, not the generated markdown.
- Use `make docs` to format and validate markdown files and links.
- Use `make check-docs` to verify formatting without modifying files.
- Proposals for new features go in `Documentation/proposals/` using the
  template at `Documentation/proposals/template.md`.

---

## CI Pipeline

PRs are validated by the following CI checks:

- **checks** — runs `make --always-make format generate && git diff --exit-code`,
  linting (`make check-golang`, `make check-api`), `make check-metrics`,
  `make tidy`, documentation formatting, and operator build.
- **unit** — runs `make test-unit` and `make test-long`.
- **e2e** — runs end-to-end tests on a KinD cluster.
- **spell-check** — checks for spelling mistakes.
- **actionlint** — lints GitHub Actions workflow files.

Before submitting, ensure:

```bash
make --always-make format generate && git diff --exit-code  # regenerate and verify no uncommitted changes
make check             # run golangci-lint and kube-api-linter
make test-unit         # run unit tests
```

---

## AI Use Policy

The project allows AI tools for contributions but requires:

- Review the change yourself before submitting.
- Ensure you can explain the why, what and how without AI assistance.
- When AI contributed significant parts, disclose it in the PR description or
  commit message.
- Avoid verbose AI-generated descriptions — keep PRs concise and focused.
- Do not submit changes unrelated to the PR's purpose.

---

## Local Development

1. Start a Kubernetes cluster (KinD recommended):

   ```bash
   kind create cluster
   ```

2. Run the operator locally:

   ```bash
   ./scripts/run-external.sh -c
   ```

---
