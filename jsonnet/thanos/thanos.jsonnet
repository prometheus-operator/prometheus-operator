local po = (import 'config.libsonnet').thanos;

{
  'prometheus.yaml': po.prometheus,
  'prometheus-cluster-role.yaml': po.clusterRole,
}
