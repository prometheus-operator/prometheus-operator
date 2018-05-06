local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    versions+:: {
      kubeStateMetrics: 'v1.3.0',
      kubeRbacProxy: 'v0.3.0',
      addonResizer: '1.0',
    },

    imageRepos+:: {
      kubeStateMetrics: 'quay.io/coreos/kube-state-metrics',
      kubeRbacProxy: 'quay.io/coreos/kube-rbac-proxy',
      addonResizer: 'quay.io/coreos/addon-resizer',
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
      local policyRule = clusterRole.rulesType;

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
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
                       policyRule.withVerbs(['list', 'watch']);

      local extensionsRule = policyRule.new() +
                             policyRule.withApiGroups(['extensions']) +
                             policyRule.withResources([
                               'daemonsets',
                               'deployments',
                               'replicasets',
                             ]) +
                             policyRule.withVerbs(['list', 'watch']);

      local appsRule = policyRule.new() +
                       policyRule.withApiGroups(['apps']) +
                       policyRule.withResources([
                         'statefulsets',
                       ]) +
                       policyRule.withVerbs(['list', 'watch']);

      local batchRule = policyRule.new() +
                        policyRule.withApiGroups(['batch']) +
                        policyRule.withResources([
                          'cronjobs',
                          'jobs',
                        ]) +
                        policyRule.withVerbs(['list', 'watch']);

      local autoscalingRule = policyRule.new() +
                              policyRule.withApiGroups(['autoscaling']) +
                              policyRule.withResources([
                                'horizontalpodautoscalers',
                              ]) +
                              policyRule.withVerbs(['list', 'watch']);

      local authenticationRole = policyRule.new() +
                                 policyRule.withApiGroups(['authentication.k8s.io']) +
                                 policyRule.withResources([
                                   'tokenreviews',
                                 ]) +
                                 policyRule.withVerbs(['create']);

      local authorizationRole = policyRule.new() +
                                policyRule.withApiGroups(['authorization.k8s.io']) +
                                policyRule.withResources([
                                  'subjectaccessreviews',
                                ]) +
                                policyRule.withVerbs(['create']);

      local rules = [coreRule, extensionsRule, appsRule, batchRule, autoscalingRule, authenticationRole, authorizationRole];

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
          '--secure-listen-address=:8443',
          '--upstream=http://127.0.0.1:8081/',
        ]) +
        container.withPorts(containerPort.newNamed('https-main', 8443)) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '20Mi' }) +
        container.mixin.resources.withLimits({ cpu: '20m', memory: '40Mi' });

      local proxySelfMetrics =
        container.new('kube-rbac-proxy-self', $._config.imageRepos.kubeRbacProxy + ':' + $._config.versions.kubeRbacProxy) +
        container.withArgs([
          '--secure-listen-address=:9443',
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
        ]) +
        container.mixin.resources.withRequests({ cpu: '102m', memory: '180Mi' }) +
        container.mixin.resources.withLimits({ cpu: '102m', memory: '180Mi' });

      local addonResizer =
        container.new('addon-resizer', $._config.imageRepos.addonResizer + ':' + $._config.versions.addonResizer) +
        container.withCommand([
          '/pod_nanny',
          '--container=kube-state-metrics',
          '--cpu=100m',
          '--extra-cpu=2m',
          '--memory=150Mi',
          '--extra-memory=30Mi',
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
        container.mixin.resources.withLimits({ cpu: '10m', memory: '30Mi' });

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
      local policyRule = role.rulesType;

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'pods',
                       ]) +
                       policyRule.withVerbs(['get']);

      local extensionsRule = policyRule.new() +
                             policyRule.withApiGroups(['extensions']) +
                             policyRule.withResources([
                               'deployments',
                             ]) +
                             policyRule.withVerbs(['get', 'update']) +
                             policyRule.withResourceNames(['kube-state-metrics']);

      local rules = [coreRule, extensionsRule];

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
          namespaceSelector: {
            matchNames: [
              'monitoring',
            ],
          },
          endpoints: [
            {
              port: 'https-main',
              scheme: 'https',
              interval: '30s',
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
