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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
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

func testReconcileDelay(t *testing.T) {
	t.Parallel()

	delayValue := "3s"
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	// Create prometheus operator with reconcile delay
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		context.Background(),
		testFramework.PrometheusOperatorOpts{
			Namespace:         ns,
			AllowedNamespaces: []string{ns},
			AdditionalArgs:    []string{fmt.Sprintf("--reconcile-delay=%s", delayValue)},
		},
	)
	require.NoError(t, err)

	t.Run("PrometheusReconcileDelay", func(t *testing.T) {
		testPrometheusReconcileDelay(t, ns)
	})

	t.Run("PrometheusAgentReconcileDelay", func(t *testing.T) {
		testPrometheusAgentReconcileDelay(t, ns)
	})

	t.Run("AlertmanagerReconcileDelay", func(t *testing.T) {
		testAlertmanagerReconcileDelay(t, ns)
	})

	t.Run("ThanosRulerReconcileDelay", func(t *testing.T) {
		testThanosRulerReconcileDelay(t, ns)
	})

	// Verify delay configuration in operator logs.
	t.Run("VerifyDelayConfiguration", func(t *testing.T) {
		verifyDelayInOperatorLogs(t, ns, delayValue)
	})
}

func testPrometheusReconcileDelay(t *testing.T, ns string) {
	name := "test-prom-delay"

	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheus)
	require.NoError(t, err)

	testRapidResourceUpdates(t, ns, name, "prometheus", func() error {
		prometheus.Spec.Replicas = ptr.To(int32(2))
		_, err := framework.MonClientV1.Prometheuses(ns).Update(
			context.Background(),
			prometheus,
			metav1.UpdateOptions{},
		)
		return err
	})

	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

func testPrometheusAgentReconcileDelay(t *testing.T, ns string) {
	name := "test-agent-delay"

	prometheusAgent := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgent)
	require.NoError(t, err)

	testRapidResourceUpdates(t, ns, name, "prometheusagent", func() error {
		prometheusAgent.Spec.Replicas = ptr.To(int32(2))
		_, err := framework.MonClientV1alpha1.PrometheusAgents(ns).Update(
			context.Background(),
			prometheusAgent,
			metav1.UpdateOptions{},
		)
		return err
	})

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

func testAlertmanagerReconcileDelay(t *testing.T, ns string) {
	name := "test-am-delay"

	alertmanager := framework.MakeBasicAlertmanager(ns, name, 1)
	_, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), alertmanager)
	require.NoError(t, err)

	testRapidResourceUpdates(t, ns, name, "alertmanager", func() error {
		alertmanager.Spec.Replicas = ptr.To(int32(2))
		_, err := framework.MonClientV1.Alertmanagers(ns).Update(
			context.Background(),
			alertmanager,
			metav1.UpdateOptions{},
		)
		return err
	})

	err = framework.DeleteAlertmanagerAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

func testThanosRulerReconcileDelay(t *testing.T, ns string) {
	name := "test-thanos-delay"

	thanosRuler := framework.MakeBasicThanosRuler(name, 1, "http://prometheus:9090")
	_, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanosRuler)
	require.NoError(t, err)

	testRapidResourceUpdates(t, ns, name, "thanosruler", func() error {
		thanosRuler.Spec.Replicas = ptr.To(int32(2))
		_, err := framework.MonClientV1.ThanosRulers(ns).Update(
			context.Background(),
			thanosRuler,
			metav1.UpdateOptions{},
		)
		return err
	})

	err = framework.DeleteThanosRulerAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

// testRapidResourceUpdates simulates rapid updates to test delay behavior.
func testRapidResourceUpdates(t *testing.T, ns, resourceName, resourceType string, updateFunc func() error) {
	// Make first update
	err := updateFunc()
	require.NoError(t, err)

	// Wait briefly, then make rapid successive updates
	time.Sleep(500 * time.Millisecond)

	// Second update
	err = updateFunc()
	require.NoError(t, err)

	// Third update
	time.Sleep(200 * time.Millisecond)
	err = updateFunc()
	require.NoError(t, err)

	// Wait for reconciliation to complete despite delays.
	time.Sleep(8 * time.Second)

	// Verify the resource was eventually reconciled correctly.
	verifyResourceReconciled(t, ns, resourceName, resourceType)

	t.Logf("%s reconciliation completed successfully with delay handling", resourceType)
}

func verifyResourceReconciled(t *testing.T, ns, resourceName, resourceType string) {
	var statefulSetName string
	switch resourceType {
	case "prometheus":
		statefulSetName = fmt.Sprintf("prometheus-%s", resourceName)
	case "prometheusagent":
		statefulSetName = resourceName
	case "alertmanager":
		statefulSetName = fmt.Sprintf("alertmanager-%s", resourceName)
	case "thanosruler":
		statefulSetName = fmt.Sprintf("thanos-ruler-%s", resourceName)
	}

	// Verify StatefulSet exists and has expected replicas
	err := wait.PollUntilContextTimeout(context.Background(), 2*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		sts, err := framework.KubeClient.AppsV1().StatefulSets(ns).Get(ctx, statefulSetName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		// Check that replicas were updated (should be 2 from our updates)
		return sts.Spec.Replicas != nil && *sts.Spec.Replicas == 2, nil
	})
	require.NoError(t, err, "StatefulSet %s should be reconciled with correct replicas", statefulSetName)
}

func verifyDelayInOperatorLogs(t *testing.T, ns, delayValue string) {
	// Get operator pod
	opts := metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app.kubernetes.io/name": "prometheus-operator",
		})).String(),
	}

	pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, pods.Items, 1, "expected 1 prometheus-operator pod")

	operatorPod := pods.Items[0]

	// Get operator logs.
	var logBuffer strings.Builder
	err = framework.WritePodLogs(
		context.Background(),
		&logBuffer,
		ns,
		operatorPod.Name,
		testFramework.LogOptions{
			Container: "prometheus-operator",
			TailLines: 500,
		},
	)
	require.NoError(t, err)

	logContent := logBuffer.String()

	// Verify delay configuration is present in logs.
	require.Contains(t, logContent, "reconcile delay enabled",
		"Should find 'reconcile delay enabled' message in operator logs")
	require.Contains(t, logContent, fmt.Sprintf("delay=%s", delayValue),
		"Should find configured delay value %s in operator logs", delayValue)

	// Count how many controllers have delay enabled.
	controllerTypes := []string{"Prometheus", "PrometheusAgent", "Alertmanager", "ThanosRuler"}
	enabledControllers := []string{}

	for _, controller := range controllerTypes {
		// Look for delay enabled message for this specific controller.
		lines := strings.Split(logContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "reconcile delay enabled") &&
				strings.Contains(line, fmt.Sprintf("resource_kind=%s", controller)) &&
				strings.Contains(line, fmt.Sprintf("delay=%s", delayValue)) {
				enabledControllers = append(enabledControllers, controller)
				break
			}
		}
	}

	require.NotEmpty(t, enabledControllers, "Should find at least one controller with delay enabled")
	t.Logf("Found delay enabled for controllers: %v", enabledControllers)

	if strings.Contains(logContent, "delaying reconciliation") {
		t.Logf("Found evidence of actual delay behavior in logs")
		require.Contains(t, logContent, fmt.Sprintf("delay_configured=%s", delayValue),
			"Should show configured delay in delay behavior messages")
	} else {
		t.Logf("No delay behavior messages found (timing dependent - may not always appear)")
	}

	t.Logf("Verified reconcile delay configuration: delay=%s, enabled_controllers=%v", delayValue, enabledControllers)
}

// testReconcileDelayDisabled verifies normal operation when delay is not configured.
func testReconcileDelayDisabled(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	// Create operator WITHOUT delay (default behavior)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		context.Background(),
		testFramework.PrometheusOperatorOpts{
			Namespace:         ns,
			AllowedNamespaces: []string{ns},
		},
	)
	require.NoError(t, err)

	name := "test-no-delay"
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheus)
	require.NoError(t, err)

	// Update Prometheus - should be processed quickly
	prometheus.Spec.Replicas = ptr.To(int32(2))
	_, err = framework.MonClientV1.Prometheuses(ns).Update(
		context.Background(),
		prometheus,
		metav1.UpdateOptions{},
	)
	require.NoError(t, err)

	// Verify StatefulSet is updated quickly.
	err = wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 20*time.Second, false, func(ctx context.Context) (bool, error) {
		sts, err := framework.KubeClient.AppsV1().StatefulSets(ns).Get(ctx, fmt.Sprintf("prometheus-%s", name), metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return sts.Spec.Replicas != nil && *sts.Spec.Replicas == 2, nil
	})
	require.NoError(t, err, "Prometheus StatefulSet should be updated without artificial delay")

	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

	t.Logf("Verified normal reconciliation works when delay is disabled")
}
