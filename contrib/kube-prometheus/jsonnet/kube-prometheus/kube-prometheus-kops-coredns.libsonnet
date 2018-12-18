local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

{
  prometheus+:: {
      kubeDnsPrometheusDiscoveryService:
      service.new('kube-dns-prometheus-discovery', { 'k8s-app': 'kube-dns' }, [servicePort.newNamed('metrics', 9153, 9153)]) +
      service.mixin.metadata.withNamespace('kube-system') +
      service.mixin.metadata.withLabels({ 'k8s-app': 'kube-dns' }) +
      service.mixin.spec.withClusterIp('None'),
  },
}
