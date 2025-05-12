---
weight: 204
toc: true
title: Prometheus Agent
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Guide for running Prometheus in Agent mode
---

{{< alert icon="👉" text="Prometheus Operator >= v0.64.0 is required."/>}}

As mentioned in [Prometheus's blog](https://prometheus.io/blog/2021/11/16/agent/), Prometheus Agent
is a deployment model optimized for environments where all collected data is forwarded to
a long-term storage solution, e.g. Cortex, Thanos or Prometheus, that do not need storage or rule evaluation.

First of all, make sure that the PrometheusAgent CRD is installed in the cluster and that the operator has the proper RBAC permissions to reconcile the PrometheusAgent resources.

```yaml mdox-exec="cat example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: 0.82.2
  name: prometheus-operator
rules:
- apiGroups:
  - monitoring.coreos.com
  resources:
  - alertmanagers
  - alertmanagers/finalizers
  - alertmanagers/status
  - alertmanagerconfigs
  - prometheuses
  - prometheuses/finalizers
  - prometheuses/status
  - prometheusagents
  - prometheusagents/finalizers
  - prometheusagents/status
  - thanosrulers
  - thanosrulers/finalizers
  - thanosrulers/status
  - scrapeconfigs
  - servicemonitors
  - podmonitors
  - probes
  - prometheusrules
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - delete
- apiGroups:
  - ""
  resources:
  - services
  - services/finalizers
  verbs:
  - get
  - create
  - update
  - delete
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - patch
  - create
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - get
  - create
  - update
  - delete
```

Similarly to Prometheus, Prometheus Agent will also require permission to scrape targets. Because of this, we will create a new service account for the Agent with the necessary permissions to scrape targets.

Start with the ServiceAccount, ClusterRole and ClusterRoleBinding:

```yaml mdox-exec="cat example/rbac/prometheus-agent/prometheus-service-account.yaml"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus-agent
```

```yaml mdox-exec="cat example/rbac/prometheus-agent/prometheus-cluster-role.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-agent
rules:
- apiGroups: [""]
  resources:
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

```yaml mdox-exec="cat example/rbac/prometheus-agent/prometheus-cluster-role-binding.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-agent
subjects:
- kind: ServiceAccount
  name: prometheus-agent
  namespace: default
```

Lastly, we can deploy the Agent. The `spec` field is very similar to the Prometheus CRD but the features that aren't applicable to the agent mode (like alerting, retention, Thanos, ...) are not available.

```yaml mdox-exec="cat example/rbac/prometheus-agent/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1alpha1
kind: PrometheusAgent
metadata:
  name: prometheus-agent
spec:
  replicas: 2
  serviceAccountName: prometheus-agent
  serviceMonitorSelector:
    matchLabels:
      team: frontend
```

Continue with the [Getting Started page]({{<ref "docs/developer/getting-started.md">}}) to learn how to monitor applications running on Kubernetes.
