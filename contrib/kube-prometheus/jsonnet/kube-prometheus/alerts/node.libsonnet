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
      {
        name: 'node-time',
        rules: [
          {
            alert: 'ClockSkewDetected',
            annotations: {
              message: 'Clock skew detected on node-exporter {{ $labels.namespace }}/{{ $labels.pod }}. Ensure NTP is configured correctly on this host.',
            },
            expr: |||
              node_ntp_offset_seconds{%(nodeExporterSelector)s} < -0.03 or node_ntp_offset_seconds{%(nodeExporterSelector)s} > 0.03
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
        ],
      },
      {
        name: 'node-network',
        rules: [
          {
            alert: 'NetworkReceiveErrors',
            annotations: {
              message: 'Network interface "{{ $labels.device }}" showing receive errors on node-exporter {{ $labels.namespace }}/{{ $labels.pod }}"',
            },
            expr: |||
              rate(node_network_receive_errs_total{%(nodeExporterSelector)s,%(hostNetworkInterfaceSelector)s}[2m]) > 0
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'NetworkTransmitErrors',
            annotations: {
              message: 'Network interface "{{ $labels.device }}" showing transmit errors on node-exporter {{ $labels.namespace }}/{{ $labels.pod }}"',
            },
            expr: |||
              rate(node_network_transmit_errs_total{%(nodeExporterSelector)s,%(hostNetworkInterfaceSelector)s}[2m]) > 0
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'NodeNetworkInterfaceDown',
            annotations: {
              message: 'Network interface "{{ $labels.device }}" down on node-exporter {{ $labels.namespace }}/{{ $labels.pod }}"',
            },
            expr: |||
              node_network_up{%(nodeExporterSelector)s,%(hostNetworkInterfaceSelector)s} == 0
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'NodeNetworkInterfaceFlapping',
            annotations: {
              message: 'Network interface "{{ $labels.device }}" changing it\'s up status often on node-exporter {{ $labels.namespace }}/{{ $labels.pod }}"',
            },
            expr: |||
              changes(node_network_up{%(nodeExporterSelector)s,%(hostNetworkInterfaceSelector)s}[2m]) > 2
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
        ],
      },
    ],
  },
}
