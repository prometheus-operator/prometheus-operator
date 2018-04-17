local k = import "ksonnet.beta.3/k.libsonnet";
local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

{
    new(namespace)::
        clusterRoleBinding.new() +
          clusterRoleBinding.mixin.metadata.withName("prometheus-k8s") +
          clusterRoleBinding.mixin.roleRef.withApiGroup("rbac.authorization.k8s.io") +
          clusterRoleBinding.mixin.roleRef.withName("prometheus-k8s") +
          clusterRoleBinding.mixin.roleRef.mixinInstance({kind: "ClusterRole"}) +
          clusterRoleBinding.withSubjects([{kind: "ServiceAccount", name: "prometheus-k8s", namespace: namespace}])
}
