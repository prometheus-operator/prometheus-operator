local po = (import 'thanos/config.libsonnet').thanos;

{
  'prometheus.yaml': po.prometheus,
  'prometheus-cluster-role.yaml': po.clusterRole,
  'prometheus-cluster-role-binding.yaml': po.clusterRoleBinding,
  'prometheus-service.yaml': po.service,
  'prometheus-servicemonitor.yaml': po.serviceMonitor,
  'query-deployment.yaml': po.queryDeployment,
  'query-service.yaml': po.queryService,
  'sidecar-service.yaml': po.sidecarService,
  'thanos-ruler.yaml': po.thanosRuler,
  'prometheus-rule.yaml': po.prometheusRule,
  'thanos-ruler-service.yaml': po.thanosRulerService,
}
