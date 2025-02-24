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

	ctx := context.Background()
	err := framework.WaitForServiceReady(ctx, namespace, prometheusOperatorServiceName)
	require.NoError(t, err)

	// Explicitly check the client-go metrics to validate the registration
	// workaround we have in place due to
	// https://github.com/kubernetes-sigs/controller-runtime/issues/3054
	err = framework.EnsureMetricsFromService(
		ctx,
		"https",
		namespace,
		prometheusOperatorServiceName,
		"https",
		"prometheus_operator_kubernetes_client_http_requests_total",
		"prometheus_operator_kubernetes_client_http_request_duration_seconds_count",
		"prometheus_operator_kubernetes_client_http_request_duration_seconds_sum",
		"prometheus_operator_kubernetes_client_rate_limiter_duration_seconds_count",
		"prometheus_operator_kubernetes_client_rate_limiter_duration_seconds_sum",
	)
	require.NoError(t, err)
}
