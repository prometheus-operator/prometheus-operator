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

    alertmanager+:: {
      name: 'main',
    },

    prometheus+:: {
      name: 'k8s',
      replicas: 2,
      rules: {},
      renderedRules: {},
    },
  },

  prometheus+:: {
    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('prometheus-' + $._config.prometheus.name) +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),
    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local prometheusPort = servicePort.newNamed('web', 9090, 'web');

      service.new('prometheus-' + $._config.prometheus.name, { app: 'prometheus', prometheus: $._config.prometheus.name }, prometheusPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ prometheus: $._config.prometheus.name }),
    rules:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'PrometheusRule',
        metadata: {
          labels: {
            prometheus: $._config.prometheus.name,
            role: 'alert-rules',
          },
          name: 'prometheus-' + $._config.prometheus.name + '-rules',
          namespace: $._config.namespace,
        },
        spec: {
          groups: $._config.prometheus.rules.groups,
        },
      },
    roleBindingDefault:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.metadata.withNamespace('default') +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.prometheus.name, namespace: $._config.namespace }]),
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
      clusterRole.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
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
      role.mixin.metadata.withName('prometheus-' + $._config.prometheus.name + '-config') +
      role.mixin.metadata.withNamespace($._config.namespace) +
      role.withRules(configmapRule),
    roleBindingConfig:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-' + $._config.prometheus.name + '-config') +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-' + $._config.prometheus.name + '-config') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.prometheus.name, namespace: $._config.namespace }]),
    roleBindingNamespace:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.prometheus.name, namespace: $._config.namespace }]),
    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('prometheus-' + $._config.prometheus.name) +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.prometheus.name, namespace: $._config.namespace }]),
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
      role.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
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
      role.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
      role.mixin.metadata.withNamespace('default') +
      role.withRules(coreRule),
    roleBindingKubeSystem:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.metadata.withNamespace('kube-system') +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-' + $._config.prometheus.name) +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.prometheus.name, namespace: $._config.namespace }]),
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
      role.mixin.metadata.withName('prometheus-' + $._config.prometheus.name) +
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
          name: $._config.prometheus.name,
          namespace: $._config.namespace,
          labels: {
            prometheus: $._config.prometheus.name,
          },
        },
        spec: {
          replicas: $._config.prometheus.replicas,
          version: $._config.versions.prometheus,
          baseImage: $._config.imageRepos.prometheus,
          serviceAccountName: 'prometheus-' + $._config.prometheus.name,
          serviceMonitorSelector: selector.withMatchExpressions({ key: 'k8s-app', operator: 'Exists' }),
          nodeSelector: { 'beta.kubernetes.io/os': 'linux' },
          ruleSelector: selector.withMatchLabels({
            role: 'alert-rules',
            prometheus: $._config.prometheus.name,
          }),
          resources: resources,
          alerting: {
            alertmanagers: [
              {
                namespace: $._config.namespace,
                name: 'alertmanager-' + $._config.alertmanager.name,
                port: 'web',
              },
            ],
          },
        },
      },
    serviceMonitor:
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
              prometheus: $._config.prometheus.name,
            },
          },
          endpoints: [
            {
              port: 'web',
              interval: '30s',
            },
          ],
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
