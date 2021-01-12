local prometheusOperator = (import 'prometheus-operator/prometheus-operator.libsonnet');
local config = (import 'config.jsonnet');

local po = prometheusOperator(config);

po.deployment {
  spec+: {
    template+: {
      spec+: {
        serviceAccountName: '',
      },
    },
  },
}
