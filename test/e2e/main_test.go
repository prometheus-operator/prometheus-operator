// Copyright 2016 The prometheus-operator Authors
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
	"flag"
	"log"
	"os"
	"testing"

	operatorFramework "github.com/coreos/prometheus-operator/test/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var (
	framework *operatorFramework.Framework
	opImage   *string
)

func skipPrometheusTests(t *testing.T) {
	if os.Getenv("EXCLUDE_PROMETHEUS_TESTS") != "" {
		t.Skip("Skipping Prometheus tests")
	}
}

func skipAlertmanagerTests(t *testing.T) {
	if os.Getenv("EXCLUDE_ALERTMANAGER_TESTS") != "" {
		t.Skip("Skipping Alertmanager tests")
	}
}

func skipThanosRulerTests(t *testing.T) {
	if os.Getenv("EXCLUDE_THANOSRULER_TESTS") != "" {
		t.Skip("Skipping ThanosRuler tests")
	}
}

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		"",
		"kube config path, e.g. $HOME/.kube/config",
	)
	opImage = flag.String(
		"operator-image",
		"",
		"operator image, e.g. quay.io/coreos/prometheus-operator",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	if framework, err = operatorFramework.New(*kubeconfig, *opImage); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

// TestAllNS tests the Prometheus Operator watching all namespaces in a
// Kubernetes cluster.
func TestAllNS(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	ns := ctx.CreateNamespace(t, framework.KubeClient)

	finalizers, err := framework.CreatePrometheusOperator(ns, *opImage, nil, nil, nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range finalizers {
		ctx.AddFinalizerFn(f)
	}

	t.Run("TestCRDs", testCRDs)

	// t.Run blocks until the function passed as the second argument (f) returns or
	// calls t.Parallel to become a parallel test. Run reports whether f succeeded
	// (or at least did not fail before calling t.Parallel). As all tests in
	// testAllNS are parallel, the defered ctx.Cleanup above would be run before
	// all tests finished. Wrapping it in testAllNSPrometheus and testAllNSAlertmanager
	// fixes this.
	t.Run("x", testAllNSAlertmanager)
	t.Run("y", testAllNSPrometheus)
	t.Run("z", testAllNSThanosRuler)

	// Check if Prometheus Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"app.kubernetes.io/name": "prometheus-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if expected := 1; len(pl.Items) != expected {
		t.Fatalf("expected %v Prometheus Operator pods, but got %v", expected, len(pl.Items))
	}
	restarts, err := framework.GetPodRestartCount(ns, pl.Items[0].GetName())
	if err != nil {
		t.Fatalf("failed to retrieve restart count of Prometheus Operator pod: %v", err)
	}
	if len(restarts) != 2 {
		t.Fatalf("expected to have 2 containers but got %d", len(restarts))
	}
	for _, restart := range restarts {
		if restart != 0 {
			t.Fatalf(
				"expected Prometheus Operator to never restart during entire test execution but got %d restarts",
				restart,
			)
		}
	}
}

func testCRDs(t *testing.T) {
	const (
		prometheusCRDName = "prometheuses.monitoring.coreos.com"
	)
	crds, err := framework.ListCRDs()
	if err != nil {
		t.Fatalf("unable to list CRDs: %v", err)
	}
	if len(crds.Items) < 5 {
		t.Fatalf("incorrect number of CRDs, want at least: %v, got %v", 5, len(crds.Items))
	}
	crd, err := framework.GetCRD(prometheusCRDName)
	if err != nil {
		t.Fatalf("unable to get prometheus custom resource definition: %v", err)
	}
	// This field might be nil in older versions of kube (<1.15)
	if crd.Spec.PreserveUnknownFields != nil && *crd.Spec.PreserveUnknownFields {
		t.Fatalf("incorrect setting for preserveUnknownFields, want: %v, got: %v", false, *crd.Spec.PreserveUnknownFields)
	}
}

func testAllNSAlertmanager(t *testing.T) {
	skipAlertmanagerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AMCreateDeleteCluster":           testAMCreateDeleteCluster,
		"AMScaling":                       testAMScaling,
		"AMVersionMigration":              testAMVersionMigration,
		"AMStorageUpdate":                 testAMStorageUpdate,
		"AMExposingWithKubernetesAPI":     testAMExposingWithKubernetesAPI,
		"AMClusterInitialization":         testAMClusterInitialization,
		"AMClusterAfterRollingUpdate":     testAMClusterAfterRollingUpdate,
		"AMClusterGossipSilences":         testAMClusterGossipSilences,
		"AMReloadConfig":                  testAMReloadConfig,
		"AMZeroDowntimeRollingDeployment": testAMZeroDowntimeRollingDeployment,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

func testAllNSPrometheus(t *testing.T) {
	skipPrometheusTests(t)
	testFuncs := map[string]func(t *testing.T){
		"PromCreateDeleteCluster":                testPromCreateDeleteCluster,
		"PromScaleUpDownCluster":                 testPromScaleUpDownCluster,
		"PromNoServiceMonitorSelector":           testPromNoServiceMonitorSelector,
		"PromVersionMigration":                   testPromVersionMigration,
		"PromResourceUpdate":                     testPromResourceUpdate,
		"PromStorageUpdate":                      testPromStorageUpdate,
		"PromReloadConfig":                       testPromReloadConfig,
		"PromAdditionalScrapeConfig":             testPromAdditionalScrapeConfig,
		"PromAdditionalAlertManagerConfig":       testPromAdditionalAlertManagerConfig,
		"PromReloadRules":                        testPromReloadRules,
		"PromMultiplePrometheusRulesSameNS":      testPromMultiplePrometheusRulesSameNS,
		"PromMultiplePrometheusRulesDifferentNS": testPromMultiplePrometheusRulesDifferentNS,
		"PromRulesExceedingConfigMapLimit":       testPromRulesExceedingConfigMapLimit,
		"PromRulesMustBeAnnotated":               testPromRulesMustBeAnnotated,
		"PromtestInvalidRulesAreRejected":        testInvalidRulesAreRejected,
		"PromOnlyUpdatedOnRelevantChanges":       testPromOnlyUpdatedOnRelevantChanges,
		"PromWhenDeleteCRDCleanUpViaOwnerRef":    testPromWhenDeleteCRDCleanUpViaOwnerRef,
		"PromDiscovery":                          testPromDiscovery,
		"PromAlertmanagerDiscovery":              testPromAlertmanagerDiscovery,
		"PromExposingWithKubernetesAPI":          testPromExposingWithKubernetesAPI,
		"PromDiscoverTargetPort":                 testPromDiscoverTargetPort,
		"PromOpMatchPromAndServMonInDiffNSs":     testPromOpMatchPromAndServMonInDiffNSs,
		"PromGetAuthSecret":                      testPromGetAuthSecret,
		"PromArbitraryFSAcc":                     testPromArbitraryFSAcc,
		"PromTLSConfigViaSecret":                 testPromTLSConfigViaSecret,
		"Thanos":                                 testThanos,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

func testAllNSThanosRuler(t *testing.T) {
	skipThanosRulerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"ThanosRulerCreateDeleteCluster": testTRCreateDeleteCluster,
	}
	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestMultiNS tests the Prometheus Operator configured to watch specific
// namespaces.
func TestMultiNS(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"OperatorNSScope": testOperatorNSScope,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestDenylist tests the Prometheus Operator configured not to watch specific namespaces.
func TestDenylist(t *testing.T) {
	skipPrometheusTests(t)
	testFuncs := map[string]func(t *testing.T){
		"Prometheus":     testDenyPrometheus,
		"ServiceMonitor": testDenyServiceMonitor,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestPromInstanceNs tests prometheus operator in different scenarios when --prometheus-instance-namespace is given
func TestPromInstanceNs(t *testing.T) {
	skipPrometheusTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AllNs":     testPrometheusInstanceNamespaces_AllNs,
		"AllowList": testPrometheusInstanceNamespaces_AllowList,
		"DenyList":  testPrometheusInstanceNamespaces_DenyList,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestAlertmanagerInstanceNs tests prometheus operator in different scenarios when --alertmanager-instance-namespace is given
func TestAlertmanagerInstanceNs(t *testing.T) {
	skipAlertmanagerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AllNs":  testAlertmanagerInstanceNamespaces_AllNs,
		"DenyNs": testAlertmanagerInstanceNamespaces_DenyNs,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}
