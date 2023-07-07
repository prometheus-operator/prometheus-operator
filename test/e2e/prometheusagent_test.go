// Copyright 2023 The prometheus-operator Authors
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

func testCreatePrometheusAgent(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}

}

func testAgentAndServerNameColision(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD); err != nil {
		t.Fatal(err)
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
	if err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}

}
