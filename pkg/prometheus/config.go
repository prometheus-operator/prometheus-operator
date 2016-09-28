package prometheus

import "text/template"

type TemplateConfig struct {
	ServiceMonitors map[string]ServiceMonitorObj
}

var configTmpl = template.Must(template.New("config").Parse(`
{{- block "scrapeConfigs" . -}}
scrape_configs:
{{- range $svc := .ServiceMonitors }}
{{- range $ep := $svc.Spec.Endpoints }}
- job_name: "{{ $svc.Name }}_{{ $ep.Port }}"
  scrape_interval: "{{ $svc.Spec.ScrapeInterval }}"
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
  {{- range $k, $v := $svc.Spec.Selector.MatchLabels }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_service_label_{{ $k }}"]
    regex: "{{ $v }}"
  {{- end }}{{ if false}}
  {{- if not eq $ep.Port.StrVal "" }}
  - action: "keep"
    source_labels: ["__meta_kubernetes_pod_port_name"]
    regex: "{{ $ep.Port.String }}"
  {{- else if not eq $ep.Port.IntVal 0 }}
  - source_labels: ["__address__"]
    regex: "(.+):[0-9]+"
    target_label: "__address__"
    replacement: "$1:{{ $ep.Port.String }}"
  {{- end }}{{ end }}
  - source_labels: [__meta_kubernetes_service_namespace]
    target_label: "namespace"
  - action: "labelmap"
    regex: "__meta_kubernetes_service_label_(.+)"
    replacement: "svc_$1"
  - action: "labelmap"
    regex: "__meta_kubernetes_pod_label_(.+)"
    replacement: "pod_$1"
  - source_labels: [__meta_kubernetes_service_name]
    target_label: "job"
    replacement: "$1_{{ $ep.Port }}"
{{- end }}
{{- end }}
{{- end }}
`))
