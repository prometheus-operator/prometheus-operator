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
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

var (
	previousVersionFramework *operatorFramework.Framework
	framework                *operatorFramework.Framework
	opImage                  *string
)

const (
	testControllerID            = "--controller-id=42"
	gitHubContentReleaseBaseURL = "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-%d.%d"
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

func skipPromVersionUpgradeTests(t *testing.T) {
	if os.Getenv("EXCLUDE_PROMETHEUS_UPGRADE_TESTS") != "" {
		t.Skip("Skipping Prometheus Version upgrade tests")
	}
}

func skipFeatureGatedTests(t *testing.T) {
	if os.Getenv("EXCLUDE_FEATURE_GATED_TESTS") != "" {
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

	prevStableVersionURL := fmt.Sprintf(gitHubContentReleaseBaseURL, currentSemVer.Major, currentSemVer.Minor-1) + "/VERSION"
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

	prevSemVer, err := semver.ParseTolerant(string(prevStableVersion))
	if err != nil {
		logger.Printf("failed to parse previous stable version: %v\n", err)
		os.Exit(1)
	}
	prevStableOpImage := fmt.Sprintf("quay.io/prometheus-operator/prometheus-operator:v%s", strings.TrimSpace(string(prevStableVersion)))
	prevExampleDir := fmt.Sprintf(gitHubContentReleaseBaseURL, prevSemVer.Major, prevSemVer.Minor) + "/example"
	prevResourcesDir := fmt.Sprintf(gitHubContentReleaseBaseURL, prevSemVer.Major, prevSemVer.Minor) + "/test/framework/resources"

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
	ctx := context.Background()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)

	finalizers, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx,
		operatorFramework.PrometheusOperatorOpts{
			Namespace:              ns,
			EnableAdmissionWebhook: true,
			ClusterRoleBindings:    true,
			EnableScrapeConfigs:    true,
		},
	)
	require.NoError(t, err)

	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	t.Run("TestServerTLS", func(t *testing.T) {
		testServerTLS(t, ns)
	})

	t.Run("TestPrometheusOperatorMetrics", func(t *testing.T) {
		t.Helper()
		testPrometheusOperatorMetrics(t, ns)
	})

	// t.Run blocks until the function passed as the second argument (f) returns or
	// calls t.Parallel to become a parallel test. Run reports whether f succeeded
	// (or at least did not fail before calling t.Parallel). As all tests in
	// testAllNS are parallel, the deferred ctx.Cleanup above would be run before
	// all tests finished. Wrapping it in testAllNSPrometheus and testAllNSAlertmanager
	// fixes this.
	t.Run("x", testAllNSAlertmanager)
	t.Run("y", testAllNSPrometheus)
	t.Run("z", testAllNSThanosRuler)
	t.Run("multipleOperators", testMultipleOperators(testCtx))

	// Check if Prometheus Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"app.kubernetes.io/name": "prometheus-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, pl.Items, 1, "expected %v Prometheus Operator pods, but got %v", 1, len(pl.Items))
	restarts, err := framework.GetPodRestartCount(ctx, ns, pl.Items[0].GetName())
	require.NoError(t, err)
	require.Len(t, restarts, 1)
	for _, restart := range restarts {
		require.Equal(t, int32(0), restart, "expected Prometheus Operator to never restart during entire test execution but got %d restarts", restart)
	}
}

func testAllNSAlertmanager(t *testing.T) {
	skipAlertmanagerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AlertmanagerConfigMatcherStrategy":       testAlertmanagerConfigMatcherStrategy,
		"AlertmanagerCRD":                         testAlertmanagerCRDValidation,
		"AMCreateDeleteCluster":                   testAMCreateDeleteCluster,
		"AMWithStatefulsetCreationFailure":        testAlertmanagerWithStatefulsetCreationFailure,
		"AMScalingReplicas":                       testAMScalingReplicas,
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
		"AMTemplateReloadConfig":                  testAMTmplateReloadConfig,
		"AMStatusScale":                           testAlertmanagerStatusScale,
		"AMServiceName":                           testAlertManagerServiceName,
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
		"PromScaleUpDownCluster":                    testPromScaleUpDownReplicas,
		"PromNoServiceMonitorSelector":              testPromNoServiceMonitorSelector,
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
		"ScrapeConfigCreation":                      testScrapeConfigCreation,
		"CreatePrometheusAgent":                     testCreatePrometheusAgent,
		"PrometheusAgentAndServerNameColision":      testAgentAndServerNameColision,
		"ScrapeConfigKubeNode":                      testScrapeConfigKubernetesNodeRole,
		"ScrapeConfigDNSSD":                         testScrapeConfigDNSSDConfig,
		"PrometheusWithStatefulsetCreationFailure":  testPrometheusWithStatefulsetCreationFailure,
		"PrometheusAgentCheckStorageClass":          testAgentCheckStorageClass,
		"PrometheusAgentStatusScale":                testPrometheusAgentStatusScale,
		"PrometheusStatusScale":                     testPrometheusStatusScale,
		"ScrapeConfigCRDValidations":                testScrapeConfigCRDValidations,
		"PrometheusServiceName":                     testPrometheusServiceName,
		"PrometheusAgentSSetServiceName":            testPrometheusAgentSSetServiceName,
		"PrometheusReconciliationOnSecretChanges":   testPrometheusReconciliationOnSecretChanges,
		"PrometheusUTF8MetricsSupport":              testPrometheusUTF8MetricsSupport,
		"PrometheusUTF8LabelSupport":                testPrometheusUTF8LabelSupport,
		"StuckStatefulSetRollout":                   testStuckStatefulSetRollout,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

func testAllNSThanosRuler(t *testing.T) {
	skipThanosRulerTests(t)
	testFuncs := map[string]func(t *testing.T){
		"ThanosRulerCreateDeleteCluster":                testThanosRulerCreateDeleteCluster,
		"ThanosRulerWithStatefulsetCreationFailure":     testThanosRulerWithStatefulsetCreationFailure,
		"ThanosRulerPrometheusRuleInDifferentNamespace": testThanosRulerPrometheusRuleInDifferentNamespace,
		"ThanosRulerPreserveUserAddedMetadata":          testTRPreserveUserAddedMetadata,
		"ThanosRulerMinReadySeconds":                    testTRMinReadySeconds,
		"ThanosRulerAlertmanagerConfig":                 testTRAlertmanagerConfig,
		"ThanosRulerQueryConfig":                        testTRQueryConfig,
		"ThanosRulerCheckStorageClass":                  testTRCheckStorageClass,
		"ThanosRulerServiceName":                        testThanosRulerServiceName,
		"ThanosRulerStateless":                          testThanosRulerStateless,
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

// TestPromInstanceNs tests prometheus operator in different scenarios when --prometheus-instance-namespace is given.
func TestPromInstanceNs(t *testing.T) {
	skipPrometheusTests(t)
	testFuncs := map[string]func(t *testing.T){
		"AllNs":                              testPrometheusInstanceNamespacesAllNs,
		"AllowList":                          testPrometheusInstanceNamespacesAllowList,
		"DenyList":                           testPrometheusInstanceNamespacesDenyList,
		"NamespaceNotFound":                  testPrometheusInstanceNamespacesNamespaceNotFound,
		"ScrapeConfigLifecycle":              testScrapeConfigLifecycle,
		"ScrapeConfigLifecycleInDifferentNs": testScrapeConfigLifecycleInDifferentNS,
		"ConfigReloaderResources":            testConfigReloaderResources,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestAlertmanagerInstanceNs tests prometheus operator in different scenarios when --alertmanager-instance-namespace is given.
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

// TestOperatorUpgrade tests the prometheus upgrade from previous stable minor version to current version.
func TestOperatorUpgrade(t *testing.T) {
	skipOperatorUpgradeTests(t)
	testFuncs := map[string]func(t *testing.T){
		"OperatorUpgrade":                          testOperatorUpgrade,
		"PromOperatorStartsWithoutScrapeConfigCRD": testPromOperatorStartsWithoutScrapeConfigCRD,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

const (
	prometheusOperatorServiceName = "prometheus-operator"
)

// TestGatedFeatures tests features that are behind feature gates.
func TestGatedFeatures(t *testing.T) {
	skipFeatureGatedTests(t)
	testFuncs := map[string]func(t *testing.T){
		"CreatePrometheusAgentDaemonSet":                       testCreatePrometheusAgentDaemonSet,
		"PromAgentDaemonSetResourceUpdate":                     testPromAgentDaemonSetResourceUpdate,
		"PromAgentReconcileDaemonSetResourceUpdate":            testPromAgentReconcileDaemonSetResourceUpdate,
		"PromAgentReconcileDaemonSetResourceDelete":            testPromAgentReconcileDaemonSetResourceDelete,
		"PrometheusAgentDaemonSetSelectPodMonitor":             testPrometheusAgentDaemonSetSelectPodMonitor,
		"PrometheusRetentionPolicies":                          testPrometheusRetentionPolicies,
		"FinalizerWhenStatusForConfigResourcesEnabled":         testFinalizerWhenStatusForConfigResourcesEnabled,
		"PrometheusAgentDaemonSetCELValidations":               testPrometheusAgentDaemonSetCELValidations,
		"ServiceMonitorStatusSubresource":                      testServiceMonitorStatusSubresource,
		"ServiceMonitorStatusWithMultipleWorkloads":            testServiceMonitorStatusWithMultipleWorkloads,
		"GarbageCollectionOfServiceMonitorBinding":             testGarbageCollectionOfServiceMonitorBinding,
		"RmServiceMonitorBindingDuringWorkloadDelete":          testRmServiceMonitorBindingDuringWorkloadDelete,
		"PodMonitorStatusSubresource":                          testPodMonitorStatusSubresource,
		"GarbageCollectionOfPodMonitorBinding":                 testGarbageCollectionOfPodMonitorBinding,
		"RmPodMonitorBindingDuringWorkloadDelete":              testRmPodMonitorBindingDuringWorkloadDelete,
		"ProbeStatusSubresource":                               testProbeStatusSubresource,
		"GarbageCollectionOfProbeBinding":                      testGarbageCollectionOfProbeBinding,
		"RmProbeBindingDuringWorkloadDelete":                   testRmProbeBindingDuringWorkloadDelete,
		"ScrapeConfigStatusSubresource":                        testScrapeConfigStatusSubresource,
		"GarbageCollectionOfScrapeConfigBinding":               testGarbageCollectionOfScrapeConfigBinding,
		"RmScrapeConfigBindingDuringWorkloadDelete":            testRmScrapeConfigBindingDuringWorkloadDelete,
		"PrometheusRuleStatusSubresource":                      testPrometheusRuleStatusSubresource,
		"FinalizerForPromAgentWhenStatusForConfigResEnabled":   testFinalizerForPromAgentWhenStatusForConfigResEnabled,
		"GarbageCollectionOfPrometheusRuleBinding":             testGarbageCollectionOfPrometheusRuleBinding,
		"RmPrometheusRuleBindingDuringWorkloadDelete":          testRmPrometheusRuleBindingDuringWorkloadDelete,
		"FinalizerForThanosRulerWhenStatusForConfigResEnabled": testFinalizerForThanosRulerWhenStatusForConfigResEnabled,
		"PrometheusRuleStatusSubresourceForThanosRuler":        testPrometheusRuleStatusSubresourceForThanosRuler,
		"GarbageCollectionOfPromRuleBindingForThanosRuler":     testGarbageCollectionOfPromRuleBindingForThanosRuler,
		"RmPromeRuleBindingDuringWorkloadDeleteForThanosRuler": testRmPromeRuleBindingDuringWorkloadDeleteForThanosRuler,
		"AlertmanagerConfigStatusSubresource":                  testAlertmanagerConfigStatusSubresource,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}

// TestPrometheusVersionUpgrade tests that all Prometheus versions in the compatibility matrix can be upgraded.
func TestPrometheusVersionUpgrade(t *testing.T) {
	skipPromVersionUpgradeTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	finalizers, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, nil, nil, nil, nil, true, true, true)
	require.NoError(t, err)

	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	t.Run("PromVersionMigration", testPromVersionMigration)
}

// testMultipleOperators checks that multiple Prometheus operator instances can
// run concurrently without stepping on each other toes when properly
// configured with ControllerID.
func testMultipleOperators(testCtx *operatorFramework.TestCtx) func(t *testing.T) {
	return func(t *testing.T) {
		skipPrometheusTests(t)

		ns := framework.CreateNamespace(context.Background(), t, testCtx)
		// Create operator-2 in a new ns and set controller-id.
		finalizers, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(context.Background(),
			operatorFramework.PrometheusOperatorOpts{
				Namespace:           ns,
				ClusterRoleBindings: true,
				EnableScrapeConfigs: true,
				AdditionalArgs:      []string{testControllerID},
			})
		require.NoError(t, err)

		for _, f := range finalizers {
			testCtx.AddFinalizerFn(f)
		}

		testFuncs := map[string]func(t *testing.T){
			"PrometheusServer": testMultipleOperatorsPrometheusServer,
			"PrometheusAgent":  testMultipleOperatorsPrometheusAgent,
			"AlertManager":     testMultipleOperatorsAlertManager,
			"ThanosRuler":      testMultipleOperatorsThanosRuler,
		}
		for name, f := range testFuncs {
			t.Run(name, f)
		}
	}
}
