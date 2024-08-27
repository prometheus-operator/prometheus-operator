---
weight: 102
toc: true
title: Installing Prometheus Operator
menu:
    docs:
        parent: getting-started
images: []
draft: false
description: Installation guide listing all the installation methods of Prometheus Operator.
date: "2020-11-16T13:59:39+01:00"
---

There are different approaches to install Prometheus Operator in your Kubernetes cluster:

- [Install using YAML files](#install-using-yaml-files)
- [Install using Kube-Prometheus](#install-using-kube-prometheus)
- [Install using Helm Chart](#install-using-helm-chart)

### Pre-requisites

For all the approaches listed on this page, you require access to a **Kubernetes cluster!** For this, you can check the official docs of Kubernetes available [here](https://kubernetes.io/docs/tasks/tools/).

Version `>=0.39.0` of the Prometheus Operator requires a Kubernetes cluster of version `>=1.16.0`. If you are just starting out with the Prometheus Operator, it is **highly recommended** to use the latest version. If you have an older version of Kubernetes and the Prometheus Operator running, we recommend upgrading Kubernetes first and then the Prometheus Operator.

> Check the appropriate versions of each of the components in the [Compatibility]({{<ref "compatibility">}}) page.

### Install using YAML files

The first step is to install the operator's Custom Resource Definitions (CRDs) as well as the operator itself with the required RBAC resources.

Run the following commands to install the CRDs and deploy the operator in the `default` namespace:

```bash
LATEST=$(curl -s https://api.github.com/repos/prometheus-operator/prometheus-operator/releases/latest | jq -cr .tag_name)
curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/${LATEST}/bundle.yaml | kubectl create -f -
```

It can take a few minutes for the operator to be up and running. You can check for completion with the following command:

```bash
kubectl wait --for=condition=Ready pods -l  app.kubernetes.io/name=prometheus-operator -n default
```

### Install using Kube-Prometheus

The easiest way of starting with the Prometheus Operator is by deploying it as part of kube-prometheus. kube-prometheus deploys the Prometheus Operator and already schedules a Prometheus called `prometheus-k8s` with alerts and rules by default.

We are going to deploy a compiled version of the Kubernetes [manifests](https://github.com/prometheus-operator/kube-prometheus/tree/main/manifests).

You can either clone the kube-prometheus from GitHub:

```shell
git clone https://github.com/prometheus-operator/kube-prometheus.git
```

or download the current main branch as zip file and extract its contents:

[github.com/prometheus-operator/kube-prometheus/archive/main.zip](https://github.com/prometheus-operator/kube-prometheus/archive/main.zip)

Once you have the files on your machine change into the project's root directory and run the following commands:

```shell
# Create the namespace and CRDs, and then wait for them to be available before creating the remaining resources
kubectl create -f manifests/setup

# Wait until the "servicemonitors" CRD is created. The message "No resources found" means success in this context.
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done

kubectl create -f manifests/
```

We create the namespace and CustomResourceDefinitions first to avoid race conditions when deploying the monitoring components. Alternatively, the resources in both folders can be applied with a single command:

```
kubectl create -f manifests/setup -f manifests
```

But it may be necessary to run the command multiple times for all components to be created successfully.

> Note: For versions before Kubernetes v1.20.z refer to the [Kubernetes compatibility matrix](https://github.com/prometheus-operator/kube-prometheus#kubernetes-compatibility-matrix) in order to choose a compatible branch.

> Note: If you used Kube-Prometheus as the installation method, we would recommend you to follow this [page](http://prometheus-operator.dev/kube-prometheus/kube/access-ui/) to learn how to access the resources provided.

### Remove Kube-Prometheus

If you're done experimenting with kube-prometheus and the Prometheus Operator you can simply teardown the deployment by running:

```shell
kubectl delete --ignore-not-found=true -f manifests/ -f manifests/setup
```

### Install Using Helm Chart

Install the [Kube-Prometheus-Stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) helm chart which provides a collection of Kubernetes manifests, [Grafana](https://grafana.com/) dashboards, and [Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with [Prometheus](https://prometheus.io/) using the Prometheus Operator.

To see more details, please check the [chart's README](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack#kube-prometheus-stack).

> This helm chart is no longer part of Prometheus-Operator and is now maintained by [Prometheus Community Helm Charts](https://github.com/prometheus-community/helm-charts).
