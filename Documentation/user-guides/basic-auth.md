<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

## Basic auth for targets

To authenticate a `ServiceMonitor`s over a metrics endpoint use [`basicAuth`](../api.md#basicauth)

[embedmd]:# (../../contrib/kube-prometheus/examples/basic-auth/service-monitor.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-apps: basic-auth-example
  name: basic-auth-example
spec:
  endpoints:
  - basicAuth:
      password:
        name: basic-auth
        key: password
      username:
        name: basic-auth
        key: user
    port: metrics
  namespaceSelector:
    matchNames:
    - logging
  selector:
    matchLabels:
      app: myapp
```

[embedmd]:# (../../contrib/kube-prometheus/examples/basic-auth/secrets.yaml)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: basic-auth
data:
  password: dG9vcg== # toor
  user: YWRtaW4= # admin
type: Opaque
```
