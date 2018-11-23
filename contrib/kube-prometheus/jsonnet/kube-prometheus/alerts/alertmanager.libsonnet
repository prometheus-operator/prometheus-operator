{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'alertmanager.rules',
        rules: [
          {
            alert: 'AlertmanagerConfigInconsistent',
            annotations: {
              message: 'The configuration of the instances of the Alertmanager cluster `{{$labels.service}}` are out of sync.',
            },
            expr: |||
              count_values("config_hash", alertmanager_config_hash{%(alertmanagerSelector)s}) BY (service) / ON(service) GROUP_LEFT() label_replace(prometheus_operator_spec_replicas{%(prometheusOperatorSelector)s,controller="alertmanager"}, "service", "alertmanager-$1", "name", "(.*)") != 1
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'AlertmanagerFailedReload',
            annotations: {
              message: "Reloading Alertmanager's configuration has failed for {{ $labels.namespace }}/{{ $labels.pod}}.",
            },
            expr: |||
              alertmanager_config_last_reload_successful{%(alertmanagerSelector)s} == 0
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert:'AlertmanagerMembersInconsistent',
            annotations:{
              message: 'Alertmanager has not found all other members of the cluster.',
            },
            expr: |||
              alertmanager_cluster_members{%(alertmanagerSelector)s}
                != on (service) GROUP_LEFT()
              count by (service) (alertmanager_cluster_members{%(alertmanagerSelector)s})
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
          },
        ],
      },
    ],
  },
}
