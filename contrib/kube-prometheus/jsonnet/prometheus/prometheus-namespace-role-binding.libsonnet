local k = import "ksonnet.beta.3/k.libsonnet";
local roleBinding = k.rbac.v1.roleBinding;

{
  new(serviceAccountNamespace, namespace, name)::
    roleBinding.new() +
      roleBinding.mixin.metadata.withName(name) +
      roleBinding.mixin.metadata.withNamespace(namespace) +
      roleBinding.mixin.roleRef.withApiGroup("rbac.authorization.k8s.io") +
      roleBinding.mixin.roleRef.withName(name) +
      roleBinding.mixin.roleRef.mixinInstance({kind: "Role"}) +
      roleBinding.withSubjects([{kind: "ServiceAccount", name: name, namespace: serviceAccountNamespace}])
}
