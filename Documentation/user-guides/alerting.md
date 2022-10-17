---
weight: 152
toc: true
title: Alerting
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Alerting guide
---

This guide assumes that you have a basic understanding of the Prometheus
operator, and that you have already followed the [Getting Started]({{< ref
"getting-started" >}}) guide.

{{< alert icon="👉" text="Prometheus Operator requires use of Kubernetes v1.16.x and up."/>}}

The Prometheus Operator introduces an `Alertmanager` resource, which allows
users to declaratively describe an Alertmanager cluster. To successfully deploy
an Alertmanager cluster, it is important to understand the contract between
Prometheus and Alertmanager. Alertmanager is used to:

* Deduplicate alerts received from Prometheus.
* Silence alerts.
* Route and send grouped notifications to various integrations (PagerDuty, OpsGenie, mail, chat, ...).

The Prometheus Operator also introduces an `AlertmanagerConfig` resource, which
allows users to declaratively describe Alertmanager configurations.

> Note: The AlertmanagerConfig resource is currently v1alpha1, testing and feedback are welcome.

Prometheus' configuration also includes "rule files", which contain the
[alerting
rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/).
When an alerting rule triggers, it fires that alert against *all* Alertmanager
instances, on *every* rule evaluation interval. The Alertmanager instances
communicate to each other which notifications have already been sent out. For
more information on this system design, see the [High Availability]({{< ref "high-availability" >}})
page.

## Pre-requisites

You have a running Prometheus operator.

## Deploying Alertmanager

First, let's create a Alertmanager cluster with three replicas:

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

## Managing Alertmanager configuration

By default, the Alertmanager instances will start with a minimal configuration
which isn't really useful since it doesn't send any notification when receiving
alerts.

You have several options to provide the [Alertmanager configuration](https://prometheus.io/docs/alerting/configuration/):
1. You can use a native Alertmanager configuration file stored in a Kubernetes secret.
2. You can use `spec.alertmanagerConfiguration` to reference an
   AlertmanagerConfig object in the same namespace which defines the main
   Alertmanager configuration.
3. You can define `spec.alertmanagerConfigSelector` and
   `spec.alertmanagerConfigNamespaceSelector` to tell the operator which
   AlertmanagerConfigs objects should be selected and merged with the main
   Alertmanager configuration.

### Using a Kubernetes Secret

The following native Alertmanager configuration sends notifications to a fictuous webhook service:

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager.yaml"
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://example.com/'
```

Save the above configuration in a file called `alertmanager.yaml` in the local directory and create a Secret from it:

```bash
kubectl create secret generic alertmanager-example --from-file=alertmanager.yaml
```

The Prometheus operator requires the Secret to be named like
`alertmanager-{ALERTMANAGER_NAME}`. In the previous example, the name of the
Alertmanager is `example`, so the secret name must be `alertmanager-example`.
The name of the key holding the configuration data in the Secret has to be
`alertmanager.yaml`.

> Note: if you want to use a different secret name, you can specify it with the `spec.configSecret` field in the Alertmanager resource.

The Alertmanager configuration may reference custom templates or password files
on disk. These can be added to the Secret along with the `alertmanager.yaml`
configuration file. For example, provided that we have the following Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: alertmanager-example
data:
  alertmanager.yaml: {BASE64_CONFIG}
  template_1.tmpl: {BASE64_TEMPLATE_1}
  template_2.tmpl: {BASE64_TEMPLATE_2}
```

Templates will be accessible to the Alertmanager container under the
`/etc/alertmanager/config` directory. The Alertmanager
configuration can reference them like this:

```yaml
templates:
- '/etc/alertmanager/config/*.tmpl'
```

### Using AlertmanagerConfig Resources

The following example configuration creates an AlertmanagerConfig resource that
sends notifications to a fictitious webhook service.

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-config-example.yaml"
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: config-example
  labels:
    alertmanagerConfig: example
spec:
  route:
    groupBy: ['job']
    groupWait: 30s
    groupInterval: 5m
    repeatInterval: 12h
    receiver: 'webhook'
  receivers:
  - name: 'webhook'
    webhookConfigs:
    - url: 'http://example.com/'
```

Create the AlertmanagerConfig resource in your cluster:

```bash
curl -sL https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/user-guides/alerting/alertmanager-config-example.yaml | kubectl create -f -
```

The `spec.alertmanagerConfigSelector` field in the Alertmanager resource
needs to be updated so the operator selects AlertmanagerConfig resources. In
the previous example, the label `alertmanagerConfig: example` is added, so the
Alertmanager object should be updated like this:

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-selector-example.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: example
spec:
  replicas: 3
  alertmanagerConfigSelector:
    matchLabels:
      alertmanagerConfig: example
```

### Using AlertmanagerConfig for global configuration

The following example configuration creates an Alertmanager resource that uses
an AlertmanagerConfig resource to be used for the Alertmanager configuration
instead of the `alertmanager-example` secret.

```yaml mdox-exec="cat example/user-guides/alerting/alertmanager-example-alertmanager-configuration.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: example
  namespace: default
spec:
  replicas: 3
  alertmanagerConfiguration:
    name: config-example
```

The AlertmanagerConfig resource named `example-config` in namespace `default`
will be a global AlertmanagerConfig. When the operator generates the
Alertmanager configuration from it, the namespace label will not be enforced
for routes and inhibition rules.

## Exposing the Alertmanager service

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

> Note: Exposing the Alertmanager web server this way may not be an applicable solution. Read more about the possible options in the [Ingress guide](exposing-prometheus-and-alertmanager.md).

## Integrating with Prometheus

### Configuring Alertmanager in Prometheus

This Alertmanager cluster is now fully functional and highly available, but no
alerts are fired against it.

First, create a Prometheus instance that will send alerts to the Alertmanger cluster:

```yaml mdox-exec="cat example/user-guides/alerting/prometheus-example.yaml"
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
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  ruleSelector:
    matchLabels:
      role: alert-rules
      prometheus: example
```

The `Prometheus` resource discovers all of the Alertmanager instances behind
the `Service` created before (pay attention to `name`, `namespace` and `port`
fields which should match with the definition of the Alertmanager Service).

Open the Prometheus web interface, go to the "Status > Runtime & Build
Information" page and check that the Prometheus has discovered 3 Alertmanager
instances.

### Deploying Prometheus Rules

The `PrometheusRule` CRD allows to define alerting and recording rules. The
operator knows which PrometheusRule objects to select for a given Prometheus
based on the `spec.ruleSelector` field.

> Note: by default, `spec.ruleSelector` is nil meaning that the operator picks up no rule.

By default, the Prometheus resources discovers only `PrometheusRule` resources
in the same namespace. This can be refined with the `ruleNamespaceSelector` field:
* To discover rules from all namespaces, pass an empty dict (`ruleNamespaceSelector: {}`).
* To discover rules from all namespaces matching a certain label, use the `matchLabels` field.

Discover `PrometheusRule` resources with `role=alert-rules` and
`prometheus=example` labels from all namespaces with `team=frontend` label:

```yaml mdox-exec="cat example/user-guides/alerting/prometheus-example-rule-namespace-selector.yaml"
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
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  ruleSelector:
    matchLabels:
      role: alert-rules
      prometheus: example
  ruleNamespaceSelector:
    matchLabels:
      team: frontend
```

In case you want to select individual namespace by their name, you can use the
`kubernetes.io/metadata.name` label, which gets populated automatically with
the
[`NamespaceDefaultLabelName`](https://kubernetes.io/docs/reference/labels-annotations-taints/#kubernetes-io-metadata-name)
feature gate.

Create a PrometheusRule object from the following manifest. Note that the
object's labels match with the `spec.ruleSelector` of the Prometheus object.

```yaml mdox-exec="cat example/user-guides/alerting/prometheus-example-rules.yaml"
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  creationTimestamp: null
  labels:
    prometheus: example
    role: alert-rules
  name: prometheus-example-rules
spec:
  groups:
  - name: ./example.rules
    rules:
    - alert: ExampleAlert
      expr: vector(1)
```

For demonstration purposes, the PrometheusRule object always fires the
`ExampleAlert` alert. To validate that everything is working properly, you can
open again the Prometheus web interface and go to the Alerts page.

Next open the Alertmanager web interface and check that it shows one active alert.
