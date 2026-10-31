---
name: release-prep
description: >-
  Prepare a Prometheus Operator release PR. Use when bumping versions,
  updating dependencies, updating CHANGELOG, or when the user mentions
  release preparation or cutting a release.
---

# Prometheus Operator Release Preparation

The project follows a **6-week release cycle**. See [RELEASE.md](../../../RELEASE.md)
for the full schedule, shepherd assignments and post-PR steps (tagging,
publishing, etc.).

> **Scope:** This skill covers minor releases only. Patch releases follow
> a different process and are not covered here — refer to RELEASE.md.

## Important

- **NEVER commit, push or create PRs without explicit human approval.**
- **NEVER push directly to `main`, `release-*`, or any protected branch.**
  Always work on a feature branch and open a PR.
- **STOP at every review gate below and wait for the user to confirm
  before proceeding.**
- Each phase depends on the previous one being merged. Do not start
  the next phase until the user confirms the prior PR has been merged.

## Commit and PR title conventions

PR titles should describe what was updated:

- Dependency bump: `chore: bump go dependencies before vX.Y.Z`
- Operand update: `chore: update default Alertmanager version to vA.B.C`
- Release cut: `chore: cut vX.Y.Z`

## Setup

Before starting, fetch the latest from upstream and create a working branch:

Run `git remote -v` to identify the remote pointing to
`prometheus-operator/prometheus-operator`. It is typically `upstream`
(fork workflow) or `origin` (direct clone). Use that remote name in the
commands below:

```bash
git fetch <remote>
git checkout -b deps-vX.Y.Z <remote>/main
```

---

## Phase 1: Dependency bump PR

Update Go dependencies across all three modules:

```bash
make update-go-deps
make tidy
```

This modifies `go.mod` and `go.sum` in the root, `pkg/apis/monitoring/`
and `pkg/client/` directories.

### Review gate 1

**STOP.** Show the user the changed files and wait for approval before
committing. Commit message and PR title:
`chore: bump go dependencies before vX.Y.Z`. Sign off with `git commit -s`.

Once the user approves, ask:
> Would you like me to create the PR using `gh pr create`, or will you
> create it manually?

If using `gh`, first verify it is authenticated by running
`gh auth status > /dev/null 2>&1` and checking the exit code. If it
fails, ask the user to run `gh auth login` before proceeding. Push the
branch, then show the full `gh pr create` command and wait for approval
before executing it.

**Wait for the PR to be reviewed and merged before proceeding to Phase 2.**

---

## Phase 2: Release cut PR

**Prerequisite:** Phase 1 PR must be merged. Fetch latest and create a
new branch (use the same remote identified during Setup):

```bash
git fetch <remote>
git checkout -b cut-vX.Y.Z <remote>/main
```

1. **Update operand versions** — bump default Prometheus, Alertmanager and
   Thanos versions in `pkg/operator/defaults.go` if newer versions exist.

### Review gate 2

**STOP.** Show the user the version changes and confirm they are correct
before continuing.

1. **Bump VERSION** — edit the `VERSION` file in the repo root.

1. **Regenerate files** — this updates CRDs, bundle, jsonnet, docs and
   example manifests (expect ~50 files to change):

   ```bash
   make clean generate
   ```

1. **Bump submodule versions in go.mod**:

   ```bash
   go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v$(< VERSION)" pkg/client/go.mod
   go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v$(< VERSION)"
   go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/client@v$(< VERSION)"
   ```

1. **Regenerate and verify** — ensure everything is consistent after
   the version bumps:

   ```bash
   make --always-make format generate && git diff --exit-code
   ```

1. **Update CHANGELOG.md** — add a new version header and entries in this
   order: `[CHANGE]`, `[FEATURE]`, `[ENHANCEMENT]`, `[BUGFIX]`. Only
   include user-facing changes. Each entry must include the PR number:

   ```text
   ## vX.Y.Z / YYYY-MM-DD

   * [CHANGE] Description of the change. #1234
   * [FEATURE] Description of the feature. #1235
   * [ENHANCEMENT] Description of the enhancement. #1236
   * [BUGFIX] Description of the fix. #1237
   ```

   Use the GitHub compare view to identify changes since the last release:
   `https://github.com/prometheus-operator/prometheus-operator/compare/vPREVIOUS...main`

### Review gate 3

**STOP.** Show the user the full diff (especially CHANGELOG.md and VERSION)
and wait for approval before committing. Commit message and PR title:
`chore: cut vX.Y.Z`. Sign off with `git commit -s`.

Once the user approves, ask:
> Would you like me to create the PR using `gh pr create`, or will you
> create it manually?

If using `gh`, first verify it is authenticated by running
`gh auth status > /dev/null 2>&1` and checking the exit code. If it
fails, ask the user to run `gh auth login` before proceeding. Push the
branch, then show the full `gh pr create` command and wait for approval
before executing it.
