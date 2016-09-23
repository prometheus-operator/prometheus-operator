package prometheus

import (
	"text/template"
)

var configTmpl = template.Must(template.New("config").Parse(`
{{ block "scrapeConfigs" . }}
scrape_configs:
- job_name: "prometheus"
  static_configs:
  - targets: ["localhost:9090"]
{{ end }}
`))
