# Monitoring other Kubernetes Namespaces
This guide will help you monitor applications in other Namespaces, which is only enabled for the `Default` Namespace during Install.

# Setup
You have to give the list of the Namespaces that you want to be able to monitor.
This is done in the variable `prometheus.roleSpecificNamespaces`. You usually set this in your `.jsonnet` file when building the manifests.

Ex to create the needed `Role` and `Rolebindig` for the Namespace `foo` : 
```
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',

    prometheus+:: {
      roleSpecificNamespaces: ["foo"],
    },
  },
};
 
{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }

```