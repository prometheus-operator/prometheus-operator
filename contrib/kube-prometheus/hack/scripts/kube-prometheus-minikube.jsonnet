local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-kubeadm.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-node-ports.libsonnet') +
  {
    _config+:: {
      namespace: 'monitoring',
    },
  };

{ ['0prometheus-operator-' + name + '.yaml']: std.manifestYamlDoc(kp.prometheusOperator[name]) for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name + '.yaml']: std.manifestYamlDoc(kp.nodeExporter[name]) for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name + '.yaml']: std.manifestYamlDoc(kp.kubeStateMetrics[name]) for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name + '.yaml']: std.manifestYamlDoc(kp.alertmanager[name]) for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name + '.yaml']: std.manifestYamlDoc(kp.prometheus[name]) for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name + '.yaml']: std.manifestYamlDoc(kp.grafana[name]) for name in std.objectFields(kp.grafana) }
