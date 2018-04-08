local k = import "ksonnet.beta.3/k.libsonnet";
local role = k.rbac.v1.role;
local policyRule = role.rulesType;

local coreRule = policyRule.new() +
  policyRule.withApiGroups([""]) +
  policyRule.withResources([
    "pods",
  ]) +
  policyRule.withVerbs(["get"]);

local extensionsRule = policyRule.new() +
  policyRule.withApiGroups(["extensions"]) +
  policyRule.withResources([
    "deployments",
  ]) +
  policyRule.withVerbs(["get", "update"]) +
  policyRule.withResourceNames(["kube-state-metrics"]);

local rules = [coreRule, extensionsRule];

{
    new(namespace)::
        role.new() +
          role.mixin.metadata.withName("kube-state-metrics") +
          role.mixin.metadata.withNamespace(namespace) +
          role.withRules(rules)
}
