{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'alertmanager.rules',
        rules: [
          {
            alert: 'AlertmanagerConfigInconsistent',
            annotations: {
              description: 'The configuration of the instances of the Alertmanager cluster `{{$labels.service}}` are out of sync.',
              runbook_url: '',
              summary: 'Configuration out of sync',
            },
            expr: |||
              count_values("config_hash", alertmanager_config_hash{%(alertmanagerSelector)s}) BY (service) / ON(service) GROUP_LEFT() label_replace(prometheus_operator_alertmanager_spec_replicas{%(prometheusOperatorSelector)s}, "service", "alertmanager-$1", "alertmanager", "(.*)") != 1
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'AlertmanagerDownOrMissing',
            annotations: {
              description: 'An unexpected number of Alertmanagers are scraped or Alertmanagers disappeared from discovery.',
              runbook_url: '',
              summary: 'Alertmanager down or missing',
            },
            expr: |||
              label_replace(prometheus_operator_alertmanager_spec_replicas{%(prometheusOperatorSelector)s}, "job", "alertmanager-$1", "alertmanager", "(.*)") / ON(job) GROUP_RIGHT() sum(up{%(alertmanagerSelector)s}) BY (job) != 1
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'AlertmanagerFailedReload',
            annotations: {
              description: "Reloading Alertmanager's configuration has failed for {{ $labels.namespace }}/{{ $labels.pod}}.",
              runbook_url: '',
              summary: "Alertmanager's configuration reload failed",
            },
            expr: |||
              alertmanager_config_last_reload_successful{%(alertmanagerSelector)s} == 0
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
