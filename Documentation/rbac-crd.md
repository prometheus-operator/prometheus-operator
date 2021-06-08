---
title: "RBAC for CRDs"
description: "Aggregate permissions on the Prometheus Operator CustomResourceDefinitions."
lead: ""
date: 2021-03-08T08:49:31+00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 420 # nice
toc: true
---

## Aggregated ClusterRoles

It can be useful to aggregate permissions on the Prometheus Operator CustomResourceDefinitions to the default user-facing roles, like `view`, `edit` and `admin`.

This can be achieved using ClusterRole aggregation. This lets admins include rules for custom resources, such as those served by CustomResourceDefinitions or Aggregated API servers, on the default roles.

> Note: ClusterRole aggregation is available starting Kubernetes 1.9. 

## Example

In order to aggregate _read_ (resp. _edit_) permissions for the Prometheus Operator CustomResourceDefinitions to the `view` (resp. `edit` / `admin`) role(s), a cluster admin can create the `ClusterRole`s below.

This grants:
- Users with `view` role permissions to view the Prometheus Operator CRDs within their namespaces,
- Users with `edit` and `admin` roles permissions to create, edit and delete Prometheus Operator CRDs within their namespaces.

[embedmd]:# (../example/rbac/prometheus-operator-crd/prometheus-operator-crd-cluster-roles.yaml)
```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: prometheus-crd-view
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-view: "true"
rules:
- apiGroups: ["monitoring.coreos.com"]
  resources: ["alertmanagers", "alertmanagerconfigs", "prometheuses", "prometheusrules", "servicemonitors", "podmonitors", "probes"]
  verbs: ["get", "list", "watch"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: prometheus-crd-edit
  labels:
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
- apiGroups: ["monitoring.coreos.com"]
  resources: ["alertmanagers", "alertmanagerconfigs", "prometheuses", "prometheusrules", "servicemonitors", "podmonitors", "probes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```
