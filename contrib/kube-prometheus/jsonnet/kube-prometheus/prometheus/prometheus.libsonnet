local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    versions+:: {
      prometheus: 'v2.2.1',
    },

    imageRepos+:: {
      prometheus: 'quay.io/prometheus/prometheus',
    },

    prometheus+:: {
      replicas: 2,
      rules: {},
      renderedRules: {},
    },
  },

  prometheus+:: {
    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('prometheus-k8s') +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),
    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local prometheusPort = servicePort.newNamed('web', 9090, 'web');

      service.new('prometheus-k8s', { app: 'prometheus', prometheus: 'k8s' }, prometheusPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ prometheus: 'k8s' }),
    rules:
      local configMap = k.core.v1.configMap;

      configMap.new('prometheus-k8s-rules', ({ 'all.rules.yaml': std.manifestYamlDoc($._config.prometheus.rules) } + $._config.prometheus.renderedRules)) +
      configMap.mixin.metadata.withLabels({ role: 'alert-rules', prometheus: 'k8s' }) +
      configMap.mixin.metadata.withNamespace($._config.namespace),
    roleBindingDefault:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-k8s') +
      roleBinding.mixin.metadata.withNamespace('default') +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-k8s') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-k8s', namespace: $._config.namespace }]),
    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local nodeMetricsRule = policyRule.new() +
                              policyRule.withApiGroups(['']) +
                              policyRule.withResources(['nodes/metrics']) +
                              policyRule.withVerbs(['get']);

      local metricsRule = policyRule.new() +
                          policyRule.withNonResourceUrls('/metrics') +
                          policyRule.withVerbs(['get']);

      local rules = [nodeMetricsRule, metricsRule];

      clusterRole.new() +
      clusterRole.mixin.metadata.withName('prometheus-k8s') +
      clusterRole.withRules(rules),
    roleConfig:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local configmapRule = policyRule.new() +
                            policyRule.withApiGroups(['']) +
                            policyRule.withResources([
                              'configmaps',
                            ]) +
                            policyRule.withVerbs(['get']);

      role.new() +
      role.mixin.metadata.withName('prometheus-k8s-config') +
      role.mixin.metadata.withNamespace($._config.namespace) +
      role.withRules(configmapRule),
    roleBindingConfig:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-k8s-config') +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-k8s-config') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-k8s', namespace: $._config.namespace }]),
    roleBindingNamespace:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-k8s') +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-k8s') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-k8s', namespace: $._config.namespace }]),
    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('prometheus-k8s') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('prometheus-k8s') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-k8s', namespace: $._config.namespace }]),
    roleKubeSystem:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'nodes',
                         'services',
                         'endpoints',
                         'pods',
                       ]) +
                       policyRule.withVerbs(['get', 'list', 'watch']);

      role.new() +
      role.mixin.metadata.withName('prometheus-k8s') +
      role.mixin.metadata.withNamespace('kube-system') +
      role.withRules(coreRule),
    roleDefault:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'nodes',
                         'services',
                         'endpoints',
                         'pods',
                       ]) +
                       policyRule.withVerbs(['get', 'list', 'watch']);

      role.new() +
      role.mixin.metadata.withName('prometheus-k8s') +
      role.mixin.metadata.withNamespace('default') +
      role.withRules(coreRule),
    roleBindingKubeSystem:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-k8s') +
      roleBinding.mixin.metadata.withNamespace('kube-system') +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-k8s') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-k8s', namespace: $._config.namespace }]),
    roleNamespace:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'nodes',
                         'services',
                         'endpoints',
                         'pods',
                       ]) +
                       policyRule.withVerbs(['get', 'list', 'watch']);

      role.new() +
      role.mixin.metadata.withName('prometheus-k8s') +
      role.mixin.metadata.withNamespace($._config.namespace) +
      role.withRules(coreRule),
    prometheus:
      local container = k.core.v1.pod.mixin.spec.containersType;
      local resourceRequirements = container.mixin.resourcesType;
      local selector = k.apps.v1beta2.deployment.mixin.spec.selectorType;

      local resources = resourceRequirements.new() +
                        resourceRequirements.withRequests({ memory: '400Mi' });

      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'Prometheus',
        metadata: {
          name: 'k8s',
          namespace: $._config.namespace,
          labels: {
            prometheus: 'k8s',
          },
        },
        spec: {
          replicas: $._config.prometheus.replicas,
          version: $._config.versions.prometheus,
          baseImage: $._config.imageRepos.prometheus,
          serviceAccountName: 'prometheus-k8s',
          serviceMonitorSelector: selector.withMatchExpressions({ key: 'k8s-app', operator: 'Exists' }),
          nodeSelector: { 'beta.kubernetes.io/os': 'linux' },
          ruleSelector: selector.withMatchLabels({
            role: 'alert-rules',
            prometheus: 'k8s',
          }),
          resources: resources,
          alerting: {
            alertmanagers: [
              {
                namespace: $._config.namespace,
                name: 'alertmanager-main',
                port: 'web',
              },
            ],
          },
        },
      },
    serviceMonitorPrometheus:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'prometheus',
          },
        },
        spec: {
          selector: {
            matchLabels: {
              prometheus: 'k8s',
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
    serviceMonitorPrometheusOperator:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus-operator',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'prometheus-operator',
          },
        },
        spec: {
          endpoints: [
            {
              port: 'http',
            },
          ],
          selector: {
            matchLabels: {
              'k8s-app': 'prometheus-operator',
            },
          },
        },
      },
    serviceMonitorKubeScheduler:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kube-scheduler',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'kube-scheduler',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'http-metrics',
              interval: '30s',
            },
          ],
          selector: {
            matchLabels: {
              'k8s-app': 'kube-scheduler',
            },
          },
          namespaceSelector: {
            matchNames: [
              'kube-system',
            ],
          },
        },
      },
    serviceMonitorKubelet:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kubelet',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'kubelet',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'https-metrics',
              scheme: 'https',
              interval: '30s',
              tlsConfig: {
                insecureSkipVerify: true,
              },
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
            {
              port: 'https-metrics',
              scheme: 'https',
              path: '/metrics/cadvisor',
              interval: '30s',
              honorLabels: true,
              tlsConfig: {
                insecureSkipVerify: true,
              },
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
          ],
          selector: {
            matchLabels: {
              'k8s-app': 'kubelet',
            },
          },
          namespaceSelector: {
            matchNames: [
              'kube-system',
            ],
          },
        },
      },
    serviceMonitorKubeControllerManager:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kube-controller-manager',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'kube-controller-manager',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'http-metrics',
              interval: '30s',
            },
          ],
          selector: {
            matchLabels: {
              'k8s-app': 'kube-controller-manager',
            },
          },
          namespaceSelector: {
            matchNames: [
              'kube-system',
            ],
          },
        },
      },
    serviceMonitorApiserver:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kube-apiserver',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'apiserver',
          },
        },
        spec: {
          jobLabel: 'component',
          selector: {
            matchLabels: {
              component: 'apiserver',
              provider: 'kubernetes',
            },
          },
          namespaceSelector: {
            matchNames: [
              'default',
            ],
          },
          endpoints: [
            {
              port: 'https',
              interval: '30s',
              scheme: 'https',
              tlsConfig: {
                caFile: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
                serverName: 'kubernetes',
              },
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
          ],
        },
      },
    serviceMonitorCoreDNS:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'coredns',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'coredns',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          selector: {
            matchLabels: {
              'k8s-app': 'coredns',
              component: 'metrics',
            },
          },
          namespaceSelector: {
            matchNames: [
              'kube-system',
            ],
          },
          endpoints: [
            {
              port: 'http-metrics',
              interval: '15s',
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            },
          ],
        },
      },
  },
}
