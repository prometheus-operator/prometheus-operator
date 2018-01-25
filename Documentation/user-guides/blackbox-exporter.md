<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# Blackbox exporter

The (blackbox exporter)[https://github.com/prometheus/blackbox_exporter] needs to be
passed the target as a parameter, relabelings can be used to label metrics with the target.


```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: blackbox-exporter
  namespace: default
  labels:
    app: blackbox-exporter
spec:
  selector:
    matchLabels:
      app: blackbox-exporter
  endpoints:
  - port: web
    path: /probe
    params:
      module:
        - http_2xx
      target:
        - http://domain.com
    relabelings:
      - sourceLabels:
          - __param_target
        targetLabel: target
      - sourceLabels:
          - __param_module
        targetLabel: module
```
