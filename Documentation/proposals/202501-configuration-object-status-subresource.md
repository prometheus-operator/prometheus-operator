## Status Subresource For Config-Based Resources

* **Owners:**
  * [yp969803](https://github.com/yp969803)
* **Status:**
  * `Accepted`
* **Related Tickets:**
  * [#3385](https://github.com/prometheus-operator/prometheus-operator/issues/3335)
* **Other docs:**
  * https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
  * https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md
  * [workload status subresource proposal](202409-status-subresource.md)

> This proposal describes how we will extend the Prometheus operator configuration Custom
> Resource Definitions (CRDs) with a Status subresource field.

## Why

This will allow users to verify whether their configurations have been successfully applied to the corresponding workload resources.

Mapping between configuration resources and their associated workloads

| Configuration Resource | Workload Resource              |
|------------------------|--------------------------------|
| ServiceMonitor         | Prometheus and PrometheusAgent |
| PodMonitor             | Prometheus and PrometheusAgent |
| Probes                 | Prometheus and PrometheusAgent |
| ScrapeConfig           | Prometheus and PrometheusAgent |
| PrometheusRule         | Prometheus and ThanosRuler     |
| AlertmanagerConfig     | Alertmanager                   |

## Pitfalls of the current solution

Prometheus operator allows users to define their observability workloads through "workload" resources like `Prometheus`, `PrometheusAgent`, `AlertManager`. The configuration of these workloads can be done dynamically by orchestrating "configuration" resources like `ServiceMonitor`, `PodMonitor`, `ScrapeConfig`, etc.

Currently, the status subresource is only implemented for workload resources. The absence of the status subresource for configuration resources makes it difficult to determine the source of the generated configuration of the workload resources. Additionally, there is no straightforward way to observe the reconciliation status of configuration resources. While Kubernetes events are available when a configuration is rejected by a workload, they are not sufficient for ongoing visibility or troubleshooting.

## Goals

* Define the structure of the status subresource for the custom resource definitions
  * `ServiceMonitor`
  * `PodMonitor`
  * `ScrapeConfig`
  * `Probes`
  * `PrometheusRule`
  * `AlertmanagerConfig`
* Provide information about the targets being scraped and their status for scrape resources (`PodMonitor`, `ServiceMonitor`, `Probes` and `ScrapeConfig`).
* Reporting in status section when a configuration resource is considered invalid during reconciliation. Examples include:
  * The Prometheus or Alertmanager version does not support a specific feature.
  * Invalid configmap/secret key reference.

## Non-Goals

* The status subresource is intended to offer only a summary (e.g., counts of up/down targets) of the scrape status.
* No information about the fired alerts in the status subresource of PrometheusRule (too heavy for the operator to query the prometheus pod for the alerts information. Prometheus may have a significant number of alerts, and fetching them repeatedly increases significant load in the operator).
* The status subresource is not intended to provide realtime information of the targets.
* Configuration resources like ServiceMonitor or PodMonitor do not expose status information that explains why they are not being selected by Prometheus or why their targets are not being scraped

### Audience

* Users of Prometheus-Operator
* Maintainers and Contributors of Prometheus-Operator

## How

Challenges that influenced the API design :

* A single config resource can be selected by different workload resources.
* The config resource might not live in the same namespace as the workload.

### CRDs

#### `ServiceMonitor`/`PodMonitor`/`Probes`/`ScrapeConfig`

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-servicemonitor
  namespace: monitoring
  labels:
    team: backend
spec:
  selector:
    matchLabels:
      app: my-service
  namespaceSelector:
    matchNames:
      - default
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scheme: http
      basicAuth:
        username:
          name: my-secret
          key: basic-auth-username
        password:
          name: my-secret
          key: basic-auth-password
  status:
    bindings:
      - group: monitoring.coreos.com
        resource: prometheuses
        name: prometheus-main
        namespace: monitoring
        totalScrapedTargets: 3
        totalUpTargets: 2
        totalDownTargets: 1
        lastCheckedTime: "2025-05-20T12:34:56Z"
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 3
            lastTransitionTime: "2025-05-20T12:34:56Z"
          - type: Reconciled
            status: "False"
            observedGeneration: 2
            lastTransitionTime: "2024-02-08T23:52:22Z"
            reason: InvalidResource
            message: "'KeepEqual' relabel action is only supported with Prometheus >= 2.41.0"
          - type: Reconciled
            status: "False"
            observedGeneration: 1
            lastTransitionTime: "2024-02-08T23:52:22Z"
            reason: InvalidSecret
            message: "Referenced Secret 'my-secret' in namespace 'monitoring' is missing or does not contain the required key 'basic-auth-password'."
```

#### `PrometheusRule`

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: example-prometheus-rules
  namespace: monitoring
  labels:
    prometheus: k8s
    role: alert-rules
spec:
  groups:
    - name: example.rules
      interval: 30s
      rules:
        - alert: HighPodCPUUsage
          expr: sum(rate(container_cpu_usage_seconds_total{container!="", pod!=""}[5m])) by (pod) > 0.5
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High CPU usage on pod {{ $labels.pod }}"
            description: "Pod {{ $labels.pod }} is using more than 0.5 cores for 5 minutes."
  status:
    bindings:
      - group: monitoring.coreos.com
        resource: prometheuses
        name: prometheus-main
        namespace: monitoring
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
```

#### `AlertManagerConfig`

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: minimal-alertmanager-config
  namespace: monitoring
spec:
  route:
    receiver: "webhook-receiver"
  receivers:
    - name: "webhook-receiver"
      webhookConfigs:
        - url: "http://my-webhook-service.monitoring.svc:8080/"
          sendResolved: true
  status:
    bindings:
      - group: monitoring.coreos.com
        resource: alertmanagers
        name: alertmanager-main
        namespace: monitoring
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
```

### Working

#### How to get targets information in scrape resources ?

The operator sends the GET request at regular interval to the config-reloader sidecar with the serviceMonitor selector labels as the request body, the sidecar container after receiving the request sends the /api/v1/targets request to prometheus-container, from the response it gets from the prometheus, it modifies the response based on the labels and send the response to the operator. The sidecar will remove critical informations from the response which can be used by attackers.

#### How to clear refrences in status section if the workload is deleted ?

Finalizers are used during the deletion of the config-resource and workload-resource to clear the refrences in the status-subresource.

## Alternatives

#### Dedicated CRD Approach for Configuration-Workload Mapping

A potential solution to mapping a configuration resource to a workload resource is the introduction of a Custom Resource Definition (CRD). This new CRD would act as an intermediary, maintaining a clear association between configurations and workloads.

The Secrets Store CSI Driver handles the Secrets Provider pod status in a similar way.

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClassPodStatus
metadata:
  creationTimestamp: "2021-01-21T19:20:11Z"
  generation: 1
  labels:
    internal.secrets-store.csi.k8s.io/node-name: kind-control-plane
    manager: secrets-store-csi
    operation: Update
    time: "2021-01-21T19:20:11Z"
  name: nginx-secrets-store-inline-crd-dev-azure-spc
  namespace: dev
  ownerReferences:
  - apiVersion: v1
    kind: Pod
    name: nginx-secrets-store-inline-crd
    uid: 10f3e31c-d20b-4e46-921a-39e4cace6db2
  resourceVersion: "1638459"
  selfLink: /apis/secrets-store.csi.x-k8s.io/v1/namespaces/dev/secretproviderclasspodstatuses/nginx-secrets-store-inline-crd
  uid: 1d078ad7-c363-4147-a7e1-234d4b9e0d53
status:
  mounted: true
  objects:
  - id: secret/secret1
    version: c55925c29c6743dcb9bb4bf091be03b0
  - id: secret/secret2
    version: 7521273d0e6e427dbda34e033558027a
  podName: nginx-secrets-store-inline-crd
  secretProviderClassName: azure-spc
  targetPath: /var/lib/kubelet/pods/10f3e31c-d20b-4e46-921a-39e4cace6db2/volumes/kubernetes.io~csi/secrets-store-inline/mount
```

It comes with the following drawbacks:
- Introducing a new CRD means extra operational complexity.
- It requires installation, maintenance, and versioning, which could increase the administrative burden.

#### Storing Information in the Workload Resource

Another approach is to store configuration mappings directly within the workload resource.

It comes with the following drawbacks:
- Workload resources could reference a high number of configuration resources.
- Storing all these mappings within a single workload resource could lead to excessive API payload sizes.
- Configuration resources won’t have a direct view of where their configurations are being used.

## Action Plan

...
