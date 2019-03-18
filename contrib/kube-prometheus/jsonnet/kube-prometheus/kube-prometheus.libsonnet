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
    dashboardDefinitions: configMapList.new(super.dashboardDefinitions),
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata: {
        name: 'grafana',
        namespace: $._config.namespace,
      },
      spec: {
        selector: {
          matchLabels: {
            app: 'grafana',
          },
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

    versions+:: {
      grafana: '6.0.1',
    },

    tlsCipherSuites: [
      'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256',  // required by h2: http://golang.org/cl/30721
      'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256',  // required by h2: http://golang.org/cl/30721

      // 'TLS_RSA_WITH_RC4_128_SHA',            // insecure: https://access.redhat.com/security/cve/cve-2013-2566
      // 'TLS_RSA_WITH_3DES_EDE_CBC_SHA',       // insecure: https://access.redhat.com/articles/2548661
      // 'TLS_RSA_WITH_AES_128_CBC_SHA',        // disabled by h2
      // 'TLS_RSA_WITH_AES_256_CBC_SHA',        // disabled by h2
      'TLS_RSA_WITH_AES_128_CBC_SHA256',
      // 'TLS_RSA_WITH_AES_128_GCM_SHA256',     // disabled by h2
      // 'TLS_RSA_WITH_AES_256_GCM_SHA384',     // disabled by h2
      // 'TLS_ECDHE_ECDSA_WITH_RC4_128_SHA',    // insecure: https://access.redhat.com/security/cve/cve-2013-2566
      // 'TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA',// disabled by h2
      // 'TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA',// disabled by h2
      // 'TLS_ECDHE_RSA_WITH_RC4_128_SHA',      // insecure: https://access.redhat.com/security/cve/cve-2013-2566
      // 'TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA', // insecure: https://access.redhat.com/articles/2548661
      // 'TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA',  // disabled by h2
      // 'TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA',  // disabled by h2
      'TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256',
      'TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256',

      // disabled by h2 means: https://github.com/golang/net/blob/e514e69ffb8bc3c76a71ae40de0118d794855992/http2/ciphers.go

      // 'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384',   // TODO: Might not work with h2
      // 'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384', // TODO: Might not work with h2
      // 'TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305',    // TODO: Might not work with h2
      // 'TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305',  // TODO: Might not work with h2
    ],

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

    alertmanagerSelector: 'job="alertmanager-main",namespace="' + $._config.namespace + '"',
    prometheusSelector: 'job="prometheus-' + $._config.prometheus.name + '",namespace="' + $._config.namespace + '"',
    prometheusOperatorSelector: 'job="prometheus-operator",namespace="' + $._config.namespace + '"',

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
