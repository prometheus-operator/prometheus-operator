local k = import "ksonnet.beta.2/k.libsonnet";
local deployment = k.extensions.v1beta1.deployment;

local po = import "./prometheus-operator.jsonnet";

local operatorDeployment = po +
  deployment.mixin.spec.template.spec.serviceAccountName("prometheus-operator");

operatorDeployment
