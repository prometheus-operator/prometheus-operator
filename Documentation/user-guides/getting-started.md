---
weight: 251
toc: true
title: Getting Started
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Getting started guide
---

The Prometheus Operator introduces custom resources in Kubernetes to declare
the desired state of a Prometheus and Alertmanager cluster as well as the
Prometheus configuration.

The `Prometheus` resource declaratively describes the desired state of a
Prometheus deployment, while `ServiceMonitor` and `PodMonitor` resources
describe the targets to be monitored by Prometheus.

This guide explains how to use `PodMonitor` and `ServiceMonitor` objects to monitor targets for a sample application.

## Pre-requisites

You will need a Kubernetes cluster with admin permissions to follow this guide.

To install Prometheus Operator, please refer to the instructions [here]({{<ref "installation.md">}}).

Also, refer to the Platform Guide to learn how to deploy a Prometheus instance before moving ahead.

<!-- do not change this link without verifying that the image will display correctly on https://prometheus-operator.dev -->

![Prometheus Operator Architecture](../img/serviceMonitor-and-podMonitor.png)

> Note: Check the [Alerting guide]({{< ref "alerting" >}}) for more information about the `Alertmanager` resource.

> Note: Check the [Design page]({{< ref "design" >}}) for an overview of all resources introduced by the Prometheus Operator.

## Deploying a sample application

First, let's deploy a simple example application with 3 replicas that listens
and exposes metrics on port `8080`.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-deployment.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      labels:
        app: example-app
    spec:
      containers:
      - name: example-app
        image: quay.io/brancz/prometheus-example-app:v0.5.0
        ports:
        - name: web
          containerPort: 8080
```

Let's expose the application with a Service object that selects all the Pods
with the `app` label having the `example-app` value. The Service object also
specifies the port on which the metrics are exposed.

## Using PodMonitors

We can utilize a `PodMonitor` object to monitor the pods. The `spec.selector` label specifies which Pods Prometheus should scrape.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-pod-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  podMetricsEndpoints:
  - port: web
```

Similarly, the Prometheus object defines which PodMonitors get selected with the
`spec.podMonitorSelector` field.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-pod-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  podMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```

## Using ServiceMonitors

To monitor the application using a `ServiceMonitor`, let’s create a Service object. This Service will select all Pods with the label `app` set to `example-app` and specify the port where the metrics are exposed.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-service.yaml"
kind: Service
apiVersion: v1
metadata:
  name: example-app
  labels:
    app: example-app
spec:
  selector:
    app: example-app
  ports:
  - name: web
    port: 8080
```

Finally, we create a `ServiceMonitor` object that selects all Service objects
with the `app: example-app` label. The ServiceMonitor object also has a `team`
label (in this case `team: frontend`) to identify which team is responsible for
monitoring the application/service.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-service-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
```

Similarly, the Prometheus object defines which ServiceMonitors get selected with the
`spec.serviceMonitorSelector` field.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-service-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```
