---
weight: 252
toc: true
title: Alerting Routes
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Alerting guide
---

This guide assumes you already have a basic understanding of the Prometheus Operator and have gone through the [Getting Started]({{< ref "getting-started" >}}) guide. We’re also expecting you to know how to run an Alertmanager instance.

In this guide, we'll explore the various methods for managing Alertmanager configurations within your Kubernetes cluster.

Prometheus' configuration also includes "rule files", which contain the
[alerting
rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/).
When an alerting rule is triggered, it fires that alert to ***all*** Alertmanager
instances, on ***every*** rule evaluation interval. The Alertmanager instances
communicate to each other which notifications have already been sent out. For
more information on this system design, see the [High Availability]({{< ref "high-availability" >}})
page.

By default, the Alertmanager instances will start with a minimal configuration
which isn't really useful since it doesn't send any notification when receiving
alerts.

You have several options to provide the [Alertmanager configuration](https://prometheus.io/docs/alerting/configuration/):
1. Using a native Alertmanager configuration file stored in a [Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/).
2. using `spec.alertmanagerConfiguration` to reference an
   `AlertmanagerConfig` object in the same namespace which defines the main
   Alertmanager configuration.
3. Using `spec.alertmanagerConfigSelector` and
   `spec.alertmanagerConfigNamespaceSelector` to tell the operator which
   `AlertmanagerConfig` objects should be selected and merged with the main
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
