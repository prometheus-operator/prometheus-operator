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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	expectedStartupProbe := prompkg.MakeExpectedStartupProbe()
	require.Equal(t, expectedStartupProbe, actualStartupProbe)

	actualLivenessProbe := dset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := prompkg.MakeExpectedLivenessProbe()
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := dset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := prompkg.MakeExpectedReadinessProbe()
	require.Equal(t, expectedReadinessProbe, actualReadinessProbe)

	testCorrectArgs(t, dset.Spec.Template.Spec.Containers[1].Args, dset.Spec.Template.Spec.Containers)
}

func TestStartupProbeTimeoutSecondsForDaemonSet(t *testing.T) {
	testcases := createTestCasesForTestStartupProbeTimeoutSeconds()

	for _, test := range testcases {
		dset, err := makeDaemonSetFromPrometheus(
			makePrometheusAgentForTestStartupProbeTimeoutSeconds(test.maximumStartupDurationSeconds))

		require.NoError(t, err)
		require.NotNil(t, dset.Spec.Template.Spec.Containers[0].StartupProbe)
		require.Equal(t, test.expectedStartupPeriodSeconds, dset.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, dset.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold)
	}
}

func makeDaemonSetFromPrometheus(p monitoringv1alpha1.PrometheusAgent) (*appsv1.DaemonSet, error) {
	logger := prompkg.NewLogger()
	cg, err := prompkg.NewConfigGenerator(logger, &p)
	if err != nil {
		return nil, err
	}

	return makeDaemonSet(
		&p,
		defaultTestConfig,
		cg,
		&operator.ShardedSecret{})
}

func TestPodTopologySpreadConstraintWithAdditionalLabelsForDaemonSet(t *testing.T) {
	testcases := createTestCasesForTestPodTopologySpreadConstraintWithAdditionalLabels()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dms, err := makeDaemonSetFromPrometheus(makePrometheusAgentForTestPodTopologySpreadConstraintWithAdditionalLabels(tc.spec))

			require.NoError(t, err)

			assert.NotEmpty(t, dms.Spec.Template.Spec.TopologySpreadConstraints)
			assert.Equal(t, tc.tsc, dms.Spec.Template.Spec.TopologySpreadConstraints[0])
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

func TestDaemonSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
		"kubectl.kubernetes.io/last-applied-configuration": "something",
		"kubectl.kubernetes.io/something":                  "something",
	}
	// kubectl annotations must not be on the daemonset so kubectl does
	// not manage the generated object
	expectedDaemonSetAnnotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	expectedDaemonSetLabels := map[string]string{
		"testlabel":                   "testlabelvalue",
		"operator.prometheus.io/name": "",
		"operator.prometheus.io/mode": "agent",
		"managed-by":                  "prometheus-operator",
	}

	expectedPodLabels := map[string]string{
		"app.kubernetes.io/name":       "prometheus-agent",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultPrometheusVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "",
		"operator.prometheus.io/name":  "",
	}

	dset, err := makeDaemonSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	})
	require.NoError(t, err)
	require.Equal(t, expectedDaemonSetLabels, dset.Labels)
	require.Equal(t, expectedDaemonSetAnnotations, dset.Annotations)
	require.Equal(t, expectedPodLabels, dset.Spec.Template.ObjectMeta.Labels)
}
