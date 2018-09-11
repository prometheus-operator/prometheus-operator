{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'kube-prometheus-node-alerting.rules',
        rules: [
          {
            alert: 'NodeDiskRunningFull',
            annotations: {
              message: 'Device {{ $labels.device }} of node-exporter {{ $labels.namespace }}/{{ $labels.pod }} will be full within the next 24 hours.',
            },
            expr: |||
              (node:node_filesystem_usage: > 0.85) and (predict_linear(node:node_filesystem_avail:[6h], 3600 * 24) < 0)
            ||| % $._config,
            'for': '30m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'NodeDiskRunningFull',
            annotations: {
              message: 'Device {{ $labels.device }} of node-exporter {{ $labels.namespace }}/{{ $labels.pod }} will be full within the next 2 hours.',
            },
            expr: |||
              (node:node_filesystem_usage: > 0.85) and (predict_linear(node:node_filesystem_avail:[30m], 3600 * 2) < 0)
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'critical',
            },
          },
        ],
      },
    ],
  },
}
