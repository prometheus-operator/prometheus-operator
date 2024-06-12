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
	"time"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

// testScrapeConfigCreation tests multiple ScrapeConfig definitions.
func testScrapeConfigCreation(t *testing.T) {
	skipPrometheusTests(t)
	t.Parallel()

	fiveMins := monitoringv1.Duration("5m")

	tests := []struct {
		name          string
		spec          monitoringv1alpha1.ScrapeConfigSpec
		expectedError bool
	}{
		{
			name: "empty-scrape-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{},
		},
		{
			name: "static-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				StaticConfigs: []monitoringv1alpha1.StaticConfig{
					{
						Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090"},
						Labels: map[monitoringv1.LabelName]string{
							"label1": "value1",
							"label2": "value2",
						},
					},
				},
			},
		},
		{
			name: "http-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
					{
						URL:             "http://localhost:8080/file.json",
						RefreshInterval: &fiveMins,
					},
				},
			},
		},
		{
			name: "file-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
					{
						Files: []monitoringv1alpha1.SDFile{
							"/etc/prometheus/sd/file.json",
							"/etc/prometheus/sd/file.yaml",
						},
						RefreshInterval: &fiveMins,
					},
				},
			},
		},
		{
			name: "kubernetes-sd-config-node-role",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.Role("Node"),
					},
				},
			},
		},
		{
			name: "dns-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{
							"demo.do.prometheus.io",
						},
						RefreshInterval: &fiveMins,
						Type:            ptr.To("A"),
						Port:            ptr.To(9090),
					},
				},
			},
		},
		{
			name: "invalid-sd-config",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
					{
						Files: []monitoringv1alpha1.SDFile{
							"/etc/prometheus/sd/file.invalid",
						},
						RefreshInterval: &fiveMins,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)

			sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
			sc.Spec = test.spec
			_, err := framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(context.Background(), sc, metav1.CreateOptions{})

			if test.expectedError {
				require.Error(t, err)
				require.Truef(t, apierrors.IsInvalid(err), "expected Invalid error but got %v", err)
				return
			}
			require.NoError(t, err)
			require.Falsef(t, test.expectedError, "expected error but got nil")
		})
	}
}

// testScrapeConfigLifecycle tests 3 things:
// 1. Creating a ScrapeConfig and checking that 2 targets appear in Prometheus
// 2. Updating that ScrapeConfig by adding a target and checking that 3 targets appear in Prometheus
// 3. Deleting that ScrapeConfig and checking that 0 targets appear in Prometheus.
func testScrapeConfigLifecycle(t *testing.T) {
	skipPrometheusTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		ns,
		[]string{ns},
		nil,
		[]string{ns},
		nil,
		false,
		true, // clusterrole
		true,
	)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "prom", "group", 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"role": "scrapeconfig",
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// 1. Create a ScrapeConfig and check that its targets appear in Prometheus
	sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090"},
		},
	}
	_, err = framework.CreateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 2)
	require.NoError(t, err)

	// 2. Update the ScrapeConfig and add a target. Then, check that 3 targets appear in Prometheus.
	sc, err = framework.GetScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090", "target3:9090"},
		},
	}

	_, err = framework.UpdateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 3)
	require.NoError(t, err)

	// 3. Remove the ScrapeConfig and check that the targets disappear in Prometheus
	err = framework.DeleteScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 0)
	require.NoError(t, err)
}

func testScrapeConfigLifecycleInDifferentNS(t *testing.T) {
	skipPrometheusTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// The ns where the prometheus CR will reside
	promns := framework.CreateNamespace(context.Background(), t, testCtx)
	// The ns where the scrapeConfig will reside
	scns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, promns)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, scns)

	_, err := framework.CreateOrUpdatePrometheusOperator(
		context.Background(),
		promns,
		[]string{scns},
		nil,
		[]string{promns},
		nil,
		false,
		true, // clusterrole
		true,
	)
	require.NoError(t, err)

	// Make a prometheus object in promns which will select any ScrapeConfig resource with
	// "group": "sc" and/or "kubernetes.io/metadata.name": <scns>
	p := framework.MakeBasicPrometheus(promns, "prom", scns, 1)
	p.Spec.ScrapeConfigNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"kubernetes.io/metadata.name": scns,
		},
	}

	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": "sc",
		},
	}

	// Make the Prometheus selection surface thin
	p.Spec.PodMonitorSelector = nil
	p.Spec.PodMonitorNamespaceSelector = nil
	p.Spec.ServiceMonitorSelector = nil
	p.Spec.ServiceMonitorNamespaceSelector = nil
	p.Spec.RuleSelector = nil
	p.Spec.RuleNamespaceSelector = nil

	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), promns, p)
	require.NoError(t, err)

	// 1. Create a ScrapeConfig in scns and check that its targets appear in Prometheus
	sc := framework.MakeBasicScrapeConfig(scns, "scrape-config")
	sc.ObjectMeta.Labels = map[string]string{
		"group": "sc"}

	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090"},
		},
	}
	_, err = framework.CreateScrapeConfig(context.Background(), scns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), promns, "prometheus-operated", 2)
	require.NoError(t, err)

	// 2. Update the ScrapeConfig and add a target. Then, check that 3 targets appear in Prometheus.
	sc, err = framework.GetScrapeConfig(context.Background(), scns, "scrape-config")
	require.NoError(t, err)

	sc.Spec.StaticConfigs = []monitoringv1alpha1.StaticConfig{
		{
			Targets: []monitoringv1alpha1.Target{"target1:9090", "target2:9090", "target3:9090"},
		},
	}

	_, err = framework.UpdateScrapeConfig(context.Background(), scns, sc)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), promns, "prometheus-operated", 3)
	require.NoError(t, err)

	// 3. Remove the ScrapeConfig and check that the targets disappear in Prometheus
	err = framework.DeleteScrapeConfig(context.Background(), scns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), promns, "prometheus-operated", 0)
	require.NoError(t, err)
}

// testPromOperatorStartsWithoutScrapeConfigCRD deletes the ScrapeConfig CRD from the cluster and then starts
// prometheus-operator to check that it doesn't crash.
func testPromOperatorStartsWithoutScrapeConfigCRD(t *testing.T) {
	skipPrometheusAllNSTests(t)

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	err := framework.DeleteCRD(context.Background(), "scrapeconfigs.monitoring.coreos.com")
	require.NoError(t, err)

	_, err = framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, false)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	// Check if Prometheus Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"app.kubernetes.io/name": "prometheus-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), opts)
	require.NoError(t, err)
	require.Lenf(t, pl.Items, 1, "expected 1 Prometheus Operator pods, but got %v", len(pl.Items))

	restarts, err := framework.GetPodRestartCount(context.Background(), ns, pl.Items[0].GetName())
	require.NoError(t, err)
	require.Emptyf(t, restarts, "expected to get 1 container but got %d", len(restarts))

	for _, restart := range restarts {
		require.Emptyf(t, restart, "expected Prometheus Operator to never restart during entire test execution but got %d restarts", restart)
	}

	// re-create Prometheus-Operator to reinstall the CRDs
	_, err = framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, true)
	require.NoError(t, err)
}

// testScrapeConfigKubernetesNodeRole tests whether Kubernetes node monitoring works as expected.
func testScrapeConfigKubernetesNodeRole(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	// Create cluster role and cluster role binding for "prometheus" service account
	// so that it has access to 'node' resource cluster scope
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, ns)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, true)
	require.NoError(t, err)

	// For prometheus to be able to scrape nodes it needs to able to authenticate
	// using mTLS certificates issued for the ServiceAccount "prometheus"
	secretName := "scraping-tls"
	createServiceAccountSecret(t, "prometheus", ns)
	createMutualTLSSecret(t, secretName, ns)

	sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
	sc.Spec.Scheme = ptr.To("HTTPS")
	sc.Spec.Authorization = &monitoringv1.SafeAuthorization{
		Credentials: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "prometheus-sa-secret",
			},
			Key: "token",
		},
	}
	sc.Spec.TLSConfig = &monitoringv1.SafeTLSConfig{
		// since we cannot validate server name in cert
		InsecureSkipVerify: ptr.To(true),
		CA: monitoringv1.SecretOrConfigMap{
			Secret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: "ca.crt",
			},
		},
		Cert: monitoringv1.SecretOrConfigMap{
			Secret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: "cert.pem",
			},
		},
		KeySecret: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: secretName,
			},
			Key: "key.pem",
		},
	}

	sc.Spec.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
		{
			Role: monitoringv1alpha1.Role("Node"),
		},
	}
	_, err = framework.CreateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "prom", "group", 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"role": "scrapeconfig",
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus and does proper scrapping
	nodes, err := framework.Nodes(context.Background())
	require.NoError(t, err)

	err = framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", len(nodes))
	require.NoError(t, err)

	// Remove the ScrapeConfig
	err = framework.DeleteScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 0)
	require.NoError(t, err)
}

// testScrapeConfigDNSSDConfig tests whether DNS SD based monitoring works as expected.
func testScrapeConfigDNSSDConfig(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, []string{ns}, nil, []string{ns}, nil, false, true, true)
	require.NoError(t, err)

	sc := framework.MakeBasicScrapeConfig(ns, "scrape-config")
	sc.Spec.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
		{
			Names: []string{"node.demo.do.prometheus.io"},
			Type:  ptr.To("A"),
			Port:  ptr.To(9100),
		},
	}
	_, err = framework.CreateScrapeConfig(context.Background(), ns, sc)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "prom", "group", 1)
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"role": "scrapeconfig",
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// Check that the targets appear in Prometheus and does proper scrapping
	if err := framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1); err != nil {
		t.Fatal(err)
	}

	// Remove the ScrapeConfig
	err = framework.DeleteScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 0)
	require.NoError(t, err)
}
