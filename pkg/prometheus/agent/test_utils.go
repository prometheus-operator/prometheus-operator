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
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/require"
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

func makeSpecForTestListenTLS() monitoringv1alpha1.PrometheusAgentSpec {
	return monitoringv1alpha1.PrometheusAgentSpec{
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
	}
}

func makeExpectedProbeHandler(probePath string) v1.ProbeHandler {
	return v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path:   probePath,
			Port:   intstr.FromString("web"),
			Scheme: "HTTPS",
		},
	}
}

func makeExpectedStartupProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
}

func makeExpectedLivenessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
}

func makeExpectedReadinessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
}

func testCorrectArgs(t *testing.T, actualArgs []string, actualContainers []v1.Container) {
	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9090/-/reload"
	reloadURLFound := false
	for _, arg := range actualArgs {
		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
			break
		}
	}
	require.True(t, reloadURLFound)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/prometheus/web_config/web-config.yaml",
		"--reload-url=https://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}
	for _, c := range actualContainers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args)
		}
	}
}

func newLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowWarn())
}

type testcaseForTestPodTopologySpreadConstraintWithAdditionalLabels struct {
	name string
	spec monitoringv1alpha1.PrometheusAgentSpec
	tsc  v1.TopologySpreadConstraint
}

func createTestCasesForTestPodTopologySpreadConstraintWithAdditionalLabels() []testcaseForTestPodTopologySpreadConstraintWithAdditionalLabels {
	return []testcaseForTestPodTopologySpreadConstraintWithAdditionalLabels{
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
	}
}

func makePrometheusAgentForTestPodTopologySpreadConstraintWithAdditionalLabels(spec monitoringv1alpha1.PrometheusAgentSpec) monitoringv1alpha1.PrometheusAgent {
	return monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "ns-test",
		},
		Spec: spec,
	}
}

type testcaseForTestAutomountServiceAccountToken struct {
	name                         string
	automountServiceAccountToken *bool
	expectedValue                bool
}

func createTestCasesForTestAutomountServiceAccountToken() []testcaseForTestAutomountServiceAccountToken {
	return []testcaseForTestAutomountServiceAccountToken{
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
	}
}

func makePrometheusAgentForTestAutomountServiceAccountToken(automountServiceAccountToken *bool) monitoringv1alpha1.PrometheusAgent {
	return monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				AutomountServiceAccountToken: automountServiceAccountToken,
			},
		},
	}
}
