local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

local ksmDeployment = import "kube-state-metrics-deployment.libsonnet";

local ksmServicePortMain = servicePort.newNamed("https-main", 8443, "https-main");
local ksmServicePortSelf = servicePort.newNamed("https-self", 9443, "https-self");

{
    new(namespace)::
        service.new("kube-state-metrics", ksmDeployment.new(namespace).spec.selector.matchLabels, [ksmServicePortMain, ksmServicePortSelf]) +
          service.mixin.metadata.withNamespace(namespace) +
          service.mixin.metadata.withLabels({"k8s-app": "kube-state-metrics"})
}
