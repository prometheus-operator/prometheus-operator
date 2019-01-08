local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local configMapList = k.core.v1.configMapList;

(import 'grafana/grafana.libsonnet') +
(import 'kube-state-metrics/kube-state-metrics.libsonnet') +
(import 'node-exporter/node-exporter.libsonnet') +
(import 'alertmanager/alertmanager.libsonnet') +
(import 'prometheus-operator/prometheus-operator.libsonnet') +
(import 'prometheus/prometheus.libsonnet') +
(import 'prometheus-adapter/prometheus-adapter.libsonnet') +
(import 'kubernetes-mixin/mixin.libsonnet') +
(import 'alerts/alerts.libsonnet') +
(import 'rules/rules.libsonnet') + {
  kubePrometheus+:: {
    namespace: k.core.v1.namespace.new($._config.namespace),
  },
  grafana+:: {
    local serviceLabels = {
      app: 'grafana',
    },
    dashboardDefinitions: configMapList.new(super.dashboardDefinitions),
    service+: {
      labels+: serviceLabels,
    },
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata: {
        name: 'grafana',
        namespace: $._config.namespace,
      },
      spec: {
        selector: {
          matchLabels: serviceLabels,
        },
        endpoints: [
          {
            port: 'http',
            interval: '15s',
          },
        ],
      },
    },
  },
} + {
  _config+:: {
    namespace: 'default',

    cadvisorSelector: 'job="kubelet"',
    kubeletSelector: 'job="kubelet"',
    kubeStateMetricsSelector: 'job="kube-state-metrics"',
    nodeExporterSelector: 'job="node-exporter"',
    notKubeDnsSelector: 'job!="kube-dns"',
    kubeSchedulerSelector: 'job="kube-scheduler"',
    kubeControllerManagerSelector: 'job="kube-controller-manager"',
    kubeApiserverSelector: 'job="apiserver"',
    coreDNSSelector: 'job="kube-dns"',
    podLabel: 'pod',

    alertmanagerSelector: 'job="alertmanager-main"',
    prometheusSelector: 'job="prometheus-k8s"',
    prometheusOperatorSelector: 'job="prometheus-operator"',

    jobs: {
      Kubelet: $._config.kubeletSelector,
      KubeScheduler: $._config.kubeSchedulerSelector,
      KubeControllerManager: $._config.kubeControllerManagerSelector,
      KubeAPI: $._config.kubeApiserverSelector,
      KubeStateMetrics: $._config.kubeStateMetricsSelector,
      NodeExporter: $._config.nodeExporterSelector,
      Alertmanager: $._config.alertmanagerSelector,
      Prometheus: $._config.prometheusSelector,
      PrometheusOperator: $._config.prometheusOperatorSelector,
      CoreDNS: $._config.coreDNSSelector,
    },

    prometheus+:: {
      rules: $.prometheusRules + $.prometheusAlerts,
    },

    grafana+:: {
      dashboards: $.grafanaDashboards,
    },
  },
}
