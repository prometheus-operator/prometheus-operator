#!/usr/bin/env python
import os
import re
from ruamel import yaml

def escape(s):
  return s.replace("{{","{{`{{").replace("}}","}}`}}")

def get_header(file_name):
  return "{{ define \"" + file_name + ".tpl\" }}\n"

#####
## Step 1 - Sync prometheus alert rules, create template file
####
charts = [
{'source':'contrib/kube-prometheus/assets/prometheus/rules/alertmanager.rules.yaml',
'destination': 'helm/alertmanager/', 'job_replace_by': '{{ template \"alertmanager.fullname\" .  }}'},
{'source': 'contrib/kube-prometheus/assets/prometheus/rules/kube-controller-manager.rules.yaml',
'destination': 'helm/exporter-kube-controller-manager/', 'job_replace_by': '{{ template \"exporter-kube-controller-manager.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/kube-scheduler.rules.yaml',
'destination': 'helm/exporter-kube-scheduler/', 'job_replace_by': '{{ template \"exporter-kube-scheduler.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/kube-state-metrics.rules.yaml',
'destination': 'helm/exporter-kube-state/', 'job_replace_by': '{{ template \"exporter-kube-state.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/node.rules.yaml',
'destination': 'helm/exporter-node/', 'job_replace_by': '{{ template \"exporter-node.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/prometheus.rules.yaml',
'destination': 'helm/prometheus/', 'job_replace_by': '{{ template \"prometheus.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/etcd3.rules.yaml',
'destination': 'helm/exporter-kube-etcd/', 'job_replace_by': '{{ template \"exporter-kube-etcd.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/general.rules.yaml',
'destination': 'helm/kube-prometheus/', 'job_replace_by': '{{ template \"kube-prometheus.fullname\" .  }}'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/kubelet.rules.yaml',
'destination': 'helm/exporter-kubelets/', 'job_replace_by': 'kubelet'},
{'source':'contrib/kube-prometheus/assets/prometheus/rules/kubernetes.rules.yaml',
'destination': 'helm/exporter-kubernetes/', 'job_replace_by': 'kubernetes'},
]

# read the rules, create a new template file
for chart in charts:

  _, name = os.path.split(chart['source'])
  lines = get_header(name)

  f = open(chart['source'], 'r')
  lines += escape(f.read())
  lines = re.sub("job=\"(.*?)\"", "job=\"" + chart['job_replace_by'] + "\"", lines) #replace the job name by chart variable

  lines += "{{ end }}" # footer

  new_f = "{}/templates/{}".format(chart['destination'], name)

  # recreate the file
  with open(new_f, 'w') as f:
      f.write(lines)

  print("Generated {}".format(new_f))
