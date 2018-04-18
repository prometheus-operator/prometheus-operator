{
    clusterRoleBinding:: import "node-exporter-cluster-role-binding.libsonnet",
    clusterRole:: import "node-exporter-cluster-role.libsonnet",
    daemonset:: import "node-exporter-daemonset.libsonnet",
    serviceAccount:: import "node-exporter-service-account.libsonnet",
    service:: import "node-exporter-service.libsonnet",
    serviceMonitor:: import "node-exporter-service-monitor.libsonnet",
}
