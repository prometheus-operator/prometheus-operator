local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

(import 'grafana/grafana.libsonnet') +
(import 'kube-state-metrics/kube-state-metrics.libsonnet') +
(import 'node-exporter/node-exporter.libsonnet') +
(import 'alertmanager/alertmanager.libsonnet') +
(import 'prometheus-operator/prometheus-operator.libsonnet') +
(import 'prometheus/prometheus.libsonnet') +
(import 'kubernetes-mixin/mixin.libsonnet') + {
  kubePrometheus+:: {
    namespace: k.core.v1.namespace.new($._config.namespace),
  },
} + {
  _config+:: {
    namespace: 'default',

    kubeStateMetricsSelector: 'job="kube-state-metrics"',
    cadvisorSelector: 'job="kubelet"',
    nodeExporterSelector: 'job="node-exporter"',
    kubeletSelector: 'job="kubelet"',
    notKubeDnsSelector: 'job!="kube-dns"',

    prometheus+:: {
      rules: $.prometheusRules + $.prometheusAlerts,
    },

    grafana+:: {
      dashboards: $.grafanaDashboards,
    },
  },
}
