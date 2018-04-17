local k = import "ksonnet.beta.3/k.libsonnet";
local role = k.rbac.v1.role;
local policyRule = role.rulesType;

{
    new(namespace)::
        local coreRule = policyRule.new() +
          policyRule.withApiGroups([""]) +
          policyRule.withResources([
              "nodes",
              "services",
              "endpoints",
              "pods",
          ]) +
          policyRule.withVerbs(["get", "list", "watch"]);

        role.new() +
          role.mixin.metadata.withName("prometheus-k8s") +
          role.mixin.metadata.withNamespace(namespace) +
          role.withRules(coreRule)
}
