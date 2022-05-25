{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prometheus-operator',
        rules: [
          {
            alert: 'PrometheusOperatorListErrors',
            expr: |||
              (sum by (%(groupLabels)s) (rate(prometheus_operator_list_operations_failed_total{%(prometheusOperatorSelector)s}[10m])) / sum by (%(groupLabels)s) (rate(prometheus_operator_list_operations_total{%(prometheusOperatorSelector)s}[10m]))) > 0.4
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: 'Errors while performing List operations in controller {{$labels.controller}} in {{$labels.namespace}} namespace.',
              summary: 'Errors while performing list operations in controller.',
            },
            'for': '15m',
          },
          {
            alert: 'PrometheusOperatorWatchErrors',
            expr: |||
              (sum by (%(groupLabels)s) (rate(prometheus_operator_watch_operations_failed_total{%(prometheusOperatorSelector)s}[5m])) / sum by (%(groupLabels)s) (rate(prometheus_operator_watch_operations_total{%(prometheusOperatorSelector)s}[5m]))) > 0.4
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
            alert: 'PrometheusOperatorSyncFailed',
            expr: |||
              min_over_time(prometheus_operator_syncs{status="failed",%(prometheusOperatorSelector)s}[5m]) > 0
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: 'Controller {{ $labels.controller }} in {{ $labels.namespace }} namespace fails to reconcile {{ $value }} objects.',
              summary: 'Last controller reconciliation failed',
            },
            'for': '10m',
          },
          {
            alert: 'PrometheusOperatorReconcileErrors',
            expr: |||
              (sum by (%(groupLabels)s) (rate(prometheus_operator_reconcile_errors_total{%(prometheusOperatorSelector)s}[5m]))) / (sum by (%(groupLabels)s) (rate(prometheus_operator_reconcile_operations_total{%(prometheusOperatorSelector)s}[5m]))) > 0.1
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
          {
            alert: 'PrometheusOperatorNotReady',
            expr: |||
              min by (%(groupLabels)s) (max_over_time(prometheus_operator_ready{%(prometheusOperatorSelector)s}[5m]) == 0)
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: "Prometheus operator in {{ $labels.namespace }} namespace isn't ready to reconcile {{ $labels.controller }} resources.",
              summary: 'Prometheus operator not ready',
            },
            'for': '5m',
          },
          {
            alert: 'PrometheusOperatorRejectedResources',
            expr: |||
              min_over_time(prometheus_operator_managed_resources{state="rejected",%(prometheusOperatorSelector)s}[5m]) > 0
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: 'Prometheus operator in {{ $labels.namespace }} namespace rejected {{ printf "%0.0f" $value }} {{ $labels.controller }}/{{ $labels.resource }} resources.',
              summary: 'Resources rejected by Prometheus operator',
            },
            'for': '5m',
          },
        ],
      },
      {
        name: 'config-reloaders',
        rules: [
          {
            alert: 'ConfigReloaderSidecarErrors',
            expr: |||
              max_over_time(reloader_last_reload_successful{%(configReloaderSelector)s}[5m]) == 0
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              description: 'Errors encountered while the {{$labels.pod}} config-reloader sidecar attempts to sync config in {{$labels.namespace}} namespace.
As a result, configuration for service running in {{$labels.pod}} may be stale and cannot be updated anymore.',
              summary: 'config-reloader sidecar has not had a successful reload for 10m',
            },
            'for': '10m',
          },
        ],
      },
    ],
  },
}
