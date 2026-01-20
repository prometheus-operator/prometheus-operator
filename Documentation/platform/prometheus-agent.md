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
    app.kubernetes.io/version: 0.88.0
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
  - scrapeconfigs/status
  - servicemonitors
  - servicemonitors/status
  - podmonitors
  - podmonitors/status
  - probes
  - probes/status
  - prometheusrules
  - prometheusrules/status
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
  - events.k8s.io
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

## Deployment Modes

PrometheusAgent supports two deployment modes that determine how the Prometheus Agent pods are deployed and managed.

### StatefulSet Mode (Default)

This is the default deployment mode where PrometheusAgent is deployed as a StatefulSet. This mode is suitable for:

- **Cluster-wide monitoring**: One or more high-availability Prometheus Agents scrape metrics from the entire cluster.
- **Persistent storage requirements**: When you need persistent volumes for WAL (Write-Ahead Log) storage.
- **Centralized management**: Easier to manage fewer agent instances with predictable scaling.

### DaemonSet Mode (Alpha)

In DaemonSet mode, PrometheusAgent is deployed as a DaemonSet, running one pod per node. This mode is ideal for:

- **Node-local monitoring**: Each agent only scrapes metrics from targets on the same node.
- **Automatic scalability**: Agents automatically scale with node additions/removals.
- **Load distribution**: Load is naturally distributed across nodes.
- **Resource efficiency**: Lower memory usage and no persistent storage requirements.

DaemonSet mode works best with `PodMonitor` resources since each agent naturally discovers and scrapes pods running on the same node.

### Comparison of Deployment Modes

| Feature               | StatefulSet Mode            | DaemonSet Mode            |
|-----------------------|-----------------------------|---------------------------|
| **Scaling**           | Manual shard management     | Automatic with node count |
| **Scraping scope**    | Cluster-wide scraping       | Node-local scraping       |
| **Storage**           | Supports persistent storage | Ephemeral storage only    |
| **Target Discovery**  | ServiceMonitor & PodMonitor | PodMonitor recommended    |
| **Resource Usage**    | Higher memory usage         | Lower memory per pod      |
| **High Availability** | Multi-replica support       | One pod per node          |
| **Use Case**          | Centralized monitoring      | Distributed monitoring    |

## Enabling DaemonSet Mode

To use DaemonSet mode, you need to:

1. **Enable the feature gate** on the Prometheus Operator:

   ```bash
   --feature-gates=PrometheusAgentDaemonSet=true
   ```

2. **Ensure DaemonSet RBAC permissions** are granted to the operator. The operator needs permissions to manage DaemonSet resources:

   ```yaml
   - apiGroups:
     - apps
     resources:
     - daemonsets
     verbs:
     - '*'
   ```

3. **Set the mode field** in your PrometheusAgent spec:

   ```yaml
   apiVersion: monitoring.coreos.com/v1alpha1
   kind: PrometheusAgent
   metadata:
     name: prometheus-agent-daemonset
   spec:
     mode: DaemonSet
     serviceAccountName: prometheus-agent
     serviceMonitorSelector:
       matchLabels:
         team: frontend
   ```

## Field Restrictions in DaemonSet Mode

When using DaemonSet mode, the following fields are **not allowed** and will be rejected by CEL validation:

- `replicas`
- `storage`
- `shards` (cannot be greater than 1)
- `persistentVolumeClaimRetentionPolicy`
- `scrapeConfigSelector`
- `scrapeConfigNamespaceSelector`
- `probeSelector`
- `probeNamespaceSelector`
- `serviceMonitorSelector`
- `serviceMonitorNamespaceSelector`
- `additionalScrapeConfigs`

## Target Discovery in DaemonSet Mode

### PrometheusAgent Configuration

Here's a minimal PrometheusAgent configuration for DaemonSet mode:

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: PrometheusAgent
metadata:
  name: prometheus-agent-daemonset
spec:
  mode: DaemonSet
  serviceAccountName: prometheus-agent
  podMonitorSelector:
    matchLabels:
      team: backend
  podMonitorNamespaceSelector:
    matchLabels:
      monitoring: enabled
```

### PodMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: app-podmonitor
  namespace: default
spec:
  selector:
    matchLabels:
      app: my-application
  podMetricsEndpoints:
  - port: metrics
    path: /metrics
```

Continue with the [Getting Started page]({{<ref "docs/developer/getting-started.md">}}) to learn how to monitor applications running on Kubernetes.
