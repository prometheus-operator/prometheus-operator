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

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

// testScrapeConfigCreation tests multiple ScrapeConfig definitions
func testScrapeConfigCreation(t *testing.T) {
	skipPrometheusTests(t)
	t.Parallel()

	fiveMins := monitoringv1.Duration("5m")

	tests := []struct {
		name          string
		spec          monitoringv1alpha1.ScrapeConfigSpec
		expectedError bool
	}{
		{
			name: "empty-scrape-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{},
		},
		{
			name: "static-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090"},
						Labels: map[monitoringv1.LabelName]string{
							"label1": "value1",
							"label2": "value2",
						},
					},
				},
			},
		},
		{
			name: "http-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL:             "http://localhost:8080/file.json",
						RefreshInterval: &fiveMins,
					},
				},
			},
		},
		{
			name: "file-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
					{
						Files: []monitoringv1alpha1.SDFile{
							"/etc/prometheus/sd/file.json",
							"/etc/prometheus/sd/file.yaml",
						},
						RefreshInterval: &fiveMins,
					},
				},
			},
		},
		{
			name: "invalid-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
					{
						Files: []monitoringv1alpha1.SDFile{
							"/etc/prometheus/sd/file.invalid",
						},
						RefreshInterval: &fiveMins,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)

			sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
			sc.Spec = test.spec
			_, err := framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(context.Background(), sc, metav1.CreateOptions{})

			if test.expectedError {
				require.Error(t, err)
				require.Truef(t, apierrors.IsInvalid(err), "expected Invalid error but got %v", err)
				return
			}
			require.NoError(t, err)
			require.Falsef(t, test.expectedError, "expected error but got nil")
		})
	}
}

// testScrapeConfigLifecycle tests 3 things:
// 1. Creating a ScrapeConfig and checking that 2 targets appear in Prometheus
// 2. Updating that ScrapeConfig by adding a target and checking that 3 targets appear in Prometheus
// 3. Deleting that ScrapeConfig and checking that 0 targets appear in Prometheus
func testScrapeConfigLifecycle(t *testing.T) {
	skipPrometheusTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, true)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "prom", "group", 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"role": "scrapeconfig",
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// 1. Create a ScrapeConfig and check that its targets appear in Prometheus
	sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090"},
		},
	}
	_, err = framework.CreateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 2)
	require.NoError(t, err)

	// 2. Update the ScrapeConfig and add a target. Then, check that 3 targets appear in Prometheus.
	sc, err = framework.GetScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090", "target3:9090"},
		},
	}

	_, err = framework.UpdateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 3)
	require.NoError(t, err)

	// 3. Remove the ScrapeConfig and check that the targets disappear in Prometheus
	err = framework.DeleteScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 0)
	require.NoError(t, err)
}

// testPromOperatorStartsWithoutScrapeConfigCRD deletes the ScrapeConfig CRD from the cluster and then starts
// prometheus-operator to check that it doesn't crash.
func testPromOperatorStartsWithoutScrapeConfigCRD(t *testing.T) {
	skipPrometheusAllNSTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	err := framework.DeleteCRD(context.Background(), "scrapeconfigs.monitoring.coreos.com")
	require.NoError(t, err)

	_, err = framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, false)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	// Check if Prometheus Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"app.kubernetes.io/name": "prometheus-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), opts)
	require.NoError(t, err)
	require.Equalf(t, 1, len(pl.Items), "expected 1 Prometheus Operator pods, but got %v", len(pl.Items))

	restarts, err := framework.GetPodRestartCount(context.Background(), ns, pl.Items[0].GetName())
	require.NoError(t, err)
	require.Emptyf(t, restarts, "expected to get 1 container but got %d", len(restarts))

	for _, restart := range restarts {
		require.Emptyf(t, restart, "expected Prometheus Operator to never restart during entire test execution but got %d restarts", restart)
	}

	// re-create Prometheus-Operator to reinstall the CRDs
	_, err = framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, true)
	require.NoError(t, err)
}
