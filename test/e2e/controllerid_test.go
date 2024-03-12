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
)

func testControllerCorrectIDPrometheusServer(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Annotations["operator.prometheus.io/controller-id"] = "88"

	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), p)
	if err == nil {
		t.Fatal("object is controlled by prometheus-operator but controllerID is different and must not")
	}

	name = "test-2"
	p = framework.MakeBasicPrometheus(ns, name, name, 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
}

func testControllerCorrectIDPrometheusAgent(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	p := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	p.Annotations["operator.prometheus.io/controller-id"] = "88"

	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), p)
	if err == nil {
		t.Fatal("object is controlled by prometheus-operator but controllerID is different and must not")
	}

	name = "test-2"
	p = framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
}

func testControllerCorrectIDAlertManager(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	a := framework.MakeBasicAlertmanager(ns, name, 1)
	a.Annotations["operator.prometheus.io/controller-id"] = "88"

	_, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err == nil {
		t.Fatal("object is controlled by prometheus-operator but controllerID is different and must not")
	}

	name = "test-2"
	a = framework.MakeBasicAlertmanager(ns, name, 1)
	_, err = framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
}

func testControllerCorrectIDThanos(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	thanos := framework.MakeBasicThanosRuler(name, 1, "")
	thanos.Annotations["operator.prometheus.io/controller-id"] = "88"

	_, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	if err == nil {
		t.Fatal("object is controlled by prometheus-operator but controllerID is different and must not")
	}

	name = "test-2"
	thanos = framework.MakeBasicThanosRuler(name, 1, "")
	_, err = framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	if err != nil {
		t.Fatal(err)
	}
}

func testControllerMultipleOperators(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	addArgs := []string{"--controller-id=42"}

	finalizers, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, nil, nil, nil, nil, true, true, true, addArgs)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	name := "operator-1"
	a := framework.MakeBasicAlertmanager(ns, name, 1)
	_, err = framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}

	name = "operator-2"
	a = framework.MakeBasicAlertmanager(ns, name, 1)
	a.Annotations["operator.prometheus.io/controller-id"] = "42"
	_, err = framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
}
