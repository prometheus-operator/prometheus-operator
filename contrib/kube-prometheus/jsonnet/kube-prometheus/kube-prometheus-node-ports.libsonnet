local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

{
  prometheus+: {
    service+:
      service.mixin.spec.withPorts(servicePort.newNamed('web', 9090, 'web') + servicePort.withNodePort(30900)) +
      service.mixin.spec.withType('NodePort'),
  },
  alertmanager+: {
    service+:
      service.mixin.spec.withPorts(servicePort.newNamed('web', 9093, 'web') + servicePort.withNodePort(30903)) +
      service.mixin.spec.withType('NodePort'),
  },
  grafana+: {
    service+:
      service.mixin.spec.withPorts(servicePort.newNamed('http', 3000, 'http') + servicePort.withNodePort(30902)) +
      service.mixin.spec.withType('NodePort'),
  },
}
