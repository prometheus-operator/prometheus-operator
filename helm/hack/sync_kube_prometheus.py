#!/usr/bin/env python
#####
## This script read the kube-prometheus rules and convert into helm charts format
####
### ----------------------------
###  Sync all prometheus rules
###
charts = [{'file_name': 'alertmanager', 'search_var': 'ruleFiles',
'source':'contrib/kube-prometheus/assets/prometheus/rules/alertmanager.rules.yaml',
'destination': 'helm/alertmanager/values.yaml'},
{'file_name': 'kube-controller-manager', 'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/kube-controller-manager.rules.yaml',
'destination': 'helm/exporter-kube-controller-manager/values.yaml'},
{'file_name': 'kube-scheduler', 'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/kube-scheduler.rules.yaml',
'destination': 'helm/exporter-kube-scheduler/values.yaml'},
{'file_name': 'kube-state-metrics',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/kube-state-metrics.rules.yaml',
'destination': 'helm/exporter-kube-state/values.yaml'},
{'file_name': 'node',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/node.rules.yaml',
'destination': 'helm/exporter-node/values.yaml'},
{'file_name': 'prometheus',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/prometheus.rules.yaml', 
'destination': 'helm/prometheus/values.yaml'},
{'file_name': 'etcd3',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/etcd3.rules.yaml', 
'destination': 'helm/exporter-kube-etcd/values.yaml'},
# //TODO add {'file_name': 'general',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/general.rules.yaml', 
# 'destination': 'helm/kube-prometheus/general_rules.yaml'},
{'file_name': 'kubelet',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/kubelet.rules.yaml', 
'destination': 'helm/exporter-kubelets/values.yaml'},
{'file_name': 'kubernetes',  'search_var': 'ruleFiles', 'source':'contrib/kube-prometheus/assets/prometheus/rules/kubernetes.rules.yaml', 
'destination': 'helm/exporter-kubernetes/values.yaml'},
### 
###  Sync grafana dashboards
###
{'file_name': 'grafana-dashboards-0', 'search_var': 'serverDashboardFiles', 'source':'contrib/kube-prometheus/manifests/grafana/grafana-dashboards.yaml', 
'destination': 'helm/grafana/values.yaml'},
]

for chart in charts:
  lines = ""
  ## parse current values.yaml file
  f = open(chart['destination'], 'r')
  for l in f.readlines():

    # stop reading file after the rule
    if "{}:".format(chart['search_var']) in l:
        break
    lines+= l

  lines+= "{}:\n".format(chart['search_var'])
  lines+= "  {}.rules: |-\n".format(chart['file_name'])


  ## parse kube-prometheus rule
  f =  open(chart['source'])
  for l in f.readlines():
    lines += "    {}".format(l)

  # recreate the file  
  with open(chart['destination'], 'w') as f:
      f.write(lines)


### ----------------------------
### 2 
###