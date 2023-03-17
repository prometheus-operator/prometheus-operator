<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.
</div>

# Running Exporters

Running exporters and scraping them with Prometheus configured by the prometheus-operator.

## The goal of ServiceMonitor(s)

The goal for one ServiceMonitor should be to cover a large number of Services. This can be achieved by creating a generic `ServiceMonitor`.

## Create a service that exposes Pod(s)

Scraping an exporter or separate metrics port requires a service that targets the Pod(s) of the exporter or application.

The following examples create a Service for the `kube-state-metrics`, and a generic ServiceMonitor which may be used for more than simply the `kube-state-metrics` Service.

The order in which the `Service` and `ServiceMonitor` is created is not important.

### kube-state-metrics Service example

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: kube-state-metrics
    k8s-app: kube-state-metrics
  annotations:
    alpha.monitoring.coreos.com/non-namespaced: "true"
  name: kube-state-metrics
spec:
  ports:
  - name: http-metrics
    port: 8080
    targetPort: metrics
    protocol: TCP
  selector:
    k8s-app: kube-state-metrics
```

This Service targets all Pods with the label `k8s-app: kube-state-metrics`.

### Generic ServiceMonitor example

This ServiceMonitor targets **all** Services with the label `k8s-app` (`spec.selector`) any value, in the namespaces `kube-system` and `monitoring` (`spec.namespaceSelector`).

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: k8s-apps-http
  labels:
    k8s-apps: http
spec:
  jobLabel: k8s-app
  selector:
    matchExpressions:
    - {key: k8s-app, operator: Exists}
  namespaceSelector:
    matchNames:
    - kube-system
    - monitoring
  endpoints:
  - port: http-metrics
    interval: 15s
```

## Default Labels

By default, the `PodMonitor` and `ServiceMonitor` objects include runtime metadata in the scraped results.

### PodMonitors

| Target Label | Source Label                         | Description                                                                                     |
|--------------|--------------------------------------|-------------------------------------------------------------------------------------------------|
| instance     | __param_target                       | The address of the scraped target                                                               |
| job          | -                                    | `{metadata.namespace}/{metadata.name}` of the `PodMonitor` or read from `jobLabel` if specified |
| namespace    | __meta_kubernetes_namespace          | `{metadata.namespace}` of the scraped pod                                                       |
| container    | __meta_kubernetes_pod_container_name | `{name}` of the container in the scraped pod                                                    |
| pod          | __meta_kubernetes_pod_name           | `{metadata.name}` of the scraped pod                                                            |
| endpoint     | -                                    | `{spec.Port}` or `{spec.TargetPort}` if specified                                               |

### ServiceMonitors

| Target Label | Source Label                         | Description                                                                   |
|--------------|--------------------------------------|-------------------------------------------------------------------------------|
| instance     | __param_target                       | The address of the scraped target                                             |
| job          | -                                    | `{metadata.name}` of the scraped service or read from `jobLabel` if specified |
| node/pod     | -                                    | Set depending on the endpoint responding to service request                   |
| namespace    | __meta_kubernetes_namespace          | `{metadata.namespace}` of the scraped pod                                     |
| service      |                                      | `{metadata.name}` of the scraped service                                      |
| pod          | __meta_kubernetes_pod_name           | `{metadata.name}` of the scraped pod                                          |
| container    | __meta_kubernetes_pod_container_name | `{name}` of the container in the scraped pod                                  |
| endpoint     | -                                    | `{spec.Port}` or `{spec.TargetPort}` if specified                             |

### Relabeling and Metric Relabeling

The Prometheus Operator provides the same capabilities as Prometheus to relabel a target before scrape or a metric before sample ingestion, below you can find examples for Service or Pod monitors.

#### Relabeling

Relabeling is a powerful feature to dynamically rewrite the label set of a target before it gets scraped, and multiple relabeling steps can be configured per scrape configuration.

> Relabel configs are applied to the label set of each target in order of their appearance in the configuration file.

**Dropping label from a target**

The following snippet drops the `pod` label from every metric scraped as part of the scrape job.

```yaml
- action: labeldrop
  regex: pod
```

**Adding label to a target**

The following snippet will add or replace the `team` label with the value `prometheus` for all the metrics scraped as part of this job.

```yaml
- action: replace
  replacement: prometheus
  targetLabel: team
```

**Filtering targets by label**

The following snippet will configure Prometheus to scrape metrics from the targets if they have the Kubernetes `team` label set to `prometheus` and the Kubernetes `datacenter` label not set to `west_europe`.

```yaml
- sourceLabels:
  - __meta_kubernetes_pod_label_team
  regex: "prometheus"
  action: keep
- sourceLabels:
  - __meta_kubernetes_pod_label_datacenter
  regex: west_europe
  action: drop
```

**Full example**

The following `ServiceMonitor` configures Prometheus to only select targets that have the `team` label set to `prometheus` and exclude the ones that have `datacenter` set to `west_europe`. The same configuration may be used with a `PodMonitor`.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
    relabelings:
      - sourceLabels:
          - __meta_kubernetes_pod_label_team
        regex: "prometheus"
        action: keep
      - sourceLabels:
          - __meta_kubernetes_pod_label_datacenter
        regex: west_europe
        action: drop
```

#### Metric Relabeling

Metric relabeling is applied to samples as the last step before ingestion, and it has the same configuration format and actions as target relabeling.

> Metric relabeling does not apply to automatically generated timeseries such as up.

**Dropping metrics**

The following snippet drops any metric which name (`__name__`) matches the regex `container_tasks_state`.

```yaml
metricRelabelings:
- sourceLabels:
  - __name__
  regex: container_tasks_state
  action: drop
```

**Dropping time series**

The following snippet drops metrics where the `id` label matches the regex `/system.slice/var-lib-docker-containers.*-shm.mount`.

```yaml
metricRelabelings:
- sourceLabels:
  - id
  regex: '/system.slice/var-lib-docker-containers.*-shm.mount'
  action: drop
```

**Full example**

The following `PodMonitor` configures Prometheus to drop metrics where the `id` label matches the regex `/system.slice/var-lib-docker-containers.*-shm.mount`. The same configuration could also be used with a `ServiceMonitor`

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
    metricRelabelings:
    - sourceLabels:
      - id
      regex: '/system.slice/var-lib-docker-containers.*-shm.mount'
      action: drop
```

## Troubleshooting

### Namespace "limits"/things to keep in mind

See the ServiceMonitor Documentation:

> By default and **before the version v0.19.0**, `ServiceMonitors` must be installed in the same namespace as the Prometheus resource. With the Prometheus Operator **v0.19.0 and above**, `ServiceMonitors` can be selected outside the Prometheus namespace via the `serviceMonitorNamespaceSelector` field of the Prometheus resource. The discovered targets may come from any namespace. This allows cross-namespace monitoring use cases, for example, for meta-monitoring.
