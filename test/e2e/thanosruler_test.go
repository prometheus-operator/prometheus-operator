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
	"fmt"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func testThanosRulerCreateDeleteCluster(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	ns := framework.CreateNamespace(t, ctx)
	framework.SetupPrometheusRBAC(t, ctx, ns)

	name := "test"

	if _, err := framework.CreateThanosRulerAndWaitUntilReady(ns, framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeleteThanosRulerAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func testThanosRulerPrometheusRuleInDifferentNamespace(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	thanosNamespace := framework.CreateNamespace(t, ctx)
	framework.SetupPrometheusRBAC(t, ctx, thanosNamespace)

	name := "test"

	// Create a Prometheus resource because Thanos ruler needs a query API.
	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(thanosNamespace, framework.MakeBasicPrometheus(thanosNamespace, name, name, 1))
	if err != nil {
		t.Fatal(err)
	}

	svc := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	if _, err := framework.CreateServiceAndWaitUntilReady(thanosNamespace, svc); err != nil {
		t.Fatal(err)
	}

	thanos := framework.MakeBasicThanosRuler(name, 1, fmt.Sprintf("http://%s:%d/", svc.Name, svc.Spec.Ports[0].Port))
	thanos.Spec.RuleSelector = &metav1.LabelSelector{}
	thanos.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"monitored": "true",
		},
	}
	thanos.Spec.EvaluationInterval = "1s"

	if _, err := framework.CreateThanosRulerAndWaitUntilReady(thanosNamespace, thanos); err != nil {
		t.Fatal(err)
	}

	ruleNamespace := framework.CreateNamespace(t, ctx)
	if err := framework.AddLabelsToNamespace(ruleNamespace, map[string]string{
		"monitored": "true",
	}); err != nil {
		t.Fatal(err)
	}

	const testAlert = "alert1"
	_, err = framework.MakeAndCreateFiringRule(ruleNamespace, "rule1", testAlert)
	if err != nil {
		t.Fatal(err)
	}

	thanosService := framework.MakeThanosRulerService(thanos.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(thanosNamespace, thanosService); err != nil {
		t.Fatalf("creating Thanos ruler service failed: %v", err)
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if err := framework.WaitForThanosFiringAlert(thanosNamespace, thanosService.Name, testAlert); err != nil {
		t.Fatal(err)
	}

	// Remove the selecting label from ruleNamespace and wait until the rule is
	// removed from the Thanos ruler.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
	if err := framework.RemoveLabelsFromNamespace(ruleNamespace, "monitored"); err != nil {
		t.Fatal(err)
	}

	var loopError error
	err = wait.Poll(time.Second, 5*framework.DefaultTimeout, func() (bool, error) {
		var firing bool
		firing, loopError = framework.CheckThanosFiringAlert(thanosNamespace, thanosService.Name, testAlert)
		return !firing, nil
	})

	if err != nil {
		t.Fatalf("waiting for alert %q to stop firing: %v: %v", testAlert, err, loopError)
	}
}

func testTRPreserveUserAddedMetadata(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := framework.CreateNamespace(t, ctx)
	framework.SetupPrometheusRBAC(t, ctx, ns)

	name := "test"

	thanosRuler := framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	thanosRuler, err := framework.CreateThanosRulerAndWaitUntilReady(ns, thanosRuler)
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
				return svcClient.Get(framework.Ctx, "thanos-ruler-operated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return svcClient.Update(framework.Ctx, asService(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "thanos-ruler stateful set",
			get: func() (metav1.Object, error) {
				return ssetClient.Get(framework.Ctx, "thanos-ruler-test", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return ssetClient.Update(framework.Ctx, asStatefulSet(t, object), metav1.UpdateOptions{})
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
	_, err = framework.UpdateThanosRulerAndWaitUntilReady(ns, thanosRuler)
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

	if err := framework.DeleteThanosRulerAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func testTRMinReadySeconds(t *testing.T) {
	runFeatureGatedTests(t)
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := framework.CreateNamespace(t, ctx)
	framework.SetupPrometheusRBAC(t, ctx, ns)

	kubeClient := framework.KubeClient

	var setMinReadySecondsInitial uint32 = 5
	thanosRuler := framework.MakeBasicThanosRuler("test-thanos", 1, "http://test.example.com")
	thanosRuler.Spec.MinReadySeconds = &setMinReadySecondsInitial
	thanosRuler, err := framework.CreateThanosRulerAndWaitUntilReady(ns, thanosRuler)
	if err != nil {
		t.Fatal(err)
	}

	trSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(framework.Ctx, "thanos-ruler-test-thanos", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if trSS.Spec.MinReadySeconds != int32(setMinReadySecondsInitial) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", setMinReadySecondsInitial, trSS.Spec.MinReadySeconds)
	}

	var updated uint32 = 10
	thanosRuler.Spec.MinReadySeconds = &updated
	if _, err = framework.UpdateThanosRulerAndWaitUntilReady(ns, thanosRuler); err != nil {
		t.Fatal("Updating ThanosRuler failed: ", err)
	}

	trSS, err = kubeClient.AppsV1().StatefulSets(ns).Get(framework.Ctx, "thanos-ruler-test-thanos", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if trSS.Spec.MinReadySeconds != int32(updated) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", updated, trSS.Spec.MinReadySeconds)
	}
}
