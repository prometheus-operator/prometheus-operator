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

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func testAlertmanagerInstanceNamespacesAllNs(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

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
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	nonInstanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNs, nil, nil, nil, []string{instanceNs}, false, true, true)
	if err != nil {
		t.Fatal(err)
	}

	am := framework.MakeBasicAlertmanager(nonInstanceNs, "non-instance", 3)
	_, err = framework.MonClientV1.Alertmanagers(nonInstanceNs).Create(context.Background(), am, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	am = framework.MakeBasicAlertmanager(instanceNs, "instance", 3)
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), am); err != nil {
		t.Fatal(err)
	}

	sts, err := framework.KubeClient.AppsV1().StatefulSets(nonInstanceNs).Get(context.Background(), "alertmanager-instance", metav1.GetOptions{})
	if !api_errors.IsNotFound(err) {
		t.Fatalf("expected not to find an Alertmanager statefulset, but did: %v/%v", sts.Namespace, sts.Name)
	}
}

func testAlertmanagerInstanceNamespacesDenyNs(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// create two namespaces:
	//
	// 1. "operator" ns:
	//   - hosts the prometheus operator deployment
	//
	// 2. "instance" ns:
	//   - will be configured on prometheus operator as --alertmanager-instance-namespaces="instance"
	//   - will additionally be configured on prometheus operator as --deny-namespaces="instance"
	//   - hosts an alertmanager CR which must be reconciled.
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNs, nil, []string{instanceNs}, nil, []string{instanceNs}, false, true, true)
	if err != nil {
		t.Fatal(err)
	}

	am := framework.MakeBasicAlertmanager(instanceNs, "instance", 3)
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), am); err != nil {
		t.Fatal(err)
	}
}

func testAlertmanagerInstanceNamespacesAllowList(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

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
	operatorNs := framework.CreateNamespace(context.Background(), t, testCtx)
	instanceNs := framework.CreateNamespace(context.Background(), t, testCtx)
	allowedNs := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, instanceNs)

	for _, ns := range []string{allowedNs, instanceNs} {
		err := framework.AddLabelsToNamespace(context.Background(), ns, map[string]string{
			"monitored": "true",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Configure the operator to watch also a non-existing namespace (e.g. "notfound").
	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNs, []string{"notfound", allowedNs}, nil, nil, []string{"notfound", instanceNs}, false, true, true)
	if err != nil {
		t.Fatal(err)
	}

	// Create the Alertmanager resource in the "allowed" namespace. We will check later that it is NOT reconciled.
	am := framework.MakeBasicAlertmanager(allowedNs, "instance", 3)

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

	// Create an Alertmanager resource in the "allowedNs" namespace which must *not* be reconciled.
	_, err = framework.MonClientV1.Alertmanagers(allowedNs).Create(context.Background(), am, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Create an Alertmanager resource in the "instance" namespace which must be reconciled.
	am.Namespace = instanceNs
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), am); err != nil {
		t.Fatal(err)
	}

	// Check that the Alertmanager resource created in the "allowed" namespace hasn't been reconciled.
	sts, err := framework.KubeClient.AppsV1().StatefulSets(allowedNs).Get(context.Background(), "alertmanager-instance", metav1.GetOptions{})
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

	if _, err = framework.MonClientV1alpha1.AlertmanagerConfigs(instanceNs).Create(context.Background(), amConfig, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err = framework.MonClientV1alpha1.AlertmanagerConfigs(allowedNs).Create(context.Background(), amConfig, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// Check that the AlertmanagerConfig resource in the "allowed" namespace is reconciled but not the one in "instance".
	err = framework.PollAlertmanagerConfiguration(context.Background(), instanceNs, "instance",
		func(config string) error {
			if !strings.Contains(config, "void") {
				return fmt.Errorf("expected generated configuration to contain %q but got %q", "void", config)
			}

			return nil
		},
		func(config string) error {
			if strings.Contains(config, instanceNs) {
				return fmt.Errorf("expected generated configuration to not contain %q but got %q", instanceNs, config)
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("failed to wait for alertmanager config: %v", err)
	}

	// FIXME(simonpasquier): the unprivileged namespace lister/watcher
	// isn't notified of updates properly so the code below fails.
	// Uncomment the test once the lister/watcher is fixed.
	//
	// Remove the selecting label on the "allowed" namespace and check that
	// the alertmanager configuration is updated.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
	//if err := framework.RemoveLabelsFromNamespace(allowedNs, "monitored"); err != nil {
	//	t.Fatal(err)
	//}

	//err = framework.PollAlertmanagerConfiguration(instanceNs, "instance",
	//	func(config string) error {
	//		if strings.Contains(config, "void") {
	//			return fmt.Errorf("expected generated configuration to not contain %q but got %q", "void", config)
	//		}

	//		return nil
	//	},
	//)

	//if err != nil {
	//	t.Fatalf("failed to wait for alertmanager config: %v", err)
	//}
}
