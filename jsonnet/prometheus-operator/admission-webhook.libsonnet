local defaults = {
  local defaults = self,
  name: 'prometheus-operator-admission-webhook',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide admission webhook image',
  // The name of the Secret containing the TLS certificate and key of the admission webhook service.
  tlsSecretName: error 'must provide tlsSecretName',
  // The Secret's key containing the TLS certificate.
  tlsCertRef: 'tls.crt',
  // The Secret's key containing the TLS private key.
  tlsPrivateKeyRef: 'tls.key',
  port: 443,
  replicas: 2,
  resources: {
    limits: { cpu: '200m', memory: '200Mi' },
    requests: { cpu: '50m', memory: '50Mi' },
  },
  commonLabels:: {
    'app.kubernetes.io/name': defaults.name,
    'app.kubernetes.io/version': defaults.version,
  },
  selectorLabels:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local aw = self,
  _config:: defaults + params,
  _metadata:: {
    name: aw._config.name,
    namespace: aw._config.namespace,
    labels: aw._config.commonLabels,
  },

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: aw._metadata,
    automountServiceAccountToken: false,
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: aw._metadata,
    spec: {
      ports: [
        { name: 'https', targetPort: 'https', port: aw._config.port },
      ],
      selector: aw._config.selectorLabels,
    },
  },

  deployment:
    local container = {
      name: aw._config.name,
      image: aw._config.image,
      ports: [{
        containerPort: 8443,
        name: 'https',
      }],
      args: [
        '--web.enable-tls=true',
        '--web.cert-file=/etc/tls/private/tls.crt',
        '--web.key-file=/etc/tls/private/tls.key',
      ],
      resources: aw._config.resources,
      terminationMessagePolicy: 'FallbackToLogsOnError',
      securityContext: {
        allowPrivilegeEscalation: false,
        readOnlyRootFilesystem: true,
        capabilities: { drop: ['ALL'] },
      },
      volumeMounts: [
        {
          mountPath: '/etc/tls/private',
          name: 'tls-certificates',
          readOnly: true,
        },
      ],
    };
    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: aw._metadata,
      spec: {
        replicas: aw._config.replicas,
        selector: { matchLabels: aw._config.selectorLabels },
        template: {
          metadata: {
            labels: aw._config.commonLabels,
            annotations: {
              'kubectl.kubernetes.io/default-container': container.name,
            },
          },
          spec: {
            containers: [container],
            securityContext: {
              runAsNonRoot: true,
              runAsUser: 65534,
            },
            serviceAccountName: aw._config.name,
            automountServiceAccountToken: false,
            volumes: [{
              name: 'tls-certificates',
              secret: {
                secretName: aw._config.tlsSecretName,
                items: [{
                  key: aw._config.tlsCertRef,
                  path: 'tls.crt',
                }, {
                  key: aw._config.tlsPrivateKeyRef,
                  path: 'tls.key',
                }],
              },
            }],
          },
        },

      } + if aw._config.replicas > 1 then {
        // configure hard anti-affinity + rolling update for proper HA.
        template+: {
          spec+: {
            affinity: {
              podAntiAffinity: {
                requiredDuringSchedulingIgnoredDuringExecution: [{
                  namespaces: [aw._config.namespace],
                  topologyKey: 'kubernetes.io/hostname',
                  labelSelector: {
                    matchLabels: aw._config.selectorLabels,
                  },
                }],
              },
            },
          },
        },
        strategy: {
          rollingUpdate: {
            maxUnavailable: 1,
          },
        },
      },
    },

  serviceMonitor: {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata: aw._metadata,
    spec: {
      endpoints: [
        {
          port: 'https',
          honorLabels: true,
        },
      ],
      selector: {
        matchLabels: aw._config.commonLabels,
      },
    },
  },

  [if (defaults + params).replicas > 1 then 'podDisruptionBudget']: {
    apiVersion: 'policy/v1',
    kind: 'PodDisruptionBudget',
    metadata: aw._metadata,
    spec: {
      minAvailable: 1,
      selector: { matchLabels: aw._config.selectorLabels },

    },
  },
}
