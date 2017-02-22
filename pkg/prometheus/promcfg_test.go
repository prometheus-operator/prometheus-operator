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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

func TestConfigGeneration(t *testing.T) {
	p := &v1alpha1.Prometheus{}
	smons := map[string]*v1alpha1.ServiceMonitor{
		"1": &v1alpha1.ServiceMonitor{
			Spec: v1alpha1.ServiceMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						metav1.LabelSelectorRequirement{
							Key:      "k8s-app",
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				Endpoints: []v1alpha1.Endpoint{
					v1alpha1.Endpoint{Port: "web"},
				},
			},
		},
	}

	expectedConfig := `alerting:
  alertmanagers: []
global:
  evaluation_interval: 30s
  scrape_interval: 30s
rule_files:
- /etc/prometheus/rules/*.rules
scrape_configs:
- job_name: //0
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - action: keep
    regex: .+
    source_labels:
    - __meta_kubernetes_service_label_k8s_app
  - action: keep
    regex: ""
    source_labels:
    - __meta_kubernetes_namespace
  - action: keep
    regex: web
    source_labels:
    - __meta_kubernetes_endpoint_port_name
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - regex: "true"
    replacement: ""
    source_labels:
    - __meta_kubernetes_service_annotation_alpha_monitoring_coreos_com_non_namespaced
    target_label: namespace
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
    replacement: svc_$1
  - action: replace
    replacement: ""
    target_label: __meta_kubernetes_pod_label_pod_template_hash
  - action: labelmap
    regex: __meta_kubernetes_pod_label_(.+)
    replacement: pod_$1
  - replacement: ${1}
    source_labels:
    - __meta_kubernetes_service_name
    target_label: job
  - replacement: web
    target_label: endpoint
`

	config, err := generateConfig(p, smons)
	if err != nil {
		t.Fatal("Config generation failed: ", err)
	}
	generatedConfig := string(config)

	if generatedConfig != expectedConfig {
		t.Fatalf("Config was not generated as expected.\n\nExpected:\n\n%s\n\nGot:\n\n%s", expectedConfig, generatedConfig)
	}
}
