local po = (import 'prometheus-operator/prometheus-operator.libsonnet').prometheusOperator;
local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local deployment = k.apps.v1beta2.deployment;

po.deployment +
deployment.mixin.spec.template.spec.withServiceAccountName('')
