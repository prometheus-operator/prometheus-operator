local po = (import 'prometheus-operator/prometheus-operator.libsonnet').prometheusOperator;
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local deployment = k.apps.v1.deployment;

po.deployment +
deployment.mixin.spec.template.spec.withServiceAccountName('')
