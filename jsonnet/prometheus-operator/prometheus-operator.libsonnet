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
        $._config.prometheusOperator.deploymentSelectorLabels
        { 'app.kubernetes.io/version': $._config.versions.prometheusOperator },
    },

    versions+:: {
      prometheusOperator: 'v0.42.1',
      prometheusConfigReloader: self.prometheusOperator,
      configmapReloader: 'v0.4.0',
    },

    imageRepos+:: {
      prometheusOperator: 'quay.io/prometheus-operator/prometheus-operator',
      configmapReloader: 'jimmidyson/configmap-reload',
      prometheusConfigReloader: 'quay.io/prometheus-operator/prometheus-config-reloader',
    },
  },

  prometheusOperator+:: {
    local po = self,

    namespace:: $._config.namespace,
    commonLabels:: $._config.prometheusOperator.commonLabels,
    deploymentSelectorLabels:: $._config.prometheusOperator.deploymentSelectorLabels,

    image:: $._config.imageRepos.prometheusOperator,
    version:: $._config.versions.prometheusOperator,
    configReloaderImage:: $._config.imageRepos.configmapReloader,
    configReloaderVersion:: $._config.versions.configmapReloader,
    prometheusConfigReloaderImage:: $._config.imageRepos.prometheusConfigReloader,
    prometheusConfigReloaderVersion:: $._config.versions.prometheusConfigReloader,

    // Prefixing with 0 to ensure these manifests are listed and therefore created first.
    '0alertmanagerCustomResourceDefinition': import 'alertmanager-crd.libsonnet',
    '0prometheusCustomResourceDefinition': import 'prometheus-crd.libsonnet',
    '0servicemonitorCustomResourceDefinition': import 'servicemonitor-crd.libsonnet',
    '0podmonitorCustomResourceDefinition': import 'podmonitor-crd.libsonnet',
    '0probeCustomResourceDefinition': import 'probe-crd.libsonnet',
    '0prometheusruleCustomResourceDefinition': import 'prometheusrule-crd.libsonnet',
    '0thanosrulerCustomResourceDefinition': import 'thanosruler-crd.libsonnet',

    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withLabels(po.commonLabels) +
      clusterRoleBinding.mixin.metadata.withName('prometheus-operator') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('prometheus-operator') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-operator', namespace: po.namespace }]),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local monitoringRule = policyRule.new() +
                             policyRule.withApiGroups(['monitoring.coreos.com']) +
                             policyRule.withResources([
                               'alertmanagers',
                               'alertmanagers/finalizers',
                               'prometheuses',
                               'prometheuses/finalizers',
                               'thanosrulers',
                               'thanosrulers/finalizers',
                               'servicemonitors',
                               'podmonitors',
                               'probes',
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

      local rules = [monitoringRule, appsRule, coreRule, podRule, routingRule, nodeRule, namespaceRule];

      clusterRole.new() +
      clusterRole.mixin.metadata.withLabels(po.commonLabels) +
      clusterRole.mixin.metadata.withName('prometheus-operator') +
      clusterRole.withRules(rules),

    deployment:
      local deployment = k.apps.v1.deployment;
      local container = k.apps.v1.deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;

      local targetPort = 8080;

      local operatorContainer =
        container.new('prometheus-operator', po.image + ':' + po.version) +
        container.withPorts(containerPort.newNamed(targetPort, 'http')) +
        container.withArgs([
          '--kubelet-service=kube-system/kubelet',
          // Prometheus Operator is run with a read-only root file system. By
          // default glog saves logfiles to /tmp. Make it log to stderr instead.
          '--logtostderr=true',
          '--config-reloader-image=' + po.configReloaderImage + ':' + po.configReloaderVersion,
          '--prometheus-config-reloader=' + po.prometheusConfigReloaderImage + ':' + po.prometheusConfigReloaderVersion,
        ]) +
        container.mixin.securityContext.withAllowPrivilegeEscalation(false) +
        container.mixin.resources.withRequests({ cpu: '100m', memory: '100Mi' }) +
        container.mixin.resources.withLimits({ cpu: '200m', memory: '200Mi' });

      deployment.new('prometheus-operator', 1, operatorContainer, po.commonLabels) +
      deployment.mixin.metadata.withNamespace(po.namespace) +
      deployment.mixin.metadata.withLabels(po.commonLabels) +
      deployment.mixin.spec.selector.withMatchLabels(po.deploymentSelectorLabels) +
      deployment.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.withServiceAccountName('prometheus-operator'),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('prometheus-operator') +
      serviceAccount.mixin.metadata.withLabels(po.commonLabels) +
      serviceAccount.mixin.metadata.withNamespace(po.namespace),

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local poServicePort = servicePort.newNamed('http', 8080, 'http');

      service.new('prometheus-operator', po.deployment.spec.selector.matchLabels, [poServicePort]) +
      service.mixin.metadata.withLabels(po.commonLabels) +
      service.mixin.metadata.withNamespace(po.namespace) +
      service.mixin.spec.withClusterIp('None'),
    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus-operator',
          namespace: po.namespace,
          labels: po.commonLabels,
        },
        spec: {
          endpoints: [
            {
              port: 'http',
              honorLabels: true,
            },
          ],
          selector: {
            matchLabels: po.commonLabels,
          },
        },
      },
  },
}
