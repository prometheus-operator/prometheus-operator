{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'general.rules',
        rules: [
          {
            alert: 'TargetDown',
            annotations: {
              description: '{{ $value }}% of {{ $labels.job }} targets are down.',
              runbook_url: '',
              summary: 'Targets are down',
            },
            expr: '100 * (count(up == 0) BY (job) / count(up) BY (job)) > 10',
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'DeadMansSwitch',
            annotations: {
              description: 'This is a DeadMansSwitch meant to ensure that the entire Alerting pipeline is functional.',
              runbook_url: '',
              summary: 'Alerting DeadMansSwitch',
            },
            expr: 'vector(1)',
            labels: {
              severity: 'none',
            },
          },
        ],
      },
    ],
  },
}
