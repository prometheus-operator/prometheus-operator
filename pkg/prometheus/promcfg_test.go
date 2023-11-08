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
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/golden"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
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
		expectedkeepDroppedTargets    uint64                = 50
	)

	for _, tc := range []struct {
		Scenario                    string
		EvaluationInterval          monitoringv1.Duration
		ScrapeInterval              monitoringv1.Duration
		ScrapeTimeout               monitoringv1.Duration
		ExternalLabels              map[string]string
		PrometheusExternalLabelName *string
		ReplicaExternalLabelName    *string
		QueryLogFile                string
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
	} {

		p := &monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ScrapeInterval:              tc.ScrapeInterval,
					ScrapeTimeout:               tc.ScrapeTimeout,
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
						Node: true,
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
				AttachMetadata: tc.ServiceMonitor.Spec.AttachMetadata,
			}
		}

		c := cg.generateK8SSDConfig(tc.ServiceMonitor.Spec.NamespaceSelector, tc.ServiceMonitor.Namespace, nil, nil, kubernetesSDRoleEndpoint, attachMetaConfig)
		s, err := yaml.Marshal(yaml.MapSlice{c})
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

	golden.Assert(t, string(s), "NamespaceSetCorrectlyForPodMonitor.golden")
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
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithLabelEnforce.golden")
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
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithJobName.golden")
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
	golden.Assert(t, string(cfg), "ProbeStaticTargetsConfigGenerationWithoutModule.golden")
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

	golden.Assert(t, string(cfg), "ProbeIngressSDConfigGeneration.golden")
}

func TestProbeIngressSDConfigGenerationWithShards(t *testing.T) {
	p := defaultPrometheus()
	p.Spec.Shards = ptr.To(int32(2))

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
	golden.Assert(t, string(cfg), "ProbeIngressSDConfigGenerationWithShards.golden")
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
		apiserverConfig *monitoringv1.APIServerConfig
		store           *assets.Store
		golden          string
	}{
		{
			nil,
			nil,
			"K8SSDConfigGenerationFirst.golden",
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
			"K8SSDConfigGenerationTwo.golden",
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
		golden.Assert(t, string(s), tc.golden)
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
					Namespace: "default",
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
				SigV4Assets: map[string]assets.SigV4Credentials{
					"alertmanager/auth/0": {
						AccessKeyID: "access-key",
						SecretKeyID: "secret-key",
					},
				},
			},
			nil,
			nil,
			nil,
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		golden.Assert(t, string(cfg), tc.golden)
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
	golden.Assert(t, string(cfg), "AlertmanagerAPIVersion.golden")
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
				Timeout:    (*monitoringv1.Duration)(ptr.To("60s")),
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
			golden.Get(t, "TestAdditionalScrapeConfigsAdditionalScrapeConfig.golden"),
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
		golden.Get(t, "AdditionalAlertRelabelConfigs.golden"),
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
	golden.Assert(t, string(cfg), "NoEnforcedNamespaceLabelServiceMonitor_Expected.golden")
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
	golden.Assert(t, string(cfg), "TestServiceMonitorWithEndpointSliceEnable_Expected.golden")
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
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelOnExcludedPodMonitor_Expected.golden")
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
	golden.Assert(t, string(cfg), "EnforcedNamespaceLabelOnExcludedServiceMonitor_Expected.golden")
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
	golden.Assert(t, string(cfg), "TestAdditionalAlertmanagers_Expected.golden")
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
	golden.Assert(t, string(cfg), "SettingHonorTimestampsInServiceMonitor.golden")
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
	golden.Assert(t, string(cfg), "SettingHonorTimestampsInPodMonitor.golden")
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
	golden.Assert(t, string(cfg), "HonorTimestampsOverriding.golden")
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
	golden.Assert(t, string(cfg), "SettingHonorLabels.golden")
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
	golden.Assert(t, string(cfg), "HonorLabelsOverriding.golden")
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
	}

	testCases := []struct {
		name              string
		sMons             map[string]*monitoringv1.ServiceMonitor
		pMons             map[string]*monitoringv1.PodMonitor
		probes            map[string]*monitoringv1.Probe
		oauth2Credentials map[string]assets.OAuth2Credentials
		golden            string
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
			golden: "probe_monitor_with_oauth2.golden",
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
			golden.Assert(t, string(cfg), tt.golden)
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
	golden.Assert(t, string(cfg), "PodTargetLabels.golden")
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
	golden.Assert(t, string(cfg), "PodTargetLabelsFromPodMonitor.golden")
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
	golden.Assert(t, string(cfg), "PodTargetLabelsFromPodMonitorAndGlobal.golden")
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
	for _, tc := range []struct {
		enforcedLimit int
		limit         int
		golden        string
	}{
		{
			enforcedLimit: -1,
			limit:         -1,
			golden:        "SampleLimits_NoLimit.golden",
		},
		{
			enforcedLimit: 1000,
			limit:         -1,
			golden:        "SampleLimits_Limit-1.golden",
		},
		{
			enforcedLimit: 1000,
			limit:         2000,
			golden:        "SampleLimits_Limit2000.golden",
		},
		{
			enforcedLimit: 1000,
			limit:         500,
			golden:        "SampleLimits_Limit500.golden",
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
								Name: "key",
							},
						},
					},
				},
			},
			golden: "RemoteReadConfig_v2.26.0_AuthorizationSafe.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestRemoteWriteConfig(t *testing.T) {
	sendNativeHistograms := true
	for _, tc := range []struct {
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
					BatchSendDeadline: "20s",
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
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
			golden: "RemoteWriteConfig_v2.10.0_1.golden",
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
			golden: "RemoteWriteConfig_v2.27.1_1.golden",
		},
		{
			version: "v2.45.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: monitoringv1.ManagedIdentity{
						ClientID: "client-id",
					},
				},
			},
			golden: "RemoteWriteConfig_v2.45.0_1.golden",
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
			golden: "RemoteWriteConfig_3.golden",
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				RemoteTimeout: "1s",
				Sigv4:         nil,
			},
			golden: "RemoteWriteConfig_v2.26.0_3.golden",
		},
		{
			version: "v2.26.0",
			remoteWrite: monitoringv1.RemoteWriteSpec{
				URL:           "http://example.com",
				Sigv4:         &monitoringv1.Sigv4{},
				RemoteTimeout: "1s",
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
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
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
					BatchSendDeadline: "20s",
					MaxRetries:        3,
					MinBackoff:        "1s",
					MaxBackoff:        "10s",
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
			golden: "RemoteWriteConfig_v2.39.0_1.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}

func TestBodySizeLimits(t *testing.T) {
	for _, tc := range []struct {
		version               string
		enforcedBodySizeLimit monitoringv1.ByteSize
		expectedErr           error
		golden                string
	}{
		{
			version:               "v2.27.0",
			enforcedBodySizeLimit: "1000MB",
			golden:                "BodySizeLimits_enforce1000MB_v2.27.0.golden",
		},
		{
			version:               "v2.28.0",
			enforcedBodySizeLimit: "1000MB",
			golden:                "BodySizeLimits_enforce1000MB_v2.28.0.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
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
			golden.Assert(t, string(cfg), tc.golden)
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
			golden.Assert(t, string(cfg), tc.golden)
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
				OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
			},
			golden: "TSDB_config_less_than_v2.39.0.golden",
		},
		{

			name: "TSDB config >= v2.39.0",
			tsdb: &monitoringv1.TSDBSpec{
				OutOfOrderTimeWindow: monitoringv1.Duration("10m"),
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
			golden.Assert(t, string(cfg), tc.golden)
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
					Scheme:   "http",
					URL:      "example.com",
					Path:     "/probe",
					ProxyURL: "socks://myproxy:9095",
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
						RelabelConfigs: []*monitoringv1.RelabelConfig{
							{
								TargetLabel: "foo",
								Replacement: "bar",
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
		name      string
		patchProm func(*monitoringv1.Prometheus)
		scSpec    monitoringv1alpha1.ScrapeConfigSpec
		golden    string
	}{
		{
			name:   "empty_scrape_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{},
			golden: "ScrapeConfigSpecConfig_Empty.golden",
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
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_HTTPSD.golden",
		},
		{
			name: "kubernetes_sd_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: "node",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD.golden",
		},
		{
			name: "kubernetes_sd_config_with_selectors",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: "node",
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  "node",
								Label: "type=infra",
								Field: "spec.unschedulable=false",
							},
						},
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_K8SSD_with_Selectors.golden",
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
				RelabelConfigs: []*monitoringv1.RelabelConfig{},
			},
			golden: "ScrapeConfigSpecConfig_EmptyRelabelConfig.golden",
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
			golden: "ScrapeConfigSpecConfig_BasicAuth.golden",
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
			golden: "ScrapeConfigSpecConfig_Authorization.golden",
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
			golden: "ScrapeConfigSpecConfig_TLSConfig.golden",
		},
		{
			name: "scheme",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				Scheme: ptr.To("HTTPS"),
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
				ScrapeInterval: (*monitoringv1.Duration)(ptr.To("15s")),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeInterval.golden",
		},
		{
			name: "scrape_timeout",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				ScrapeTimeout: (*monitoringv1.Duration)(ptr.To("10s")),
			},
			golden: "ScrapeConfigSpecConfig_ScrapeTimeout.golden",
		},
		{
			name: "non_empty_metric_relabel_config",
			scSpec: monitoringv1alpha1.ScrapeConfigSpec{
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
					{
						Regex:  "noisy_labels.*",
						Action: "labeldrop",
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_NonEmptyMetricRelabelConfig.golden",
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
						Type:  ptr.To("A"),
						Port:  ptr.To(9100),
					},
				},
			},
			golden: "ScrapeConfigSpecConfig_DNSSD_ARecord.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
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
		name      string
		patchProm func(*monitoringv1.Prometheus)
		scSpec    monitoringv1alpha1.ScrapeConfigSpec
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
						Scheme:       ptr.To("https"),
						Services:     []string{"prometheus", "alertmanager"},
						Tags:         []string{"tag1"},
						TagSeparator: ptr.To(";"),
						NodeMeta: map[string]string{
							"service": "service_name",
							"name":    "node_name",
						},
						AllowStale:           ptr.To(false),
						RefreshInterval:      (*monitoringv1.Duration)(ptr.To("30s")),
						ProxyUrl:             ptr.To("http://no-proxy.com"),
						NoProxy:              ptr.To("0.0.0.0"),
						ProxyFromEnvironment: ptr.To(true),
						ProxyConnectHeader: map[string]v1.SecretKeySelector{
							"header": {
								LocalObjectReference: v1.LocalObjectReference{
									Name: "foo",
								},
								Key: "proxy-header",
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
									Name: "foo",
								},
								Key: "credential",
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
			golden: "ConsulScrapeConfigTLSConfig.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
		})

	}
}

func TestScrapeConfigSpecConfigWithEC2SD(t *testing.T) {
	c := fake.NewSimpleClientset(
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
	for _, tc := range []struct {
		name        string
		scSpec      monitoringv1alpha1.ScrapeConfigSpec
		golden      string
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
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
						Port:            ptr.To(9100),
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
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
						Port:            ptr.To(9100),
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
						RefreshInterval: (*monitoringv1.Duration)(ptr.To("30s")),
						Port:            ptr.To(9100),
						Filters: []*monitoringv1alpha1.EC2Filter{
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
			golden: "ScrapeConfigSpecConfig_EC2SDConfigFilters.golden",
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
				assets.NewStore(c.CoreV1(), c.CoreV1()),
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

func TestTracingConfig(t *testing.T) {
	samplingTwo := resource.MustParse("0.5")
	testCases := []struct {
		tracingConfig *monitoringv1.PrometheusTracingConfig
		name          string
		expectedErr   bool
		golden        string
	}{
		{
			name: "Config only with endpoint",
			tracingConfig: &monitoringv1.PrometheusTracingConfig{
				Endpoint: "https://otel-collector.default.svc.local:3333",
			},
			golden:      "TracingConfig_Config_only_with_endpoint.golden",
			expectedErr: false,
		},
		{
			tracingConfig: &monitoringv1.PrometheusTracingConfig{
				ClientType:       ptr.To("grpc"),
				Endpoint:         "https://otel-collector.default.svc.local:3333",
				SamplingFraction: &samplingTwo,
				Headers: map[string]string{
					"custom": "header",
				},
				Compression: ptr.To("gzip"),
				Timeout:     (*monitoringv1.Duration)(ptr.To("10s")),
				Insecure:    ptr.To(false),
			},
			name:        "Expect valid config",
			expectedErr: false,
			golden:      "TracingConfig_Expect_valid_config.golden",
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
			golden.Assert(t, string(cfg), tc.golden)
		})
	}
}
