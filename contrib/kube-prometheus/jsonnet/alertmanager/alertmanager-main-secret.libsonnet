local k = import "ksonnet.beta.3/k.libsonnet";
local secret = k.core.v1.secret;

{
    new(namespace, plainConfig)::
        secret.new("alertmanager-main", {"alertmanager.yaml": std.base64(plainConfig)}) +
          secret.mixin.metadata.withNamespace(namespace)
}
