{
    clusterRoleBinding:: import "kube-state-metrics-cluster-role-binding.libsonnet",
    clusterRole:: import "kube-state-metrics-cluster-role.libsonnet",
    deployment:: import "kube-state-metrics-deployment.libsonnet",
    roleBinding:: import "kube-state-metrics-role-binding.libsonnet",
    role:: import "kube-state-metrics-role.libsonnet",
    serviceAccount:: import "kube-state-metrics-service-account.libsonnet",
    service:: import "kube-state-metrics-service.libsonnet",
    serviceMonitor:: import "kube-state-metrics-service-monitor.libsonnet",
}
