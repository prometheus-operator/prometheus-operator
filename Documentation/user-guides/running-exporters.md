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

```
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
    app: kube-state-metrics
```

This Service targets all Pods with the label `k8s-app: kube-state-metrics`.

### Generic ServiceMonitor example

This ServiceMonitor targets **all** Services with the label `k8s-app` (`spec.selector`) any value, in the namespaces `kube-system` and `monitoring` (`spec.namespaceSelector`).

```
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

## Troubleshooting

### Namespace "limits"/things to keep in mind

See the ServiceMonitor Documentation:

> While `ServiceMonitors` must live in the same namespace as the Prometheus
resource, discovered targets may come from any namespace. This allows
cross-namespace monitoring use cases, for example, for meta-monitoring. Use the
`namespaceSelector` of the `ServiceMonitorSpec` to restrict the
namespaces from which Endpoints objects may be discovered.
