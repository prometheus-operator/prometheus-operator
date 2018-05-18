{
  prometheus+:: {
    serviceMonitorKubelet+:
      {
        spec+: {
          endpoints: [
            {
              port: 'http-metrics',
              scheme: 'http',
              interval: '30s',
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
            {
              port: 'http-metrics',
              scheme: 'http',
              path: '/metrics/cadvisor',
              interval: '30s',
              honorLabels: true,
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
          ],
        },
      },
  },
}
