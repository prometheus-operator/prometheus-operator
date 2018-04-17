local k = import "ksonnet.beta.3/k.libsonnet";
local role = k.rbac.v1.role;
local policyRule = role.rulesType;

local configmapRule = policyRule.new() +
  policyRule.withApiGroups([""]) +
  policyRule.withResources([
    "configmaps",
  ]) +
  policyRule.withVerbs(["get"]);

{
    new(namespace)::
        role.new() +
          role.mixin.metadata.withName("prometheus-k8s-config") +
          role.mixin.metadata.withNamespace(namespace) +
          role.withRules(configmapRule),
}
