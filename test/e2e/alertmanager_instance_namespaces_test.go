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

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testAlertmanagerInstanceNamespaces_AllNs(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	operatorNs := ctx.CreateNamespace(t, framework.KubeClient)
	instanceNs := ctx.CreateNamespace(t, framework.KubeClient)
	nonInstanceNs := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBACGlobal(t, instanceNs, framework.KubeClient)

	_, err := framework.CreatePrometheusOperator(operatorNs, *opImage, nil, nil, nil, []string{instanceNs}, false, true)
	if err != nil {
		t.Fatal(err)
	}

	am := framework.MakeBasicAlertmanager("non-instance", 3)
	am.Namespace = nonInstanceNs
	_, err = framework.MonClientV1.Alertmanagers(nonInstanceNs).Create(context.TODO(), am, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	am = framework.MakeBasicAlertmanager("instance", 3)
	am.Namespace = instanceNs
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(instanceNs, am); err != nil {
		t.Fatal(err)
	}

	sts, err := framework.KubeClient.AppsV1().StatefulSets(nonInstanceNs).Get(context.TODO(), "alertmanager-instance", metav1.GetOptions{})
	if !api_errors.IsNotFound(err) {
		t.Fatalf("expected not to find an Alertmanager statefulset, but did: %v/%v", sts.Namespace, sts.Name)
	}
}

func testAlertmanagerInstanceNamespaces_DenyNs(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	// create two namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --alertmanager-instance-namespaces="instance"
	//   - will additionally be configured on prometheus operator as --deny-namespaces="instance"
	//   - hosts an alertmanager CR which must be reconciled.
	operatorNs := ctx.CreateNamespace(t, framework.KubeClient)
	instanceNs := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBACGlobal(t, instanceNs, framework.KubeClient)

	_, err := framework.CreatePrometheusOperator(operatorNs, *opImage, nil, []string{instanceNs}, nil, []string{instanceNs}, false, true)
	if err != nil {
		t.Fatal(err)
	}

	am := framework.MakeBasicAlertmanager("instance", 3)
	am.Namespace = instanceNs
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(instanceNs, am); err != nil {
		t.Fatal(err)
	}
}
