// Copyright 2019 The prometheus-operator Authors
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
	v1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testPrometheusInstanceNamespacesAllNs(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	nonInstanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		operatorNs,
		nil,
		nil,
		[]string{instanceNs},
		nil,
		false,
		true, // clusterrole
		true,
	)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(nonInstanceNs, "non-instance", "non-instance", 1)
	_, err = framework.MonClientV1.Prometheuses(nonInstanceNs).Create(context.Background(), p, metav1.CreateOptions{})
	require.NoError(t, err, "creating %v Prometheus instances failed (%v): %v", p.Spec.Replicas, p.Name, err)

	p = framework.MakeBasicPrometheus(instanceNs, "instance", "instance", 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), instanceNs, p)
	require.NoError(t, err)

	// this is not ideal, as we cannot really find out if prometheus operator did not reconcile the denied prometheus.
	// nevertheless it is very likely that it reconciled it as the allowed prometheus is up.
	sts, err := framework.KubeClient.AppsV1().StatefulSets(nonInstanceNs).Get(context.Background(), "prometheus-instance", metav1.GetOptions{})
	require.True(t, api_errors.IsNotFound(err), "expected not to find a Prometheus statefulset, but did: %v/%v", sts.Namespace, sts.Name)
}

func testPrometheusInstanceNamespacesDenyList(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// create three namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --prometheus-instance-namespaces="instance"
	//   - will additionally be configured on prometheus operator as --deny-namespaces="instance"
	//   - hosts a service monitor CR which must NOT be reconciled.
	//   - hosts a prometheus CR which must be reconciled.
	//     This prometheus instance must pick up targets (service monitors)
	//     in the "allowed" namespace.
	//
	// 3. "denied" ns:
	//   - will be configured on prometheus operator as --deny-namespaces="denied"
	//   - hosts a service monitor CR which must NOT be reconciled
	//   - hosts a prometheus CR which must NOT be reconciled
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	deniedNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	for _, ns := range []string{deniedNs, instanceNs} {
		err := framework.AddLabelsToNamespace(context.Background(), ns, map[string]string{
			"monitored": "true",
		})
		require.NoError(t, err)
	}

	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		operatorNs,
		nil,
		[]string{deniedNs, instanceNs},
		[]string{instanceNs},
		nil,
		false,
		true, // clusterrole
		true,
	)
	require.NoError(t, err)

	{
		// create Prometheus custom resources in "denied" namespaces.
		// This must NOT be reconciled as the prometheus-instance-namespaces option points to somewhere else.
		p := framework.MakeBasicPrometheus(deniedNs, "denied", "denied", 1)
		_, err = framework.MonClientV1.Prometheuses(deniedNs).Create(context.Background(), p, metav1.CreateOptions{})
		require.NoError(t, err, "creating %v Prometheus instances failed (%v): %v", p.Spec.Replicas, p.Name, err)

		// create a simple echo server in the "denied" namespace,
		// expose a service pointing to it,
		// and create a service monitor pointing to that service.
		// Wait, until that service appears as a target in the "instance" Prometheus.
		echo := framework.MakeEchoDeployment("denied")

		err = framework.CreateDeployment(context.Background(), deniedNs, echo)
		require.NoError(t, err)

		svc := framework.MakeEchoService("denied", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), deniedNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		s := framework.MakeBasicServiceMonitor("monitored")
		_, err = framework.MonClientV1.ServiceMonitors(deniedNs).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	// create Prometheus custom resource in the "instance" namespace.
	// This one must be reconciled.
	// Let this Prometheus custom resource match service monitors in namespaces having the label `"group": "monitored"`.
	// This will match the service monitors created in the "denied" namespace.
	// Also create a service monitor in this namespace. This one must not be reconciled.
	// Expose the created Prometheus service.
	{
		p := framework.MakeBasicPrometheus(instanceNs, "instance", "instance", 1)

		p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"monitored": "true",
			},
		}

		p.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"group": "monitored",
			},
		}

		s := framework.MakeBasicServiceMonitor("monitored")
		_, err = framework.MonClientV1.ServiceMonitors(instanceNs).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)

		// create the prometheus service and wait until it is ready
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), instanceNs, p)
		require.NoError(t, err)

		svc := framework.MakePrometheusService("instance", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), instanceNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)
	}

	// this is not ideal, as we cannot really find out if prometheus operator did not reconcile the denied prometheus.
	// nevertheless it is very likely that it reconciled it as the allowed prometheus is up.
	sts, err := framework.KubeClient.AppsV1().StatefulSets(deniedNs).Get(context.Background(), "prometheus-instance", metav1.GetOptions{})
	require.True(t, api_errors.IsNotFound(err), "expected not to find a Prometheus statefulset, but did: %v/%v", sts.Namespace, sts.Name)

	err = framework.WaitForActiveTargets(context.Background(), instanceNs, "prometheus-instance", 0)
	require.NoError(t, err)
}

func testPrometheusInstanceNamespacesAllowList(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// create three namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --prometheus-instance-namespaces="instance"
	//   - hosts a service monitor CR which must NOT be reconciled.
	//   - hosts a prometheus CR which must be reconciled.
	//     This prometheus instance must pick up targets (service monitors)
	//     in the "allowed" namespace.
	//
	// 3. "allowed" ns:
	//   - will be configured on prometheus operator as --namespaces="allowed"
	//   - hosts a service monitor CR which must be reconciled
	//   - hosts a prometheus CR which must NOT be reconciled
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	for _, ns := range []string{allowedNs, instanceNs} {
		err := framework.AddLabelsToNamespace(context.Background(), ns, map[string]string{
			"monitored": "true",
		})
		require.NoError(t, err)
	}

	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		operatorNs,
		[]string{allowedNs},
		nil,
		[]string{instanceNs},
		nil,
		false,
		false, // not clusterrole
		true)
	require.NoError(t, err)

	// create Prometheus custom resources in "allowed" namespaces.
	// This must NOT be reconciled as the prometheus-instance-namespaces option points to somewhere else.
	p := framework.MakeBasicPrometheus(allowedNs, "allowed", "allowed", 1)
	_, err = framework.MonClientV1.Prometheuses(allowedNs).Create(context.Background(), p, metav1.CreateOptions{})
	require.NoError(t, err, "creating %v Prometheus instances failed (%v): %v", p.Spec.Replicas, p.Name, err)

	// create Prometheus custom resource in the "instance" namespace.
	// This one must be reconciled.
	// Let this Prometheus custom resource match service monitors in namespaces having the label `"monitored": "true"`.
	// This will match the service monitors created in the "allowed" namespace.
	// Also create a service monitor in this namespace. This one must not be reconciled.
	// Expose the created Prometheus service.
	{
		p = framework.MakeBasicPrometheus(instanceNs, "instance", "instance", 1)

		p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"monitored": "true",
			},
		}

		p.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"group": "monitored",
			},
		}

		// create the prometheus service and wait until it is ready
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), instanceNs, p)
		require.NoError(t, err)

		svc := framework.MakePrometheusService("instance", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), instanceNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		s := framework.MakeBasicServiceMonitor("monitored")
		_, err = framework.MonClientV1.ServiceMonitors(instanceNs).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	{
		// create a simple echo server in the "allowed" namespace,
		// expose a service pointing to it,
		// and create a service monitor pointing to that service.
		// Wait, until that service appears as a target in the "instance" Prometheus.
		echo := framework.MakeEchoDeployment("allowed")

		err = framework.CreateDeployment(context.Background(), allowedNs, echo)
		require.NoError(t, err)

		svc := framework.MakeEchoService("allowed", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), allowedNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		s := framework.MakeBasicServiceMonitor("monitored")
		_, err = framework.MonClientV1.ServiceMonitors(allowedNs).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)

		err = framework.WaitForActiveTargets(context.Background(), instanceNs, "prometheus-instance", 1)
		require.NoError(t, err)

		// Remove the selecting label on the "allowed" namespace and check that
		// the target is removed.
		// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
		err = framework.RemoveLabelsFromNamespace(context.Background(), allowedNs, "monitored")
		require.NoError(t, err)

		err = framework.WaitForActiveTargets(context.Background(), instanceNs, "prometheus-instance", 0)
		require.NoError(t, err)
	}

	// this is not ideal, as we cannot really find out if prometheus operator did not reconcile the denied prometheus.
	// nevertheless it is very likely that it reconciled it as the allowed prometheus is up.
	sts, err := framework.KubeClient.AppsV1().StatefulSets(allowedNs).Get(context.Background(), "prometheus-instance", metav1.GetOptions{})
	require.True(t, api_errors.IsNotFound(err), "expected not to find a Prometheus statefulset, but did: %v/%v", sts.Namespace, sts.Name)

	// assert that no prometheus target points to the "instance" namespace
	targets, err := framework.GetActiveTargets(context.Background(), instanceNs, "prometheus-instance")
	require.NoError(t, err)

	for _, target := range targets {
		for k, v := range target.Labels {
			if k == "namespace" && v == instanceNs {
				t.Fatalf("expected namespace %s not be to have reconciled a service monitor but it has", instanceNs)
			}
		}
	}
}

// testPrometheusInstanceNamespacesNamespaceNotFound verifies that the
// operator can reconcile Prometheus and associated resources even when
// it's configured to watch namespaces that don't exist.
// See https://github.com/prometheus-operator/prometheus-operator/issues/3347
func testPrometheusInstanceNamespacesNamespaceNotFound(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// create three namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --prometheus-instance-namespaces="instance"
	//   - hosts a prometheus CR which must be reconciled.
	//     This prometheus instance must pick up targets (service monitors)
	//     in the "allowed" namespace.
	//
	// 3. "allowed" ns:
	//   - will be configured on prometheus operator as --namespaces="allowed"
	//   - hosts a service monitor CR which must be reconciled
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	for _, ns := range []string{allowedNs, instanceNs} {
		err := framework.AddLabelsToNamespace(context.Background(), ns, map[string]string{
			"monitored": "true",
		})
		require.NoError(t, err)
	}

	// Configure the operator to watch also a non-existing namespace (e.g. "notfound").
	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		operatorNs,
		[]string{"notfound", allowedNs},
		nil,
		[]string{"notfound", instanceNs},
		nil,
		false,
		true, // clusterrole
		true,
	)
	require.NoError(t, err)

	// Create Prometheus custom resource in the "instance" namespace.
	// Let this Prometheus custom resource match service monitors in namespaces having the label `"monitored": "true"`.
	// This will match the service monitors created in the "allowed" namespace.
	// Expose the created Prometheus service.
	{
		p := framework.MakeBasicPrometheus(instanceNs, "instance", "instance", 1)

		p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"monitored": "true",
			},
		}

		p.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"group": "monitored",
			},
		}

		// Create the prometheus service and wait until it is ready.
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), instanceNs, p)
		require.NoError(t, err)

		svc := framework.MakePrometheusService("instance", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), instanceNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)
	}

	{
		// Create a simple echo server in the "allowed" namespace,
		// Expose a service pointing to it, and create a service monitor
		// pointing to that service.
		// Wait, until that service appears as a target in the "instance" Prometheus.
		echo := framework.MakeEchoDeployment("allowed")

		err = framework.CreateDeployment(context.Background(), allowedNs, echo)
		require.NoError(t, err)

		svc := framework.MakeEchoService("allowed", "monitored", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), allowedNs, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		s := framework.MakeBasicServiceMonitor("monitored")
		_, err = framework.MonClientV1.ServiceMonitors(allowedNs).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)

		err = framework.WaitForActiveTargets(context.Background(), instanceNs, "prometheus-instance", 1)
		require.NoError(t, err)
	}
}
