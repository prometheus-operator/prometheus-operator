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
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

func TestListenTLSForDaemonSet(t *testing.T) {
	dset, err := makeDaemonSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: makeSpecForTestListenTLS(),
	})
	require.NoError(t, err)

	actualStartupProbe := dset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := makeExpectedStartupProbe()
	require.Equal(t, expectedStartupProbe, actualStartupProbe)

	actualLivenessProbe := dset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := makeExpectedLivenessProbe()
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := dset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := makeExpectedReadinessProbe()
	require.Equal(t, expectedReadinessProbe, actualReadinessProbe)

	testCorrectArgs(t, dset.Spec.Template.Spec.Containers[1].Args, dset.Spec.Template.Spec.Containers)
}

func TestWALCompressionForDaemonSet(t *testing.T) {
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
		dset, err := makeDaemonSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version:        test.version,
					WALCompression: test.enabled,
				},
			},
		})
		require.NoError(t, err)

		promArgs := dset.Spec.Template.Spec.Containers[0].Args
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

func TestStartupProbeTimeoutSecondsForDaemonSet(t *testing.T) {
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
		dset, err := makeDaemonSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
			Spec: monitoringv1alpha1.PrometheusAgentSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					MaximumStartupDurationSeconds: test.maximumStartupDurationSeconds,
				},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, dset.Spec.Template.Spec.Containers[0].StartupProbe)
		require.Equal(t, test.expectedStartupPeriodSeconds, dset.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, dset.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold)
	}
}

func makeDaemonSetFromPrometheus(p monitoringv1alpha1.PrometheusAgent) (*appsv1.DaemonSet, error) {
	logger := newLogger()
	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	if err != nil {
		return nil, err
	}

	return makeDaemonSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		&operator.ShardedSecret{})
}

func TestPodTopologySpreadConstraintWithAdditionalLabelsForDaemonSet(t *testing.T) {
	testcases := createTestCasesForTestPodTopologySpreadConstraintWithAdditionalLabels()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sts, err := makeDaemonSetFromPrometheus(makePrometheusAgentForTestPodTopologySpreadConstraintWithAdditionalLabels(tc.spec))

			require.NoError(t, err)

			assert.NotEmpty(t, sts.Spec.Template.Spec.TopologySpreadConstraints)
			assert.Equal(t, tc.tsc, sts.Spec.Template.Spec.TopologySpreadConstraints[0])
		})
	}
}

func TestAutomountServiceAccountTokenForDaemonSet(t *testing.T) {
	testcases := createTestCasesForTestAutomountServiceAccountToken()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dset, err := makeDaemonSetFromPrometheus(makePrometheusAgentForTestAutomountServiceAccountToken(tc.automountServiceAccountToken))
			require.NoError(t, err)
			require.NotNil(t, dset.Spec.Template.Spec.AutomountServiceAccountToken)
			require.Equal(t, tc.expectedValue, *dset.Spec.Template.Spec.AutomountServiceAccountToken)
		})
	}
}
