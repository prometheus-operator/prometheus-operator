{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prometheus-operator',
        rules: [
          {
            alert: 'PrometheusOperatorReconcileErrors',
            expr: |||
              rate(prometheus_operator_reconcile_errors_total{%(prometheusOperatorSelector)s}[5m]) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Errors while reconciling {{ $labels.controller }} in {{ $labels.namespace }} Namespace.',
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
              message: 'Errors while reconciling Prometheus in {{ $labels.namespace }} Namespace.',
            },
            'for': '10m',
          },
        ],
      },
    ],
  },
}
