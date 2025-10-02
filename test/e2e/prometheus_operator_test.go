// Copyright 2020 The prometheus-operator Authors
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

	"github.com/stretchr/testify/require"
)

// testServerTLS verifies that the Prometheus operator web server is working
// with HTTPS.
func testServerTLS(t *testing.T, namespace string) {
	skipPrometheusTests(t)

	ctx := context.Background()
	err := framework.WaitForServiceReady(ctx, namespace, prometheusOperatorServiceName)
	require.NoError(t, err)

	operatorService := framework.KubeClient.CoreV1().Services(namespace)
	request := operatorService.ProxyGet("https", prometheusOperatorServiceName, "https", "/healthz", nil)
	_, err = request.DoRaw(ctx)
	require.NoError(t, err)
}

// testPrometheusOperatorMetrics verifies that the Prometheus operator exposes
// the expected metrics.
func testPrometheusOperatorMetrics(t *testing.T, namespace string) {
	skipPrometheusTests(t)
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	operatorMetrics := []string{
		// Kubernetes client metrics.
		"prometheus_operator_kubernetes_client_http_requests_total",
		"prometheus_operator_kubernetes_client_http_request_duration_seconds_count",
		"prometheus_operator_kubernetes_client_http_request_duration_seconds_sum",
		"prometheus_operator_kubernetes_client_rate_limiter_duration_seconds_count",
		"prometheus_operator_kubernetes_client_rate_limiter_duration_seconds_sum",

		// Operator's info metrics.
		"prometheus_operator_build_info",
		"prometheus_operator_feature_gate",
		"prometheus_operator_kubelet_managed_resource",

		"prometheus_operator_list_operations_failed_total",
		"prometheus_operator_list_operations_total",
		"prometheus_operator_node_address_lookup_errors_total",
		"prometheus_operator_node_syncs_failed_total",
		"prometheus_operator_node_syncs_total",
		"prometheus_operator_ready",

		// Resource reconciler metrics.
		"prometheus_operator_reconcile_duration_seconds_bucket",
		"prometheus_operator_reconcile_duration_seconds_count",
		"prometheus_operator_reconcile_duration_seconds_sum",
		"prometheus_operator_reconcile_errors_total",
		"prometheus_operator_reconcile_operations_total",
		"prometheus_operator_reconcile_sts_delete_create_total",

		// Kubernetes work queue metrics.
		"prometheus_operator_workqueue_depth",
		"prometheus_operator_workqueue_adds_total",
		"prometheus_operator_workqueue_latency_seconds_bucket",
		"prometheus_operator_workqueue_latency_seconds_count",
		"prometheus_operator_workqueue_latency_seconds_sum",
		"prometheus_operator_workqueue_work_duration_seconds_bucket",
		"prometheus_operator_workqueue_work_duration_seconds_count",
		"prometheus_operator_workqueue_work_duration_seconds_sum",
		"prometheus_operator_workqueue_unfinished_work_seconds",
		"prometheus_operator_workqueue_longest_running_processor_seconds",
		"prometheus_operator_workqueue_retries_total",

		"prometheus_operator_status_update_errors_total",
		"prometheus_operator_status_update_operations_total",
		"prometheus_operator_syncs",
		"prometheus_operator_triggered_total",
		"prometheus_operator_watch_operations_failed_total",
		"prometheus_operator_watch_operations_total",

		"prometheus_operator_managed_resources",
		"prometheus_operator_spec_replicas",
		"prometheus_operator_spec_shards",
	}

	name := "test"
	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusCRD.Namespace = ns

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	err := framework.EnsureMetricsFromService(
		context.Background(),
		"https",
		namespace,
		prometheusOperatorServiceName,
		"https",
		operatorMetrics...,
	)

	require.NoError(t, err)
}
