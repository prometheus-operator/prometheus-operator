local prometheusOperator = (import 'prometheus-operator/prometheus-operator.libsonnet');
local config = (import 'config.jsonnet');

local po = prometheusOperator(config);

{
  'prometheus-operator-cluster-role-binding.yaml': po.clusterRoleBinding,
  'prometheus-operator-cluster-role.yaml': po.clusterRole,
  'prometheus-operator-service-account.yaml': po.serviceAccount,
  'prometheus-operator-deployment.yaml': po.deployment,
  'prometheus-operator-service.yaml': po.service,
  'prometheus-operator-service-monitor.yaml': po.serviceMonitor,
}
