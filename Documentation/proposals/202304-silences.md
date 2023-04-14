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

## Pitfalls of the current solution

Using Alertmanager API Directly comes with drawbacks:

* Teams have to build an automation to add silences in a centralized manner
* There is no input validation, which can lead to an invalid silence configuration
* Teams need to manage expiration on their own
* If a silence was deleted by accident, it will be permanent.

# Goals

* Provide a way for users to manage Alertmanager silences with the Kubernetes Way.
* Make it easier to manage Alertmanager silences via centralised repo with e.g. ArgoCD, Flux...
* Provide a way to manage silence expiration with a silence CR.

## Audience

* Users who serve Prometheus as a service and want to have an interface in defining silences exposed to developers.
* Users who want to manage silences the same way as for services running within the Kubernetes cluster
* Users who want a supported Kubernetes way of silences outside the Kubernetes cluster

# Non-Goals

* This proposal doesn't aim to remove the CR after the expiration date has reached
* Refactoring of the other CRDs is not in scope for the first version


# How

Creating a new cluster-scope Silence CRD that will act as an interface by adding silences via the Alertmanager API.
Usage of Silences doesn't exclude the use of the other CRDs, they are not mutually exclusive.
`Silence` will allow the creation of any silences to Alertmanager, while the other CRDs provide sane defaults. This will allow
for isolated testing of the new `Silence` CRD.


A typical Silence resouce should looks like the following:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Silence
metadata:
  name: my-silence
  labels:
    test: value
spec:
  selector:
    app: test-alertmanager
  validUntil: "2023-06-01"
  matchers:
    - name: namespace
      value: test-ns
      isRegex: false
```

The above resource will result in creating a silence in the Alertmanager with label selector, 
`app: test-alertmanager`.

* `selector` used to select which Kubernetes Service to use as an Alertmanager API.
* `validUntil` to define the expiration of the silence.
* `matchers` field corresponds to the Alertmanager silence matchers each of which consists of:
    - `name` - name of tag on an alert to match
    - `value` - fixed string or expression to match against the value of the tag named by name above on an alert
    - `isRegex` - a boolean specifying whether to treat value as a regex (=~) or a fixed string (=)
    - `isEqual` - a boolean specifying whether to use equal signs (= or =~) or to negate the matcher (!= or !~)


This example doesn't list all the fields that are offered by prometheus. The implementation of all the fields will be
done in an iterative process and as such, the expectation is not for all of them to be implemented in the first version.

Also, to help selecting `Silence`, a new field will be added to the Alertmanager CRD:

```yaml
[...]
spec:
  silenceSelector: ...
  silenceNamespaceSelector: ...
```

# Alternatives

* Use Alertmanager API, with the pitfalls described earlier

# Action Plan

1. Create the `Silence` CRD, covering `selector`, `matchers` and `validUntil`. 
2. Once released, add other mechanisms to the CRD and complete the implementation.
