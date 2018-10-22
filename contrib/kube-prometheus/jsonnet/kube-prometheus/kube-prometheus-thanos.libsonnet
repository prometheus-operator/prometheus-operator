local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;


{
  _config+:: {
    versions+:: {
      thanos: 'v0.1.0',
    },
    imageRepos+:: {
      thanos: 'improbable/thanos',
    },
  },
  prometheus+:: {
    prometheus+: {
      spec+: {
        podMetadata+: {
          labels+: { 'thanos-peer': 'true' },
        },
        thanos+: {
          peers: 'thanos-peers.' + $._config.namespace + '.svc:10900',
          version: $._config.versions.thanos,
          baseImage: $._config.imageRepos.thanos,
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
          '--cluster.peers=thanos-peers.' + $._config.namespace + '.svc:10900',
        ]);
      local podLabels = { app: 'thanos-query', 'thanos-peer': 'true' };
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
    thanosPeerService:
      local thanosPeerPort = servicePort.newNamed('cluster', 10900, 'cluster');
      service.new('thanos-peers', { 'thanos-peer': 'true' }, thanosPeerPort) +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.spec.withType('ClusterIP') +
      service.mixin.spec.withClusterIp('None'),
  },
}
