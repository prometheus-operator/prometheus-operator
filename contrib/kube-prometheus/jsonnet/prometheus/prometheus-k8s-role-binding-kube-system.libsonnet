local prometheusNamespaceRoleBinding = import "prometheus-namespace-role-binding.libsonnet";

{
    new(namespace):: prometheusNamespaceRoleBinding.new(namespace, "kube-system", "prometheus-k8s", "prometheus-k8s")
}
