local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

local prometheusPort = servicePort.newNamed("web", 9090, "web");


{
    new(namespace)::
        service.new("prometheus-k8s", {app: "prometheus", prometheus: "k8s"}, prometheusPort) +
          service.mixin.metadata.withNamespace(namespace) +
          service.mixin.metadata.withLabels({prometheus: "k8s"})
}
