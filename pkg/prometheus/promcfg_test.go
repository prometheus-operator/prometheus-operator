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
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func defaultPrometheus() *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
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
}

func mustNewConfigGenerator(t *testing.T, p *monitoringv1.Prometheus) *ConfigGenerator {
	t.Helper()

	if p == nil {
		p = &monitoringv1.Prometheus{}
	}
	logger := level.NewFilter(log.NewLogfmtLogger(os.Stderr), level.AllowWarn())

	cg, err := NewConfigGenerator(log.With(logger, "test", t.Name()), p, false)
	require.NoError(t, err)

	return cg
}

func TestConfigGeneration(t *testing.T) {
	for i := range operator.PrometheusCompatibilityMatrix {
		v := operator.PrometheusCompatibilityMatrix[i]
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			cfg, err := generateTestConfig(t, v)
			require.NoError(t, err)

			reps := 1000
			if testing.Short() {
				reps = 100
			}

			for i := 0; i < reps; i++ {
				testcfg, err := generateTestConfig(t, v)
				require.NoError(t, err)

				require.Equal(t, testcfg, cfg)
			}
		})
	}
}

func TestGlobalSettings(t *testing.T) {
	var (
		expectedBodySizeLimit         monitoringv1.ByteSize = "1000MB"
		expectedSampleLimit           uint64                = 10000
		expectedTargetLimit           uint64                = 1000
		expectedLabelLimit            uint64                = 50
		expectedLabelNameLengthLimit  uint64                = 40
		expectedLabelValueLengthLimit uint64                = 30
	)

	for _, tc := range []struct {
		Scenario              string
		EvaluationInterval    monitoringv1.Duration
		ScrapeInterval        monitoringv1.Duration
		ScrapeTimeout         monitoringv1.Duration
		ExternalLabels        map[string]string
		QueryLogFile          string
		Version               string
		BodySizeLimit         *monitoringv1.ByteSize
		SampleLimit           *uint64
		TargetLimit           *uint64
		LabelLimit            *uint64
		LabelNameLengthLimit  *uint64
		LabelValueLengthLimit *uint64
		Expected              string
		ExpectError           bool
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
		{
			Scenario:           "valid global limits",
			Version:            "v2.45.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			BodySizeLimit:      &expectedBodySizeLimit,
			SampleLimit:        &expectedSampleLimit,
			TargetLimit:        &expectedTargetLimit,
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: /
    prometheus_replica: $(POD_NAME)
  body_size_limit: 1000MB
  sample_limit: 10000
  target_limit: 1000
scrape_configs: []
`,
		},
		{
			Scenario:              "valid global config with label limits",
			Version:               "v2.45.0",
			ScrapeInterval:        "30s",
			EvaluationInterval:    "30s",
			BodySizeLimit:         &expectedBodySizeLimit,
			SampleLimit:           &expectedSampleLimit,
			TargetLimit:           &expectedTargetLimit,
			LabelLimit:            &expectedLabelLimit,
			LabelNameLengthLimit:  &expectedLabelNameLengthLimit,
			LabelValueLengthLimit: &expectedLabelValueLengthLimit,
			Expected: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: /
    prometheus_replica: $(POD_NAME)
  body_size_limit: 1000MB
  sample_limit: 10000
  target_limit: 1000
  label_limit: 50
  label_name_length_limit: 40
  label_value_length_limit: 30
scrape_configs: []
`,
		},
	} {

		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ScrapeInterval:        tc.ScrapeInterval,
					ScrapeTimeout:         tc.ScrapeTimeout,
					ExternalLabels:        tc.ExternalLabels,
					Version:               tc.Version,
					TracingConfig:         nil,
					BodySizeLimit:         tc.BodySizeLimit,
					SampleLimit:           tc.SampleLimit,
					TargetLimit:           tc.TargetLimit,
					LabelLimit:            tc.LabelLimit,
					LabelNameLengthLimit:  tc.LabelNameLengthLimit,
					LabelValueLengthLimit: tc.LabelValueLengthLimit,
				},
				EvaluationInterval: tc.EvaluationInterval,
				QueryLogFile:       tc.QueryLogFile,
			},
		}

		cg := mustNewConfigGenerator(t, p)
		t.Run(fmt.Sprintf("case %s", tc.Scenario), func(t *testing.T) {
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)

			if tc.ExpectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.Expected, string(cfg))
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
		require.NoError(t, err)
		require.Equal(t, tc.Expected, string(s))
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
	require.NoError(t, err)

	expected := `kubernetes_sd_configs:
- role: pod
  namespaces:
    names:
    - test
  attach_metadata:
    node: true
`

	require.Equal(t, expected, string(s))
}

func TestProbeStaticTargetsConfigGenerationWithLabelEnforce(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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
  - regex: noisy_labels.*
    action: labeldrop
  - target_label: namespace
    replacement: default
`

	require.Equal(t, expected, string(cfg))
}

func TestProbeStaticTargetsConfigGenerationWithJobName(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestProbeStaticTargetsConfigGenerationWithoutModule(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestProbeIngressSDConfigGeneration(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestProbeIngressSDConfigGenerationWithShards(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Shards = pointer.Int32(2)

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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
	require.Equal(t, expected, string(cfg))
}

func TestProbeIngressSDConfigGenerationWithLabelEnforce(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.EnforcedNamespaceLabel = "namespace"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
		require.NoError(t, err)
		require.Equal(t, tc.expected, string(s))
	}
}

func TestAlertmanagerBearerToken(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:            "alertmanager-main",
				Namespace:       "default",
				Port:            intstr.FromString("web"),
				BearerTokenFile: "/some/file/on/disk",
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
			context.Background(),
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

		require.Equal(t, tc.expectedConfig, string(cfg))
	}
}

func TestAlertmanagerAPIVersion(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:       "alertmanager-main",
				Namespace:  "default",
				Port:       intstr.FromString("web"),
				APIVersion: "v2",
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestAlertmanagerTimeoutConfig(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:       "alertmanager-main",
				Namespace:  "default",
				Port:       intstr.FromString("web"),
				APIVersion: "v2",
				Timeout:    (*monitoringv1.Duration)(pointer.String("60s")),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.Version = tc.version
			p.Spec.Alerting = &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:        "alertmanager-main",
						Namespace:   "default",
						Port:        intstr.FromString("web"),
						APIVersion:  "v2",
						EnableHttp2: swag.Bool(tc.enableHTTP2),
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestAdditionalScrapeConfigs(t *testing.T) {
	getCfg := func(shards *int32) string {
		p := defaultPrometheus()
		p.Spec.Shards = shards

		cg := mustNewConfigGenerator(t, p)
		cfg, err := cg.GenerateServerConfiguration(
			context.Background(),
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
		require.NoError(t, err)

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
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.result, tt.expected)
		},
		)
	}
}

func TestAdditionalAlertRelabelConfigs(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:      "alertmanager-main",
				Namespace: "default",
				Port:      intstr.FromString("web"),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestNoEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
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

	require.Equal(t, expected, string(cfg))
}

func TestServiceMonitorWithEndpointSliceEnable(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "ns-key"

	cg := mustNewConfigGenerator(t, p)
	cg.endpointSliceSupported = true
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestEnforcedNamespaceLabelPodMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "ns-key"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	expected := `global:
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

	require.Equal(t, expected, string(cfg))
}

func TestEnforcedNamespaceLabelOnExcludedPodMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.ExcludedFromEnforcement = []monitoringv1.ObjectReference{
		{
			Namespace: "pod-monitor-ns",
			Group:     monitoring.GroupName,
			Resource:  monitoringv1.PodMonitorName,
			Name:      "testpodmonitor1",
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	expected := `global:
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

	require.Equal(t, expected, string(cfg))
}

func TestEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "ns-key"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestEnforcedNamespaceLabelOnExcludedServiceMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.ExcludedFromEnforcement = []monitoringv1.ObjectReference{
		{
			Namespace: "service-monitor-ns",
			Group:     monitoring.GroupName,
			Resource:  monitoringv1.ServiceMonitorName,
			Name:      "", // exclude all servicemonitors in this namespace
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
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

	require.Equal(t, expected, string(cfg))
}

func TestAdditionalAlertmanagers(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:      "alertmanager-main",
				Namespace: "default",
				Port:      intstr.FromString("web"),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestSettingHonorTimestampsInServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestSettingHonorTimestampsInPodMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestHonorTimestampsOverriding(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.OverrideHonorTimestamps = true

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestSettingHonorLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestHonorLabelsOverriding(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.OverrideHonorLabels = true

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestTargetLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
		sMons             map[string]*monitoringv1.ServiceMonitor
		pMons             map[string]*monitoringv1.PodMonitor
		probes            map[string]*monitoringv1.Probe
		oauth2Credentials map[string]assets.OAuth2Credentials
		expectedCfg       string
	}{
		{
			name: "service monitor with oauth2",
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
			p := defaultPrometheus()

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				tt.sMons,
				tt.pMons,
				tt.probes,
				nil,
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
			require.NoError(t, err)
			result := string(cfg)

			if !strings.Contains(result, tt.expectedCfg) {
				t.Fatalf("expected Prometheus configuration to contain:\n %s\nFull config:\n %s", tt.expectedCfg, result)
			}
		})
	}
}

func TestPodTargetLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestPodTargetLabelsFromPodMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestPodTargetLabelsFromPodMonitorAndGlobal(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.PodTargetLabels = []string{"global"}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

func TestEmptyEndpointPorts(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
		context.Background(),
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
		require.Equal(t, "honor_labels", k)
		require.Equal(t, tc.Expected, v)
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
			require.Equal(t, tc.Expected, string(hl))
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
			p := defaultPrometheus()
			if tc.enforcedLimit >= 0 {
				i := uint64(tc.enforcedLimit)
				p.Spec.EnforcedSampleLimit = &i
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
				sampleLimit := uint64(tc.limit)
				serviceMonitor.Spec.SampleLimit = &sampleLimit
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLimit >= 0 {
				i := uint64(tc.enforcedLimit)
				p.Spec.EnforcedTargetLimit = &i
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
				limit := uint64(tc.limit)
				serviceMonitor.Spec.TargetLimit = &limit
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestRemoteReadConfig(t *testing.T) {
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
			version: "v2.25.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:             "http://example.com",
				FollowRedirects: pointer.Bool(true),
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
				URL:             "http://example.com",
				FollowRedirects: pointer.Bool(false),
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
  follow_redirects: false
`,
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: pointer.Bool(true),
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
				FilterExternalLabels: pointer.Bool(false),
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
				FilterExternalLabels: pointer.Bool(true),
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version
			p.Spec.RemoteRead = []monitoringv1.RemoteReadSpec{tc.remoteRead}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestRemoteWriteConfig(t *testing.T) {
	sendNativeHistograms := true
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
		{
			version: "v2.43.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:                  "http://example.com",
				SendNativeHistograms: &sendNativeHistograms,
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
  send_native_histograms: true
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
		{
			version: "v2.39.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:                  "http://example.com",
				SendNativeHistograms: &sendNativeHistograms,
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version
			p.Spec.CommonPrometheusFields.RemoteWrite = []monitoringv1.RemoteWriteSpec{tc.remoteWrite}
			p.Spec.CommonPrometheusFields.Secrets = []string{"sigv4-secret"}

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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				store,
				nil,
				nil,
				nil,
				nil)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelLimit >= 0 {
				p.Spec.EnforcedLabelLimit = pointer.Uint64(uint64(tc.enforcedLabelLimit))
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
				labelLimit := uint64(tc.labelLimit)
				serviceMonitor.Spec.LabelLimit = &labelLimit
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelNameLengthLimit >= 0 {
				p.Spec.EnforcedLabelNameLengthLimit = pointer.Uint64(uint64(tc.enforcedLabelNameLengthLimit))
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
				labelNameLengthLimit := uint64(tc.labelNameLengthLimit)
				podMonitor.Spec.LabelNameLengthLimit = &labelNameLengthLimit
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelValueLengthLimit >= 0 {
				p.Spec.EnforcedLabelValueLengthLimit = pointer.Uint64(uint64(tc.enforcedLabelValueLengthLimit))
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
				labelValueLengthLimit := uint64(tc.labelValueLengthLimit)
				probe.Spec.LabelValueLengthLimit = &labelValueLengthLimit
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
					"testprobe1": &probe,
				},
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedBodySizeLimit != "" {
				p.Spec.EnforcedBodySizeLimit = tc.enforcedBodySizeLimit
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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestMatchExpressionsServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

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

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestPodMonitorPhaseFilter(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
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
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				p.Spec.RuleSelector,
				p.Spec.Exemplars,
				p.Spec.TSDB,
				p.Spec.Alerting,
				p.Spec.RemoteRead,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestStorageSettingMaxExemplars(t *testing.T) {
	for _, tc := range []struct {
		Scenario       string
		Version        string
		Exemplars      *monitoringv1.Exemplars
		ExpectedConfig string
	}{
		{
			Scenario: "Exemplars maxSize is set to 5000000",
			Exemplars: &monitoringv1.Exemplars{
				MaxSize: pointer.Int64(5000000),
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
			Version:  "v2.28.0",
			Exemplars: &monitoringv1.Exemplars{
				MaxSize: pointer.Int64(5000000),
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
			p := defaultPrometheus()
			if tc.Version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.Version
			}
			if tc.Exemplars != nil {
				p.Spec.Exemplars = tc.Exemplars
			}
			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedConfig, string(cfg))
		})
	}
}

func TestTSDBConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		p        *monitoringv1.Prometheus
		version  string
		tsdb     *monitoringv1.TSDBSpec
		expected string
	}{
		{
			name: "no TSDB config",
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
			name:    "TSDB config < v2.39.0",
			version: "v2.38.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
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
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
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
			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.tsdb != nil {
				p.Spec.TSDB = *tc.tsdb
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(cfg))
		})
	}
}

func TestGenerateRelabelConfig(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		context.Background(),
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
		nil,
		&assets.Store{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

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

	require.Equal(t, expected, string(cfg))
}

// When adding new test cases the developer should specify a name, a Probe Spec
// (pbSpec) and an expectedConfig. (Optional) It's also possible to specify a
// function that modifies the default Prometheus CR used if necessary for the test
// case.
func TestProbeSpecConfig(t *testing.T) {
	for _, tc := range []struct {
		name        string
		patchProm   func(*monitoringv1.Prometheus)
		pbSpec      monitoringv1.ProbeSpec
		expectedCfg string
	}{
		{
			name:   "empty_probe",
			pbSpec: monitoringv1.ProbeSpec{},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/probe1
  honor_timestamps: true
  metrics_path: ""
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
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
`,
		},
		{
			name: "prober_spec",
			pbSpec: monitoringv1.ProbeSpec{
				ProberSpec: monitoringv1.ProberSpec{
					Scheme:   "http",
					URL:      "example.com",
					Path:     "/probe",
					ProxyURL: "socks://myproxy:9095",
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/probe1
  honor_timestamps: true
  metrics_path: /probe
  scheme: http
  proxy_url: socks://myproxy:9095
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
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
`,
		},
		{
			name: "targets_static_config",
			pbSpec: monitoringv1.ProbeSpec{
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
				}},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/probe1
  honor_timestamps: true
  metrics_path: ""
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
    replacement: ""
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
`,
		},
		{
			name: "module_config",
			pbSpec: monitoringv1.ProbeSpec{
				Module: "http_2xx",
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: probe/default/probe1
  honor_timestamps: true
  metrics_path: ""
  params:
    module:
    - http_2xx
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
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
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pbs := map[string]*monitoringv1.Probe{
				"probe1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "probe1",
						Namespace: "default",
					},
					Spec: tc.pbSpec,
				},
			}

			p := defaultPrometheus()
			if tc.patchProm != nil {
				tc.patchProm(p)
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				nil,
				nil,
				p.Spec.TSDB,
				nil,
				nil,
				nil,
				nil,
				pbs,
				nil,
				&assets.Store{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expectedCfg, string(cfg))
		})

	}
}

// When adding new test cases the developer should specify a name, a ScrapeConfig Spec
// (scSpec) and an expectedConfig. (Optional) It's also possible to specify a
// function (patchProm) that modifies the default Prometheus CR used if necessary for the test
// case.
func TestScrapeConfigSpecConfig(t *testing.T) {
	refreshInterval := monitoringv1.Duration("5m")
	for _, tc := range []struct {
		name        string
		patchProm   func(*monitoringv1.Prometheus)
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		expectedCfg string
	}{
		{
			name:   "empty_scrape_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
`,
		},
		{
			name: "static_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"http://localhost:9100"},
						Labels: map[monitoringv1.LabelName]string{
							"label1": "value1",
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  static_configs:
  - targets:
    - http://localhost:9100
    labels:
      label1: value1
`,
		},
		{
			name: "file_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
					{
						Files:           []monitoringv1alpha1.SDFile{"/tmp/myfile.json"},
						RefreshInterval: &refreshInterval,
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  file_sd_configs:
  - files:
    - /tmp/myfile.json
    refresh_interval: 5m
`,
		},
		{
			name: "http_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL:             "http://localhost:9100/sd.json",
						RefreshInterval: &refreshInterval,
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  http_sd_configs:
  - url: http://localhost:9100/sd.json
    refresh_interval: 5m
`,
		},
		{
			name: "metrics_path",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricsPath: pointer.String("/metrics"),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  metrics_path: /metrics
`,
		},
		{
			name: "empty_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				RelabelConfigs: []*monitoringv1.RelabelConfig{},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
`,
		},
		{
			name: "non_empty_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				RelabelConfigs: []*monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						Regex:        "(.+)(?::d+)",
						Replacement:  "$1:9537",
						SourceLabels: []monitoringv1.LabelName{"__address__"},
						TargetLabel:  "__address__",
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  relabel_configs:
  - source_labels:
    - job
    target_label: __tmp_prometheus_job_name
  - source_labels:
    - __address__
    target_label: __address__
    regex: (.+)(?::d+)
    replacement: $1:9537
    action: replace
`,
		},
		{
			name: "honor_timestamp",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HonorTimestamps: pointer.Bool(true),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  honor_timestamps: true
`,
		},
		{
			name: "honor_labels",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HonorLabels: pointer.Bool(true),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  honor_labels: true
`,
		},
		{
			name: "basic_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
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
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
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
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  basic_auth:
    username: scrape-bob
    password: scrape-alice
  http_sd_configs:
  - url: http://localhost:9100/sd.json
    basic_auth:
      username: http-sd-bob
      password: http-sd-alice
`,
		},
		{
			name: "authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				Authorization: &monitoringv1.SafeAuthorization{
					Credentials: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "key",
						},
					},
				},
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "key",
								},
							},
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  authorization:
    type: Bearer
    credentials: scrape-secret
  http_sd_configs:
  - url: http://localhost:9100/sd.json
    authorization:
      type: Bearer
      credentials: http-sd-secret
`,
		},
		{
			name: "tlsconfig",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				TLSConfig: &monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret-ca-global",
							},
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret-cert",
							},
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key",
					},
				},
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							InsecureSkipVerify: true,
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret-ca-http",
									},
								},
							},
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  tls_config:
    insecure_skip_verify: false
    ca_file: /etc/prometheus/certs/secret_default_secret-ca-global_
    cert_file: /etc/prometheus/certs/secret_default_secret-cert_
    key_file: /etc/prometheus/certs/secret_default_secret_key
  http_sd_configs:
  - url: http://localhost:9100/sd.json
    tls_config:
      insecure_skip_verify: true
      ca_file: /etc/prometheus/certs/secret_default_secret-ca-http_
`,
		},
		{
			name: "scheme",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				Scheme: pointer.String("HTTPS"),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  scheme: https
`,
		},
		{
			name: "limits",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				SampleLimit:           pointer.Uint64(10000),
				TargetLimit:           pointer.Uint64(1000),
				LabelLimit:            pointer.Uint64(50),
				LabelNameLengthLimit:  pointer.Uint64(40),
				LabelValueLengthLimit: pointer.Uint64(30),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  sample_limit: 10000
  target_limit: 1000
  label_limit: 50
  label_name_length_limit: 40
  label_value_length_limit: 30
`,
		},
		{
			name: "params",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricsPath: pointer.String("/federate"),
				Params:      map[string][]string{"match[]": {"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"}},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  metrics_path: /federate
  params:
    match[]:
    - '{job="prometheus"}'
    - '{__name__=~"job:.*"}'
`,
		},
		{
			name: "scrape_interval",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeInterval: (*monitoringv1.Duration)(pointer.String("15s")),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  scrape_interval: 15s
`,
		},
		{
			name: "scrape_timeout",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeTimeout: (*monitoringv1.Duration)(pointer.String("10s")),
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  scrape_timeout: 10s
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scs := map[string]*monitoringv1alpha1.ScrapeConfig{
				"sc": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testscrapeconfig1",
						Namespace: "default",
					},
					Spec: tc.scSpec,
				},
			}

			p := defaultPrometheus()
			if tc.patchProm != nil {
				tc.patchProm(p)
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				nil,
				nil,
				p.Spec.TSDB,
				nil,
				nil,
				nil,
				nil,
				nil,
				scs,
				&assets.Store{
					BasicAuthAssets: map[string]assets.BasicAuthCredentials{
						"scrapeconfig/default/testscrapeconfig1": {
							Username: "scrape-bob",
							Password: "scrape-alice",
						},
						"scrapeconfig/default/testscrapeconfig1/httpsdconfig/0": {
							Username: "http-sd-bob",
							Password: "http-sd-alice",
						},
					},
					TokenAssets: map[string]assets.Token{
						"scrapeconfig/auth/default/testscrapeconfig1":                assets.Token("scrape-secret"),
						"scrapeconfig/auth/default/testscrapeconfig1/httpsdconfig/0": assets.Token("http-sd-secret"),
					},
				},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expectedCfg, string(cfg))
		})

	}
}

func TestScrapeConfigSpecConfigWithConsulSD(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"proxy-header": []byte("value"),
				"token":        []byte("value"),
			},
		},
	)
	for _, tc := range []struct {
		name        string
		patchProm   func(*monitoringv1.Prometheus)
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		expectedCfg string
	}{
		{
			name: "consul_scrape_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:       "localhost",
						Datacenter:   pointer.String("we1"),
						Namespace:    pointer.String("observability"),
						Partition:    pointer.String("1"),
						Scheme:       pointer.String("https"),
						Services:     []string{"prometheus", "alertmanager"},
						Tags:         []string{"tag1"},
						TagSeparator: pointer.String(";"),
						NodeMeta: map[string]string{
							"service": "service_name",
							"name":    "node_name",
						},
						AllowStale:           pointer.Bool(false),
						RefreshInterval:      (*monitoringv1.Duration)(pointer.String("30s")),
						ProxyUrl:             pointer.String("http://no-proxy.com"),
						NoProxy:              pointer.String("0.0.0.0"),
						ProxyFromEnvironment: pointer.Bool(true),
						ProxyConnectHeader: map[string]v1.SecretKeySelector{
							"header": {
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "proxy-header",
							},
						},
						FollowRedirects: pointer.Bool(true),
						EnableHttp2:     pointer.Bool(true),
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "foo",
							},
							Key: "token",
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  consul_sd_configs:
  - server: localhost
    token: value
    datacenter: we1
    namespace: observability
    partition: "1"
    scheme: https
    services:
    - prometheus
    - alertmanager
    tags:
    - tag1
    tag_separator: ;
    node_meta:
      name: node_name
      service: service_name
    allow_stale: false
    refresh_interval: 30s
    proxy_url: http://no-proxy.com
    no_proxy: 0.0.0.0
    proxy_from_environment: true
    proxy_connect_header:
      header: value
    follow_redirects: true
    enable_http2: true
`,
		}, {
			name: "consul_scrape_config_basic_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
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
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  consul_sd_configs:
  - basic_auth:
      username: consul-sd-bob
      password: consul-sd-alice
    server: localhost:8500
`,
		}, {
			name: "consul_scrape_config_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "credential",
							},
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  consul_sd_configs:
  - authorization:
      type: Bearer
      credentials: authorization
    server: localhost:8500
`,
		}, {
			name: "consul_scrape_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
						Oauth2: &monitoringv1.OAuth2{
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
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  consul_sd_configs:
  - oauth2:
      client_id: client-id
      client_secret: client-secret
      token_url: http://test.url
      scopes:
      - scope 1
      - scope 2
      endpoint_params:
        param1: value1
        param2: value2
    server: localhost:8500
`,
		}, {
			name: "consul_scrape_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret-ca-global",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret-cert",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key",
							},
						},
					},
				},
			},
			expectedCfg: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs:
- job_name: scrapeconfig/default/testscrapeconfig1
  consul_sd_configs:
  - tls_config:
      insecure_skip_verify: false
      ca_file: /etc/prometheus/certs/secret_default_secret-ca-global_
      cert_file: /etc/prometheus/certs/secret_default_secret-cert_
      key_file: /etc/prometheus/certs/secret_default_secret_key
    server: localhost:8500
`,
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewStore(c.CoreV1(), c.CoreV1())
			store.BasicAuthAssets = map[string]assets.BasicAuthCredentials{
				"scrapeconfig/default/testscrapeconfig1/consulsdconfig/0": {
					Username: "consul-sd-bob",
					Password: "consul-sd-alice",
				},
			}

			store.OAuth2Assets = map[string]assets.OAuth2Credentials{
				"scrapeconfig/default/testscrapeconfig1/consulsdconfig/0": {
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			}

			store.TokenAssets = map[string]assets.Token{
				"scrapeconfig/auth/default/testscrapeconfig1/consulsdconfig/0": assets.Token("authorization"),
			}

			scs := map[string]*monitoringv1alpha1.ScrapeConfig{
				"sc": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testscrapeconfig1",
						Namespace: "default",
					},
					Spec: tc.scSpec,
				},
			}

			p := defaultPrometheus()
			if tc.patchProm != nil {
				tc.patchProm(p)
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
				p.Spec.EvaluationInterval,
				p.Spec.QueryLogFile,
				nil,
				nil,
				p.Spec.TSDB,
				nil,
				nil,
				nil,
				nil,
				nil,
				scs,
				store,
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, tc.expectedCfg, string(cfg))
		})

	}
}

func TestTracingConfig(t *testing.T) {
	samplingTwo := resource.MustParse("0.5")
	testCases := []struct {
		tracingConfig  *monitoringv1.PrometheusTracingConfig
		name           string
		expectedConfig string
		expectedErr    bool
	}{
		{
			name: "Config only with endpoint",
			tracingConfig: &monitoringv1.PrometheusTracingConfig{
				Endpoint: "https://otel-collector.default.svc.local:3333",
			},

			expectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
tracing:
  endpoint: https://otel-collector.default.svc.local:3333
`,
			expectedErr: false,
		},
		{
			tracingConfig: &monitoringv1.PrometheusTracingConfig{
				ClientType:       pointer.String("grpc"),
				Endpoint:         "https://otel-collector.default.svc.local:3333",
				SamplingFraction: &samplingTwo,
				Headers: map[string]string{
					"custom": "header",
				},
				Compression: pointer.String("gzip"),
				Timeout:     (*monitoringv1.Duration)(pointer.String("10s")),
				Insecure:    pointer.Bool(false),
			},
			name:        "Expect valid config",
			expectedErr: false,
			expectedConfig: `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
scrape_configs: []
tracing:
  endpoint: https://otel-collector.default.svc.local:3333
  client_type: grpc
  sampling_fraction: 0.5
  insecure: false
  headers:
    custom: header
  compression: gzip
  timeout: 10s
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()

			p.Spec.CommonPrometheusFields.TracingConfig = tc.tracingConfig

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				context.Background(),
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
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedConfig, string(cfg))
		})
	}
}
