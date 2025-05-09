---
weight: 203
toc: true
title: Admission webhook
menu:
    docs:
        parent: user-guides
lead: ""
images: []
draft: false
description: Guide to deploy and run the admission webhook service
---

This guide describes how to deploy and use the Prometheus operator's admission webhook service.

The admission webhook service is able to
* Validate requests ensuring that `PrometheusRule` and `AlertmanagerConfig` objects
  are semantically valid.
* Mutate requests enforcing that all annotations of `PrometheusRule` objects are
  coerced into string values.
* Convert `AlertmanagerConfig` objects between `v1alpha1` and `v1beta1` versions.

This guide assumes that you have already [deployed the Prometheus
Operator]({{< ref "docs/developer/getting-started.md" >}}) and that [admission controllers are
enabled](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller)
on your cluster.

## Prerequisites

The Kubernetes API server expects admission webhook services to communicate
over HTTPS so we need:
1. Valid TLS certificate and key provisioned for the admission webhook service.
2. Kubernetes Secret containing the TLS certificate and key.

For this guide, we assume that a Secret named `admission-webhook-certs` exists
in the same namespace as the webhook deployment and that it contains the
base64-encoded [PEM](https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail)
certificate (`tls.crt`) and key (`tls.key`) for the admission webhook service.

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

The recommended approach is to use [cert-manager](https://cert-manager.io/)
which manages both the lifecycle of the TLS certificates and the integration
with the Kubernetes API with respect to the webhook configuration (e.g.
automatic injection of the CA bundle).

While installing cert-manager is beyond the scope of this guide, below is an
example of a `Certificate` object which triggers the creation of the
`admission-webhook-certs` secret using a [SelfSigned
issuer](https://cert-manager.io/docs/configuration/selfsigned/). The
certificate is valid for the Kubernetes service
`prometheus-operator-admission-webhook` in the `default` namespace.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: prometheus-operator-admission-webhook
  namespace: default
spec:
  dnsNames:
    - prometheus-operator-admission-webhook.default.svc
  secretName: admission-webhook-certs
  issuerRef:
    name: selfsigned-cluster-issuer
    kind: ClusterIssuer
```

## Deploying the admission webhook

You can apply the following manifests to run a deployment of the webhook with 2 replicas.

```yaml mdox-exec="cat example/admission-webhook/service-account.yaml"
apiVersion: v1
automountServiceAccountToken: false
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: prometheus-operator-admission-webhook
    app.kubernetes.io/version: 0.82.2
  name: prometheus-operator-admission-webhook
  namespace: default
```

```yaml mdox-exec="cat example/admission-webhook/deployment.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: prometheus-operator-admission-webhook
    app.kubernetes.io/version: 0.82.2
  name: prometheus-operator-admission-webhook
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: prometheus-operator-admission-webhook
  strategy:
    rollingUpdate:
      maxUnavailable: 1
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: prometheus-operator-admission-webhook
      labels:
        app.kubernetes.io/name: prometheus-operator-admission-webhook
        app.kubernetes.io/version: 0.82.2
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app.kubernetes.io/name: prometheus-operator-admission-webhook
            namespaces:
            - default
            topologyKey: kubernetes.io/hostname
      automountServiceAccountToken: false
      containers:
      - args:
        - --web.enable-tls=true
        - --web.cert-file=/etc/tls/private/tls.crt
        - --web.key-file=/etc/tls/private/tls.key
        image: quay.io/prometheus-operator/admission-webhook:v0.82.2
        name: prometheus-operator-admission-webhook
        ports:
        - containerPort: 8443
          name: https
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 50m
            memory: 50Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /etc/tls/private
          name: tls-certificates
          readOnly: true
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        seccompProfile:
          type: RuntimeDefault
      serviceAccountName: prometheus-operator-admission-webhook
      volumes:
      - name: tls-certificates
        secret:
          items:
          - key: tls.crt
            path: tls.crt
          - key: tls.key
            path: tls.key
          secretName: admission-webhook-certs
```

You can now expose the webhook as a Kubernetes service by applying the following manifest.

```yaml mdox-exec="cat example/admission-webhook/service.yaml"
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: prometheus-operator-admission-webhook
    app.kubernetes.io/version: 0.82.2
  name: prometheus-operator-admission-webhook
  namespace: default
spec:
  ports:
  - name: https
    port: 443
    targetPort: https
  selector:
    app.kubernetes.io/name: prometheus-operator-admission-webhook
```

## Managing webhook configurations

Once the Prometheus operator's admission webhook service is up and running, you
can create `ValidatingWebhookConfiguration` and/or `MutatingWebhookConfiguration`
API objects that defines when/how the Kubernetes API should contact the service.

For more details, refer to the [Kubernetes
documentation](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#webhook-configuration).

### PrometheusRule

#### Validating PrometheusRule resources

The `/admission-prometheusrules/validate` endpoint of the admission webhook
service ensures that `PrometheusRule` objects are semantically valid.

The following example configures a validating admission webhook rejecting
invalid PrometheusRule objects.

> Note: If you're not using cert-manager, check the [CA Bundle]({{< ref "#ca-bundle" >}}) section.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesvalidation
  annotations:
    cert-manager.io/inject-ca-from: default/prometheus-operator-admission-webhook
webhooks:
  - clientConfig:
      service:
        name: prometheus-operator-admission-webhook
        namespace: default
        path: /admission-prometheusrules/validate
    failurePolicy: Fail
    name: prometheusrulevalidate.monitoring.coreos.com
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

#### Mutating PrometheusRule resources

The `/admission-prometheusrules/mutate` endpoint mutates `PrometheusRule`
objects so that integer and boolean YAML data elements are coerced into
strings.

The following example deploys a mutating admission webhook.

> Note: If you're not using cert-manager, check the [CA Bundle]({{< ref "#ca-bundle" >}}) section.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: prometheus-operator-rulesmutation
  annotations:
    cert-manager.io/inject-ca-from: default/prometheus-operator-admission-webhook
webhooks:
  - clientConfig:
      service:
        name: prometheus-operator-admission-webhook
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

### AlertmanagerConfig

The `/admission-alertmanagerconfigs/validate` endpoint rejects
`AlertmanagerConfig` objects that are not semantically valid.

The following example configures a validating admission webhook rejecting
invalid `AlertmanagerConfig` objects.

> Note: If you're not using cert-manager, check the [CA Bundle]({{< ref "#ca-bundle" >}}) section.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prometheus-operator-alertmanager-config-validation
  annotations:
    cert-manager.io/inject-ca-from: default/prometheus-operator-admission-webhook
webhooks:
  - clientConfig:
      service:
        name: prometheus-operator-admission-webhook
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

## Converting AlertmanagerConfig resources

The `/convert` endpoint converts `Alertmanagerconfig` objects between `v1alpha1`
and `v1beta1` versions.

For more details, refer to the [Kubernetes
documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#webhook-conversion).

The following command patches the
`alertmanagerconfigs.monitoring.coreos.com` CRD to enable the conversion.

```bash
cat <<EOF | kubectl patch crds/alertmanagerconfigs.monitoring.coreos.com --patch-file /dev/stdin
{
   "spec": {
      "conversion": {
         "strategy": "Webhook",
         "webhook": {
            "clientConfig": {
               "service": {
                  "name": "prometheus-operator-admission-webhook",
                  "namespace": "default",
                  "path": "/convert",
                  "port": 443
               }
            },
            "conversionReviewVersions": [
               "v1beta1",
               "v1alpha1"
            ]
         }
      }
   }
}
EOF
```

Annotate the AlertmanagerConfig CRD to let cert-manager inject the service CA bundle.

```bash
kubectl annotate crds alertmanagerconfigs.monitoring.coreos.com cert-manager.io/inject-ca-from=default/prometheus-operator-admission-webhook
```

> Note: If you're not using cert-manager, check the [CA Bundle]({{< ref "#ca-bundle" >}}) section.

## CA bundle

When contacting the webhook service during request admissions or CRD
conversion, the Kubernetes API verifies the server certificate using the
`caBundle` field defined in `clientConfig`. The field should contain the
base64-encoded CA certificate that signed the webhook's TLS certificate.

Certificate managers like [cert-manager](https://cert-manager.io/) supports
automatic CA injection into webhook configurations and custom resource
definitions. If you are **not using** a certificate manager, you need to
manually specify the `caBundle` field in `ValidatingWebhookConfiguration`,
`MutatingWebhookConfiguration` and `CustomResourceDefinition`.
