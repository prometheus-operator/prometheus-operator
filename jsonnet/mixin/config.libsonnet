{
  _config+:: {
    prometheusOperatorSelector: 'job="prometheus-operator"',
    configReloaderSelector: 'namespace=~".+"',
    groupLabels: 'controller,namespace',
  },
}
