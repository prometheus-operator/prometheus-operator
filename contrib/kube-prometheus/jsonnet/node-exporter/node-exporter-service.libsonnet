local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;

local nodeExporterDaemonset = import "node-exporter-daemonset.libsonnet";

local nodeExporterPort = servicePort.newNamed("https", 9100, "https");

{
    new(namespace)::
        service.new("node-exporter", nodeExporterDaemonset.new(namespace).spec.selector.matchLabels, nodeExporterPort) +
          service.mixin.metadata.withNamespace(namespace) +
          service.mixin.metadata.withLabels({"k8s-app": "node-exporter"})
}
