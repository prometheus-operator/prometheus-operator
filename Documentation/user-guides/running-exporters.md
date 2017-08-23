# Running Exporters
Running exporters and scraping them with Prometheus configured by the prometheus-operator.

## The goal of ServiceMonitor(s)
The goal for one ServiceMonitor should be to cover a large number of `Service`s.
This can be achieved by creating a generic `ServiceMonitor`.

## Create a service that exposes Pod(s)
For scraping an exporter or separate metrics port you need a service that targets the pod(s) of the exporter or application.
An example for the `kube-state-metrics` is the below `Service` and a generic `ServiceMonitor` that covers more than just the `kube-state-metrics` `Service`.
After you have created the `Service`, you then continue on by creating the `ServiceMonitor` for it.
The order in which the `Service` and `ServiceMonitor` is created is not important.

### `kube-state-metrics` Service example
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
This Service targets all pods with the label `k8s-app: kube-state-metrics`.

## Create a matching ServiceMonitor
### Generic `ServiceMonitor` example
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
(A better example for monitoring Kubernetes cluster components can be found [User Guide "Cluster Monitoring"](user-guides/cluster-monitoring.md))
This ServiceMonitor targets **all** Services with the label `k8s-app` (`spec.selector`) any value, in the namespaces `kube-system` and `monitoring` (`spec.namespaceSelector`).

## Troubleshooting
### Namespace "limits"/things to keep in mind
See the ServiceMonitor Documentation:
> While `ServiceMonitor`s must live in the same namespace as the `Prometheus`
TPR, discovered targets may come from any namespace. This is important to allow
cross-namespace monitoring use cases, e.g. for meta-monitoring. Using the
`namespaceSelector` of the `ServiceMonitorSpec`, one can restrict the
namespaces the `Endpoints` objects are allowed to be discovered from.
