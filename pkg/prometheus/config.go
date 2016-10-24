package prometheus

import "html/template"

type TemplateConfig struct {
	ServiceMonitors map[string]ServiceMonitorObj
}

var configTmpl = template.Must(template.New("config").Parse(`
{{- block "scrapeConfigs" . -}}
scrape_configs:
{{- range $mon := .ServiceMonitors }}
{{- range $i, $ep := $mon.Spec.Endpoints }}
- job_name: "{{ $mon.Name }}-{{ $i }}"

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
