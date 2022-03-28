# Release schedule

Following [Prometheus](https://github.com/prometheus/prometheus/blob/main/RELEASE.md) and [Thanos](https://github.com/thanos-io/thanos/blob/main/docs/release-process.md), this project aims for a predictable release schedule.

Release cadence of first pre-releases being cut is 6 weeks.

| Release | Date of first pre-release (year-month-day) | Release shepherd                        |
|---------|--------------------------------------------|-----------------------------------------|
| v0.39   | 2020-05-06                                 | Pawel Krupa (GitHub: @paulfantom)       |
| v0.40   | 2020-06-17                                 | Lili Cosic (GitHub: @lilic)             |
| v0.41   | 2020-07-29                                 | Sergiusz Urbaniak (GitHub: @s-urbaniak) |
| v0.42   | 2020-09-09                                 | Matthias Loibl (GitHub: @metalmatze)    |
| v0.43   | 2020-10-21                                 | Simon Pasquier (GitHub: @simonpasquier) |
| v0.44   | 2020-12-02                                 | Pawel Krupa (GitHub: @paulfantom)       |
| v0.45   | 2021-01-13                                 | Lili Cosic (GitHub: @lilic)             |
| v0.46   | 2021-02-24                                 | Sergiusz Urbaniak (GitHub: @s-urbaniak) |
| v0.47   | 2021-04-07                                 | Simon Pasquier (GitHub: @simonpasquier) |
| v0.48   | 2021-05-19                                 | Matthias Loibl (GitHub: @metalmatze)    |
| v0.49   | 2021-06-30                                 | Pawel Krupa (GitHub: @paulfantom)       |
| v0.50   | 2021-08-11                                 | Pawel Krupa (GitHub: @paulfantom)       |
| v0.51   | 2021-09-22                                 | Simon Pasquier (GitHub: @simonpasquier) |
| v0.52   | 2021-11-03                                 | Filip Petkovski (Github: @fpetkovski)   |
| v0.53   | 2021-12-15                                 | Simon Pasquier (GitHub: @simonpasquier) |
| v0.54   | 2022-01-26                                 | Filip Petkovski (Github: @fpetkovski)   |
| v0.55   | 2022-03-09                                 | Philip Gough (GitHub: @PhilipGough)     |
| v0.56   | 2022-04-20                                 | Simon Pasquier (GitHub: @simonpasquier) |
| v0.57   | 2022-06-01                                 | **searching for volunteer**             |
| v0.58   | 2022-07-13                                 | **searching for volunteer**             |

# How to cut a new release

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

Entries in the `CHANGELOG.md` are meant to be in this order:

* `[CHANGE]`
* `[FEATURE]`
* `[ENHANCEMENT]`
* `[BUGFIX]`

Create a PR for the changes to be reviewed.

## Publish the new release

For new minor and major releases, create the `release-<major>.<minor>` branch starting at the PR merge commit.
Push the branch to the remote repository with

```
git push origin release-<major>.<minor>
```

From now on, all work happens on the `release-<major>.<minor>` branch.

Tag the new release with a tag named `v<major>.<minor>.<patch>`, e.g. `v2.1.3`. Note the `v` prefix. Tag also the `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` module with `pkg/apis/monitoring/v<major>.<minor>.<patch>` and the `github.com/prometheus-operator/prometheus-operator/pkg/client` module with `pkg/client/v<major>.<minor>.<patch>`. You can do the tagging on the commandline:

```bash
tag="v$(< VERSION)"
git tag -s "${tag}" -m "${tag}"
git tag -s "pkg/apis/monitoring/${tag}" -m "pkg/apis/monitoring/${tag}"
git tag -s "pkg/client/${tag}" -m "pkg/client/${tag}"
git push origin "${tag}" "pkg/apis/monitoring/${tag}" "pkg/client/${tag}"
```

Signed tag with a GPG key is appreciated, but in case you can't add a GPG key to your Github account using the following [procedure](https://help.github.com/articles/generating-a-gpg-key/), you can replace the `-s` flag by `-a` flag of the `git tag` command to only annotate the tag without signing.

Our CI pipeline will automatically push the container images to [quay.io](https://quay.io/organization/prometheus-operator) and [ghcr.io](https://github.com/prometheus-operator/prometheus-operator/pkgs/container/prometheus-operator)

Go to https://github.com/prometheus-operator/prometheus-operator/releases/new, associate the new release with the before pushed tag, paste in changes made to `CHANGELOG.md` and click "Publish release".

For patch releases, submit a pull request to merge back the release branch into the `main` branch.

## Update website

Bump the operator's version in the [website](https://github.com/prometheus-operator/website/blob/main/data/prometheusOperator.json) repository.

Take a breath. You're done releasing.
