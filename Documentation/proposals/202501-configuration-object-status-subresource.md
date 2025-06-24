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

> This proposal outlines the implementation of a Status subresource field extension to the Prometheus operator's configuration-based Custom Resource Definitions (CRDs).

## Why

The solution will allow users to verify whether their configurations have been successfully applied to the corresponding workload resources.

Mapping between configuration resources and their associated workload resources:

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
* Report when a configuration resource is considered invalid during reconciliation. For example:
  * Feature not being supported by the version of the workload.
  * Invalid configmap/secret key reference.
  * Invalid PromQL expression in PrometheusRule resources.

## Non-Goals

* The solution does not aim to expose the full live configuration or runtime status of Prometheus. Doing so would be expensive for the workload, the operator and the Kubernetes system in general.
  * It will not include per-target scrape status, only summary information (e.g., number of targets up/down).
  * It will not surface fired alerts from PrometheusRule resources, as querying Prometheus for this data can be expensive and places undue load on Prometheus, the operator and the Kubernetes platform in general.
* Configuration resources won't expose status information that explains why they are not being selected by Prometheus or why their targets are not being scraped.
  * This non-goal can be partially addressed by tools like [`poctl`](https://github.com/prometheus-operator/poctl), which provide insights into configuration resource selection and target matching.

### Audience

* Users of Prometheus-Operator
* Maintainers and Contributors of Prometheus-Operator

## How

There are different challenges that influence the API design:

* A single config resource can be selected by multiple workload resources.
* The configuration resource may not be in the same namespace as the workload resource.
* Which workload selects which configuration resources can vary over time depending on the workload resource's label selectors and on the configuration resource's labels.

### CRDs

#### `ServiceMonitor`/`PodMonitor`/`Probes`/`ScrapeConfig`

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-servicemonitor
  namespace: monitoring
  generation: 3
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
        targets: 
          up: 2
          down: 1
          lastCheckedTime: "2025-05-20T12:34:56Z"
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 3
            lastTransitionTime: "2025-05-20T12:34:56Z"
    bindings:
      - group: monitoring.coreos.com
        resource: prometheuses
        name: prometheus-example
        namespace: default
        targets: 
          up: 3
          down: 1
          lastCheckedTime: "2025-05-20T12:34:56Z"
        conditions:
          - type: Reconciled
            status: "False"
            observedGeneration: 2
            lastTransitionTime: "2024-02-08T23:52:22Z"
            reason: InvalidConfiguration
            message: "'KeepEqual' relabel action is only supported with Prometheus >= 2.41.0"
    bindings:
      - group: monitoring.coreos.com
        resource: prometheusagents
        name: prometheus-agent-main
        namespace: monitor
        targets: 
          up: 3
          down: 1
          lastCheckedTime: "2025-05-20T12:34:56Z"
        conditions:
          - type: Reconciled
            status: "False"
            observedGeneration: 3
            lastTransitionTime: "2024-02-08T23:52:22Z"
            reason: InvalidConfiguration
            message: "Referenced Secret 'my-secret' in namespace 'monitoring' is missing or does not contain the required key 'basic-auth-password'"
```

#### `PrometheusRule`

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: example-prometheus-rules
  namespace: monitoring
  generation: 1
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
            status: "False"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: InvalidConfiguration
            message: "rule 0, alert: 'HighPodCPUUsage', parse error: expected type vector in aggregation expression, got scalar"
```

#### `AlertManagerConfig`

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: minimal-alertmanager-config
  namespace: monitoring
  generation: 1
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

#### How to Enable/Disable the Status Subresource for Configuration Resources

This feature is controlled by a feature flag: `StatusForConfigurationResources=true|false`. Cluster administrators can toggle this flag to enable or disable the status subresource support for configuration resources as needed.

#### Details of the Status API fields

* `bindings`: Lists the workload resources that reference or use the configuration resource.
  * `conditions`: Describes the latest status of the configuration resource.
    * `type`: The only condition type used is `Reconciled`, which indicates that the workload resource controller has successfully discovered the associated configuration resource and completed the necessary reconciliation steps.
    * `status`: It can be either `true` or `false`.
      * `true` indicates that the configuration resource was successfully accepted by the controller and written to the configuration secret.
      * `false` means the controller rejected the configuration due to an error.
    * `reason`: Specifies the reason for rejection, if the configuration was not accepted.
    * `message`: Provides the detailed error message returned by the controller during reconciliation.
    * `observedGeneration`: Represents the generation of the configuration resource that the controller has most recently reconciled.

#### How to get targets information in scrape resources ?

The operator periodically sends a gRPC request to the `config-reloader` sidecar, including the configuration resources selector labels in the request body. Upon receiving the request, the sidecar queries the Prometheus container by calling the /api/v1/targets endpoint. It then filters and modifies the response based on the provided labels, removing sensitive information that could potentially be exploited by attackers. The sanitized response is finally sent back to the operator.

##### Authentication/Authorization of the request ?

To ensure secure and authenticated communication between the `config-reloader` sidecar and the `operator`, mutual TLS (mTLS) can be used. With mTLS, both parties validate each other’s identity and encrypt all communication, mitigating the risk of man-in-the-middle attacks or unauthorized access. The necessary TLS certificates and keys are provisioned as Kubernetes Secrets and mounted into both the Prometheus and operator pods. This setup ensures that only trusted components within the cluster can participate in the gRPC communication, further strengthening the overall security posture.

##### Alternative (Not recommended)

An alternative approach would involve allowing the workload's ServiceAccount to update its own binding in the status subresource of the configuration resource. In this method, the job_name field could be used to infer a reference to the originating resource—such as ServiceMonitor/namespace/name/index.

Cons:
* The workload's ServiceAccount would need additional permissions to update its own resource’s status subresource.
* Complexity in Resource Mapping.

#### How to remove a binding when the workload is deleted ?

When a workload resource is created, we add a finalizer to it to ensure proper cleanup before deletion. If a user later requests deletion of the resource, Kubernetes does not immediately remove it; instead, it sets a deletionTimestamp on the resource. This triggers an update event, which the controller receives and processes. The controller checks if the deletionTimestamp is set to determine if the resource is in the process of being deleted. If so, the controller proceeds to clean up associated references from relevant configuration resources (e.g., ServiceMonitors, PrometheusRules). Once the cleanup is complete, the controller removes the finalizer from the workload resource, allowing Kubernetes to complete the deletion process.

#### How to remove invalid bindings from config-resources status ?

A dedicated goroutine runs continuously to monitor configuration resources for any invalid bindings.

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
* Introducing a new CRD could lead to additional operational complexity.
* It requires installation, maintenance and versioning, which could increase the administrative burden.

#### Storing Information in the Workload Resource

Another approach is to store configuration mappings directly within the workload resource.

It comes with the following drawbacks:
* Workload resources could reference a high number of configuration resources.
* Storing all these mappings within a single workload resource could lead to excessive API payload sizes.
* Configuration resources won’t have a direct view of where their configurations are being used.

## Action Plan

...
