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

package prometheusagent

import (
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

var (
	defaultTestConfig = &prompkg.Config{
		LocalHost:                  "localhost",
		ReloaderConfig:             operator.DefaultReloaderTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: operator.DefaultPrometheusBaseImage,
		ThanosDefaultBaseImage:     operator.DefaultThanosBaseImage,
	}
)

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Web: &monitoringv1.PrometheusWebSpec{
					WebConfigFileFields: monitoringv1.WebConfigFileFields{
						TLSConfig: &monitoringv1.WebTLSConfig{
							KeySecret: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "some-secret",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								ConfigMap: &v1.ConfigMapKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "some-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	expectedProbeHandler := func(probePath string) v1.ProbeHandler {
		return v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   probePath,
				Port:   intstr.FromString("web"),
				Scheme: "HTTPS",
			},
		}
	}

	actualStartupProbe := sset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
	if !reflect.DeepEqual(actualStartupProbe, expectedStartupProbe) {
		t.Fatalf("Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)
	}

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9090/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	if !reloadURLFound {
		t.Fatalf("expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)
	}

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/prometheus/web_config/web-config.yaml",
		"--reload-url=https://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
		}
	}
}

func TestWALCompression(t *testing.T) {
	var (
		tr = true
		fa = false
	)
	tests := []struct {
		version       string
		enabled       *bool
		expectedArg   string
		shouldContain bool
	}{
		// Nil should not have either flag.
		{"v2.30.0", &fa, "--storage.agent.wal-compression", false},
		{"v2.32.0", nil, "--storage.agent.wal-compression", false},
		{"v2.32.0", &fa, "--no-storage.agent.wal-compression", true},
		{"v2.32.0", &tr, "--storage.agent.wal-compression", true},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version:        test.version,
					WALCompression: test.enabled,
				},
			},
		})
		require.NoError(t, err)

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		found := false
		for _, flag := range promArgs {
			if flag == test.expectedArg {
				found = true
				break
			}
		}

		if found != test.shouldContain {
			if test.shouldContain {
				t.Fatalf("expected Prometheus args to contain %v, but got %v", test.expectedArg, promArgs)
			} else {
				t.Fatalf("expected Prometheus args to NOT contain %v, but got %v", test.expectedArg, promArgs)
			}
		}
	}
}

func TestStartupProbeTimeoutSeconds(t *testing.T) {
	tests := []struct {
		maximumStartupDurationSeconds   *int32
		expectedStartupPeriodSeconds    int32
		expectedStartupFailureThreshold int32
	}{
		{
			maximumStartupDurationSeconds:   nil,
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(600)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 10,
		},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					MaximumStartupDurationSeconds: test.maximumStartupDurationSeconds,
				},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, sset.Spec.Template.Spec.Containers[0].StartupProbe)
		require.Equal(t, test.expectedStartupPeriodSeconds, sset.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, sset.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold)
	}
}

func newLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowWarn())
}

func makeStatefulSetFromPrometheus(p monitoringv1alpha1.PrometheusAgent) (*appsv1.StatefulSet, error) {
	logger := newLogger()
	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	if err != nil {
		return nil, err
	}

	return makeStatefulSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		"",
		0,
		&operator.ShardedSecret{})
}

func TestPodTopologySpreadConstraintWithAdditionalLabels(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec monitoringv1alpha1.PrometheusAgentSpec
		tsc  v1.TopologySpreadConstraint
	}{
		{
			name: "without labelSelector and additionalLabels",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
			},
		},
		{
			name: "with labelSelector and without additionalLabels",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "prometheus",
					},
				},
			},
		},
		{
			name: "with labelSelector and additionalLabels as ShardAndNameResource",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							AdditionalLabelSelectors: ptr.To(monitoringv1.ShardAndResourceNameLabelSelector),
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":                          "prometheus",
						"app.kubernetes.io/instance":   "test",
						"app.kubernetes.io/managed-by": "prometheus-operator",
						"app.kubernetes.io/name":       "prometheus-agent",
						"operator.prometheus.io/name":  "test",
						"operator.prometheus.io/shard": "0",
					},
				},
			},
		},
		{
			name: "with labelSelector and additionalLabels as ResourceName",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							AdditionalLabelSelectors: ptr.To(monitoringv1.ResourceNameLabelSelector),
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":                          "prometheus",
						"app.kubernetes.io/instance":   "test",
						"app.kubernetes.io/managed-by": "prometheus-operator",
						"app.kubernetes.io/name":       "prometheus-agent",
						"operator.prometheus.io/name":  "test",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sts, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns-test",
				},
				Spec: tc.spec,
			})

			require.NoError(t, err)

			assert.NotEmpty(t, sts.Spec.Template.Spec.TopologySpreadConstraints)
			assert.Equal(t, tc.tsc, sts.Spec.Template.Spec.TopologySpreadConstraints[0])
		})
	}
}

func TestAutomountServiceAccountToken(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		automountServiceAccountToken *bool
		expectedValue                bool
	}{
		{
			name:                         "automountServiceAccountToken not set",
			automountServiceAccountToken: nil,
			expectedValue:                true,
		},
		{
			name:                         "automountServiceAccountToken set to true",
			automountServiceAccountToken: ptr.To(true),
			expectedValue:                true,
		},
		{
			name:                         "automountServiceAccountToken set to false",
			automountServiceAccountToken: ptr.To(false),
			expectedValue:                false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: monitoringv1alpha1.PrometheusAgentSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						AutomountServiceAccountToken: tc.automountServiceAccountToken,
					},
				},
			})
			require.NoError(t, err)

			if sset.Spec.Template.Spec.AutomountServiceAccountToken == nil {
				t.Fatalf("expected automountServiceAccountToken to be set")
			}

			if *sset.Spec.Template.Spec.AutomountServiceAccountToken != tc.expectedValue {
				t.Fatalf("expected automountServiceAccountToken to be %v", tc.expectedValue)
			}
		})
	}
}
