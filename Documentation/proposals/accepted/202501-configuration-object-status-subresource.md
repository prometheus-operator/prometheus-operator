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
  * [workload status subresource proposal](../implemented/202409-status-subresource.md)

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

Prometheus operator allows users to define their observability workloads through "workload" resources like `Prometheus`, `PrometheusAgent`, `Alertmanager`. The configuration of these workloads can be done dynamically by orchestrating "configuration" resources like `ServiceMonitor`, `PodMonitor`, `ScrapeConfig`, etc.

Currently, the status subresource is only implemented for workload resources. The absence of the status subresource for configuration resources makes it difficult to determine the source of the generated configuration of the workload resources. Additionally, there is no straightforward way to observe the reconciliation status of configuration resources. While Kubernetes events are available when a configuration is rejected by a workload, they are not sufficient for ongoing visibility or troubleshooting.

## Goals

* Define the structure of the status subresource for the custom resource definitions
  * `ServiceMonitor`
  * `PodMonitor`
  * `ScrapeConfig`
  * `Probes`
  * `PrometheusRule`
  * `AlertmanagerConfig`
* Report when a configuration resource is considered invalid during reconciliation. For example:
  * Feature not being supported by the version of the workload.
  * Invalid configmap/secret key reference.
  * Invalid PromQL expression in PrometheusRule resources.

## Non-Goals

* The solution does not aim to expose the full live configuration or runtime status of Prometheus. Doing so would be expensive for the workload, the operator and the Kubernetes system in general.
  * It will not provide information about the targets being scraped and their status for scrape resources (`PodMonitor`, `ServiceMonitor`, `Probes` and `ScrapeConfig`).
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

### API

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
        name: main
        namespace: monitoring
        conditions:
          - type: Accepted
            status: "True"
            observedGeneration: 3
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: ""
            message: ""
      - group: monitoring.coreos.com
        resource: prometheuses
        name: example
        namespace: default
        conditions:
          - type: Accepted
            status: "False"
            observedGeneration: 2
            lastTransitionTime: "2024-02-08T23:52:22Z"
            reason: InvalidConfiguration
            message: "'KeepEqual' relabel action is only supported with Prometheus >= 2.41.0"
      - group: monitoring.coreos.com
        resource: prometheusagents
        name: agent
        namespace: monitoring
        conditions:
          - type: Accepted
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
          expr: sum(rate(container_cpu_usage_seconds_total{container!="", pod!=""}[5m)) by (pod) > 0.5
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
          - type: Accepted
            status: "False"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: InvalidConfiguration
            message: "rule 0, alert: 'HighPodCPUUsage', parse error: expected type vector in aggregation expression, got scalar"
```

#### `AlertmanagerConfig`

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
          - type: Accepted
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: ""
            message: ""
```

#### Details of the Status API fields

* `bindings`: Lists the workload resources that select the configuration resource.
  * `conditions`: Describes the latest conditions of the configuration resource in relation to the workload resource.
    * `type`: The only condition type used is `Accepted`, indicating whether the workload controller has successfully accepted the configuration resource and updated the configuration of the workload accordingly.
    * `status`: It can be either `True` , `False` or `Unknown`.
      * `True` indicates that the configuration resource was successfully accepted by the controller and written to the configuration secret.
      * `False` means the controller rejected the configuration due to an error.
    * `reason`: Specifies the reason why the configuration was not accepted.
    * `message`: Provides the detailed error message returned by the controller during reconciliation.
    * `observedGeneration`: Represents the generation of the configuration resource that the controller has most recently observed. When the value doesn't match the object metadata's `generation` value, the condition is stale.

### Implementation

#### Feature gate

This feature is controlled by a feature gate: `StatusForConfigurationResources`. Cluster administrators can toggle this flag to enable or disable the status subresource support for configuration resources as needed.

This feature gate is disabled by default. To enable it, pass the following argument to the prometheus-operator container during installation:

```bash
 --feature-gates=StatusForConfigurationResources=true
```

Once we've reached feature completeness and we're confident about the stability, we will toggle the feature gate to be enabled by default.

#### Keeping the status up-to-date

##### Removing a binding when the workload is deleted

When a workload resource is created, we add a finalizer to it to ensure proper cleanup before deletion. If a user later requests deletion of the resource, Kubernetes does not immediately remove it; instead, it sets a deletionTimestamp on the resource. This triggers an update event, which the workload controller receives and processes. When the deletionTimestamp is set, the controller proceeds to clean up the configuration resources (e.g., ServiceMonitors, PrometheusRules) with bindings to the workload. Once the cleanup is complete, the controller removes the finalizer from the workload resource, allowing Kubernetes to complete the deletion process.

##### Removing invalid bindings from configuration resources status

A configuration resource may contain a reference to a workload resource in its bindings which is not relevant anymore. This can occur for instance when:
* A workload resource A selects a configuration resource X in namespace Y.
* The operator updates the status of resource X to reference workload A.
* At a later time, changes may happen that break this association:
  * The labels of X and/or its namespace Y are modified.
  * The label selectors and/or namespace selectors of workload A are updated.

These changes can result in the workload A no longer selecting the configuration resource X which requires the operator to update the configuration resource's status.

A separate Go routine can be used to remove invalid bindings, offloading this responsibility from the main workload controller's reconciliation loop.
This background routine can periodically:
* Query all workload resources.
* List all configuration resources.
* For each configuration resource, check if its bindings are still valid.
* Remove any bindings that no longer have an active association with a workload resource.

This approach ensures that the cleanup process does not interfere with the primary reconciliation loop and improves controller efficiency.

## Alternatives

#### Dedicated CRD Approach for Configuration-Workload Mapping

A potential solution to mapping a configuration resource to a workload resource is the introduction of a Custom Resource Definition (CRD). This new CRD would act as an intermediary, maintaining a clear association between configurations and workloads.

The [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/introduction) handles the Secrets Provider pod status in a similar way.

It comes with the following drawbacks:
* Introducing a new CRD could lead to additional operational complexity.
* It requires installation, maintenance and versioning, which could increase the administrative burden.

#### Storing Information in the Workload Resource

Another approach is to store configuration mappings directly within the workload resource.

It comes with the following drawbacks:
* Owners of configuration resources not having permissions to view the workload resource won’t have a view of the status which is one of the main goals of this effort.
* Workload resources could reference a high number of configuration resources (it isn't uncommon for a Prometheus resource to select more than a hundred of service monitors + pod monitors + rules).
* Storing all these mappings within a single workload resource could lead to excessive API payload sizes.

## Action Plan

* Introduce a new feature gate.
* Manage finalizers on workload resources.
  * Add or remove a finalizer based on the feature gate status.
* Create bindings in associated configuration resources.
  * During workload reconciliation, populate status.bindings in matching configuration resources.
* Handle updates affecting selection logic
  * On changes to:
    * Namespace labels
    * Configuration resource labels
    * Workload’s label/namespace selectors
  * Recalculate and remove invalid bindings from affected configuration resources.
* Clean up bindings on workload deletion
  * When a workload is deleted, remove its references from all associated configuration resources

## Follow-ups

Once the goals of this proposal are achieved, we can extend the implementation to populate scrape targets information in the status subresource of ServiceMonitor, PodMonitor, ScrapeConfig, and Probe resources.
