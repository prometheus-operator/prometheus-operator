local k = import "ksonnet.beta.3/k.libsonnet";
local roleBinding = k.rbac.v1.roleBinding;

{
  new(serviceAccountNamespace, namespace, roleName, serviceAccountName)::
    roleBinding.new() +
      roleBinding.mixin.metadata.withName(roleName) +
      roleBinding.mixin.metadata.withNamespace(namespace) +
      roleBinding.mixin.roleRef.withApiGroup("rbac.authorization.k8s.io") +
      roleBinding.mixin.roleRef.withName(roleName) +
      roleBinding.mixin.roleRef.mixinInstance({kind: "Role"}) +
      roleBinding.withSubjects([{kind: "ServiceAccount", name: serviceAccountName, namespace: serviceAccountNamespace}])
}
