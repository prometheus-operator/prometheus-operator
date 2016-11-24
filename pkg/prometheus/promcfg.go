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
	"fmt"

	yaml "gopkg.in/yaml.v1"

	"github.com/coreos/prometheus-operator/pkg/spec"
)

func generateConfig(p *spec.Prometheus, mons map[string]*spec.ServiceMonitor) ([]byte, error) {
	cfg := map[string]interface{}{}

	global := map[string]string{
		"evaluation_interval": "30s",
	}
	if p.Spec.EvaluationInterval != "" {
		global["evaluation_interval"] = p.Spec.EvaluationInterval
	}

	cfg["global"] = global
	cfg["rule_files"] = []string{"/etc/prometheus/rules/*.rules"}

	var scrapeConfigs []interface{}
	for _, mon := range mons {
		for i, ep := range mon.Spec.Endpoints {
			scrapeConfigs = append(scrapeConfigs, generateServiceMonitorConfig(mon, ep, i))
		}
	}

	cfg["scrape_configs"] = scrapeConfigs

	return yaml.Marshal(cfg)
}

func generateServiceMonitorConfig(m *spec.ServiceMonitor, ep spec.Endpoint, i int) interface{} {
	cfg := map[string]interface{}{
		"job_name": fmt.Sprintf("%s/%s/%d", m.Namespace, m.Name, i),
		"kubernetes_sd_configs": []map[string]interface{}{
			{
				"role": "endpoints",
			},
		},
	}

	if ep.Interval != "" {
		cfg["scrape_interval"] = ep.Interval
	}
	if ep.Path != "" {
		cfg["metrics_path"] = ep.Path
	}
	if ep.Scheme != "" {
		cfg["scheme"] = ep.Scheme
	}

	var relabelings []interface{}

	// Filter targets by services selected by the monitor.
	for k, v := range m.Spec.Selector.MatchLabels {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_service_label_" + k},
			"regex":         v,
		})
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_endpoint_port_name"},
			"regex":         ep.Port,
		})
	} else if ep.TargetPort.StrVal != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_container_port_name"},
			"regex":         ep.TargetPort.String(),
		})
	} else if ep.TargetPort.IntVal != 0 {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_container_port_number"},
			"regex":         ep.TargetPort.String(),
		})
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabelings = append(relabelings, []interface{}{
		map[string]interface{}{
			"source_labels": []string{"__meta_kubernetes_namespace"},
			"target_label":  "namespace",
		},
		map[string]interface{}{
			"action":      "labelmap",
			"regex":       "__meta_kubernetes_service_label_(.+)",
			"replacement": "svc_$1",
		},
		map[string]interface{}{
			"action":       "replace",
			"target_label": "__meta_kubernetes_pod_label_pod_template_hash",
			"replacement":  "",
		},
		map[string]interface{}{
			"action":      "labelmap",
			"regex":       "__meta_kubernetes_pod_label_(.+)",
			"replacement": "pod_$1",
		},
	}...)

	// By default, generate a safe job name from the service name and scraped port.
	// We also keep this around if a jobLabel is set in case the targets don't actually
	// have a value for it.
	if ep.Port != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"source_labels": []string{"__meta_kubernetes_service_name"},
			"target_label":  "job",
			"replacement":   "${1}-" + ep.Port,
		})
	} else if ep.TargetPort.String() != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"source_labels": []string{"__meta_kubernetes_service_name"},
			"target_label":  "job",
			"replacement":   "${1}-" + ep.TargetPort.String(),
		})
	}
	// Generate a job name with a base label. Same as above, just that we
	// get the base from the label, if present.
	if m.Spec.JobLabel != "" {
		if ep.Port != "" {
			relabelings = append(relabelings, map[string]interface{}{
				"source_labels": []string{"__meta_kubernetes_service_label_" + m.Spec.JobLabel},
				"target_label":  "job",
				"regex":         "(.+)",
				"replacement":   "${1}-" + ep.Port,
			})
		} else if ep.TargetPort.String() != "" {
			relabelings = append(relabelings, map[string]interface{}{
				"source_labels": []string{"__meta_kubernetes_service_label_" + m.Spec.JobLabel},
				"target_label":  "job",
				"regex":         "(.+)",
				"replacement":   "${1}-" + ep.TargetPort.String(),
			})
		}
	}

	cfg["relabel_configs"] = relabelings

	return cfg
}
