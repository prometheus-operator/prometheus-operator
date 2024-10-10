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

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testDenyPrometheus(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	operatorNamespace := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}
	deniedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, operatorNamespace)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNamespace, nil, deniedNamespaces, nil, nil, false, true, true)
	require.NoError(t, err)

	for _, denied := range deniedNamespaces {
		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, denied)
		p := framework.MakeBasicPrometheus(denied, "denied", "denied", 1)
		_, err = framework.MonClientV1.Prometheuses(denied).Create(context.Background(), p, metav1.CreateOptions{})
		require.NoError(t, err, "creating %v Prometheus instances failed (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	for _, allowed := range allowedNamespaces {
		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, allowed)
		p := framework.MakeBasicPrometheus(allowed, "allowed", "allowed", 1)
		_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), allowed, p)
		require.NoError(t, err)
	}

	for _, denied := range deniedNamespaces {
		// this is not ideal, as we cannot really find out if prometheus operator did not reconcile the denied prometheus.
		// nevertheless it is very likely that it reconciled it as the allowed prometheus is up.
		sts, err := framework.KubeClient.AppsV1().StatefulSets(denied).Get(context.Background(), "prometheus-denied", metav1.GetOptions{})
		require.True(t, api_errors.IsNotFound(err), "expected not to find a Prometheus statefulset, but did: %v/%v", sts.Namespace, sts.Name)
	}

	for _, allowed := range allowedNamespaces {
		err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), allowed, "allowed")
		require.NoError(t, err)
	}
}

func testDenyServiceMonitor(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	operatorNamespace := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}
	deniedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, operatorNamespace)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNamespace, nil, deniedNamespaces, nil, nil, false, true, true)
	require.NoError(t, err)

	for _, denied := range deniedNamespaces {
		echo := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ehoserver",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: proto.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"prometheus": "denied",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"prometheus": "denied",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "echoserver",
								Image: "k8s.gcr.io/echoserver:1.10",
								Ports: []v1.ContainerPort{
									{
										Name:          "web",
										ContainerPort: 8443,
									},
								},
							},
						},
					},
				},
			},
		}

		err = framework.CreateDeployment(context.Background(), denied, echo)
		require.NoError(t, err)

		svc := framework.MakePrometheusService("denied", "denied", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), denied, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		// create the service monitor in a way, that it matches the label selector used in the allowed namespace.
		s := framework.MakeBasicServiceMonitor("allowed")
		_, err = framework.MonClientV1.ServiceMonitors(denied).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	for _, allowed := range allowedNamespaces {
		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, allowed)
		p := framework.MakeBasicPrometheus(allowed, "allowed", "allowed", 1)
		_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), allowed, p)
		require.NoError(t, err)

		svc := framework.MakePrometheusService("allowed", "allowed", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), allowed, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		s := framework.MakeBasicServiceMonitor("allowed")
		_, err = framework.MonClientV1.ServiceMonitors(allowed).Create(context.Background(), s, metav1.CreateOptions{})
		require.NoError(t, err)

		err = framework.WaitForActiveTargets(context.Background(), allowed, svc.Name, 1)
		require.NoError(t, err)
	}

	// just iterate again, so we have a chance to catch a faulty reconciliation of denied namespaces.
	for _, allowed := range allowedNamespaces {
		targets, err := framework.GetActiveTargets(context.Background(), allowed, "prometheus-allowed")
		require.NoError(t, err)

		got := len(targets)
		require.Equal(t, 1, got)
	}

	for _, allowed := range allowedNamespaces {
		err = framework.MonClientV1.ServiceMonitors(allowed).Delete(context.Background(), "allowed", metav1.DeleteOptions{})
		require.NoError(t, err)

		err = framework.WaitForActiveTargets(context.Background(), allowed, "prometheus-allowed", 0)
		require.NoError(t, err)
	}
}

func testDenyThanosRuler(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	operatorNamespace := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}
	deniedNamespaces := []string{framework.CreateNamespace(context.Background(), t, testCtx), framework.CreateNamespace(context.Background(), t, testCtx)}

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, operatorNamespace)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNamespace, nil, deniedNamespaces, nil, nil, false, true, true)
	require.NoError(t, err)

	for _, denied := range deniedNamespaces {
		tr := framework.MakeBasicThanosRuler("denied", 1, "http://test.example.com")
		_, err = framework.MonClientV1.ThanosRulers(denied).Create(context.Background(), tr, metav1.CreateOptions{})
		require.NoError(t, err, "creating %v Prometheus instances failed (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	for _, allowed := range allowedNamespaces {
		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, allowed)

		_, err = framework.CreateThanosRulerAndWaitUntilReady(context.Background(), allowed, framework.MakeBasicThanosRuler("allowed", 1, "http://test.example.com"))
		require.NoError(t, err)
	}

	for _, denied := range deniedNamespaces {
		// this is not ideal, as we cannot really find out if prometheus operator did not reconcile the denied thanos ruler.
		// nevertheless it is very likely that it reconciled it as the allowed prometheus is up.
		sts, err := framework.KubeClient.AppsV1().StatefulSets(denied).Get(context.Background(), "thanosruler-denied", metav1.GetOptions{})
		require.True(t, api_errors.IsNotFound(err), "expected not to find a Prometheus statefulset, but did: %v/%v", sts.Namespace, sts.Name)
	}

	for _, allowed := range allowedNamespaces {
		err := framework.DeleteThanosRulerAndWaitUntilGone(context.Background(), allowed, "allowed")
		require.NoError(t, err)
	}
}
