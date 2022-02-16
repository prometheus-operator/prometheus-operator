---
id: 0001
title: How to write a Proposal
status: published
authors: Michael Deffenbaugh <mike@zeroent.net>
---

## Summary

_(**These notes in parenthesis describe the purpose of each section.** The
summary contains a short summary of the Proposal. Ideally it has to be useful for
people who do not know what problem the Proposal is attempting to solve. The summary
gives readers context and helps them evaluate if this is a topic that interests
or affects them.)_

This is the first Proposal and it covers:

* What Proposals are and why we need them
* How to write an Proposal
* When to avoid a Proposal
* Lifecycle of an Proposal

Rename the [proposal](https://github.com/tinkerbell/proposals) repository to Proposal.

## Goals and non-Goals

_(Goals and non-Goals contains two bullet point sections that briefly describe
the end goals of what is being discussed and what is out-of-scope or otherwise
not open for discussion.)_

Goals

* Explain what a Proposal is, how and when to use it
* Share the Proposal lifecycle
* Be the first Proposal for the project

Non-Goals

* Offering all of the automation, scripts and bots described as part of the Proposal
  lifecycle
* Converting the already open PR against the
    [proposal](https://github.com/tinkerbell/proposals) repo to this new format

## Content

_(This is the main section. It contains the core of the discussion, the design
you are proposing, the changes this imposes and the new features it provides.
This should include a high-level overview of how this proposal can be
implemented.)_

### Why a Proposal is needed and what it looks like

Discussing and deciding how and why a feature or a piece of code should or
should not be written is an important part of the engineering process.

As an open source project, Tinkerbell needs to have a public place where
community members, contributors can share their ideas in detail, asking for
help, feedback and more in general to spot problems addressable as early as
possible.

A maintainer can ask contributors to write an Proposal in order to open a discussion
about a feature that looks trivial, helping contributors to sit down and think
about the design of the request or the code they will build. It is a way to
translate Slack discussions to something that won’t get lost for example.

Currently, we have a repo called
[proposal](https://github.com/tinkerbell/proposals) that I would like to rename
to Proposal. Every Proposal will have its own directory with a markdown file inside.

### Metadata

The Proposal is written in markdown and includes metadata in the header of the file:

* `ID`: identifies the Proposal
* `Title`: title of the Proposal
* `Status`: description of the current state for the Proposal
* `Authors`: in the format of a list of `Name <email@emai.com>`

Here is a complete example:

```yaml
---
id: 0001
title: Request for discussion HOWTO
pr: 2
status: ideation
authors: Gianluca Arbezzano <gianarb92@gmail.com>
---
```

### Lifecycle

I don’t want to rewrite the lifecycle here because the one in use by [Oxide
Computer](https://oxide.computer/blog/rfd-1-requests-for-discussion/#rfd-life-cycle)
looks familiar and is functional for our purpose. They wrote the lifecycle in
detail and, for now, we will follow it as it is.

What follows here is copy/pasted from ["RFD 1 Requests for
Discussion"](https://oxide.computer/blog/rfd-1-requests-for-discussion/) and all
the credit goes to [Oxide Computer](https://oxide.computer).

~~~markdown

### RFD Metadata and State

At the start of every RFD document, we'd like to include a brief amount of metadata. The metadata format is based on the python-markdown2 metadata format. It'd look like:

---
authors: Jenny Smith <jenny@example.computer>, Neal Jones <neal@example.computer>
state: prediscussion
---
We keep track of three pieces of metadata:

authors: The authors (and therefore owners) of an RFD. They should be listed with their name and e-mail address.
state: Must be one of the states discussed below.
discussion: For RFDs that are in or beyond the discussion state, this should be a link to the PR to integrate the RFD; see below for details.
An RFD can be in one of the following six states:

* prediscussion
* ideation
* discussion
* published
* committed
* abandoned

A document in the `prediscussion` state indicates that the work is not yet ready
for discussion, but that the RFD is effectively a placeholder. The `prediscussion`
state signifies that work iterations are being done quickly on the RFD in
its branch in order to advance the RFD to the `discussion` state.

A document in the `ideation` state contains only a description of the topic that
the RFD will cover, providing an indication of the scope of the eventual RFD.
Unlike the `prediscussion` state, there is no expectation that is undergoing
active revision. Such a document is scratchpad for related ideas. Any member of
the team is encouraging to start active development of such an RFD (moving it to
the `prediscussion` state) with or without the participation of the original
author. It is critical that RFDs in the `ideation` state are clear and narrowly
defined.

Documents under active discussion should be in the `discussion` state. At this
point a discussion is being had for the RFD in a Pull Request.

Once (or if) discussion has converged and the Pull Request is ready to be
merged, it should be updated to the published state before merge. Note that just
because something is in the `published` state does not mean that it cannot be
updated and corrected. See the Making changes to an RFD section for more
information.

The `prediscussion` state should be viewed as essentially a collaborative
extension of an engineer's notebook, and the `discussion` state should be used
when an idea is being actively discussed. These states shouldn't be used for
ideas that have been committed to, organizationally or otherwise; by the time an
idea represents the consensus or direction, it should be in the `published` state.

Once an idea has been entirely implemented, it should be in the `committed` state.
Comments on ideas in the `committed` state should generally be raised as issues --
but if the comment represents a call for a significant divergence from or
extension to committed functionality, a new RFD may be called for; as in all
things, use your best judgment.

Finally, if an idea is found to be non-viable (that is, deliberately never
implemented) or if an RFD should be otherwise indicated that it should be
ignored, it can be moved into the `abandoned` state.

We will go over this in more detail. Let's walk through the life of a RFD.

### RFD life-cycle

There is a prototype script in this repository, scripts/new.sh, that will
automate the process.

```sh
$ scripts/new.sh 0042 "My title here"
```

If you wish to create a new RFD by hand, or understand the process in greater detail, read on.

NOTE: Never at anytime through the process do you push directly to the main
branch. Once your pull request (PR) with your RFD in your branch is merged into
main, then the RFD will appear in the master branch.

### RESERVE A RFD NUMBER

You will first need to reserve the number you wish to use for your RFC. This
number should be the next available RFD number from looking at the current git
branch -r output.

### CREATE A BRANCH FOR YOUR RFD

Now you will need to create a new git branch, named after the RFD number you
wish to reserve. This number should have leading zeros if less than 4 digits.
Before creating the branch, verify that it does not already exist:

```sh
$ git branch -rl *0042
```

If you see a branch there (but not a corresponding sub-directory in rfd in
main), it is possible that the RFD is currently being created; stop and check
with co-workers before proceeding! Once you have verified that the branch
doesn't exist, create it locally and switch to it:

```sh
$ git checkout -b 0042
```

### CREATE A PLACEHOLDER RFD

Now create a placeholder RFD. You can do so with the following commands:

```sh
$ mkdir -p rfd/0042
$ cp prototypes/prototype.md rfd/0042/README.md
```


Fill in the RFD number and title placeholders in the new doc and add your name
as an author. The status of the RFD at this point should be prediscussion.

If your preference is to use asciidoc, that is acceptable as well, however the
examples in this flow will assume markdown.

### PUSH YOUR RFD BRANCH REMOTELY

Push your changes to your RFD branch in the RFD repo.

```sh
$ git add rfd/0042/README.md
$ git commit -m '0042: Adding placeholder for RFD <Title>'
$ git push origin 0042
```

After your branch is pushed, the table in the README on the main branch will
update automatically with the new RFD. If you ever change the name of the RFD in
the future, the table will update as well. Whenever information about the state
of the RFD changes, this updates the table as well. The single source of truth
for information about the RFD comes from the RFD in the branch until it is
merged.

### ITERATE ON YOUR RFD IN YOUR BRANCH

Now, you can work on writing your RFD in your branch.

```sh
$ git checkout 0042
```

Now you can gather your thoughts and get your RFD to a state where you would
like to get feedback and discuss with others. It's recommended to push your
branch remotely to make sure the changes you make stay in sync with the remote
in case your local gets damaged.

It is up to you as to whether you would like to squash all your commits down to
one before opening up for feedback, or if you would like to keep the commit
history for the sake of history.

### DISCUSS YOUR RFD

When you are ready to get feedback on your RFD, make sure all your local changes
are pushed to the remote branch. At this point you are likely at the stage where
you will want to change the status of the RFD from prediscussion to discussion
for a fully formed RFD or to ideation for one where only the topic is specified.
Do this in your branch.

### PUSH YOUR RFD BRANCH REMOTELY

Along with your RFD content, update the RFD's state to discussion in your
branch, then:

```sh
$ git commit -am '0042: Add RFD for <Title>'
$ git push origin 0042
```

### OPEN A PULL REQUEST

Open a pull request on GitHub to merge your branch, in this case 0042 into the
main branch.

If you move your RFD into discussion but fail to open a pull request, a friendly
bot will do it for you. If you open a pull request but fail to update the state
of the RFD to discussion, the bot will automatically correct the state by moving
it into discussion. The bot will also cleanup the title of the pull request to
be RFD {num} {title}. The bot will automatically add the link to the pull
request to the discussion: metadata.

After the pull request is opened, anyone subscribed to the repo will get a
notification that you have opened a pull request and can read your RFD and give
any feedback.

### DISCUSS THE RFD ON THE PULL REQUEST

The comments you choose to accept from the discussion are up to you as the owner
of the RFD, but you should remain empathetic in the way you engage in the
discussion.

For those giving feedback on the pull request, be sure that all feedback is
constructive. Put yourself in the other person's shoes and if the comment you
are about to make is not something you would want someone commenting on an RFD
of yours, then do not make the comment.

### MERGE THE PULL REQUEST

After there has been time for folks to leave comments, the RFD can be merged
into main and changed from the discussion state to the published state. The
timing is left to your discretion: you decide when to open the pull request, and
you decide when to merge it. As a guideline, 3-5 business days to comment on
your RFD before merging seems reasonable -- but circumstances (e.g., time zones,
availability of particular expertise, length of RFD) may dictate a different
timeline, and you should use your best judgment. In general, RFDs shouldn't be
merged if no one else has read or commented on it; if no one is reading your
RFD, it's time to explicitly ask someone to give it a read!

Discussion can continue on published RFDs! The discussion: link in the metadata
should be retained, allowing discussion to continue on the original pull
request. If an issue merits more attention or a larger discussion of its own, an
issue may be opened, with the synopsis directing the discussion.

Any discussion on an RFD in the can still be made on the original pull request
to keep the sprawl to a minimum. Or if you feel your comment post-merge requires
a larger discussion, an issue may be opened on it -- but be sure to reflect the
focus of the discussion in the issue synopsis (e.g., "RFD 42: add consideration
of RISC-V"), and be sure to link back to the original PR in the issue
description so that one may find one from the other.

### MAKING CHANGES TO AN RFD

After your RFD has been merged, there is always opportunity to make changes. The
easiest way to make a change to an RFD is to make a pull request with the change
you would like to make. If you are not the original author of the RFD name your
branch after the RFD # (e.g. 0001) and be sure to @ the original authors on your
pull request to make sure they see and approve of the changes.

Changes to an RFD will go through the same discussion and merge process as
described above.

### COMMITTING TO AN RFD

Once an RFD has become implemented -- that is, once it is not an idea of some
future state but rather an explanation of how a system works -- its state should
be moved to be committed. This state is essentially no different from published,
but represents ideas that have been more fully developed. While discussion on
committed RFDs is permitted (and changes allowed), they would be expected to be
infrequent.
~~~

_end of copy/paste from ["Proposal 1 Requests for
Discussion"](https://oxide.computer/blog/rfd-1-requests-for-discussion/)._

### Table of contents

This is what the table of contents for a Proposal should look like:

1. Summary
2. Goals and non-goals
3. Content
4. System-context-diagram
5. APIs
6. Alternatives

### Tooling

I would like to get a generated version of the Proposal in markdown with a table in
the homepage listing all of the Proposals with their status. This will allow us to
quickly identify those open for discussion.

### Credits, links and inspiration

* Previously designed Proposal workflows that are not publicly available
* [Kubernetes
    KEPS](https://github.com/kubernetes/enhancements/tree/master/keps)
* [https://www.industrialempathy.com/posts/design-docs-at-google/](https://www.industrialempathy.com/posts/design-docs-at-google/)
* [https://oxide.computer/blog/rfd-1-requests-for-discussion/#rfd-metadata-and-state](https://oxide.computer/blog/rfd-1-requests-for-discussion/#rfd-metadata-and-state)
* [https://github.com/crossplane/crossplane/blob/master/design/README.md](https://github.com/crossplane/crossplane/blob/master/design/README.md)
* [https://github.com/joyent/rfd](https://github.com/joyent/rfd)

### When you can avoid a Proposal

Proposals are not necessary when small fixes are made or minor features are
introduced.  An Proposal is not necessary in when the community is satisfied with the
discussions happening in other channels like Slack, voice, or in pull requests
and issues.

## System-context-diagram

_(How does this feature or discussion fit into the big picture? Tinkerbell, like
many other modern software stacks, have a high level of interaction between
component, including external components. A system context diagram helps the
reader visually understand how this Proposal plays a role.)_

## APIs

_(Describe any API changes, including new APIs, here with JSON, YAML, `curl`, or
GRPC Interface examples.)_

## Alternatives

_(List or describe any alternatives considered that didn’t fit this Proposal. This
could be links to similar services or detailed explorations that were abandoned
in favor of the proposed Proposal.)_
