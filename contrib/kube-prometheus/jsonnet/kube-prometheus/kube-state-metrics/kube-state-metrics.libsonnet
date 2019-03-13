local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    kubeStateMetrics+:: {
      collectors: '',  // empty string gets a default set
      scrapeInterval: '30s',
      scrapeTimeout: '30s',

      baseCPU: '100m',
      baseMemory: '150Mi',
      cpuPerNode: '2m',
      memoryPerNode: '30Mi',
    },

    versions+:: {
      kubeStateMetrics: 'v1.5.0',
      kubeRbacProxy: 'v0.4.1',
      addonResizer: '1.8.4',
    },

    imageRepos+:: {
      kubeStateMetrics: 'quay.io/coreos/kube-state-metrics',
      kubeRbacProxy: 'quay.io/coreos/kube-rbac-proxy',
      addonResizer: 'k8s.gcr.io/addon-resizer',
    },
  },

  kubeStateMetrics+:: {
    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('kube-state-metrics') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('kube-state-metrics') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'kube-state-metrics', namespace: $._config.namespace }]),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local rulesType = clusterRole.rulesType;

      local coreRule = rulesType.new() +
                       rulesType.withApiGroups(['']) +
                       rulesType.withResources([
                         'configmaps',
                         'secrets',
                         'nodes',
                         'pods',
                         'services',
                         'resourcequotas',
                         'replicationcontrollers',
                         'limitranges',
                         'persistentvolumeclaims',
                         'persistentvolumes',
                         'namespaces',
                         'endpoints',
                       ]) +
                       rulesType.withVerbs(['list', 'watch']);

      local extensionsRule = rulesType.new() +
                             rulesType.withApiGroups(['extensions']) +
                             rulesType.withResources([
                               'daemonsets',
                               'deployments',
                               'replicasets',
                             ]) +
                             rulesType.withVerbs(['list', 'watch']);

      local appsRule = rulesType.new() +
                       rulesType.withApiGroups(['apps']) +
                       rulesType.withResources([
                         'statefulsets',
                         'daemonsets',
                         'deployments',
                         'replicasets',
                       ]) +
                       rulesType.withVerbs(['list', 'watch']);

      local batchRule = rulesType.new() +
                        rulesType.withApiGroups(['batch']) +
                        rulesType.withResources([
                          'cronjobs',
                          'jobs',
                        ]) +
                        rulesType.withVerbs(['list', 'watch']);

      local autoscalingRule = rulesType.new() +
                              rulesType.withApiGroups(['autoscaling']) +
                              rulesType.withResources([
                                'horizontalpodautoscalers',
                              ]) +
                              rulesType.withVerbs(['list', 'watch']);

      local authenticationRole = rulesType.new() +
                                 rulesType.withApiGroups(['authentication.k8s.io']) +
                                 rulesType.withResources([
                                   'tokenreviews',
                                 ]) +
                                 rulesType.withVerbs(['create']);

      local authorizationRole = rulesType.new() +
                                rulesType.withApiGroups(['authorization.k8s.io']) +
                                rulesType.withResources([
                                  'subjectaccessreviews',
                                ]) +
                                rulesType.withVerbs(['create']);

      local policyRule = rulesType.new() +
                         rulesType.withApiGroups(['policy']) +
                         rulesType.withResources([
                           'poddisruptionbudgets',
                         ]) +
                         rulesType.withVerbs(['list', 'watch']);

      local rules = [coreRule, extensionsRule, appsRule, batchRule, autoscalingRule, authenticationRole, authorizationRole, policyRule];

      clusterRole.new() +
      clusterRole.mixin.metadata.withName('kube-state-metrics') +
      clusterRole.withRules(rules),
    deployment:
      local deployment = k.apps.v1beta2.deployment;
      local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
      local volume = k.apps.v1beta2.deployment.mixin.spec.template.spec.volumesType;
      local containerPort = container.portsType;
      local containerVolumeMount = container.volumeMountsType;
      local podSelector = deployment.mixin.spec.template.spec.selectorType;

      local podLabels = { app: 'kube-state-metrics' };

      local proxyClusterMetrics =
        container.new('kube-rbac-proxy-main', $._config.imageRepos.kubeRbacProxy + ':' + $._config.versions.kubeRbacProxy) +
        container.withArgs([
          '--logtostderr',
          '--secure-listen-address=:8443',
          '--tls-cipher-suites=' + std.join(',', $._config.tlsCipherSuites),
          '--upstream=http://127.0.0.1:8081/',
        ]) +
        container.withPorts(containerPort.newNamed('https-main', 8443)) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '20Mi' }) +
        container.mixin.resources.withLimits({ cpu: '20m', memory: '40Mi' });

      local proxySelfMetrics =
        container.new('kube-rbac-proxy-self', $._config.imageRepos.kubeRbacProxy + ':' + $._config.versions.kubeRbacProxy) +
        container.withArgs([
          '--logtostderr',
          '--secure-listen-address=:9443',
          '--tls-cipher-suites=' + std.join(',', $._config.tlsCipherSuites),
          '--upstream=http://127.0.0.1:8082/',
        ]) +
        container.withPorts(containerPort.newNamed('https-self', 9443)) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '20Mi' }) +
        container.mixin.resources.withLimits({ cpu: '20m', memory: '40Mi' });

      local kubeStateMetrics =
        container.new('kube-state-metrics', $._config.imageRepos.kubeStateMetrics + ':' + $._config.versions.kubeStateMetrics) +
        container.withArgs([
          '--host=127.0.0.1',
          '--port=8081',
          '--telemetry-host=127.0.0.1',
          '--telemetry-port=8082',
        ] + if $._config.kubeStateMetrics.collectors != '' then ['--collectors=' + $._config.kubeStateMetrics.collectors] else []) +
        container.mixin.resources.withRequests({ cpu: $._config.kubeStateMetrics.baseCPU, memory: $._config.kubeStateMetrics.baseMemory }) +
        container.mixin.resources.withLimits({ cpu: $._config.kubeStateMetrics.baseCPU, memory: $._config.kubeStateMetrics.baseMemory });

      local addonResizer =
        container.new('addon-resizer', $._config.imageRepos.addonResizer + ':' + $._config.versions.addonResizer) +
        container.withCommand([
          '/pod_nanny',
          '--container=kube-state-metrics',
          '--cpu=' + $._config.kubeStateMetrics.baseCPU,
          '--extra-cpu=' + $._config.kubeStateMetrics.cpuPerNode,
          '--memory=' + $._config.kubeStateMetrics.baseMemory,
          '--extra-memory=' + $._config.kubeStateMetrics.memoryPerNode,
          '--threshold=5',
          '--deployment=kube-state-metrics',
        ]) +
        container.withEnv([
          {
            name: 'MY_POD_NAME',
            valueFrom: {
              fieldRef: { apiVersion: 'v1', fieldPath: 'metadata.name' },
            },
          },
          {
            name: 'MY_POD_NAMESPACE',
            valueFrom: {
              fieldRef: { apiVersion: 'v1', fieldPath: 'metadata.namespace' },
            },
          },
        ]) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '30Mi' }) +
        container.mixin.resources.withLimits({ cpu: '50m', memory: '30Mi' });

      local c = [proxyClusterMetrics, proxySelfMetrics, kubeStateMetrics, addonResizer];

      deployment.new('kube-state-metrics', 1, c, podLabels) +
      deployment.mixin.metadata.withNamespace($._config.namespace) +
      deployment.mixin.metadata.withLabels(podLabels) +
      deployment.mixin.spec.selector.withMatchLabels(podLabels) +
      deployment.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.withServiceAccountName('kube-state-metrics'),

    roleBinding:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('kube-state-metrics') +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('kube-state-metrics') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'kube-state-metrics' }]),

    role:
      local role = k.rbac.v1.role;
      local rulesType = role.rulesType;

      local coreRule = rulesType.new() +
                       rulesType.withApiGroups(['']) +
                       rulesType.withResources([
                         'pods',
                       ]) +
                       rulesType.withVerbs(['get']);

      local extensionsRule = rulesType.new() +
                             rulesType.withApiGroups(['extensions']) +
                             rulesType.withResources([
                               'deployments',
                             ]) +
                             rulesType.withVerbs(['get', 'update']) +
                             rulesType.withResourceNames(['kube-state-metrics']);

      local appsRule = rulesType.new() +
                       rulesType.withApiGroups(['apps']) +
                       rulesType.withResources([
                         'deployments',
                       ]) +
                       rulesType.withVerbs(['get', 'update']) +
                       rulesType.withResourceNames(['kube-state-metrics']);

      local rules = [coreRule, extensionsRule, appsRule];

      role.new() +
      role.mixin.metadata.withName('kube-state-metrics') +
      role.mixin.metadata.withNamespace($._config.namespace) +
      role.withRules(rules),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('kube-state-metrics') +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local ksmServicePortMain = servicePort.newNamed('https-main', 8443, 'https-main');
      local ksmServicePortSelf = servicePort.newNamed('https-self', 9443, 'https-self');

      service.new('kube-state-metrics', $.kubeStateMetrics.deployment.spec.selector.matchLabels, [ksmServicePortMain, ksmServicePortSelf]) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ 'k8s-app': 'kube-state-metrics' }) +
      service.mixin.spec.withClusterIp('None'),

    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kube-state-metrics',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'kube-state-metrics',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          selector: {
            matchLabels: {
              'k8s-app': 'kube-state-metrics',
            },
          },
          endpoints: [
            {
              port: 'https-main',
              scheme: 'https',
              interval: $._config.kubeStateMetrics.scrapeInterval,
              scrapeTimeout: $._config.kubeStateMetrics.scrapeTimeout,
              honorLabels: true,
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
              tlsConfig: {
                insecureSkipVerify: true,
              },
            },
            {
              port: 'https-self',
              scheme: 'https',
              interval: '30s',
              bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
              tlsConfig: {
                insecureSkipVerify: true,
              },
            },
          ],
        },
      },
  },
}
