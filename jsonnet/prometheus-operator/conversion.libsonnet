local defaults = {
  local defaults = self,
  name: error 'must provide the name of the webhook service',
  namespace: error 'must provide the namespace of the webhook service',
  annotations: {},
  port: 8443,
  path: '/convert',
  versions: ['v1beta1', 'v1alpha1'],
  caBundle: '',
};


function(params) {
  local c = self,
  _config:: defaults + params,

  metadata+: {
    annotations+: c._config.annotations,
  },
  spec+: {
    conversion: {
      strategy: 'Webhook',
      webhook: {
        conversionReviewVersions: c._config.versions,
        clientConfig: {
          service: {
            namespace: c._config.namespace,
            name: c._config.name,
            path: c._config.path,
            port: c._config.port,
          } + if c._config.caBundle != '' then
            { caBundle: c._config.caBundle }
          else
            {},
        },
      },
    },
  },
}
