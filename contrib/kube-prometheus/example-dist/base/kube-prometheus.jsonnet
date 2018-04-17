local kubePrometheus = import "kube-prometheus.libsonnet";

local namespace = "monitoring";
local objects = kubePrometheus.new(namespace);

{[path]: std.manifestYamlDoc(objects[path]) for path in std.objectFields(objects)}
