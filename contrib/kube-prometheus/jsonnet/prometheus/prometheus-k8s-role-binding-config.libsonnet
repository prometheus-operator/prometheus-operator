local prometheusNamespaceRoleBinding = import "prometheus-namespace-role-binding.libsonnet";

{
    new(namespace):: prometheusNamespaceRoleBinding.new(namespace, namespace, "prometheus-k8s-config")
}
