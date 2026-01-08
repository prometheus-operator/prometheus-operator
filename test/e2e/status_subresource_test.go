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
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
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

// testGarbageCollectionOfProbeBinding validates that the operator removes the reference to the Prometheus resource when the Probe isn't selected anymore by the workload.
func testGarbageCollectionOfProbeBinding(t *testing.T) {
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

	name := "probe-status-binding-cleanup-test"
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

	probe := framework.MakeBasicStaticProbe("probe", proberURL, targets)
	probe.Labels["group"] = name
	probe, err = framework.MonClientV1.Probes(ns).Create(ctx, probe, metav1.CreateOptions{})
	require.NoError(t, err)

	probe, err = framework.WaitForProbeCondition(ctx, probe, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the Probe's labels, Prometheus doesn't select the resource anymore.
	probe.Labels = map[string]string{}
	probe, err = framework.MonClientV1.Probes(ns).Update(ctx, probe, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForProbeWorkloadBindingCleanup(ctx, probe, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testRmProbeBindingDuringWorkloadDelete validates that the operator removes the reference to the Prometheus resource from Probe's status when workload is deleted.
func testRmProbeBindingDuringWorkloadDelete(t *testing.T) {
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

	name := "workload-del-probe-test"
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

	probe := framework.MakeBasicStaticProbe("probe", proberURL, targets)
	probe.Labels["group"] = name
	probe, err = framework.MonClientV1.Probes(ns).Create(ctx, probe, metav1.CreateOptions{})
	require.NoError(t, err)

	probe, err = framework.WaitForProbeCondition(ctx, probe, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForProbeWorkloadBindingCleanup(ctx, probe, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testPrometheusRuleStatusSubresource validates PrometheusRule status updates upon Prometheus selection.
func testPrometheusRuleStatusSubresource(t *testing.T) {
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

	name := "prometheusrule-status-subresource-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	// Create a first PrometheusRule to check that the operator only updates the binding when needed.
	pr1 := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr1.Labels["group"] = name
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	pr1, err = framework.WaitForRuleCondition(ctx, pr1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(pr1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second PrometheusRule to check that the operator updates the binding when the condition changes.
	pr2 := framework.MakeBasicRule(ns, "rule2", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert2",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert2",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr2.Labels["group"] = name
	pr2, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr2, v1.CreateOptions{})
	require.NoError(t, err)

	pr2, err = framework.WaitForRuleCondition(ctx, pr2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first PrometheusRule. A label update doesn't
	// change the status of the PrometheusRule and the observed timestamp
	// should be the same as before.
	pr1.Labels["test"] = "test"
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second PrometheusRule to have an invalid rule expression.
	pr2.Spec.Groups[0].Rules = append(pr2.Spec.Groups[0].Rules, monitoringv1.Rule{
		Record: "test:invalid",
		Expr:   intstr.FromString("invalid_expr{"),
	})
	pr2, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second PrometheusRule should change to Accepted=False.
	_, err = framework.WaitForRuleCondition(ctx, pr2, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first PrometheusRule should remain unchanged.
	pr1, err = framework.WaitForRuleCondition(ctx, pr1, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(pr1.Status.Bindings, p, monitoringv1.PrometheusName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testGarbageCollectionOfPrometheusRuleBinding validates that the operator removes the reference to the Prometheus resource when the PrometheusRule isn't selected anymore by the workload.
func testGarbageCollectionOfPrometheusRuleBinding(t *testing.T) {
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

	name := "prom-rule-status-binding-cleanup-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	pr1 := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr1.Labels["group"] = name
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr1, v1.CreateOptions{})
	require.NoError(t, err)

	pr1.Labels = map[string]string{}
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr1, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForRuleWorkloadBindingCleanup(ctx, pr1, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testRmPrometheusRuleBindingDuringWorkloadDelete validates that the operator removes the reference to the Prometheus resource when workload is deleted.
func testRmPrometheusRuleBindingDuringWorkloadDelete(t *testing.T) {
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

	name := "prom-rule-status-binding-cleanup-test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	pr := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr.Labels["group"] = name
	pr, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr, v1.CreateOptions{})
	require.NoError(t, err)

	pr, err = framework.WaitForRuleCondition(ctx, pr, p, monitoringv1.PrometheusName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForRuleWorkloadBindingCleanup(ctx, pr, p, monitoringv1.PrometheusName, 1*time.Minute)
	require.NoError(t, err)
}

// testFinalizerForThanosRulerWhenStatusForConfigResEnabled tests the adding/removing of status-cleanup finalizer for ThanosRuler when StatusForConfigurationResourcesFeature is enabled.
func testFinalizerForThanosRulerWhenStatusForConfigResEnabled(t *testing.T) {
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

	name := "tr-status-finalizer"

	tr := framework.MakeBasicThanosRuler(name, 1, name)
	_, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	require.NoError(t, err)

	ruler, err := framework.MonClientV1.ThanosRulers(ns).Get(ctx, name, v1.GetOptions{})
	require.NoError(t, err)

	finalizers := ruler.GetFinalizers()
	require.NotEmpty(t, finalizers)

	err = framework.DeleteThanosRulerAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)
}

// testPrometheusRuleStatusSubresourceForThanosRuler validates PrometheusRule status updates upon ThanosRuler selection.
func testPrometheusRuleStatusSubresourceForThanosRuler(t *testing.T) {
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

	name := "promrule-status-subres-tr-test"

	tr := framework.MakeBasicThanosRuler(name, 1, name)
	tr.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	tr, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	require.NoError(t, err)

	// Create a first PrometheusRule to check that the operator only updates the binding when needed.
	pr1 := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr1.Labels["group"] = name
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	pr1, err = framework.WaitForRuleCondition(ctx, pr1, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 3*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(pr1.Status.Bindings, tr, monitoringv1.ThanosRulerName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second PrometheusRule to check that the operator updates the binding when the condition changes.
	pr2 := framework.MakeBasicRule(ns, "rule2", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert2",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert2",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr2.Labels["group"] = name
	pr2, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr2, v1.CreateOptions{})
	require.NoError(t, err)

	pr2, err = framework.WaitForRuleCondition(ctx, pr2, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first PrometheusRule. A label update doesn't
	// change the status of the PrometheusRule and the observed timestamp
	// should be the same as before.
	pr1.Labels["test"] = "test"
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second PrometheusRule to have an invalid rule expression.
	pr2.Spec.Groups[0].Rules = append(pr2.Spec.Groups[0].Rules, monitoringv1.Rule{
		Record: "test:invalid",
		Expr:   intstr.FromString("invalid_expr{"),
	})
	pr2, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second PrometheusRule should change to Accepted=False.
	_, err = framework.WaitForRuleCondition(ctx, pr2, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first PrometheusRule should remain unchanged.
	pr1, err = framework.WaitForRuleCondition(ctx, pr1, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(pr1.Status.Bindings, tr, monitoringv1.ThanosRulerName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}

// testGarbageCollectionOfPromRuleBindingForThanosRuler validates that the operator removes the reference to the thanoousRuler resource when the PrometheusRule isn't selected anymore by the workload.
func testGarbageCollectionOfPromRuleBindingForThanosRuler(t *testing.T) {
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

	name := "prom-rule-status-binding-cleanup-tr"

	tr := framework.MakeBasicThanosRuler(name, 1, name)
	tr.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	tr, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	require.NoError(t, err)

	pr1 := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr1.Labels["group"] = name
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr1, v1.CreateOptions{})
	require.NoError(t, err)

	pr1, err = framework.WaitForRuleCondition(ctx, pr1, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	pr1.Labels = map[string]string{}
	pr1, err = framework.MonClientV1.PrometheusRules(ns).Update(ctx, pr1, v1.UpdateOptions{})
	require.NoError(t, err)

	_, err = framework.WaitForRuleWorkloadBindingCleanup(ctx, pr1, tr, monitoringv1.ThanosRulerName, 1*time.Minute)
	require.NoError(t, err)
}

// testRmPromeRuleBindingDuringWorkloadDeleteForThanosRuler validates that the operator removes the reference to the ThanosRuler resource when workload is deleted.
func testRmPromeRuleBindingDuringWorkloadDeleteForThanosRuler(t *testing.T) {
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

	name := "prom-rule-status-binding-cleanup-tr"

	tr := framework.MakeBasicThanosRuler(name, 1, name)
	tr.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	tr, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	require.NoError(t, err)

	pr := framework.MakeBasicRule(ns, "rule1", []monitoringv1.RuleGroup{
		{
			Name: "TestAlert1",
			Rules: []monitoringv1.Rule{
				{
					Alert: "TestAlert1",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	})
	pr.Labels["group"] = name
	pr, err = framework.MonClientV1.PrometheusRules(ns).Create(ctx, pr, v1.CreateOptions{})
	require.NoError(t, err)

	pr, err = framework.WaitForRuleCondition(ctx, pr, tr, monitoringv1.ThanosRulerName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 3*time.Minute)
	require.NoError(t, err)

	err = framework.DeleteThanosRulerAndWaitUntilGone(ctx, ns, name)
	require.NoError(t, err)

	_, err = framework.WaitForRuleWorkloadBindingCleanup(ctx, pr, tr, monitoringv1.ThanosRulerName, 1*time.Minute)
	require.NoError(t, err)
}

// testAlertmanagerConfigStatusSubresource validates AlertmanagerConfig status updates upon Alertmanager selection.
func testAlertmanagerConfigStatusSubresource(t *testing.T) {
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

	name := "am-cfg-status-subresource-test"

	am := framework.MakeBasicAlertmanager(ns, name, 1)
	am.Spec.AlertmanagerConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": name,
		},
	}

	_, err = framework.CreateAlertmanagerAndWaitUntilReady(ctx, am)
	require.NoError(t, err)

	// Create a first AlertmanagerConfig to check that the operator only updates the binding when needed.
	alc1 := &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "amcfg1",
			Namespace: ns,
			Labels:    map[string]string{},
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "default",
			},
			Receivers: []monitoringv1alpha1.Receiver{
				{
					Name: "default",
					WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
						{
							URL: func(s string) *monitoringv1alpha1.URL {
								u := monitoringv1alpha1.URL(s)
								return &u
							}("http://test.url"),
						},
					},
				},
			},
		},
	}
	alc1.Labels["group"] = name

	alc1, err = framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(ctx, alc1, v1.CreateOptions{})
	require.NoError(t, err)

	// Record the lastTransitionTime value.
	alc1, err = framework.WaitForAlertmanagerConfigCondition(ctx, alc1, am, monitoringv1alpha1.AlertmanagerConfigName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err := framework.GetWorkloadBinding(alc1.Status.Bindings, am, monitoringv1alpha1.AlertmanagerConfigName)
	require.NoError(t, err)
	cond, err := framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	ts := cond.LastTransitionTime.String()
	require.NotEqual(t, "", ts)

	// Create a second AlertmanagerConfig to check that the operator updates the binding when the condition changes.
	alc2 := &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "amcfg2",
			Namespace: ns,
			Labels:    map[string]string{},
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "default",
			},
			Receivers: []monitoringv1alpha1.Receiver{
				{
					Name: "default",
					WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
						{
							URL: func(s string) *monitoringv1alpha1.URL {
								u := monitoringv1alpha1.URL(s)
								return &u
							}("http://test.url"),
						},
					},
				},
			},
		},
	}
	alc2.Labels["group"] = name
	alc2, err = framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(ctx, alc2, v1.CreateOptions{})
	require.NoError(t, err)

	alc2, err = framework.WaitForAlertmanagerConfigCondition(ctx, alc2, am, monitoringv1alpha1.AlertmanagerConfigName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)

	// Update the labels of the first AlertmanagerConfig. A label update doesn't
	// change the status of the AlertmanagerConfig and the observed timestamp
	// should be the same as before.
	alc1.Labels["test"] = "test"
	alc1, err = framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Update(ctx, alc1, v1.UpdateOptions{})
	require.NoError(t, err)

	// Update the second AlertmanagerConfig to have an invalid rule expression.
	invalidURL := monitoringv1alpha1.URL("//invalid-url")
	alc2.Spec.Receivers[0].WebhookConfigs[0].URL = &invalidURL
	alc2, err = framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Update(ctx, alc2, v1.UpdateOptions{})
	require.NoError(t, err)

	// The second AlertmanagerConfig should change to Accepted=False.
	_, err = framework.WaitForAlertmanagerConfigCondition(ctx, alc2, am, monitoringv1alpha1.AlertmanagerConfigName, monitoringv1.Accepted, monitoringv1.ConditionFalse, 1*time.Minute)
	require.NoError(t, err)

	// The first AlertmanagerConfig should remain unchanged.
	alc1, err = framework.WaitForAlertmanagerConfigCondition(ctx, alc1, am, monitoringv1alpha1.AlertmanagerConfigName, monitoringv1.Accepted, monitoringv1.ConditionTrue, 1*time.Minute)
	require.NoError(t, err)
	binding, err = framework.GetWorkloadBinding(alc1.Status.Bindings, am, monitoringv1alpha1.AlertmanagerConfigName)
	require.NoError(t, err)
	cond, err = framework.GetConfigResourceCondition(binding.Conditions, monitoringv1.Accepted)
	require.NoError(t, err)
	require.Equal(t, ts, cond.LastTransitionTime.String())
}
