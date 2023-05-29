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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func testCreatePrometheusAgent(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}

}

func testAgentAndServerNameColision(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD); err != nil {
		t.Fatal(err)
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
	if err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}

}

// testMetricsFromOperatorWithAgentAndServer tests if scraping metrics from the Prometheus-Operator container
// succeeds with Prometheus server and agent deployed.
func testMetricsFromOperatorWithAgentAndServer(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	errorG, _ := errgroup.WithContext(context.Background())

	createPromOperatorFunc := func() error {
		_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, nil, nil, false, false, false)
		return err
	}
	sm := framework.MakeBasicServiceMonitor(name)
	createServiceMonitorFunc := func() error {
		sm.Spec = monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "prometheus-operator",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					TargetPort: &intstr.IntOrString{IntVal: 8080},
				},
			},
		}
		_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{})
		return err
	}
	createPromAgentFunc := func() error {
		prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
		_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD)
		return err
	}
	createPromFunc := func() error {
		prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
		prometheusCRD.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"group": name,
			},
		}
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
		return err
	}
	pSVC := framework.MakePrometheusService(name, name, v1.ServiceTypeClusterIP)
	createPromSvcFunc := func() error {
		_, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
		return err
	}

	errorG.Go(createPromOperatorFunc)
	errorG.Go(createServiceMonitorFunc)
	errorG.Go(createPromAgentFunc)
	errorG.Go(createPromFunc)
	errorG.Go(createPromSvcFunc)

	// Wait for the creation of all resources
	err := errorG.Wait()
	require.NoError(t, err)
	defer func() {
		err := framework.MonClientV1.ServiceMonitors(ns).Delete(context.Background(), sm.Name, metav1.DeleteOptions{})
		t.Fatal(err)
	}()

	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
		res, err := framework.PrometheusQuery(ns, pSVC.Name, "http", "prometheus_operator_spec_replicas")
		if err != nil {
			return false, errors.Wrap(err, "failed to query Prometheus")
		}

		if len(res) == 0 {
			return false, errors.Wrap(err, "query didn't return any metrics")
		}

		return true, nil
	})
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}
