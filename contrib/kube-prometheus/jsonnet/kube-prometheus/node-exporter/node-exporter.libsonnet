local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

local daemonset = k.apps.v1beta2.daemonSet;
local toleration = daemonset.mixin.spec.template.spec.tolerationsType;
{
  _config+:: {
    namespace: 'default',

    versions+:: {
      nodeExporter: 'v0.16.0',
      kubeRbacProxy: 'v0.4.0',
    },

    imageRepos+:: {
      nodeExporter: 'quay.io/prometheus/node-exporter',
      kubeRbacProxy: 'quay.io/coreos/kube-rbac-proxy',
    },
    tolerations+::[
        toleration.new() +
            toleration.withEffect('NoSchedule') +
            toleration.withKey('node-role.kubernetes.io/master'),
    ],
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
      local containerEnv = container.envType;

      local podLabels = { app: 'node-exporter' };

      local procVolumeName = 'proc';
      local procVolume = volume.fromHostPath(procVolumeName, '/proc');
      local procVolumeMount = containerVolumeMount.new(procVolumeName, '/host/proc');

      local sysVolumeName = 'sys';
      local sysVolume = volume.fromHostPath(sysVolumeName, '/sys');
      local sysVolumeMount = containerVolumeMount.new(sysVolumeName, '/host/sys');

      local rootVolumeName = 'root';
      local rootVolume = volume.fromHostPath(rootVolumeName, '/');
      local rootVolumeMount = containerVolumeMount.new(rootVolumeName, '/host/root').
        withMountPropagation('HostToContainer').
        withReadOnly(true);

      local nodeExporter =
        container.new('node-exporter', $._config.imageRepos.nodeExporter + ':' + $._config.versions.nodeExporter) +
        container.withArgs([
          '--web.listen-address=127.0.0.1:9100',
          '--path.procfs=/host/proc',
          '--path.sysfs=/host/sys',

          // The following settings have been taken from
          // https://github.com/prometheus/node_exporter/blob/0662673/collector/filesystem_linux.go#L30-L31
          // Once node exporter is being released with those settings, this can be removed.
          '--collector.filesystem.ignored-mount-points=^/(dev|proc|sys|var/lib/docker/.+)($|/)',
          '--collector.filesystem.ignored-fs-types=^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$',
        ]) +
        container.withVolumeMounts([procVolumeMount, sysVolumeMount, rootVolumeMount]) +
        container.mixin.resources.withRequests({ cpu: '102m', memory: '180Mi' }) +
        container.mixin.resources.withLimits({ cpu: '250m', memory: '180Mi' });

      local ip = containerEnv.fromFieldPath('IP', 'status.podIP');
      local proxy =
        container.new('kube-rbac-proxy', $._config.imageRepos.kubeRbacProxy + ':' + $._config.versions.kubeRbacProxy) +
        container.withArgs([
          '--secure-listen-address=$(IP):9100',
          '--upstream=http://127.0.0.1:9100/',
        ]) +
        // Keep `hostPort` here, rather than in the node-exporter container
        // because Kubernetes mandates that if you define a `hostPort` then
        // `containerPort` must match. In our case, we are splitting the
        // host port and container port between the two containers.
        // We'll keep the port specification here so that the named port
        // used by the service is tied to the proxy container. We *could*
        // forgo declaring the host port, however it is important to declare
        // it so that the scheduler can decide if the pod is schedulable.
        container.withPorts(containerPort.new(9100) + containerPort.withHostPort(9100) + containerPort.withName('https')) +
        container.mixin.resources.withRequests({ cpu: '10m', memory: '20Mi' }) +
        container.mixin.resources.withLimits({ cpu: '20m', memory: '40Mi' }) +
        container.withEnv([ip]);

      local c = [nodeExporter, proxy];

      daemonset.new() +
      daemonset.mixin.metadata.withName('node-exporter') +
      daemonset.mixin.metadata.withNamespace($._config.namespace) +
      daemonset.mixin.metadata.withLabels(podLabels) +
      daemonset.mixin.spec.selector.withMatchLabels(podLabels) +
      daemonset.mixin.spec.template.metadata.withLabels(podLabels) +
      daemonset.mixin.spec.template.spec.withTolerations($._config.tolerations) +
      daemonset.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      daemonset.mixin.spec.template.spec.withContainers(c) +
      daemonset.mixin.spec.template.spec.withVolumes([procVolume, sysVolume, rootVolume]) +
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
