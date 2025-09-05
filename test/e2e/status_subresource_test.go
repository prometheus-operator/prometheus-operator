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
	p, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	finalizers := p.GetFinalizers()
	require.NotEmpty(t, finalizers)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Create a first service monitor to check that the operator only updates the binding when needed.
	sm1 := framework.MakeBasicServiceMonitor("smon1")
	sm1.Labels["group"] = name
	sm1, err = framework.MonClientV1.ServiceMonitors(ns).Create(ctx, sm1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	sm1, err = framework.WaitForServiceMonitorCondition(ctx, sm1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(sm1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second service monitor to check that the operator updates the binding when the condition changes.
	sm2 := framework.MakeBasicServiceMonitor("smon2")
	sm2.Labels["group"] = name
	sm2, err = framework.MonClientV1.ServiceMonitors(ns).Create(ctx, sm2, v1.CreateOptions{})
	require.NoError(t, err)

	sm2, err = framework.WaitForServiceMonitorCondition(ctx, sm2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first service monitor. A label update doesn't
	// change the status of the service monitor and the observed timetstamp
	// should be the same as before.
	sm1.Labels["test"] = "test"
	sm1, err = framework.MonClientV1.ServiceMonitors(ns).Update(ctx, sm1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second service monitor to reference an non-existing Secret.
	sm2.Spec.Endpoints[0].BasicAuth = &monitoringv1.BasicAuth{
		Username: corev1.SecretKeySelector{
			Key: "username",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
	sm2, err = framework.MonClientV1.ServiceMonitors(ns).Update(ctx, sm2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second ServiceMonitor should change to Accepted=False.
	_, err = framework.WaitForServiceMonitorCondition(ctx, sm2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first ServiceMonitor should remain unchanged.
	sm1, err = framework.WaitForServiceMonitorCondition(ctx, sm1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(sm1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testGarbageCollectionOfServiceMonitorBinding validates that the operator removes the reference to the Prometheus resource when the ServiceMonitor isn't selected anymore by the workload.
func testGarbageCollectionOfServiceMonitorBinding(t *testing.T) {
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

	name := "smon-status-binding-cleanup-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	sm := framework.MakeBasicServiceMonitor(name)
	sm, err = framework.MonClientV1.ServiceMonitors(ns).Create(ctx, sm, v1.CreateOptions{})
	require.NoError(t, err)

	sm, err = framework.WaitForServiceMonitorCondition(ctx, sm, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the ServiceMonitor's labels, Prometheus doesn't select the resource anymore.
	sm.Labels = map[string]string{}
	sm, err = framework.MonClientV1.ServiceMonitors(ns).Update(ctx, sm, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForServiceMonitorWorkloadBindingCleanup(ctx, sm, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

func testServiceMonitorStatusWithMultipleWorkloads(t *testing.T) {
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

	name := "servicemonitor-status-multiple-workloads"
	p1 := framework.MakeBasicPrometheus(ns, "server1", name, 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p1)
	require.NoError(t, err)

	p2 := framework.MakeBasicPrometheus(ns, "server2", name, 1)
	// Forbid access to the container's filesystem.
	p2.Spec.ArbitraryFSAccessThroughSMs.Deny = true
	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p2)
	require.NoError(t, err)

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints[0].BearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	sm, err = framework.MonClientV1.ServiceMonitors(ns).Create(ctx, sm, v1.CreateOptions{})
	require.NoError(t, err)

	// The ServiceMonitor should be accepted by the Prometheus "server1" resource.
	_, err = framework.WaitForServiceMonitorCondition(ctx, sm, p1, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// ServiceMonitor should be rejected by the Prometheus "server2" resource because it wants to access the SA token file.
	_, err = framework.WaitForServiceMonitorCondition(ctx, sm, p2, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)
}

// testRmServiceMonitorBindingDuringWorkloadDelete validates that the operator removes the reference to the Prometheus resource when workload is deleted.
func testRmServiceMonitorBindingDuringWorkloadDelete(t *testing.T) {
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

	name := "workload-del-smon-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err, "failed to create Prometheus")
	smon := framework.MakeBasicServiceMonitor(name)

	sm, err := framework.MonClientV1.ServiceMonitors(ns).Create(ctx, smon, v1.CreateOptions{})
	require.NoError(t, err)

	sm, err = framework.WaitForServiceMonitorCondition(ctx, sm, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForServiceMonitorWorkloadBindingCleanup(ctx, sm, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}
