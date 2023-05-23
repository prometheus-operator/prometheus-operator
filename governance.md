# Governance

This document describes the rules and governance of the project. It is a slightly modified version of the [Prometheus Governance](https://prometheus.io/governance/#governance-changes).

It is meant to be followed by all the developers of the Prometheus Operator project and the Prometheus Operator community. Common terminology used in this governance document are listed below:

* **Maintainers Team**: Contributors who have the `Owner` role on the [prometheus-operator organization](https://github.com/prometheus-operator) and all projects within it. Members are listed in the [MAINTAINERS.md](MAINTAINERS.md) file.

* **Triage Team**: Contributors who have the `Member` role and `Triage` GitHub permission on the [prometheus-operator organization](https://github.com/prometheus-operator) and all projects within it, allowing them to modify GitHub issues and PRs statuses and labels. Triage members are listed in the [MAINTAINERS.md](MAINTAINERS.md#triage) file.

* **The Prometheus Operator project**: The sum of all activities performed under the [prometheus-operator organization](https://github.com/prometheus-operator), concerning one or more repositories or the community.

Both Triage and Maintainers are part of the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).

## Values

The Prometheus Operator developers and community are expected to follow the values defined in the [Code of Conduct](code-of-conduct.md).

Furthermore, the Prometheus Operator community strives for kindness, giving feedback effectively, and building a welcoming environment. The Prometheus Operator developers generally decide by consensus and only resort to conflict resolution by a majority vote if consensus cannot be reached.

## Decision making

### Maintainers Team

Team member status may be given to those who have made ongoing contributions to the Prometheus Operator project for at least 3 months.
This is usually in the form of code improvements, pull-request reviews, issue triaging or notable work on documentation, but organizing events or user support could also be taken into account.

New members may be proposed by any existing Maintainer by email to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). It is highly desirable to reach consensus about acceptance of a new member.
However, the proposal is ultimately voted on by a formal [supermajority vote](#supermajority-vote) of Team Maintainers.

If the new member proposal is accepted, the proposed team member should be contacted privately via email to confirm or deny their acceptance of team membership.
This email will also be CC'd to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team) for record-keeping purposes.

If they choose to accept, the following steps are taken:

* The new maintainer is added to the [prometheus-operator organization](https://github.com/prometheus-operator) with the `Owner` role.
* The new maintainer is added to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).
* The new maintainer is added to the [MAINTAINERS.md](MAINTAINERS.md) file.
* The new maintainer is announced on the [Prometheus Operator Twitter](https://twitter.com/PromOperator) by an existing team member.

Team members may retire at any time by emailing the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).

Team members can be removed by [supermajority vote](#supermajority-vote) on the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). For this vote, the member in question is not eligible to vote and does not count towards the quorum.

Upon death of a member, their team membership ends automatically.

### Triage Team

Triage team has similar rules, however the contributions made to the projects does not need to be as significant as expected from a maintainer.

New members may be proposed by any existing Maintainer or Triage person by email to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). It is highly desirable to reach consensus about acceptance of a new member.
However, the proposal is ultimately voted on by a formal [majority vote](#majority-vote) (in comparison to Maintainer's vote which requires a supermajority vote).

If the new member proposal is accepted, the proposed team member should be contacted privately via email to confirm or deny their acceptance of team membership.
This email will also be CC'd to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team) for record-keeping purposes.

If they choose to accept, the following steps are taken:

* The new triage team member is added to the [Prometheus Operator project](http://github.com/prometheus-operator/prometheus-operator) with `Member` role and `Triage` permission.
* The new triage team member is added to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).
* The new triage team member is added to the [MAINTAINERS.md](MAINTAINERS.md) file under the Triage section.
* The new triage team member is announced on the [Prometheus Operator Twitter](https://twitter.com/PromOperator) by an existing team member.

Triage member may retire at any time by emailing the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).

Triage member can be removed by [majority vote](#majority-vote) on the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). Only the [Maintainers team members](#maintainers-team) have right to vote.

Upon death of a member, their Triage team membership ends automatically.

### Technical decisions

Smaller technical decisions are made informally and [lazy consensus](#consensus) is assumed. Technical decisions that span multiple parts of the Prometheus Operator project
should be discussed and made on the [GitHub issues](https://github.com/prometheus-operator/prometheus-operator/issues) and in most cases followed by proposal as described [here](https://github.com/prometheus-operator/prometheus-operator/blob/main/CONTRIBUTING.md).

Decisions are usually made by [lazy consensus](#consensus). If no consensus can be reached, the matter may be resolved by [majority vote](#majority-vote).

### Governance changes

Material changes to this document are discussed publicly on the [Prometheus Operator GitHub](http://github.com/prometheus-operator/prometheus-operator).
Any change requires a [supermajority](#supermajority-vote) in favor. Editorial changes may be made by [lazy consensus](#consensus) unless challenged.

### Other matters

Any matter that needs a decision, including but not limited to financial matters, may be called to a vote by any Maintainer if they deem it necessary.
For financial, private, or personnel matters, discussion and voting takes place on the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). Otherwise discussion and votes are held in public on the GitHub issues or #prometheus-operator-dev Kubernetes slack channel.

## Voting

The Prometheus Operator project usually runs by informal consensus, however sometimes a formal decision must be made.

Depending on the subject matter, as laid out [above](#decision-making), different methods of voting are used.

For all votes, voting must be open for at least one week. The end date should be clearly stated in the call to vote.
A vote may be called and closed early if enough votes have come in one way so that further votes cannot change the final decision.

In all cases, all and only [Maintainers team members](#maintainers-team) are eligible to vote, with the sole exception of the forced removal of a team member, in which said member is not eligible to vote.

Discussion and votes on personnel matters (including but not limited to team membership and maintainership) are held in private on the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). All other discussion and votes are held in public on the GitHub issues or #prometheus-operator-dev CNCF slack channel.

For public discussions, anyone interested is encouraged to participate. Formal power to object or vote is limited to [Maintainers team members](#maintainers-team).

### Governance

It's important for the project to stay independent and focused on shared interest instead of single use case of one company or organization.

We value open source values and freedom, that's why we limit Maintainers Team **votes to maximum two from a single organization or company.**

We also encourage any other company interested in helping maintaining Prometheus Operator to join us to make sure we stay independent.

### Consensus

The default decision making mechanism for the Prometheus Operator project is [lazy consensus](https://couchdb.apache.org/bylaws.html#lazy). This means that any decision on technical issues is considered supported by the [Maintainers team members](#maintainers-team) as long as nobody objects.

Silence on any consensus decision is implicit agreement and equivalent to explicit agreement. Explicit agreement may be stated at will.

Consensus decisions can never override or go against the spirit of an earlier explicit vote.

If any of the [Maintainers team members](#maintainers-team) raises objections, the team members work together towards a solution that all involved can accept.
This solution is again subject to lazy consensus.

In case no consensus can be found, but a decision one way or the other must be made, [Maintainers team members](#maintainers-team) may call a formal [majority vote](#majority-vote).

### Majority vote

Majority votes must be called explicitly in a separate thread on the appropriate mailing list. The subject must be prefixed with `[VOTE]`.
In the body, the call to vote must state the proposal being voted on. It should reference any discussion leading up to this point.

Votes may take the form of a single proposal, with the option to vote yes or no, or the form of multiple alternatives.

A vote on a single proposal is considered successful if more vote in favor than against.

If there are multiple alternatives, members may vote for one or more alternatives, or vote “no” to object to all alternatives.
It is not possible to cast an “abstain” vote. A vote on multiple alternatives is considered decided in favor of one alternative if it has received the most votes in favor, and a vote from more than half of those voting. Should no alternative reach this quorum, another vote on a reduced number of options may be called separately.

### Supermajority vote

Supermajority votes must be called explicitly in a separate thread on the appropriate mailing list.
The subject must be prefixed with `[VOTE]`. In the body, the call to vote must state the proposal being voted on. It should reference any discussion leading up to this point.

Votes may take the form of a single proposal, with the option to vote yes or no, or the form of multiple alternatives.

A vote on a single proposal is considered successful if at least two thirds of those eligible to vote vote in favor.

If there are multiple alternatives, members may vote for one or more alternatives, or vote “no” to object to all alternatives.
A vote on multiple alternatives is considered decided in favor of one alternative if it has received the most votes in favor, and a vote from at least two thirds of those eligible to vote. Should no alternative reach this quorum, another vote on a reduced number of options may be called separately.

## Frequently Asked Questions

This section is informational. In case of disagreement, the rules above overrule any FAQ.

### For a majority vote, what if there is an even number of maintainers and an equal amount of votes in favor than against?

It has to be the majority so the vote will be declined.

### So what's the TLDR difference between majority vs supermajority?

It's about number of up votes to agree on the decision.

* majority: the majority of voters have to agree.
* supermajority: 2/3 of the voters have to agree.

### How do I propose a decision?

See the [Contributing guide](https://github.com/prometheus-operator/prometheus-operator/blob/main/CONTRIBUTING.md).

### How do I become a team member?

To become an official member of the [Maintainers Team](#maintainers-team), you should make ongoing contributions to one or more project(s) for at least three months.
At that point, a team member (typically a maintainer of the project) may propose you for membership.
The discussion about this will be held in private, and you will be informed privately when a decision has been made. A possible, but not required, graduation path is to become a triage member first.

Should the decision be in favor, your new membership will also be announced on the [Prometheus Operator Twitter](https://twitter.com/PromOperator)

### How do I add a project?

As a team member, propose the new project on the [Prometheus Operator GitHub Issue](https://github.com/prometheus-operator/prometheus-operator/issues). However, currently to maintain a project in our organization, you have to become a [Maintainers Team](#maintainers-team) member.

### How do I remove a Maintainer or Triage member?

All members may resign by notifying the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team). If you think a team member should be removed against their will, propose this to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).
Discussions will be held there in private.

### Can majority/supermajority vote be done on GitHub PR by just approving PR?

No, an email with the `[VOTE]` prefix has to be sent to the [prometheus-operator mailing list](https://groups.google.com/forum/#!forum/prometheus-operator-team).

### What if during majority/supermajority vote there is no answer after 1 week?

For majority voting, all members that didn't reply explicitly are assumed to agree with the proposal.

For supermajority voting, the team has to wait for the replies from all members.
