# RBAC

RBAC for the Prometheus Operator involves two parts, RBAC rules for the
Operator itself and RBAC rules for the Prometheus Pods themselves created by
the Operator as Prometheus requires access to the Kubernetes API for target and
Alertmanager discovery.

## Prometheus Operator RBAC

In order for the Prometheus Operator to work in an RBAC based authorization
environment, a `ClusterRole` with access to all the resources the Operator
requires for the Kubernetes API needs to be created. This section is intended
to describe, why the specified rules are required.

Here is a ready to use yaml definition of a `ClusterRole` that can be used to
start the Prometheus Operator:

```yaml
apiVersion: rbac.authorization.k8s.io/v1alpha1
kind: ClusterRole
metadata:
  name: prometheus-operator
rules:
- apiGroups:
  - extensions
  resources:
  - thirdpartyresources
  verbs:
  - create
- apiGroups:
  - monitoring.coreos.com
  resources:
  - alertmanagers
  - prometheuses
  - servicemonitors
  verbs:
  - "*"
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs: ["*"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["*"]
- apiGroups: [""]
  resources:
  - pods
  verbs: ["list", "delete"]
- apiGroups: [""]
  resources:
  - services
  - endpoints
  verbs: ["create", "update"]
- apiGroups: [""]
  resources:
  - nodes
  verbs: ["list", "watch"]
```

> Note: A cluster admin is required to create this `ClusterRole` and create a
> `ClusterRoleBinding` to the `ServiceAccount` used by the Prometheus Operator
> `Pod`.  The `ServiceAccount` used by the Prometheus Operator `Pod` can be
> specified in the `Deployment` object used to deploy it.


When the Prometheus Operator boots up for the first time it registers the
`thirdpartyresources` it uses, therefore the `create` action on those is
required.

As the Prometheus Operator work extensively with the `thirdpartyresources` it
registers, it requires all actions on those objects. Those are:

* `alertmanagers`
* `prometheuses`
* `servicemonitors`

Alertmanager and Prometheus clusters are created using `statefulsets` therefore
all changes to an Alertmanager or Prometheus object result in a change to the
`statefulsets`, which means all actions must be permitted.

Additionally as the Prometheus Operator takes care of generating configurations
for Prometheus to run, it requires all actions on `configmaps`.

When the Prometheus Operator performs version migrations from one version of
Prometheus or Alertmanager to the other it needs to `list` `pods` running an
old version and `delete` those.

The Prometheus Operator creates and updates `services` called
`prometheus-operated` and `alertmanager-operated`, which are used as governing
`Service`s for the `StatefulSet`s.

As the kubelet is currently not self-hosted, the Prometheus Operator has a
feature to synchronize the IPs of the kubelets into an `Endpoints` object,
which requires access to `list` and `watch` of `nodes` (kubelets) and `create`
and `update` for `endpoints`.

## Prometheus RBAC

The Prometheus server itself accesses the Kubernetes API to discover targets
and Alertmanagers. Therefore a separate `ClusterRole` for those Prometheus
servers need to exist.

As Prometheus does not modify any Objects in the Kubernetes API, but just reads
them it simply requires the `get`, `list`, and `watch` actions.

```yaml
apiVersion: rbac.authorization.k8s.io/v1alpha1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
```

> Note: A cluster admin is required to create this `ClusterRole` and create a
> `ClusterRoleBinding` to the `ServiceAccount` used by the Prometheus `Pod`s.
> The `ServiceAccount` used by the Prometheus `Pod`s can be specified in the
> `Prometheus` object.

