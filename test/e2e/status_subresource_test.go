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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// testPodMonitorStatusSubresource validates PodMonitor status updates upon Prometheus selection.
func testPodMonitorStatusSubresource(t *testing.T) {
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

	name := "podmonitor-status-subresource-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	// Create a first podmonitor to check that the operator only updates the binding when needed.
	pm1 := framework.MakeBasicPodMonitor("pmon1")
	pm1.Labels["group"] = name
	pm1, err = framework.MonClientV1.PodMonitors(ns).Create(ctx, pm1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	pm1, err = framework.WaitForPodMonitorCondition(ctx, pm1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(pm1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second podmonitor to check that the operator updates the binding when the condition changes.
	pm2 := framework.MakeBasicPodMonitor("pmon2")
	pm2.Labels["group"] = name
	pm2, err = framework.MonClientV1.PodMonitors(ns).Create(ctx, pm2, v1.CreateOptions{})
	require.NoError(t, err)

	pm2, err = framework.WaitForPodMonitorCondition(ctx, pm2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first podmonitor. A label update doesn't
	// change the status of the podmonitor and the observed timetstamp
	// should be the same as before.
	pm1.Labels["test"] = "test"
	pm1, err = framework.MonClientV1.PodMonitors(ns).Update(ctx, pm1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second podmonitor to reference an non-existing Secret.
	pm2.Spec.PodMetricsEndpoints[0].BasicAuth = &monitoringv1.BasicAuth{
		Username: corev1.SecretKeySelector{
			Key: "username",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
	pm2, err = framework.MonClientV1.PodMonitors(ns).Update(ctx, pm2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second PodMonitor should change to Accepted=False.
	_, err = framework.WaitForPodMonitorCondition(ctx, pm2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first PodMonitor should remain unchanged.
	pm1, err = framework.WaitForPodMonitorCondition(ctx, pm1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(pm1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testProbeStatusSubresource validates Probe status updates upon Prometheus selection.
func testProbeStatusSubresource(t *testing.T) {
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

	name := "probe-status-subresource-test"
	svc := framework.MakePrometheusService(name, name, corev1.ServiceTypeClusterIP)

	proberURL := "localhost:9115"
	targets := []string{svc.Name + ":9090"}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ProbeSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	if finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(ctx, ns, svc); err != nil {
		require.NoError(t, fmt.Errorf("creating prometheus service failed: %w", err))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	probe1 := framework.MakeBasicStaticProbe("probe1", proberURL, targets)
	probe1.Labels["group"] = name
	probe1, err = framework.MonClientV1.Probes(ns).Create(ctx, probe1, metav1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	probe1, err = framework.WaitForProbeCondition(ctx, probe1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(probe1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second probe to check that the operator updates the binding when the condition changes.
	probe2 := framework.MakeBasicStaticProbe("probe2", proberURL, targets)
	probe2.Labels["group"] = name
	probe2, err = framework.MonClientV1.Probes(ns).Create(ctx, probe2, metav1.CreateOptions{})
	require.NoError(t, err)

	probe2, err = framework.WaitForProbeCondition(ctx, probe2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first probe. A label update doesn't
	// change the status of the probe and the observed timetstamp
	// should be the same as before.
	probe1.Labels["test"] = "test"
	probe1, err = framework.MonClientV1.Probes(ns).Update(ctx, probe1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second probe to reference an non-existing Secret.
	probe2.Spec.BasicAuth = &monitoringv1.BasicAuth{
		Username: corev1.SecretKeySelector{
			Key: "username",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
	probe2, err = framework.MonClientV1.Probes(ns).Update(ctx, probe2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second Probe should change to Accepted=False.
	_, err = framework.WaitForProbeCondition(ctx, probe2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first Probe should remain unchanged.
	probe1, err = framework.WaitForProbeCondition(ctx, probe1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(probe1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testGarbageCollectionOfPodMonitorBinding validates that the operator removes the reference to the Prometheus resource when the PodMonitor isn't selected anymore by the workload.
func testGarbageCollectionOfPodMonitorBinding(t *testing.T) {
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

	name := "pmon-status-binding-cleanup-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	pm := framework.MakeBasicPodMonitor(name)
	pm, err = framework.MonClientV1.PodMonitors(ns).Create(ctx, pm, v1.CreateOptions{})
	require.NoError(t, err)

	pm, err = framework.WaitForPodMonitorCondition(ctx, pm, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the PodMonitor's labels, Prometheus doesn't select the resource anymore.
	pm.Labels = map[string]string{}
	pm, err = framework.MonClientV1.PodMonitors(ns).Update(ctx, pm, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForPodMonitorWorkloadBindingCleanup(ctx, pm, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testRmPodMonitorBindingDuringWorkloadDelete validates that the operator removes the reference to the Prometheus resource from PodMonitor's status when workload is deleted.
func testRmPodMonitorBindingDuringWorkloadDelete(t *testing.T) {
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

	name := "workload-del-pmon-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err, "failed to create Prometheus")
	pmon := framework.MakeBasicPodMonitor(name)

	pm, err := framework.MonClientV1.PodMonitors(ns).Create(ctx, pmon, v1.CreateOptions{})
	require.NoError(t, err)

	pm, err = framework.WaitForPodMonitorCondition(ctx, pm, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForPodMonitorWorkloadBindingCleanup(ctx, pm, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testScrapeConfigStatusSubresource validates ScrapeConfig status updates upon Prometheus selection.
func testScrapeConfigStatusSubresource(t *testing.T) {
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

	name := "scfg-status-subresource-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	// Create a first scrapeConfig to check that the operator only updates the binding when needed.
	sc1 := framework.MakeBasicScrapeConfig(ns, "sc1")
	sc1.Labels["group"] = name
	sc1, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(ctx, sc1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	sc1, err = framework.WaitForScrapeConfigCondition(ctx, sc1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(sc1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second scrapeConfig to check that the operator updates the binding when the condition changes.
	sc2 := framework.MakeBasicScrapeConfig(ns, "sc2")
	sc2.Labels["group"] = name
	sc2, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(ctx, sc2, v1.CreateOptions{})
	require.NoError(t, err)

	sc2, err = framework.WaitForScrapeConfigCondition(ctx, sc2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first scrapeConfig. A label update doesn't
	// change the status of the scrapeConfig and the observed timetstamp
	// should be the same as before.
	sc1.Labels["test"] = "test"
	sc1, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Update(ctx, sc1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second scrapeConfig to reference an non-existing Secret.
	sc2.Spec.BasicAuth = &monitoringv1.BasicAuth{
		Username: corev1.SecretKeySelector{
			Key: "username",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
	sc2, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Update(ctx, sc2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second ScrapeConfig should change to Accepted=False.
	_, err = framework.WaitForScrapeConfigCondition(ctx, sc2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first ScrapeConfig should remain unchanged.
	sc1, err = framework.WaitForScrapeConfigCondition(ctx, sc1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(sc1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testGarbageCollectionOfScrapeConfigBinding validates that the operator removes the reference to the Prometheus resource when the ScrapeConfig isn't selected anymore by the workload.
func testGarbageCollectionOfScrapeConfigBinding(t *testing.T) {
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

	name := "scfg-status-binding-cleanup-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	sc := framework.MakeBasicScrapeConfig(ns, name)
	sc.Labels["group"] = name

	sc, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(ctx, sc, v1.CreateOptions{})
	require.NoError(t, err)

	sc, err = framework.WaitForScrapeConfigCondition(ctx, sc, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the ScrapeConfig's labels, Prometheus doesn't select the resource anymore.
	sc.Labels = map[string]string{}
	sc, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Update(ctx, sc, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForScrapeConfigWorkloadBindingCleanup(ctx, sc, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testRmScrapeConfigBindingDuringWorkloadDelete validates that the operator removes the reference to the Prometheus resource when workload is deleted.
func testRmScrapeConfigBindingDuringWorkloadDelete(t *testing.T) {
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

	name := "workload-del-scfg-test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err, "failed to create Prometheus")

	sc := framework.MakeBasicScrapeConfig(ns, name)
	sc.Labels["group"] = name

	sc, err = framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(ctx, sc, v1.CreateOptions{})
	require.NoError(t, err)

	sc, err = framework.WaitForScrapeConfigCondition(ctx, sc, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForScrapeConfigWorkloadBindingCleanup(ctx, sc, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testFinalizerForPromAgentWhenStatusForConfigResEnabled tests the adding/removing of status-cleanup finalizer for PrometheusAgent when StatusForConfigurationResourcesFeature is enabled.
func testFinalizerForPromAgentWhenStatusForConfigResEnabled(t *testing.T) {
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

	p := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	p, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	finalizers := p.GetFinalizers()
	require.NotEmpty(t, finalizers)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)
}
