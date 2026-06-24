# Agents Guide for Prometheus Operator

This document captures patterns and preferences observed from maintainer reviews
of recently merged pull requests. Use it to align your contributions with what
maintainers expect.

---

## Project Overview

Prometheus Operator uses Kubernetes CRDs to declaratively manage Prometheus,
Alertmanager, Thanos Ruler and related monitoring infrastructure. See the
[introduction](https://prometheus-operator.dev/docs/getting-started/introduction/)
and [design](https://prometheus-operator.dev/docs/getting-started/design/)
docs for architecture details.

The project ships three binaries: `operator`, `prometheus-config-reloader`
and `admission-webhook`. CRDs are defined under the `monitoring.coreos.com`
API group across `v1`, `v1alpha1` and `v1beta1` versions.

Many files (CRDs, client-go libraries, bundle manifests, docs) are
auto-generated — never edit `zz_generated.*` files or generated manifests
by hand. Always run `make generate` and commit the generated changes before
submitting a pull request.

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
```

Common subsystem prefixes: `prometheus`, `alertmanager`, `thanos`, `scrapeconfig`,
`operator`, `admission`, `reloader`, `pkg/<name>`, `docs`, `chore`, `feat`, `fix`,
`test`, `refactor`, `build(deps)`, `ci`. Use `*` when the change spans many packages.

---

## Commits

- Each commit must compile and pass tests independently.
- Keep commits small and focused. Do not bundle unrelated changes in one commit.
  If a refactor is necessary, do it in a separate PR when possible.
- Sign off every commit with `git commit -s` to satisfy the DCO requirement.
- Verified (GPG signed) commits are appreciated.

---

## CHANGELOG

The CHANGELOG uses the following prefixes in this order. When adding an entry,
include the PR number at the end.

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

- All Go source files must carry the Apache 2.0 license header. Run
  `make fix-license` to add missing headers.
- Run `make check` before submitting. This runs `golangci-lint` and the
  kube-api-linter — all coding conventions (import ordering, import aliases,
  forbidden packages, comment style, etc.) are enforced by the linter
  configuration in `.golangci.yml`. Fix violations rather than suppressing
  them with `//nolint`.

---

## CRD / API Design

See the [API changes](https://prometheus-operator.dev/docs/community/contributing/#changes-to-the-apis)
section in the contributing guide for conventions and stability guarantees.
Run `make check-api` to validate CRD types against Kubernetes API conventions
(configuration is in `.golangci-kal.yml`).

---

## Tests

- Bug fixes require a test that reproduces the bug.
- New behaviour or API changes require unit and/or e2e tests.
- Golden files are used for testing generated configs — update them with
  `make test-unit-update-golden` when modifying config generation.
- Run `make test-unit` for unit tests. See [TESTING.md](TESTING.md) for the
  full testing guide including e2e tests.

---

## Pull Requests

Every PR description must include:

- A concise description of the change and its motivation.
- A **type of change** checkbox (`CHANGE`, `FEATURE`, `BUGFIX`, `ENHANCEMENT`
  or `NONE`).
- A `release-note` fenced code block with a one-line user-facing summary.
  If there is no user-facing change, leave the block empty.

````text
```release-note
Add metadataConfig field to the Prometheus CRD for configuring how remote-write sends metadata information.
```
````

- Use GitHub closing keywords so linked issues close automatically on merge
  (e.g. `Fixes #8243`).
- Do not include unrelated changes — make a separate PR instead.
- If a PR is large, split it into preparatory and follow-up PRs and reference
  them with "Part of #NNNN" or "Depends on #NNNN".

---

## Pre-submission Checklist

```bash
make --always-make format generate && git diff --exit-code  # regenerate and verify no uncommitted changes
make check             # run golangci-lint and kube-api-linter
make test-unit         # run unit tests
```

---

## AI Use Policy

See the [AI use policy](CONTRIBUTING.md#ai-use-policy) in the contributing guide.

---
