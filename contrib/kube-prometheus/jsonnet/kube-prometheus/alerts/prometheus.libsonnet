{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prometheus.rules',
        rules: [
          {
            alert: 'PrometheusConfigReloadFailed',
            annotations: {
              description: "Reloading Prometheus' configuration has failed for {{$labels.namespace}}/{{$labels.pod}}",
              summary: "Reloading Prometheus' configuration failed",
            },
            expr: |||
              prometheus_config_last_reload_successful{%(prometheusSelector)s} == 0
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusNotificationQueueRunningFull',
            annotations: {
              description: "Prometheus' alert notification queue is running full for {{$labels.namespace}}/{{ $labels.pod}}",
              summary: "Prometheus' alert notification queue is running full",
            },
            expr: |||
              predict_linear(prometheus_notifications_queue_length{%(prometheusSelector)s}[5m], 60 * 30) > prometheus_notifications_queue_capacity{%(prometheusSelector)s}
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusErrorSendingAlerts',
            annotations: {
              description: 'Errors while sending alerts from Prometheus {{$labels.namespace}}/{{ $labels.pod}} to Alertmanager {{$labels.Alertmanager}}',
              summary: 'Errors while sending alert from Prometheus',
            },
            expr: |||
              rate(prometheus_notifications_errors_total{%(prometheusSelector)s}[5m]) / rate(prometheus_notifications_sent_total{%(prometheusSelector)s}[5m]) > 0.01
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusErrorSendingAlerts',
            annotations: {
              description: 'Errors while sending alerts from Prometheus {{$labels.namespace}}/{{ $labels.pod}} to Alertmanager {{$labels.Alertmanager}}',
              summary: 'Errors while sending alerts from Prometheus',
            },
            expr: |||
              rate(prometheus_notifications_errors_total{%(prometheusSelector)s}[5m]) / rate(prometheus_notifications_sent_total{%(prometheusSelector)s}[5m]) > 0.03
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'PrometheusNotConnectedToAlertmanagers',
            annotations: {
              description: 'Prometheus {{ $labels.namespace }}/{{ $labels.pod}} is not connected to any Alertmanagers',
              summary: 'Prometheus is not connected to any Alertmanagers',
            },
            expr: |||
              prometheus_notifications_alertmanagers_discovered{%(prometheusSelector)s} < 1
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusTSDBReloadsFailing',
            annotations: {
              description: '{{$labels.job}} at {{$labels.instance}} had {{$value | humanize}} reload failures over the last four hours.',
              summary: 'Prometheus has issues reloading data blocks from disk',
            },
            expr: |||
              increase(prometheus_tsdb_reloads_failures_total{%(prometheusSelector)s}[2h]) > 0
            ||| % $._config,
            'for': '12h',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusTSDBCompactionsFailing',
            annotations: {
              description: '{{$labels.job}} at {{$labels.instance}} had {{$value | humanize}} compaction failures over the last four hours.',
              summary: 'Prometheus has issues compacting sample blocks',
            },
            expr: |||
              increase(prometheus_tsdb_compactions_failed_total{%(prometheusSelector)s}[2h]) > 0
            ||| % $._config,
            'for': '12h',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusTSDBWALCorruptions',
            annotations: {
              description: '{{$labels.job}} at {{$labels.instance}} has a corrupted write-ahead log (WAL).',
              summary: 'Prometheus write-ahead log is corrupted',
            },
            expr: |||
              prometheus_tsdb_wal_corruptions_total{%(prometheusSelector)s} > 0
            ||| % $._config,
            'for': '4h',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusNotIngestingSamples',
            annotations: {
              description: "Prometheus {{ $labels.namespace }}/{{ $labels.pod}} isn't ingesting samples.",
              summary: "Prometheus isn't ingesting samples",
            },
            expr: |||
              rate(prometheus_tsdb_head_samples_appended_total{%(prometheusSelector)s}[5m]) <= 0
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PrometheusTargetScrapesDuplicate',
            annotations: {
              description: '{{$labels.namespace}}/{{$labels.pod}} has many samples rejected due to duplicate timestamps but different values',
              summary: 'Prometheus has many samples rejected',
            },
            expr: |||
              increase(prometheus_target_scrapes_sample_duplicate_timestamp_total{%(prometheusSelector)s}[5m]) > 0
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
        ],
      },
    ],
  },
}
