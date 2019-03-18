local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') + {
    _config+:: {
      namespace: 'monitoring',
    },
  };

local kustomization = {
  apiVersion: 'kustomize.config.k8s.io/v1beta1',
  kind: 'Kustomization',
  resources:
    ['00namespace-' + name + '.yaml' for name in std.objectFields(kp.kubePrometheus)] +
    ['0prometheus-operator-' + name + '.yaml' for name in std.objectFields(kp.prometheusOperator)] +
    ['node-exporter-' + name + '.yaml' for name in std.objectFields(kp.nodeExporter)] +
    ['kube-state-metrics-' + name + '.yaml' for name in std.objectFields(kp.kubeStateMetrics)] +
    ['alertmanager-' + name + '.yaml' for name in std.objectFields(kp.alertmanager)] +
    ['prometheus-' + name + '.yaml' for name in std.objectFields(kp.prometheus)] +
    ['prometheus-adapter-' + name + '.yaml' for name in std.objectFields(kp.prometheusAdapter)] +
    ['grafana-' + name + '.yaml' for name in std.objectFields(kp.grafana)],
};

local manifests =
  { ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
  { ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
  { ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
  { ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
  { ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
  { ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
  { ['prometheus-adapter-' + name]: kp.prometheusAdapter[name] for name in std.objectFields(kp.prometheusAdapter) } +
  { ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) };

manifests { kustomization: kustomization }
