local prometheusNamespaceRole = import "prometheus-namespace-role.libsonnet";

{
    new():: prometheusNamespaceRole.new("kube-system")
}
