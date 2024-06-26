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
	"context"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func newLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowWarn())
}

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
			scenario: "valid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: ptr.To("${2}"),
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
		// Test valid relabel config with empty replacement
		{
			scenario: "valid replace config with empty replacement",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "dummyTarget",
				Regex:       "replica",
				Replacement: ptr.To(""),
			},
			prometheus: defaultPrometheusSpec,
		},
		{
			scenario: "valid labeldrop config with default values",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: defaultSourceLabels,
				Separator:    ptr.To(relabel.DefaultRelabelConfig.Separator),
				TargetLabel:  relabel.DefaultRelabelConfig.TargetLabel,
				Regex:        defaultRegex,
				Modulus:      relabel.DefaultRelabelConfig.Modulus,
				Replacement:  &relabel.DefaultRelabelConfig.Replacement,
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
				Replacement:  ptr.To("$1:80"),
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

func TestSelectProbes(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.ProbeSpec)
		selected    bool
		scrapeClass *string
	}{
		{
			scenario: "url starting with http",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "http://blackbox-exporter.example.com"
			},
			selected: false,
		},
		{
			scenario: "url starting with https",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "https://blackbox-exporter.example.com"
			},
			selected: false,
		},
		{
			scenario: "url starting with ftp",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "ftp://fileserver.com"
			},
			selected: false,
		},
		{
			scenario: "ip address as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "192.168.178.3"
			},
			selected: true,
		},
		{
			scenario: "ip address:port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "192.168.178.3:9090"
			},
			selected: true,
		},
		{
			scenario: "dnsname as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "blackbox-exporter.example.com"
			},
			selected: true,
		},
		{
			scenario: "dnsname:port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "blackbox-exporter.example.com:8080"
			},
			selected: true,
		},
		{
			scenario: "hostname as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "localhost"
			},
			selected: true,
		},
		{
			scenario: "hostname starting with a digit as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "12-exporter.example.com"
			},
			selected: true,
		},
		{
			scenario: "invalid proxyurl",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyURL = "http://xxx-${dev}.svc.cluster.local:80"
			},
			selected: false,
		},
		{
			scenario: "valid proxyurl",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyURL = "123-proxy.example.com"
			},
			selected: true,
		},
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid metric relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
		},
		{
			scenario: "valid static relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid static relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "valid ingress relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid ingress relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario:    "inexistent scrape class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario:    "existent scrape class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			selected: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset()

			rs := NewResourceSelector(
				newLogger(),
				&monitoringv1.Prometheus{
					Spec: monitoringv1.PrometheusSpec{
						CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
							ScrapeClasses: []monitoringv1.ScrapeClass{
								{
									Name: "existent",
								},
							},
						},
					},
				},
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				record.NewFakeRecorder(1),
			)

			probe := &monitoringv1.Probe{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						URL: "example.com:80",
					},
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{},
					},
				},
			}

			if tc.scrapeClass != nil {
				probe.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&probe.Spec)

			probes, err := rs.SelectProbes(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(probe)
				return nil
			})

			require.NoError(t, err)
			if tc.selected {
				require.Len(t, probes, 1)
			} else {
				require.Empty(t, probes)
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
		t.Run(tc.scenario, func(t *testing.T) {
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

func TestSelectServiceMonitors(t *testing.T) {
	ca, err := os.ReadFile("../../test/e2e/remote_write_certs/ca.crt")
	require.NoError(t, err)

	cert, err := os.ReadFile("../../test/e2e/remote_write_certs/client.crt")
	require.NoError(t, err)

	key, err := os.ReadFile("../../test/e2e/remote_write_certs/client.key")
	require.NoError(t, err)

	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.ServiceMonitorSpec)
		selected    bool
		scrapeClass *string
	}{
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: true,
		},
		{
			scenario: "invalid metric relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "valid relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: true,
		},
		{
			scenario: "invalid relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "valid TLS config with CA, cert and key",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				})
			},
			selected: true,
		},
		{
			scenario: "invalid TLS config with both CA and CAFile",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						CAFile: "/etc/secrets/tls/ca.crt",
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid TLS config with both CA Secret and Configmap",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
								ConfigMap: &v1.ConfigMapKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "configmap",
									},
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid TLS config with invalid CA data",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid TLS config with cert and missing key",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid TLS config with key and missing cert",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid TLS config with key and invalid cert",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid proxyurl",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					ProxyURL: ptr.To("http://xxx-${dev}.svc.cluster.local:80"),
				})
			},
			selected: false,
		},
		{
			scenario: "valid proxyurl",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					ProxyURL: ptr.To("http://proxy.svc.cluster.local:80"),
				})
			},
			selected: true,
		},
		{
			scenario:    "inexistent Scrape Class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(_ *monitoringv1.ServiceMonitorSpec) {
			},
			selected: false,
		},
		{
			scenario:    "existent Scrape Class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(_ *monitoringv1.ServiceMonitorSpec) {
			},
			selected: true,
		},
		{
			scenario: "Mixed Endpoints",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"ca":         ca,
						"invalid_ca": []byte("garbage"),
						"cert":       cert,
						"key":        key,
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "test",
					},
					Data: map[string]string{
						"ca":   string(ca),
						"cert": string(cert),
					},
				},
			)

			rs := NewResourceSelector(
				newLogger(),
				&monitoringv1.Prometheus{
					Spec: monitoringv1.PrometheusSpec{
						CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
							ScrapeClasses: []monitoringv1.ScrapeClass{
								{
									Name: "existent",
								},
							},
						},
					},
				},
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				record.NewFakeRecorder(1),
			)

			sm := &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.ServiceMonitorSpec{},
			}

			if tc.scrapeClass != nil {
				sm.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&sm.Spec)

			sms, err := rs.SelectServiceMonitors(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(sm)
				return nil
			})

			require.NoError(t, err)
			if tc.selected {
				require.Len(t, sms, 1)
			} else {
				require.Empty(t, sms)
			}
		})
	}
}

func TestSelectPodMonitors(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.PodMonitorSpec)
		selected    bool
		scrapeClass *string
	}{
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: true,
		},
		{
			scenario: "invalid metric relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "valid relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: true,
		},
		{
			scenario: "invalid relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
		{
			scenario: "invalid proxyurl",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					ProxyURL: ptr.To("http://xxx-${dev}.svc.cluster.local:80"),
				})
			},
			selected: false,
		},
		{
			scenario: "valid proxyurl",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					ProxyURL: ptr.To("http://proxy.svc.cluster.local:80"),
				})
			},
			selected: true,
		},
		{
			scenario:    "Inexistent Scrape Class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(_ *monitoringv1.PodMonitorSpec) {
			},
			selected: false,
		},
		{
			scenario:    "existent Scrape Class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(_ *monitoringv1.PodMonitorSpec) {
			},
			selected: true,
		},
		{
			scenario: "Mixed Endpoints",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			selected: false,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			rs := NewResourceSelector(
				newLogger(),
				&monitoringv1.Prometheus{
					Spec: monitoringv1.PrometheusSpec{
						CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
							ScrapeClasses: []monitoringv1.ScrapeClass{
								{
									Name: "existent",
								},
							},
						},
					},
				},
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				record.NewFakeRecorder(1),
			)

			pm := &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}

			if tc.scrapeClass != nil {
				pm.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&pm.Spec)

			sms, err := rs.SelectPodMonitors(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(pm)
				return nil
			})

			require.NoError(t, err)

			if tc.selected {
				require.Len(t, sms, 1)
				return
			}

			require.Empty(t, sms)
		})
	}
}

func TestSelectScrapeConfigs(t *testing.T) {
	ca, err := os.ReadFile("../../test/e2e/remote_write_certs/ca.crt")
	require.NoError(t, err)

	cert, err := os.ReadFile("../../test/e2e/remote_write_certs/client.crt")
	require.NoError(t, err)

	key, err := os.ReadFile("../../test/e2e/remote_write_certs/client.key")
	require.NoError(t, err)
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1alpha1.ScrapeConfigSpec)
		selected    bool
		promVersion string
		scrapeClass *string
	}{
		{
			scenario: "valid relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid metric relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "valid proxy config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid proxy config with proxyConnectHeaders but no proxyUrl defined or proxyFromEnvironment set to true",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "invalid proxy config with proxy from environment set to true but proxyUrl defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "invalid proxy config with proxyFromEnvironment set to true but noProxy defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "invalid proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "invalid-key",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "invalid proxy config with noProxy defined and but no proxyUrl defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					NoProxy: ptr.To("0.0.0.0"),
				}
			},
			selected: false,
		},
		{
			scenario: "valid proxy config with muti header values",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "invalid proxy config with one invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "invalid-key",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "HTTP SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "HTTP SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "HTTP SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "HTTP SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "HTTP SD proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "invalid-key",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with invalid label selector",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Label: "app=example,env!=production,release in (v1, v2",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with invalid field selector",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Field: "status.phase=Running,metadata.name!=worker,)",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with valid Selector Role",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: "node",
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role: "node",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with invalid Selector Role",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: "node",
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role: "pod",
							},
						},
					},
				}

			},
			selected: false,
		},
		{
			scenario: "Kubernetes SD config with valid label and field selectors",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: "node",
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  "node",
								Label: "app=example,env!=production,release in (v1, v2)",
								Field: "status.phase=Running,metadata.name!=worker",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with only apiServer specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						APIServer: ptr.To("https://kube-api-server-address:6443"),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with only namespaces.ownNamespace specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							IncludeOwnNamespace: ptr.To(true),
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kubernetes SD config with both apiServer and namespaces.ownNamespace specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						APIServer: ptr.To("https://kube-api-server-address:6443"),
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							IncludeOwnNamespace: ptr.To(true),
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Consul SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Consul SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Consul SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Consul SD proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "foo",
										},
										Key: "invalid-key",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "DNS SD config with port for type other than SRV record",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To("A"),
						Port:  ptr.To(9100),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "DNS SD config with no port specified for type other than SRV record",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To("A"),
					},
				}
			},
			selected: false,
		},
		{
			scenario: "EC2 SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "EC2 SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "EC2 SD config with invalid secret ref for accessKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "EC2 SD config with invalid secret ref for secretKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Azure SD config with valid options for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						TenantID: ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID: ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Azure SD config with no client secret ref provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("OAuth"),
						TenantID:             ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID:             ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Azure SD config with no tenant id provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("OAuth"),
						ClientID:             ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Azure SD config with no client id provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("OAuth"),
						TenantID:             ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Azure SD config without options provided for ManagedIdentity authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("ManagedIdentity"),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Azure SD config without options provided for SDK authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("SDK"),
					},
				}
			},
			promVersion: "2.52.0",
			selected:    true,
		},
		{
			scenario: "Azure SD config with SDK authentication method but unsupported prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To("SDK"),
					},
				}
			},
			promVersion: "2.51.0",
			selected:    false,
		},
		{
			scenario: "OpenStack SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   "Instance",
						Region: "RegionOne",
						Password: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						ApplicationCredentialSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "OpenStack SD config with invalid secret ref for password",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Password: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "invalid",
							},
							Key: "key1",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "OpenStack SD config with invalid secret ref for application credentials",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						ApplicationCredentialSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key3",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "OpenStack SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   "hypervisor",
						Region: "RegionTwo",
					},
				}
			},
			selected: true,
		},
		{
			scenario: "DigitalOcean SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DigitalOceanSDConfigs = []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "DigitalOcean SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DigitalOceanSDConfigs = []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kuma SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kuma SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kuma SD config with invalid server",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "aaaaaa",
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Kuma SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Kuma SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Eureka SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Eureka SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Eureka SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Eureka SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},

		{
			scenario: "Docker SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Docker SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Docker SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Docker SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Linode SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Linode SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Linode SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Linode SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Hetzner SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Hetzner SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Hetzener SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Hetzner SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Hetzner SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Hetzner SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Nomad SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Nomad SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Nomad SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Nomad SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Dockerswarm SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Dockerswarm SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "Dockerswarm SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Dockerswarm SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "PuppetDB SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "PuppetDB SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "PuppetDB SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "PuppetDB SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "PuppetDB SD config with invalid URL",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "www.percent-off.com",
					},
				}
			},
			selected: false,
		},
		{
			scenario: "LightSail SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "LightSail SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "LightSail SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "LightSail SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "LightSail SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			selected: false,
		},

		{
			scenario: "LightSail SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: true,
		},
		{
			scenario: "LightSail SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "LightSail SD config with invalid secret ref for accessKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "LightSail SD config with invalid secret ref for secretKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key2",
						},
					},
				}
			},
			selected: false,
		},
		{
			scenario: "OVHCloud SD config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OVHCloudSDConfigs = []monitoringv1alpha1.OVHCloudSDConfig{
					{
						ApplicationKey: "ApplicationKey",
						ApplicationSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						ConsumerKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
						Service:  monitoringv1alpha1.VPS,
						Endpoint: ptr.To("127.0.0.1"),
					},
				}
			},
			selected: true,
		},
		{
			scenario: "Inexistent Scrape Class",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   "hypervisor",
						Region: "RegionTwo",
					},
				}
			},
			selected:    false,
			scrapeClass: ptr.To("inexistent"),
		},
		{
			scenario: "inexistent Scrape Class",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   "hypervisor",
						Region: "RegionTwo",
					},
				}
			},
			selected:    true,
			scrapeClass: ptr.To("existent"),
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1":       []byte("val1"),
						"key2":       []byte("val2"),
						"ca":         ca,
						"invalid_ca": []byte("garbage"),
						"cert":       cert,
						"key":        key,
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "test",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			)

			rs := NewResourceSelector(
				newLogger(),
				&monitoringv1.Prometheus{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: monitoringv1.PrometheusSpec{
						CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
							Version: tc.promVersion,
							ScrapeClasses: []monitoringv1.ScrapeClass{
								{
									Name: "existent",
								},
							},
						},
					},
				},
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				record.NewFakeRecorder(1),
			)

			sc := &monitoringv1alpha1.ScrapeConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}

			if tc.scrapeClass != nil {
				sc.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&sc.Spec)

			sms, err := rs.SelectScrapeConfigs(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(sc)
				return nil
			})

			require.NoError(t, err)
			if tc.selected {
				require.Len(t, sms, 1)
			} else {
				require.Empty(t, sms)
			}
		})
	}
}
