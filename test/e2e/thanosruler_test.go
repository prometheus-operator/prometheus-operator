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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testThanosRulerCreateDeleteCluster(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	if _, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeleteThanosRulerAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
}

func testThanosRulerWithStatefulsetCreationFailure(t *testing.T) {
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)

	tr := framework.MakeBasicThanosRuler("test", 1, "")
	// Empty queryEndpoints and queryConfigFile prevent the controller from
	// creating the statefulset.
	tr.Spec.QueryEndpoints = []string{}

	_, err := framework.MonClientV1.ThanosRulers(ns).Create(ctx, tr, metav1.CreateOptions{})
	require.NoError(t, err)

	var loopError error
	err = wait.PollUntilContextTimeout(ctx, time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1.ThanosRulers(ns).Get(ctx, "test", metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err != nil {
			loopError = err
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Available, monitoringv1.ConditionFalse); err != nil {
			loopError = err
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		t.Fatalf("%v: %v", err, loopError)
	}

	require.NoError(t, framework.DeleteThanosRulerAndWaitUntilGone(ctx, ns, "test"))
}

func testThanosRulerPrometheusRuleInDifferentNamespace(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	thanosNamespace := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, thanosNamespace)

	name := "test"

	// Create a Prometheus resource because Thanos ruler needs a query API.
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), thanosNamespace, framework.MakeBasicPrometheus(thanosNamespace, name, name, 1))
	if err != nil {
		t.Fatal(err)
	}

	svc := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	if _, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), thanosNamespace, svc); err != nil {
		t.Fatal(err)
	}

	thanos := framework.MakeBasicThanosRuler(name, 1, fmt.Sprintf("http://%s:%d/", svc.Name, svc.Spec.Ports[0].Port))
	thanos.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"monitored": "true",
		},
	}
	thanos.Spec.EvaluationInterval = "1s"

	if _, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), thanosNamespace, thanos); err != nil {
		t.Fatal(err)
	}

	ruleNamespace := framework.CreateNamespace(context.Background(), t, testCtx)
	if err := framework.AddLabelsToNamespace(context.Background(), ruleNamespace, map[string]string{
		"monitored": "true",
	}); err != nil {
		t.Fatal(err)
	}

	const testAlert = "alert1"
	_, err = framework.MakeAndCreateFiringRule(context.Background(), ruleNamespace, "rule1", testAlert)
	if err != nil {
		t.Fatal(err)
	}

	thanosService := framework.MakeThanosRulerService(thanos.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), thanosNamespace, thanosService); err != nil {
		t.Fatalf("creating Thanos ruler service failed: %v", err)
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	if err := framework.WaitForThanosFiringAlert(context.Background(), thanosNamespace, thanosService.Name, testAlert); err != nil {
		t.Fatal(err)
	}

	// Remove the selecting label from ruleNamespace and wait until the rule is
	// removed from the Thanos ruler.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
	if err := framework.RemoveLabelsFromNamespace(context.Background(), ruleNamespace, "monitored"); err != nil {
		t.Fatal(err)
	}

	var loopError error
	err = wait.PollUntilContextTimeout(context.Background(), time.Second, 5*framework.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
		var firing bool
		firing, loopError = framework.CheckThanosFiringAlert(ctx, thanosNamespace, thanosService.Name, testAlert)
		return !firing, nil
	})

	if err != nil {
		t.Fatalf("waiting for alert %q to stop firing: %v: %v", testAlert, err, loopError)
	}
}

func testTRPreserveUserAddedMetadata(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	thanosRuler := framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	thanosRuler, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanosRuler)
	if err != nil {
		t.Fatal(err)
	}

	updatedLabels := map[string]string{
		"user-defined-label": "custom-label-value",
	}
	updatedAnnotations := map[string]string{
		"user-defined-annotation": "custom-annotation-val",
	}

	svcClient := framework.KubeClient.CoreV1().Services(ns)
	ssetClient := framework.KubeClient.AppsV1().StatefulSets(ns)

	resourceConfigs := []struct {
		name   string
		get    func() (metav1.Object, error)
		update func(object metav1.Object) (metav1.Object, error)
	}{
		{
			name: "thanos-ruler-operated service",
			get: func() (metav1.Object, error) {
				return svcClient.Get(context.Background(), "thanos-ruler-operated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return svcClient.Update(context.Background(), asService(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "thanos-ruler stateful set",
			get: func() (metav1.Object, error) {
				return ssetClient.Get(context.Background(), "thanos-ruler-test", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return ssetClient.Update(context.Background(), asStatefulSet(t, object), metav1.UpdateOptions{})
			},
		},
	}

	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		if err != nil {
			t.Fatal(err)
		}

		updateObjectLabels(res, updatedLabels)
		updateObjectAnnotations(res, updatedAnnotations)

		_, err = rConf.update(res)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Ensure resource reconciles
	thanosRuler.Spec.Replicas = proto.Int32(2)
	_, err = framework.PatchThanosRulerAndWaitUntilReady(context.Background(), thanosRuler.Name, ns, thanosRuler.Spec)
	if err != nil {
		t.Fatal(err)
	}

	// Assert labels preserved
	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		if err != nil {
			t.Fatal(err)
		}

		labels := res.GetLabels()
		if !containsValues(labels, updatedLabels) {
			t.Errorf("%s: labels do not contain updated labels, found: %q, should contain: %q", rConf.name, labels, updatedLabels)
		}

		annotations := res.GetAnnotations()
		if !containsValues(annotations, updatedAnnotations) {
			t.Fatalf("%s: annotations do not contain updated annotations, found: %q, should contain: %q", rConf.name, annotations, updatedAnnotations)
		}
	}

	if err := framework.DeleteThanosRulerAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
}

func testTRMinReadySeconds(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	kubeClient := framework.KubeClient

	thanosRuler := framework.MakeBasicThanosRuler("test-thanos", 1, "http://test.example.com")
	thanosRuler.Spec.MinReadySeconds = ptr.To(int32(5))
	thanosRuler, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanosRuler)
	require.NoError(t, err)

	trSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "thanos-ruler-test-thanos", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, int32(5), trSS.Spec.MinReadySeconds)

	thanosRuler.Spec.MinReadySeconds = ptr.To(int32(10))
	_, err = framework.PatchThanosRulerAndWaitUntilReady(context.Background(), thanosRuler.Name, ns, thanosRuler.Spec)
	require.NoError(t, err)

	trSS, err = kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "thanos-ruler-test-thanos", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, int32(10), trSS.Spec.MinReadySeconds)
}

// Tests Thanos ruler -> Alertmanager path
// This is done by creating a firing rule that will be picked up by
// Thanos Ruler which will send it to Alertmanager, finally we will
// use the Alertmanager API to validate that the alert is there.
func testTRAlertmanagerConfig(t *testing.T) {
	const (
		name       = "test"
		group      = "thanos-alertmanager-test"
		secretName = "thanos-ruler-alertmanagers-config"
		configKey  = "alertmanagers.yaml"
		testAlert  = "alert1"
	)
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	// Create Alertmanager resource and service
	alertmanager, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), framework.MakeBasicAlertmanager(ns, name, 1))
	require.NoError(t, err)

	amSVC := framework.MakeAlertmanagerService(alertmanager.Name, group, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, amSVC)
	require.NoError(t, err)

	// Create a Prometheus resource because Thanos ruler needs a query API.
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	require.NoError(t, err)

	svc := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	// Create Secret with Alertmanager config,
	trAmConfigSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			configKey: fmt.Appendf(nil, `
alertmanagers:
- scheme: http
  api_version: v2
  static_configs:
    - dnssrv+_web._tcp.%s.%s.svc.cluster.local
`, amSVC.Name, ns),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), trAmConfigSecret, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create Thanos ruler resource and service
	thanos := framework.MakeBasicThanosRuler(name, 1, fmt.Sprintf("http://%s:%d/", svc.Name, svc.Spec.Ports[0].Port))
	thanos.Spec.EvaluationInterval = "1s"
	thanos.Spec.AlertManagersConfig = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: secretName,
		},
		Key: configKey,
	}

	_, err = framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	require.NoError(t, err)

	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, framework.MakeThanosRulerService(thanos.Name, group, v1.ServiceTypeClusterIP))
	require.NoError(t, err)

	// Create firing rule
	_, err = framework.MakeAndCreateFiringRule(context.Background(), ns, "rule1", testAlert)
	require.NoError(t, err)

	err = framework.WaitForAlertmanagerFiringAlert(context.Background(), ns, amSVC.Name, testAlert)
	require.NoError(t, err)
}

// Tests Thanos ruler query Config
// This is done by creating a firing rule that will be picked up by
// Thanos Ruler which will only fire the rule if it's able to query prometheus
// it has to pull configuration from queryConfig file.
func testTRQueryConfig(t *testing.T) {
	const (
		name       = "test"
		group      = "thanos-ruler-query-config"
		secretName = "thanos-ruler-query-config"
		configKey  = "query.yaml"
		testAlert  = "alert1"
	)
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	// Create a Thanos querier which is used
	// - by the Thanos Ruler as a query API endpoint.
	// - by the test to query metrics from the Thanos Ruler.
	querier, err := testFramework.MakeThanosQuerier(
		fmt.Sprintf("dnssrv+_grpc._tcp.thanos-ruler-operated.%s.svc.cluster.local", ns),
	)
	require.NoError(t, err)

	err = framework.CreateDeployment(context.Background(), ns, querier)
	require.NoError(t, err)

	querierSvc := framework.MakeThanosQuerierService(querier.Name)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, querierSvc)
	require.NoError(t, err)

	// Create Secret with query config,
	trQueryConfSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			configKey: fmt.Appendf(nil, `
- scheme: http
  static_configs:
  - %s.%s.svc:%d
`, querierSvc.Name, ns, querierSvc.Spec.Ports[0].Port),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), trQueryConfSecret, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create Thanos ruler resource and service
	// setting queryEndpoint to "" as it will be ignored because we set QueryConfig
	thanos := framework.MakeBasicThanosRuler(name, 1, "")
	thanos.Spec.EvaluationInterval = "1s"
	thanos.Spec.QueryConfig = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: secretName,
		},
		Key: configKey,
	}

	_, err = framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	require.NoError(t, err)

	svc := framework.MakeThanosRulerService(thanos.Name, group, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	// Create an always firing rule.
	_, err = framework.MakeAndCreateFiringRule(context.Background(), ns, "rule1", testAlert)
	require.NoError(t, err)

	err = framework.WaitForThanosFiringAlert(context.Background(), ns, svc.Name, testAlert)
	require.NoError(t, err)

	// Check that the ALERTS metric is present via Thanos querier.
	err = framework.WaitForPrometheusFiringAlert(context.Background(), ns, querierSvc.Name, testAlert)
	require.NoError(t, err)
}

func testTRCheckStorageClass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)

	tr := framework.MakeBasicThanosRuler("test", 1, "http://test.example.com")

	tr, err := framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	if err != nil {
		t.Fatal(err)
	}

	// Invalid storageclass e2e test

	_, err = framework.PatchThanosRuler(
		context.Background(),
		tr.Name,
		ns,
		monitoringv1.ThanosRulerSpec{
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						StorageClassName: ptr.To("unknown-storage-class"),
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var loopError error
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1.ThanosRulers(ns).Get(ctx, tr.Name, metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err == nil {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		t.Fatalf("%v: %v", err, loopError)
	}
}

func testThanosRulerServiceName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(ctx, t, testCtx)
	name := "test-servicename"

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", name),
			Namespace: ns,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name: "web",
					Port: 9090,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":       "thanos-ruler",
				"app.kubernetes.io/managed-by": "prometheus-operator",
				"app.kubernetes.io/instance":   name,
				"thanos-ruler":                 name,
			},
		},
	}

	_, err := framework.KubeClient.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
	require.NoError(t, err)

	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)

	tr := framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	tr.Spec.ServiceName = &svc.Name

	_, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, tr)
	require.NoError(t, err)

	// Ensure that the default governing service was not created by the operator.
	svcList, err := framework.KubeClient.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, svcList.Items, 1)
	require.Equal(t, svcList.Items[0].Name, svc.Name)
}

func testThanosRulerStateless(t *testing.T) {
	const (
		name       = "test"
		group      = "thanos-ruler-query-config"
		secretName = "thanos-ruler-query-config"
		configKey  = "query.yaml"
		testAlert  = "alert1"
	)
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ctx := context.Background()
	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)

	// Create a Prometheus resource which will act as the query API endpoint +
	// remote-write receiver for Thanos ruler.
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheus.Spec.EnableRemoteWriteReceiver = true
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(ctx, ns, prometheus)
	// Ensure that the Promehteus resource selects no rule.
	prometheus.Spec.RuleSelector = nil
	require.NoError(t, err)

	promSVC := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(ctx, ns, promSVC)
	require.NoError(t, err)

	// Create the query config secret.
	trQueryConfSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			configKey: fmt.Appendf(nil, `
- scheme: http
  static_configs:
  - %s.%s.svc:%d
`, promSVC.Name, ns, promSVC.Spec.Ports[0].Port),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(ctx, trQueryConfSecret, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create the Thanos ruler resource.
	thanos := framework.MakeBasicThanosRuler(name, 1, "")
	thanos.Spec.EvaluationInterval = "1s"
	thanos.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{
		{
			URL: fmt.Sprintf("http://%s.%s.svc:%d/api/v1/write", promSVC.Name, ns, promSVC.Spec.Ports[0].Port),
			// Ensure that samples are sent ASAP to the remote write receiver.
			QueueConfig: &monitoringv1.QueueConfig{
				MaxSamplesPerSend: 1,
			},
		},
	}
	thanos.Spec.QueryConfig = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: secretName,
		},
		Key: configKey,
	}

	_, err = framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, thanos)
	require.NoError(t, err)

	svc := framework.MakeThanosRulerService(thanos.Name, group, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(ctx, ns, svc)
	require.NoError(t, err)

	// Create the always firing alerting rule and check that it is active.
	_, err = framework.MakeAndCreateFiringRule(ctx, ns, "rule1", testAlert)
	require.NoError(t, err)

	err = framework.WaitForThanosFiringAlert(ctx, ns, svc.Name, testAlert)
	require.NoError(t, err)

	// Check that the ALERTS metric is present in Prometheus.
	err = framework.WaitForPrometheusFiringAlert(context.Background(), ns, promSVC.Name, testAlert)
	require.NoError(t, err)
}

func testThanosRulerScaleUpWithoutLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)

	name := "test"

	// Create a ThanosRuler resource with 1 replica
	tr, err := framework.CreateThanosRulerAndWaitUntilReady(ctx, ns, framework.MakeBasicThanosRuler(name, 1, "http://test.example.com"))
	require.NoError(t, err)

	// Remove all labels on the StatefulSet using Patch
	stsName := fmt.Sprintf("thanos-ruler-%s", name)
	err = framework.RemoveAllLabelsFromStatefulSet(ctx, stsName, ns)
	require.NoError(t, err)

	// Scale up the ThanosRuler resource to 2 replicas
	_, err = framework.UpdateThanosRulerReplicasAndWaitUntilReady(ctx, tr.Name, ns, 2)
	require.NoError(t, err)

	// Verify the StatefulSet now has labels again (restored by the operator)
	stsClient := framework.KubeClient.AppsV1().StatefulSets(ns)
	sts, err := stsClient.Get(ctx, stsName, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, sts.GetLabels(), "expected labels to be restored on the StatefulSet by the operator")
}
