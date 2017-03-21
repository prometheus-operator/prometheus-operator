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

	promconfig "github.com/prometheus/prometheus/config"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

func TestConfigGenerationNonNamespacedAnnotation(t *testing.T) {
	p := &v1alpha1.Prometheus{}
	smons := map[string]*v1alpha1.ServiceMonitor{
		"1": &v1alpha1.ServiceMonitor{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.ServiceMonitorSpec{
				Selector: unversioned.LabelSelector{
					MatchExpressions: []unversioned.LabelSelectorRequirement{
						unversioned.LabelSelectorRequirement{
							Key:      "k8s-app",
							Operator: unversioned.LabelSelectorOpExists,
						},
					},
				},
				Endpoints: []v1alpha1.Endpoint{
					v1alpha1.Endpoint{Port: "web"},
				},
			},
		},
	}

	config, err := generateConfig(p, smons, 0)
	if err != nil {
		t.Fatal("Config generation failed: ", err)
	}
	generatedConfig := string(config)

	cfg, err := promconfig.Load(generatedConfig)
	if err != nil {
		t.Fatalf("Generated config cannot be parsed by Prometheus. Config:\n\n%s\n\nError: %s", generatedConfig, err)
	}

	success := false
	for _, c := range cfg.ScrapeConfigs[0].RelabelConfigs {
		for _, label := range c.SourceLabels {
			if label == "__meta_kubernetes_service_annotation_alpha_monitoring_coreos_com_non_namespaced" {
				success = true
			}
		}
	}

	if !success {
		t.Fatal("No action on the `non-namespaced` annotation taken during relabelling")
	}
}
