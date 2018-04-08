local k = import "ksonnet.beta.3/k.libsonnet";

local alertmanager = import "alertmanager/alertmanager.libsonnet";
local ksm = import "kube-state-metrics/kube-state-metrics.libsonnet";
local nodeExporter = import "node-exporter/node-exporter.libsonnet";
local po = import "prometheus-operator/prometheus-operator.libsonnet";
local prometheus = import "prometheus/prometheus.libsonnet";

local namespace = "monitoring";

local objects = {
    "alertmanager-main/alertmanager-main-secret.yaml":          alertmanager.config.new(namespace),
    "alertmanager-main/alertmanager-main-service-account.yaml": alertmanager.serviceAccount.new(namespace),
    "alertmanager-main/alertmanager-main-service.yaml":         alertmanager.service.new(namespace),
    "alertmanager-main/alertmanager-main.yaml":                 alertmanager.alertmanager.new(namespace),

    "kube-state-metrics/kube-state-metrics-cluster-role-binding": ksm.clusterRoleBinding.new(namespace),
    "kube-state-metrics/kube-state-metrics-cluster-role.yaml":    ksm.clusterRole.new(),
    "kube-state-metrics/kube-state-metrics-deployment.yaml":      ksm.deployment.new(namespace),
    "kube-state-metrics/kube-state-metrics-role-binding.yaml":    ksm.roleBinding.new(namespace),
    "kube-state-metrics/kube-state-metrics-role.yaml":            ksm.role.new(namespace),
    "kube-state-metrics/kube-state-metrics-service-account.yaml": ksm.serviceAccount.new(namespace),
    "kube-state-metrics/kube-state-metrics-service.yaml":         ksm.service.new(namespace),

    "node-exporter/node-exporter-cluster-role-binding.yaml": nodeExporter.clusterRoleBinding.new(namespace),
    "node-exporter/node-exporter-cluster-role.yaml": nodeExporter.clusterRole.new(),
    "node-exporter/node-exporter-daemonset.yaml": nodeExporter.daemonset.new(namespace),
    "node-exporter/node-exporter-service-account.yaml": nodeExporter.serviceAccount.new(namespace),
    "node-exporter/node-exporter-service.yaml": nodeExporter.service.new(namespace),

    "prometheus-operator/prometheus-operator-cluster-role-binding.yaml": po.clusterRoleBinding.new(namespace),
    "prometheus-operator/prometheus-operator-cluster-role.yaml":         po.clusterRole.new(),
    "prometheus-operator/prometheus-operator-deployment.yaml":           po.deployment.new(namespace),
    "prometheus-operator/prometheus-operator-service.yaml":              po.service.new(namespace),
    "prometheus-operator/prometheus-operator-service-account.yaml":      po.serviceAccount.new(namespace),

    "prometheus-k8s/prometheus-k8s-cluster-role-binding.yaml":                    prometheus.clusterRoleBinding.new(namespace),
    "prometheus-k8s/prometheus-k8s-cluster-role.yaml":                            prometheus.clusterRole.new(),
    "prometheus-k8s/prometheus-k8s-service-account.yaml":                         prometheus.serviceAccount.new(namespace),
    "prometheus-k8s/prometheus-k8s-service.yaml":                                 prometheus.service.new(namespace),
    "prometheus-k8s/prometheus-k8s.yaml":                                         prometheus.prometheus.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-binding-config.yaml":                     prometheus.roleBindingConfig.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-binding-namespace.yaml":                  prometheus.roleBindingNamespace.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-binding-kube-system.yaml":                prometheus.roleBindingKubeSystem.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-binding-default.yaml":                    prometheus.roleBindingDefault.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-config.yaml":                             prometheus.roleConfig.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-namespace.yaml":                          prometheus.roleNamespace.new(namespace),
    "prometheus-k8s/prometheus-k8s-role-kube-system.yaml":                        prometheus.roleKubeSystem.new(),
    "prometheus-k8s/prometheus-k8s-role-default.yaml":                            prometheus.roleDefault.new(),
    "prometheus-k8s/prometheus-k8s-service-monitor-alertmanager.yaml":            prometheus.serviceMonitorAlertmanager.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-apiserver.yaml":               prometheus.serviceMonitorApiserver.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-coredns.yaml":                 prometheus.serviceMonitorCoreDNS.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-kube-controller-manager.yaml": prometheus.serviceMonitorControllerManager.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-kube-scheduler.yaml":          prometheus.serviceMonitorScheduler.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-kube-state-metrics.yaml":      prometheus.serviceMonitorKubeStateMetrics.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-kubelet.yaml":                 prometheus.serviceMonitorKubelet.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-node-exporter.yaml":           prometheus.serviceMonitorNodeExporter.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-prometheus-operator.yaml":     prometheus.serviceMonitorPrometheusOperator.new(namespace),
    "prometheus-k8s/prometheus-k8s-service-monitor-prometheus.yaml":              prometheus.serviceMonitorPrometheus.new(namespace),
};

{[path]: std.manifestYamlDoc(objects[path]) for path in std.objectFields(objects)}
