local k = import "ksonnet.beta.3/k.libsonnet";

local alertmanager = import "alertmanager/alertmanager.libsonnet";
local ksm = import "kube-state-metrics/kube-state-metrics.libsonnet";
local nodeExporter = import "node-exporter/node-exporter.libsonnet";
local po = import "prometheus-operator/prometheus-operator.libsonnet";
local prometheus = import "prometheus/prometheus.libsonnet";
local grafana = import "grafana/grafana.libsonnet";

local alertmanagerConfig = importstr "../assets/alertmanager/alertmanager.yaml";

local ruleFiles = {
    "alertmanager.rules.yaml":            importstr "../assets/prometheus/rules/alertmanager.rules.yaml",
    "etcd3.rules.yaml":                   importstr "../assets/prometheus/rules/etcd3.rules.yaml",
    "general.rules.yaml":                 importstr "../assets/prometheus/rules/general.rules.yaml",
    "kube-controller-manager.rules.yaml": importstr "../assets/prometheus/rules/kube-controller-manager.rules.yaml",
    "kube-scheduler.rules.yaml":          importstr "../assets/prometheus/rules/kube-scheduler.rules.yaml",
    "kube-state-metrics.rules.yaml":      importstr "../assets/prometheus/rules/kube-state-metrics.rules.yaml",
    "kubelet.rules.yaml":                 importstr "../assets/prometheus/rules/kubelet.rules.yaml",
    "kubernetes.rules.yaml":              importstr "../assets/prometheus/rules/kubernetes.rules.yaml",
    "node.rules.yaml":                    importstr "../assets/prometheus/rules/node.rules.yaml",
    "prometheus.rules.yaml":              importstr "../assets/prometheus/rules/prometheus.rules.yaml",
};

{
    new(namespace)::
        {
            "grafana/grafana-dashboard-definitions.yaml": grafana.dashboardDefinitions.new(namespace),
            "grafana/grafana-dashboard-sources.yaml":     grafana.dashboardSources.new(namespace),
            "grafana/grafana-datasources.yaml":           grafana.dashboardDatasources.new(namespace),
            "grafana/grafana-deployment.yaml":            grafana.deployment.new(namespace),
            "grafana/grafana-service-account.yaml":       grafana.serviceAccount.new(namespace),
            "grafana/grafana-service.yaml":               grafana.service.new(namespace),

            "alertmanager-main/alertmanager-main-secret.yaml":          alertmanager.config.new(namespace, alertmanagerConfig),
            "alertmanager-main/alertmanager-main-service-account.yaml": alertmanager.serviceAccount.new(namespace),
            "alertmanager-main/alertmanager-main-service.yaml":         alertmanager.service.new(namespace),
            "alertmanager-main/alertmanager-main-service-monitor.yaml": alertmanager.serviceMonitor.new(namespace),
            "alertmanager-main/alertmanager-main.yaml":                 alertmanager.alertmanager.new(namespace),

            "kube-state-metrics/kube-state-metrics-cluster-role-binding.yaml": ksm.clusterRoleBinding.new(namespace),
            "kube-state-metrics/kube-state-metrics-cluster-role.yaml":         ksm.clusterRole.new(),
            "kube-state-metrics/kube-state-metrics-deployment.yaml":           ksm.deployment.new(namespace),
            "kube-state-metrics/kube-state-metrics-role-binding.yaml":         ksm.roleBinding.new(namespace),
            "kube-state-metrics/kube-state-metrics-role.yaml":                 ksm.role.new(namespace),
            "kube-state-metrics/kube-state-metrics-service-account.yaml":      ksm.serviceAccount.new(namespace),
            "kube-state-metrics/kube-state-metrics-service.yaml":              ksm.service.new(namespace),
            "kube-state-metrics/kube-state-metrics-service-monitor.yaml":      ksm.serviceMonitor.new(namespace),

            "node-exporter/node-exporter-cluster-role-binding.yaml": nodeExporter.clusterRoleBinding.new(namespace),
            "node-exporter/node-exporter-cluster-role.yaml":         nodeExporter.clusterRole.new(),
            "node-exporter/node-exporter-daemonset.yaml":            nodeExporter.daemonset.new(namespace),
            "node-exporter/node-exporter-service-account.yaml":      nodeExporter.serviceAccount.new(namespace),
            "node-exporter/node-exporter-service.yaml":              nodeExporter.service.new(namespace),
            "node-exporter/node-exporter-service-monitor.yaml":      nodeExporter.serviceMonitor.new(namespace),

            "prometheus-operator/prometheus-operator-cluster-role-binding.yaml": po.clusterRoleBinding.new(namespace),
            "prometheus-operator/prometheus-operator-cluster-role.yaml":         po.clusterRole.new(),
            "prometheus-operator/prometheus-operator-deployment.yaml":           po.deployment.new(namespace),
            "prometheus-operator/prometheus-operator-service.yaml":              po.service.new(namespace),
            "prometheus-operator/prometheus-operator-service-monitor.yaml":      po.serviceMonitor.new(namespace),
            "prometheus-operator/prometheus-operator-service-account.yaml":      po.serviceAccount.new(namespace),

            "prometheus-k8s/prometheus-k8s-cluster-role-binding.yaml":                    prometheus.clusterRoleBinding.new(namespace),
            "prometheus-k8s/prometheus-k8s-cluster-role.yaml":                            prometheus.clusterRole.new(),
            "prometheus-k8s/prometheus-k8s-service-account.yaml":                         prometheus.serviceAccount.new(namespace),
            "prometheus-k8s/prometheus-k8s-service.yaml":                                 prometheus.service.new(namespace),
            "prometheus-k8s/prometheus-k8s.yaml":                                         prometheus.prometheus.new(namespace),
            "prometheus-k8s/prometheus-k8s-rules.yaml":                                   prometheus.rules.new(namespace, ruleFiles),
            "prometheus-k8s/prometheus-k8s-role-binding-config.yaml":                     prometheus.roleBindingConfig.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-binding-namespace.yaml":                  prometheus.roleBindingNamespace.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-binding-kube-system.yaml":                prometheus.roleBindingKubeSystem.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-binding-default.yaml":                    prometheus.roleBindingDefault.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-config.yaml":                             prometheus.roleConfig.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-namespace.yaml":                          prometheus.roleNamespace.new(namespace),
            "prometheus-k8s/prometheus-k8s-role-kube-system.yaml":                        prometheus.roleKubeSystem.new(),
            "prometheus-k8s/prometheus-k8s-role-default.yaml":                            prometheus.roleDefault.new(),
            "prometheus-k8s/prometheus-k8s-service-monitor-apiserver.yaml":               prometheus.serviceMonitorApiserver.new(namespace),
            "prometheus-k8s/prometheus-k8s-service-monitor-coredns.yaml":                 prometheus.serviceMonitorCoreDNS.new(namespace),
            "prometheus-k8s/prometheus-k8s-service-monitor-kube-controller-manager.yaml": prometheus.serviceMonitorControllerManager.new(namespace),
            "prometheus-k8s/prometheus-k8s-service-monitor-kube-scheduler.yaml":          prometheus.serviceMonitorScheduler.new(namespace),
            "prometheus-k8s/prometheus-k8s-service-monitor-kubelet.yaml":                 prometheus.serviceMonitorKubelet.new(namespace),
            "prometheus-k8s/prometheus-k8s-service-monitor-prometheus.yaml":              prometheus.serviceMonitorPrometheus.new(namespace),
        }
}
