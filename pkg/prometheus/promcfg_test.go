// Copyright 2017 The prometheus-operator Authors
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
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-openapi/swag"
	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func mustNewConfigGenerator(t *testing.T, p *monitoringv1.Prometheus) *ConfigGenerator {
	t.Helper()

	if p == nil {
		p = &monitoringv1.Prometheus{}
	}

	logger := level.NewFilter(log.NewLogfmtLogger(os.Stderr), level.AllowWarn())

	cg, err := NewConfigGenerator(log.With(logger, "test", t.Name()), p, false)
	if err != nil {
		t.Fatalf("failed to create config generator: %v", err)
	}

	return cg
}

func TestConfigGeneration(t *testing.T) {
	for i := range operator.PrometheusCompatibilityMatrix {
		v := operator.PrometheusCompatibilityMatrix[i]
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			cfg, err := generateTestConfig(t, v)
			if err != nil {
				t.Fatal(err)
			}

			reps := 1000
			if testing.Short() {
				reps = 100
			}

			for i := 0; i < reps; i++ {
				testcfg, err := generateTestConfig(t, v)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(cfg, testcfg) {
					t.Fatalf("Config generation is not deterministic.\n\n\nFirst generation: \n\n%s\n\nDifferent generation: \n\n%s\n\n", string(cfg), string(testcfg))
				}
			}
		})
	}
}

func TestGlobalSettings(t *testing.T) {
	for _, tc := range []struct {
		Scenario           string
		EvaluationInterval monitoringv1.Duration
		ScrapeInterval     monitoringv1.Duration
		ScrapeTimeout      monitoringv1.Duration
		ExternalLabels     map[string]string
		QueryLogFile       string
		Version            string
		Expected           string
		ExpectError        bool
	}{
		{
			Scenario:           "valid config",
			Version:            "v2.15.2",
			ScrapeInterval:     "15s",
			EvaluationInterval: "30s",
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 15s
  external_labels:
    prometheus: /
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			Scenario:       "invalid scrape timeout specified when scrape interval specified",
			Version:        "v2.30.0",
			ScrapeInterval: "30s",
			ScrapeTimeout:  "60s",
			ExpectError:    true,
		},
		{
			Scenario:           "valid scrape timeout along with valid scrape interval specified",
			Version:            "v2.15.2",
			ScrapeInterval:     "60s",
			ScrapeTimeout:      "10s",
			EvaluationInterval: "30s",
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 60s
  scrape_timeout: 10s
  external_labels:
    prometheus: /
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			Scenario:           "external label specified",
			Version:            "v2.15.2",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			ExternalLabels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    key1: value1
    key2: value2
    prometheus: /
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			Scenario:           "query log file",
			Version:            "v2.16.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			QueryLogFile:       "test.log",
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: /
    prometheus_replica: $(POD_NAME)
  query_log_file: /var/log/prometheus/test.log
scrape_configs: []
`,
		},
	} {

		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ScrapeInterval: tc.ScrapeInterval,
					ScrapeTimeout:  tc.ScrapeTimeout,
					ExternalLabels: tc.ExternalLabels,
					Version:        tc.Version,
				},
				EvaluationInterval: tc.EvaluationInterval,
				QueryLogFile:       tc.QueryLogFile,
			},
		}

		cg := mustNewConfigGenerator(t, p)
		t.Run(fmt.Sprintf("case %s", tc.Scenario), func(t *testing.T) {
			cfg, err := cg.GenerateServerConfiguration(
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)

			if err != nil && !tc.ExpectError {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tc.ExpectError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			result := string(cfg)
			if tc.Expected != string(cfg) {
				t.Log(pretty.Compare(tc.Expected, result))
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestNamespaceSetCorrectly(t *testing.T) {
	type testCase struct {
		ServiceMonitor           *monitoringv1.ServiceMonitor
		IgnoreNamespaceSelectors bool
		Expected                 string
	}

	testcases := []testCase{
		// Test that namespaces from 'MatchNames' are returned instead of the current namespace
		{
			ServiceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					NamespaceSelector: monitoringv1.NamespaceSelector{
						MatchNames: []string{"test1", "test2"},
					},
					AttachMetadata: &monitoringv1.AttachMetadata{
						Node: true,
					},
				},
			},
			IgnoreNamespaceSelectors: false,
			Expected: `kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test1
    - test2
  attach_metadata:
    node: true
`,
		},
		// Test that 'Any' returns an empty list instead of the current namespace
		{
			ServiceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor2",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					NamespaceSelector: monitoringv1.NamespaceSelector{
						Any: true,
					},
				},
			},
			IgnoreNamespaceSelectors: false,
			Expected: `kubernetes_sd_configs:
- role: endpoints
`,
		},
		// Test that Any takes precedence over MatchNames
		{
			ServiceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor2",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					NamespaceSelector: monitoringv1.NamespaceSelector{
						Any:        true,
						MatchNames: []string{"foo", "bar"},
					},
				},
			},
			IgnoreNamespaceSelectors: false,
			Expected: `kubernetes_sd_configs:
- role: endpoints
`,
		},
		// Test that IgnoreNamespaceSelectors overrides Any and MatchNames
		{
			ServiceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor2",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					NamespaceSelector: monitoringv1.NamespaceSelector{
						Any:        true,
						MatchNames: []string{"foo", "bar"},
					},
				},
			},
			IgnoreNamespaceSelectors: true,
			Expected: `kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - default
`,
		},
	}

	for _, tc := range testcases {
		cg := mustNewConfigGenerator(
			t,
			&monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						IgnoreNamespaceSelectors: tc.IgnoreNamespaceSelectors,
					},
				},
			},
		)

		var attachMetaConfig *attachMetadataConfig
		if tc.ServiceMonitor.Spec.AttachMetadata != nil {
			attachMetaConfig = &attachMetadataConfig{
				MinimumVersion: "2.37.0",
				AttachMetadata: tc.ServiceMonitor.Spec.AttachMetadata,
			}
		}

		c := cg.generateK8SSDConfig(tc.ServiceMonitor.Spec.NamespaceSelector, tc.ServiceMonitor.Namespace, nil, nil, kubernetesSDRoleEndpoint, attachMetaConfig)
		s, err := yaml.Marshal(yaml.MapSlice{c})
		if err != nil {
			t.Fatal(err)
		}

		if tc.Expected != string(s) {
			t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", string(s), tc.Expected)
		}
	}
}

func TestNamespaceSetCorrectlyForPodMonitor(t *testing.T) {
	pm := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{"test"},
			},
			AttachMetadata: &monitoringv1.AttachMetadata{
				Node: true,
			},
		},
	}

	cg := mustNewConfigGenerator(
		t,
		&monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					IgnoreNamespaceSelectors: false,
				},
			},
		},
	)

	attachMetadataConfig := &attachMetadataConfig{
		MinimumVersion: "2.35.0",
		AttachMetadata: pm.Spec.AttachMetadata,
	}
	c := cg.generateK8SSDConfig(pm.Spec.NamespaceSelector, pm.Namespace, nil, nil, kubernetesSDRolePod, attachMetadataConfig)

	s, err := yaml.Marshal(yaml.MapSlice{c})
	if err != nil {
		t.Fatal(err)
	}

	expected := `kubernetes_sd_configs:
- role: pod
  namespaces:
    names:
    - test
  attach_metadata:
    node: true
`

	result := string(s)
	if expected != result {
		t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", result, expected)
	}
}

func TestProbeStaticTargetsConfigGeneration(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme:   "http",
						URL:      "blackbox.exporter.io",
						Path:     "/probe",
						ProxyURL: "socks://myproxy:9095",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
							Targets: []string{
								"prometheus.io",
								"promcon.io",
							},
							Labels: map[string]string{
								"static": "label",
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: "bar",
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  proxy_url: socks://myproxy:9095
  params:
    module:
    - http_2xx
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: default
      static: label
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  - target_label: foo
    replacement: bar
    action: replace
  metric_relabel_configs: []
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestProbeStaticTargetsConfigGenerationWithLabelEnforce(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "namespace",
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
							Targets: []string{
								"prometheus.io",
								"promcon.io",
							},
							Labels: map[string]string{
								"namespace": "custom",
								"static":    "label",
							},
						},
					},
					MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Regex:  "noisy_labels.*",
							Action: "labeldrop",
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  params:
    module:
    - http_2xx
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: custom
      static: label
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  - target_label: namespace
    replacement: default
  metric_relabel_configs:
  - regex: noisy_labels.*
    action: labeldrop
  - target_label: namespace
    replacement: default
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestProbeStaticTargetsConfigGenerationWithJobName(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					JobName: "blackbox",
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
							Targets: []string{
								"prometheus.io",
								"promcon.io",
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  params:
    module:
    - http_2xx
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: default
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - target_label: job
    replacement: blackbox
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  metric_relabel_configs: []
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestProbeStaticTargetsConfigGenerationWithoutModule(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					JobName: "blackbox",
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
							Targets: []string{
								"prometheus.io",
								"promcon.io",
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: default
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - target_label: job
    replacement: blackbox
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  metric_relabel_configs: []
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestProbeIngressSDConfigGeneration(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						Ingress: &monitoringv1.ProbeTargetIngress{
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"prometheus.io/probe": "true",
								},
							},
							NamespaceSelector: monitoringv1.NamespaceSelector{
								Any: true,
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: "bar",
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  params:
    module:
    - http_2xx
  kubernetes_sd_configs:
  - role: ingress
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_ingress_label_prometheus_io_probe
    - __meta_kubernetes_ingress_labelpresent_prometheus_io_probe
    regex: (true);true
  - source_labels:
    - __meta_kubernetes_ingress_scheme
    - __address__
    - __meta_kubernetes_ingress_path
    separator: ;
    regex: (.+);(.+);(.+)
    target_label: __param_target
    replacement: ${1}://${2}${3}
    action: replace
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_ingress_name
    target_label: ingress
  - source_labels:
    - __address__
    separator: ;
    regex: (.*)
    target_label: __tmp_ingress_address
    replacement: $1
    action: replace
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  - target_label: foo
    replacement: bar
    action: replace
  - source_labels:
    - __param_target
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestProbeIngressSDConfigGenerationWithShards(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				Shards:         pointer.Int32(2),
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						Ingress: &monitoringv1.ProbeTargetIngress{
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"prometheus.io/probe": "true",
								},
							},
							NamespaceSelector: monitoringv1.NamespaceSelector{
								Any: true,
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: "bar",
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  params:
    module:
    - http_2xx
  kubernetes_sd_configs:
  - role: ingress
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_ingress_label_prometheus_io_probe
    - __meta_kubernetes_ingress_labelpresent_prometheus_io_probe
    regex: (true);true
  - source_labels:
    - __meta_kubernetes_ingress_scheme
    - __address__
    - __meta_kubernetes_ingress_path
    separator: ;
    regex: (.+);(.+);(.+)
    target_label: __param_target
    replacement: ${1}://${2}${3}
    action: replace
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_ingress_name
    target_label: ingress
  - source_labels:
    - __address__
    separator: ;
    regex: (.*)
    target_label: __tmp_ingress_address
    replacement: $1
    action: replace
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  - target_label: foo
    replacement: bar
    action: replace
  - source_labels:
    - __param_target
    target_label: __tmp_hash
    modulus: 2
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)
	if expected != result {
		t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", result, expected)
	}
}

func TestProbeIngressSDConfigGenerationWithLabelEnforce(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "namespace",
				ProbeSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Version:        operator.DefaultPrometheusVersion,
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		map[string]*monitoringv1.Probe{
			"probe1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme: "http",
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						Ingress: &monitoringv1.ProbeTargetIngress{
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"prometheus.io/probe": "true",
								},
							},
							NamespaceSelector: monitoringv1.NamespaceSelector{
								Any: true,
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: "bar",
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  params:
    module:
    - http_2xx
  kubernetes_sd_configs:
  - role: ingress
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_ingress_label_prometheus_io_probe
    - __meta_kubernetes_ingress_labelpresent_prometheus_io_probe
    regex: (true);true
  - source_labels:
    - __meta_kubernetes_ingress_scheme
    - __address__
    - __meta_kubernetes_ingress_path
    separator: ;
    regex: (.+);(.+);(.+)
    target_label: __param_target
    replacement: ${1}://${2}${3}
    action: replace
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_ingress_name
    target_label: ingress
  - source_labels:
    - __address__
    separator: ;
    regex: (.*)
    target_label: __tmp_ingress_address
    replacement: $1
    action: replace
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  - target_label: foo
    replacement: bar
    action: replace
  - target_label: namespace
    replacement: default
  - source_labels:
    - __param_target
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - target_label: namespace
    replacement: default
`

	result := string(cfg)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected result got(-) want(+)\n%s\n", diff)
	}
}

func TestK8SSDConfigGeneration(t *testing.T) {
	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{"test"},
			},
		},
	}

	testcases := []struct {
		apiserverConfig *monitoringv1.APIServerConfig
		store           *assets.Store
		expected        string
	}{
		{
			nil,
			nil,
			`kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test
`,
		},
		{
			&monitoringv1.APIServerConfig{
				Host:            "example.com",
				BasicAuth:       &monitoringv1.BasicAuth{},
				BearerToken:     "bearer_token",
				BearerTokenFile: "bearer_token_file",
				TLSConfig:       nil,
			},
			&assets.Store{
				BasicAuthAssets: map[string]assets.BasicAuthCredentials{
					"apiserver": {
						Username: "foo",
						Password: "bar",
					},
				},
				OAuth2Assets: map[string]assets.OAuth2Credentials{},
				TokenAssets:  map[string]assets.Token{},
			},
			`kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test
  api_server: example.com
  basic_auth:
    username: foo
    password: bar
  bearer_token: bearer_token
  bearer_token_file: bearer_token_file
`,
		},
	}

	for _, tc := range testcases {
		cg := mustNewConfigGenerator(
			t,
			&monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						IgnoreNamespaceSelectors: false,
					},
				},
			},
		)

		var attachMetaConfig *attachMetadataConfig
		if sm.Spec.AttachMetadata != nil {
			attachMetaConfig = &attachMetadataConfig{
				MinimumVersion: "2.37.0",
				AttachMetadata: sm.Spec.AttachMetadata,
			}
		}
		c := cg.generateK8SSDConfig(
			sm.Spec.NamespaceSelector,
			sm.Namespace,
			tc.apiserverConfig,
			tc.store,
			kubernetesSDRoleEndpoint,
			attachMetaConfig,
		)
		s, err := yaml.Marshal(yaml.MapSlice{c})
		if err != nil {
			t.Fatal(err)
		}
		result := string(s)

		if result != tc.expected {
			t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", result, tc.expected)
		}
	}
}

func TestAlertmanagerBearerToken(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:            "alertmanager-main",
						Namespace:       "default",
						Port:            intstr.FromString("web"),
						BearerTokenFile: "/some/file/on/disk",
					},
				},
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `bearer_token_file` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    bearer_token_file: /some/file/on/disk
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestAlertmanagerBasicAuth(t *testing.T) {
	for _, tc := range []struct {
		name           string
		version        string
		expectedConfig string
	}{
		{
			name:    "Valid Prom Version",
			version: "2.26.0",
			expectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    basic_auth:
      username: bob
      password: alice
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`,
		},
		{
			name:    "Invalid Prom Version",
			version: "2.25.0",
			expectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`,
		},
	} {

		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ScrapeInterval: "30s",
					Version:        tc.version,
				},
				EvaluationInterval: "30s",
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:      "alertmanager-main",
							Namespace: "default",
							Port:      intstr.FromString("web"),
							BasicAuth: &monitoringv1.BasicAuth{
								Username: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "foo",
									},
									Key: "username",
								},
								Password: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "foo",
									},
									Key: "password",
								},
							},
						},
					},
				},
			},
		}

		cg := mustNewConfigGenerator(t, p)

		cfg, err := cg.GenerateServerConfiguration(
			p.Spec.EvaluationInterval,
			p.Spec.QueryLogFile,
			p.Spec.RuleSelector,
			p.Spec.Exemplars,
			p.Spec.TSDB,
			p.Spec.Alerting,
			p.Spec.RemoteRead,
			nil,
			nil,
			nil,
			&assets.Store{BasicAuthAssets: map[string]assets.BasicAuthCredentials{
				"alertmanager/auth/0": {
					Username: "bob",
					Password: "alice",
				},
			}},
			nil,
			nil,
			nil,
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}

		result := string(cfg)

		if diff := cmp.Diff(tc.expectedConfig, result); diff != "" {
			t.Logf("\n%s", diff)
			t.Fatal("expected Prometheus configuration and actual configuration do not match")
		}
	}
}

func TestAlertmanagerAPIVersion(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:        "v2.11.0",
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  "default",
						Port:       intstr.FromString("web"),
						APIVersion: "v2",
					},
				},
			},
		},
	}
	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `api_version` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    api_version: v2
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestAlertmanagerTimeoutConfig(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:        "v2.11.0",
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  "default",
						Port:       intstr.FromString("web"),
						APIVersion: "v2",
						Timeout:    (*monitoringv1.Duration)(pointer.String("60s")),
					},
				},
			},
		},
	}
	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `api_version` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    timeout: 60s
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    api_version: v2
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestAlertmanagerEnableHttp2(t *testing.T) {
	expectedWithHTTP2Unsupported := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    api_version: v2
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	expectedWithHTTP2Disabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    enable_http2: false
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    api_version: v2
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	expectedWithHTTP2Enabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    enable_http2: true
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    api_version: v2
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	for _, tc := range []struct {
		version     string
		expected    string
		enableHTTP2 bool
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Enabled,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Disabled,
		},
	} {
		t.Run(fmt.Sprintf("%s TestAlertmanagerEnableHttp2(%t)", tc.version, tc.enableHTTP2), func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},

				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
					},
					EvaluationInterval: "30s",
					Alerting: &monitoringv1.AlertingSpec{
						Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
							{
								Name:        "alertmanager-main",
								Namespace:   "default",
								Port:        intstr.FromString("web"),
								APIVersion:  "v2",
								EnableHttp2: swag.Bool(tc.enableHTTP2),
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestAdditionalScrapeConfigs(t *testing.T) {
	getCfg := func(shards *int32) string {
		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Shards:         shards,
					ScrapeInterval: "30s",
				},
				EvaluationInterval: "30s",
			},
		}

		cg := mustNewConfigGenerator(t, p)

		cfg, err := cg.GenerateServerConfiguration(
			p.Spec.EvaluationInterval,
			p.Spec.QueryLogFile,
			p.Spec.RuleSelector,
			p.Spec.Exemplars,
			p.Spec.TSDB,
			p.Spec.Alerting,
			p.Spec.RemoteRead,
			nil,
			nil,
			nil,
			&assets.Store{},
			[]byte(`- job_name: prometheus
  scrape_interval: 15s
  static_configs:
  - targets: ['localhost:9090']
- job_name: gce_app_bar
  scrape_interval: 5s
  gce_sd_config:
    - project: foo
      zone: us-central1
  relabel_configs:
    - action: keep
      source_labels:
      - __meta_gce_label_app
      regex: my_app
`),
			nil,
			nil,
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		return string(cfg)
	}

	testCases := []struct {
		name     string
		result   string
		expected string
	}{
		{
			name:   "unsharded prometheus",
			result: getCfg(nil),
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: prometheus
  scrape_interval: 15s
  static_configs:
  - targets:
    - localhost:9090
- job_name: gce_app_bar
  scrape_interval: 5s
  gce_sd_config:
  - project: foo
    zone: us-central1
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_gce_label_app
    regex: my_app
`,
		},
		{
			name:   "one prometheus shard",
			result: getCfg(pointer.Int32(1)),
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: prometheus
  scrape_interval: 15s
  static_configs:
  - targets:
    - localhost:9090
- job_name: gce_app_bar
  scrape_interval: 5s
  gce_sd_config:
  - project: foo
    zone: us-central1
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_gce_label_app
    regex: my_app
`,
		},
		{
			name:   "sharded prometheus",
			result: getCfg(pointer.Int32(3)),
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: prometheus
  scrape_interval: 15s
  static_configs:
  - targets:
    - localhost:9090
  relabel_configs:
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 3
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
- job_name: gce_app_bar
  scrape_interval: 5s
  gce_sd_config:
  - project: foo
    zone: us-central1
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_gce_label_app
    regex: my_app
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 3
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
`,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.expected != tt.result {
					fmt.Println(pretty.Compare(tt.expected, tt.result))
					t.Fatal("expected Prometheus configuration and actual configuration do not match")
				}
			},
		)
	}
}

func TestAdditionalAlertRelabelConfigs(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "alertmanager-main",
						Namespace: "default",
						Port:      intstr.FromString("web"),
					},
				},
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		nil,
		&assets.Store{},
		nil,
		[]byte(`- action: drop
  source_labels: [__meta_kubernetes_node_name]
  regex: spot-(.+)

`),
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  - action: drop
    source_labels:
    - __meta_kubernetes_node_name
    regex: spot-(.+)
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestNoEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:        "https-metrics",
							HonorLabels: true,
							Interval:    "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)",
									SourceLabels: []monitoringv1.LabelName{"__name__"},
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.+)(?::d+)",
									Replacement:  "$1:9537",
									SourceLabels: []monitoringv1.LabelName{"__address__"},
									TargetLabel:  "__address__",
								},
								{
									Action:      "replace",
									Replacement: "crio",
									TargetLabel: "job",
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: true
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: https-metrics
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: https-metrics
  - source_labels:
    - __address__
    target_label: __address__
    regex: (.+)(?::d+)
    replacement: $1:9537
    action: replace
  - target_label: job
    replacement: crio
    action: replace
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - __name__
    regex: container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)
    action: drop
`

	result := string(cfg)
	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestServiceMonitorWithEndpointSliceEnable(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "ns-key",
				ScrapeInterval:         "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cg.endpointSliceSupported = true

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "alpha",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"beta", "gamma"},
							},
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "ns-key",
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  "$1",
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpointslice
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_alpha
    - __meta_kubernetes_service_labelpresent_alpha
    regex: (beta|gamma);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpointslice_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpointslice_address_target_kind
    - __meta_kubernetes_endpointslice_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpointslice_address_target_kind
    - __meta_kubernetes_endpointslice_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __meta_kubernetes_pod_ready
    regex: (.*)
    replacement: $1
    action: replace
  - target_label: ns-key
    replacement: default
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - pod_name
    target_label: ns-key
    regex: my-job-pod-.+
    action: drop
  - target_label: ns-key
    replacement: default
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match for enforced namespace label test:\n%s", diff)
	}
}

func TestEnforcedNamespaceLabelPodMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "ns-key",
				ScrapeInterval:         "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "pod-monitor-ns",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     "web",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "my-ns",
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  "$1",
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
								},
							},
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __meta_kubernetes_pod_ready
    regex: (.*)
    replacement: $1
    action: replace
  - target_label: ns-key
    replacement: pod-monitor-ns
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - pod_name
    target_label: my-ns
    regex: my-job-pod-.+
    action: drop
  - target_label: ns-key
    replacement: pod-monitor-ns
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match\n%s", diff)
	}
}

func TestEnforcedNamespaceLabelOnExcludedPodMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "ns-key",
				ExcludedFromEnforcement: []monitoringv1.ObjectReference{
					{
						Namespace: "pod-monitor-ns",
						Group:     monitoring.GroupName,
						Resource:  monitoringv1.PodMonitorName,
						Name:      "testpodmonitor1",
					},
				},
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "pod-monitor-ns",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
					Kind:       monitoringv1.PodMonitorsKind,
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     "web",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "my-ns",
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  "$1",
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
								},
							},
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __meta_kubernetes_pod_ready
    regex: (.*)
    replacement: $1
    action: replace
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - pod_name
    target_label: my-ns
    regex: my-job-pod-.+
    action: drop
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match\n%s", diff)
	}
}

func TestEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnforcedNamespaceLabel: "ns-key",
				ScrapeInterval:         "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "alpha",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"beta", "gamma"},
							},
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "ns-key",
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  "$1",
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_alpha
    - __meta_kubernetes_service_labelpresent_alpha
    regex: (beta|gamma);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __meta_kubernetes_pod_ready
    regex: (.*)
    replacement: $1
    action: replace
  - target_label: ns-key
    replacement: default
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - pod_name
    target_label: ns-key
    regex: my-job-pod-.+
    action: drop
  - target_label: ns-key
    replacement: default
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match for enforced namespace label test:\n%s", diff)
	}
}

func TestEnforcedNamespaceLabelOnExcludedServiceMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval:         "30s",
				EnforcedNamespaceLabel: "ns-key",
				ExcludedFromEnforcement: []monitoringv1.ObjectReference{
					{
						Namespace: "service-monitor-ns",
						Group:     monitoring.GroupName,
						Resource:  monitoringv1.ServiceMonitorName,
						Name:      "", // exclude all servicemonitors in this namespace
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "servicemonitor1",
					Namespace: "service-monitor-ns",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
					Kind:       monitoringv1.ServiceMonitorsKind,
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "ns-key",
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  "$1",
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/service-monitor-ns/servicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - service-monitor-ns
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __meta_kubernetes_pod_ready
    regex: (.*)
    replacement: $1
    action: replace
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - pod_name
    target_label: ns-key
    regex: my-job-pod-.+
    action: drop
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match for enforced namespace label test:\n%s", diff)
	}
}

func TestAdditionalAlertmanagers(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "alertmanager-main",
						Namespace: "default",
						Port:      intstr.FromString("web"),
					},
				},
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		[]byte(`- static_configs:
  - targets:
    - localhost
`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
  - static_configs:
    - targets:
      - localhost
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestSettingHonorTimestampsInServiceMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:        "v2.9.0",
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							HonorTimestamps: swag.Bool(false),
							Port:            "web",
							Interval:        "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  honor_timestamps: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestSettingHonorTimestampsInPodMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:        "v2.9.0",
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							HonorTimestamps: swag.Bool(false),
							Port:            "web",
							Interval:        "30s",
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  honor_timestamps: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestHonorTimestampsOverriding(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:                 "v2.9.0",
				ScrapeInterval:          "30s",
				OverrideHonorTimestamps: true,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							HonorTimestamps: swag.Bool(true),
							Port:            "web",
							Interval:        "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  honor_timestamps: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestSettingHonorLabels(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							HonorLabels: true,
							Port:        "web",
							Interval:    "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: true
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestHonorLabelsOverriding(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval:      "30s",
				OverrideHonorLabels: true,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							HonorLabels: true,
							Port:        "web",
							Interval:    "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestTargetLabels(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval:      "30s",
				OverrideHonorLabels: false,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestEndpointOAuth2(t *testing.T) {
	oauth2 := monitoringv1.OAuth2{
		ClientID: monitoringv1.SecretOrConfigMap{
			ConfigMap: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "oauth2",
				},
				Key: "client_id",
			},
		},
		ClientSecret: v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "oauth2",
			},
			Key: "client_secret",
		},
		TokenURL: "http://test.url",
		Scopes:   []string{"scope 1", "scope 2"},
		EndpointParams: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	expectedCfg := strings.TrimSpace(`
oauth2:
    client_id: test_client_id
    client_secret: test_client_secret
    token_url: http://test.url
    scopes:
    - scope 1
    - scope 2
    endpoint_params:
      param1: value1
      param2: value2`)

	testCases := []struct {
		name              string
		p                 *monitoringv1.Prometheus
		sMons             map[string]*monitoringv1.ServiceMonitor
		pMons             map[string]*monitoringv1.PodMonitor
		probes            map[string]*monitoringv1.Probe
		oauth2Credentials map[string]assets.OAuth2Credentials
		expectedCfg       string
	}{
		{
			name: "service monitor with oauth2",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						OverrideHonorLabels: false,
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
						Version: "v2.27.0",
					},
				},
			},
			sMons: map[string]*monitoringv1.ServiceMonitor{
				"testservicemonitor1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testservicemonitor1",
						Namespace: "default",
						Labels: map[string]string{
							"group": "group1",
						},
					},
					Spec: monitoringv1.ServiceMonitorSpec{
						Endpoints: []monitoringv1.Endpoint{
							{
								Port:   "web",
								OAuth2: &oauth2,
							},
						},
					},
				},
			},
			oauth2Credentials: map[string]assets.OAuth2Credentials{
				"serviceMonitor/default/testservicemonitor1/0": {
					ClientID:     "test_client_id",
					ClientSecret: "test_client_secret",
				},
			},
			expectedCfg: expectedCfg,
		},
		{
			name: "pod monitor with oauth2",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						OverrideHonorLabels: false,
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
						Version: "v2.27.0",
					},
				},
			},
			pMons: map[string]*monitoringv1.PodMonitor{
				"testpodmonitor1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpodmonitor1",
						Namespace: "default",
						Labels: map[string]string{
							"group": "group1",
						},
					},
					Spec: monitoringv1.PodMonitorSpec{
						PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
							{
								Port:   "web",
								OAuth2: &oauth2,
							},
						},
					},
				},
			},
			oauth2Credentials: map[string]assets.OAuth2Credentials{
				"podMonitor/default/testpodmonitor1/0": {
					ClientID:     "test_client_id",
					ClientSecret: "test_client_secret",
				},
			},
			expectedCfg: expectedCfg,
		},
		{
			name: "probe monitor with oauth2",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						OverrideHonorLabels: false,
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
						Version: "v2.27.0",
					},
				},
			},
			probes: map[string]*monitoringv1.Probe{
				"testprobe1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testprobe1",
						Namespace: "default",
						Labels: map[string]string{
							"group": "group1",
						},
					},
					Spec: monitoringv1.ProbeSpec{
						OAuth2: &oauth2,
						Targets: monitoringv1.ProbeTargets{
							StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
								Targets: []string{"127.0.0.1"},
							},
						},
					},
				},
			},
			oauth2Credentials: map[string]assets.OAuth2Credentials{
				"probe/default/testprobe1": {
					ClientID:     "test_client_id",
					ClientSecret: "test_client_secret",
				},
			},
			expectedCfg: expectedCfg,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cg := mustNewConfigGenerator(t, tt.p)
			cfg, err := cg.GenerateServerConfiguration(
				tt.p.Spec.EvaluationInterval,
				tt.p.Spec.QueryLogFile,
				tt.p.Spec.RuleSelector,
				tt.p.Spec.Exemplars,
				tt.p.Spec.TSDB,
				tt.p.Spec.Alerting,
				tt.p.Spec.RemoteRead,
				tt.sMons,
				tt.pMons,
				tt.probes,
				&assets.Store{
					BasicAuthAssets: map[string]assets.BasicAuthCredentials{},
					OAuth2Assets:    tt.oauth2Credentials,
					TokenAssets:     map[string]assets.Token{},
				},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if !strings.Contains(result, tt.expectedCfg) {
				t.Fatalf("expected Prometheus configuration to contain:\n %s\nFull config:\n %s", tt.expectedCfg, result)
			}
		})
	}
}

func TestPodTargetLabels(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestPodTargetLabelsFromPodMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestPodTargetLabelsFromPodMonitorAndGlobal(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				PodTargetLabels: []string{"global"},
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodTargetLabels: []string{"local"},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_local
    target_label: local
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_global
    target_label: global
    regex: (.+)
    replacement: ${1}
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestEmptyEndpointPorts(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						// Add a single endpoint with empty configuration.
						{},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `bearer_token_file` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)
	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func generateTestConfig(t *testing.T, version string) ([]byte, error) {
	t.Helper()

	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "alertmanager-main",
						Namespace: "default",
						Port:      intstr.FromString("web"),
					},
				},
			},
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ExternalLabels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				Version:  version,
				Replicas: func(i int32) *int32 { return &i }(1),
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				PodMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
				RemoteWrite: []monitoringv1.RemoteWriteSpec{{
					URL: "https://example.com/remote_write",
				}},
			},
			RuleSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"role": "rulefile",
				},
			},
			RemoteRead: []monitoringv1.RemoteReadSpec{{
				URL: "https://example.com/remote_read",
			}},
		},
	}
	cg := mustNewConfigGenerator(t, p)

	return cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		makeServiceMonitors(),
		makePodMonitors(),
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
}

func makeServiceMonitors() map[string]*monitoringv1.ServiceMonitor {
	res := map[string]*monitoringv1.ServiceMonitor{}

	res["servicemonitor1"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": "group1",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["servicemonitor2"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor2",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group2",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group2",
					"group3": "group3",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["servicemonitor3"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor3",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group4",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group4",
					"group3": "group5",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					Path:     "/federate",
					Params:   map[string][]string{"metrics[]": {"{__name__=~\"job:.*\"}"}},
				},
			},
		},
	}

	res["servicemonitor4"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group6",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group6",
					"group3": "group7",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "drop",
							Regex:        "my-job-pod-.+",
							SourceLabels: []monitoringv1.LabelName{"pod_name"},
						},
						{
							Action:       "drop",
							Regex:        "test",
							SourceLabels: []monitoringv1.LabelName{"namespace"},
						},
					},
				},
			},
		},
	}

	res["servicemonitor5"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group8",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group8",
					"group3": "group9",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					RelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
							TargetLabel:  "pod_ready",
						},
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
							TargetLabel:  "nodename",
						},
					},
				},
			},
		},
	}

	return res
}

func makePodMonitors() map[string]*monitoringv1.PodMonitor {
	res := map[string]*monitoringv1.PodMonitor{}

	res["podmonitor1"] = &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": "group1",
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["podmonitor2"] = &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor2",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group2",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group2",
					"group3": "group3",
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["podmonitor3"] = &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor3",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group4",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group4",
					"group3": "group5",
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
					Path:     "/federate",
					Params:   map[string][]string{"metrics[]": {"{__name__=~\"job:.*\"}"}},
				},
			},
		},
	}

	res["podmonitor4"] = &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group6",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group6",
					"group3": "group7",
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
					MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "drop",
							Regex:        "my-job-pod-.+",
							SourceLabels: []monitoringv1.LabelName{"pod_name"},
						},
						{
							Action:       "drop",
							Regex:        "test",
							SourceLabels: []monitoringv1.LabelName{"namespace"},
						},
					},
				},
			},
		},
	}

	res["podmonitor5"] = &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodmonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group8",
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group8",
					"group3": "group9",
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
					RelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
							TargetLabel:  "pod_ready",
						},
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
							TargetLabel:  "nodename",
						},
					},
				},
			},
		},
	}

	return res
}

func TestHonorLabels(t *testing.T) {
	type testCase struct {
		UserHonorLabels     bool
		OverrideHonorLabels bool
		Expected            bool
	}

	testCases := []testCase{
		{
			UserHonorLabels:     false,
			OverrideHonorLabels: true,
			Expected:            false,
		},
		{
			UserHonorLabels:     true,
			OverrideHonorLabels: false,
			Expected:            true,
		},
		{
			UserHonorLabels:     true,
			OverrideHonorLabels: true,
			Expected:            false,
		},
		{
			UserHonorLabels:     false,
			OverrideHonorLabels: false,
			Expected:            false,
		},
	}

	for _, tc := range testCases {
		cg := mustNewConfigGenerator(
			t,
			&monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						OverrideHonorLabels: tc.OverrideHonorLabels,
					},
				},
			},
		)
		cfg := cg.AddHonorLabels(yaml.MapSlice{}, tc.UserHonorLabels)
		k, v := cfg[0].Key.(string), cfg[0].Value.(bool)
		if k != "honor_labels" {
			t.Fatalf("expected key 'honor_labels', got %q", k)
		}
		if tc.Expected != v {
			t.Fatalf("\nGot: %t, \nExpected: %t\nFor values UserHonorLabels %t, OverrideHonorLabels %t\n", v, tc.Expected, tc.UserHonorLabels, tc.OverrideHonorLabels)
		}
	}
}

func TestHonorTimestamps(t *testing.T) {
	type testCase struct {
		UserHonorTimestamps     *bool
		OverrideHonorTimestamps bool
		Expected                string
	}

	testCases := []testCase{
		{
			UserHonorTimestamps:     nil,
			OverrideHonorTimestamps: true,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     nil,
			OverrideHonorTimestamps: false,
			Expected:                "{}\n",
		},
		{
			UserHonorTimestamps:     swag.Bool(false),
			OverrideHonorTimestamps: true,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     swag.Bool(false),
			OverrideHonorTimestamps: false,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     swag.Bool(true),
			OverrideHonorTimestamps: true,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     swag.Bool(true),
			OverrideHonorTimestamps: false,
			Expected:                "honor_timestamps: true\n",
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			cg := mustNewConfigGenerator(t, &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:                 "2.9.0",
						OverrideHonorTimestamps: tc.OverrideHonorTimestamps,
					},
				},
			})

			hl, _ := yaml.Marshal(cg.AddHonorTimestamps(yaml.MapSlice{}, tc.UserHonorTimestamps))
			cfg := string(hl)
			if tc.Expected != cfg {
				t.Fatalf("\nGot: %s, \nExpected: %s\nFor values UserHonorTimestamps %+v, OverrideHonorTimestamps %t\n", cfg, tc.Expected, tc.UserHonorTimestamps, tc.OverrideHonorTimestamps)
			}
		})
	}
}

func TestSampleLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  sample_limit: %d
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		enforcedLimit int
		limit         int
		expected      string
	}{
		{
			enforcedLimit: -1,
			limit:         -1,
			expected:      expectNoLimit,
		},
		{
			enforcedLimit: 1000,
			limit:         -1,
			expected:      fmt.Sprintf(expectLimit, 1000),
		},
		{
			enforcedLimit: 1000,
			limit:         2000,
			expected:      fmt.Sprintf(expectLimit, 1000),
		},
		{
			enforcedLimit: 1000,
			limit:         500,
			expected:      fmt.Sprintf(expectLimit, 500),
		},
	} {
		t.Run(fmt.Sprintf("enforcedlimit(%d) limit(%d)", tc.enforcedLimit, tc.limit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        "v2.20.0",
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}
			if tc.enforcedLimit >= 0 {
				i := uint64(tc.enforcedLimit)
				prometheus.Spec.EnforcedSampleLimit = &i
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}
			if tc.limit >= 0 {
				serviceMonitor.Spec.SampleLimit = uint64(tc.limit)
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestTargetLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  target_limit: %d
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version       string
		enforcedLimit int
		limit         int
		expected      string
	}{
		{
			version:       "v2.15.0",
			enforcedLimit: -1,
			limit:         -1,
			expected:      expectNoLimit,
		},
		{
			version:       "v2.21.0",
			enforcedLimit: -1,
			limit:         -1,
			expected:      expectNoLimit,
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         -1,
			expected:      expectNoLimit,
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         -1,
			expected:      fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         2000,
			expected:      expectNoLimit,
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         2000,
			expected:      fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         500,
			expected:      expectNoLimit,
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         500,
			expected:      fmt.Sprintf(expectLimit, 500),
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedlimit(%d) limit(%d)", tc.version, tc.enforcedLimit, tc.limit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}
			if tc.enforcedLimit >= 0 {
				i := uint64(tc.enforcedLimit)
				prometheus.Spec.EnforcedTargetLimit = &i
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}
			if tc.limit >= 0 {
				serviceMonitor.Spec.TargetLimit = uint64(tc.limit)
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestRemoteReadConfig(t *testing.T) {
	boolTrue := true
	boolFalse := false

	for _, tc := range []struct {
		version     string
		remoteRead  monitoringv1.RemoteReadSpec
		expected    string
		expectedErr error
	}{
		{
			version: "v2.27.1",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
  oauth2:
    client_id: client-id
    client_secret: client-secret
    token_url: http://token-url
    scopes:
    - scope1
    endpoint_params:
      param: value
`,
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
`,
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: &boolTrue,
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
`,
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
`,
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: &boolFalse,
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
  filter_external_labels: false
`,
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: &boolTrue,
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
  filter_external_labels: true
`,
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				Authorization: &monitoringv1.Authorization{
					SafeAuthorization: monitoringv1.SafeAuthorization{
						Credentials: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "key",
							},
						},
					},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_read:
- url: http://example.com
  remote_timeout: 30s
  authorization:
    type: Bearer
    credentials: secret
`,
		},
	} {
		t.Run(fmt.Sprintf("version=%s", tc.version), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
					RemoteRead:         []monitoringv1.RemoteReadSpec{tc.remoteRead},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				nil,
				nil,
				&assets.Store{
					BasicAuthAssets: map[string]assets.BasicAuthCredentials{},
					OAuth2Assets: map[string]assets.OAuth2Credentials{
						"remoteRead/0": {
							ClientID:     "client-id",
							ClientSecret: "client-secret",
						},
					},
					TokenAssets: map[string]assets.Token{
						"remoteRead/auth/0": assets.Token("secret"),
					},
				},
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr != nil {
				if tc.expectedErr.Error() != err.Error() {
					t.Logf("\n%s", pretty.Compare(tc.expectedErr.Error(), err.Error()))
					t.Fatal("expected error and actual error do not match")
				}
				return
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestRemoteWriteConfig(t *testing.T) {
	for _, tc := range []struct {
		version     string
		remoteWrite monitoringv1.RemoteWriteSpec
		expected    string
		expectedErr error
	}{
		{
			version: "v2.22.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    min_backoff: 1s
    max_backoff: 10s
`,
		},
		{
			version: "v2.23.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    min_backoff: 1s
    max_backoff: 10s
  metadata_config:
    send: false
    send_interval: 1m
`,
		},
		{
			version: "v2.23.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    min_backoff: 1s
    max_backoff: 10s
  metadata_config:
    send: false
    send_interval: 1m
`,
		},
		{
			version: "v2.10.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    max_retries: 3
    min_backoff: 1s
    max_backoff: 10s
`,
		},
		{
			version: "v2.27.1",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  oauth2:
    client_id: client-id
    client_secret: client-secret
    token_url: http://token-url
    scopes:
    - scope1
    endpoint_params:
      param: value
`,
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				Authorization: &monitoringv1.Authorization{
					SafeAuthorization: monitoringv1.SafeAuthorization{
						Credentials: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "key",
							},
						},
					},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  authorization:
    type: Bearer
    credentials: secret
`,
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				Sigv4: &monitoringv1.Sigv4{
					Profile: "profilename",
					RoleArn: "arn:aws:iam::123456789012:instance-profile/prometheus",
					AccessKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "sigv4-secret",
						},
						Key: "access-key",
					},
					SecretKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "sigv4-secret",
						},
						Key: "secret-key",
					},
					Region: "us-central-0",
				},
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  sigv4:
    region: us-central-0
    access_key: access-key
    secret_key: secret-key
    profile: profilename
    role_arn: arn:aws:iam::123456789012:instance-profile/prometheus
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    min_backoff: 1s
    max_backoff: 10s
  metadata_config:
    send: false
    send_interval: 1m
`,
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				RemoteTimeout: "1s",
				Sigv4:         nil,
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 1s
`,
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				Sigv4:         &monitoringv1.Sigv4{},
				RemoteTimeout: "1s",
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 1s
  sigv4: {}
`,
		},
		{
			version: "v2.30.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
					RetryOnRateLimit:  true,
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
remote_write:
- url: http://example.com
  remote_timeout: 30s
  queue_config:
    capacity: 1000
    min_shards: 1
    max_shards: 10
    max_samples_per_send: 100
    batch_send_deadline: 20s
    min_backoff: 1s
    max_backoff: 10s
    retry_on_http_429: true
`,
		},
	} {
		t.Run(fmt.Sprintf("version=%s", tc.version), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
						RemoteWrite: []monitoringv1.RemoteWriteSpec{tc.remoteWrite},
						Secrets:     []string{"sigv4-secret"},
					},
					EvaluationInterval: "30s",
				},
			}

			store := &assets.Store{
				BasicAuthAssets: map[string]assets.BasicAuthCredentials{},
				OAuth2Assets: map[string]assets.OAuth2Credentials{
					"remoteWrite/0": {
						ClientID:     "client-id",
						ClientSecret: "client-secret",
					},
				},
				TokenAssets: map[string]assets.Token{
					"remoteWrite/auth/0": assets.Token("secret"),
				},
			}
			if tc.remoteWrite.Sigv4 != nil && tc.remoteWrite.Sigv4.AccessKey != nil {
				store.SigV4Assets = map[string]assets.SigV4Credentials{
					"remoteWrite/0": {
						AccessKeyID: "access-key",
						SecretKeyID: "secret-key",
					},
				}
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				nil,
				nil,
				store,
				nil,
				nil,
				nil,
				nil)
			if tc.expectedErr != nil {
				if tc.expectedErr.Error() != err.Error() {
					t.Logf("\n%s", pretty.Compare(tc.expectedErr.Error(), err.Error()))
					t.Fatal("expected error and actual error do not match")
				}
				return
			}
			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestLabelLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  label_limit: %d
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version            string
		enforcedLabelLimit int
		labelLimit         int
		expected           string
	}{
		{
			version:            "v2.26.0",
			enforcedLabelLimit: -1,
			labelLimit:         -1,
			expected:           expectNoLimit,
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: -1,
			labelLimit:         -1,
			expected:           expectNoLimit,
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         -1,
			expected:           expectNoLimit,
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         -1,
			expected:           fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         2000,
			expected:           expectNoLimit,
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         2000,
			expected:           fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         500,
			expected:           expectNoLimit,
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         500,
			expected:           fmt.Sprintf(expectLimit, 500),
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelLimit(%d) labelLimit(%d)", tc.version, tc.enforcedLabelLimit, tc.labelLimit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			if tc.enforcedLabelLimit >= 0 {
				i := uint64(tc.enforcedLabelLimit)
				prometheus.Spec.EnforcedLabelLimit = &i
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}
			if tc.labelLimit >= 0 {
				serviceMonitor.Spec.LabelLimit = uint64(tc.labelLimit)
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestLabelNameLengthLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  label_name_length_limit: %d
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version                      string
		enforcedLabelNameLengthLimit int
		labelNameLengthLimit         int
		expected                     string
	}{
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: -1,
			labelNameLengthLimit:         -1,
			expected:                     expectNoLimit,
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: -1,
			labelNameLengthLimit:         -1,
			expected:                     expectNoLimit,
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         -1,
			expected:                     expectNoLimit,
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         -1,
			expected:                     fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         2000,
			expected:                     expectNoLimit,
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         2000,
			expected:                     fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         500,
			expected:                     expectNoLimit,
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         500,
			expected:                     fmt.Sprintf(expectLimit, 500),
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelNameLengthLimit(%d) labelNameLengthLimit(%d)", tc.version, tc.enforcedLabelNameLengthLimit, tc.labelNameLengthLimit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			if tc.enforcedLabelNameLengthLimit >= 0 {
				i := uint64(tc.enforcedLabelNameLengthLimit)
				prometheus.Spec.EnforcedLabelNameLengthLimit = &i
			}

			podMonitor := monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}
			if tc.labelNameLengthLimit >= 0 {
				podMonitor.Spec.LabelNameLengthLimit = uint64(tc.labelNameLengthLimit)
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestLabelValueLengthLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  proxy_url: socks://myproxy:9095
  params:
    module:
    - http_2xx
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: default
      static: label
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/testprobe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  proxy_url: socks://myproxy:9095
  params:
    module:
    - http_2xx
  label_value_length_limit: %d
  static_configs:
  - targets:
    - prometheus.io
    - promcon.io
    labels:
      namespace: default
      static: label
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - source_labels:
    - __address__
    target_label: __param_target
  - source_labels:
    - __param_target
    target_label: instance
  - target_label: __address__
    replacement: blackbox.exporter.io
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version                       string
		enforcedLabelValueLengthLimit int
		labelValueLengthLimit         int
		expected                      string
	}{
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: -1,
			labelValueLengthLimit:         -1,
			expected:                      expectNoLimit,
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: -1,
			labelValueLengthLimit:         -1,
			expected:                      expectNoLimit,
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         -1,
			expected:                      expectNoLimit,
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         -1,
			expected:                      fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         2000,
			expected:                      expectNoLimit,
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         2000,
			expected:                      fmt.Sprintf(expectLimit, 1000),
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         500,
			expected:                      expectNoLimit,
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         500,
			expected:                      fmt.Sprintf(expectLimit, 500),
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelValueLengthLimit(%d) labelValueLengthLimit(%d)", tc.version, tc.enforcedLabelValueLengthLimit, tc.labelValueLengthLimit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}
			if tc.enforcedLabelValueLengthLimit >= 0 {
				i := uint64(tc.enforcedLabelValueLengthLimit)
				prometheus.Spec.EnforcedLabelValueLengthLimit = &i
			}

			probe := monitoringv1.Probe{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testprobe1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						Scheme:   "http",
						URL:      "blackbox.exporter.io",
						Path:     "/probe",
						ProxyURL: "socks://myproxy:9095",
					},
					Module: "http_2xx",
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
							Targets: []string{
								"prometheus.io",
								"promcon.io",
							},
							Labels: map[string]string{
								"static": "label",
							},
						},
					},
				},
			}
			if tc.labelValueLengthLimit >= 0 {
				probe.Spec.LabelValueLengthLimit = uint64(tc.labelValueLengthLimit)
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				nil,
				map[string]*monitoringv1.Probe{
					"testprobe1": &probe,
				},
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestBodySizeLimits(t *testing.T) {
	expectNoLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectLimit := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  body_size_limit: %s
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version               string
		enforcedBodySizeLimit monitoringv1.ByteSize
		expected              string
		expectedErr           error
	}{
		{
			version:               "v2.27.0",
			enforcedBodySizeLimit: "1000MB",
			expected:              expectNoLimit,
		},
		{
			version:               "v2.28.0",
			enforcedBodySizeLimit: "1000MB",
			expected:              fmt.Sprintf(expectLimit, "1000MB"),
		},
		{
			version:               "v2.28.0",
			enforcedBodySizeLimit: "",
			expected:              expectNoLimit,
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedBodySizeLimit(%s)", tc.version, tc.enforcedBodySizeLimit), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}
			if tc.enforcedBodySizeLimit != "" {
				i := tc.enforcedBodySizeLimit
				prometheus.Spec.EnforcedBodySizeLimit = i
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)
			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)

			if tc.expectedErr != nil {
				if tc.expectedErr.Error() != err.Error() {
					t.Logf("\n%s", pretty.Compare(tc.expectedErr.Error(), err.Error()))
					t.Fatal("expected error and actual error do not match")
				}
				return
			}
			result := string(cfg)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestMatchExpressionsServiceMonitor(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-value",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "alpha",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"beta", "gamma"},
							},
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: ns-value/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_alpha
    - __meta_kubernetes_service_labelpresent_alpha
    regex: (beta|gamma);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)
	if expected != result {
		diff := cmp.Diff(expected, result)
		t.Fatalf("expected Prometheus configuration and actual configuration do not match for enforced namespace label test:\n%s", diff)
	}
}

func TestServiceMonitorEndpointFollowRedirects(t *testing.T) {
	expectedWithRedirectsUnsupported := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectedWithRedirectsDisabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  follow_redirects: false
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`
	expectedWithRedirectsEnabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  follow_redirects: true
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version         string
		expected        string
		followRedirects bool
	}{
		{
			version:         "v2.25.0",
			followRedirects: false,
			expected:        expectedWithRedirectsUnsupported,
		},
		{
			version:         "v2.25.0",
			followRedirects: true,
			expected:        expectedWithRedirectsUnsupported,
		},
		{
			version:         "v2.28.0",
			followRedirects: true,
			expected:        expectedWithRedirectsEnabled,
		},
		{
			version:         "v2.28.0",
			followRedirects: false,
			expected:        expectedWithRedirectsDisabled,
		},
	} {
		t.Run(fmt.Sprintf("%s TestServiceMonitorEndpointFollowRedirects(%t)", tc.version, tc.followRedirects), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:            "web",
							Interval:        "30s",
							FollowRedirects: swag.Bool(tc.followRedirects),
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestPodMonitorEndpointFollowRedirects(t *testing.T) {
	expectedWithRedirectsUnsupported := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectedWithRedirectsDisabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  follow_redirects: false
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`
	expectedWithRedirectsEnabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  follow_redirects: true
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version         string
		expected        string
		followRedirects bool
	}{
		{
			version:         "v2.25.0",
			followRedirects: false,
			expected:        expectedWithRedirectsUnsupported,
		},
		{
			version:         "v2.25.0",
			followRedirects: true,
			expected:        expectedWithRedirectsUnsupported,
		},
		{
			version:         "v2.28.0",
			followRedirects: true,
			expected:        expectedWithRedirectsEnabled,
		},
		{
			version:         "v2.28.0",
			followRedirects: false,
			expected:        expectedWithRedirectsDisabled,
		},
	} {
		t.Run(fmt.Sprintf("%s TestServiceMonitorEndpointFollowRedirects(%t)", tc.version, tc.followRedirects), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			podMonitor := monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "pod-monitor-ns",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:            "web",
							Interval:        "30s",
							FollowRedirects: swag.Bool(tc.followRedirects),
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestServiceMonitorEndpointEnableHttp2(t *testing.T) {
	expectedWithHTTP2Unsupported := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectedWithHTTP2Disabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  enable_http2: false
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`
	expectedWithHTTP2Enabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  enable_http2: true
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version     string
		expected    string
		enableHTTP2 bool
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Enabled,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Disabled,
		},
	} {
		t.Run(fmt.Sprintf("%s TestServiceMonitorEndpointEnableHttp2(%t)", tc.version, tc.enableHTTP2), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			serviceMonitor := monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:        "web",
							Interval:    "30s",
							EnableHttp2: swag.Bool(tc.enableHTTP2),
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestPodMonitorPhaseFilter(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:        "v2.9.0",
				ScrapeInterval: "30s",
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
			EvaluationInterval: "30s",
		},
	}
	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		nil,
		map[string]*monitoringv1.PodMonitor{
			"testpodmonitor1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							FilterRunning: swag.Bool(false),
							Port:          "test",
						},
					},
				},
			},
		},
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/default/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - default
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: test
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: default/testpodmonitor1
  - target_label: endpoint
    replacement: test
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	result := string(cfg)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Logf("\n%s", diff)
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestPodMonitorEndpointEnableHttp2(t *testing.T) {
	expectedWithHTTP2Unsupported := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	expectedWithHTTP2Disabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  enable_http2: false
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`
	expectedWithHTTP2Enabled := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: podMonitor/pod-monitor-ns/testpodmonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - pod-monitor-ns
  scrape_interval: 30s
  enable_http2: true
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - action: keep
    source_labels:
    - __meta_kubernetes_pod_container_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - target_label: job
    replacement: pod-monitor-ns/testpodmonitor1
  - target_label: endpoint
    replacement: web
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs: []
`

	for _, tc := range []struct {
		version     string
		expected    string
		enableHTTP2 bool
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Unsupported,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			expected:    expectedWithHTTP2Enabled,
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			expected:    expectedWithHTTP2Disabled,
		},
	} {
		t.Run(fmt.Sprintf("%s TestServiceMonitorEndpointEnableHttp2(%t)", tc.version, tc.enableHTTP2), func(t *testing.T) {
			prometheus := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        tc.version,
						ScrapeInterval: "30s",
						ServiceMonitorSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"group": "group1",
							},
						},
					},
					EvaluationInterval: "30s",
				},
			}

			podMonitor := monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpodmonitor1",
					Namespace: "pod-monitor-ns",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:        "web",
							Interval:    "30s",
							EnableHttp2: swag.Bool(tc.enableHTTP2),
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, &prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				prometheus.Spec.EvaluationInterval,
				prometheus.Spec.QueryLogFile,
				prometheus.Spec.RuleSelector,
				prometheus.Spec.Exemplars,
				prometheus.Spec.TSDB,
				prometheus.Spec.Alerting,
				prometheus.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfg)

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Logf("\n%s", diff)
				t.Fatal("expected Prometheus configuration and actual configuration do not match")
			}
		})
	}
}

func TestStorageSettingMaxExemplars(t *testing.T) {
	for _, tc := range []struct {
		Scenario       string
		Prometheus     *monitoringv1.Prometheus
		ExpectedConfig string
	}{
		{
			Scenario: "Exemplars maxSize is set to 5000000",
			Prometheus: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					Exemplars: &monitoringv1.Exemplars{
						MaxSize: getInt64Pointer(5000000),
					},
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ScrapeInterval: "30s",
					},
					EvaluationInterval: "30s",
				},
			},
			ExpectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
storage:
  exemplars:
    max_exemplars: 5000000
`,
		},
		{
			Scenario: "max_exemplars is not set if version is less than v2.29.0",
			Prometheus: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        "v2.28.0",
						ScrapeInterval: "30s",
					},
					Exemplars: &monitoringv1.Exemplars{
						MaxSize: getInt64Pointer(5000000),
					},
					EvaluationInterval: "30s",
				},
			},
			ExpectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			Scenario: "Exemplars maxSize is not set",
			Prometheus: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ScrapeInterval: "30s",
						EnableFeatures: []string{"exemplar-storage"},
					},
					EvaluationInterval: "30s",
				},
			},
			ExpectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.Scenario), func(t *testing.T) {
			cg := mustNewConfigGenerator(t, tc.Prometheus)

			cfg, err := cg.GenerateServerConfiguration(
				tc.Prometheus.Spec.EvaluationInterval,
				tc.Prometheus.Spec.QueryLogFile,
				tc.Prometheus.Spec.RuleSelector,
				tc.Prometheus.Spec.Exemplars,
				tc.Prometheus.Spec.TSDB,
				tc.Prometheus.Spec.Alerting,
				tc.Prometheus.Spec.RemoteRead,
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			actualConfig := string(cfg)

			if tc.ExpectedConfig != actualConfig {
				t.Logf("\n%s", pretty.Compare(tc.ExpectedConfig, actualConfig))
				t.Fatal("expected prometheus configuration with storage and actual configuration do not match")
			}
		})
	}
}

func TestTSDBConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		p        *monitoringv1.Prometheus
		expected string
	}{
		{
			name: "no TSDB config",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ScrapeInterval: "30s",
					},
					EvaluationInterval: "30s",
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			name: "TSDB config < v2.39.0",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        "v2.38.0",
						ScrapeInterval: "30s",
					},
					EvaluationInterval: "30s",
					TSDB: monitoringv1.TSDBSpec{
						OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
					},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
`,
		},
		{
			name: "TSDB config >= v2.39.0",
			p: &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:        "v2.39.0",
						ScrapeInterval: "30s",
					},
					EvaluationInterval: "30s",
					TSDB: monitoringv1.TSDBSpec{
						OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
					},
				},
			},
			expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
storage:
  tsdb:
    out_of_order_time_window: 10m
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cg := mustNewConfigGenerator(t, tc.p)

			cfg, err := cg.GenerateServerConfiguration(
				tc.p.Spec.EvaluationInterval,
				tc.p.Spec.QueryLogFile,
				tc.p.Spec.RuleSelector,
				tc.p.Spec.Exemplars,
				tc.p.Spec.TSDB,
				tc.p.Spec.Alerting,
				tc.p.Spec.RemoteRead,
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			actualConfig := string(cfg)

			if tc.expected != actualConfig {
				t.Logf("\n%s", pretty.Compare(tc.expected, actualConfig))
				t.Fatal("expected TSDB configuration doesn't match with actual configuration")
			}
		})
	}
}

func TestGenerateRelabelConfig(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-relabel",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeInterval: "30s",
			},
			EvaluationInterval: "30s",
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		map[string]*monitoringv1.ServiceMonitor{
			"test": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "https-metrics",
							Interval: "30s",
							MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "Drop",
									Regex:        "container_fs*",
									SourceLabels: []monitoringv1.LabelName{"__name__"},
								},
							},
							RelabelConfigs: []*monitoringv1.RelabelConfig{
								{
									Action:       "Uppercase",
									SourceLabels: []monitoringv1.LabelName{"instance"},
									TargetLabel:  "instance",
								},
								{
									Action:       "Replace",
									Regex:        "(.+)(?::d+)",
									Replacement:  "$1:9537",
									SourceLabels: []monitoringv1.LabelName{"__address__"},
									TargetLabel:  "__address__",
								},
								{
									Action:      "Replace",
									Replacement: "crio",
									TargetLabel: "job",
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: test-relabel/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: serviceMonitor/default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    - __meta_kubernetes_service_labelpresent_foo
    regex: (bar);true
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: https-metrics
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container
  - action: drop
    source_labels:
    - __meta_kubernetes_pod_phase
    regex: (Failed|Succeeded)
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: https-metrics
  - source_labels:
    - instance
    target_label: instance
    action: uppercase
  - source_labels:
    - __address__
    target_label: __address__
    regex: (.+)(?::d+)
    replacement: $1:9537
    action: replace
  - target_label: job
    replacement: crio
    action: replace
  - source_labels:
    - __address__
    target_label: __tmp_hash
    modulus: 1
    action: hashmod
  - source_labels:
    - __tmp_hash
    regex: $(SHARD)
    action: keep
  metric_relabel_configs:
  - source_labels:
    - __name__
    regex: container_fs*
    action: drop
`

	result := string(cfg)
	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func getInt64Pointer(i int64) *int64 {
	return &i
}
