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

const testAnnotationControllerID = "42"

func testMultipleOperatorsPrometheusServer(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test-op-1"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	name = "test-op-2"
	p = framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Annotations["operator.prometheus.io/controller-id"] = testAnnotationControllerID
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}
}

func testMultipleOperatorsPrometheusAgent(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test-op-1"
	p := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	name = "test-op-2"
	p = framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	p.Annotations["operator.prometheus.io/controller-id"] = testAnnotationControllerID
	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}
}

func testMultipleOperatorsAlertManager(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test-op-1"
	a := framework.MakeBasicAlertmanager(ns, name, 1)
	_, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}

	name = "test-op-2"
	a = framework.MakeBasicAlertmanager(ns, name, 1)
	a.Annotations["operator.prometheus.io/controller-id"] = testAnnotationControllerID
	_, err = framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
}

func testMultipleOperatorsThanosRuler(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test-op-1"
	thanos := framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	_, err := framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	if err != nil {
		t.Fatal(err)
	}

	name = "test-op-2"
	thanos = framework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	thanos.Annotations["operator.prometheus.io/controller-id"] = testAnnotationControllerID
	_, err = framework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanos)
	if err != nil {
		t.Fatal(err)
	}
}
