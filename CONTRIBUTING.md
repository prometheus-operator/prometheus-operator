---
weight: 601
toc: true
title: Contributing
menu:
    docs:
        parent: community
lead: ""
lastmod: "2021-03-08T08:48:57+00:00"
images: []
draft: false
description: How can I contribute to the Prometheus Operator and kube-prometheus?
date: "2021-03-08T08:48:57+00:00"
---

This project is licensed under the [Apache 2.0 license](LICENSE) and accept
contributions via GitHub pull requests. This document outlines some of the
conventions on development workflow, commit message formatting, contact points
and other resources to make it easier to get your contribution accepted.

To maintain a safe and welcoming community, all participants must adhere to the
project's [Code of Conduct](code-of-conduct.md).

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

## Email and Chat

The project currently uses the [Kubernetes Slack](https://kubernetes.slack.com):

- [#prometheus-operator](https://kubernetes.slack.com/archives/CFFDS2Z7F)
- [#prometheus-operator-dev](https://kubernetes.slack.com/archives/C01B03QCSMN)

Please avoid emailing maintainers found in the MAINTAINERS file directly. They
are very busy and read the mailing lists.

## Office Hours Meetings

The project also holds bi-weekly public meetings where maintainers,
contributors and users of the Prometheus Operator and kube-prometheus can
discuss issues, pull requests or any topic related to the projects. The
meetings happen at 11:00 UTC on Monday, check the [online
notes](https://docs.google.com/document/d/1-fjJmzrwRpKmSPHtXN5u6VZnn39M28KqyQGBEJsqUOk/edit?usp=sharing)
to know the exact dates and the connection details.

An invite is also available on the [project's public calendar](https://calendar.google.com/calendar/u/1/embed?src=c_331fefe21da6f878f17e5b752d63e19d58b1e3bb24cb82e5ac65e5fd14e81878@group.calendar.google.com&csspa=1).

## Getting Started

- Fork the repository on GitHub
- Read the [README](README.md) for build and test instructions
- Play with the project, submit bugs, submit patches!

## Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where you want to base your work (usually `main`).
- Make commits of logical units.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- Make sure the tests pass, and add any new tests as appropriate. ([Testing guidelines](TESTING.md))
- Submit a pull request to the original repository.

Many files (documentation, manifests, ...) in this repository are auto-generated. For instance, `bundle.yaml` is generated from the *Jsonnet* files in `/jsonnet/prometheus-operator`. Before submitting a pull request, make sure that you've executed `make generate` and committed the generated changes.

We also use [golangci-lint](https://golangci-lint.run/docs/) to lint the Go code (including the API definitions). Make sure to execute `make check` before creating/updating your PR.

Thanks for your contributions!

### Changes to the APIs

When designing Custom Resource Definitions (CRDs), please refer to the existing Kubernetes guidelines:

- [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md).
- [API changes](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md).

In particular, this project follows the API stability guidelines:

- For alpha API versions (e.g. `v1alpha1`, `v1alpha2`, ...), we may allow to break forward and backward compatibility (but we'll try hard to avoid it).
- For beta API versions (e.g. `v1beta1`, `v1beta2`, ...), we may allow to break backward compatibility but not forward compatibility.
- For stable API versions (e.g. `v1`), we don't allow to break backward and forward compatibility.

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```bash
scripts: add the test-cluster command

This uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38
```

The format can be described more formally as follows:

```bash
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various
Git tools.

## AI use policy

We allow the use of AI tools when contributing to the project (issues and pull
requests). At the same time, you need to be mindful of maintainers' time and
attention which is why we ask you to comply with the following guidelines.

### When using AI for learning

* Keep in mind that while AI assistants help with navigating and understanding the code base, you need to take their claims with a grain of salt.
* Using AI tools doesn't prevent you from becoming familiar with the code and the development workflow.

### When using AI for communication

* Follow the proposed template when submitting GitHub issues.
* **Avoid verbose descriptions**, provide enough information for the maintainers to understand the request but do not overwhelm them with unrelated details.

### When using AI for code contribution

* Review the change by yourself before submitting the pull request.
* Ensure that you can explain the why, what and how of the change without help from the AI tool.
* If necessary call out the parts which are unclear to you.
* When AI tools have contributed significant parts of the code change, communicate the information in the pull request's description and/or the commit message.
* Don't submit changes which are unrelated to the purpose of the pull request.
* **Avoid verbose AI-generated descriptions in PRs** keep descriptions concise, focused, and relevant to the actual changes being made.

## Local Development

If you want to run Prometheus Operator on your local environment, you can follow the steps below.

1. First start a Kubernetes cluster. We recommend [KinD](https://kind.sigs.k8s.io/) because it is lightweight (it can run on small notebooks) and this is what the project's CI uses. [MiniKube](https://minikube.sigs.k8s.io/docs/start/) is also another option.

2. Run the utility script [scripts/run-external.sh](scripts/run-external.sh), it will check all the requirements and run your local version of the Prometheus Operator on your Kind cluster.

```bash
./scripts/run-external.sh -c
```

3. You should now be able to see the logs from the operator in your terminal. The Operator is successfully running in your local system and can be debugged, checked for behaviour etc.

Similarly, if you work on a specific branch, you can run the `scripts/run-external.sh` script in this branch to deploy it.

## Proposal Process

The Prometheus Operator project accepts proposals for new features,
enhancements and design documents. The document should be created in the
`Documentation/proposals` directory using the template below, prefixed by
`<YEAR><MONTH>-` and submitted in the form of a GitHub Pull Request.

The process is adopted from the Thanos community.

```markdown mdox-exec="cat Documentation/proposals/template.md"
## Your Proposal Title

* **Owners:**
  * `<@author: single champion for the moment of writing>`
* **Status:**
  * `<Accepted/Rejected/Implemented>`
* **Related Tickets:**
  * `<JIRA, GH Issues>`
* **Other docs:**
  * `<Links…>`

> TL;DR: Give a summary of what this document is proposing and what components it is touching.
>
> *For example: This design doc is proposing a consistent design template for “example.com” organization.*

## Why

Provide a motivation behind the change proposed by this design document, give context.

*For example: It’s important to clearly explain the reasons behind certain design decisions in order to have a
consensus between team members, as well as external stakeholders.
Such a design document can also be used as a reference and for knowledge-sharing purposes.
That’s why we are proposing a consistent style of the design document that will be used for future designs.*

### Pitfalls of the current solution

What specific problems are we hitting with the current solution? Why is it not enough?

*For example: We were missing a consistent design doc template, so each team/person was creating their own.
Because of inconsistencies, those documents were harder to understand, and it was easy to miss important sections.
This was causing certain engineering time to be wasted.*

## Audience

Provide the target audience for this change.

## Goals

Goals and use cases for the solution as proposed in [How](#how):

* Allow easy collaboration and decision making on design ideas.
* Have a consistent design style that is readable and understandable.
* Have a design style that is concise and covers all the essential information.

## Non-Goals

* Move old designs to the new format.
* Not doing X,Y,Z.

## How

Explain the full overview of the proposed solution. Some guidelines:

* Make it concise and **simple**; put diagrams; be concrete, avoid using “really”, “amazing” and “great” (:
* How will you test and verify?
* How will you migrate users, without downtime. How do we solve incompatibilities?
* What open questions are left? (“Known unknowns”)

## Alternatives

This section should state potential alternatives.
Highlight the objections the reader should have towards your proposal as they read it.
Tell them why you still think you should take this path.

1. This is why not solution Z...

## Action Plan

The tasks to do in order to migrate to the new idea.

* [ ] Task one

  <gh issue="">

* [ ] Task two

  <gh issue="">

  ...
```
