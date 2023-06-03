---
weight: 120
toc: true
title: Contributing
menu:
    docs:
        parent: prologue
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

# Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

# Email and Chat

The project currently uses the [Kubernetes Slack](https://slack.k8s.io/):
- [#prometheus-operator](https://kubernetes.slack.com/archives/CFFDS2Z7F)
- [#prometheus-operator-dev](https://kubernetes.slack.com/archives/C01B03QCSMN)

Please avoid emailing maintainers found in the MAINTAINERS file directly. They
are very busy and read the mailing lists.

# Office Hours Meetings

The project also holds bi-weekly public meetings where maintainers,
contributors and users of the Prometheus Operator and kube-prometheus can
discuss issues, pull requests or any topic related to the projects. The
meetings happen at 09:00 UTC on Monday, check the [online
notes](https://docs.google.com/document/d/1-fjJmzrwRpKmSPHtXN5u6VZnn39M28KqyQGBEJsqUOk/edit?usp=sharing)
to know the exact dates and the connection details.

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
- Make sure the tests pass, and add any new tests as appropriate.
- Submit a pull request to the original repository.

Many files (documentation, manifests, ...) in this repository are auto-generated. For instance, `bundle.yaml` is generated from the *Jsonnet* files in `/jsonnet/prometheus-operator`. Before submitting a pull request, make sure that you've executed `make generate` and committed the generated changes.

Thanks for your contributions!

### Changes to the APIs

When designing Custom Resource Definitions (CRDs), please refer to the existing Kubernetes guidelines:
* [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md).
* [API changes](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md).

In particular, this project follows the API stability guidelines:
* For alpha API versions (e.g. `v1alpha1`, `v1alpha2`, ...), we may allow to break forward and backward compatibility (but we'll try hard to avoid it).
* For beta API versions (e.g. `v1beta1`, `v1beta2`, ...), we may allow to break backward compatibility but not forward compatibility.
* For stable API versions (e.g. `v1`), we don't allow to break backward and forward compatibility.

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
scripts: add the test-cluster command

This uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38
```

The format can be described more formally as follows:

```
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

# Proposal Process

The Prometheus Operator project accepts proposals for new features, enhancements and design documents.
Proposals can be submitted in the form of a pull request using the template below.

The process is adopted from the Thanos community.

## Your Proposal Title

* **Owners:**
  * `<@author: single champion for the moment of writing>`

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

## Goals

Goals and use cases for the solution as proposed in [How](#how):

* Allow easy collaboration and decision making on design ideas.
* Have a consistent design style that is readable and understandable.
* Have a design style that is concise and covers all the essential information.

### Audience

If this is not clear already, provide the target audience for this change.

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
