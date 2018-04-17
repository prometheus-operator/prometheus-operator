{
  clusterRoleBinding:: import "prometheus-operator-cluster-role-binding.libsonnet",
  clusterRole:: import "prometheus-operator-cluster-role.libsonnet",
  deployment:: import "prometheus-operator-deployment.libsonnet",
  serviceAccount:: import "prometheus-operator-service-account.libsonnet",
  service:: import "prometheus-operator-service.libsonnet",
}
