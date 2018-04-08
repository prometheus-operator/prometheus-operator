local prometheusNamespaceRoleBinding = import "prometheus-namespace-role-binding.libsonnet";

{
    new(namespace):: prometheusNamespaceRoleBinding.new(namespace, "default", "prometheus-k8s")
}
