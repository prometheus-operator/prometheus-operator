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
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/golden"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

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
				PodMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
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

func mustNewConfigGenerator(t *testing.T, p *monitoringv1.Prometheus, opts ...ConfigGeneratorOption) *ConfigGenerator {
	t.Helper()

	if p == nil {
		p = &monitoringv1.Prometheus{}
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	if p.Spec.ServiceDiscoveryRole != nil && *p.Spec.ServiceDiscoveryRole == monitoringv1.EndpointSliceRole {
		opts = append(opts, WithEndpointSliceSupport())
	}

	cg, err := NewConfigGenerator(logger.With("test", t.Name()), p, opts...)
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
		expectedRuleQueryOffset       monitoringv1.Duration = "30s"
		expectedSampleLimit           uint64                = 10000
		expectedTargetLimit           uint64                = 1000
		expectedLabelLimit            uint64                = 50
		expectedLabelNameLengthLimit  uint64                = 40
		expectedLabelValueLengthLimit uint64                = 30
		expectedkeepDroppedTargets    uint64                = 50
	)

	for _, tc := range []struct {
		Scenario                    string
		RuleQueryOffset             *monitoringv1.Duration
		EvaluationInterval          monitoringv1.Duration
		ScrapeInterval              monitoringv1.Duration
		ScrapeTimeout               monitoringv1.Duration
		ScrapeProtocols             []monitoringv1.ScrapeProtocol
		ExternalLabels              map[string]string
		PrometheusExternalLabelName *string
		ReplicaExternalLabelName    *string
		QueryLogFile                string
		ScrapeFailureLogFile        *string
		Version                     string
		BodySizeLimit               *monitoringv1.ByteSize
		SampleLimit                 *uint64
		TargetLimit                 *uint64
		LabelLimit                  *uint64
		LabelNameLengthLimit        *uint64
		LabelValueLengthLimit       *uint64
		KeepDroppedTargets          *uint64
		ExpectError                 bool
		Golden                      string
	}{
		{
			Scenario:           "valid config",
			Version:            "v2.15.2",
			ScrapeInterval:     "15s",
			EvaluationInterval: "30s",
			Golden:             "global_settings_valid_config_v2.15.2.golden",
		},
		{
			Scenario:       "invalid scrape timeout specified when scrape interval specified",
			Version:        "v2.30.0",
			ScrapeInterval: "30s",
			ScrapeTimeout:  "60s",
			Golden:         "invalid_scrape_timeout_specified_when_scrape_interval_specified.golden",
			ExpectError:    true,
		},
		{
			Scenario:           "valid scrape timeout along with valid scrape interval specified",
			Version:            "v2.15.2",
			ScrapeInterval:     "60s",
			ScrapeTimeout:      "10s",
			EvaluationInterval: "30s",
			Golden:             "valid_scrape_timeout_along_with_valid_scrape_interval_specified.golden",
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
			Golden: "external_label_specified.golden",
		},
		{
			Scenario:           "external label specified along with reserved labels",
			Version:            "v2.45.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			ExternalLabels: map[string]string{
				"prometheus_replica": "1",
				"prometheus":         "prometheus-k8s-1",
				"some-other-key":     "some-value",
			},
			PrometheusExternalLabelName: ptr.To("prometheus"),
			ReplicaExternalLabelName:    ptr.To("prometheus_replica"),
			Golden:                      "external_label_specified_along_with_reserved_labels.golden",
		},
		{
			Scenario:           "query log file",
			Version:            "v2.16.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			QueryLogFile:       "test.log",
			Golden:             "query_log_file.golden",
		},
		{
			Scenario:             "scrape_failure_log_file",
			Version:              "v2.55.0",
			ScrapeInterval:       "30s",
			EvaluationInterval:   "30s",
			ScrapeFailureLogFile: ptr.To("/tmp/file.log"),
			Golden:               "scrape_failure_log_file.golden",
		},
		{
			Scenario:             "scrape_failure_log_file_empty_path",
			Version:              "v2.55.0",
			ScrapeInterval:       "30s",
			EvaluationInterval:   "30s",
			ScrapeFailureLogFile: ptr.To("file.log"),
			Golden:               "scrape_failure_log_file_empty_path.golden",
		},
		{
			Scenario:             "scrape_failure_log_file_unsupported_version",
			Version:              "v2.54.0",
			ScrapeInterval:       "30s",
			EvaluationInterval:   "30s",
			ScrapeFailureLogFile: ptr.To("file.log"),
			Golden:               "scrape_failure_log_file_unsupported_version.golden",
		},
		{
			Scenario:           "valid global limits",
			Version:            "v2.45.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			BodySizeLimit:      &expectedBodySizeLimit,
			SampleLimit:        &expectedSampleLimit,
			TargetLimit:        &expectedTargetLimit,
			Golden:             "valid_global_limits.golden",
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
			Golden:                "valid_global_config_with_label_limits.golden",
		},
		{
			Scenario:           "valid global config with keep dropped targets",
			Version:            "v2.47.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			KeepDroppedTargets: &expectedkeepDroppedTargets,
			Golden:             "valid_global_config_with_keep_dropped_targets.golden",
		},
		{
			Scenario:           "valid global config with scrape protocols",
			Version:            "v2.49.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				monitoringv1.OpenMetricsText1_0_0,
				monitoringv1.OpenMetricsText0_0_1,
				monitoringv1.PrometheusProto,
				monitoringv1.PrometheusText0_0_4,
				monitoringv1.PrometheusText1_0_0,
			},
			Golden: "valid_global_config_with_scrape_protocols.golden",
		},
		{
			Scenario:           "valid global config with new scrape protocol",
			Version:            "v3.0.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				monitoringv1.PrometheusText1_0_0,
			},
			Golden: "valid_global_config_with_new_scrape_protocol.golden",
		},
		{
			Scenario:           "valid global config with unsupported scrape protocols",
			Version:            "v2.48.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				monitoringv1.PrometheusProto,
				monitoringv1.PrometheusText0_0_4,
				monitoringv1.OpenMetricsText0_0_1,
				monitoringv1.OpenMetricsText1_0_0,
			},
			Golden: "valid_global_config_with_unsupported_scrape_protocols.golden",
		},
		{
			Scenario:           "valid global config without rule query offset if prometheus version less required",
			Version:            "v2.52.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			RuleQueryOffset:    &expectedRuleQueryOffset,
			Golden:             "valid_global_config_without_rule_query_offset.golden",
		},
		{
			Scenario:           "valid global config with rule query offset if prometheus version meets the requirement",
			Version:            "v2.53.0",
			ScrapeInterval:     "30s",
			EvaluationInterval: "30s",
			RuleQueryOffset:    &expectedRuleQueryOffset,
			Golden:             "valid_global_config_with_rule_query_offset.golden",
		},
	} {

		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "test",
			},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ScrapeInterval:              tc.ScrapeInterval,
					ScrapeTimeout:               tc.ScrapeTimeout,
					ScrapeProtocols:             tc.ScrapeProtocols,
					ExternalLabels:              tc.ExternalLabels,
					PrometheusExternalLabelName: tc.PrometheusExternalLabelName,
					ReplicaExternalLabelName:    tc.ReplicaExternalLabelName,
					Version:                     tc.Version,
					TracingConfig:               nil,
					BodySizeLimit:               tc.BodySizeLimit,
					SampleLimit:                 tc.SampleLimit,
					TargetLimit:                 tc.TargetLimit,
					LabelLimit:                  tc.LabelLimit,
					LabelNameLengthLimit:        tc.LabelNameLengthLimit,
					LabelValueLengthLimit:       tc.LabelValueLengthLimit,
					KeepDroppedTargets:          tc.KeepDroppedTargets,
					ScrapeFailureLogFile:        tc.ScrapeFailureLogFile,
				},
				EvaluationInterval: tc.EvaluationInterval,
				RuleQueryOffset:    tc.RuleQueryOffset,
				QueryLogFile:       tc.QueryLogFile,
			},
		}

		cg := mustNewConfigGenerator(t, p)
		t.Run(fmt.Sprintf("case %s", tc.Scenario), func(t *testing.T) {
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
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

			golden.Assert(t, string(cfg), tc.Golden)
		})
	}
}

func TestNamespaceSetCorrectly(t *testing.T) {
	type testCase struct {
		ServiceMonitor           *monitoringv1.ServiceMonitor
		IgnoreNamespaceSelectors bool
		Golden                   string
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
						Node: ptr.To(true),
					},
				},
			},
			IgnoreNamespaceSelectors: false,
			Golden:                   "namespaces_from_MatchNames_are_returned_instead_of_the_current_namespace.golden",
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
			Golden:                   "Any_returns_an_empty_list_instead_of_the_current_namespace.golden",
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
			Golden:                   "Any_takes_precedence_over_MatchNames.golden",
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
			Golden:                   "IgnoreNamespaceSelectors_overrides_Any_and_MatchNames.golden",
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
				attachMetadata: tc.ServiceMonitor.Spec.AttachMetadata,
			}
		}

		c := cg.generateK8SSDConfig(tc.ServiceMonitor.Spec.NamespaceSelector, tc.ServiceMonitor.Namespace, nil, assets.NewTestStoreBuilder().ForNamespace(tc.ServiceMonitor.Namespace), kubernetesSDRoleEndpoint, attachMetaConfig)
		// Wrap partial K8s SD config in a full Prometheus config to satisfy promtool validation.
		fullConfig := yaml.MapSlice{
			{Key: "scrape_configs", Value: []yaml.MapSlice{
				{
					{Key: "job_name", Value: "k8s-sd-test"},
					c,
				},
			}},
		}

		s, err := yaml.Marshal(fullConfig)
		require.NoError(t, err)
		golden.Assert(t, string(s), tc.Golden)
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
				Node: ptr.To(true),
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
		attachMetadata: pm.Spec.AttachMetadata,
	}
	c := cg.generateK8SSDConfig(pm.Spec.NamespaceSelector, pm.Namespace, nil, assets.NewTestStoreBuilder().ForNamespace(pm.Namespace), kubernetesSDRolePod, attachMetadataConfig)

	// Wrap partial K8s SD config in a full Prometheus config to satisfy promtool validation.
	fullConfig := yaml.MapSlice{
		{Key: "scrape_configs", Value: []yaml.MapSlice{
			{
				{Key: "job_name", Value: "k8s-sd-test"},
				c,
			},
		}},
	}

	s, err := yaml.Marshal(fullConfig)
	require.NoError(t, err)

	golden.Assert(t, string(s), "NamespaceSetCorrectlyForPodMonitor.golden")
}

func TestProbeStaticTargetsConfigGenerationWithLabelEnforce(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Regex:  "noisy_labels.*",
							Action: "labeldrop",
						},
					},
				},
			},
		},
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithLabelEnforce.golden")
}

func TestProbeStaticTargetsConfigGenerationWithJobName(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithJobName.golden")
}

func TestProbeStaticTargetsConfigGenerationWithoutModule(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithoutModule.golden")
}

func TestProbeIngressSDConfigGeneration(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: ptr.To("bar"),
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	golden.Assert(t, string(cfg), "ProbeIngressSDConfigGeneration.golden")
}

func TestProbeIngressSDConfigGenerationWithShards(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Shards = ptr.To(int32(2))

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: ptr.To("bar"),
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ProbeIngressSDConfigGenerationWithShards.golden")
}

func TestProbeIngressSDConfigGenerationWithLabelEnforce(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.EnforcedNamespaceLabel = "namespace"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									TargetLabel: "foo",
									Replacement: ptr.To("bar"),
									Action:      "replace",
								},
							},
						},
					},
				},
			},
		},
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ProbeIngressSDConfigGenerationWithLabelEnforce.golden")
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
		apiServerConfig *monitoringv1.APIServerConfig
		store           *assets.StoreBuilder
		role            string
		golden          string
	}{
		{
			apiServerConfig: nil,
			store:           assets.NewTestStoreBuilder(),
			role:            "endpoints",
			golden:          "K8SSDConfigGenerationFirst.golden",
		},
		{
			apiServerConfig: &monitoringv1.APIServerConfig{
				Host: "example.com",
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
			store: assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"username": []byte("foo"),
						"password": []byte("bar"),
					},
				},
			),
			role:   "endpoints",
			golden: "K8SSDConfigGenerationTwo.golden",
		},
		{
			apiServerConfig: nil,
			store:           assets.NewTestStoreBuilder(),
			role:            "endpointslice",
			golden:          "K8SSDConfigGenerationThree.golden",
		},
		{
			apiServerConfig: &monitoringv1.APIServerConfig{
				Host: "example.com",
				TLSConfig: &monitoringv1.TLSConfig{
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						CA: monitoringv1.SecretOrConfigMap{
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "ca",
							},
						},
						Cert: monitoringv1.SecretOrConfigMap{
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "cert",
							},
						},
						KeySecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "private-key",
						},
					},
				},
			},
			store:  assets.NewTestStoreBuilder(),
			role:   "endpoints",
			golden: "K8SSDConfigGenerationTLSConfig.golden",
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
				attachMetadata: sm.Spec.AttachMetadata,
			}
		}
		c := cg.generateK8SSDConfig(
			sm.Spec.NamespaceSelector,
			sm.Namespace,
			tc.apiServerConfig,
			tc.store.ForNamespace(sm.Namespace),
			tc.role,
			attachMetaConfig,
		)
		// Wrap partial K8s SD config in a full Prometheus config to satisfy promtool validation.
		fullConfig := yaml.MapSlice{
			{Key: "scrape_configs", Value: []yaml.MapSlice{
				{
					{Key: "job_name", Value: "k8s-sd-test"},
					c,
				},
			}},
		}

		s, err := yaml.Marshal(fullConfig)
		require.NoError(t, err)
		golden.Assert(t, string(s), tc.golden)
	}
}

func TestAlertmanagerBearerToken(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:            "alertmanager-main",
				Namespace:       ptr.To("default"),
				Port:            intstr.FromString("web"),
				BearerTokenFile: "/some/file/on/disk",
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "AlertmanagerBearerToken.golden")
}

func TestAlertmanagerBasicAuth(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		golden  string
	}{
		{
			name:    "Valid Prom Version",
			version: "2.26.0",
			golden:  "AlertmanagerBasicAuth_Valid_Prom_Version.golden",
		},
		{
			name:    "Invalid Prom Version",
			version: "2.25.0",
			golden:  "AlertmanagerBasicAuth_Invalid_Prom_Version.golden",
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
							Namespace: ptr.To("default"),
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
			p,
			nil,
			nil,
			nil,
			nil,
			assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"username": []byte("bob"),
						"password": []byte("alice"),
					},
				},
			),
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)

		golden.Assert(t, string(cfg), tc.golden)
	}
}

func TestAlertmanagerSigv4(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		golden  string
	}{
		{
			name:    "Valid Prom Version",
			version: "2.48.0",
			golden:  "AlertmanagerSigv4_Valid_Prom_Version.golden",
		},
		{
			name:    "Invalid Prom Version",
			version: "2.47.0",
			golden:  "AlertmanagerSigv4_Invalid_Prom_Version.golden",
		},
	} {
		p := defaultPrometheus()
		p.Spec.Version = tc.version
		p.Spec.Alerting = &monitoringv1.AlertingSpec{
			Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
				{
					Name:      "alertmanager-main",
					Namespace: ptr.To("default"),
					Port:      intstr.FromString("web"),
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
				},
			},
		}

		cg := mustNewConfigGenerator(t, p)
		cfg, err := cg.GenerateServerConfiguration(
			p,
			nil,
			nil,
			nil,
			nil,
			assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sigv4-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"access-key": []byte("access-key"),
						"secret-key": []byte("secret-key"),
					},
				},
			),
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		golden.Assert(t, string(cfg), tc.golden)
	}
}

func TestAlertmanagerAPIVersion(t *testing.T) {
	testCases := []struct {
		alerting *monitoringv1.AlertingSpec
		name     string
		version  string
		golden   string
	}{
		{
			name:    "Alertmanager APIV1 Compatible",
			version: "v2.11.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion1),
					},
				},
			},
			golden: "v2/AlertmanagerAPIVersionV1.golden",
		},
		{
			name:    "Alertmanager API Compatible version",
			version: "v2.11.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
					},
				},
			},
			golden: "AlertmanagerAPIVersion.golden",
		},
		{
			name:    "Alertmanager APIV2 Prometheus Version 3",
			version: "3.0.0-rc.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
					},
				},
			},
			golden: "AlertmanagerAPIVersionPrometheusV3.golden",
		},
		{
			name:    "Alertmanager APIv1 Incompatible with Prometheus V3",
			version: "3.0.0-rc.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion1),
					},
				},
			},
			golden: "v2/AlertmanagerAPIVersionV1LowerCasePrometheusV3.golden",
		},
		{
			name:    "Alertmanager APIV1 Incompatible with Prometheus V3",
			version: "3.0.0-rc.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion1),
					},
				},
			},
			golden: "v2/AlertmanagerAPIVersionV1UpperCasePrometheusV3.golden",
		},
		{
			name:    "Alertmanager APIV2 Incompatible with Prometheus V3",
			version: "3.0.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
					},
				},
			},
			golden: "AlertmanagerAPIVersionV2UpperCasePrometheusV3.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			p.Spec.Alerting = tc.alerting

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAlertmanagerTimeoutConfig(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:       "alertmanager-main",
				Namespace:  ptr.To("default"),
				Port:       intstr.FromString("web"),
				APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
				Timeout:    ptr.To(monitoringv1.Duration("60s")),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "AlertmanagerTimeoutConfig.golden")
}
func TestAlertmanagerEnableHttp2(t *testing.T) {
	for _, tc := range []struct {
		version     string
		enableHTTP2 bool
		golden      string
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			golden:      "AlertmanagerEnableHttp2_false_expectedWithHTTP2Unsupported.golden",
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			golden:      "AlertmanagerEnableHttp2_true_expectedWithHTTP2Unsupported.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			golden:      "AlertmanagerEnableHttp2_true_expectedWithHTTP2Enabled.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			golden:      "AlertmanagerEnableHttp2_false_expectedWithHTTP2Enabled.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s TestAlertmanagerEnableHttp2(%t)", tc.version, tc.enableHTTP2), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.Version = tc.version
			p.Spec.Alerting = &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:        "alertmanager-main",
						Namespace:   ptr.To("default"),
						Port:        intstr.FromString("web"),
						APIVersion:  ptr.To(monitoringv1.AlertmanagerAPIVersion2),
						EnableHttp2: ptr.To(tc.enableHTTP2),
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAlertmanagerRelabelConfigs(t *testing.T) {
	t.Run("TestAlertmanagerRelabelConfigs", func(t *testing.T) {
		p := defaultPrometheus()
		p.Spec.Alerting = &monitoringv1.AlertingSpec{
			Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
				{
					Name:       "alertmanager-main",
					Namespace:  ptr.To("default"),
					Port:       intstr.FromString("web"),
					APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							TargetLabel: "namespace",
							Replacement: ptr.To("ns1"),
						},
						{
							Action:       "replace",
							Regex:        "(.+)(?::d+)",
							Replacement:  ptr.To("$1:9537"),
							SourceLabels: []monitoringv1.LabelName{"__address__"},
							TargetLabel:  "__address__",
						},
						{
							Action:      "replace",
							Replacement: ptr.To("crio"),
							TargetLabel: "job",
						},
					},
				},
			},
		}

		cg := mustNewConfigGenerator(t, p)
		cfg, err := cg.GenerateServerConfiguration(
			p,
			nil,
			nil,
			nil,
			nil,
			&assets.StoreBuilder{},
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		golden.Assert(t, string(cfg), "Alertmanager_with_RelabelConfigs.golden")
	})
}

func TestAlertmanagerAlertRelabelConfigs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		golden  string
	}{
		{
			name:    "Invalid Prometheus Version",
			version: "2.40.1",
			golden:  "AlertmanagerAlertRelabel_Invalid_Version.golden",
		},
		{
			name:    "Valid Prometheus Version",
			version: "2.51.0",
			golden:  "AlertmanagerAlertRelabel_Valid_Version.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.Version = tc.version
			p.Spec.Alerting = &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:       "alertmanager-main",
						Namespace:  ptr.To("default"),
						Port:       intstr.FromString("web"),
						APIVersion: ptr.To(monitoringv1.AlertmanagerAPIVersion2),
						AlertRelabelConfigs: []monitoringv1.RelabelConfig{
							{
								TargetLabel: "namespace",
								Replacement: ptr.To("ns1"),
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAdditionalScrapeConfigs(t *testing.T) {
	getCfg := func(shards *int32) string {
		p := defaultPrometheus()
		p.Spec.Shards = shards

		cg := mustNewConfigGenerator(t, p)
		cfg, err := cg.GenerateServerConfiguration(
			p,
			nil,
			nil,
			nil,
			nil,
			&assets.StoreBuilder{},
			golden.Get(t, "input/TestAdditionalScrapeConfigsAdditionalScrapeConfig.golden"),
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)

		return string(cfg)
	}

	testCases := []struct {
		name   string
		result string
		golden string
	}{
		{
			name:   "unsharded prometheus",
			result: getCfg(nil),
			golden: "AdditionalScrapeConfigs_unsharded_prometheus.golden",
		},
		{
			name:   "one prometheus shard",
			result: getCfg(ptr.To(int32(1))),
			golden: "AdditionalScrapeConfigs_one_prometheus_shard.golden",
		},
		{
			name:   "sharded prometheus",
			result: getCfg(ptr.To(int32(3))),
			golden: "AdditionalScrapeConfigs_sharded prometheus.golden",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			golden.Assert(t, tt.result, tt.golden)
		})
	}
}

func TestAdditionalAlertRelabelConfigs(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:      "alertmanager-main",
				Namespace: ptr.To("default"),
				Port:      intstr.FromString("web"),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		golden.Get(t, "input/AdditionalAlertRelabelConfigs.golden"),
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "AdditionalAlertRelabelConfigs_Expected.golden")
}

func TestNoEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)",
									SourceLabels: []monitoringv1.LabelName{"__name__"},
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.+)(?::d+)",
									Replacement:  ptr.To("$1:9537"),
									SourceLabels: []monitoringv1.LabelName{"__address__"},
									TargetLabel:  "__address__",
								},
								{
									Action:      "replace",
									Replacement: ptr.To("crio"),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "NoEnforcedNamespaceLabelServiceMonitor_Expected.golden")
}

func TestServiceMonitorWithServiceDiscoveryRole(t *testing.T) {
	for _, tc := range []struct {
		name     string
		promRole *monitoringv1.ServiceDiscoveryRole
		smonRole *monitoringv1.ServiceDiscoveryRole
		golden   string
	}{
		{
			name:   "Default",
			golden: "TestServiceMonitorWithDefaultServiceDiscoveryRole.golden",
		},
		{
			name:     "Prometheus with endpoints",
			promRole: ptr.To(monitoringv1.EndpointsRole),
			golden:   "TestServiceMonitorWithDefaultServiceDiscoveryRole.golden",
		},
		{
			name:     "Prometheus with endpointslice",
			promRole: ptr.To(monitoringv1.EndpointSliceRole),
			golden:   "TestServiceMonitorWithEndpointSliceEnable_Expected.golden",
		},
		{
			name:     "Prometheus with endpointslice and ServiceMonitor with endpoints",
			promRole: ptr.To(monitoringv1.EndpointSliceRole),
			smonRole: ptr.To(monitoringv1.EndpointsRole),
			golden:   "TestServiceMonitorWithEndpointsAndPrometheusEndpointSlice.golden",
		},
		{
			name:     "Prometheus with endpoints and ServiceMonitor with endpointslice",
			promRole: ptr.To(monitoringv1.EndpointsRole),
			smonRole: ptr.To(monitoringv1.EndpointSliceRole),
			golden:   "TestServiceMonitorWithEndpointSliceAndPrometheusEndpoints.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.ServiceDiscoveryRole = tc.promRole

			smon := &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					ServiceDiscoveryRole: ptr.To(monitoringv1.EndpointsRole),
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
						},
					},
				},
			}
			smon.Spec.ServiceDiscoveryRole = tc.smonRole

			cg := mustNewConfigGenerator(t, p, WithEndpointSliceSupport())
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"test": smon,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestEnforcedNamespaceLabelPodMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "ns-key"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							Port:     ptr.To("web"),
							Interval: "30s",
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "my-ns",
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  ptr.To("$1"),
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
									TargetLabel:  "pod_ready",
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelPodMonitor_Expected.golden")
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
		p,
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
							Port:     ptr.To("web"),
							Interval: "30s",
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "my-ns",
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  ptr.To("$1"),
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
									TargetLabel:  "pod_ready",
								},
							},
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelOnExcludedPodMonitor_Expected.golden")
}

func TestEnforcedNamespaceLabelServiceMonitor(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "ns-key"

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "ns-key",
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  ptr.To("$1"),
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
									TargetLabel:  "pod_ready",
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelServiceMonitor_Expected.golden")
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
		p,
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
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "drop",
									Regex:        "my-job-pod-.+",
									SourceLabels: []monitoringv1.LabelName{"pod_name"},
									TargetLabel:  "ns-key",
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "replace",
									Regex:        "(.*)",
									Replacement:  ptr.To("$1"),
									SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
									TargetLabel:  "pod_ready",
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelOnExcludedServiceMonitor_Expected.golden")
}

func TestAdditionalAlertmanagers(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Name:      "alertmanager-main",
				Namespace: ptr.To("default"),
				Port:      intstr.FromString("web"),
			},
		},
	}

	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		[]byte(`- static_configs:
  - targets:
    - localhost
`),
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "TestAdditionalAlertmanagers_Expected.golden")
}

func TestSettingHonorTimestampsInServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							HonorTimestamps: ptr.To(false),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "SettingHonorTimestampsInServiceMonitor.golden")
}

func TestSettingHonorTimestampsInPodMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							HonorTimestamps: ptr.To(false),
							Port:            ptr.To("web"),
							Interval:        "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "SettingHonorTimestampsInPodMonitor.golden")
}

func TestSettingTrackTimestampsStalenessInServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							TrackTimestampsStaleness: ptr.To(false),
							Port:                     "web",
							Interval:                 "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "SettingTrackTimestampsStalenessInServiceMonitor.golden")
}

func TestSettingTrackTimestampsStalenessInPodMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							TrackTimestampsStaleness: ptr.To(false),
							Port:                     ptr.To("web"),
							Interval:                 "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "SettingTrackTimestampsStalenessInPodMonitor.golden")
}

func TestSettingScrapeProtocolsInServiceMonitor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		scrape  []monitoringv1.ScrapeProtocol
		golden  string
	}{
		{
			name:    "setting ScrapeProtocols in ServiceMonitor with prometheus old version",
			version: "v2.48.0",
			scrape: []monitoringv1.ScrapeProtocol{
				monitoringv1.ScrapeProtocol("OpenMetricsText1.0.0"),
				monitoringv1.ScrapeProtocol("OpenMetricsText0.0.1"),
			},
			golden: "SettingScrapeProtocolsInServiceMonitor_OldVersion.golden",
		},
		{
			name:    "setting ScrapeProtocols in ServiceMonitor with prometheus new version",
			version: "v2.49.0",
			scrape: []monitoringv1.ScrapeProtocol{
				monitoringv1.ScrapeProtocol("OpenMetricsText1.0.0"),
				monitoringv1.ScrapeProtocol("OpenMetricsText0.0.1"),
			},
			golden: "SettingScrapeProtocolsInServiceMonitor_NewVersion.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testservicemonitor1",
							Namespace: "default",
						},
						Spec: monitoringv1.ServiceMonitorSpec{
							TargetLabels:    []string{"example", "env"},
							ScrapeProtocols: tc.scrape,
							Endpoints: []monitoringv1.Endpoint{
								{
									HonorTimestamps: ptr.To(false),
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
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestSettingScrapeFallbackProtocolInServiceMonitor(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		version                string
		fallbackScrapeProtocol *monitoringv1.ScrapeProtocol
		golden                 string
	}{
		{
			name:                   "setting FallbackScrapeProtocol in ServiceMonitor with prometheus old version",
			version:                "v2.55.0",
			fallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
			golden:                 "SettingScrapeFallbackProtocolInServiceMonitor_OldVersion.golden",
		},
		{
			name:                   "setting FallbackScrapeProtocol in ServiceMonitor with prometheus new version",
			version:                "v3.0.0",
			fallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText0_0_1),
			golden:                 "SettingScrapeFallbackProtocolInServiceMonitor_NewVersion.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testservicemonitor1",
							Namespace: "default",
						},
						Spec: monitoringv1.ServiceMonitorSpec{
							TargetLabels:           []string{"example", "env"},
							FallbackScrapeProtocol: tc.fallbackScrapeProtocol,
							Endpoints: []monitoringv1.Endpoint{
								{
									HonorTimestamps: ptr.To(false),
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
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestSettingScrapeProtocolsInPodMonitor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		scrape  []monitoringv1.ScrapeProtocol
		golden  string
	}{
		{
			name:    "setting ScrapeProtocols in PodMonitor with prometheus old version",
			version: "v2.48.0",
			scrape: []monitoringv1.ScrapeProtocol{
				monitoringv1.ScrapeProtocol("OpenMetricsText1.0.0"),
				monitoringv1.ScrapeProtocol("OpenMetricsText0.0.1"),
			},
			golden: "SettingScrapeProtocolsInPodMonitor_OldVersion.golden",
		},
		{
			name:    "setting ScrapeProtocols in PodMonitor with prometheus new version",
			version: "v2.49.0",
			scrape: []monitoringv1.ScrapeProtocol{
				monitoringv1.ScrapeProtocol("OpenMetricsText1.0.0"),
				monitoringv1.ScrapeProtocol("OpenMetricsText0.0.1"),
			},
			golden: "SettingScrapeProtocolsInPodMonitor_NewVersion.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpodmonitor1",
							Namespace: "default",
						},
						Spec: monitoringv1.PodMonitorSpec{
							PodTargetLabels: []string{"example", "env"},
							ScrapeProtocols: tc.scrape,
							PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
								{
									TrackTimestampsStaleness: ptr.To(false),
									Port:                     ptr.To("web"),
									Interval:                 "30s",
								},
							},
						},
					},
				},
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestSettingScrapeFallbackProtocolInPodMonitor(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		version                string
		fallbackScrapeProtocol *monitoringv1.ScrapeProtocol
		golden                 string
	}{
		{
			name:                   "setting FallbackScrapeProtocol in PodMonitor with prometheus old version",
			version:                "v2.55.0",
			fallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText0_0_1),
			golden:                 "SettingScrapeFallbackProtocolInPodMonitor_OldVersion.golden",
		},
		{
			name:                   "setting FallbackScrapeProtocol in PodMonitor with prometheus new version",
			version:                "v3.0.0",
			fallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
			golden:                 "SettingScrapeFallbackProtocolInPodMonitor_NewVersion.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpodmonitor1",
							Namespace: "default",
						},
						Spec: monitoringv1.PodMonitorSpec{
							PodTargetLabels:        []string{"example", "env"},
							FallbackScrapeProtocol: tc.fallbackScrapeProtocol,
							PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
								{
									TrackTimestampsStaleness: ptr.To(false),
									Port:                     ptr.To("web"),
									Interval:                 "30s",
								},
							},
						},
					},
				},
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestHonorTimestampsOverriding(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.OverrideHonorTimestamps = true

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							HonorTimestamps: ptr.To(true),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "HonorTimestampsOverriding.golden")
}

func TestSettingHonorLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "SettingHonorLabels.golden")
}

func TestHonorLabelsOverriding(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.OverrideHonorLabels = true

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "HonorLabelsOverriding.golden")
}

func TestTargetLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "TargetLabels.golden")
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
		TLSConfig: &monitoringv1.SafeTLSConfig{
			InsecureSkipVerify: ptr.To(true),
			CA: monitoringv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "tls",
					},
					Key: "ca2",
				},
			},
		},
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
						Key: "proxy-header",
					},
				},
			},
		},
	}

	s := assets.NewTestStoreBuilder(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "oauth2",
				Namespace: "default",
			},
			Data: map[string]string{
				"client_id": "test_client_id",
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "oauth2",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"client_secret": []byte("test_client_secret"),
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"proxy-header": []byte("value"),
				"token":        []byte("value"),
				"Username":     []byte("kube-admin"),
				"Password":     []byte("password"),
			},
		},
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

	testCases := []struct {
		name   string
		sMons  map[string]*monitoringv1.ServiceMonitor
		pMons  map[string]*monitoringv1.PodMonitor
		probes map[string]*monitoringv1.Probe
		golden string
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
								Port: "web",
								HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
									HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
										HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
											OAuth2: &oauth2,
										},
									},
								},
							},
						},
					},
				},
			},
			golden: "service_monitor_with_oauth2.golden",
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
								Port: ptr.To("web"),
								HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
									HTTPConfig: monitoringv1.HTTPConfig{
										HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
											OAuth2: &oauth2,
										},
									},
								},
							},
						},
					},
				},
			},
			golden: "pod_monitor_with_oauth2.golden",
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
						HTTPConfig: monitoringv1.HTTPConfig{
							HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
								OAuth2: &oauth2,
							},
						},
						Targets: monitoringv1.ProbeTargets{
							StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
								Targets: []string{"127.0.0.1"},
							},
						},
					},
				},
			},
			golden: "probe_monitor_with_oauth2.golden",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			p := defaultPrometheus()

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				tt.sMons,
				tt.pMons,
				tt.probes,
				nil,
				s,
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tt.golden)
		})
	}
}

func TestPodTargetLabels(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "PodTargetLabels.golden")
}

func TestPodTargetLabelsFromPodMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "PodTargetLabelsFromPodMonitor.golden")
}

func TestPodTargetLabelsFromPodMonitorAndGlobal(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.PodTargetLabels = []string{"global"}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "PodTargetLabelsFromPodMonitorAndGlobal.golden")
}

func TestEmptyEndpointPorts(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "EmptyEndpointPorts.golden")
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
						Namespace: ptr.To("default"),
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
				Replicas: ptr.To(int32(1)),
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
		p,
		makeServiceMonitors(),
		makePodMonitors(),
		nil,
		nil,
		&assets.StoreBuilder{},
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
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
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
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  ptr.To("$1"),
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
							TargetLabel:  "pod_ready",
						},
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  ptr.To("$1"),
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
					Port:     ptr.To("web"),
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
					Port:     ptr.To("web"),
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
					Port:     ptr.To("web"),
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
					Port:     ptr.To("web"),
					Interval: "30s",
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
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
					Port:     ptr.To("web"),
					Interval: "30s",
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  ptr.To("$1"),
							SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ready"},
							TargetLabel:  "pod_ready",
						},
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  ptr.To("$1"),
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
			UserHonorTimestamps:     ptr.To(false),
			OverrideHonorTimestamps: true,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     ptr.To(false),
			OverrideHonorTimestamps: false,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     ptr.To(true),
			OverrideHonorTimestamps: true,
			Expected:                "honor_timestamps: false\n",
		},
		{
			UserHonorTimestamps:     ptr.To(true),
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

func TestTrackTimestampsStaleness(t *testing.T) {
	type testCase struct {
		UserTrackTimestampsStaleness *bool
		Expected                     string
	}

	testCases := []testCase{
		{
			UserTrackTimestampsStaleness: nil,
			Expected:                     "{}\n",
		},
		{
			UserTrackTimestampsStaleness: ptr.To(false),
			Expected:                     "track_timestamps_staleness: false\n",
		},
		{
			UserTrackTimestampsStaleness: ptr.To(true),
			Expected:                     "track_timestamps_staleness: true\n",
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			cg := mustNewConfigGenerator(t, &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "2.48.0",
					},
				},
			})

			hl, _ := yaml.Marshal(cg.AddTrackTimestampsStaleness(yaml.MapSlice{}, tc.UserTrackTimestampsStaleness))
			require.Equal(t, tc.Expected, string(hl))
		})
	}
}

func TestSampleLimits(t *testing.T) {
	for _, tc := range []struct {
		globalLimit   int
		enforcedLimit int
		limit         int
		golden        string
		version       string
	}{
		{
			version:       "v2.15.0",
			globalLimit:   -1,
			enforcedLimit: -1,
			limit:         -1,
			golden:        "SampleLimits_NoLimit.golden",
		},
		{
			version:       "v2.47.0",
			globalLimit:   -1,
			enforcedLimit: 1000,
			limit:         -1,
			golden:        "SampleLimits_Limit-1.golden",
		},
		{
			version:       "v2.47.0",
			globalLimit:   -1,
			enforcedLimit: 1000,
			limit:         2000,
			golden:        "SampleLimits_Limit2000.golden",
		},
		{
			version:       "v2.47.0",
			globalLimit:   -1,
			enforcedLimit: 1000,
			limit:         500,
			golden:        "SampleLimits_Limit500.golden",
		},
		{
			version:       "v2.47.0",
			globalLimit:   1000,
			enforcedLimit: 2000,
			limit:         -1,
			golden:        "SampleLimits_GlobalLimit1000_Enforce2000.golden",
		},
		{
			version:       "v2.21.0",
			globalLimit:   1000,
			enforcedLimit: 2000,
			limit:         -1,
			golden:        "SampleLimits_GlobalLimit1000_Enforce2000-2.21.golden",
		},
		{
			version:       "v2.21.0",
			globalLimit:   1000,
			enforcedLimit: -1,
			limit:         -1,
			golden:        "SampleLimits_GlobalLimit1000_Enforce-1-2.21.golden",
		},
		{
			version:       "v2.21.0",
			globalLimit:   -1,
			enforcedLimit: 2000,
			limit:         500,
			golden:        "SampleLimits_GlobalLimit-1_Enforce-2000-Limit-500-2.21.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedlimit(%d) limit(%d)", tc.version, tc.enforcedLimit, tc.limit), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version
			if tc.globalLimit >= 0 {
				p.Spec.SampleLimit = ptr.To(uint64(tc.globalLimit))
			}

			if tc.golden == "SampleLimits_GlobalLimit1000_Enforce2000.golden" {
				fmt.Print("test")
			}

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
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestTargetLimits(t *testing.T) {
	for _, tc := range []struct {
		version       string
		enforcedLimit int
		limit         int
		expected      string
		golden        string
	}{
		{
			version:       "v2.15.0",
			enforcedLimit: -1,
			limit:         -1,
			golden:        "TargetLimits-1_Versionv2.15.0.golden",
		},
		{
			version:       "v2.21.0",
			enforcedLimit: -1,
			limit:         -1,
			golden:        "TargetLimits-1_Versionv2.21.0.golden",
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         -1,
			golden:        "TargetLimits-1_Versionv2.15.0.golden",
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         -1,
			golden:        "TargetLimits-1_Versionv2.21.0_Enforce1000.golden",
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         2000,
			golden:        "TargetLimits2000_Versionv2.15.0_Enforce1000.golden",
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         2000,
			golden:        "TargetLimits2000_Versionv2.21.0_Enforce1000.golden",
		},
		{
			version:       "v2.15.0",
			enforcedLimit: 1000,
			limit:         500,
			golden:        "TargetLimits500_Versionv2.15.0_Enforce1000.golden",
		},
		{
			version:       "v2.21.0",
			enforcedLimit: 1000,
			limit:         500,
			golden:        "TargetLimits1000_Versionv2.21.0_Enforce1000.golden",
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
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestRemoteReadConfig(t *testing.T) {
	for _, tc := range []struct {
		version     string
		remoteRead  monitoringv1.RemoteReadSpec
		golden      string
		expectedErr error
	}{
		{
			version: "v2.27.1",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
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
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			golden: "RemoteReadConfig_v2.27.1.golden",
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
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
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			golden: "RemoteReadConfig_v2.26.0.golden",
		},
		{
			version: "v2.25.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:             "http://example.com",
				FollowRedirects: ptr.To(true),
			},
			golden: "RemoteReadConfig_v2.25.0.golden",
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:             "http://example.com",
				FollowRedirects: ptr.To(false),
			},
			golden: "RemoteReadConfig_v2.26.0_NotFollowRedirects.golden",
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: ptr.To(true),
			},
			golden: "RemoteReadConfig_v2.26.0_FilterExternalLabels.golden",
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
			},
			golden: "RemoteReadConfig_v2.34.0.golden",
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: ptr.To(false),
			},
			golden: "RemoteReadConfig_v2.34.0_NotFilterExternalLabels.golden",
		},
		{
			version: "v2.34.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL:                  "http://example.com",
				FilterExternalLabels: ptr.To(true),
			},
			golden: "RemoteReadConfig_v2.34.0_FilterExternalLabels.golden",
		},
		{
			version: "v2.26.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
				Authorization: &monitoringv1.Authorization{
					SafeAuthorization: monitoringv1.SafeAuthorization{
						Credentials: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "auth",
							},
							Key: "bearer",
						},
					},
				},
			},
			golden: "RemoteReadConfig_v2.26.0_AuthorizationSafe.golden",
		},
		{
			version: "v2.43.0",
			remoteRead: monitoringv1.RemoteReadSpec{
				URL: "http://example.com",
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
								Key: "proxy-header",
							},
						},
					},
				},
			},
			golden: "RemoteReadConfig_v2.43.0_ProxyConfig.golden",
		},
	} {
		t.Run(fmt.Sprintf("version=%s", tc.version), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version
			p.Spec.RemoteRead = []monitoringv1.RemoteReadSpec{tc.remoteRead}

			s := assets.NewTestStoreBuilder(
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "auth",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"bearer": []byte("secret"),
					},
				},
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

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				s,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestRemoteWriteConfig(t *testing.T) {
	sendNativeHistograms := true
	enableHTTP2 := false
	followRedirects := true
	for i, tc := range []struct {
		version     string
		remoteWrite monitoringv1.RemoteWriteSpec
		golden      string
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_v2.22.0_1.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_v2.23.0_1.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_v2.23.0_2.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "legacy-versions/RemoteWriteConfig_v2.10.0_1.golden",
		},
		{
			version: "v2.27.1",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				OAuth2: &monitoringv1.OAuth2{
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
					TokenURL:       "http://token-url",
					Scopes:         []string{"scope1"},
					EndpointParams: map[string]string{"param": "value"},
				},
			},
			golden: "RemoteWriteConfig_v2.27.1_1.golden",
		},
		{
			version: "v2.45.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("00000000-0000-0000-0000-000000000000"),
					},
				},
			},
			golden: "RemoteWriteConfig_v2.45.0_1.golden",
		},
		{
			version: "v2.48.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "00000000-0000-0000-0000-000000000000",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
				},
			},
			golden: "RemoteWriteConfigAzureADOAuth_v2.48.0_1.golden",
		},
		{
			version: "v2.52.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
				},
			},
			golden: "RemoteWriteConfigAzureADSDK_v2.52.0.golden",
		},
		{
			version: "v3.5.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("00000000-0000-0000-0000-000000000000"),
					},
				},
			},
			golden: "RemoteWriteConfig_AzureADManagedIdentity_v3.5.0.golden",
		},
		{
			version: "v3.7.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			golden: "RemoteWriteConfigAzureADWorkloadIdentity_v3.7.0.golden",
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				Authorization: &monitoringv1.Authorization{
					SafeAuthorization: monitoringv1.SafeAuthorization{
						Credentials: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "auth",
							},
							Key: "token",
						},
					},
				},
			},
			golden: "RemoteWriteConfig_v2.26.0_2.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_3.golden",
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				RemoteTimeout: ptr.To(monitoringv1.Duration("1s")),
				Sigv4:         nil,
			},
			golden: "RemoteWriteConfig_v2.26.0_3.golden",
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				Sigv4:         &monitoringv1.Sigv4{},
				RemoteTimeout: ptr.To(monitoringv1.Duration("1s")),
			},
			golden: "RemoteWriteConfig_v2.26.0_4.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
					RetryOnRateLimit:  true,
				},
			},
			golden: "RemoteWriteConfig_v2.30.0_2.golden",
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
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
					RetryOnRateLimit:  true,
				},
			},
			golden: "RemoteWriteConfig_v2.43.0_2.golden",
		},
		{
			version: "v2.39.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:                  "http://example.com",
				SendNativeHistograms: &sendNativeHistograms,
				EnableHttp2:          &enableHTTP2,
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
					RetryOnRateLimit:  true,
				},
			},
			golden: "RemoteWriteConfig_v2.39.0_1.golden",
		},
		{
			version: "v2.49.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:                  "http://example.com",
				SendNativeHistograms: &sendNativeHistograms,
				EnableHttp2:          &enableHTTP2,
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
					RetryOnRateLimit:  true,
					SampleAgeLimit:    ptr.To(monitoringv1.Duration("1s")),
				},
			},
			golden: "RemoteWriteConfig_v2.49.0.golden",
		},
		{
			version: "v2.50.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:                  "http://example.com",
				SendNativeHistograms: &sendNativeHistograms,
				EnableHttp2:          &enableHTTP2,
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
					RetryOnRateLimit:  true,
					SampleAgeLimit:    ptr.To(monitoringv1.Duration("1s")),
				},
			},
			golden: "RemoteWriteConfig_v2.50.0.golden",
		},
		{
			version: "v2.43.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:             "http://example.com",
				FollowRedirects: &followRedirects,
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
								Key: "proxy-header",
							},
						},
					},
				},
			},
			golden: "RemoteWriteConfig_v2.43.0_ProxyConfig.golden",
		},
		{
			version: "v2.43.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:             "http://example.com",
				FollowRedirects: &followRedirects,
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
								Key: "proxy-header",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "proxy-header",
							},
						},
					},
				},
			},
			golden: "RemoteWriteConfig_v2.43.0_ProxyConfigWithMutiValues.golden",
		},
		{
			version: "v2.53.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:            "http://example.com",
				MessageVersion: ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
			},
			golden: "RemoteWriteConfig_v2.53.0_MessageVersion2.golden",
		},
		{
			version: "v2.54.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:            "http://example.com",
				MessageVersion: ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
			},
			golden: "RemoteWriteConfig_v2.54.0_MessageVersion2.golden",
		},
		{
			version: "v3.1.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:            "http://example.com",
				MessageVersion: ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
				RoundRobinDNS:  ptr.To(true),
			},
			golden: "RemoteWriteConfig_v3.1.0.golden",
		},
		{
			version: "v2.28.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				MetadataConfig: &monitoringv1.MetadataConfig{
					MaxSamplesPerSend: ptr.To(int32(10)),
				},
			},
			golden: "RemoteWriteConfig_v2.28.0_MaxSamplesPerSendMetadataConfig.golden",
		},
		{
			version: "v2.29.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				MetadataConfig: &monitoringv1.MetadataConfig{
					MaxSamplesPerSend: ptr.To(int32(10)),
				},
			},
			golden: "RemoteWriteConfig_v2.29.0_MaxSamplesPerSendMetadataConfig.golden",
		},
		{
			version: "v2.53.0",
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
					Region:             "us-central-0",
					UseFIPSSTSEndpoint: ptr.To(true),
				},
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_v2.53.0.golden",
		},
		{
			version: "v2.54.0",
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
					Region:             "us-central-0",
					UseFIPSSTSEndpoint: ptr.To(true),
				},
				QueueConfig: &monitoringv1.QueueConfig{
					Capacity:          1000,
					MinShards:         1,
					MaxShards:         10,
					MaxSamplesPerSend: 100,
					BatchSendDeadline: ptr.To(monitoringv1.Duration("20s")),
					MaxRetries:        3,
					MinBackoff:        ptr.To(monitoringv1.Duration("1s")),
					MaxBackoff:        ptr.To(monitoringv1.Duration("10s")),
				},
				MetadataConfig: &monitoringv1.MetadataConfig{
					Send:         false,
					SendInterval: "1m",
				},
			},
			golden: "RemoteWriteConfig_v2.54.0.golden",
		},
		{
			// Empty metadataConfig defaults to no metadata being sent.
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:            "http://example.com",
				MetadataConfig: &monitoringv1.MetadataConfig{},
			},
			golden: "RemoteWriteConfigWithEmptyMetadataConfig.golden",
		},
		{
			version: "v3.8.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzurePublic"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
					Scope: ptr.To("https://custom.monitor.azure.com/.default"),
				},
			},
			golden: "RemoteWriteConfig_AzureADScope_v3.8.0.golden",
		},
		{
			version: "v3.9.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzurePublic"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
					Scope: ptr.To("https://custom.monitor.azure.com/.default"),
				},
			},
			golden: "RemoteWriteConfig_AzureADScope_v3.9.0.golden",
		},
	} {
		t.Run(fmt.Sprintf("i=%d,version=%s", i, tc.version), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version
			p.Spec.CommonPrometheusFields.RemoteWrite = []monitoringv1.RemoteWriteSpec{tc.remoteWrite}
			p.Spec.CommonPrometheusFields.Secrets = []string{"sigv4-secret"}

			store := assets.NewTestStoreBuilder(
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
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
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value1"),
						"token":        []byte("value1"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "auth",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"token": []byte("secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sigv4-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"access-key": []byte("access-key"),
						"secret-key": []byte("secret-key"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "azure-oauth-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"secret-key": []byte("secret-key"),
					},
				},
			)

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestLabelLimits(t *testing.T) {
	for _, tc := range []struct {
		version            string
		enforcedLabelLimit int
		labelLimit         int
		golden             string
	}{
		{
			version:            "v2.26.0",
			enforcedLabelLimit: -1,
			labelLimit:         -1,
			golden:             "LabelLimits_NoLimit_v2.26.0.golden",
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: -1,
			labelLimit:         -1,

			golden: "LabelLimits_NoLimit_v2.27.0.golden",
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         -1,

			golden: "LabelLimits_NoLimit_v2.26.0_enforceLimit1000.golden",
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         -1,
			golden:             "LabelLimits_NoLimit_v2.27.0_enforceLimit1000.golden",
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         2000,

			golden: "LabelLimits_Limit2000_v2.26.0_enforceLimit1000.golden",
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         2000,
			golden:             "LabelLimits_Limit2000_v2.27.0_enforceLimit1000.golden",
		},
		{
			version:            "v2.26.0",
			enforcedLabelLimit: 1000,
			labelLimit:         500,

			golden: "LabelLimits_Limit500_v2.26.0_enforceLimit1000.golden",
		},
		{
			version:            "v2.27.0",
			enforcedLabelLimit: 1000,
			labelLimit:         500,
			golden:             "LabelLimits_Limit500_v2.27.0_enforceLimit1000.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelLimit(%d) labelLimit(%d)", tc.version, tc.enforcedLabelLimit, tc.labelLimit), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelLimit >= 0 {
				p.Spec.EnforcedLabelLimit = ptr.To(uint64(tc.enforcedLabelLimit))
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
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestLabelNameLengthLimits(t *testing.T) {
	for _, tc := range []struct {
		version                      string
		enforcedLabelNameLengthLimit int
		labelNameLengthLimit         int
		golden                       string
	}{
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: -1,
			labelNameLengthLimit:         -1,
			golden:                       "LabelNameLengthLimits_Limit-1_Enforce-1_v2.26.0.golden",
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: -1,
			labelNameLengthLimit:         -1,
			golden:                       "LabelNameLengthLimits_Limit-1_Enforce-1_v2.27.0.golden",
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         -1,
			golden:                       "LabelNameLengthLimits_Limit-1_Enforc1000_v2.26.0.golden",
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         -1,
			golden:                       "LabelNameLengthLimits_Limit-1_Enforce1000_v2.27.0.golden",
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         2000,
			golden:                       "LabelNameLengthLimits_Limit2000_Enforce1000_v2.26.0.golden",
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         2000,
			golden:                       "LabelNameLengthLimits_Limit2000_Enforce1000_v2.27.0.golden",
		},
		{
			version:                      "v2.26.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         500,
			golden:                       "LabelNameLengthLimits_Limit500_Enforce1000_v2.26.0.golden",
		},
		{
			version:                      "v2.27.0",
			enforcedLabelNameLengthLimit: 1000,
			labelNameLengthLimit:         500,
			golden:                       "LabelNameLengthLimits_Limit500_Enforce1000_v2.27.0.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelNameLengthLimit(%d) labelNameLengthLimit(%d)", tc.version, tc.enforcedLabelNameLengthLimit, tc.labelNameLengthLimit), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelNameLengthLimit >= 0 {
				p.Spec.EnforcedLabelNameLengthLimit = ptr.To(uint64(tc.enforcedLabelNameLengthLimit))
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
							Port:     ptr.To("web"),
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
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestLabelValueLengthLimits(t *testing.T) {
	for _, tc := range []struct {
		version                       string
		enforcedLabelValueLengthLimit int
		labelValueLengthLimit         int
		golden                        string
	}{
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: -1,
			labelValueLengthLimit:         -1,
			golden:                        "LabelValueLengthLimits_Enforce-1_LabelValue-1_v2.26.0.golden",
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: -1,
			labelValueLengthLimit:         -1,
			golden:                        "LabelValueLengthLimits_Enforce-1_LabelValue-1_v2.27.0.golden",
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         -1,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue-1_v2.26.0.golden",
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         -1,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue-1_v2.27.0.golden",
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         2000,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue2000_v2.26.0.golden",
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         2000,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue2000_v2.27.0.golden",
		},
		{
			version:                       "v2.26.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         500,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue500_v2.26.0.golden",
		},
		{
			version:                       "v2.27.0",
			enforcedLabelValueLengthLimit: 1000,
			labelValueLengthLimit:         500,
			golden:                        "LabelValueLengthLimits_Enforce1000_LabelValue500_v2.27.0.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedLabelValueLengthLimit(%d) labelValueLengthLimit(%d)", tc.version, tc.enforcedLabelValueLengthLimit, tc.labelValueLengthLimit), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			if tc.enforcedLabelValueLengthLimit >= 0 {
				p.Spec.EnforcedLabelValueLengthLimit = ptr.To(uint64(tc.enforcedLabelValueLengthLimit))
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
						Scheme: ptr.To(monitoringv1.SchemeHTTP),
						URL:    "blackbox.exporter.io",
						Path:   "/probe",
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
										Key: "proxy-header",
									},
								},
							},
						},
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

			s := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
					},
				},
			)

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				map[string]*monitoringv1.Probe{
					"testprobe1": &probe,
				},
				nil,
				s,
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestKeepDroppedTargets(t *testing.T) {
	for _, tc := range []struct {
		version                    string
		enforcedKeepDroppedTargets *uint64
		keepDroppedTargets         *uint64
		golden                     string
	}{
		{
			version:                    "v2.46.0",
			enforcedKeepDroppedTargets: ptr.To(uint64(1000)),
			keepDroppedTargets:         ptr.To(uint64(50)),
			golden:                     "KeepDroppedTargetsNotAddedInConfig.golden",
		},
		{
			version:                    "v2.47.0",
			enforcedKeepDroppedTargets: ptr.To(uint64(1000)),
			keepDroppedTargets:         ptr.To(uint64(2000)),
			golden:                     "KeepDroppedTargetsOverridedWithEnforcedValue.golden",
		},
		{
			version:                    "v2.47.0",
			enforcedKeepDroppedTargets: ptr.To(uint64(1000)),
			keepDroppedTargets:         ptr.To(uint64(500)),
			golden:                     "KeepDroppedTargets.golden",
		},
	} {
		t.Run(fmt.Sprintf("%s enforcedKeepDroppedTargets(%d) keepDroppedTargets(%d)", tc.version, tc.enforcedKeepDroppedTargets, tc.keepDroppedTargets), func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = tc.version

			p.Spec.EnforcedKeepDroppedTargets = tc.enforcedKeepDroppedTargets

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

			serviceMonitor.Spec.KeepDroppedTargets = tc.keepDroppedTargets

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestNativeHistogramConfig(t *testing.T) {
	for _, tc := range []struct {
		version               string
		nativeHistogramConfig monitoringv1.NativeHistogramConfig
		golden                string
	}{
		{
			version: "v3.0.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "NativeHistogramConfig.golden",
		},
		{
			version: "v2.54.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "native-histograms/NativeHistogramConfigMissConvertClassicHistogramsToNHCB.golden",
		},
		{
			version: "v2.46.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "native-histograms/NativeHistogramConfigWithMissNativeHistogramMinBucketFactor.golden",
		},
		{
			version: "v2.44.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "NativeHistogramConfigWithMissALL.golden",
		},
		{
			version: "3.0.0-rc.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "NativeHistogramConfigAlwaysScrapeClassicHistograms.golden",
		},
		{
			version: "v3.8.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				ScrapeNativeHistograms:         ptr.To(true),
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "NativeHistogramConfigWithScrapeNativeHistograms.golden",
		},
		{
			version: "v3.7.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				ScrapeNativeHistograms:         ptr.To(true),
				NativeHistogramBucketLimit:     ptr.To(uint64(10)),
				ScrapeClassicHistograms:        ptr.To(true),
				NativeHistogramMinBucketFactor: ptr.To(resource.MustParse("12.124")),
				ConvertClassicHistogramsToNHCB: ptr.To(true),
			},
			golden: "NativeHistogramConfigMissScrapeNativeHistograms.golden",
		},
		{
			version: "v3.8.0",
			nativeHistogramConfig: monitoringv1.NativeHistogramConfig{
				ScrapeNativeHistograms: ptr.To(true),
			},
			golden: "NativeHistogramConfigOnlyScrapeNativeHistograms.golden",
		},
	} {
		t.Run(fmt.Sprintf("version=%s", tc.version), func(t *testing.T) {
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
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			}

			serviceMonitor.Spec.NativeHistogramConfig = tc.nativeHistogramConfig

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestBodySizeLimits(t *testing.T) {
	for _, tc := range []struct {
		version                     string
		enforcedBodySizeLimit       monitoringv1.ByteSize
		serviceMonitorBodySizeLimit *monitoringv1.ByteSize
		expectedErr                 error
		golden                      string
	}{
		{
			version:                     "v2.27.0",
			enforcedBodySizeLimit:       "1000MB",
			serviceMonitorBodySizeLimit: ptr.To[monitoringv1.ByteSize]("2GB"),
			golden:                      "BodySizeLimits_enforce_v2.27.0.golden",
		},
		{
			version:               "v2.28.0",
			enforcedBodySizeLimit: "1000MB",
			golden:                "BodySizeLimits_enforce1000MB_v2.28.0.golden",
		},
		{
			version:                     "v2.28.0",
			enforcedBodySizeLimit:       "4000MB",
			serviceMonitorBodySizeLimit: ptr.To[monitoringv1.ByteSize]("2GB"),
			golden:                      "BodySizeLimits_enforce2GB_v2.28.0.golden",
		},
		{
			version:               "v2.28.0",
			enforcedBodySizeLimit: "",
			golden:                "BodySizeLimits_enforce0MB_v2.28.0.golden",
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
					BodySizeLimit: tc.serviceMonitorBodySizeLimit,
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigBodySizeLimit(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.Version = "v2.28.0"

	scrapeConfig := monitoringv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testscrapeconfig1",
			Namespace: "default",
		},
		Spec: monitoringv1alpha1.ScrapeConfigSpec{
			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{
					Targets: []monitoringv1alpha1.Target{"localhost:9090"},
				},
			},
			BodySizeLimit: ptr.To[monitoringv1.ByteSize]("500MB"),
		},
	}

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		map[string]*monitoringv1alpha1.ScrapeConfig{
			"testscrapeconfig1": &scrapeConfig,
		},
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	golden.Assert(t, string(cfg), "ScrapeConfigBodySizeLimit.golden")
}

func TestMatchExpressionsServiceMonitor(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	golden.Assert(t, string(cfg), "MatchExpressionsServiceMonitor.golden")
}

func TestServiceMonitorEndpointFollowRedirects(t *testing.T) {
	for _, tc := range []struct {
		version         string
		expected        string
		golden          string
		followRedirects bool
	}{
		{
			version:         "v2.25.0",
			followRedirects: false,
			golden:          "ServiceMonitorEndpointFollowRedirects_FollowRedirectFalse_v2.25.0.golden",
		},
		{
			version:         "v2.25.0",
			followRedirects: true,
			golden:          "ServiceMonitorEndpointFollowRedirects_FollowRedirectTrue_v2.25.0.golden",
		},
		{
			version:         "v2.28.0",
			followRedirects: true,
			golden:          "ServiceMonitorEndpointFollowRedirects_FollowRedirectTrue_v2.28.0.golden",
		},
		{
			version:         "v2.28.0",
			followRedirects: false,
			golden:          "ServiceMonitorEndpointFollowRedirects_FollowRedirectFalse_v2.28.0.golden",
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
							Port:     "web",
							Interval: "30s",
							HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
								HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
									HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
										FollowRedirects: ptr.To(tc.followRedirects),
									},
								},
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestPodMonitorEndpointFollowRedirects(t *testing.T) {
	for _, tc := range []struct {
		version         string
		golden          string
		followRedirects bool
	}{
		{
			version:         "v2.25.0",
			followRedirects: false,
			golden:          "PodMonitorEndpointFollowRedirects_FollowRedirectsFalse_v2.25.0.golden",
		},
		{
			version:         "v2.25.0",
			followRedirects: true,
			golden:          "PodMonitorEndpointFollowRedirects_FollowRedirectsTrue_v2.25.0.golden",
		},
		{
			version:         "v2.28.0",
			followRedirects: true,
			golden:          "PodMonitorEndpointFollowRedirects_FollowRedirectsTrue_v2.28.0.golden",
		},
		{
			version:         "v2.28.0",
			followRedirects: false,
			golden:          "PodMonitorEndpointFollowRedirects_FollowRedirectsFalse_v2.28.0.golden",
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
							Port:     ptr.To("web"),
							Interval: "30s",
							HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
								HTTPConfig: monitoringv1.HTTPConfig{
									HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
										FollowRedirects: ptr.To(tc.followRedirects),
									},
								},
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestServiceMonitorEndpointEnableHttp2(t *testing.T) {
	for _, tc := range []struct {
		version     string
		golden      string
		enableHTTP2 bool
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			golden:      "ServiceMonitorEndpointEnableHttp2_EnableHTTP2False_v2.34.0.golden",
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			golden:      "ServiceMonitorEndpointEnableHttp2_EnableHTTP2True_v2.34.0.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			golden:      "ServiceMonitorEndpointEnableHttp2_EnableHTTP2True_v2.35.0.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			golden:      "ServiceMonitorEndpointEnableHttp2_EnableHTTP2False_v2.35.0.golden",
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
							Port:     "web",
							Interval: "30s",
							HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
								HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
									HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
										EnableHTTP2: ptr.To(tc.enableHTTP2),
									},
								},
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{
					"testservicemonitor1": &serviceMonitor,
				},
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestPodMonitorPhaseFilter(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							FilterRunning: ptr.To(false),
							Port:          ptr.To("test"),
						},
					},
				},
			},
		},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "PodMonitorPhaseFilter.golden")
}

func TestPodMonitorEndpointEnableHttp2(t *testing.T) {
	for _, tc := range []struct {
		version     string
		golden      string
		enableHTTP2 bool
	}{
		{
			version:     "v2.34.0",
			enableHTTP2: false,
			golden:      "PodMonitorEndpointEnableHttp2_EnableHTTP2False_v2.34.0.golden",
		},
		{
			version:     "v2.34.0",
			enableHTTP2: true,
			golden:      "PodMonitorEndpointEnableHttp2_EnableHTTP2True_v2.34.0.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: true,
			golden:      "PodMonitorEndpointEnableHttp2_EnableHTTP2True_v2.35.0.golden",
		},
		{
			version:     "v2.35.0",
			enableHTTP2: false,
			golden:      "PodMonitorEndpointEnableHttp2_EnableHTTP2False_v2.35.0.golden",
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
							Port:     ptr.To("web"),
							Interval: "30s",
							HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
								HTTPConfig: monitoringv1.HTTPConfig{
									HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
										EnableHTTP2: ptr.To(tc.enableHTTP2),
									},
								},
							},
						},
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{
					"testpodmonitor1": &podMonitor,
				},
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestRuntimeConfig(t *testing.T) {
	for _, tc := range []struct {
		Scenario string
		Version  string
		Runtime  *monitoringv1.RuntimeConfig
		Golden   string
	}{
		{
			Scenario: "Runtime GoGC is set to 25",
			Version:  "v2.53.0",
			Runtime: &monitoringv1.RuntimeConfig{
				GoGC: ptr.To(int32(25)),
			},
			Golden: "RuntimeConfig_GoGC25.golden",
		},
		{
			Scenario: "Runtime GoGC is set to 25 but unsupported Prometheus Version",
			Version:  "v2.52.0",
			Runtime: &monitoringv1.RuntimeConfig{
				GoGC: ptr.To(int32(25)),
			},
			Golden: "RuntimeConfig_GoGC_Not_Set.golden",
		},
		{
			Scenario: "Runtime GoGC not specified",
			Golden:   "RuntimeConfig_GoGC_Not_Set.golden",
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.Scenario), func(t *testing.T) {
			p := defaultPrometheus()
			if tc.Version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.Version
			}
			p.Spec.Runtime = tc.Runtime
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.Golden)
		})
	}
}

func TestStorageSettingMaxExemplars(t *testing.T) {
	for _, tc := range []struct {
		Scenario  string
		Version   string
		Exemplars *monitoringv1.Exemplars
		Golden    string
	}{
		{
			Scenario: "Exemplars maxSize is set to 5000000",
			Exemplars: &monitoringv1.Exemplars{
				MaxSize: ptr.To(int64(5000000)),
			},
			Golden: "StorageSettingMaxExemplars_MaxSize5000000.golden",
		},
		{
			Scenario: "max_exemplars is not set if version is less than v2.29.0",
			Version:  "v2.28.0",
			Exemplars: &monitoringv1.Exemplars{
				MaxSize: ptr.To(int64(5000000)),
			},
			Golden: "StorageSettingMaxExemplars_MaxSizeNotSet_v2.29.0.golden",
		},
		{
			Scenario: "Exemplars maxSize is not set",
			Golden:   "StorageSettingMaxExemplars_MaxSizeNotSetAtAll.golden",
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
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.Golden)
		})
	}
}

func TestTSDBConfig(t *testing.T) {
	for _, tc := range []struct {
		name    string
		p       *monitoringv1.Prometheus
		version string
		tsdb    *monitoringv1.TSDBSpec
		golden  string
	}{
		{
			name:   "no TSDB config",
			golden: "no_TSDB_config.golden",
		},
		{
			name:    "TSDB config < v2.39.0",
			version: "v2.38.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: ptr.To(monitoringv1.Duration("10m")),
			},
			golden: "TSDB_config_less_than_v2.39.0.golden",
		},
		{

			name: "TSDB config >= v2.39.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: ptr.To(monitoringv1.Duration("10m")),
			},
			golden: "TSDB_config_greater_than_or_equal_to_v2.39.0.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.tsdb != nil {
				p.Spec.TSDB = tc.tsdb
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestTSDBConfigPrometheusAgent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		p       *monitoringv1.Prometheus
		version string
		tsdb    *monitoringv1.TSDBSpec
		golden  string
	}{
		{
			name:   "PrometheusAgent no TSDB config",
			golden: "PrometheusAgent_no_TSDB_config.golden",
		},
		{
			name:    "PrometheusAgent TSDB config < v2.54.0",
			version: "v2.53.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: ptr.To(monitoringv1.Duration("10m")),
			},
			golden: "PrometheusAgent_TSDB_config_less_than_v2.53.0.golden",
		},
		{

			name:    "PrometheusAgent TSDB config >= v2.54.0",
			version: "v2.54.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: ptr.To(monitoringv1.Duration("10m")),
			},
			golden: "PrometheusAgent_TSDB_config_greater_than_or_equal_to_v2.54.0.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.tsdb != nil {
				p.Spec.TSDB = tc.tsdb
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateAgentConfiguration(
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeFailureLogFilePrometheusAgent(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		p                    *monitoringv1.Prometheus
		version              string
		scrapeFailureLogFile *string
		golden               string
	}{
		{
			name:   "PrometheusAgent no setting scrape failure log file",
			golden: "PrometheusAgent_no_scrapeFailureLogFile.golden",
		},
		{
			name:                 "PrometheusAgent version < v2.55.0",
			version:              "v2.54.0",
			scrapeFailureLogFile: ptr.To("file.log"),
			golden:               "PrometheusAgent_scrapeFailureLogFile_less_than_v2.54.0.golden",
		},
		{

			name:                 "PrometheusAgent version >= v2.55.0",
			version:              "v2.55.0",
			scrapeFailureLogFile: ptr.To("/tmp/file.log"),
			golden:               "PrometheusAgent_scrapeFailureLogFile_greater_than_or_equal_to_v2.55.0.golden",
		},
		{

			name:                 "PrometheusAgent version >= v2.55.0 and scrapeFailureLogFile with empty path",
			version:              "v2.55.0",
			scrapeFailureLogFile: ptr.To("file.log"),
			golden:               "PrometheusAgent_scrapeFailureLogFile_empty_path_v2.55.0.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.scrapeFailureLogFile != nil {
				p.Spec.ScrapeFailureLogFile = tc.scrapeFailureLogFile
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateAgentConfiguration(
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestPromAgentDaemonSetPodMonitorConfig(t *testing.T) {
	p := defaultPrometheus()
	cg := mustNewConfigGenerator(t, p)
	cg.daemonSet = true
	pmons := map[string]*monitoringv1.PodMonitor{
		"pm": defaultPodMonitor(),
	}
	cfg, err := cg.GenerateAgentConfiguration(
		nil,
		pmons,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "PromAgentDaemonSetPodMonitorConfig.golden")
}

func TestGenerateRelabelConfig(t *testing.T) {
	p := defaultPrometheus()

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
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
							MetricRelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "Drop",
									Regex:        "container_fs*",
									SourceLabels: []monitoringv1.LabelName{"__name__"},
								},
								{
									// Test empty replacement
									Action:      "Replace",
									Replacement: ptr.To(""),
									TargetLabel: "job",
								},
							},
							RelabelConfigs: []monitoringv1.RelabelConfig{
								{
									Action:       "Uppercase",
									SourceLabels: []monitoringv1.LabelName{"instance"},
									TargetLabel:  "instance",
								},
								{
									Action:       "Replace",
									Regex:        "(.+)(?::d+)",
									Replacement:  ptr.To("$1:9537"),
									SourceLabels: []monitoringv1.LabelName{"__address__"},
									TargetLabel:  "__address__",
								},
								{
									Action:      "Replace",
									Replacement: ptr.To("crio"),
									TargetLabel: "job",
								},
								{
									// Test empty replacement
									Action:      "Replace",
									Replacement: ptr.To(""),
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
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "GenerateRelabelConfig.golden")
}

// When adding new test cases the developer should specify a name, a Probe Spec
// (pbSpec) and an expectedConfig. (Optional) It's also possible to specify a
// function that modifies the default Prometheus CR used if necessary for the test
// case.
func TestProbeSpecConfig(t *testing.T) {
	for _, tc := range []struct {
		name      string
		patchProm func(*monitoringv1.Prometheus)
		pbSpec    monitoringv1.ProbeSpec
		golden    string
	}{
		{
			name:   "empty_probe",
			golden: "ProbeSpecConfig_empty_probe.golden",
			pbSpec: monitoringv1.ProbeSpec{},
		},
		{
			name:   "prober_spec",
			golden: "ProbeSpecConfig_prober_spec.golden",
			pbSpec: monitoringv1.ProbeSpec{
				ProberSpec: monitoringv1.ProberSpec{
					Scheme: ptr.To(monitoringv1.SchemeHTTP),
					URL:    "example.com",
					Path:   "/probe",
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
									Key: "proxy-header",
								},
							},
						},
					},
				},
			},
		},
		{
			name:   "targets_static_config",
			golden: "ProbeSpecConfig_targets_static_config.golden",
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
						RelabelConfigs: []monitoringv1.RelabelConfig{
							{
								TargetLabel: "foo",
								Replacement: ptr.To("bar"),
								Action:      "replace",
							},
							// Empty replacement case
							{
								TargetLabel: "foobar",
								Replacement: ptr.To(""),
								Action:      "replace",
							},
						},
					},
				}},
		},
		{
			name:   "module_config",
			golden: "ProbeSpecConfig_module_config.golden",
			pbSpec: monitoringv1.ProbeSpec{
				Module: "http_2xx",
			},
		},
		{
			name:   "module_config_with_params",
			golden: "ProbeSpecConfig_module_config_with_params.golden",
			pbSpec: monitoringv1.ProbeSpec{
				Module: "http_2xx",
				Params: []monitoringv1.ProbeParam{
					{
						Name:   "foo",
						Values: []string{"bar"},
					},
				},
			},
		},
		{
			name:   "module_config_with_params_skip_module",
			golden: "ProbeSpecConfig_module_config_with_params_skip_module.golden",
			pbSpec: monitoringv1.ProbeSpec{
				Module: "http_2xx",
				Params: []monitoringv1.ProbeParam{
					{
						Name:   "foo",
						Values: []string{"bar"},
					},
					{
						Name:   "module",
						Values: []string{"tcp_connect"},
					},
				},
			},
		},
		{
			name:   "module_config_with_params_define_module_in_param",
			golden: "ProbeSpecConfig_module_config_with_params_define_module_in_param.golden",
			pbSpec: monitoringv1.ProbeSpec{
				Params: []monitoringv1.ProbeParam{
					{
						Name:   "foo",
						Values: []string{"bar"},
					},
					{
						Name:   "module",
						Values: []string{"tcp_connect"},
					},
				},
			},
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

			s := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
					},
				},
			)

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				pbs,
				nil,
				s,
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})

	}
}

// When adding new test cases the developer should specify a name, a ScrapeConfig Spec
// (scSpec) and an expected config in golden file (.golden file in testdata folder). (Optional) It's also possible to specify a
// function (patchProm) that modifies the default Prometheus CR used if necessary for the test
// case.
func TestScrapeConfigSpecConfig(t *testing.T) {
	refreshInterval := monitoringv1.Duration("5m")
	for _, tc := range []struct {
		name            string
		version         string
		patchProm       func(*monitoringv1.Prometheus)
		scSpec          monitoringv1alpha1.ScrapeConfigSpec
		inlineTLSConfig bool
		golden          string
	}{
		{
			name:   "empty_scrape_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{},
			golden: "ScrapeConfigSpecConfig_Empty.golden",
		},
		{
			name: "explicit_job_name",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				JobName: ptr.To("explicit-test-scrape-config3"),
			},
			golden: "ScrapeConfigSpecConfig_WithJobName.golden",
		},
		{
			name: "explicit_job_name_with_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				JobName: ptr.To("explicit-test-scrape-config5"),
				RelabelConfigs: []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						Regex:        "(.+)(?::d+)",
						Replacement:  ptr.To("$1:9537"),
						SourceLabels: []monitoringv1.LabelName{"__address__"},
						TargetLabel:  "__address__",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_WithJobNameAndRelabelConfig.golden",
		},
		{
			name: "shard_config",
			patchProm: func(p *monitoringv1.Prometheus) {
				p.Spec.Shards = ptr.To(int32(2))
			},
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"localhost:9100"},
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_Sharded.golden",
		},
		{
			name: "already_sharded_config",
			patchProm: func(p *monitoringv1.Prometheus) {
				p.Spec.Shards = ptr.To(int32(2))
			},
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"localhost:9100"},
						Labels: map[string]string{
							"label1": "value1_sharded",
						},
					},
				},
				RelabelConfigs: []monitoringv1.RelabelConfig{
					{
						SourceLabels: []monitoringv1.LabelName{"__address__"},
						TargetLabel:  "__tmp_hash",
						Modulus:      999,
						Action:       "hashmod",
					},
					{
						SourceLabels: []monitoringv1.LabelName{"__tmp_hash", "__tmp_disable_sharding"},
						Regex:        "$(SHARD);|.+;.+",
						Action:       "keep",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_Already_Sharded.golden",
		},
		{
			name: "static_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"localhost:9100"},
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_Static.golden",
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
			golden: "ScrapeConfigSpecConfig_FileSD.golden",
		},
		{
			name: "http_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL:             "http://localhost:9100/sd.json",
						RefreshInterval: &refreshInterval,
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
										Key: "proxy-header",
									},
								},
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD.golden",
		},
		{
			name: "metrics_path",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricsPath: ptr.To("/metrics"),
			},
			golden: "ScrapeConfigSpecConfig_MetricPath.golden",
		},
		{
			name: "empty_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				RelabelConfigs: []monitoringv1.RelabelConfig{},
			},
			golden: "ScrapeConfigSpecConfig_EmptyRelabelConfig.golden",
		},
		{
			name: "non_empty_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				RelabelConfigs: []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						Regex:        "(.+)(?::d+)",
						Replacement:  ptr.To("$1:9537"),
						SourceLabels: []monitoringv1.LabelName{"__address__"},
						TargetLabel:  "__address__",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NonEmptyRelabelConfig.golden",
		},
		{
			name: "honor_timestamp",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HonorTimestamps: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_HonorTimeStamp.golden",
		},
		{
			name: "track_timestamps_staleness",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				TrackTimestampsStaleness: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_TrackTimestampsStaleness.golden",
		},
		{
			name: "honor_labels",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HonorLabels: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_HonorLabels.golden",
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
								Key: "username-sd",
							},
							Password: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "password-sd",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_BasicAuth.golden",
		},
		{
			name: "authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				Authorization: &monitoringv1.SafeAuthorization{
					Credentials: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "auth",
						},
						Key: "scrape-key",
					},
				},
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "auth",
								},
								Key: "http-sd-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_Authorization.golden",
		},
		{
			name: "inline_tlsconfig",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				TLSConfig: &monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "cert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls",
						},
						Key: "private-key",
					},
				},
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							InsecureSkipVerify: ptr.To(false),
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca2",
								},
							},
						},
					},
				},
			},
			inlineTLSConfig: true,
			golden:          "ScrapeConfigSpecConfig_Inline_TLSConfig.golden",
		},
		{
			name: "tlsconfig",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				TLSConfig: &monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "cert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls",
						},
						Key: "private-key",
					},
				},
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							InsecureSkipVerify: ptr.To(true),
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca2",
								},
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_TLSConfig.golden",
		},
		{
			name: "scheme",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				Scheme: ptr.To(monitoringv1.SchemeHTTPS),
			},
			golden: "ScrapeConfigSpecConfig_Scheme.golden",
		},
		{
			name: "limits",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				SampleLimit:           ptr.To(uint64(10000)),
				TargetLimit:           ptr.To(uint64(1000)),
				LabelLimit:            ptr.To(uint64(50)),
				LabelNameLengthLimit:  ptr.To(uint64(40)),
				LabelValueLengthLimit: ptr.To(uint64(30)),
			},
			golden: "ScrapeConfigSpecConfig_Limits.golden",
		},
		{
			name: "params",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricsPath: ptr.To("/federate"),
				Params:      map[string][]string{"match[]": {"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"}},
			},
			golden: "ScrapeConfigSpecConfig_Params.golden",
		},
		{
			name: "scrape_interval",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeInterval: ptr.To(monitoringv1.Duration("15s")),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeInterval.golden",
		},
		{
			name: "scrape_timeout",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeTimeout: ptr.To(monitoringv1.Duration("10s")),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeTimeout.golden",
		},
		{
			name: "scrape_protocols",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeProtocols: []monitoringv1.ScrapeProtocol{
					monitoringv1.ScrapeProtocol("PrometheusProto"),
					monitoringv1.ScrapeProtocol("OpenMetricsText1.0.0"),
					monitoringv1.ScrapeProtocol("OpenMetricsText0.0.1"),
					monitoringv1.ScrapeProtocol("PrometheusText0.0.4"),
				},
			},
			golden: "ScrapeConfigSpecConfig_ScrapeProtocols.golden",
		},
		{
			name:    "fallback_scrape_protocol",
			version: "v3.0.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeFallbackProtocol.golden",
		},
		{
			name:    "fallback_scrape_protocol_with_unsupported_version",
			version: "v2.55.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeFallbackProtocol_OldVersion.golden",
		},
		{
			name: "non_empty_metric_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricRelabelConfigs: []monitoringv1.RelabelConfig{
					{
						Regex:  "noisy_labels.*",
						Action: "labeldrop",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NonEmptyMetricRelabelConfig.golden",
		},
		{
			name: "proxy_settings",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
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
								Key: "proxy-header",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_ProxySettings.golden",
		},
		{
			name: "proxy_settings_with_muti_proxy_connect_header_values",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
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
								Key: "proxy-header",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "token",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "proxy-header",
							},
						},
						"token": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "token",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "token",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_ProxySettingsWithMutiProxyConnectHeaderValues.golden",
		},
		{
			name: "proxy_settings_incompatible_prometheus_version",
			patchProm: func(p *monitoringv1.Prometheus) {
				p.Spec.CommonPrometheusFields.Version = "v2.42.0"
			},
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
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
								Key: "proxy-header",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_ProxySettingsIncompatiblePrometheusVersion.golden",
		},
		{
			name: "dns_sd_config-srv-record",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"web.example.com"},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_SRVRecord.golden",
		},
		{
			name: "dns_sd_config-a-record",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
						Port:  ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_ARecord.golden",
		},
		{
			name:    "dns_sd_config-ns-record",
			version: "v2.49.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeNS),
						Port:  ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_NSRecord.golden",
		},
		{
			name:    "dns_sd_config-ns-record-unsupported-version",
			version: "v2.48.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeNS),
						Port:  ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_NSRecord_OldVersion.golden",
		},
		{
			name:    "dns_sd_config-mx-record",
			version: "v2.39.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeMX),
						Port:  ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_MXRecord.golden",
		},
		{
			name:    "dns_sd_config-mx-record-unsupported-version",
			version: "v2.28.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeNS),
						Port:  ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_NSRecord_OldVersion.golden",
		},
		{
			name: "enable_compression_is_set_to_true",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableCompression: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_EnableCompression_True.golden",
		},
		{
			name: "enable_compression_is_set_to_false",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableCompression: ptr.To(false),
			},
			golden: "ScrapeConfigSpecConfig_EnableCompression_False.golden",
		},
		{
			name:    "enable_http2_is_set_to_true_unsupported",
			version: "v2.34.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableHTTP2: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_EnableHTTP2_Unsupported.golden",
		},
		{
			name:    "enable_http2_is_set_to_false_unsupported",
			version: "v2.34.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableHTTP2: ptr.To(false),
			},
			golden: "ScrapeConfigSpecConfig_EnableHTTP2_Unsupported.golden",
		},
		{
			name:    "enable_http2_is_set_to_true",
			version: "v2.35.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableHTTP2: ptr.To(true),
			},
			golden: "ScrapeConfigSpecConfig_EnableHTTP2_True.golden",
		},
		{
			name:    "enable_http2_is_set_to_false",
			version: "v2.35.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EnableHTTP2: ptr.To(false),
			},
			golden: "ScrapeConfigSpecConfig_EnableHTTP2_False.golden",
		},
		{
			name:    "config_oauth",
			version: "v2.27.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_WithOAuth.golden",
		},
		{
			name:    "config_oauth_unsupported",
			version: "v2.26.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_WithOAuth_Unsupported.golden",
		},
		{
			name:    "config_NameValidationScheme_utf8",
			version: "v3.0.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameValidationScheme: ptr.To(monitoringv1.UTF8NameValidationScheme),
			},
			golden: "NameValidationScheme_utf8.golden",
		},
		{
			name:    "config_NameValidationScheme_legacy",
			version: "v3.0.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameValidationScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			},
			golden: "NameValidationScheme_legacy.golden",
		},
		{
			name:    "config_NameValidationScheme_unsupported",
			version: "v2.55.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameValidationScheme: ptr.To(monitoringv1.UTF8NameValidationScheme),
			},
			golden: "NameValidationScheme_unsupported.golden",
		},
		{
			name:    "config_NameEscapingScheme_AllowUTF8",
			version: "v3.4.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameEscapingScheme: ptr.To(monitoringv1.AllowUTF8NameEscapingScheme),
			},
			golden: "NameEscapingScheme_AllowUTF8.golden",
		},
		{
			name:    "config_NameEscapingScheme_Underscores",
			version: "v3.4.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameEscapingScheme: ptr.To(monitoringv1.UnderscoresNameEscapingScheme),
			},
			golden: "NameEscapingScheme_Underscores.golden",
		},
		{
			name:    "config_NameEscapingScheme_Dots",
			version: "v3.4.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameEscapingScheme: ptr.To(monitoringv1.DotsNameEscapingScheme),
			},
			golden: "NameEscapingScheme_Dots.golden",
		},
		{
			name:    "config_NameEscapingScheme_Values",
			version: "v3.4.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameEscapingScheme: ptr.To(monitoringv1.ValuesNameEscapingScheme),
			},
			golden: "NameEscapingScheme_Values.golden",
		},
		{
			name:    "config_NameEscapingScheme_Unsupported",
			version: "v3.3.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NameEscapingScheme: ptr.To(monitoringv1.ValuesNameEscapingScheme),
			},
			golden: "NameEscapingScheme_Unsupported.golden",
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
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.patchProm != nil {
				tc.patchProm(p)
			}

			cg := mustNewConfigGenerator(t, p)
			cg.inlineTLSConfig = tc.inlineTLSConfig

			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"username":     []byte("scrape-bob"),
						"password":     []byte("scrape-alice"),
						"username-sd":  []byte("http-sd-bob"),
						"password-sd":  []byte("http-sd-alice"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("bar-value"),
						"token":        []byte("bar-value"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "auth",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"scrape-key":  []byte("scrape-secret"),
						"http-sd-key": []byte("http-sd-secret"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tls",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"ca":          []byte("ca"),
						"ca2":         []byte("ca2"),
						"cert":        []byte("cert"),
						"private-key": []byte("private-key"),
					},
				},
			)

			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}
func TestScrapeConfigSpecConfigWithHTTPSD(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scSpec  monitoringv1alpha1.ScrapeConfigSpec
		golden  string
		version string
	}{
		{
			name: "http_sd_config_with_proxy_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL:             "http://localhost:9100/sd.json",
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
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
										Key: "proxy-header",
									},
								},
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD_with_ProxyConfig.golden",
		},
		{
			name: "http_sd_config_basic_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						BasicAuth: &monitoringv1.BasicAuth{
							Username: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Username",
							},
							Password: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Password",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD_with_BasicAuth.golden",
		}, {
			name: "http_sd_config_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD_with_Authorization.golden",
		}, {
			name: "http_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_HTTPSD_with_OAuth.golden",
		}, {
			name: "http_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://localhost:9100/sd.json",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD_with_TLSConfig.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"Username":     []byte("kube-admin"),
						"Password":     []byte("password"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			p.Spec.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})

	}
}

func TestScrapeConfigSpecConfigWithKubernetesSD(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scSpec  monitoringv1alpha1.ScrapeConfigSpec
		golden  string
		version string
	}{
		{
			name: "kubernetes_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role:      monitoringv1alpha1.KubernetesRoleNode,
						APIServer: ptr.To("https://kubernetes.default.svc"),
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD.golden",
		},
		{
			name: "kubernetes_sd_config_with_namespace_discovery_and_own_namespace",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRolePod,
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							IncludeOwnNamespace: ptr.To(true),
							Names:               []string{"ns1", "ns2"},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_NamespaceDiscovery_and_OwnNamespace.golden",
		},
		{
			name: "kubernetes_sd_config_with_namespace_discovery",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRolePod,
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							Names: []string{"ns1", "ns2"},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_NamespaceDiscovery.golden",
		},
		{
			name: "kubernetes_sd_config_with_supported_role_attach_metadata",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRolePod,
						AttachMetadata: &monitoringv1alpha1.AttachMetadata{
							Node: ptr.To(true),
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_AttachMetadata.golden",
		},
		{
			name: "kubernetes_sd_config_with_unsupported_role_attach_metadata",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleService,
						AttachMetadata: &monitoringv1alpha1.AttachMetadata{
							Node: ptr.To(true),
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_Unsupported_Role_AttachMetadata.golden",
		},
		{
			name: "kubernetes_sd_config_with_selectors",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  monitoringv1alpha1.KubernetesRoleNode,
								Label: ptr.To("component=executor"),
							},
						},
					},
				},
			},
			version: "2.18.0",
			golden:  "ScrapeConfigSpecConfig_K8SSD_with_Selectors.golden",
		},
		{
			name: "kubernetes_sd_config_with_selectors_unsupported_version",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  monitoringv1alpha1.KubernetesRoleNode,
								Label: ptr.To("type=infra"),
								Field: ptr.To("spec.unschedulable=false"),
							},
						},
					},
				},
			},
			version: "2.16.0",
			golden:  "ScrapeConfigSpecConfig_K8SSD_with_Selectors_Unsupported_Version.golden",
		},
		{
			name: "kubernetes_sd_config_basic_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role:      monitoringv1alpha1.KubernetesRoleNode,
						APIServer: ptr.To("https://kubernetes.default.svc"),
						BasicAuth: &monitoringv1.BasicAuth{
							Username: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Username",
							},
							Password: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Password",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_BasicAuth.golden",
		}, {
			name: "kubernetes_sd_config_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role:      monitoringv1alpha1.KubernetesRoleNode,
						APIServer: ptr.To("https://kubernetes.default.svc"),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_Authorization.golden",
		}, {
			name: "kubernetes_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role:      monitoringv1alpha1.KubernetesRoleNode,
						APIServer: ptr.To("https://kubernetes.default.svc"),
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_K8SSD_with_OAuth.golden",
		}, {
			name: "kubernetes_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role:      monitoringv1alpha1.KubernetesRoleNode,
						APIServer: ptr.To("https://kubernetes.default.svc"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"Username":     []byte("kube-admin"),
						"Password":     []byte("password"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			p.Spec.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})

	}
}

func TestScrapeConfigSpecConfigWithConsulSD(t *testing.T) {
	for _, tc := range []struct {
		name      string
		patchProm func(*monitoringv1.Prometheus)
		scSpec    monitoringv1alpha1.ScrapeConfigSpec
		version   string
		golden    string
	}{
		{
			name: "consul_scrape_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:       "localhost",
						Datacenter:   ptr.To("we1"),
						Namespace:    ptr.To("observability"),
						Partition:    ptr.To("1"),
						Scheme:       ptr.To(monitoringv1.SchemeHTTPS),
						Services:     []string{"prometheus", "alertmanager"},
						Tags:         []string{"tag1"},
						TagSeparator: ptr.To(";"),
						NodeMeta: map[string]string{
							"service": "service_name",
							"name":    "node_name",
						},
						AllowStale:      ptr.To(false),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHttp2:     ptr.To(true),
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "foo",
							},
							Key: "token",
						},
					},
				},
			},
			golden: "ConsulScrapeConfig.golden",
		}, {
			name:    "consul_scrape_config_with_filter",
			version: "3.0.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost",
						Filter: ptr.To("Meta.env == \"qa\""),
					},
				},
			},
			golden: "ConsulScrapeConfigWithFilter.golden",
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
			golden: "ConsulScrapeConfigBasicAuth.golden",
		}, {
			name: "consul_scrape_config_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "auth",
								},
								Key: "token",
							},
						},
					},
				},
			},
			golden: "ConsulScrapeConfigAuthorization.golden",
		}, {
			name: "consul_scrape_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "localhost:8500",
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ConsulScrapeConfigOAuth.golden",
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
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ConsulScrapeConfigTLSConfig.golden",
		},
		{
			name: "Consul SD config with PathPrefix",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:     "server",
						PathPrefix: ptr.To("example-path-prefix"),
					},
				},
			},
			version: "2.45.0",
			golden:  "ConsulScrapeConfigPathPrefix.golden",
		},
		{
			name: "Consul SD config with PathPrefix but unsupported version",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:     "server",
						PathPrefix: ptr.To("example-path-prefix"),
					},
				},
			},
			version: "2.35.0",
			golden:  "ConsulScrapeConfigPathPrefix_unsupported_version.golden",
		},
		{
			name: "Consul SD config with Namespace",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:    "server",
						Namespace: ptr.To("example-namespace"),
					},
				},
			},
			version: "2.31.0",
			golden:  "ConsulScrapeConfigNamespace.golden",
		},
		{
			name: "Consul SD config with Namespace but unsupported version",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
					{
						Server:    "server",
						Namespace: ptr.To("example-namespace"),
					},
				},
			},
			version: "2.21.0",
			golden:  "ConsulScrapeConfigNamespace_unsupported_version.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"username":     []byte("consul-sd-bob"),
						"password":     []byte("consul-sd-alice"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "auth",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"token": []byte("secret"),
					},
				},
			)

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
			p.Spec.Version = tc.version

			if tc.patchProm != nil {
				tc.patchProm(p)
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})

	}
}

func TestScrapeConfigSpecConfigWithEC2SD(t *testing.T) {
	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-access-api",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"accessKey": []byte("access-key"),
			"secretKey": []byte("secret-key"),
		},
	}
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
		version     string
		expectedErr bool
	}{
		{
			name: "ec2_sd_config_valid_with_api_keys",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "aws-access-api",
							},
							Key: "accessKey",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "aws-access-api",
							},
							Key: "secretKey",
						},
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EC2SDConfigValidAPIKeys.golden",
		},
		{
			name: "ec2_sd_config_valid_with_role_arn",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region:          ptr.To("us-east-1"),
						RoleARN:         ptr.To("arn:aws:iam::123456789:role/prometheus-role"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EC2SDConfigValidRoleARN.golden",
		},
		{
			name: "ec2_sd_config_valid_with_filters",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region:          ptr.To("us-east-1"),
						RoleARN:         ptr.To("arn:aws:iam::123456789:role/prometheus-role"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
						Filters: []monitoringv1alpha1.Filter{
							{
								Name:   "tag:environment",
								Values: []string{"prod"},
							},
							{
								Name:   "tag:service",
								Values: []string{"web", "db"},
							},
						},
					},
				},
			},
			version: "2.3.0",
			golden:  "ScrapeConfigSpecConfig_EC2SDConfigFilters.golden",
		},
		{
			name: "ec2_sd_config_valid_with_filters_unsupported_version",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region:          ptr.To("us-east-1"),
						RoleARN:         ptr.To("arn:aws:iam::123456789:role/prometheus-role"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
						Filters: []monitoringv1alpha1.Filter{
							{
								Name:   "tag:environment",
								Values: []string{"prod"},
							},
							{
								Name:   "tag:service",
								Values: []string{"web", "db"},
							},
						},
					},
				},
			},
			version: "2.2.0",
			golden:  "ScrapeConfigSpecConfig_EC2SDConfigFilters_Unsupported_Version.golden",
		},
		{
			name: "ec2_sd_config_invalid",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong-secret-name",
							},
							Key: "accessKey",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "aws-access-api",
							},
							Key: "secretKey",
						},
					},
				},
			},
			expectedErr: true,
		},
		{
			name: "ec2_sd_config_empty",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{},
			},
			golden: "ScrapeConfigSpecConfig_EC2SDConfigEmpty.golden",
		},
		{
			name: "ec2_sd_config_proxyconfig",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
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
										Key: "proxy-header",
									},
								},
							},
						},
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EC2SD_withProxyConfig.golden",
		},
		{
			name: "ec2_sd_config_http_and_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
					},
				},
			},
			version: "2.41.0",
			golden:  "ScrapeConfigSpecConfig_EC2SD_with_TLSConfig.golden",
		},
		{
			name: "ec2_sd_config_http_and_tls_unsupported_version",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
					},
				},
			},
			version: "2.31.0",
			golden:  "ScrapeConfigSpecConfig_EC2SD_with_TLSConfig_Unsupported_Version.golden",
		}} {
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
			p.Spec.Version = tc.version

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				scs,
				assets.NewTestStoreBuilder(sec),
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithAzureSD(t *testing.T) {
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
		expectedErr bool
	}{
		{
			name: "azure_sd_config_valid",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						Environment:          ptr.To("AzurePublicCloud"),
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						TenantID:             ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID:             ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "clientSecret",
						},
						ResourceGroup:   ptr.To("my-resource-group"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid.golden",
		},
		{
			name: "azure_sd_config_valid_sdk_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						Environment:          ptr.To("AzurePublicCloud"),
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						ResourceGroup:        ptr.To("my-resource-group"),
						RefreshInterval:      ptr.To(monitoringv1.Duration("30s")),
						Port:                 ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid_SDKAuth.golden",
		},
		{
			name: "azure_sd_config_valid_managedidendity_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						Environment:          ptr.To("AzurePublicCloud"),
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeManagedIdentity),
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid_ManagedIdentityAuth.golden",
		},
		{
			name: "azure_sd_config_with_subscription_id",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValidWithSubscriptionID.golden",
		},
		{
			name: "azure_sd_config_invalid",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong-secret-name",
							},
							Key: "clientSecret",
						},
					},
				},
			},
			expectedErr: true,
		},
		{
			name: "azure_sd_config_with_http_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid_http_config.golden",
		},
		{
			name: "azure_sd_config_with_http_config_basic_auth_and_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "key",
							},
						},
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
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid_basic_auth_and_TLS.golden",
		},
		{
			name: "azure_sd_config_with_http_config_with_oauth2",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
					{
						SubscriptionID:       "11AAAA11-A11A-111A-A111-1111A1111A11",
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_AzureSDConfigValid_with_oauth2.golden",
		},
		{
			name: "azure_sd_config_empty",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{},
			},
			golden: "ScrapeConfigSpecConfig_AzureSDConfigEmpty.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
						"clientSecret": []byte("my-secret"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"username": []byte("value"),
						"password": []byte("value"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithGCESD(t *testing.T) {
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
		expectedErr bool
	}{
		{
			name: "gce_sd_config_valid",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
					{
						Project:         "devops-dev",
						Zone:            "us-west-1",
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_GCESDConfigValid.golden",
		},
		{
			name: "gce_sd_config_valid_all_options",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
					{
						Project:         "devops-dev",
						Zone:            "us-west-1",
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
						Filter:          ptr.To("filter-expression"),
						TagSeparator:    ptr.To("tag-value"),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_GCESDConfigValid_all_options.golden",
		},
		{
			name: "gce_sd_config_valid_required_fields",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
					{
						Project: "devops-dev",
						Zone:    "us-west-1",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_GCESDConfigValidRequiredFields.golden",
		},
		{
			name: "gce_sd_config_empty",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				GCESDConfigs: []monitoringv1alpha1.GCESDConfig{},
			},
			golden: "ScrapeConfigSpecConfig_GCESDConfigEmpty.golden",
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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				scs,
				assets.NewTestStoreBuilder(),
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithOpenStackSD(t *testing.T) {
	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-access-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"password":                     []byte("password"),
			"applicationCredentialsSecret": []byte("application-credentials"),
		},
	}
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
		expectedErr bool
	}{
		{
			name: "openstack_sd_config_valid",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:             monitoringv1alpha1.OpenStackRoleInstance,
						Region:           "region-1",
						IdentityEndpoint: ptr.To("http://identity.example.com:5000/v2.0"),
						Username:         ptr.To("nova-user-1"),
						Password: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "openstack-access-secret",
							},
							Key: "password",
						},
						DomainName:      ptr.To("devops-project-1"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_OpenStackSDConfigValid.golden",
		},
		{
			name: "openstack_sd_config_invalid_secret_ref",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleInstance,
						Region: "region-1",
						ApplicationCredentialSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "openstack-access-secret",
							},
							Key: "invalid-key",
						},
					},
				},
			},
			expectedErr: true,
		},
		{
			name: "openstack_sd_config_empty",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{},
			},
			golden: "ScrapeConfigSpecConfig_OpenStackSDConfigEmpty.golden",
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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				scs,
				assets.NewTestStoreBuilder(sec),
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithDigitalOceanSD(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scSpec  monitoringv1alpha1.ScrapeConfigSpec
		golden  string
		version string
	}{
		{
			name: "digitalocean_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Port:            ptr.To(int32(9100)),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
					},
				},
			},
			version: "2.40.0",
			golden:  "ScrapeConfigSpecConfig_DigitalOceanSD.golden",
		}, {
			name: "digitalocean_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						OAuth2: &monitoringv1.OAuth2{
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
			version: "2.40.0",
			golden:  "ScrapeConfigSpecConfig_DigitalOceanSD_with_OAuth.golden",
		}, {
			name: "digitalocean_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			version: "2.40.0",
			golden:  "ScrapeConfigSpecConfig_DigitalOceanSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			p.Spec.Version = tc.version
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithDockerSDConfig(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		scSpec  monitoringv1alpha1.ScrapeConfigSpec
		golden  string
	}{
		{
			name: "docker_sd_config_with_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSDConfigs: []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects:    ptr.To(true),
						EnableHTTP2:        ptr.To(true),
						Port:               ptr.To(int32(9100)),
						RefreshInterval:    ptr.To(monitoringv1.Duration("30s")),
						HostNetworkingHost: ptr.To("localhost"),
						Filters: []monitoringv1alpha1.Filter{
							{Name: "dummy_label_1",
								Values: []string{"dummy_value_1"}},
							{Name: "dummy_label_2",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
							{Name: "a_dummy_label_1",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerSDConfig.golden",
		},
		{
			name: "docker_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSDConfigs: []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						OAuth2: &monitoringv1.OAuth2{
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
						Filters: []monitoringv1alpha1.Filter{
							{Name: "dummy_label_1",
								Values: []string{"dummy_value_1"}},
							{Name: "dummy_label_2",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerSD_with_OAuth.golden",
		},
		{
			name: "docker_sd_config_basicauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSDConfigs: []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Filters: []monitoringv1alpha1.Filter{
							{Name: "dummy_label_1",
								Values: []string{"dummy_value_1"}},
							{Name: "dummy_label_2",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
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
			golden: "ScrapeConfigSpecConfig_DockerSD_with_BasicAuth.golden",
		},
		{
			name:    "docker_sd_config_match_first_network",
			version: "v2.54.1",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSDConfigs: []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Filters: []monitoringv1alpha1.Filter{
							{Name: "dummy_label_1",
								Values: []string{"dummy_value_1"}},
							{Name: "dummy_label_2",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
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
						MatchFirstNetwork: ptr.To(true),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerSD_with_MatchFirstNetwork.golden",
		},
		{
			name:    "docker_sd_config_match_first_network_with_old_version",
			version: "v2.53.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSDConfigs: []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Filters: []monitoringv1alpha1.Filter{
							{Name: "dummy_label_1",
								Values: []string{"dummy_value_1"}},
							{Name: "dummy_label_2",
								Values: []string{"dummy_value_2", "dummy_value_3"}},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
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
						MatchFirstNetwork: ptr.To(true),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerSD_with_MatchFirstNetwork_OldVersion.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithLinodeSDConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "linode_sd_config_with_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LinodeSDConfigs: []monitoringv1alpha1.LinodeSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						Port:            ptr.To(int32(9100)),
						TagSeparator:    ptr.To(","),
						RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
						EnableHTTP2:     ptr.To(true),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LinodeSDConfig.golden",
		},
		{
			name: "linode_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LinodeSDConfigs: []monitoringv1alpha1.LinodeSDConfig{
					{
						OAuth2: &monitoringv1.OAuth2{
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
						TagSeparator:    ptr.To(","),
						RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
						EnableHTTP2:     ptr.To(true),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LinodeSD_with_OAuth.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}
func TestScrapeConfigSpecConfigWithHetznerSD(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		scSpec  monitoringv1alpha1.ScrapeConfigSpec
		golden  string
	}{
		{
			name: "hetzner_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Port:            ptr.To(int32(9100)),
						RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSD.golden",
		},
		{
			name:    "hetzner_sd_config_label_selector",
			version: "v3.5.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Port:            ptr.To(int32(9100)),
						RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
						LabelSelector:   ptr.To("label_value"),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSDLabelSelector.golden",
		},
		{
			name:    "hetzner_sd_config_no_label_selector",
			version: "v3.0.0",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Port:            ptr.To(int32(9100)),
						RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
						LabelSelector:   ptr.To("label_value"),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSDNoLabelSelector.golden",
		},
		{
			name: "hetzner_sd_config_basic_auth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						BasicAuth: &monitoringv1.BasicAuth{
							Username: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Username",
							},
							Password: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "Password",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSD_with_BasicAuth.golden",
		}, {
			name: "hetzner_sd_config_authorization",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "token",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSD_with_Authorization.golden",
		}, {
			name: "hetzner_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_HetznerSD_with_OAuth.golden",
		}, {
			name: "hetzner_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				HetznerSDConfigs: []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HetznerSD_with_TLSConfig.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"Username":     []byte("kube-admin"),
						"Password":     []byte("password"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAppendNameValidationScheme(t *testing.T) {
	testCases := []struct {
		name                 string
		version              string
		nameValidationScheme *monitoringv1.NameValidationSchemeOptions
		expectedCfg          string
	}{
		{
			name:                 "UTF8 nameValidationScheme withPrometheus Version 3",
			version:              "v3.0.0",
			nameValidationScheme: ptr.To(monitoringv1.UTF8NameValidationScheme),
			expectedCfg:          "NameValidationSchemeUTF8WithPrometheusV3.golden",
		},
		{
			name:                 "Legacy nameValidationScheme with Prometheus Version 3",
			version:              "v3.0.0",
			nameValidationScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			expectedCfg:          "NameValidationSchemeLegacyWithPrometheusV3.golden",
		},
		{
			name:                 "Legacy nameValidationScheme with Prometheus Version 3",
			version:              "v2.55.0",
			nameValidationScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			expectedCfg:          "NameValidationSchemeLegacyWithPrometheusLowerThanV3.golden",
		},
		{
			name:                 "Legacy nameValidationScheme with Prometheus Version 3",
			version:              "v2.55.0",
			nameValidationScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			expectedCfg:          "NameValidationSchemeLegacyWithPrometheusLowerThanV3.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.nameValidationScheme != nil {
				p.Spec.CommonPrometheusFields.NameValidationScheme = tc.nameValidationScheme
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)

			golden.Assert(t, string(cfg), tc.expectedCfg)
		})
	}
}

func TestAppendNameEscapingScheme(t *testing.T) {
	testCases := []struct {
		name               string
		version            string
		nameEscapingScheme *monitoringv1.NameEscapingSchemeOptions
		expectedCfg        string
	}{
		{
			name:               "AllowUTF8 nameEscapingScheme withPrometheus Version 3",
			version:            "v3.4.0",
			nameEscapingScheme: ptr.To(monitoringv1.AllowUTF8NameEscapingScheme),
			expectedCfg:        "NameEscapingSchemeUTF8WithPrometheusV3.golden",
		},
		{
			name:               "Underscores nameEscapingScheme with Prometheus Version 3",
			version:            "v3.4.0",
			nameEscapingScheme: ptr.To(monitoringv1.UnderscoresNameEscapingScheme),
			expectedCfg:        "NameEscapingSchemeUnderscoresWithPrometheusV3.golden",
		},
		{
			name:               "Underscores nameEscapingScheme with Prometheus Version 3",
			version:            "v2.55.0",
			nameEscapingScheme: ptr.To(monitoringv1.UnderscoresNameEscapingScheme),
			expectedCfg:        "NameEscapingSchemeUnderscoresWithPrometheusLowerThanV3.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.nameEscapingScheme != nil {
				p.Spec.CommonPrometheusFields.NameEscapingScheme = tc.nameEscapingScheme
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)

			golden.Assert(t, string(cfg), tc.expectedCfg)
		})
	}
}

func TestAppendConvertClassicHistogramsToNHCB(t *testing.T) {
	testCases := []struct {
		name                           string
		version                        string
		convertClassicHistogramsToNHCB *bool
		expectedCfg                    string
	}{
		{
			name:                           "ConvertClassicHistogramsToNHCB true with Prometheus Version 3.4",
			version:                        "v3.4.0",
			convertClassicHistogramsToNHCB: ptr.To(true),
			expectedCfg:                    "ConvertClassicHistogramsToNHCBTrueWithPrometheusV3.golden",
		},
		{
			name:                           "ConvertClassicHistogramsToNHCB false with Prometheus Version 3.4",
			version:                        "v3.4.0",
			convertClassicHistogramsToNHCB: ptr.To(false),
			expectedCfg:                    "ConvertClassicHistogramsToNHCBFalseWithPrometheusV3.golden",
		},
		{
			name:                           "ConvertClassicHistogramsToNHCB true with Prometheus Version 2",
			version:                        "v2.55.0",
			convertClassicHistogramsToNHCB: ptr.To(true),
			expectedCfg:                    "ConvertClassicHistogramsToNHCBTrueWithPrometheusLowerThanV3.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.convertClassicHistogramsToNHCB != nil {
				p.Spec.CommonPrometheusFields.ConvertClassicHistogramsToNHCB = tc.convertClassicHistogramsToNHCB
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)

			golden.Assert(t, string(cfg), tc.expectedCfg)
		})
	}
}

func TestOTLPConfig(t *testing.T) {
	testCases := []struct {
		otlpConfig    *monitoringv1.OTLPConfig
		nameValScheme *monitoringv1.NameValidationSchemeOptions
		name          string
		version       string
		expectedErr   bool
		golden        string
	}{
		{
			name:    "Config promote resource attributes",
			version: "v2.55.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{"aa", "bb", "cc"},
			},
			golden: "OTLPConfig_Config_promote_resource_attributes.golden",
		},
		{
			name:    "Config promote resource attributes with old version",
			version: "v2.53.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{"aa", "bb", "cc"},
			},
			expectedErr: true,
		},
		{
			name:    "Config Empty attributes",
			version: "v2.55.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{},
			},
			golden: "OTLPConfig_Config_empty_attributes.golden",
		},
		{
			name:    "Config otlp translation strategy",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: ptr.To(monitoringv1.UnderscoreEscapingWithSuffixes),
			},
			golden: "OTLPConfig_Config_translation_strategy.golden",
		},
		{
			name:    "Config Empty translation strategy",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: nil,
			},
			golden: "OTLPConfig_Config_empty_translation_strategy.golden",
		},
		{
			name:    "Config Empty translation strategy",
			version: "v2.55.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: ptr.To(monitoringv1.UnderscoreEscapingWithSuffixes),
			},
			golden: "OTLPConfig_Config_translation_strategy_with_unsupported_version.golden",
		},
		{
			name:    "Config Legacy nameValidationScheme with OTLP translation strategy NoUTF8",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: ptr.To(monitoringv1.NoUTF8EscapingWithSuffixes),
			},
			nameValScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			expectedErr:   true,
		},
		{
			name:    "Config Legacy nameValidationScheme with OTLP translation strategy UnderscoreEscapingWithSuffixes",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: ptr.To(monitoringv1.UnderscoreEscapingWithSuffixes),
			},
			nameValScheme: ptr.To(monitoringv1.LegacyNameValidationScheme),
			golden:        "OTLPConfig_Config_translation_strategy_with_suffixes_and_name_validation_scheme.golden",
		},
		{
			name:    "Config KeepIdentifyingResourceAttributes",
			version: "v3.1.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				KeepIdentifyingResourceAttributes: ptr.To(true),
			},
			golden: "OTLPConfig_Config_keep_identifying_resource_attributes.golden",
		},
		{
			name:    "Config ConvertHistogramsToNHCB old version",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				KeepIdentifyingResourceAttributes: ptr.To(false),
			},
			golden: "OTLPConfig_Config_keep_identifying_resource_attributes_with_old_version.golden",
		},
		{
			name:    "Config NoTranslation translation strategy",
			version: "v3.4.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: nil,
			},
			golden: "OTLPConfig_Config_translation_strategy_with_notranslation.golden",
		},
		{
			name:    "Config NoTranslation translation strategye unsupported version",
			version: "v3.0.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				TranslationStrategy: ptr.To(monitoringv1.NoTranslation),
			},
			expectedErr: true,
		},
		{
			name:    "Config ConvertHistogramsToNHCB",
			version: "v3.4.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				ConvertHistogramsToNHCB: ptr.To(true),
			},
			golden: "OTLPConfig_Config_convert_histograms_to_nhcb.golden",
		},
		{
			name:    "Config ConvertHistogramsToNHCB with old version",
			version: "v3.3.1",
			otlpConfig: &monitoringv1.OTLPConfig{
				ConvertHistogramsToNHCB: ptr.To(true),
			},
			golden: "OTLPConfig_Config_convert_histograms_to_nhcb_with_old_version.golden",
		},
		{
			name:    "Config IgnoreResourceAttributes and PromoteAllResourceAttributes true",
			version: "v3.5.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				IgnoreResourceAttributes:     []string{"aa", "bb", "cc"},
				PromoteAllResourceAttributes: ptr.To(true),
			},
			golden: "OTLPConfig_Config_ignore_resource_attributes_and_promote_all_resource_attributes.golden",
		},
		{
			name:    "Config IgnoreResourceAttributes with old prometheus version",
			version: "v3.4.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				IgnoreResourceAttributes: []string{"aa", "bb", "cc"},
			},
			golden: "OTLPConfig_Config_ignore_resource_attributes_wrong_prom.golden",
		},
		{
			name:    "Config IgnoreResourceAttributes with correct prometheus version but missing PromoteAllResourceAttributes ",
			version: "v3.5.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				IgnoreResourceAttributes: []string{"aa", "bb", "cc"},
			},
			expectedErr: true,
		},
		{
			name:    "Config PromoteAllResourceAttributes with correct prometheus version",
			version: "v3.5.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteAllResourceAttributes: ptr.To(true),
			},
			golden: "OTLPConfig_Config_promote_all_resource_attributes.golden",
		},
		{
			name:    "Config PromoteAllResourceAttributes and PromoteResourceAttributes",
			version: "v3.5.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes:    []string{"aa", "bb", "cc"},
				PromoteAllResourceAttributes: ptr.To(true),
			},
			expectedErr: true,
		},
		{
			name:    "Config PromoteAllResourceAttributes with old prometheus version",
			version: "v3.4.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteAllResourceAttributes: ptr.To(true),
			},
			golden: "OTLPConfig_Config_promote_all_resource_attributes_wrong_prom.golden",
		},
		{
			name:    "Config PromoteScopeMetadata with compatible versiopn",
			version: "v3.6.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteScopeMetadata: ptr.To(true),
			},
			golden: "OTLPConfig_Config_promote_scope_metadata.golden",
		},
		{
			name:    "Config PromoteScopeMetadata with wrong version",
			version: "v3.5.0",
			otlpConfig: &monitoringv1.OTLPConfig{
				PromoteScopeMetadata: ptr.To(true),
			},
			golden: "OTLPConfig_Config_promote_scope_metadata_wrong_version.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder()
			p := defaultPrometheus()

			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}

			p.Spec.CommonPrometheusFields.OTLP = tc.otlpConfig

			p.Spec.CommonPrometheusFields.NameValidationScheme = tc.nameValScheme

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				store,
				nil,
				nil,
				nil,
				nil,
			)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				golden.Assert(t, string(cfg), tc.golden)
			}
		})
	}
}

func TestTracingConfig(t *testing.T) {
	samplingTwo := resource.MustParse("0.5")
	testCases := []struct {
		tracingConfig *monitoringv1.TracingConfig
		name          string
		expectedErr   bool
		golden        string
	}{
		{
			name: "Config only with endpoint",
			tracingConfig: &monitoringv1.TracingConfig{
				Endpoint: "https://otel-collector.default.svc.local:3333",
			},
			golden:      "TracingConfig_Config_only_with_endpoint.golden",
			expectedErr: false,
		},
		{
			tracingConfig: &monitoringv1.TracingConfig{
				ClientType:       ptr.To("grpc"),
				Endpoint:         "https://otel-collector.default.svc.local:3333",
				SamplingFraction: &samplingTwo,
				Headers: map[string]string{
					"custom": "header",
				},
				Compression: ptr.To("gzip"),
				Timeout:     ptr.To(monitoringv1.Duration("10s")),
				Insecure:    ptr.To(false),
			},
			name:        "Expect valid config",
			expectedErr: false,
			golden:      "TracingConfig_Expect_valid_config.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder()
			p := defaultPrometheus()

			p.Spec.CommonPrometheusFields.TracingConfig = tc.tracingConfig

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				store,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithKumaSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "kuma_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KumaSDConfigs: []monitoringv1alpha1.KumaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Server:          "http://127.0.0.1:5681",
						ClientID:        ptr.To("client"),
						FetchTimeout:    (*monitoringv1.Duration)(ptr.To("5s")),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_KumaSD.golden",
		},
		{
			name: "kuma_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KumaSDConfigs: []monitoringv1alpha1.KumaSDConfig{
					{
						OAuth2: &monitoringv1.OAuth2{
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
						Server: "http://localhost:5681",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_KumaSD_with_OAuth.golden",
		}, {
			name: "kuma_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KumaSDConfigs: []monitoringv1alpha1.KumaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "key",
							},
						},
						Server: "http://localhost:5681",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_KumaSD_with_TLSConfig.golden",
		}, {
			name: "kuma_sd_config_tls_tlsversion",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KumaSDConfigs: []monitoringv1alpha1.KumaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "key",
							},
							MaxVersion: ptr.To(monitoringv1.TLSVersion12),
							MinVersion: ptr.To(monitoringv1.TLSVersion10),
						},
						Server: "http://localhost:5681",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_KumaSD_with_TLSConfig_TLSVersion.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func defaultServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaultServiceMonitor",
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
}

func defaultProbe() *monitoringv1.Probe {
	return &monitoringv1.Probe{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaultProbe",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ProbeSpec{
			ProberSpec: monitoringv1.ProberSpec{
				Scheme: ptr.To(monitoringv1.SchemeHTTP),
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
			MetricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Regex:  "noisy_labels.*",
					Action: "labeldrop",
				},
			},
		},
	}
}

func defaultPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaultPodMonitor",
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
					Port:     ptr.To("web"),
					Interval: "30s",
				},
			},
		},
	}
}

func defaultScrapeConfig() *monitoringv1alpha1.ScrapeConfig {
	return &monitoringv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaultScrapeConfig",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{
					URL:             "http://localhost:9100/sd.json",
					RefreshInterval: ptr.To(monitoringv1.Duration("5m")),
					ProxyConfig: monitoringv1.ProxyConfig{
						ProxyURL:             ptr.To("http://no-proxy.com"),
						NoProxy:              ptr.To("0.0.0.0"),
						ProxyFromEnvironment: ptr.To(false),
					},
				},
			},
		},
	}
}

func TestScrapeClass(t *testing.T) {
	testCases := []struct {
		name        string
		scrapeClass []monitoringv1.ScrapeClass
		golden      string
	}{
		{
			name:   "Monitor Object without Scrape Class",
			golden: "monitorObjectWithoutScrapeClass.golden",
		},
		{
			name:   "Monitor object with Non Default Scrape Class",
			golden: "monitorObjectWithNonDefaultScrapeClassAndTLSConfig.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name: "test-tls-scrape-class",
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/ca.crt",
							CertFile: "/etc/prometheus/secrets/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/tls.key",
						},
					},
				},
			},
		},
		{
			name:   "Monitor object with Default Scrape Class",
			golden: "monitorObjectWithDefaultScrapeClassAndTLSConfig.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name:    "test-tls-scrape-class",
					Default: ptr.To(true),
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/default/ca.crt",
							CertFile: "/etc/prometheus/secrets/default/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/default/tls.key",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			serviceMonitor := defaultServiceMonitor()
			podMonitor := defaultPodMonitor()
			probe := defaultProbe()
			scrapeConfig := defaultScrapeConfig()

			for _, sc := range tc.scrapeClass {
				p.Spec.ScrapeClasses = append(p.Spec.ScrapeClasses, sc)
				if !ptr.Deref(sc.Default, false) {
					serviceMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
					podMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
					probe.Spec.ScrapeClassName = ptr.To(sc.Name)
					scrapeConfig.Spec.ScrapeClassName = ptr.To(sc.Name)
				}
			}

			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitor},
				map[string]*monitoringv1.PodMonitor{"monitor": podMonitor},
				map[string]*monitoringv1.Probe{"monitor": probe},
				map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": scrapeConfig},
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestServiceMonitorScrapeClassWithDefaultTLS(t *testing.T) {
	testCases := []struct {
		name        string
		scrapeClass []monitoringv1.ScrapeClass
		tlsConfig   *monitoringv1.TLSConfig
		golden      string
	}{
		{
			name:   "Monitor object with Non Default Scrape Class and existing TLS Config",
			golden: "serviceMonitorObjectWithNonDefaultScrapeClassAndExistingTLSConfig.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name: "test-tls-scrape-class",
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/ca.crt",
							CertFile: "/etc/prometheus/secrets/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/tls.key",
						},
					},
				},
			},
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "tls",
							},
							Key: "cert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls",
						},
						Key: "private-key",
					},
				},
			},
		},
		{
			name:   "Monitor object with Non Default Scrape Class and existing TLS config missing ca",
			golden: "serviceMonitorObjectWithNonDefaultScrapeClassAndExistingTLSConfigMissingCA.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name: "test-tls-scrape-class",
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/ca.crt",
							CertFile: "/etc/prometheus/secrets/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/tls.key",
						},
					},
				},
			},
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
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
	}

	for _, tc := range testCases {
		p := defaultPrometheus()
		serviceMonitor := defaultServiceMonitor()

		for _, sc := range tc.scrapeClass {
			p.Spec.ScrapeClasses = append(p.Spec.ScrapeClasses, sc)
			if sc.Default == nil {
				serviceMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
			}
		}

		serviceMonitor.Spec.Endpoints[0].TLSConfig = tc.tlsConfig

		cg := mustNewConfigGenerator(t, p)

		cfg, err := cg.GenerateServerConfiguration(
			p,
			map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitor},
			nil,
			nil,
			nil,
			&assets.StoreBuilder{},
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		golden.Assert(t, string(cfg), tc.golden)
	}
}

func TestPodMonitorScrapeClassWithDefaultTLS(t *testing.T) {
	testCases := []struct {
		name        string
		scrapeClass []monitoringv1.ScrapeClass
		tlsConfig   *monitoringv1.SafeTLSConfig
		golden      string
	}{
		{
			name:   "Monitor object with Non Default Scrape Class and existing TLS Config",
			golden: "podMonitorObjectWithNonDefaultScrapeClassAndExistingTLSConfig.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name: "test-tls-scrape-class",
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/ca.crt",
							CertFile: "/etc/prometheus/secrets/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/tls.key",
						},
					},
				},
			},
			tlsConfig: &monitoringv1.SafeTLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls",
						},
						Key: "ca",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls",
						},
						Key: "cert",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "tls",
					},
					Key: "private-key",
				},
			},
		},
		{
			name:   "Monitor object with Non Default Scrape Class and existing TLS config missing ca",
			golden: "podMonitorObjectWithNonDefaultScrapeClassAndExistingTLSConfigMissingCA.golden",
			scrapeClass: []monitoringv1.ScrapeClass{
				{
					Name: "test-tls-scrape-class",
					TLSConfig: &monitoringv1.TLSConfig{
						TLSFilesConfig: monitoringv1.TLSFilesConfig{
							CAFile:   "/etc/prometheus/secrets/ca.crt",
							CertFile: "/etc/prometheus/secrets/tls.crt",
							KeyFile:  "/etc/prometheus/secrets/tls.key",
						},
					},
				},
			},
			tlsConfig: &monitoringv1.SafeTLSConfig{
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
	}

	for _, tc := range testCases {
		p := defaultPrometheus()
		podMonitor := defaultPodMonitor()

		for _, sc := range tc.scrapeClass {
			p.Spec.ScrapeClasses = append(p.Spec.ScrapeClasses, sc)
			if sc.Default == nil {
				podMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
			}
		}
		podMonitor.Spec.PodMetricsEndpoints[0].TLSConfig = tc.tlsConfig

		cg := mustNewConfigGenerator(t, p)

		cfg, err := cg.GenerateServerConfiguration(
			p,
			nil,
			map[string]*monitoringv1.PodMonitor{"monitor": podMonitor},
			nil,
			nil,
			&assets.StoreBuilder{},
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		golden.Assert(t, string(cfg), tc.golden)
	}
}

func TestPodMonitorPortNumber(t *testing.T) {
	testCases := []struct {
		name       string
		port       string
		portNumber int32
		targetPort intstr.IntOrString
		golden     string
	}{
		{
			name:       "PodMonitor with Pod Name",
			golden:     "podMonitorObjectWithPodName.golden",
			port:       "podname",
			portNumber: 1024,
			targetPort: intstr.FromString("10240"),
		},
		{
			name:       "PodMonitor with Pod Port Number",
			golden:     "podMonitorObjectWithPortNumber.golden",
			portNumber: 1024,
			targetPort: intstr.FromString("10240"),
		},
		{
			name:       "PodMonitor with TargetPort Int",
			golden:     "podMonitorObjectWithTargetPortInt.golden",
			targetPort: intstr.FromInt(10240),
		},
		{
			name:       "PodMonitor with TargetPort string",
			golden:     "podMonitorObjectWithTargetPortString.golden",
			targetPort: intstr.FromString("10240"),
		},
	}

	for _, tc := range testCases {
		p := defaultPrometheus()
		podMonitor := defaultPodMonitor()

		podMonitor.Spec.PodMetricsEndpoints[0].Port = ptr.To(tc.port)
		podMonitor.Spec.PodMetricsEndpoints[0].PortNumber = ptr.To(tc.portNumber)
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		podMonitor.Spec.PodMetricsEndpoints[0].TargetPort = ptr.To(tc.targetPort)

		cg := mustNewConfigGenerator(t, p)

		cfg, err := cg.GenerateServerConfiguration(
			p,
			nil,
			map[string]*monitoringv1.PodMonitor{"monitor": podMonitor},
			nil,
			nil,
			&assets.StoreBuilder{},
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		golden.Assert(t, string(cfg), tc.golden)
	}
}

func TestNewConfigGeneratorWithMultipleDefaultScrapeClass(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	p := defaultPrometheus()
	p.Spec.ScrapeClasses = []monitoringv1.ScrapeClass{
		{
			Name:    "test-default-scrape-class",
			Default: ptr.To(true),
			TLSConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "/etc/prometheus/secrets/ca.crt",
					CertFile: "/etc/prometheus/secrets/tls.crt",
					KeyFile:  "/etc/prometheus/secrets/tls.key",
				},
			},
		},
		{
			Name:    "test-default-scrape-class-2",
			Default: ptr.To(true),
			TLSConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "/etc/prometheus/secrets/ca.crt",
					CertFile: "/etc/prometheus/secrets/tls.crt",
					KeyFile:  "/etc/prometheus/secrets/tls.key",
				},
			},
		},
	}
	_, err := NewConfigGenerator(logger.With("test", "NewConfigGeneratorWithMultipleDefaultScrapeClass"), p)
	require.Error(t, err)
	require.Equal(t, "failed to parse scrape classes: multiple default scrape classes defined", err.Error())
}

func TestMergeTLSConfigWithScrapeClass(t *testing.T) {
	tests := []struct {
		name        string
		tlsConfig   *monitoringv1.TLSConfig
		scrapeClass monitoringv1.ScrapeClass

		expectedConfig *monitoringv1.TLSConfig
	}{
		{
			name: "nil TLSConfig and ScrapeClass",
			scrapeClass: monitoringv1.ScrapeClass{
				TLSConfig: &monitoringv1.TLSConfig{
					TLSFilesConfig: monitoringv1.TLSFilesConfig{
						CAFile:   "defaultCAFile",
						CertFile: "defaultCertFile",
						KeyFile:  "defaultKeyFile",
					},
				},
			},

			expectedConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "defaultCAFile",
					CertFile: "defaultCertFile",
					KeyFile:  "defaultKeyFile",
				},
			},
		},
		{
			name: "nil TLSConfig and empty ScrapeClass",
		},
		{
			name: "non-nil TLSConfig and empty ScrapeClass",
			tlsConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "caFile",
					CertFile: "certFile",
					KeyFile:  "keyFile",
				},
			},
			scrapeClass: monitoringv1.ScrapeClass{},

			expectedConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "caFile",
					CertFile: "certFile",
					KeyFile:  "keyFile",
				},
			},
		},
		{
			name: "non-nil TLSConfig and ScrapeClass",
			tlsConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "caFile",
					CertFile: "certFile",
					KeyFile:  "keyFile",
				},
			},
			scrapeClass: monitoringv1.ScrapeClass{
				TLSConfig: &monitoringv1.TLSConfig{
					TLSFilesConfig: monitoringv1.TLSFilesConfig{
						CAFile:   "defaultCAFile",
						CertFile: "defaultCertFile",
						KeyFile:  "defaultKeyFile",
					},
				},
			},

			expectedConfig: &monitoringv1.TLSConfig{
				TLSFilesConfig: monitoringv1.TLSFilesConfig{
					CAFile:   "caFile",
					CertFile: "certFile",
					KeyFile:  "keyFile",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeTLSConfigWithScrapeClass(tt.tlsConfig, tt.scrapeClass)
			require.Equal(t, tt.expectedConfig, result, "expected %v, got %v", tt.expectedConfig, result)
		})
	}
}

func TestScrapeConfigSpecConfigWithEurekaSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "eureka_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EurekaSDConfigs: []monitoringv1alpha1.EurekaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Server:          "http://localhost:8761/eureka",
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EurekaSD.golden",
		},
		{
			name: "eureka_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EurekaSDConfigs: []monitoringv1alpha1.EurekaSDConfig{
					{
						OAuth2: &monitoringv1.OAuth2{
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
						Server: "http://localhost:8761/eureka",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EurekaSD_with_OAuth.golden",
		}, {
			name: "eureka_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				EurekaSDConfigs: []monitoringv1alpha1.EurekaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
						Server: "http://localhost:8761/eureka",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_EurekaSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithNomadSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "nomad_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NomadSDConfigs: []monitoringv1alpha1.NomadSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						AllowStale:      ptr.To(true),
						TagSeparator:    ptr.To(","),
						Namespace:       ptr.To("default"),
						Region:          ptr.To("default"),
						Server:          "127.0.0.1",
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NomadSD.golden",
		},
		{
			name: "nomad_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NomadSDConfigs: []monitoringv1alpha1.NomadSDConfig{
					{
						OAuth2: &monitoringv1.OAuth2{
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
						Server: "http://localhost:4646",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NomadSD_with_OAuth.golden",
		}, {
			name: "nomad_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				NomadSDConfigs: []monitoringv1alpha1.NomadSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "key",
							},
						},
						Server: "http://localhost:4646",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NomadSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithDockerswarmSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "dockerswarm_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSwarmSDConfigs: []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Host: "https://www.example.com",
						Role: "Nodes",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
						Filters: []monitoringv1alpha1.Filter{
							{
								Name:   "foo",
								Values: []string{"bar"},
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerswarmSD.golden",
		},

		{
			name: "dockerswarm_sd_config_basicauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSwarmSDConfigs: []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Host: "http://www.example.com",
						Role: "Tasks",
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerswarmSD_withBasicAuth.golden",
		},
		{
			name: "dockerswarm_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSwarmSDConfigs: []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Host: "ftp://www.example.com",
						Role: "Nodes",
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_DockerswarmSD_with_OAuth.golden",
		}, {
			name: "dockerswarm_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				DockerSwarmSDConfigs: []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Host: "acp://www.example.com",
						Role: "Nodes",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DockerswarmSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithPuppetDBSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "puppetdb_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				PuppetDBSDConfigs: []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL:   "https://www.example.com",
						Query: "vhost", // This is not a valid PuppetDB query, just a mock.
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_PuppetDBSD.golden",
		},

		{
			name: "puppetdb_sd_config_basicauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				PuppetDBSDConfigs: []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL:   "http://www.example.com",
						Query: "vhost", // This is not a valid PuppetDB query, just a mock.
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_PuppetDBSD_withBasicAuth.golden",
		},
		{
			name: "puppetdb_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				PuppetDBSDConfigs: []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL:   "https://www.example.com",
						Query: "vhost", // This is not a valid PuppetDB query, just a mock.
						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_PuppetDBSD_with_OAuth.golden",
		}, {
			name: "puppetdb_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				PuppetDBSDConfigs: []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL:   "https://www.example.com",
						Query: "vhost", // This is not a valid PuppetDB query, just a mock.
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_PuppetDBSD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}
func TestScrapeConfigSpecConfigWithLightSailSD(t *testing.T) {
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
		expectedErr bool
	}{
		{
			name: "lightSail_sd_config_valid_with_api_keys",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:   ptr.To("us-east-1"),
						Endpoint: ptr.To("https://lightsail.us-east-1.amazonaws.com/"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "aws-access-api",
							},
							Key: "accessKey",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "aws-access-api",
							},
							Key: "secretKey",
						},
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LightSailSDConfigValidAPIKeys.golden",
		},
		{
			name: "lightSail_sd_config_valid_with_role_arn",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:          ptr.To("us-east-1"),
						Endpoint:        ptr.To("https://lightsail.us-east-1.amazonaws.com/"),
						RoleARN:         ptr.To("arn:aws:iam::123456789:role/prometheus-role"),
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LightSailSDConfigValidRoleARN.golden",
		},
		{
			name: "lightSail_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:   ptr.To("us-east-1"),
						Endpoint: ptr.To("https://lightsail.us-east-1.amazonaws.com/"),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LightSailSD.golden",
		},
		{
			name: "lightSail_sd_config_basicauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:   ptr.To("us-east-1"),
						Endpoint: ptr.To("https://lightsail.us-east-1.amazonaws.com/"),
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LightSailSD_withBasicAuth.golden",
		},
		{
			name: "lightSail_sd_config_oauth",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:   ptr.To("us-east-1"),
						Endpoint: ptr.To("https://lightsail.us-east-1.amazonaws.com/"),

						OAuth2: &monitoringv1.OAuth2{
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
			golden: "ScrapeConfigSpecConfig_LightSailSD_with_OAuth.golden",
		},
		{
			name: "lightSail_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
					{
						Region:   ptr.To("us-east-1"),
						Endpoint: ptr.To("https://lightsail.us-east-1.amazonaws.com/"),
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_LightSailSD_with_TLSConfig.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-access-api",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"accessKey": []byte("access-key"),
						"secretKey": []byte("secret-key"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithOVHCloudSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "ovhcloud_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
					{
						ApplicationKey: "application-key",
						ApplicationSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "as",
						},
						ConsumerKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "ck",
						},
						Service:         monitoringv1alpha1.OVHServiceDedicatedServer,
						Endpoint:        ptr.To("127.0.0.1"),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_OVHCloudSD.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"as": []byte("application-secret"),
						"ck": []byte("consumer-key"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithScalewaySD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "scaleway_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "SCWABCDEFGH123456789",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "credential",
						},
						ProjectID:  "00000000-0000-0000-0000-000000000001",
						Role:       monitoringv1alpha1.ScalewayRoleInstance,
						Zone:       ptr.To("fr-par-1"),
						Port:       ptr.To(int32(23456)),
						ApiURL:     ptr.To("https://api.scaleway.com"),
						NameFilter: ptr.To("name"),
						TagsFilter: []string{"aa", "bb"},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_ScaleWaySD.golden",
		}, {
			name: "scaleway_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "SCWABCDEFGH123456789",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "credential",
						},
						ProjectID: "00000000-0000-0000-0000-000000000001",
						Role:      monitoringv1alpha1.ScalewayRoleInstance,
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_ScaleWaySD_with_TLSConfig.golden",
		}} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestScrapeConfigSpecConfigWithIonosSD(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scSpec monitoringv1alpha1.ScrapeConfigSpec
		golden string
	}{
		{
			name: "ionos_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
					{
						DataCenterID: "11111111-1111-1111-1111-111111111111",
						Authorization: monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
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
										Key: "proxy-header",
									},
								},
							},
						},
						FollowRedirects: ptr.To(true),
						EnableHTTP2:     ptr.To(true),
						Port:            ptr.To(int32(9100)),
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_IonosSD.golden",
		},
		{
			name: "ionos_sd_config_tls",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
					{
						DataCenterID: "11111111-1111-1111-1111-111111111111",
						Authorization: monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "credential",
							},
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "tls",
								},
								Key: "private-key",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_IonosSD_withTLSConfig.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewTestStoreBuilder(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"proxy-header": []byte("value"),
						"token":        []byte("value"),
						"credential":   []byte("00000000-0000-0000-0000-000000000000"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string]string{
						"client_id": "client-id",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "oauth2",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"client_secret": []byte("client-secret"),
					},
				},
			)

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
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestServiceMonitorWithDefaultScrapeClassRelabelings(t *testing.T) {
	p := defaultPrometheus()
	serviceMonitor := defaultServiceMonitor()
	scrapeClasses := []monitoringv1.ScrapeClass{
		{
			Name:    "default",
			Default: ptr.To(true),
			Relabelings: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_app_name"},
					TargetLabel:  "app",
				},
			},
		},
		{
			Name: "not-default",
			Relabelings: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
					TargetLabel:  "node",
				},
			},
		},
	}

	p.Spec.ScrapeClasses = scrapeClasses
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p,
		map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitor},
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "serviceMonitorObjectWithDefaultScrapeClassWithRelabelings.golden")
}

func TestServiceMonitorWithNonDefaultScrapeClassRelabelings(t *testing.T) {
	p := defaultPrometheus()
	serviceMonitor := defaultServiceMonitor()
	sc := monitoringv1.ScrapeClass{
		Name: "test-extra-relabelings-scrape-class",
		Relabelings: []monitoringv1.RelabelConfig{
			{
				Action:       "replace",
				SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
				TargetLabel:  "node",
			},
		},
	}

	p.Spec.ScrapeClasses = append(p.Spec.ScrapeClasses, sc)
	serviceMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p,
		map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitor},
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "serviceMonitorObjectWithNonDefaultScrapeClassWithRelabelings.golden")
}

func TestPodMonitorWithDefaultScrapeClassRelabelings(t *testing.T) {
	p := defaultPrometheus()
	podMonitor := defaultPodMonitor()
	scrapeClasses := []monitoringv1.ScrapeClass{
		{
			Name:    "default",
			Default: ptr.To(true),
			Relabelings: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_app_name"},
					TargetLabel:  "app",
				},
			},
		},
		{
			Name: "not-default",
			Relabelings: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
					TargetLabel:  "node",
				},
			},
		},
	}

	p.Spec.ScrapeClasses = scrapeClasses
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		map[string]*monitoringv1.PodMonitor{"monitor": podMonitor},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "podMonitorObjectWithDefaultScrapeClassWithRelabelings.golden")
}

func TestPodMonitorWithNonDefaultScrapeClassRelabelings(t *testing.T) {
	p := defaultPrometheus()
	podMonitor := defaultPodMonitor()
	sc := monitoringv1.ScrapeClass{
		Name: "test-extra-relabelings-scrape-class",
		Relabelings: []monitoringv1.RelabelConfig{
			{
				Action:       "replace",
				SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_node_name"},
				TargetLabel:  "node",
			},
		},
	}

	p.Spec.ScrapeClasses = append(p.Spec.ScrapeClasses, sc)
	podMonitor.Spec.ScrapeClassName = ptr.To(sc.Name)
	cg := mustNewConfigGenerator(t, p)

	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		map[string]*monitoringv1.PodMonitor{"monitor": podMonitor},
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	golden.Assert(t, string(cfg), "podMonitorObjectWithNonDefaultScrapeClassWithRelabelings.golden")
}

func TestScrapeClassMetricRelabelings(t *testing.T) {
	serviceMonitorWithNonDefaultScrapeClass := defaultServiceMonitor()
	serviceMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-extra-relabelings-scrape-class")
	podMonitorWithNonDefaultScrapeClass := defaultPodMonitor()
	podMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-extra-relabelings-scrape-class")
	for _, tc := range []struct {
		name            string
		scrapeClasses   []monitoringv1.ScrapeClass
		serviceMonitors map[string]*monitoringv1.ServiceMonitor
		podMonitors     map[string]*monitoringv1.PodMonitor
		probes          map[string]*monitoringv1.Probe
		scrapeConfigs   map[string]*monitoringv1alpha1.ScrapeConfig
		goldenFile      string
	}{
		{
			name: "ServiceMonitor with default ScrapeClass MetricRelabelings",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							SourceLabels: []monitoringv1.LabelName{"namespace"},
							Regex:        "tenant1-.*",
							TargetLabel:  "tenant",
							Replacement:  ptr.To("tenant1"),
						},
						{
							SourceLabels: []monitoringv1.LabelName{"namespace"},
							Regex:        "tenant2-.*",
							TargetLabel:  "tenant",
							Replacement:  ptr.To("tenant2"),
						},
					},
				},
				{
					Name: "not-default",
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							TargetLabel: "tenant",
							Replacement: ptr.To("not-default"),
						},
					},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": defaultServiceMonitor()},
			goldenFile:      "serviceMonitorObjectWithDefaultScrapeClassWithMetricRelabelings.golden",
		},
		{
			name: "ServiceMonitor with non-default ScrapeClass MetricRelabelings",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: "test-extra-relabelings-scrape-class",
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							TargetLabel: "extra",
							Replacement: ptr.To("value1"),
						},
					},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitorWithNonDefaultScrapeClass},
			goldenFile:      "serviceMonitorObjectWithNonDefaultScrapeClassWithMetricRelabelings.golden",
		},
		{
			name: "PodMonitor with default ScrapeClass MetricRelabelings",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							SourceLabels: []monitoringv1.LabelName{"namespace"},
							Regex:        "tenant1-.*",
							TargetLabel:  "tenant",
							Replacement:  ptr.To("tenant1"),
						},
						{
							SourceLabels: []monitoringv1.LabelName{"namespace"},
							Regex:        "tenant2-.*",
							TargetLabel:  "tenant",
							Replacement:  ptr.To("tenant2"),
						},
					},
				},
				{
					Name: "not-default",
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							TargetLabel: "tenant",
							Replacement: ptr.To("not-default"),
						},
					},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": defaultPodMonitor()},
			goldenFile:  "podMonitorObjectWithDefaultScrapeClassWithMetricRelabelings.golden",
		},
		{
			name: "PodMonitor with non-default ScrapeClass MetricRelabelings",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: "test-extra-relabelings-scrape-class",
					MetricRelabelings: []monitoringv1.RelabelConfig{
						{
							TargetLabel: "extra",
							Replacement: ptr.To("value1"),
						},
					},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": podMonitorWithNonDefaultScrapeClass},
			goldenFile:  "podMonitorObjectWithNonDefaultScrapeClassWithMetricRelabelings.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

			p.Spec.ScrapeClasses = tc.scrapeClasses
			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				tc.serviceMonitors,
				tc.podMonitors,
				tc.probes,
				tc.scrapeConfigs,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.goldenFile)
		})
	}
}

func TestScrapeClassAuthorization(t *testing.T) {
	const (
		scrapeClassName = "test-attach-authz-scrape-class"

		secretName      = "secret"
		secretNamespace = "default"
		secretKey       = "token"
		secretValue     = "token"
		authType        = "Bearer"
	)

	scn := ptr.To(scrapeClassName)

	safeAuthz := &monitoringv1.SafeAuthorization{
		Type: authType,
		Credentials: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{Name: secretName},
			Key:                  secretKey,
		},
	}

	authzSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			secretKey: []byte(secretValue),
		},
	}

	serviceMonitorWithNonDefaultScrapeClass := defaultServiceMonitor()
	serviceMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = scn
	serviceMonitorWithUserConfiguredAuthz := defaultServiceMonitor()
	serviceMonitorWithUserConfiguredAuthz.Spec.Endpoints[0].Authorization = safeAuthz

	podMonitorWithNonDefaultScrapeClass := defaultPodMonitor()
	podMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = scn
	podMonitorWithUserConfiguredAuthz := defaultPodMonitor()
	podMonitorWithUserConfiguredAuthz.Spec.PodMetricsEndpoints[0].Authorization = safeAuthz

	probeWithNonDefaultScrapeClass := defaultProbe()
	probeWithNonDefaultScrapeClass.Spec.ScrapeClassName = scn
	probeWithUserConfiguredAuthz := defaultProbe()
	probeWithUserConfiguredAuthz.Spec.Authorization = safeAuthz

	scrapeConfigWithNonDefaultScrapeClass := defaultScrapeConfig()
	scrapeConfigWithNonDefaultScrapeClass.Spec.ScrapeClassName = scn
	scrapeConfigWithUserConfiguredAuthz := defaultScrapeConfig()
	scrapeConfigWithUserConfiguredAuthz.Spec.Authorization = safeAuthz

	for _, tc := range []struct {
		name            string
		scrapeClasses   []monitoringv1.ScrapeClass
		serviceMonitors map[string]*monitoringv1.ServiceMonitor
		podMonitors     map[string]*monitoringv1.PodMonitor
		probes          map[string]*monitoringv1.Probe
		scrapeConfigs   map[string]*monitoringv1alpha1.ScrapeConfig
		storeBuilder    *assets.StoreBuilder
		goldenFile      string
	}{
		{
			name: "ServiceMonitor with default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": defaultServiceMonitor()},
			goldenFile:      "serviceMonitorObjectWithDefaultScrapeClassAuthz.golden",
		},
		{
			name: "ServiceMonitor with non-default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitorWithNonDefaultScrapeClass},
			goldenFile:      "serviceMonitorObjectWithNonDefaultScrapeClassAuthz.golden",
		},
		{
			name: "ServiceMonitor with user defined Authorization not overridden by ScrapeClass",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			storeBuilder:    assets.NewTestStoreBuilder(authzSecret),
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitorWithUserConfiguredAuthz},
			goldenFile:      "serviceMonitorObjectWithNonDefaultScrapeClassUserDefinedAuthz.golden",
		},
		{
			name: "PodMonitor with default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/default",
					},
				},
				{
					Name: "not-default",
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/non-default",
					},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": defaultPodMonitor()},
			goldenFile:  "podMonitorObjectWithDefaultScrapeClassAuthz.golden",
		},
		{
			name: "PodMonitor with non-default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": podMonitorWithNonDefaultScrapeClass},
			goldenFile:  "podMonitorObjectWithNonDefaultScrapeClassAuthz.golden",
		},
		{
			name: "PodMonitor with user defined Authorization not overridden by ScrapeClass",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			storeBuilder: assets.NewTestStoreBuilder(authzSecret),
			podMonitors:  map[string]*monitoringv1.PodMonitor{"monitor": podMonitorWithUserConfiguredAuthz},
			goldenFile:   "podMonitorObjectWithNonDefaultScrapeClassUserDefinedAuthz.golden",
		},
		{
			name: "Probe with default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/default",
					},
				},
				{
					Name: "not-default",
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/non-default",
					},
				},
			},
			probes:     map[string]*monitoringv1.Probe{"probe": defaultProbe()},
			goldenFile: "ProbeWithDefaultScrapeClassAuthz.golden",
		},
		{
			name: "Probe with non-default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			probes:     map[string]*monitoringv1.Probe{"probe": probeWithNonDefaultScrapeClass},
			goldenFile: "ProbeWithNonDefaultScrapeClassAuthz.golden",
		},
		{
			name: "Probe with user defined Authorization not overridden by ScrapeClass",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			storeBuilder: assets.NewTestStoreBuilder(authzSecret),
			probes:       map[string]*monitoringv1.Probe{"probe": probeWithUserConfiguredAuthz},
			goldenFile:   "ProbeWithNonDefaultScrapeClassUserDefinedAuthz.golden",
		},
		{
			name: "ScrapeConfig with default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:    "default",
					Default: ptr.To(true),
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/default",
					},
				},
				{
					Name: "not-default",
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials/non-default",
					},
				},
			},
			scrapeConfigs: map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": defaultScrapeConfig()},
			goldenFile:    "ScrapeConfigWithDefaultScrapeClassAuthz.golden",
		},
		{
			name: "ScrapeConfig with non-default ScrapeClass attach Authorization",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			scrapeConfigs: map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": scrapeConfigWithNonDefaultScrapeClass},
			goldenFile:    "ScrapeConfigWithNonDefaultScrapeClassAuthz.golden",
		},
		{
			name: "ScrapeConfig with user defined Authorization not overridden by ScrapeClass",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name: scrapeClassName,
					Authorization: &monitoringv1.Authorization{
						CredentialsFile: "/etc/secret/credentials",
					},
				},
			},
			storeBuilder:  assets.NewTestStoreBuilder(authzSecret),
			scrapeConfigs: map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": scrapeConfigWithUserConfiguredAuthz},
			goldenFile:    "ScrapeConfigWithNonDefaultScrapeClassUserDefinedAuthz.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

			p.Spec.ScrapeClasses = tc.scrapeClasses
			cg := mustNewConfigGenerator(t, p)

			if tc.storeBuilder == nil {
				tc.storeBuilder = &assets.StoreBuilder{}
			}

			cfg, err := cg.GenerateServerConfiguration(
				p,
				tc.serviceMonitors,
				tc.podMonitors,
				tc.probes,
				tc.scrapeConfigs,
				tc.storeBuilder,
				nil,
				nil,
				nil,
				nil,
			)

			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.goldenFile)
		})
	}
}

func TestScrapeClassAttachMetadata(t *testing.T) {
	serviceMonitorWithNonDefaultScrapeClass := defaultServiceMonitor()
	serviceMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-attachmetadata-scrape-class")
	podMonitorWithNonDefaultScrapeClass := defaultPodMonitor()
	podMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-attachmetadata-scrape-class")
	for _, tc := range []struct {
		name            string
		scrapeClasses   []monitoringv1.ScrapeClass
		serviceMonitors map[string]*monitoringv1.ServiceMonitor
		podMonitors     map[string]*monitoringv1.PodMonitor
		probes          map[string]*monitoringv1.Probe
		scrapeConfigs   map[string]*monitoringv1alpha1.ScrapeConfig
		goldenFile      string
	}{
		{
			name: "ServiceMonitor with default ScrapeClass AttachMetadata",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:           "default",
					Default:        ptr.To(true),
					AttachMetadata: &monitoringv1.AttachMetadata{Node: ptr.To(true)},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": defaultServiceMonitor()},
			goldenFile:      "serviceMonitorObjectWithDefaultScrapeClassWithAttachMetadata.golden",
		},
		{
			name: "ServiceMonitor with non-default ScrapeClass AttachMetadata",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:           "test-attachmetadata-scrape-class",
					AttachMetadata: &monitoringv1.AttachMetadata{Node: ptr.To(true)},
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitorWithNonDefaultScrapeClass},
			goldenFile:      "serviceMonitorObjectWithNonDefaultScrapeClassWithAttachMetadata.golden",
		},
		{
			name: "PodMonitor with default ScrapeClass AttachMetadata",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:           "default",
					Default:        ptr.To(true),
					AttachMetadata: &monitoringv1.AttachMetadata{Node: ptr.To(true)},
				},
				{
					Name:           "not-default",
					AttachMetadata: &monitoringv1.AttachMetadata{Node: ptr.To(true)},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": defaultPodMonitor()},
			goldenFile:  "podMonitorObjectWithDefaultScrapeClassWithAttachMetadata.golden",
		},
		{
			name: "PodMonitor with non-default ScrapeClass AttachMetadata",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:           "test-attachmetadata-scrape-class",
					AttachMetadata: &monitoringv1.AttachMetadata{Node: ptr.To(true)},
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": podMonitorWithNonDefaultScrapeClass},
			goldenFile:  "podMonitorObjectWithNonDefaultScrapeClassWithAttachMetadata.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

			p.Spec.ScrapeClasses = tc.scrapeClasses
			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				tc.serviceMonitors,
				tc.podMonitors,
				tc.probes,
				tc.scrapeConfigs,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)

			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.goldenFile)
		})
	}
}

func TestScrapeClassFallbackScrapeProtocol(t *testing.T) {
	serviceMonitorWithNonDefaultScrapeClass := defaultServiceMonitor()
	serviceMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-fallback-scrapeprotocol-scrape-class")
	podMonitorWithNonDefaultScrapeClass := defaultPodMonitor()
	podMonitorWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-fallback-scrapeprotocol-scrape-class")
	probeWithNonDefaultScrapeClass := defaultProbe()
	probeWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-fallback-scrapeprotocol-scrape-class")
	scrapeConfigWithNonDefaultScrapeClass := defaultScrapeConfig()
	scrapeConfigWithNonDefaultScrapeClass.Spec.ScrapeClassName = ptr.To("test-fallback-scrapeprotocol-scrape-class")
	for _, tc := range []struct {
		name            string
		scrapeClasses   []monitoringv1.ScrapeClass
		serviceMonitors map[string]*monitoringv1.ServiceMonitor
		podMonitors     map[string]*monitoringv1.PodMonitor
		probes          map[string]*monitoringv1.Probe
		scrapeConfigs   map[string]*monitoringv1alpha1.ScrapeConfig
		goldenFile      string
	}{
		{
			name: "ServiceMonitor with default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "default",
					Default:                ptr.To(true),
					FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": defaultServiceMonitor()},
			goldenFile:      "serviceMonitorObjectWithDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "ServiceMonitor with non-default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "test-fallback-scrapeprotocol-scrape-class",
					FallbackScrapeProtocol: ptr.To(monitoringv1.PrometheusText0_0_4),
				},
			},
			serviceMonitors: map[string]*monitoringv1.ServiceMonitor{"monitor": serviceMonitorWithNonDefaultScrapeClass},
			goldenFile:      "serviceMonitorObjectWithNonDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "PodMonitor with default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "default",
					Default:                ptr.To(true),
					FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": defaultPodMonitor()},
			goldenFile:  "podMonitorObjectWithDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "PodMonitor with non-default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "test-fallback-scrapeprotocol-scrape-class",
					FallbackScrapeProtocol: ptr.To(monitoringv1.PrometheusText0_0_4),
				},
			},
			podMonitors: map[string]*monitoringv1.PodMonitor{"monitor": podMonitorWithNonDefaultScrapeClass},
			goldenFile:  "podMonitorObjectWithNonDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "Probe with default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "default",
					Default:                ptr.To(true),
					FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
				},
			},
			probes:     map[string]*monitoringv1.Probe{"monitor": defaultProbe()},
			goldenFile: "probeObjectWithDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "Probe with non-default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "test-fallback-scrapeprotocol-scrape-class",
					FallbackScrapeProtocol: ptr.To(monitoringv1.PrometheusText0_0_4),
				},
			},
			probes:     map[string]*monitoringv1.Probe{"monitor": probeWithNonDefaultScrapeClass},
			goldenFile: "probeObjectWithNonDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "ScrapeConfig with default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "default",
					Default:                ptr.To(true),
					FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText1_0_0),
				},
			},
			scrapeConfigs: map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": defaultScrapeConfig()},
			goldenFile:    "scrapeConfigObjectWithDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
		{
			name: "ScrapeConfig with non-default ScrapeClass FallbackScrapeProtocol",
			scrapeClasses: []monitoringv1.ScrapeClass{
				{
					Name:                   "test-fallback-scrapeprotocol-scrape-class",
					FallbackScrapeProtocol: ptr.To(monitoringv1.PrometheusText0_0_4),
				},
			},
			scrapeConfigs: map[string]*monitoringv1alpha1.ScrapeConfig{"monitor": scrapeConfigWithNonDefaultScrapeClass},
			goldenFile:    "scrapeConfigObjectWithNonDefaultScrapeClassWithFallbackScrapeProtocol.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.EnforcedNamespaceLabel = "namespace"

			p.Spec.ScrapeClasses = tc.scrapeClasses
			cg := mustNewConfigGenerator(t, p)

			cfg, err := cg.GenerateServerConfiguration(
				p,
				tc.serviceMonitors,
				tc.podMonitors,
				tc.probes,
				tc.scrapeConfigs,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)

			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.goldenFile)
		})
	}
}

func TestGenerateAlertmanagerConfig(t *testing.T) {
	for _, tc := range []struct {
		alerting *monitoringv1.AlertingSpec
		sdRole   *monitoringv1.ServiceDiscoveryRole
		golden   string
	}{
		{
			alerting: nil,
			golden:   "AlertmanagerConfigEmpty.golden",
		},
		{
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						Port:      intstr.FromString("web"),
					},
				},
			},
			golden: "AlertmanagerConfigOtherNamespace.golden",
		},
		{
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("default"),
						Port:      intstr.FromString("web"),
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
										Key: "proxy-header",
									},
								},
							},
						},
					},
				},
			},
			golden: "AlertmanagerConfigProxyconfig.golden",
		},
		{
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("default"),
						Port:      intstr.FromString("web"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
							},
						},
					},
				},
			},
			golden: "AlertmanagerConfigTLSconfig.golden",
		},
		{
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						Port:      intstr.FromString("web"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
							},
						},
					},
				},
			},
			golden: "AlertmanagerConfigTLSconfigOtherNamespace.golden",
		},
		{
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("default"),
						Port:      intstr.FromString("web"),
					},
				},
			},
			sdRole: ptr.To(monitoringv1.EndpointSliceRole),
			golden: "AlertmanagerConfigEndpointSlice.golden",
		},
	} {
		t.Run("", func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.PrometheusSpec{
					Alerting:           tc.alerting,
					EvaluationInterval: "30s",
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ServiceDiscoveryRole: tc.sdRole,
						ScrapeInterval:       "30s",
					},
				},
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{},
				nil,
				nil,
				nil,
				assets.NewTestStoreBuilder(
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
				),
				nil,
				nil,
				nil,
				nil,
			)

			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAlertmanagerTLSConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		version  string
		alerting *monitoringv1.AlertingSpec
		golden   string
	}{
		{
			name:    "Valid Prom Version with TLSConfig",
			version: "2.26.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
								MaxVersion: ptr.To(monitoringv1.TLSVersion12),
								MinVersion: ptr.To(monitoringv1.TLSVersion10),
							},
						},
					},
				},
			},
			golden: "AlertmanagerTLSConfig_Valid_Prom_TLSConfig.golden",
		},
		{
			name:    "Invalid Prom Version with TLSConfig MinVersion",
			version: "2.36.0",
			alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
								MinVersion: ptr.To(monitoringv1.TLSVersion10),
							},
						},
					},
				},
			},
			golden: "AlertmanagerTLSConfig_Valid_Prom_TLSConfig_MinVersion.golden",
		},
		{
			name:    "Invalid Prom Version with TLSConfig MaxVersion",
			version: "2.41.0",
			alerting: &monitoringv1.AlertingSpec{

				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
								MaxVersion: ptr.To(monitoringv1.TLSVersion12),
							},
						},
					},
				},
			},
			golden: "AlertmanagerTLSConfig_Valid_Prom_TLSConfig_MaxVersion.golden",
		},
		{
			name:    "Invalid Prom Version with TLSConfig MaxVersion and MinVersion",
			version: "2.51.0",
			alerting: &monitoringv1.AlertingSpec{

				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "foo",
						Namespace: ptr.To("other"),
						TLSConfig: &monitoringv1.TLSConfig{
							SafeTLSConfig: monitoringv1.SafeTLSConfig{
								CA: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "ca",
									},
								},
								Cert: monitoringv1.SecretOrConfigMap{
									Secret: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "tls",
										},
										Key: "cert",
									},
								},
								KeySecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "tls",
									},
									Key: "private-key",
								},
								MaxVersion: ptr.To(monitoringv1.TLSVersion12),
								MinVersion: ptr.To(monitoringv1.TLSVersion10),
							},
						},
					},
				},
			},
			golden: "AlertmanagerTLSConfig_Valid_Prom_TLSConfig_MaxVersion_MinVersion.golden",
		},
	} {

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
				Alerting:           tc.alerting,
				EvaluationInterval: "30s",
			},
		}

		cg := mustNewConfigGenerator(t, p)
		cfg, err := cg.GenerateServerConfiguration(
			p,
			map[string]*monitoringv1.ServiceMonitor{},
			nil,
			nil,
			nil,
			assets.NewTestStoreBuilder(),
			nil,
			nil,
			nil,
			nil,
		)

		require.NoError(t, err)
		golden.Assert(t, string(cfg), tc.golden)

	}
}

func TestServiceMonitorSelectors(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		golden               string
		serviceMonitor       *monitoringv1.ServiceMonitor
		serviceDiscoveryRole monitoringv1.ServiceDiscoveryRole
	}{
		{
			name:                 "ServiceMonitor with Match Label Selector",
			golden:               "ServiceMonitorObjectWithMatchLabelSelector.golden",
			serviceDiscoveryRole: monitoringv1.EndpointsRole,
			serviceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultServiceMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
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
			},
		},
		{
			name:                 "ServiceMonitor with Match Expression Selector",
			golden:               "ServiceMonitorObjectWithMatchExpressionSelector.golden",
			serviceDiscoveryRole: monitoringv1.EndpointsRole,
			serviceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultServiceMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "group",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"group1"},
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
		{
			name:                 "ServiceMonitor with endpoint slice selector and match label selector",
			golden:               "ServiceMonitorObjectWithEndpointSliceSelectorAndMatchLabelSelector.golden",
			serviceDiscoveryRole: monitoringv1.EndpointSliceRole,
			serviceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultServiceMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
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
			},
		},
		{
			name:                 "ServiceMonitor with selector and match expression selector",
			golden:               "ServiceMonitorObjectWithSelectorAndMatchExpressionSelector.golden",
			serviceDiscoveryRole: monitoringv1.EndpointSliceRole,
			serviceMonitor: &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultServiceMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"group": "group1",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "group",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"group2"},
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			p.Spec.ServiceDiscoveryRole = &tc.serviceDiscoveryRole
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				map[string]*monitoringv1.ServiceMonitor{"monitor": tc.serviceMonitor},
				nil,
				nil,
				nil,
				assets.NewTestStoreBuilder(),
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestPodMonitorSelectors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		golden     string
		podMonitor *monitoringv1.PodMonitor
	}{
		{
			name:   "PodMonitor with Match Label Selector",
			golden: "PodMonitorObjectWithMatchLabelSelector.golden",
			podMonitor: &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultPodMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"group": "group1",
						},
					},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
		{
			name:   "PodMonitor with Match Expression Selector",
			golden: "PodMonitorObjectWithMatchExpressionSelector.golden",
			podMonitor: &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultPodMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "group",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"group1"},
							},
						},
					},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
		{
			name:   "PodMonitor with endpoint slice selector and match label selector",
			golden: "PodMonitorObjectWithEndpointSliceSelectorAndMatchLabelSelector.golden",
			podMonitor: &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultPodMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"group": "group1",
						},
					},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
		{
			name:   "PodMonitor with selector and match expression selector",
			golden: "PodMonitorObjectWithSelectorAndMatchExpressionSelector.golden",
			podMonitor: &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaultPodMonitor",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					SelectorMechanism: ptr.To(monitoringv1.SelectorMechanismRole),
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"group": "group1",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "group",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"group2"},
							},
							{
								Key:      "groupb",
								Operator: metav1.LabelSelectorOpDoesNotExist,
							},
						},
					},
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
						{
							Port:     ptr.To("web"),
							Interval: "30s",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()
			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				map[string]*monitoringv1.PodMonitor{"monitor": tc.podMonitor},
				nil,
				nil,
				assets.NewTestStoreBuilder(),
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestAppendConvertScrapeClassicHistograms(t *testing.T) {
	testCases := []struct {
		name                    string
		version                 string
		ScrapeClassicHistograms *bool
		expectedCfg             string
	}{
		{
			name:                    "ScrapeClassicHistograms true with Prometheus Version 3.5",
			version:                 "v3.5.0",
			ScrapeClassicHistograms: ptr.To(true),
			expectedCfg:             "ScrapeClassicHistogramsTrueProperPromVersion.golden",
		},
		{
			name:                    "ScrapeClassicHistograms false with Prometheus Version 3.5",
			version:                 "v3.5.0",
			ScrapeClassicHistograms: ptr.To(false),
			expectedCfg:             "ScrapeClassicHistogramsFalseProperPromVersion.golden",
		},
		{
			name:                    "ScrapeClassicHistograms true with Prometheus Version 2",
			version:                 "v2.45.0",
			ScrapeClassicHistograms: ptr.To(true),
			expectedCfg:             "ScrapeClassicHistogramsTrueWrongPromVersion.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.ScrapeClassicHistograms != nil {
				p.Spec.CommonPrometheusFields.ScrapeClassicHistograms = tc.ScrapeClassicHistograms
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)

			golden.Assert(t, string(cfg), tc.expectedCfg)
		})
	}
}

func TestAppendScrapeNativeHistograms(t *testing.T) {
	testCases := []struct {
		name                   string
		version                string
		ScrapeNativeHistograms *bool
		expectedCfg            string
	}{
		{
			name:                   "ScrapeNativeHistograms true with Prometheus Version 3.8",
			version:                "v3.8.0",
			ScrapeNativeHistograms: ptr.To(true),
			expectedCfg:            "ScrapeNativeHistogramsTrueProperPromVersion.golden",
		},
		{
			name:                   "ScrapeNativeHistograms false with Prometheus Version 3.8",
			version:                "v3.8.0",
			ScrapeNativeHistograms: ptr.To(false),
			expectedCfg:            "ScrapeNativeHistogramsFalseProperPromVersion.golden",
		},
		{
			name:                   "ScrapeNativeHistograms true with Lower Prometheus version",
			version:                "v3.7.0",
			ScrapeNativeHistograms: ptr.To(true),
			expectedCfg:            "ScrapeNativeHistogramsTrueWrongPromVersion.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			p := defaultPrometheus()
			if tc.version != "" {
				p.Spec.CommonPrometheusFields.Version = tc.version
			}
			if tc.ScrapeNativeHistograms != nil {
				p.Spec.CommonPrometheusFields.ScrapeNativeHistograms = tc.ScrapeNativeHistograms
			}

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)

			golden.Assert(t, string(cfg), tc.expectedCfg)
		})
	}
}
