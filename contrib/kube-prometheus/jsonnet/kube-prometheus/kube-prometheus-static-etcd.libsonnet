local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

(import 'etcd-mixin/mixin.libsonnet') + {
  _config+:: {
    etcd: {
      ips: [],
      clientCA: null,
      clientKey: null,
      clientCert: null,
      serverName: null,
    },
  },
  prometheus+:: {
    serviceEtcd:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local etcdServicePort = servicePort.newNamed('metrics', 2379, 2379);

      service.new('etcd', null, etcdServicePort) +
      service.mixin.metadata.withNamespace('kube-system') +
      service.mixin.metadata.withLabels({ 'k8s-app': 'etcd' }) +
      service.mixin.spec.withClusterIp('None'),
    endpointsEtcd:
      local endpoints = k.core.v1.endpoints;
      local endpointSubset = endpoints.subsetsType;
      local endpointPort = endpointSubset.portsType;

      local etcdPort = endpointPort.new() +
                       endpointPort.withName('metrics') +
                       endpointPort.withPort(2379) +
                       endpointPort.withProtocol('TCP');

      local subset = endpointSubset.new() +
                     endpointSubset.withAddresses([
                       { ip: etcdIP }
                       for etcdIP in $._config.etcd.ips
                     ]) +
                     endpointSubset.withPorts(etcdPort);

      endpoints.new() +
      endpoints.mixin.metadata.withName('etcd') +
      endpoints.mixin.metadata.withNamespace('kube-system') +
      endpoints.mixin.metadata.withLabels({ 'k8s-app': 'etcd' }) +
      endpoints.withSubsets(subset),
    serviceMonitorEtcd:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'etcd',
          namespace: 'kube-system',
          labels: {
            'k8s-app': 'etcd',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'metrics',
              interval: '30s',
              scheme: 'https',
              tlsConfig: {
                caFile: '/etc/prometheus/secrets/kube-etcd-client-certs/etcd-client-ca.crt',
                keyFile: '/etc/prometheus/secrets/kube-etcd-client-certs/etcd-client.key',
                certFile: '/etc/prometheus/secrets/kube-etcd-client-certs/etcd-client.crt',
                serverName: $._config.etcd.serverName,
              },
            },
          ],
          selector: {
            matchLabels: {
              'k8s-app': 'etcd',
            },
          },
        },
      },
    secretEtcdCerts:
      local secret = k.core.v1.secret;

      secret.new('kube-etcd-client-certs', {
        'etcd-client-ca.crt': std.base64($._config.etcd.clientCA),
        'etcd-client.key': std.base64($._config.etcd.clientKey),
        'etcd-client.crt': std.base64($._config.etcd.clientCert),
      }) +
      secret.mixin.metadata.withNamespace($._config.namespace),
    prometheus+:
      {
        spec+: {
          secrets+: [$.prometheus.secretEtcdCerts.metadata.name],
        },
      },
  },
}
