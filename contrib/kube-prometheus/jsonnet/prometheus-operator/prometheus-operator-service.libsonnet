local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

local poDeployment = import "prometheus-operator-deployment.libsonnet";

local poServicePort = servicePort.newNamed("http", 8080, "http");


{
    new(namespace)::
        service.new("prometheus-operator", poDeployment.new(namespace).spec.selector.matchLabels, [poServicePort]) +
        service.mixin.metadata.withNamespace(namespace)
}
