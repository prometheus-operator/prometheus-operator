## Status Subresource For Config-Based Resources

* **Owners:**
  * [yp969803](https://github.com/yp969803)
* **Status:**
  * `In-Progress`
* **Related Tickets:**
  * [#3385](https://github.com/prometheus-operator/prometheus-operator/issues/3335)
* **Other docs:**
  * NA

> This proposal describes how we will extend the Prometheus operator configuration Custom
> Resource Definitions (CRDs) with a Status subresource field.

## Why

This will enhance observability by allowing users to verify whether their configurations have been successfully applied to Workload.

## Pitfalls of the current solution

Prometheus operator allows users to define their observability workloads through "workload" resources like `Prometheus`, `PrometheusAgent`, `AlertManager`. The configuration of these workloads can be done dynamically by orchestrating "configuration" resources like `ServiceMonitor`, `PodMonitor`, `ScrapeConfig`, etc which in turn are used to generate the configuration for the workloads that the Prometheus-Operator manages.

Currently, the status subresource is only implemented for workload resources. The absence of the status subresource for configuration resources makes it difficult to determine the source of the generated configuration of the workload resources. Additionally, there is no straightforward way to observe the reconciliation status of configuration resources. While Kubernetes events are available when a configuration is rejected by a workload, they are not sufficient for ongoing visibility or troubleshooting.

## Goals

* Define the structure of the status subresource for the custom resource definitions
  * `ServiceMonitor`
  * `PodMonitor`
  * `ScrapeConfig`
  * `Probes`
  * `PrometheusRule`
  * `AlertmanagerConfig`
* Information about number of Up/Down targets in status subresource for PodMonitor, ServiceMonitor, Probes and ScrapeConfig.
* Define how the operator would reconcile the status subresource.
* Feature gate to help the user to enable/disable the feature.

### Audience

* Users of Prometheus-Operator
* Maintainers and Contributors of Prometheus-Operator

## Non-Goals

* The status subresource is intended to offer only a summary (e.g., counts of up/down targets) of the scrape status.
* No information about the fired alerts in the status subresource of PrometheusRule (too heavy for the operator to query the prometheus pod for the alerts information. Prometheus may have a significant number of alerts, and fetching them repeatedly increases significant load in the operator).
* The status subresource is not intended to provide realtime information of the targets.

## How

### CRDs

#### `ServiceMonitor`

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
  status:
    bindings:
      - resource: prometheus
        name: prometheus-main
        namespace: monitoring
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: ReconcileSucceeded
            message: "Successfully reconciled with Prometheus"
        pods:
          - name: prometheus-main-0
            namespace: monitoring
            endpoint: http://10.42.0.10:9090
            totalScrapedTargets: 3
            totalUpTargets: 2
            totalDownTargets: 1
            lastCheckedTime: "2025-05-20T12:34:56Z"
```

CRDs of podMonitor, scrapeCOnfig and Probes looks similar to above.

The operator sends the GET request at regular interval to the config-reloader sidecar with the serviceMonitor selector labels as the request body, the sidecar container after receiving the request sends the /api/v1/targets request to prometheus-container, from the response it gets from the prometheus, it modifies the response based on the labels and send the response to the operator. The sidecar will remove information from the response which can be used by attackers.

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
      - resource: prometheus
        name: prometheus-main
        namespace: monitoring
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: ReconcileSucceeded
            message: "Successfully reconciled with Prometheus"
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
      - resource: alertmanager
        name: alertmanager-main
        namespace: monitoring
        conditions:
          - type: Reconciled
            status: "True"
            observedGeneration: 1
            lastTransitionTime: "2025-05-20T12:34:56Z"
            reason: ReconcileSucceeded
            message: "Successfully reconciled with Prometheus"
```

Finalizers are used during the deletion of the config-resource and workload-resource to clear the refrences in the status-subresource.

## Alternatives

#### Dedicated CRD Approach for Configuration-Workload Mapping

A potential solution to mapping a configuration resource to a workload resource is the introduction of a Custom Resource Definition (CRD). This new CRD would act as an intermediary, maintaining a clear association between configurations and workloads.

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
