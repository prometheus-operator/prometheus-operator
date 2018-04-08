local k = import "ksonnet.beta.3/k.libsonnet";
local secret = k.core.v1.secret;

local plainConfig = "global:
  resolve_timeout: 5m
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'null'
  routes:
  - match:
      alertname: DeadMansSwitch
    receiver: 'null'
receivers:
- name: 'null'";

local config = std.base64(plainConfig);

{
    new(namespace)::
        secret.new("alertmanager-main", {"alertmanager.yaml": config}) +
          secret.mixin.metadata.withNamespace(namespace)
}
