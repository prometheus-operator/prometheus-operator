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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		found := false
		for _, flag := range promArgs {
			if flag == test.expectedArg {
				found = true
				break
			}
		}
		require.Equal(t, test.shouldContain, found)
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
