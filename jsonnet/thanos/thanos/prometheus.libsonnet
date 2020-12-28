{
	"apiVersion": "monitoring.coreos.com/v1",
	"kind": "Prometheus",
	"metadata": {
		"labels": {
			"prometheus": "self"
		},
		"name": "self",
		"namespace": "default"
	},
	"spec": {
		"replicas": "2",
		"serviceMonitorSelector": {
			"matchLabels": {
				"app": "prometheus"
			}
		},
		"ruleSelector": {
			"matchLabels": {
				"role": "prometheus-rulefiles",
				"prometheus": "k8s"
			}
		},
		"thanos": {
			"version": "v0.11.0"
		}
	}
}