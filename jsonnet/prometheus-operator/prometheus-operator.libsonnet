local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    prometheusOperator+:: {
      deploymentSelectorLabels: {
        'app.kubernetes.io/name': 'prometheus-operator',
        'app.kubernetes.io/component': 'controller',
      },
      commonLabels:
        $._config.prometheusOperator.deploymentSelectorLabels +
        { 'app.kubernetes.io/version': $._config.versions.prometheusOperator, },
    },

    versions+:: {
      prometheusOperator: 'v0.32.0',
      prometheusConfigReloader: self.prometheusOperator,
      configmapReloader: 'v0.0.1',
    },

    imageRepos+:: {
      prometheusOperator: 'quay.io/coreos/prometheus-operator',
      configmapReloader: 'quay.io/coreos/configmap-reload',
      prometheusConfigReloader: 'quay.io/coreos/prometheus-config-reloader',
    },
  },

  prometheusOperator+:: {
    // Prefixing with 0 to ensure these manifests are listed and therefore created first.
    '0alertmanagerCustomResourceDefinition': import 'alertmanager-crd.libsonnet',
    '0prometheusCustomResourceDefinition': import 'prometheus-crd.libsonnet',
    '0servicemonitorCustomResourceDefinition': import 'servicemonitor-crd.libsonnet',
    '0podmonitorCustomResourceDefinition': import 'podmonitor-crd.libsonnet',
    '0prometheusruleCustomResourceDefinition': import 'prometheusrule-crd.libsonnet',

    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withLabels($._config.prometheusOperator.commonLabels) +
      clusterRoleBinding.mixin.metadata.withName('prometheus-operator') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('prometheus-operator') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-operator', namespace: $._config.namespace }]),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local apiExtensionsRule = policyRule.new() +
                                policyRule.withApiGroups(['apiextensions.k8s.io']) +
                                policyRule.withResources([
                                  'customresourcedefinitions',
                                ]) +
                                policyRule.withVerbs(['*']);

      local monitoringRule = policyRule.new() +
                             policyRule.withApiGroups(['monitoring.coreos.com']) +
                             policyRule.withResources([
                               'alertmanagers',
                               'prometheuses',
                               'prometheuses/finalizers',
                               'alertmanagers/finalizers',
                               'servicemonitors',
                               'podmonitors',
                               'prometheusrules',
                             ]) +
                             policyRule.withVerbs(['*']);

      local appsRule = policyRule.new() +
                       policyRule.withApiGroups(['apps']) +
                       policyRule.withResources([
                         'statefulsets',
                       ]) +
                       policyRule.withVerbs(['*']);

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'configmaps',
                         'secrets',
                       ]) +
                       policyRule.withVerbs(['*']);

      local podRule = policyRule.new() +
                      policyRule.withApiGroups(['']) +
                      policyRule.withResources([
                        'pods',
                      ]) +
                      policyRule.withVerbs(['list', 'delete']);

      local routingRule = policyRule.new() +
                          policyRule.withApiGroups(['']) +
                          policyRule.withResources([
                            'services',
                            'services/finalizers',
                            'endpoints',
                          ]) +
                          policyRule.withVerbs(['get', 'create', 'update', 'delete']);

      local nodeRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'nodes',
                       ]) +
                       policyRule.withVerbs(['list', 'watch']);

      local namespaceRule = policyRule.new() +
                            policyRule.withApiGroups(['']) +
                            policyRule.withResources([
                              'namespaces',
                            ]) +
                            policyRule.withVerbs(['get', 'list', 'watch']);

      local rules = [apiExtensionsRule, monitoringRule, appsRule, coreRule, podRule, routingRule, nodeRule, namespaceRule];

      clusterRole.new() +
      clusterRole.mixin.metadata.withLabels($._config.prometheusOperator.commonLabels) +
      clusterRole.mixin.metadata.withName('prometheus-operator') +
      clusterRole.withRules(rules),

    deployment:
      local deployment = k.apps.v1.deployment;
      local container = k.apps.v1.deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;

      local targetPort = 8080;

      local operatorContainer =
        container.new('prometheus-operator', $._config.imageRepos.prometheusOperator + ':' + $._config.versions.prometheusOperator) +
        container.withPorts(containerPort.newNamed(targetPort, 'http')) +
        container.withArgs([
          '--kubelet-service=kube-system/kubelet',
          // Prometheus Operator is run with a read-only root file system. By
          // default glog saves logfiles to /tmp. Make it log to stderr instead.
          '--logtostderr=true',
          '--config-reloader-image=' + $._config.imageRepos.configmapReloader + ':' + $._config.versions.configmapReloader,
          '--prometheus-config-reloader=' + $._config.imageRepos.prometheusConfigReloader + ':' + $._config.versions.prometheusConfigReloader,
        ]) +
        container.mixin.securityContext.withAllowPrivilegeEscalation(false) +
        container.mixin.resources.withRequests({ cpu: '100m', memory: '100Mi' }) +
        container.mixin.resources.withLimits({ cpu: '200m', memory: '200Mi' });

      deployment.new('prometheus-operator', 1, operatorContainer, $._config.prometheusOperator.commonLabels) +
      deployment.mixin.metadata.withNamespace($._config.namespace) +
      deployment.mixin.metadata.withLabels($._config.prometheusOperator.commonLabels) +
      deployment.mixin.spec.selector.withMatchLabels($._config.prometheusOperator.deploymentSelectorLabels) +
      deployment.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.withServiceAccountName('prometheus-operator'),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('prometheus-operator') +
      serviceAccount.mixin.metadata.withLabels($._config.prometheusOperator.commonLabels) +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local poServicePort = servicePort.newNamed('http', 8080, 'http');

      service.new('prometheus-operator', $.prometheusOperator.deployment.spec.selector.matchLabels, [poServicePort]) +
      service.mixin.metadata.withLabels($._config.prometheusOperator.commonLabels) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.spec.withClusterIp('None'),
    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus-operator',
          namespace: $._config.namespace,
          labels: $._config.prometheusOperator.commonLabels,
        },
        spec: {
          endpoints: [
            {
              port: 'http',
              honorLabels: true,
            },
          ],
          selector: {
            matchLabels: $._config.prometheusOperator.commonLabels,
          },
        },
      },
  },
}
