local k = import "ksonnet.beta.3/k.libsonnet";
local clusterRole = k.rbac.v1.clusterRole;
local policyRule = clusterRole.rulesType;

local nodeMetricsRule = policyRule.new() +
  policyRule.withApiGroups([""]) +
  policyRule.withResources(["nodes/metrics"]) +
  policyRule.withVerbs(["get"]);

local metricsRule = policyRule.new() +
  policyRule.withNonResourceUrls("/metrics") +
  policyRule.withVerbs(["get"]);

local rules = [nodeMetricsRule, metricsRule];

{
    new()::
        clusterRole.new() +
          clusterRole.mixin.metadata.withName("prometheus-k8s") +
          clusterRole.withRules(rules)
}
