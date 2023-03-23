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
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var (
	previousVersionFramework *operatorFramework.Framework
	framework                *operatorFramework.Framework
	opImage                  *string
)

func skipPrometheusAllNSTests(t *testing.T) {
	if os.Getenv("EXCLUDE_PROMETHEUS_ALL_NS_TESTS") != "" {
		t.Skip("Skipping Prometheus all namespace tests")
	}
}

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

func skipOperatorUpgradeTests(t *testing.T) {
	if os.Getenv("EXCLUDE_OPERATOR_UPGRADE_TESTS") != "" {
		t.Skip("Skipping Operator upgrade tests")
	}
}

// feature gated tests need to be explicitly included
func runFeatureGatedTests(t *testing.T) {
	if os.Getenv("FEATURE_GATED_TESTS") != "include" {
		t.Skip("Skipping Feature Gated tests")
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
		"operator image, e.g. quay.io/prometheus-operator/prometheus-operator",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	logger := log.New(os.Stdout, "", log.Lshortfile)

	currentVersion, err := os.ReadFile("../../VERSION")
	if err != nil {
		logger.Printf("failed to read version file: %v\n", err)
		os.Exit(1)
	}
	currentSemVer, err := semver.ParseTolerant(string(currentVersion))
	if err != nil {
		logger.Printf("failed to parse current version: %v\n", err)
		os.Exit(1)
	}

	prevStableVersionURL := fmt.Sprintf("https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-%d.%d/VERSION", currentSemVer.Major, currentSemVer.Minor-1)
	reader, err := operatorFramework.URLToIOReader(prevStableVersionURL)
	if err != nil {
		logger.Printf("failed to get previous version file content: %v\n", err)
		os.Exit(1)
	}

	prevStableVersion, err := io.ReadAll(reader)
	if err != nil {
		logger.Printf("failed to read previous stable version: %v\n", err)
		os.Exit(1)
	}

	prometheusOperatorGithubBranchURL := "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator"

	prevSemVer, err := semver.ParseTolerant(string(prevStableVersion))
	if err != nil {
		logger.Printf("failed to parse previous stable version: %v\n", err)
		os.Exit(1)
	}
	prevStableOpImage := fmt.Sprintf("%s:v%s", "quay.io/prometheus-operator/prometheus-operator", strings.TrimSpace(string(prevStableVersion)))
	prevExampleDir := fmt.Sprintf("%s/release-%d.%d/example", prometheusOperatorGithubBranchURL, prevSemVer.Major, prevSemVer.Minor)
	prevResourcesDir := fmt.Sprintf("%s/release-%d.%d/test/framework/resources", prometheusOperatorGithubBranchURL, prevSemVer.Major, prevSemVer.Minor)

	if previousVersionFramework, err = operatorFramework.New(*kubeconfig, prevStableOpImage, prevExampleDir, prevResourcesDir, prevSemVer); err != nil {
		logger.Printf("failed to setup previous version framework: %v\n", err)
		os.Exit(1)
	}

	exampleDir := "../../example"
	resourcesDir := "../framework/resources"

	nextSemVer, err := semver.ParseTolerant(fmt.Sprintf("0.%d.0", currentSemVer.Minor))
	if err != nil {
		logger.Printf("failed to parse next version: %v\n", err)
		os.Exit(1)
	}

	// init with next minor version since we are developing toward it.
	if framework, err = operatorFramework.New(*kubeconfig, *opImage, exampleDir, resourcesDir, nextSemVer); err != nil {
		logger.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

// TestAllNS tests the Prometheus Operator watching all namespaces in a
// Kubernetes cluster.
func TestAllNS(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	finalizers, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, nil, nil, nil, nil, true, true, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	t.Run("TestServerTLS", testServerTLS(context.Background(), t, ns))

	// t.Run blocks until the function passed as the second argument (f) returns or
	// calls t.Parallel to become a parallel test. Run reports whether f succeeded
	// (or at least did not fail before calling t.Parallel). As all tests in
	// testAllNS are parallel, the deferred ctx.Cleanup above would be run before
	// all tests finished. Wrapping it in testAllNSPrometheus and testAllNSAlertmanager
	// fixes this.
	t.Run("x", testAllNSAlertmanager)
	t.Run("y", testAllNSPrometheus)
	t.Run("z", testAllNSThanosRuler)

	// Check if Prometheus Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"app.kubernetes.io/name": "prometheus-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if expected := 1; len(pl.Items) != expected {
		t.Fatalf("expected %v Prometheus Operator pods, but got %v", expected, len(pl.Items))
	}
	restarts, err := framework.GetPodRestartCount(context.Background(), ns, pl.Items[0].GetName())
	if err != nil {
		t.Fatalf("failed to retrieve restart count of Prometheus Operator pod: %v", err)
	}
	if len(restarts) != 1 {
		t.Fatalf("expected to have 1 container but got %d", len(restarts))
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

func testAllNSAlertmanager(t *testing.T) {
	skipAlertmanagerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AlertmanagerCRD":                         testAlertmanagerCRDValidation,
		"AMCreateDeleteCluster":                   testAMCreateDeleteCluster,
		"AMScaling":                               testAMScaling,
		"AMVersionMigration":                      testAMVersionMigration,
		"AMStorageUpdate":                         testAMStorageUpdate,
		"AMExposingWithKubernetesAPI":             testAMExposingWithKubernetesAPI,
		"AMClusterInitialization":                 testAMClusterInitialization,
		"AMClusterAfterRollingUpdate":             testAMClusterAfterRollingUpdate,
		"AMClusterGossipSilences":                 testAMClusterGossipSilences,
		"AMReloadConfig":                          testAMReloadConfig,
		"AMZeroDowntimeRollingDeployment":         testAMZeroDowntimeRollingDeployment,
		"AMAlertmanagerConfigCRD":                 testAlertmanagerConfigCRD,
		"AMAlertmanagerConfigVersions":            testAlertmanagerConfigVersions,
		"AMUserDefinedAMConfigFromSecret":         testUserDefinedAlertmanagerConfigFromSecret,
		"AMUserDefinedAMConfigFromCustomResource": testUserDefinedAlertmanagerConfigFromCustomResource,
		"AMPreserveUserAddedMetadata":             testAMPreserveUserAddedMetadata,
		"AMRollbackManualChanges":                 testAMRollbackManualChanges,
		"AMMinReadySeconds":                       testAlertManagerMinReadySeconds,
		"AMWeb":                                   testAMWeb,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

func testAllNSPrometheus(t *testing.T) {
	skipPrometheusAllNSTests(t)
	testFuncs := map[string]func(t *testing.T){
		"PrometheusCRDValidation":                   testPrometheusCRDValidation,
		"PromRemoteWriteWithTLS":                    testPromRemoteWriteWithTLS,
		"PromCreateDeleteCluster":                   testPromCreateDeleteCluster,
		"PromScaleUpDownCluster":                    testPromScaleUpDownCluster,
		"PromNoServiceMonitorSelector":              testPromNoServiceMonitorSelector,
		"PromVersionMigration":                      testPromVersionMigration,
		"PromResourceUpdate":                        testPromResourceUpdate,
		"PromStorageLabelsAnnotations":              testPromStorageLabelsAnnotations,
		"PromStorageUpdate":                         testPromStorageUpdate,
		"PromReloadConfig":                          testPromReloadConfig,
		"PromAdditionalScrapeConfig":                testPromAdditionalScrapeConfig,
		"PromAdditionalAlertManagerConfig":          testPromAdditionalAlertManagerConfig,
		"PromReloadRules":                           testPromReloadRules,
		"PromMultiplePrometheusRulesSameNS":         testPromMultiplePrometheusRulesSameNS,
		"PromMultiplePrometheusRulesDifferentNS":    testPromMultiplePrometheusRulesDifferentNS,
		"PromRulesExceedingConfigMapLimit":          testPromRulesExceedingConfigMapLimit,
		"PromRulesMustBeAnnotated":                  testPromRulesMustBeAnnotated,
		"PromtestInvalidRulesAreRejected":           testInvalidRulesAreRejected,
		"PromOnlyUpdatedOnRelevantChanges":          testPromOnlyUpdatedOnRelevantChanges,
		"PromWhenDeleteCRDCleanUpViaOwnerRef":       testPromWhenDeleteCRDCleanUpViaOwnerRef,
		"PromDiscovery":                             testPromDiscovery,
		"ShardingProvisioning":                      testShardingProvisioning,
		"Resharding":                                testResharding,
		"PromAlertmanagerDiscovery":                 testPromAlertmanagerDiscovery,
		"PromExposingWithKubernetesAPI":             testPromExposingWithKubernetesAPI,
		"PromDiscoverTargetPort":                    testPromDiscoverTargetPort,
		"PromOpMatchPromAndServMonInDiffNSs":        testPromOpMatchPromAndServMonInDiffNSs,
		"PromGetAuthSecret":                         testPromGetAuthSecret,
		"PromArbitraryFSAcc":                        testPromArbitraryFSAcc,
		"PromTLSConfigViaSecret":                    testPromTLSConfigViaSecret,
		"Thanos":                                    testThanos,
		"PromStaticProbe":                           testPromStaticProbe,
		"PromSecurePodMonitor":                      testPromSecurePodMonitor,
		"PromSharedResourcesReconciliation":         testPromSharedResourcesReconciliation,
		"PromPreserveUserAddedMetadata":             testPromPreserveUserAddedMetadata,
		"PromWebWithThanosSidecar":                  testPromWebWithThanosSidecar,
		"PromMinReadySeconds":                       testPromMinReadySeconds,
		"PromEnforcedNamespaceLabel":                testPromEnforcedNamespaceLabel,
		"PromNamespaceEnforcementExclusion":         testPromNamespaceEnforcementExclusion,
		"PromQueryLogFile":                          testPromQueryLogFile,
		"PromDegradedCondition":                     testPromDegradedConditionStatus,
		"PromUnavailableCondition":                  testPromUnavailableConditionStatus,
		"PromStrategicMergePatch":                   testPromStrategicMergePatch,
		"RelabelConfigCRDValidation":                testRelabelConfigCRDValidation,
		"PromReconcileStatusWhenInvalidRuleCreated": testPromReconcileStatusWhenInvalidRuleCreated,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

func testAllNSThanosRuler(t *testing.T) {
	skipThanosRulerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"ThanosRulerCreateDeleteCluster":                testThanosRulerCreateDeleteCluster,
		"ThanosRulerPrometheusRuleInDifferentNamespace": testThanosRulerPrometheusRuleInDifferentNamespace,
		"ThanosRulerPreserveUserAddedMetadata":          testTRPreserveUserAddedMetadata,
		"ThanosRulerMinReadySeconds":                    testTRMinReadySeconds,
		"ThanosRulerAlertmanagerConfig":                 testTRAlertmanagerConfig,
		"ThanosRulerQueryConfig":                        testTRQueryConfig,
	}
	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestMultiNS tests the Prometheus Operator configured to watch specific
// namespaces.
func TestMultiNS(t *testing.T) {
	skipPrometheusTests(t)
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
		"ThanosRuler":    testDenyThanosRuler,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestPromInstanceNs tests prometheus operator in different scenarios when --prometheus-instance-namespace is given
func TestPromInstanceNs(t *testing.T) {
	skipPrometheusTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AllNs":             testPrometheusInstanceNamespacesAllNs,
		"AllowList":         testPrometheusInstanceNamespacesAllowList,
		"DenyList":          testPrometheusInstanceNamespacesDenyList,
		"NamespaceNotFound": testPrometheusInstanceNamespacesNamespaceNotFound,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestAlertmanagerInstanceNs tests prometheus operator in different scenarios when --alertmanager-instance-namespace is given
func TestAlertmanagerInstanceNs(t *testing.T) {
	skipAlertmanagerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AllNs":     testAlertmanagerInstanceNamespacesAllNs,
		"AllowList": testAlertmanagerInstanceNamespacesAllowList,
		"DenyNs":    testAlertmanagerInstanceNamespacesDenyNs,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestOperatorUpgrade tests the prometheus upgrade from previous stable minor version to current version
func TestOperatorUpgrade(t *testing.T) {
	skipOperatorUpgradeTests(t)
	testFuncs := map[string]func(t *testing.T){
		"OperatorUpgrade": testOperatorUpgrade,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

const (
	prometheusOperatorServiceName = "prometheus-operator"
)

func testServerTLS(ctx context.Context, t *testing.T, namespace string) func(t *testing.T) {
	return func(t *testing.T) {
		skipPrometheusTests(t)
		if err := framework.WaitForServiceReady(context.Background(), namespace, prometheusOperatorServiceName); err != nil {
			t.Fatal("waiting for prometheus operator service: ", err)
		}

		operatorService := framework.KubeClient.CoreV1().Services(namespace)
		request := operatorService.ProxyGet("https", prometheusOperatorServiceName, "https", "/healthz", make(map[string]string))
		_, err := request.DoRaw(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}
