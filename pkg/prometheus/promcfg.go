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
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

func generateConfig(p *v1alpha1.Prometheus, mons map[string]*v1alpha1.ServiceMonitor) ([]byte, error) {
	cfg := map[string]interface{}{}

	cfg["global"] = map[string]string{
		"evaluation_interval": "30s",
		"scrape_interval":     "30s",
	}

	cfg["rule_files"] = []string{"/etc/prometheus/rules/*.rules"}

	var scrapeConfigs []interface{}
	for _, mon := range mons {
		for i, ep := range mon.Spec.Endpoints {
			scrapeConfigs = append(scrapeConfigs, generateServiceMonitorConfig(mon, ep, i))
		}
	}
	var alertmanagerConfigs []interface{}
	for _, am := range p.Spec.Alerting.Alertmanagers {
		alertmanagerConfigs = append(alertmanagerConfigs, generateAlertmanagerConfig(am))
	}

	cfg["scrape_configs"] = scrapeConfigs
	cfg["alerting"] = map[string]interface{}{
		"alertmanagers": alertmanagerConfigs,
	}

	return yaml.Marshal(cfg)
}

func generateServiceMonitorConfig(m *v1alpha1.ServiceMonitor, ep v1alpha1.Endpoint, i int) interface{} {
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
	if ep.TLSConfig != nil {
		tlsConfig := map[string]interface{}{
			"insecure_skip_verify": ep.TLSConfig.InsecureSkipVerify,
		}
		if ep.TLSConfig.CAFile != "" {
			tlsConfig["ca_file"] = ep.TLSConfig.CAFile
		}
		if ep.TLSConfig.CertFile != "" {
			tlsConfig["cert_file"] = ep.TLSConfig.CertFile
		}
		if ep.TLSConfig.KeyFile != "" {
			tlsConfig["key_file"] = ep.TLSConfig.KeyFile
		}
		if ep.TLSConfig.ServerName != "" {
			tlsConfig["server_name"] = ep.TLSConfig.ServerName
		}
		cfg["tls_config"] = tlsConfig
	}
	if ep.BearerTokenFile != "" {
		cfg["bearer_token_file"] = ep.BearerTokenFile
	}

	var relabelings []interface{}

	// Filter targets by services selected by the monitor.

	// Exact label matches.
	for k, v := range m.Spec.Selector.MatchLabels {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(k)},
			"regex":         v,
		})
	}
	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabelings = append(relabelings, map[string]interface{}{
				"action":        "keep",
				"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)},
				"regex":         strings.Join(exp.Values, "|"),
			})
		case metav1.LabelSelectorOpNotIn:
			relabelings = append(relabelings, map[string]interface{}{
				"action":        "drop",
				"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)},
				"regex":         strings.Join(exp.Values, "|"),
			})
		case metav1.LabelSelectorOpExists:
			relabelings = append(relabelings, map[string]interface{}{
				"action":        "keep",
				"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)},
				"regex":         ".+",
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabelings = append(relabelings, map[string]interface{}{
				"action":        "drop",
				"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)},
				"regex":         ".+",
			})
		}
	}

	// Filter targets based on the namespace selection configuration.
	// By default we only discover services within the namespace of the
	// ServiceMonitor.
	// Selections allow extending this to all namespaces or to a subset
	// of them specified by label or name matching.
	//
	// Label selections are not supported yet as they require either supported
	// in the upstream SD integration or require out-of-band implementation
	// in the operator with configuration reload.
	//
	// There's no explicit nil for the selector, we decide for the default
	// case if it's all zero values.
	nsel := m.Spec.NamespaceSelector

	if !nsel.Any && len(nsel.MatchNames) == 0 {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_namespace"},
			"regex":         m.Namespace,
		})
	} else if len(nsel.MatchNames) > 0 {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_namespace"},
			"regex":         strings.Join(nsel.MatchNames, "|"),
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

	// By default, generate a safe job name from the service name.  We also keep
	// this around if a jobLabel is set in case the targets don't actually have a
	// value for it. A single service may potentially have multiple metrics
	// endpoints, therefore the endpoints labels is filled with the ports name or
	// as a fallback the port number.

	relabelings = append(relabelings, map[string]interface{}{
		"source_labels": []string{"__meta_kubernetes_service_name"},
		"target_label":  "job",
		"replacement":   "${1}",
	})
	if m.Spec.JobLabel != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"source_labels": []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(m.Spec.JobLabel)},
			"target_label":  "job",
			"regex":         "(.+)",
			"replacement":   "${1}",
		})
	}

	if ep.Port != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"target_label": "endpoint",
			"replacement":  ep.Port,
		})
	} else if ep.TargetPort.String() != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"target_label": "endpoint",
			"replacement":  ep.TargetPort.String(),
		})
	}

	cfg["relabel_configs"] = relabelings

	return cfg
}

func generateAlertmanagerConfig(am v1alpha1.AlertmanagerEndpoints) interface{} {
	if am.Scheme == "" {
		am.Scheme = "http"
	}
	cfg := map[string]interface{}{
		"kubernetes_sd_configs": []map[string]interface{}{
			{
				"role": "endpoints",
			},
		},
		"scheme": am.Scheme,
	}

	var relabelings []interface{}

	relabelings = append(relabelings, map[string]interface{}{
		"action":        "keep",
		"source_labels": []string{"__meta_kubernetes_service_name"},
		"regex":         am.Name,
	})
	relabelings = append(relabelings, map[string]interface{}{
		"action":        "keep",
		"source_labels": []string{"__meta_kubernetes_namespace"},
		"regex":         am.Namespace,
	})

	if am.Port.StrVal != "" {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_endpoint_port_name"},
			"regex":         am.Port.String(),
		})
	} else if am.Port.IntVal != 0 {
		relabelings = append(relabelings, map[string]interface{}{
			"action":        "keep",
			"source_labels": []string{"__meta_kubernetes_container_port_number"},
			"regex":         am.Port.String(),
		})
	}

	cfg["relabel_configs"] = relabelings

	return cfg
}
