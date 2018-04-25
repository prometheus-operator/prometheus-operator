local k = import "ksonnet/ksonnet.beta.3/k.libsonnet";
local deployment = k.apps.v1beta2.deployment;

local po = import "./prometheus-operator.jsonnet";

local operatorDeployment = po +
  deployment.mixin.spec.template.spec.withServiceAccountName("prometheus-operator");

operatorDeployment
