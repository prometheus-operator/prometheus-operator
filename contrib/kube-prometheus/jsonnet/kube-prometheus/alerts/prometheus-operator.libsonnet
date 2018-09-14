{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prometheus-operator',
        rules: [
          {
            alert: 'PrometheusOperatorAlertmanagerReconcileErrors',
            expr: |||
              rate(prometheus_operator_alertmanager_reconcile_errors_total{%(prometheusOperatorSelector)s}[5m]) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Errors while reconciling Alertmanager in {{$labels.namespace}} namespace.',
            },
            'for': '10m',
          },
          {
            alert: 'PrometheusOperatorPrometheusReconcileErrors',
            expr: |||
              rate(prometheus_operator_prometheus_reconcile_errors_total{%(prometheusOperatorSelector)s}[5m]) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Errors while reconciling Prometheus in {{$labels.namespace}} namespace.',
            },
            'for': '10m',
          },
          {
            alert: 'PrometheusOperatorNodeLookupErrors',
            expr: |||
              rate(prometheus_operator_node_address_lookup_errors_total{%(prometheusOperatorSelector)s}[5m]) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Errors while reconciling Prometheus in {{$labels.namespace}} namespace.',
            },
            'for': '10m',
          },
        ],
      },
    ],
  },
}
