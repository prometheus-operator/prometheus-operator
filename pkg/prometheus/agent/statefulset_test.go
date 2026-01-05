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
	"slices"
	"strings"
	"testing"

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

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: makeSpecForTestListenTLS(),
	})
	require.NoError(t, err)

	actualStartupProbe := sset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := prompkg.MakeExpectedStartupProbe()
	require.Equal(t, expectedStartupProbe, actualStartupProbe)

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := prompkg.MakeExpectedLivenessProbe()
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := prompkg.MakeExpectedReadinessProbe()
	require.Equal(t, expectedReadinessProbe, actualReadinessProbe)

	testCorrectArgs(t, sset.Spec.Template.Spec.Containers[1].Args, sset.Spec.Template.Spec.Containers)
}

func TestWALCompression(t *testing.T) {
	tests := []struct {
		version       string
		enabled       *bool
		expectedArg   string
		shouldContain bool
	}{
		// Nil should not have either flag.
		{"v2.30.0", ptr.To(false), "--storage.agent.wal-compression", false},
		{"v2.32.0", nil, "--storage.agent.wal-compression", false},
		{"v2.32.0", ptr.To(false), "--no-storage.agent.wal-compression", true},
		{"v2.32.0", ptr.To(true), "--storage.agent.wal-compression", true},
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
		require.Equal(t, test.shouldContain, slices.Contains(promArgs, test.expectedArg))
	}
}

func TestPrometheusAgentCommandLineFlag(t *testing.T) {
	tests := []struct {
		version       string
		expectedArg   string
		shouldContain bool
	}{
		{"v3.0.0", "--agent", true},
		{"v3.0.0-beta.0", "--agent", true},
		{"v2.53.0", "--agent", false},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: test.version,
				},
			},
		})
		require.NoError(t, err)

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		require.Equal(t, test.shouldContain, slices.Contains(promArgs, test.expectedArg))
	}
}

func TestStartupProbeTimeoutSeconds(t *testing.T) {
	testcases := createTestCasesForTestStartupProbeTimeoutSeconds()

	for _, test := range testcases {
		sset, err := makeStatefulSetFromPrometheus(
			makePrometheusAgentForTestStartupProbeTimeoutSeconds(test.maximumStartupDurationSeconds))

		require.NoError(t, err)
		require.NotNil(t, sset.Spec.Template.Spec.Containers[0].StartupProbe)
		require.Equal(t, test.expectedStartupPeriodSeconds, sset.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, sset.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold)
	}
}

func makeStatefulSetFromPrometheus(p monitoringv1alpha1.PrometheusAgent) (*appsv1.StatefulSet, error) {
	logger := prompkg.NewLogger()
	cg, err := prompkg.NewConfigGenerator(logger, &p)
	if err != nil {
		return nil, err
	}

	return makeStatefulSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		"abc",
		0,
		&operator.ShardedSecret{})
}

func TestPodTopologySpreadConstraintWithAdditionalLabels(t *testing.T) {
	testcases := createTestCasesForTestPodTopologySpreadConstraintWithAdditionalLabels()

	// The appended test case is specific for StatefulSet mode (not for DaemonSet mode)
	// because it has operator.prometheus.io/shard label (DaemonSet doesn't support sharding).
	testcases = append(testcases, testcaseForTestPodTopologySpreadConstraintWithAdditionalLabels{
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
	})

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sts, err := makeStatefulSetFromPrometheus(makePrometheusAgentForTestPodTopologySpreadConstraintWithAdditionalLabels(tc.spec))

			require.NoError(t, err)

			assert.NotEmpty(t, sts.Spec.Template.Spec.TopologySpreadConstraints)
			assert.Equal(t, tc.tsc, sts.Spec.Template.Spec.TopologySpreadConstraints[0])
		})
	}
}

func TestAutomountServiceAccountToken(t *testing.T) {
	testcases := createTestCasesForTestAutomountServiceAccountToken()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(makePrometheusAgentForTestAutomountServiceAccountToken(tc.automountServiceAccountToken))
			require.NoError(t, err)
			require.NotNil(t, sset.Spec.Template.Spec.AutomountServiceAccountToken)
			require.Equal(t, tc.expectedValue, *sset.Spec.Template.Spec.AutomountServiceAccountToken)
		})
	}
}

func TestStatefulSetDNSPolicyAndDNSConfig(t *testing.T) {
	// Monitoring DNS settings
	monitoringDNSPolicy := v1.DNSClusterFirst
	monitoringDNSConfig := &monitoringv1.PodDNSConfig{
		Nameservers: []string{"8.8.8.8", "8.8.4.4"},
		Searches:    []string{"custom.search"},
		Options: []monitoringv1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: ptr.To("5"),
			},
		},
	}
	monitoringDNSPolicyPtr := ptr.To(monitoringv1.DNSPolicy(monitoringDNSPolicy))

	// Create the PrometheusAgent object with DNS settings
	prometheusAgent := monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				DNSPolicy: monitoringDNSPolicyPtr,
				DNSConfig: monitoringDNSConfig,
			},
		},
	}

	// Generate the StatefulSet
	sset, err := makeStatefulSetFromPrometheus(prometheusAgent)
	require.NoError(t, err)

	// Validate the DNS Policy
	require.Equal(t, v1.DNSClusterFirst, sset.Spec.Template.Spec.DNSPolicy, "expected DNS policy to match")

	// Validate the DNS Config
	require.NotNil(t, sset.Spec.Template.Spec.DNSConfig, "expected DNS config to be set")
	require.Equal(t, monitoringDNSConfig.Nameservers, sset.Spec.Template.Spec.DNSConfig.Nameservers, "expected nameservers to match")
	require.Equal(t, monitoringDNSConfig.Searches, sset.Spec.Template.Spec.DNSConfig.Searches, "expected searches to match")

	require.Equal(t, len(monitoringDNSConfig.Options), len(sset.Spec.Template.Spec.DNSConfig.Options), "expected options length to match")
	for i, option := range monitoringDNSConfig.Options {
		k8sOption := sset.Spec.Template.Spec.DNSConfig.Options[i]
		require.Equal(t, option.Name, k8sOption.Name, "expected option names to match")
		require.Equal(t, option.Value, k8sOption.Value, "expected option values to match")
	}
}

func TestScrapeFailureLogFileVolumeMountPresent(t *testing.T) {
	sset, err := makeDaemonSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeFailureLogFile: ptr.To("file.log"),
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.True(t, found, "Volume for scrape failure log file not found.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Scrape failure log file not mounted.")
}

func TestScrapeFailureLogFileVolumeMountNotPresent(t *testing.T) {
	// An emptyDir is only mounted by the Operator if the given
	// path is only a base filename.
	sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeFailureLogFile: ptr.To("/tmp/file.log"),
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.False(t, found, "Volume for scrape failure file found, when it shouldn't be.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.False(t, found, "Scrape failure log file mounted, when it shouldn't be.")
}

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
		"kubectl.kubernetes.io/last-applied-configuration": "something",
		"kubectl.kubernetes.io/something":                  "something",
	}

	// kubectl annotations must not be on the statefulset so kubectl does
	// not manage the generated object
	expectedStatefulSetAnnotations := map[string]string{
		"prometheus-operator-input-hash": "abc",
		"testannotation":                 "testannotationvalue",
	}

	expectedStatefulSetLabels := map[string]string{
		"testlabel":                    "testlabelvalue",
		"operator.prometheus.io/name":  "test",
		"operator.prometheus.io/shard": "0",
		"operator.prometheus.io/mode":  "agent",
		"managed-by":                   "prometheus-operator",
		"app.kubernetes.io/instance":   "test",
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/name":       "prometheus-agent",
	}

	expectedPodLabels := map[string]string{
		"app.kubernetes.io/name":       "prometheus-agent",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultPrometheusVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "test",
		"operator.prometheus.io/name":  "test",
		"operator.prometheus.io/shard": "0",
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "ns",
			Labels:      labels,
			Annotations: annotations,
		},
	})
	require.NoError(t, err)

	require.Equal(t, expectedStatefulSetLabels, sset.Labels)
	require.Equal(t, expectedStatefulSetAnnotations, sset.Annotations)
	require.Equal(t, expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels)
}

func TestStatefulSetenableServiceLinks(t *testing.T) {
	tests := []struct {
		enableServiceLinks         *bool
		expectedEnableServiceLinks *bool
	}{
		{enableServiceLinks: ptr.To(false), expectedEnableServiceLinks: ptr.To(false)},
		{enableServiceLinks: ptr.To(true), expectedEnableServiceLinks: ptr.To(true)},
		{enableServiceLinks: nil, expectedEnableServiceLinks: nil},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					EnableServiceLinks: test.enableServiceLinks,
				},
			},
		})
		require.NoError(t, err)

		if test.expectedEnableServiceLinks != nil {
			require.NotNil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to be non-nil")
			require.Equal(t, *test.expectedEnableServiceLinks, *sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to match")
		} else {
			require.Nil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to be nil")
		}
	}
}

func TestStatefulPodManagementPolicy(t *testing.T) {
	for _, tc := range []struct {
		podManagementPolicy *monitoringv1.PodManagementPolicyType
		exp                 appsv1.PodManagementPolicyType
	}{
		{
			podManagementPolicy: nil,
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.ParallelPodManagement),
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.OrderedReadyPodManagement),
			exp:                 appsv1.OrderedReadyPodManagement,
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
				Spec: monitoringv1alpha1.PrometheusAgentSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						PodManagementPolicy: tc.podManagementPolicy,
					},
				},
			})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.PodManagementPolicy)
		})
	}
}

func TestStatefulSetUpdateStrategy(t *testing.T) {
	for _, tc := range []struct {
		updateStrategy *monitoringv1.StatefulSetUpdateStrategy
		exp            appsv1.StatefulSetUpdateStrategy
	}{
		{
			updateStrategy: nil,
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.OnDeleteStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
				Spec: monitoringv1alpha1.PrometheusAgentSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						UpdateStrategy: tc.updateStrategy,
					},
				},
			})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.UpdateStrategy)
		})
	}
}
