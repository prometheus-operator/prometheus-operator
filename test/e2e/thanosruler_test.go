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

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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

	var setMinReadySecondsInitial uint32 = 5
	thanosRuler := framework.MakeBasicThanosRuler("test-thanos", 1, "http://test.example.com")
	thanosRuler.Spec.MinReadySeconds = &setMinReadySecondsInitial
	thanosRuler, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanosRuler)
	if err != nil {
		t.Fatal(err)
	}

	trSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "thanos-ruler-test-thanos", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if trSS.Spec.MinReadySeconds != int32(setMinReadySecondsInitial) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", setMinReadySecondsInitial, trSS.Spec.MinReadySeconds)
	}

	var updated uint32 = 10
	thanosRuler.Spec.MinReadySeconds = &updated
	if _, err = framework.PatchThanosRulerAndWaitUntilReady(context.Background(), thanosRuler.Name, ns, thanosRuler.Spec); err != nil {
		t.Fatal("patching ThanosRuler failed: ", err)
	}

	trSS, err = kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "thanos-ruler-test-thanos", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if trSS.Spec.MinReadySeconds != int32(updated) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", updated, trSS.Spec.MinReadySeconds)
	}
}

// Tests Thanos ruler -> Alertmanger path
// This is done by creating a firing rule that will be picked up by
// Thanos Ruler which will send it to Alertmanager, finally we will
// use the Alertmanager API to validate that the alert is there
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
	assert.NoError(t, err)

	amSVC := framework.MakeAlertmanagerService(alertmanager.Name, group, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, amSVC)
	assert.NoError(t, err)

	// Create a Prometheus resource because Thanos ruler needs a query API.
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	assert.NoError(t, err)

	svc := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	assert.NoError(t, err)

	// Create Secret with Alermanager config,
	trAmConfigSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			configKey: []byte(fmt.Sprintf(`
alertmanagers:
- scheme: http
  api_version: v2
  static_configs:
    - dnssrv+_web._tcp.%s.%s.svc.cluster.local
`, amSVC.Name, ns)),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), trAmConfigSecret, metav1.CreateOptions{})
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, framework.MakeThanosRulerService(thanos.Name, group, v1.ServiceTypeClusterIP))
	assert.NoError(t, err)

	// Create firing rule
	_, err = framework.MakeAndCreateFiringRule(context.Background(), ns, "rule1", testAlert)
	assert.NoError(t, err)

	err = framework.WaitForAlertmanagerFiringAlert(context.Background(), ns, amSVC.Name, testAlert)
	assert.NoError(t, err)
}

// Tests Thanos ruler query Config
// This is done by creating a firing rule that will be picked up by
// Thanos Ruler which will only fire the rule if it's able to query prometheus
// it has to pull configuration from queryConfig file
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
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	// Create a Prometheus resource because Thanos ruler needs a query API.
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	assert.NoError(t, err)

	promSVC := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, promSVC)
	assert.NoError(t, err)

	// Create Secret with query config,
	trQueryConfSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			configKey: []byte(fmt.Sprintf(`
- scheme: http
  static_configs:
  - %s.%s.svc:%d
`, promSVC.Name, ns, promSVC.Spec.Ports[0].Port)),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), trQueryConfSecret, metav1.CreateOptions{})
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	svc := framework.MakeThanosRulerService(thanos.Name, group, v1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	assert.NoError(t, err)

	// Create firing rule
	_, err = framework.MakeAndCreateFiringRule(context.Background(), ns, "rule1", testAlert)
	assert.NoError(t, err)

	if err := framework.WaitForThanosFiringAlert(context.Background(), ns, svc.Name, testAlert); err != nil {
		t.Fatal(err)
	}
}
