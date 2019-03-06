// On managed Kubernetes clusters some of the control plane components are not exposed to customers.
// Disable scrape jobs and service monitors for these components by overwriting 'kube-prometheus.libsonnet' defaults
// Note this doesn't disable generation of associated alerting rules but the rules don't trigger

{
  _config+:: {
    // This snippet walks the original object (super.jobs, set as temp var j) and creates a replacement jobs object
    //     excluding any members of the set specified (eg: controller and scheduler).
    local j = super.jobs,
    jobs: {
      [k]: j[k]
      for k in std.objectFields(j)
      if !std.setMember(k, ['KubeControllerManager', 'KubeScheduler'])
    },
  },

  // Same as above but for ServiceMonitor's
  local p = super.prometheus,
  prometheus: {
    [q]: p[q]
    for q in std.objectFields(p)
    if !std.setMember(q, ['serviceMonitorKubeControllerManager', 'serviceMonitorKubeScheduler'])
  },

  // TODO: disable generationg of alerting rules
  // manifests/prometheus-rules.yaml:52:  - name: kube-scheduler.rules

}
