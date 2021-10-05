module github.com/prometheus-operator/prometheus-operator/pkg/client

go 1.14

require (
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.51.2
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
)

replace github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => ../apis/monitoring
