# Admission webhooks

This document describes how to deploy the Prometheus operator's admission webhook service.

The admission webhook can be used to:
* Ensure that all annotations of PrometheusRule objects are coerced into string values.
* Check that PrometheusRule objects are semantically valid.
* Check that AlertmanagerConfig objects are semantically valid.
* Convert AlertmanagerConfig objects between v1alpha1 and v1beta1 versions.

## Prerequisites

This guide assumes that you have already [deployed the Prometheus
Operator](getting-started.md) and that [admission controllers are
enabled](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller)
on your cluster.

The Kubernetes API server expects admission webhooks to communicate over HTTPS
so this guide also assumes the following:
1. A valid TLS certificate and key has been provisioned for the admission webhook service.
2. A secret holding the TLS certificate and key has been created.

If you don't want to manually provision the TLS materials,
[cert-manager](https://cert-manager.io/) is a good option.

## Admission webhook deployment

Assuming that the following secret exists and contains the base64-encoded
[PEM](https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail) certificate
(`tls.crt`) and key (`tls.key`) for the admission webhook service.

```yaml
apiVersion: v1
data:
  tls.crt: LS0tLS...LS0tCg==
  tls.key: LS0tLS...LS0tCg==
kind: Secret
metadata:
  name: admission-webhook-certs
  namespace: default
```

The admission webhook's pod template should be modified to mount the secret as a
volume and the container arguments should be modified to include the
certificate and key:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-operator-admission-webhook
  namespace: default
spec:
  template:
    spec:
      containers:
      - name: prometheus-operator-admission-webhook
        args:
        - --web.enable-tls=true
        - --web.cert-file=/etc/tls/private/tls.crt
        - --web.key-file=/etc/tls/private/tls.key
        volumeMounts:
        - mountPath: /etc/tls/private
          name: tls-certificates
          readOnly: true
     volumes:
     - name: tls-certificates
     - secret:
         secretName: admission-webhook-certs
```

## Webhook endpoints

### caBundle note

The `caBundle` field contains the base64-encoded CA certificate used to sign the
webhook's certificate. It is used by the Kubernetes API server to validate the
certificate of the webhook service.

Certificate managers like [cert-manager](https://cert-manager.io/) supports CA
injection into webhook configurations and custom resource definitions.

### `/admission-prometheusrules/validate`

The endpoint `/admission-prometheusrules/validate` rejects `PrometheusRule`
objects that are not valid.

The following example deploys the validating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesvalidation
webhooks:
  - clientConfig:
      caBundle: LS0tLS...LS0tCg==
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

The endpoint `/admission-prometheusrules/mutate` ensures that integers and
boolean yaml data elements are coerced into strings.

The following example deploys the mutating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesmutation
webhooks:
  - clientConfig:
      caBundle: LS0tLS...LS0tCg==
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

The endpoint `/admission-alertmanagerconfigs/validate` rejects
`AlertmanagerConfig` objects that are not valid.

The following example deploys the validating admission webhook:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-alertmanager-config-validation
webhooks:
  - clientConfig:
      caBundle: LS0tLS...LS0tCg==
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

### `/convert`

The endpoint `/convert` convert `Alertmanagerconfig` objects between `v1alpha1`
and `v1beta1` versions.

The following example is a patch for the
`alertmanagerconfigs.monitoring.coreos.com` CRD to configure the conversion
webhook.

```json
{
   "apiVersion": "apiextensions.k8s.io/v1",
   "kind": "CustomResourceDefinition",
   "metadata": {
      "name": "alertmanagerconfigs.monitoring.coreos.com"
   },
   "spec": {
      "conversion": {
         "strategy": "Webhook",
         "webhook": {
            "clientConfig": {
               "service": {
                  "name": "prometheus-operator-admission-webhook",
                  "namespace": "default",
                  "path": "/convert",
                  "port": 8443
               },
               "caBundle": "LS0tLS...LS0tCg=="
            },
            "conversionReviewVersions": [
               "v1beta1",
               "v1alpha1"
            ]
         }
      }
   }
}
```
