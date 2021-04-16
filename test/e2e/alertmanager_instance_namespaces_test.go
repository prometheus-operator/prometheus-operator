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
	"fmt"
	"strings"
	"testing"
	"time"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testAlertmanagerInstanceNamespacesAllNs(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	// create 3 namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --alertmanager-instance-namespaces="instance"
	//
	// 3. "nonInstance" ns:
	//   - hosts an Alertmanager CR which must not be reconciled
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

func testAlertmanagerInstanceNamespacesDenyNs(t *testing.T) {
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

func testAlertmanagerInstanceNamespacesAllowList(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	// create 3 namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --alertmanager-instance-namespaces="instance"
	//   - hosts an Alertmanager CR which will select AlertmanagerConfig resources in all "allowed" namespaces.
	//   - hosts an AlertmanagerConfig CR which must not be reconciled.
	//
	// 3. "allowed" ns:
	//   - will be configured on prometheus operator as --namespaces="allowed"
	//   - hosts an AlertmanagerConfig CR which must be reconciled
	//   - hosts an Alertmanager CR which must not reconciled.
	operatorNs := ctx.CreateNamespace(t, framework.KubeClient)
	instanceNs := ctx.CreateNamespace(t, framework.KubeClient)
	allowedNs := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBACGlobal(t, instanceNs, framework.KubeClient)

	for _, ns := range []string{allowedNs, instanceNs} {
		err := testFramework.AddLabelsToNamespace(framework.KubeClient, ns, map[string]string{
			"monitored": "true",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Configure the operator to watch also a non-existing namespace (e.g. "notfound").
	_, err := framework.CreatePrometheusOperator(operatorNs, *opImage, []string{"notfound", allowedNs}, nil, nil, []string{"notfound", instanceNs}, false, true)
	if err != nil {
		t.Fatal(err)
	}

	// Create the Alertmanager resource in the "allowed" namespace. We will check later that it is NOT reconciled.
	am := framework.MakeBasicAlertmanager("instance", 3)

	am.Spec.AlertmanagerConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": "monitored",
		},
	}

	am.Spec.AlertmanagerConfigNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"monitored": "true",
		},
	}

	_, err = framework.MonClientV1.Alertmanagers(allowedNs).Create(context.TODO(), am, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Create an Alertmanager resource in the "instance" namespace which must be reconciled.
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(instanceNs, am); err != nil {
		t.Fatal(err)
	}

	// Check that the Alertmanager resource created in the "allowed" namespace hasn't been reconciled.
	sts, err := framework.KubeClient.AppsV1().StatefulSets(allowedNs).Get(context.TODO(), "alertmanager-instance", metav1.GetOptions{})
	if !api_errors.IsNotFound(err) {
		t.Fatalf("expected not to find an Alertmanager statefulset, but did: %v/%v", sts.Namespace, sts.Name)
	}

	// Create the AlertmanagerConfig resources in the "instance" and "allowed" namespaces.
	amConfig := &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "e2e-test-amconfig-multi-namespace",
			Labels: map[string]string{
				"group": "monitored",
			},
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "void",
			},
			Receivers: []monitoringv1alpha1.Receiver{{
				Name: "void",
			}},
		},
	}

	if _, err = framework.MonClientV1alpha1.AlertmanagerConfigs(instanceNs).Create(context.TODO(), amConfig, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err = framework.MonClientV1alpha1.AlertmanagerConfigs(allowedNs).Create(context.TODO(), amConfig, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// Check that the AlertmanagerConfig resource in the "allowed" namespace is reconciled but not the one in "instance".
	var pollError error
	err = wait.Poll(10*time.Second, time.Minute*5, func() (bool, error) {
		amStatus, err := framework.GetAlertmanagerStatus(instanceNs, "alertmanager-instance-0")
		if err != nil {
			pollError = fmt.Errorf("failed to query Alertmanager: %s", err)
			return false, nil
		}

		if !strings.Contains(*amStatus.Config.Original, "void") {
			pollError = fmt.Errorf("expected generated configuration to contain %q but got %q", "void", *amStatus.Config.Original)
			return false, nil
		}

		if strings.Contains(*amStatus.Config.Original, instanceNs) {
			pollError = fmt.Errorf("expected generated configuration to not contain %q but got %q", instanceNs, *amStatus.Config.Original)
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		t.Fatalf("failed to wait for alertmanager config: %v: %v", err, pollError)
	}
}
