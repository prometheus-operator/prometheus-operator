local prometheusNamespaceRole = import "prometheus-namespace-role.libsonnet";

{
    new(namespace):: prometheusNamespaceRole.new(namespace)
}
