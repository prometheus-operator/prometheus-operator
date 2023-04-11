# Silence CRD

* Owners:
  * [mcbenjemaa](https://github.com/mcbenjemaa)
* Related Tickets:
  * [#5452](https://github.com/prometheus-operator/prometheus-operator/issues/5452)
  * [#2398](https://github.com/prometheus-operator/prometheus-operator/issues/2398)
* Other docs:
  * n/a

This document describes the creation of `Silence` Custom Resource Definition that defines Alertmanager silences
configurations in the Kubernetes way.

# Why

prometheus-operator doesn't have a way to automate the management of Alertmanager silences. Many users have either been using some internal scripts
or an additional operator that does the job e.g. [silence-operator](https://github.com/giantswarm/silence-operator).

* Users who are using CI/CD jobs to manage silences have reported that this is really cumbersome.
* Users that use scripts are not in better shape obviously.
* Users that use a standalone operator that exposes a Silence CRD are in a better situation, some folks have reported that they are using GitOps
to fully manage the life cycle of Silences Custom resources, Nevertheless, this comes with a drawback 
as the team must be able to manage a new component in their stack.

Additionally, Having a new component in the stack and keeping it maintained is not always ideal (said the folks at [Giant Swarm](https://giantswarm.io) the owners of [silence-operator](https://github.com/giantswarm/silence-operator)).
It's really better to have this feature as part of prometheus-operator and keep it maintained and available for the community.

By Adding support for `Silence` CRD in the prometheus-operator, this will make users more flexible in terms of choosing the tool 
to deploy their silences and will free others from managing a standalone component within the stack.
