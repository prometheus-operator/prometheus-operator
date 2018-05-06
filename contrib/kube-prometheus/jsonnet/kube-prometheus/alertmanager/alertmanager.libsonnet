local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

local alertmanagerConfig = "\nglobal:\n  resolve_timeout: 5m\nroute:\n  group_by: ['job']\n  group_wait: 30s\n  group_interval: 5m\n  repeat_interval: 12h\n  receiver: 'null'\n  routes:\n  - match:\n      alertname: DeadMansSwitch\n    receiver: 'null'\nreceivers:\n- name: 'null'\n";

{
  _config+:: {
    namespace: 'default',

    versions+:: {
      alertmanager: 'v0.14.0',
    },

    imageRepos+:: {
      alertmanager: 'quay.io/prometheus/alertmanager',
    },

    alertmanager+:: {
      config: alertmanagerConfig,
      replicas: 3,
    },
  },

  alertmanager+:: {
    secret:
      local secret = k.core.v1.secret;

      secret.new('alertmanager-main', { 'alertmanager.yaml': std.base64($._config.alertmanager.config) }) +
      secret.mixin.metadata.withNamespace($._config.namespace),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('alertmanager-main') +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local alertmanagerPort = servicePort.newNamed('web', 9093, 'web');

      service.new('alertmanager-main', { app: 'alertmanager', alertmanager: 'main' }, alertmanagerPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ alertmanager: 'main' }),

    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'alertmanager',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'alertmanager',
          },
        },
        spec: {
          selector: {
            matchLabels: {
              alertmanager: 'main',
            },
          },
          namespaceSelector: {
            matchNames: [
              'monitoring',
            ],
          },
          endpoints: [
            {
              port: 'web',
              interval: '30s',
            },
          ],
        },
      },

    alertmanager:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'Alertmanager',
        metadata: {
          name: 'main',
          namespace: $._config.namespace,
          labels: {
            alertmanager: 'main',
          },
        },
        spec: {
          replicas: $._config.alertmanager.replicas,
          version: $._config.versions.alertmanager,
          baseImage: $._config.imageRepos.alertmanager,
          nodeSelector: { 'beta.kubernetes.io/os': 'linux' },
          serviceAccountName: 'alertmanager-main',
        },
      },
  },
}
