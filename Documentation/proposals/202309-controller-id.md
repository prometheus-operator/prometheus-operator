# Controller-ID Proposal

* Owners:
  * [danielmellado](https://github.com/danielmellado)
* Status:
  * `Implemented`
* Related Tickets:
  * [#4281](https://github.com/prometheus-operator/prometheus-operator/issues/4281)
  * [#4498](https://github.com/prometheus-operator/prometheus-operator/pull/4498)
* Other Docs:
  * n/a

# Introduction

This proposal aims to implement a solution to support multiple cluster-level
Prometheus instances running concurrently without conflicting over the same
custom resources. This solution isn't limited to the Prometheus resources as
it'll also be available for `AlertManager`and `ThanosRuler` ones, as well as for
any pod-based resource that could be added in the future.

This issue can significantly impact use cases where multiple Prometheus operator
instances run at the same time in the Kubernetes cluster.

# Why

Currently, we encounter issues when different users deploy different instances
of the Prometheus operator, that will try to reconcile the same resources.

In the worst-case scenario, these operators may not only compete for ownership
of the CRD resources but also attempt to rewrite or redeploy different versions
of the CRD, causing disruptions to all pods.

The remediation for this scenario, where users deploy their Prometheus operator
instances in parallel, involves using one of the many CLI arguments such as
`--deny-namespaces`, `--namespaces`, `--prometheus-instance-selector`  or
`prometheus-instance-namespaces`. But this requires cooperation between the
different parties and there's no way to ensure that a specific monitoring
resource is managed only by a specific operator instance.

# How

After some research from @machine424, we have identified a potential solution
already implemented by the
[zalando/postgres-operator](https://github.com/zalando/postgres-operator). When
an operator is configured with a specific "controller ID" value, it will only
reconcile resources that have a matching "controller ID" annotation.

Conversely, if the operator is not configured with a "controller ID," it will
skip all resources that have a "controller ID" annotation. More details can be
found in the
[zalando/postgres-operator documentation](https://github.com/zalando/postgres-operator/blob/master/docs/administrator.md#operators-with-defined-ownership-of-certain-postgres-clusters).

# Goals

* Guarantee that a custom resource will be managed by a specific Prometheus
  operator instance.

## Audience

This proposal is relevant to the following audience:

* Users who provide Prometheus as a service and want to run multiple Prometheus
  operator instances in different namespaces.
* Users seeking to mitigate the impact of rogue Prometheus instances.

# Non-Goals

* Provide a solution that works with user intervantion. It'll require work from
  the user deploying the operator and resources. (e.g. if the operator is
  started without any specific argument, it'll attempt to reconcile all
  resources in all namespaces).

# Alternatives

Although the initial discussion for this proposal considered adding an owner
reference within the scope, the
[ControllerRef](https://github.com/kubernetes/design-proposals-archive/blob/acc25e14ca83dfda4f66d8cb1f1b491f26e78ffe/api-machinery/controller-ref.md)
model does not directly address this problem because the `ControllerRef`model
solves the problem of controllers that fight over controlled objects due to
overlapping selectors (e.g. a ReplicaSet fighting with a ReplicationController
over Pods because both controllers have label selectors that match those Pods)
