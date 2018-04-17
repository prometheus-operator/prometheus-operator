local k = import "ksonnet.beta.3/k.libsonnet";
local clusterRole = k.rbac.v1.clusterRole;
local policyRule = clusterRole.rulesType;

local coreRule = policyRule.new() +
  policyRule.withApiGroups([""]) +
  policyRule.withResources([
    "configmaps",
    "secrets",
    "nodes",
    "pods",
    "services",
    "resourcequotas",
    "replicationcontrollers",
    "limitranges",
    "persistentvolumeclaims",
    "persistentvolumes",
    "namespaces",
    "endpoints",
  ]) +
  policyRule.withVerbs(["list", "watch"]);

local extensionsRule = policyRule.new() +
  policyRule.withApiGroups(["extensions"]) +
  policyRule.withResources([
    "daemonsets",
    "deployments",
    "replicasets",
  ]) +
  policyRule.withVerbs(["list", "watch"]);

local appsRule = policyRule.new() +
  policyRule.withApiGroups(["apps"]) +
  policyRule.withResources([
    "statefulsets",
  ]) +
  policyRule.withVerbs(["list", "watch"]);

local batchRule = policyRule.new() +
  policyRule.withApiGroups(["batch"]) +
  policyRule.withResources([
    "cronjobs",
    "jobs",
  ]) +
  policyRule.withVerbs(["list", "watch"]);

local autoscalingRule = policyRule.new() +
  policyRule.withApiGroups(["autoscaling"]) +
  policyRule.withResources([
    "horizontalpodautoscalers",
  ]) +
  policyRule.withVerbs(["list", "watch"]);

local authenticationRole = policyRule.new() +
  policyRule.withApiGroups(["authentication.k8s.io"]) +
  policyRule.withResources([
    "tokenreviews",
  ]) +
  policyRule.withVerbs(["create"]);

local authorizationRole = policyRule.new() +
  policyRule.withApiGroups(["authorization.k8s.io"]) +
  policyRule.withResources([
    "subjectaccessreviews",
  ]) +
  policyRule.withVerbs(["create"]);

local rules = [coreRule, extensionsRule, appsRule, batchRule, autoscalingRule, authenticationRole, authorizationRole];

{
    new()::
        clusterRole.new() +
          clusterRole.mixin.metadata.withName("kube-state-metrics") +
          clusterRole.withRules(rules)
}
