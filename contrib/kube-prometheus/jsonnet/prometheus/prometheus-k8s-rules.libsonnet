local k = import "ksonnet.beta.3/k.libsonnet";
local configMap = k.core.v1.configMap;

{
    new(namespace, ruleFiles)::
        configMap.new("prometheus-k8s-rules", ruleFiles) +
          configMap.mixin.metadata.withNamespace(namespace)
}
