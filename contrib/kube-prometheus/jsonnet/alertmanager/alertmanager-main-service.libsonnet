local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

local alertmanagerPort = servicePort.newNamed("web", 9093, "web");

{
    new(namespace)::
        service.new("alertmanager-main", {app: "alertmanager", alertmanager: "main"}, alertmanagerPort) +
          service.mixin.metadata.withNamespace(namespace) +
          service.mixin.metadata.withLabels({alertmanager: "main"})
}
