// Copyright 2025 The prometheus-operator Authors
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

package validation

import (
	"testing"

	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestValidateRelabelConfig(t *testing.T) {
	defaultRegexp, err := relabel.DefaultRelabelConfig.Regex.MarshalYAML()
	if err != nil {
		t.Errorf("Could not marshal relabel.DefaultRelabelConfig.Regex: %v", err)
	}
	defaultRegex, ok := defaultRegexp.(string)
	if !ok {
		t.Errorf("Could not assert marshaled defaultRegexp as string: %v", defaultRegexp)
	}

	defaultSourceLabels := []monitoringv1.LabelName{}
	for _, label := range relabel.DefaultRelabelConfig.SourceLabels {
		defaultSourceLabels = append(defaultSourceLabels, monitoringv1.LabelName(label))
	}

	defaultPrometheusSpec := monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version: "v2.36.0",
			},
		},
	}

	for _, tc := range []struct {
		scenario      string
		relabelConfig monitoringv1.RelabelConfig
		prometheus    monitoringv1.Prometheus
		expectedErr   bool
	}{
		// Test invalid regex expression
		{
			scenario: "Invalid regex",
			relabelConfig: monitoringv1.RelabelConfig{
				Regex: "invalid regex)",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test invalid target label
		{
			scenario: "invalid target label",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "l\\${3}",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test empty target label for action replace
		{
			scenario: "empty target label for replace action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test empty target label for action hashmod
		{
			scenario: "empty target label for hashmod action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "hashmod",
				TargetLabel: "",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test empty target label for action uppercase
		{
			scenario: "empty target label for uppercase action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "uppercase",
				TargetLabel: "",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test empty target label for action lowercase
		{
			scenario: "empty target label for lowercase action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "lowercase",
				TargetLabel: "",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test replacement set for action uppercase
		{
			scenario: "replacement set for uppercase action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "uppercase",
				Replacement: ptr.To("some-replace-value"),
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test invalid hashmod relabel config
		{
			scenario: "invalid hashmod config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "hashmod",
				Modulus:      0,
				TargetLabel:  "__tmp_hashmod",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test invalid labelmap relabel config
		{
			scenario: "invalid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: ptr.To("some-name-value"),
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test valid labelmap relabel config when replacement not specified
		{
			scenario: "valid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labelmap",
				Regex:  "__meta_kubernetes_service_label_(.+)",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid labelmap relabel config with replacement specified
		{
			scenario: "valid labelmap config with replacement",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: ptr.To("abc"),
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test invalid labelkeep relabel config
		{
			scenario: "invalid labelkeep config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "labelkeep",
				TargetLabel:  "__tmp_labelkeep",
			},
			prometheus:  defaultPrometheusSpec,
			expectedErr: true,
		},
		// Test valid labelkeep relabel config
		{
			scenario: "valid labelkeep config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labelkeep",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid labeldrop relabel config
		{
			scenario: "valid labeldrop config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labeldrop",
				Regex:  "replica",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid replace config with empty replacement
		{
			scenario: "valid replace config with empty replacement",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				Replacement: ptr.To(""),
				TargetLabel: "abc",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid labeldrop relabel config with default value
		{
			scenario: "valid labeldrop config with default values",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labeldrop",
				Regex:  defaultRegex,
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid hashmod relabel config
		{
			scenario: "valid hashmod config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: defaultSourceLabels,
				Action:       "hashmod",
				Modulus:      1,
				TargetLabel:  "abc",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid replace relabel config
		{
			scenario: "valid replace config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "abc",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid uppercase relabel config
		{
			scenario: "valid uppercase config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "uppercase",
				TargetLabel: "ABC",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid lowercase relabel config
		{
			scenario: "valid lowercase config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "lowercase",
				TargetLabel: "abc",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test uppercase relabel config with lower prometheus version
		{
			scenario: "uppercase config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "uppercase",
				TargetLabel: "abc",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.30.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test lowercase relabel config with lower prometheus version
		{
			scenario: "lowercase config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "lowercase",
				TargetLabel: "abc",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.30.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test keepequal relabel config with lower prometheus version
		{
			scenario: "keepequal config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "keepequal",
				TargetLabel: "abc",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.30.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test dropequal relabel config with lower prometheus version
		{
			scenario: "dropequal config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "dropequal",
				TargetLabel: "abc",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.30.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test valid keepequal relabel config
		{
			scenario: "valid keepequal config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__tmp_port"},
				TargetLabel:  "__port1",
				Action:       "keepequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.41.0",
					},
				},
			},
		},
		// Test valid dropequal relabel config
		{
			scenario: "valid dropequal config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__tmp_port"},
				TargetLabel:  "__port1",
				Action:       "dropequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.41.0",
					},
				},
			},
		},
		// Test valid keepequal with non default values for other fields
		{
			scenario: "valid keepequal config with non default values for other fields",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__tmp_port"},
				TargetLabel:  "__port1",
				Separator:    ptr.To("^"),
				Regex:        "validregex",
				Replacement:  ptr.To("replacevalue"),
				Action:       "keepequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.41.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test valid keepequal with default values for other fields
		{
			scenario: "valid keepequal config with default values for other fields",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__tmp_port"},
				TargetLabel:  "__port1",
				Separator:    ptr.To(relabel.DefaultRelabelConfig.Separator),
				Regex:        relabel.DefaultRelabelConfig.Regex.String(),
				Modulus:      relabel.DefaultRelabelConfig.Modulus,
				Replacement:  &relabel.DefaultRelabelConfig.Replacement,
				Action:       "keepequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.41.0",
					},
				},
			},
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			lcv, err := NewLabelConfigValidator(&tc.prometheus)
			require.NoError(t, err)

			err = lcv.ValidateRelabelConfig(tc.relabelConfig)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
