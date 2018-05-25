local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

{
  prometheus+:: {
    kubeControllerManagerPrometheusDiscoveryService:
      service.new('kube-controller-manager-prometheus-discovery', { 'k8s-app': 'kube-controller-manager' }, servicePort.newNamed('http-metrics', 10252, 10252)) +
      service.mixin.metadata.withNamespace('kube-system') +
      service.mixin.metadata.withLabels({ 'k8s-app': 'kube-controller-manager' }) +
      service.mixin.spec.withClusterIp('None'),
    kubeSchedulerPrometheusDiscoveryService:
      service.new('kube-scheduler-prometheus-discovery', { 'k8s-app': 'kube-scheduler' }, servicePort.newNamed('http-metrics', 10251, 10251)) +
      service.mixin.metadata.withNamespace('kube-system') +
      service.mixin.metadata.withLabels({ 'k8s-app': 'kube-scheduler' }) +
      service.mixin.spec.withClusterIp('None'),
    kubeDnsPrometheusDiscoveryService:
      service.new('kube-dns-prometheus-discovery', { 'k8s-app': 'kube-dns' }, [servicePort.newNamed('http-metrics-skydns', 10055, 10055), servicePort.newNamed('http-metrics-dnsmasq', 10054, 10054)]) +
      service.mixin.metadata.withNamespace('kube-system') +
      service.mixin.metadata.withLabels({ 'k8s-app': 'kube-dns' }) +
      service.mixin.spec.withClusterIp('None'),
  },
}
