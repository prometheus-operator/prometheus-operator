local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

{
  _config+:: {
    versions+:: {
      thanos: 'v0.3.2',
    },
    imageRepos+:: {
      thanos: 'improbable/thanos',
    },
    thanos+:: {
      objectStorageConfig: {
        key: 'thanos.yaml',  // How the file inside the secret is called
        name: 'thanos-objstore-config',  // This is the name of your Kubernetes secret with the config
      },
    },
  },
  prometheus+:: {
    prometheus+: {
      spec+: {
        podMetadata+: {
          labels+: { 'thanos-peers': 'true' },
        },
        thanos+: {
          peers: 'thanos-peers.' + $._config.namespace + '.svc:10900',
          version: $._config.versions.thanos,
          baseImage: $._config.imageRepos.thanos,
          objectStorageConfig: $._config.thanos.objectStorageConfig,
        },
      },
    },
    thanosPeerService:
      service.new('thanos-peers', { 'thanos-peers': 'true' }, [
        servicePort.newNamed('cluster', 10900, 'cluster'),
        servicePort.newNamed('http', 10902, 'http'),
      ]) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ 'thanos-peers': 'true' }) +
      service.mixin.spec.withType('ClusterIP') +
      service.mixin.spec.withClusterIp('None'),

    serviceMonitorThanosPeer:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'thanos-peers',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'thanos-peers',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'http',
              interval: '30s',
            },
          ],
          selector: {
            matchLabels: {
              'thanos-peers': 'true',
            },
          },
        },
      },
    thanosQueryDeployment:
      local deployment = k.apps.v1beta2.deployment;
      local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;

      local thanosQueryContainer =
        container.new('thanos-query', $._config.imageRepos.thanos + ':' + $._config.versions.thanos) +
        container.withPorts([
          containerPort.newNamed('http', 10902),
          containerPort.newNamed('grpc', 10901),
          containerPort.newNamed('cluster', 10900),
        ]) +
        container.withArgs([
          'query',
          '--log.level=debug',
          '--query.replica-label=prometheus_replica',
          '--query.auto-downsampling',
          '--cluster.peers=thanos-peers.' + $._config.namespace + '.svc:10900',
        ]);
      local podLabels = { app: 'thanos-query', 'thanos-peers': 'true' };
      deployment.new('thanos-query', 1, thanosQueryContainer, podLabels) +
      deployment.mixin.metadata.withNamespace($._config.namespace) +
      deployment.mixin.metadata.withLabels(podLabels) +
      deployment.mixin.spec.selector.withMatchLabels(podLabels) +
      deployment.mixin.spec.template.spec.withServiceAccountName('prometheus-' + $._config.prometheus.name),
    thanosQueryService:
      local thanosQueryPort = servicePort.newNamed('http-query', 9090, 'http');
      service.new('thanos-query', { app: 'thanos-query' }, thanosQueryPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ app: 'thanos-query' }),

    thanosStoreStatefulset:
      local statefulSet = k.apps.v1beta2.statefulSet;
      local volume = statefulSet.mixin.spec.template.spec.volumesType;
      local container = statefulSet.mixin.spec.template.spec.containersType;
      local containerEnv = container.envType;
      local containerVolumeMount = container.volumeMountsType;

      local labels = { app: 'thanos', 'thanos-peers': 'true' };

      local c =
        container.new('thanos-store', $._config.imageRepos.thanos + ':' + $._config.versions.thanos) +
        container.withArgs([
          'store',
          '--log.level=debug',
          '--data-dir=/var/thanos/store',
          '--cluster.peers=thanos-peers.' + $._config.namespace + '.svc:10900',
          '--objstore.config=$(OBJSTORE_CONFIG)',
        ]) +
        container.withEnv([
          containerEnv.fromSecretRef(
            'OBJSTORE_CONFIG',
            $._config.thanos.objectStorageConfig.name,
            $._config.thanos.objectStorageConfig.key,
          ),
        ]) +
        container.withPorts([
          { name: 'cluster', containerPort: 10900 },
          { name: 'grpc', containerPort: 10901 },
          { name: 'http', containerPort: 10902 },
        ]) +
        container.withVolumeMounts([
          containerVolumeMount.new('data', '/var/thanos/store', false),
        ]);

      statefulSet.new('thanos-store', 1, c, [], labels) +
      statefulSet.mixin.metadata.withNamespace($._config.namespace) +
      statefulSet.mixin.spec.selector.withMatchLabels(labels) +
      statefulSet.mixin.spec.withServiceName('thanos-store') +
      statefulSet.mixin.spec.template.spec.withVolumes([
        volume.fromEmptyDir('data'),
      ]),

    serviceMonitorThanosCompactor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'thanos-compactor',
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'thanos-compactor',
          },
        },
        spec: {
          jobLabel: 'k8s-app',
          endpoints: [
            {
              port: 'http',
              interval: '30s',
            },
          ],
          selector: {
            matchLabels: {
              app: 'thanos-compactor',
            },
          },
        },
      },

    thanosCompactorService:
      service.new(
        'thanos-compactor',
        { app: 'thanos-compactor' },
        servicePort.newNamed('http', 9090, 'http'),
      ) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ app: 'thanos-compactor' }),

    thanosCompactorStatefulset:
      local statefulSet = k.apps.v1beta2.statefulSet;
      local volume = statefulSet.mixin.spec.template.spec.volumesType;
      local container = statefulSet.mixin.spec.template.spec.containersType;
      local containerEnv = container.envType;
      local containerVolumeMount = container.volumeMountsType;

      local labels = { app: 'thanos-compactor' };

      local c =
        container.new('thanos-compactor', $._config.imageRepos.thanos + ':' + $._config.versions.thanos) +
        container.withArgs([
          'compact',
          '--log.level=debug',
          '--data-dir=/var/thanos/store',
          '--objstore.config=$(OBJSTORE_CONFIG)',
          '--wait',
        ]) +
        container.withEnv([
          containerEnv.fromSecretRef(
            'OBJSTORE_CONFIG',
            $._config.thanos.objectStorageConfig.name,
            $._config.thanos.objectStorageConfig.key,
          ),
        ]) +
        container.withPorts([
          { name: 'http', containerPort: 10902 },
        ]) +
        container.withVolumeMounts([
          containerVolumeMount.new('data', '/var/thanos/store', false),
        ]);

      statefulSet.new('thanos-compactor', 1, c, [], labels) +
      statefulSet.mixin.metadata.withNamespace($._config.namespace) +
      statefulSet.mixin.spec.selector.withMatchLabels(labels) +
      statefulSet.mixin.spec.withServiceName('thanos-compactor') +
      statefulSet.mixin.spec.template.spec.withVolumes([
        volume.fromEmptyDir('data'),
      ]),
  },
}
