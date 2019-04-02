local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    versions+:: {
      prometheusAdapter: 'v0.4.1',
    },

    imageRepos+:: {
      prometheusAdapter: 'quay.io/coreos/k8s-prometheus-adapter-amd64',
    },

    prometheusAdapter+:: {
      name: 'prometheus-adapter',
      labels: { name: $._config.prometheusAdapter.name },
      prometheusURL: 'http://prometheus-' + $._config.prometheus.name + '.' + $._config.namespace + '.svc:9090/',
      config: |||
        resourceRules:
          cpu:
            containerQuery: sum(rate(container_cpu_usage_seconds_total{<<.LabelMatchers>>,container_name!="POD",container_name!="",pod_name!=""}[1m])) by (<<.GroupBy>>)
            nodeQuery: sum(1 - rate(node_cpu_seconds_total{mode="idle"}[1m]) * on(namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{<<.LabelMatchers>>}) by (<<.GroupBy>>)
            resources:
              overrides:
                node:
                  resource: node
                namespace:
                  resource: namespace
                pod_name:
                  resource: pod
            containerLabel: container_name
          memory:
            containerQuery: sum(container_memory_working_set_bytes{<<.LabelMatchers>>,container_name!="POD",container_name!="",pod_name!=""}) by (<<.GroupBy>>)
            nodeQuery: sum(node:node_memory_bytes_total:sum{<<.LabelMatchers>>} - node:node_memory_bytes_available:sum{<<.LabelMatchers>>}) by (<<.GroupBy>>)
            resources:
              overrides:
                node:
                  resource: node
                namespace:
                  resource: namespace
                pod_name:
                  resource: pod
            containerLabel: container_name
          window: 1m
      |||,
    },
  },

  prometheusAdapter+:: {
    apiService:
      {
        apiVersion: 'apiregistration.k8s.io/v1',
        kind: 'APIService',
        metadata: {
          name: 'v1beta1.metrics.k8s.io',
        },
        spec: {
          service: {
            name: $.prometheusAdapter.service.metadata.name,
            namespace: $._config.namespace,
          },
          group: 'metrics.k8s.io',
          version: 'v1beta1',
          insecureSkipTLSVerify: true,
          groupPriorityMinimum: 100,
          versionPriority: 100,
        },
      },

    configMap:
      local configmap = k.core.v1.configMap;

      configmap.new('adapter-config', { 'config.yaml': $._config.prometheusAdapter.config }) +
      configmap.mixin.metadata.withNamespace($._config.namespace),

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      service.new(
        $._config.prometheusAdapter.name,
        $._config.prometheusAdapter.labels,
        servicePort.newNamed('https', 443, 6443),
      ) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels($._config.prometheusAdapter.labels),

    deployment:
      local deployment = k.apps.v1beta2.deployment;
      local volume = deployment.mixin.spec.template.spec.volumesType;
      local container = deployment.mixin.spec.template.spec.containersType;
      local containerVolumeMount = container.volumeMountsType;

      local c =
        container.new($._config.prometheusAdapter.name, $._config.imageRepos.prometheusAdapter + ':' + $._config.versions.prometheusAdapter) +
        container.withArgs([
          '--cert-dir=/var/run/serving-cert',
          '--config=/etc/adapter/config.yaml',
          '--logtostderr=true',
          '--metrics-relist-interval=1m',
          '--prometheus-url=' + $._config.prometheusAdapter.prometheusURL,
          '--secure-port=6443',
        ]) +
        container.withPorts([{ containerPort: 6443 }]) +
        container.withVolumeMounts([
          containerVolumeMount.new('tmpfs', '/tmp'),
          containerVolumeMount.new('volume-serving-cert', '/var/run/serving-cert'),
          containerVolumeMount.new('config', '/etc/adapter'),
        ],);

      deployment.new($._config.prometheusAdapter.name, 1, c, $._config.prometheusAdapter.labels) +
      deployment.mixin.metadata.withNamespace($._config.namespace) +
      deployment.mixin.spec.selector.withMatchLabels($._config.prometheusAdapter.labels) +
      deployment.mixin.spec.template.spec.withServiceAccountName($.prometheusAdapter.serviceAccount.metadata.name) +
      deployment.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(1) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(0) +
      deployment.mixin.spec.template.spec.withVolumes([
        volume.fromEmptyDir(name='tmpfs'),
        volume.fromEmptyDir(name='volume-serving-cert'),
        { name: 'config', configMap: { name: 'adapter-config' } },
      ]),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new($._config.prometheusAdapter.name) +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local rules =
        policyRule.new() +
        policyRule.withApiGroups(['']) +
        policyRule.withResources(['nodes', 'namespaces', 'pods', 'services']) +
        policyRule.withVerbs(['get', 'list', 'watch']);

      clusterRole.new() +
      clusterRole.mixin.metadata.withName($._config.prometheusAdapter.name) +
      clusterRole.withRules(rules),

    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName($._config.prometheusAdapter.name) +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName($.prometheusAdapter.clusterRole.metadata.name) +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{
        kind: 'ServiceAccount',
        name: $.prometheusAdapter.serviceAccount.metadata.name,
        namespace: $._config.namespace,
      }]),

    clusterRoleBindingDelegator:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('resource-metrics:system:auth-delegator') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('system:auth-delegator') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{
        kind: 'ServiceAccount',
        name: $.prometheusAdapter.serviceAccount.metadata.name,
        namespace: $._config.namespace,
      }]),

    clusterRoleServerResources:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local rules =
        policyRule.new() +
        policyRule.withApiGroups(['metrics.k8s.io']) +
        policyRule.withResources(['*']) +
        policyRule.withVerbs(['*']);

      clusterRole.new() +
      clusterRole.mixin.metadata.withName('resource-metrics-server-resources') +
      clusterRole.withRules(rules),

    roleBindingAuthReader:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('resource-metrics-auth-reader') +
      roleBinding.mixin.metadata.withNamespace('kube-system') +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('extension-apiserver-authentication-reader') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{
        kind: 'ServiceAccount',
        name: $.prometheusAdapter.serviceAccount.metadata.name,
        namespace: $._config.namespace,
      }]),
  },
}
