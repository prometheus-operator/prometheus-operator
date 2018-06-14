local po = (import 'prometheus-operator/prometheus-operator.libsonnet').prometheusOperator;

{
  'prometheus-operator-cluster-role-binding.yaml': po.clusterRoleBinding,
  'prometheus-operator-cluster-role.yaml': po.clusterRole,
  'prometheus-operator-service-account.yaml': po.serviceAccount,
  'prometheus-operator-deployment.yaml': po.deployment,
}
