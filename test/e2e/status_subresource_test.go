// Copyright 2022 The prometheus-operator Authors
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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

// testFinalizerWhenStatusForConfigResourcesEnabled tests the adding/removing of status-cleanup finalizer for Prometheus when StatusForConfigurationResourcesFeature is enabled.
func testFinalizerWhenStatusForConfigResourcesEnabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.StatusForConfigurationResourcesFeature},
		},
	)
	require.NoError(t, err)

	name := "status-cleanup-finalizer-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	pm, err := framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)
	finalizers := pm.GetFinalizers()
	require.NotEmpty(t, finalizers, "finalizers list should not be empty")
	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err, "failed to delete Prometheus with status-cleanup finalizer")
}

// testServiceMonitorStatusSubresource validates ServiceMonitor status updates upon Prometheus selection.
func testServiceMonitorStatusSubresource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.StatusForConfigurationResourcesFeature},
		},
	)
	require.NoError(t, err)

	name := "servicemonitor-status-subresource-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err, "failed to create Prometheus")
	smon := framework.MakeBasicServiceMonitor(name)

	sm, err := framework.MonClientV1.ServiceMonitors(ns).Create(ctx, smon, v1.CreateOptions{})
	require.NoError(t, err)

	sm, err = framework.WaitForServiceMonitorAcceptedCondition(ctx, sm, p, monitoringv1.PrometheusName, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	sm.Spec.Endpoints[0].BasicAuth = &monitoringv1.BasicAuth{
		Username: corev1.SecretKeySelector{
			Key: "username",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
	sm, err = framework.MonClientV1.ServiceMonitors(ns).Update(ctx, sm, v1.UpdateOptions{})
	require.NoError(t, err)
	_, err = framework.WaitForServiceMonitorAcceptedCondition(ctx, sm, p, monitoringv1.PrometheusName, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)
}
