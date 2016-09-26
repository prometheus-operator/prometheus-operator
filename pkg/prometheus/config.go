package prometheus

import "text/template"

type TemplateConfig struct {
	ServiceMonitors []ServiceMonitorSpec
}

var configTmpl = template.Must(template.New("config").Parse(`
{{- block "scrapeConfigs" . -}}
scrape_configs:
{{- range _, $svc := range .ServiceMonitors }}
{{- range _, $ep := range $svc.Endpoints }}
- job_name: {{ $svc.Service }}
  scrape_interval: 10s
  metrics_path: "{{ $ep.Path }}"
  kubernetes_sd_configs:
  - in_cluster: true
    api_servers:
    - "https://kubernetes"
    role: endpoint
  tls_config:
    ca_file: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
  bearer_token_file: "/var/run/secrets/kubernetes.io/serviceaccount/token"
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name]
    action: keep
    regex: "{{ $svc.Service }}"
  - source_labels: [__meta_kubernetes_service_namespace]
    target_label: "namespace"
  - action: "labelmap"
    regex: "__meta_kubernetes_service_label_(.+)"
    replacement: "svc_$1"
{{ end -}}
{{ end -}}
{{ end -}}
`))
