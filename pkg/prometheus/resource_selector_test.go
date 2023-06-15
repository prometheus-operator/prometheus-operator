// Copyright 2023 The prometheus-operator Authors
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
	"testing"

	"github.com/prometheus/prometheus/model/relabel"

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
				Replacement: "some-replace-value",
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
				Replacement: "some-name-value",
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
			scenario: "valid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: "${2}",
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
		{
			scenario: "valid labeldrop config with default values",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: defaultSourceLabels,
				Separator:    relabel.DefaultRelabelConfig.Separator,
				TargetLabel:  relabel.DefaultRelabelConfig.TargetLabel,
				Regex:        defaultRegex,
				Modulus:      relabel.DefaultRelabelConfig.Modulus,
				Replacement:  relabel.DefaultRelabelConfig.Replacement,
				Action:       "labeldrop",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid hashmod relabel config
		{
			scenario: "valid hashmod config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "hashmod",
				Modulus:      10,
				TargetLabel:  "__tmp_hashmod",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid replace relabel config
		{
			scenario: "valid replace config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__address__"},
				Action:       "replace",
				Regex:        "([^:]+)(?::\\d+)?",
				Replacement:  "$1:80",
				TargetLabel:  "__address__",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid uppercase relabel config
		{
			scenario: "valid uppercase config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"foo"},
				Action:       "uppercase",
				TargetLabel:  "foo_uppercase",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test valid lowercase relabel config
		{
			scenario: "valid lowercase config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"bar"},
				Action:       "lowercase",
				TargetLabel:  "bar_lowercase",
			},
			prometheus: defaultPrometheusSpec,
		},
		// Test uppercase relabel config but lower prometheus version
		{
			scenario: "uppercase config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"foo"},
				Action:       "uppercase",
				TargetLabel:  "foo_uppercase",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.35.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test lowercase relabel config but lower prometheus version
		{
			scenario: "lowercase config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"bar"},
				Action:       "lowercase",
				TargetLabel:  "bar_lowercase",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.35.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test keepequal relabel config but lower prometheus version
		{
			scenario: "keepequal config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"foo"},
				Action:       "keepequal",
				TargetLabel:  "foo_keepequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.37.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test dropequal relabel config but lower prometheus version
		{
			scenario: "dropequal config with lower prometheus version",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"bar"},
				Action:       "keepequal",
				TargetLabel:  "bar_keepequal",
			},
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v2.37.0",
					},
				},
			},
			expectedErr: true,
		},
		// Test valid keepequal config
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
		// Test valid dropequal config
		{
			scenario: "valid dropequal config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__tmp_port"},
				TargetLabel:  "__port2",
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
				Separator:    "^",
				Regex:        "validregex",
				Replacement:  "replacevalue",
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
				Separator:    relabel.DefaultRelabelConfig.Separator,
				Regex:        relabel.DefaultRelabelConfig.Regex.String(),
				Modulus:      relabel.DefaultRelabelConfig.Modulus,
				Replacement:  relabel.DefaultRelabelConfig.Replacement,
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
		t.Run(fmt.Sprintf("case %s", tc.scenario), func(t *testing.T) {
			err := validateRelabelConfig(&tc.prometheus, tc.relabelConfig)
			if err != nil && !tc.expectedErr {
				t.Fatalf("expected no error, got: %v", err)
			}
			if err == nil && tc.expectedErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestValidateProberUrl(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		proberSpec  monitoringv1.ProberSpec
		expectedErr bool
	}{
		{
			scenario: "url starting with http",
			proberSpec: monitoringv1.ProberSpec{
				URL: "http://blackbox-exporter.example.com",
			},
			expectedErr: true,
		},
		{
			scenario: "url starting with https",
			proberSpec: monitoringv1.ProberSpec{
				URL: "https://blackbox-exporter.example.com",
			},
			expectedErr: true,
		},
		{
			scenario: "url starting with ftp",
			proberSpec: monitoringv1.ProberSpec{
				URL: "ftp://fileserver.com",
			},
			expectedErr: true,
		},
		{
			scenario: "ip address as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "192.168.178.3",
			},
		},
		{
			scenario: "ip address:port as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "192.168.178.3:9090",
			},
		},
		{
			scenario: "dnsname as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "blackbox-exporter.example.com",
			},
		},
		{
			scenario: "dnsname:port as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "blackbox-exporter.example.com:8080",
			},
		},
		{
			scenario: "hostname as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "localhost",
			},
		},
		{
			scenario: "hostname starting with a digit as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "12-exporter.example.com",
			},
		},
	} {
		t.Run(fmt.Sprintf("case %s %s", tc.scenario, tc.proberSpec.URL), func(t *testing.T) {
			err := validateProberURL(tc.proberSpec.URL)
			if err != nil && !tc.expectedErr {
				t.Fatalf("expected no error, got: %v", err)
			}
			if err == nil && tc.expectedErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestValidateScrapeIntervalAndTimeout(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		prometheus  monitoringv1.Prometheus
		smSpec      monitoringv1.ServiceMonitorSpec
		expectedErr bool
	}{
		{
			scenario: "scrape interval and timeout specified at service monitor spec but invalid #1",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "30s",
						ScrapeTimeout: "45s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "scrape interval and timeout specified at service monitor spec but invalid #2",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "30 s",
						ScrapeTimeout: "10s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "scrape interval and timeout specified at service monitor spec are valid",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "15s",
						ScrapeTimeout: "10s",
					},
				},
			},
		},
		{
			scenario: "only scrape timeout specified at service monitor spec but invalid compared to global scrapeInterval",
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ScrapeInterval: "15s",
					},
				},
			},
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						ScrapeTimeout: "20s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "only scrape timeout specified at service monitor spec but invalid compared to default global scrapeInterval",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						ScrapeTimeout: "60s",
					},
				},
			},
			expectedErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.scenario), func(t *testing.T) {
			for _, endpoint := range tc.smSpec.Endpoints {
				err := validateScrapeIntervalAndTimeout(&tc.prometheus, endpoint.Interval, endpoint.ScrapeTimeout)
				t.Logf("err %v", err)
				if err != nil && !tc.expectedErr {
					t.Fatalf("expected no error, got: %v", err)
				}
				if err == nil && tc.expectedErr {
					t.Fatalf("expected an error, got nil")
				}
			}
		})
	}
}
