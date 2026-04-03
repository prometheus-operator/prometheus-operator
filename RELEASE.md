---
weight: 606
toc: true
title: Release
menu:
    docs:
        parent: community
lead: ""
images: []
draft: false
---

# Release schedule

Following [Prometheus](https://github.com/prometheus/prometheus/blob/main/RELEASE.md) and [Thanos](https://github.com/thanos-io/thanos/blob/main/docs/release-process.md), this project aims for a predictable release schedule.

The release cycle for cutting releases is every 6 weeks

| Release | Date of release (year-month-day) | Release shepherd                          |
|---------|----------------------------------|-------------------------------------------|
| v0.92   | 2026-06-10                       | **searching for volunteer**               |
| v0.91   | 2026-04-29                       | **searching for volunteer**               |
| v0.90   | 2026-03-18                       | Jayapriya Pai (Github: @slashpai)         |
| v0.89   | 2026-02-04                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.88   | 2025-12-24                       | Jayapriya Pai (Github: @slashpai)         |
| v0.87   | 2025-11-12                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.86   | 2025-10-01                       | Jayapriya Pai (Github: @slashpai)         |
| v0.85   | 2025-08-20                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.84   | 2025-07-09                       | M Viswanath Sai (Github: @mviswanathsai)  |
| v0.83   | 2025-05-28                       | M Viswanath Sai (Github: @mviswanathsai)  |
| v0.82   | 2025-04-16                       | Jayapriya Pai (Github: @slashpai)         |
| v0.81   | 2025-03-05                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.80   | 2025-01-22                       | Jayapriya Pai (Github: @slashpai)         |
| v0.79   | 2024-12-11                       | Jayapriya Pai (Github: @slashpai)         |
| v0.78   | 2024-10-30                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.77   | 2024-09-18                       | Jayapriya Pai (Github: @slashpai)         |
| v0.76   | 2024-08-07                       | Nicolas Takashi (Github: @nicolastakashi) |
| v0.75   | 2024-06-26                       | Jayapriya Pai (Github: @slashpai)         |
| v0.74   | 2024-05-15                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.73   | 2024-04-03                       | Jayapriya Pai (Github: @slashpai)         |
| v0.72   | 2024-02-21                       | Arthur Sens (Github: @ArthurSens)         |
| v0.71   | 2024-01-10                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.70   | 2023-11-29                       | Pawel Krupa (GitHub: @paulfantom)         |
| v0.69   | 2023-10-18                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.68   | 2023-09-06                       | Arthur Sens (Github: @ArthurSens)         |
| v0.67   | 2023-07-26                       | Simon Pasquier (GitHub: @simonpasquier)   |
| v0.66   | 2023-06-14                       | Arthur Sens (Github: @ArthurSens)         |
| v0.65   | 2023-05-03                       | Philip Gough (GitHub: @PhilipGough)       |

If any of the maintainers is interested in volunteering please create a pull request against the [prometheus-operator/prometheus-operator](https://github.com/prometheus-operator/prometheus-operator) repository and propose yourself for the release series of your choice.

## Release shepherd responsibilities

The release shepherd is responsible for the entire release series of a major or minor release, including all patch releases. Some preparations should be done a few days in advance.

* We aim to keep the main branch in a working state at all times. In principle, it should be possible to cut a release from main at any time. In practice, things might not work out as nicely. A few days before the release is scheduled, the shepherd should check the state of main. Following their best judgement, the shepherd should try to expedite features/bug fixes that are still in progress but should make it into the release. On the other hand, the shepherd may hold back merging last-minute invasive and risky changes that are better suited for the next major release.
* On the date listed in the table above, the release shepherd cuts the release and creates a new branch called `release-<major>.<minor>` starting at the commit tagged for the release.
* If regressions or critical bugs are detected, they need to get fixed before cutting a new release.

See the next section for details on cutting an individual release.

## How to cut a new release

> This guide is strongly based on the [Prometheus release instructions](https://github.com/prometheus/prometheus/blob/main/RELEASE.md).

## Branch management and versioning strategy

We use [Semantic Versioning](http://semver.org/).

We maintain a separate branch for each minor release, named `release-<major>.<minor>`, e.g. `release-1.1`, `release-2.0`.

The usual flow is to merge new features and changes into the `main` branch and to merge bug fixes into the latest release branch. Bug fixes are then merged into `main` from the latest release branch. The `main` branch should always contain all commits from the latest release branch.

If a bug fix got accidentally merged into `main`, cherry-pick commits have to be created in the latest release branch, which then have to be merged back into `main`. Try to avoid that situation.

Maintaining the release branches for older minor releases happens on a best effort basis.

## Update Go dependencies

A couple of days before the release, consider submitting a PR against the `main` branch to update the Go dependencies.

```bash
make update-go-deps
make tidy
```

## Update operand versions

A couple of days before the release, update the [default versions](https://github.com/prometheus-operator/prometheus-operator/blob/f6ce472ecd6064fb6769e306b55b149dfb6af903/pkg/operator/defaults.go#L20-L31) of Prometheus, Alertmanager and Thanos if newer versions are available.

## Prepare your release

For a new major or minor release, work from the `main` branch. For a patch release, work in the branch of the minor release you want to patch (e.g. `release-0.43` if you're releasing `v0.43.2`).

Bump the version in the `VERSION` file in the root of the repository.

A number of files have to be re-generated, this is automated with the following make target:

```bash
make clean generate
```

Bump the version of the `pkg/apis/monitoring` and `pkg/client` packages in `go.mod`:

```bash
go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v$(< VERSION)" pkg/client/go.mod
go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v$(< VERSION)"
go mod edit -require "github.com/prometheus-operator/prometheus-operator/pkg/client@v$(< VERSION)"
```

Now that all version information has been updated, an entry for the new version can be added to the `CHANGELOG.md` file.

Note that CHANGELOG.md should only document changes relevant to users of prometheus-operator, including external API changes, performance improvements, and new features. Do not document changes of internal interfaces, code refactoring and clean-ups, doc changes and changes to the build process, etc.

Entries in the `CHANGELOG.md` are meant to be in this order:

* `[CHANGE]`
* `[FEATURE]`
* `[ENHANCEMENT]`
* `[BUGFIX]`

Create a PR for the changes to be reviewed.

You can use the GitHub UI to see the difference between the release branch and the latest stable release.

For example: https://github.com/prometheus-operator/prometheus-operator/compare/v0.72.0...release-0.73

Unless exception, the latest tag shouldn't contain commits that don't exist in the release branch.

## Publish the new release

For new minor and major releases, create the `release-<major>.<minor>` branch starting at the PR merge commit.
Push the branch to the remote repository with

**Note:** The remote name `origin` is assumed to be pointed to `github.com/prometheus-operator/prometheus-operator`. If you have a different remote name, use that instead of `origin`. Verify this using `git remote -v`.

```bash
git push origin release-<major>.<minor>
```

You could also create the release branch directly from Github UI as well if the current main branch HEAD is what release branch should be based on.

From now on, all work happens on the `release-<major>.<minor>` branch.

Tag the new release with a tag named `v<major>.<minor>.<patch>`, e.g. `v2.1.3`. Note the `v` prefix. Tag also the `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` module with `pkg/apis/monitoring/v<major>.<minor>.<patch>` and the `github.com/prometheus-operator/prometheus-operator/pkg/client` module with `pkg/client/v<major>.<minor>.<patch>`. You can do the tagging on the commandline:

```bash
tag="v$(< VERSION)"
git tag -s "${tag}" -m "${tag}"
git tag -s "pkg/apis/monitoring/${tag}" -m "pkg/apis/monitoring/${tag}"
git tag -s "pkg/client/${tag}" -m "pkg/client/${tag}"
git push origin "${tag}" "pkg/apis/monitoring/${tag}" "pkg/client/${tag}"
```

Signed tag with a GPG key is appreciated, but in case you can't add a GPG key to your Github account using the following [procedure](https://docs.github.com/articles/generating-a-gpg-key), you can replace the `-s` flag by `-a` flag of the `git tag` command to only annotate the tag without signing.

Once a tag is created, the `publish` Github action will push the container images to [quay.io](https://quay.io/organization/prometheus-operator) and [ghcr.io](https://github.com/prometheus-operator/prometheus-operator/pkgs/container/prometheus-operator). Wait until the [publish](https://github.com/prometheus-operator/prometheus-operator/actions/workflows/publish.yaml) workflow is complete before going to the next step.

We have observed in the past that if we create a draft release and publish it later assets are not attached correctly hence its advised to wait till all workflow jobs (at least the publish job) are completed to create the release.

Go to https://github.com/prometheus-operator/prometheus-operator/releases/new, associate the new release with the before pushed tag, paste in changes made to `CHANGELOG.md` and click "Publish release".

Once release is published, [release job](https://github.com/prometheus-operator/prometheus-operator/actions/workflows/release.yaml) will be triggered to upload assets to the newly created release.

For patch releases, submit a pull request to merge back the release branch into the `main` branch.

## Update website

Bump the operator's version in the [website](https://github.com/prometheus-operator/website/blob/main/data/prometheusOperator.json) repository.

Take a breath. You're done releasing.
