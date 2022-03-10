# Admission webhooks

This document describes how to set up an admission webhook to validate
PrometheusRules, and thus preventing Prometheus from loading invalid
configuration.

## Prerequisites

This guide assumes that you have already [deployed the Prometheus
Operator](getting-started.md) and that [admission
controllers are
enabled](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller)
on your cluster.

Admission webhooks require TLS, and as such this guide also assumes that you
have a TLS certificate and key ready.

## Preparing the Operator

A secret needs to be created from the TLS certificate and key, assuming the
certificate is in `tls.crt` and the key in `tls.key`:

```bash
kubectl create secret tls prometheus-operator-certs --cert=tls.crt --key=tls.key
```

The Prometheus Operator will serve the admission webhook. However, to do so, it
requires being available over TLS, and not only plain HTTP. Thus the following
flags need to be added to the Prometheus Operator deployment:

* `--web.enable-tls=true` to enable the Prometheus Operator to serve its API
  over TLS,

* `--web.cert-file` to load the TLS certificate to use,

* `--web.key-file` to load the associate key.

## Webhook endpoints

### caBundle note

The `caBundle` contains the base64-encoded CA certificate used to sign the
webhook's certificate.

### `/admission-prometheusrules/validate`

The endpoint `/admission-prometheusrules/validate` rejects rules that are not valid prometheus config.

The following example deploys the validating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesvalidation
webhooks:
  - clientConfig:
      caBundle: SOMECABASE64ENCODED==
      service:
        name: prometheus-operator
        namespace: default
        path: /admission-prometheusrules/validate
    failurePolicy: Fail
    name: prometheusrulemutate.monitoring.coreos.com
    namespaceSelector: {}
    rules:
      - apiGroups:
          - monitoring.coreos.com
        apiVersions:
          - '*'
        operations:
          - CREATE
          - UPDATE
        resources:
          - prometheusrules
    admissionReviewVersions: ["v1", "v1beta1"]
    sideEffects: None
```

### `/admission-prometheusrules/mutate`

The endpoint `/admission-prometheusrules/mutate` ensures that integers and boolean yaml data elements are cooerced into strings.

The following example deploys the mutating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesmutation
webhooks:
  - clientConfig:
      caBundle: SOMECABASE64ENCODED==
      service:
        name: prometheus-operator
        namespace: default
        path: /admission-prometheusrules/mutate
    failurePolicy: Fail
    name: prometheusrulemutate.monitoring.coreos.com
    namespaceSelector: {}
    rules:
      - apiGroups:
          - monitoring.coreos.com
        apiVersions:
          - '*'
        operations:
          - CREATE
          - UPDATE
        resources:
          - prometheusrules
    admissionReviewVersions: ["v1", "v1beta1"]
    sideEffects: None
```

### `/admission-alertmanagerconfigs/validate`

The endpoint `/admission-alertmanagerconfigs/validate` rejects alertmanagerconfigs that are not valid alertmanager config.

The following example deploys the validating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-alertmanager-config-validation
webhooks:
  - clientConfig:
      caBundle: SOMECABASE64ENCODED==
      service:
        name: prometheus-operator
        namespace: default
        path: /admission-alertmanagerconfigs/validate
    failurePolicy: Fail
    name: alertmanagerconfigsvalidate.monitoring.coreos.com
    namespaceSelector: {}
    rules:
      - apiGroups:
          - monitoring.coreos.com
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - alertmanagerconfigs
    admissionReviewVersions: ["v1", "v1beta1"]
    sideEffects: None
```
