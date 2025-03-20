---
weight: 201
toc: true
title: Getting Started
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Getting started page for Platform Guide
---

This guide assumes you have a basic understanding of the Prometheus Operator. If you are new to it, please start with the [Introduction]({{<ref "introduction.md">}}) page before proceeding. This guide will walk you through deploying Prometheus and Alertmanager instances.

## Deploying Prometheus

To deploy a Prometheus instance, you must create the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/authorization/) rules for the Prometheus service account.

First, create a ServiceAccount for Prometheus.

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-service-account.yaml"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
```

Next, create a ClusterRole that grants Prometheus the necessary permissions to discover and scrape the targets within the cluster.

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-cluster-role.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/metrics
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs: ["get", "list", "watch"]
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

Now, create a ClusterRoleBinding to bind the ClusterRole to the Prometheus ServiceAccount.

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-cluster-role-binding.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: default
```

Apply all these manifests to create the necessary RBAC resources. Now you are all set to deploy a Prometheus instance. Here is an example of a basic Prometheus instance manifest.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
```

To verify that the instance is up and running, run:

```bash
kubectl get -n default prometheus prometheus -w
```

For more information, see the [Prometheus Operator RBAC guide]({{< ref "rbac" >}}).

## Deploying Alertmanager

Let us take a simple example that creates 3 replicas of Alertmanager.

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-example.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: example
spec:
  replicas: 3
```

Wait for all Alertmanager pods to be ready:

```bash
kubectl get pods -l alertmanager=example -w
```

However, Alertmanager as it is now is of no use to us. To properly use Alertmanager, it is important to understand the relationship between Prometheus and Alertmanager. Alertmanager is used to:

* Deduplicate alerts received from Prometheus.
* Silence alerts.
* Route and send grouped notifications to various integrations (PagerDuty, OpsGenie, mail, chat, …).

So, to put Alertmanager instances to use, you would need to integrate it with Prometheus.

## Integrating Alertmanager With Prometheus

### Exposing the Alertmanager service

To access the Alertmanager interface, you have to expose the service to the outside. For
simplicity, we use a `NodePort` Service.

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-example-service.yaml"
apiVersion: v1
kind: Service
metadata:
  name: alertmanager-example
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30903
    port: 9093
    protocol: TCP
    targetPort: web
  selector:
    alertmanager: example
```

Once the Service is created, the Alertmanager web server is available under the
node's IP address on port `30903`.

> Note: Exposing the Alertmanager web server this way may not be an applicable solution. Read more about the possible options in the [Ingress guide]({{<ref "exposing-prometheus-and-alertmanager.md">}}).

### Configuring Alertmanager in Prometheus

The Alertmanager cluster is now fully functional and highly available, but no
alerts are fired against it.

First, create a Prometheus instance that will send alerts to the Alertmanger cluster:

```
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: example
spec:
  serviceAccountName: prometheus
  replicas: 2
  alerting:
    alertmanagers:
    - namespace: default
      name: alertmanager-example
      port: web
```

The `Prometheus` resource discovers all of the Alertmanager instances behind
the `Service` created before (pay attention to `name`, `namespace` and `port`
fields which should match with the definition of the Alertmanager Service).

Open the Prometheus web interface, go to the "Status > Runtime & Build
Information" page and check that the Prometheus has discovered 3 Alertmanager
instances.
