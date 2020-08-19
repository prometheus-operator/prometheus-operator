{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prometheus-operator',
        rules: [
          {
            alert: 'PrometheusOperatorWatchErrors',
            expr: |||
              (sum by (controller,namespace) (rate(prometheus_operator_watch_operations_failed_total{%(prometheusOperatorSelector)s}[1h])) / sum by (controller,namespace) (rate(prometheus_operator_watch_operations_total{%(prometheusOperatorSelector)s}[1h]))) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: 'Errors while performing watch operations in controller {{$labels.controller}} in {{$labels.namespace}} namespace.',
              summary: 'Errors while performing watch operations in controller.',
            },
            'for': '15m',
          },
          {
            alert: 'PrometheusOperatorReconcileErrors',
            expr: |||
              (sum by (controller,namespace) (rate(prometheus_operator_reconcile_errors_total{%(prometheusOperatorSelector)s}[5m])) / (sum by (controller,namespace) (rate(prometheus_operator_reconcile_operations_total{%(prometheusOperatorSelector)s}[5m])) > 0.1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: '{{ $value | humanizePercentage }} of reconciling operations failed for {{ $labels.controller }} controller in {{ $labels.namespace }} namespace.',
              summary: 'Errors while reconciling controller.',
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
              description: 'Errors while reconciling Prometheus in {{ $labels.namespace }} Namespace.',
              summary: 'Errors while reconciling Prometheus.',
            },
            'for': '10m',
          },
        ],
      },
    ],
  },
}
