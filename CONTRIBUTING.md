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

Thanks for your contributions!

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that you can easily kill and
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
git tools.
