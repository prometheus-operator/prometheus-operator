local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'default',

    versions+:: {
      nodeExporter: 'v0.15.2',
      kubeRbacProxy: 'v0.3.1',
    },

    imageRepos+:: {
      nodeExporter: 'quay.io/prometheus/node-exporter',
      kubeRbacProxy: 'quay.io/coreos/kube-rbac-proxy',
    },
  },

  nodeExporter+:: {
    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('node-exporter') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('node-exporter') +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'node-exporter', namespace: $._config.namespace }]),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

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

      local rules = [authenticationRole, authorizationRole];

      clusterRole.new() +
      clusterRole.mixin.metadata.withName('node-exporter') +
      clusterRole.withRules(rules),

    daemonset:
      local daemonset = k.apps.v1beta2.daemonSet;
      local container = daemonset.mixin.spec.template.spec.containersType;
      local volume = daemonset.mixin.spec.template.spec.volumesType;
      local containerPort = container.portsType;
      local containerVolumeMount = container.volumeMountsType;
      local podSelector = daemonset.mixin.spec.template.spec.selectorType;
      local toleration = daemonset.mixin.spec.template.spec.tolerationsType;

      local podLabels = { app: 'node-exporter' };

      local masterToleration = toleration.new() +
                               toleration.withEffect('NoSchedule') +
                               toleration.withKey('node-role.kubernetes.io/master');

      local procVolumeName = 'proc';
      local procVolume = volume.fromHostPath(procVolumeName, '/proc');
      local procVolumeMount = containerVolumeMount.new(procVolumeName, '/host/proc');

      local sysVolumeName = 'sys';
      local sysVolume = volume.fromHostPath(sysVolumeName, '/sys');
      local sysVolumeMount = containerVolumeMount.new(sysVolumeName, '/host/sys');

      local nodeExporter =
        container.new('node-exporter', $._config.imageRepos.nodeExporter + ':' + $._config.versions.nodeExporter) +
        container.withArgs([
          '--web.listen-address=127.0.0.1:9101',
          '--path.procfs=/host/proc',
          '--path.sysfs=/host/sys',
        ]) +
        container.withVolumeMounts([procVolumeMount, sysVolumeMount]) +
        container.mixin.resources.withRequests({ cpu: '102m', memory: '180Mi' }) +
        container.mixin.resources.withLimits({ cpu: '102m', memory: '180Mi' });

      local proxy =
        container.new('kube-rbac-proxy', $._config.imageRepos.kubeRbacProxy + ':' + $._config.versions.kubeRbacProxy) +
        container.withArgs([
          '--secure-listen-address=:9100',
          '--upstream=http://127.0.0.1:9101/',
        ]) +
        container.withPorts(containerPort.new(9100) + containerPort.withHostPort(9100) + containerPort.withName('https')) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '20Mi' }) +
        container.mixin.resources.withLimits({ cpu: '20m', memory: '40Mi' });

      local c = [nodeExporter, proxy];

      daemonset.new() +
      daemonset.mixin.metadata.withName('node-exporter') +
      daemonset.mixin.metadata.withNamespace($._config.namespace) +
      daemonset.mixin.metadata.withLabels(podLabels) +
      daemonset.mixin.spec.selector.withMatchLabels(podLabels) +
      daemonset.mixin.spec.template.metadata.withLabels(podLabels) +
      daemonset.mixin.spec.template.spec.withTolerations([masterToleration]) +
      daemonset.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      daemonset.mixin.spec.template.spec.withContainers(c) +
      daemonset.mixin.spec.template.spec.withVolumes([procVolume, sysVolume]) +
      daemonset.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      daemonset.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      daemonset.mixin.spec.template.spec.withServiceAccountName('node-exporter') +
      daemonset.mixin.spec.template.spec.withHostPid(true) +
      daemonset.mixin.spec.template.spec.withHostNetwork(true),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('node-exporter') +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),

    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'node-exporter',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'node-exporter',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          selector: {
            matchLabels: {
              'k8s-app': 'node-exporter',
            },
          },
          endpoints: [
            {
              port: 'https',
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

    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local nodeExporterPort = servicePort.newNamed('https', 9100, 'https');

      service.new('node-exporter', $.nodeExporter.daemonset.spec.selector.matchLabels, nodeExporterPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ 'k8s-app': 'node-exporter' }) +
      service.mixin.spec.withClusterIp('None'),
  },
}
