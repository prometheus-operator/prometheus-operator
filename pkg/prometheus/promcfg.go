// Copyright 2016 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"html/template"

	"github.com/coreos/prometheus-operator/pkg/spec"
)

type templateConfig struct {
	Prometheus      spec.PrometheusSpec
	ServiceMonitors map[string]*spec.ServiceMonitor
}

var configTmpl = template.Must(template.New("config").Parse(`
{{- block "globals" . }}
global:
  {{- if ne .Prometheus.EvaluationInterval "" }}
  evaluation_interval: {{ .Prometheus.EvaluationInterval }}
  {{- else }}
  evaluation_interval: 30s
  {{- end }}
{{- end}}

rule_files:
- /etc/prometheus/rules/*.rules

{{ block "scrapeConfigs" . }}
scrape_configs:
{{- range $mon := .ServiceMonitors }}
{{- range $i, $ep := $mon.Spec.Endpoints }}
- job_name: "{{ $mon.Namespace }}/{{ $mon.Name }}/{{ $i }}"

  {{- if ne $ep.Interval "" }}
  scrape_interval: "{{ $ep.Interval }}"
  {{- end }}
  {{- if ne $ep.Path "" }}
  metrics_path: "{{ $ep.Path }}"
  {{- end }}
  {{- if ne $ep.Scheme "" }}
  scheme: "{{ $ep.Scheme }}"
  {{- end }}

  kubernetes_sd_configs:
  - role: endpoints

  relabel_configs:
  # 
  # FILTERING
  #
  {{- range $k, $v := $mon.Spec.Selector.MatchLabels }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_service_label_{{ $k }}"]
    regex: "{{ $v }}"
  {{- end }}
  {{- if ne $ep.Port "" }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_endpoint_port_name"]
    regex: "{{ $ep.Port }}"
  {{- else if ne $ep.TargetPort.StrVal "" }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_pod_container_port_name"]
    regex: "{{ $ep.TargetPort.String }}"
  {{- else if ne $ep.TargetPort.IntVal 0 }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_pod_container_port_number"]
    regex: "{{ $ep.TargetPort.String }}"
  {{- end }}
  # 
  # TARGET LABELS
  #
  - source_labels: ["__meta_kubernetes_namespace"]
    target_label: "namespace"
  - action: "labelmap"
    regex: "__meta_kubernetes_service_label_(.+)"
    replacement: "svc_$1"
  - # Drop 'pod_template_hash' label that's set by deployment controller.
    action: replace
    target_label: "__meta_kubernetes_pod_label_pod_template_hash"
    replacement: ""
  - action: "labelmap"
    regex: "__meta_kubernetes_pod_label_(.+)"
    replacement: "pod_$1"
  # 
  # JOB LABEL
  #
  {{- if ne $ep.Port "" }}
  - source_labels: ["__meta_kubernetes_service_name"]
    target_label: "job"
    replacement: "${1}-{{ $ep.Port }}"
  {{- else if ne $ep.TargetPort.String "" }}
  - source_labels: ["__meta_kubernetes_service_name"]
    target_label: "job"
    replacement: "${1}-{{ $ep.TargetPort.String }}"
  {{- end}}
{{- end }}
{{- end }}
{{- end }}
`))
