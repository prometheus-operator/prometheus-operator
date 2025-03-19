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
						Labels: map[string]string{
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
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  "Pod",
								Label: ptr.To("component=executor"),
							},
						},
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
						Type:            ptr.To(monitoringv1alpha1.DNSRecordType("A")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
		},
		{
			name: "invalid-dns-sd-config-with-empty-name",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
					{
						Names:           []string{""},
						RefreshInterval: &fiveMins,
						Type:            ptr.To(monitoringv1alpha1.DNSRecordType("A")),
						Port:            ptr.To(int32(9100)),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid-scaleway-sd-config-with-empty-tagfilter",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "ak",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key.pem",
						},
						ProjectID:  "1",
						Role:       monitoringv1alpha1.ScalewayRoleInstance,
						TagsFilter: []string{}, // empty
					},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid-scaleway-sd-config-with-empty-string-tagfilter",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "ak",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key.pem",
						},
						ProjectID:  "1",
						Role:       monitoringv1alpha1.ScalewayRoleInstance,
						TagsFilter: []string{""},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid-scaleway-sd-config-tagfilter-items-repeat",
			spec: monitoringv1alpha1.ScrapeConfigSpec{
				ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "ak",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key.pem",
						},
						ProjectID:  "1",
						Role:       monitoringv1alpha1.ScalewayRoleInstance,
						TagsFilter: []string{"do", "do"}, // repeat
					},
				},
			},
			expectedError: true,
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
		"group": "sc",
	}

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
			Role: monitoringv1alpha1.KubernetesRoleNode,
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
	t.Skip("DNS service discovery tests are disabled until we find a replacement for node.demo.do.prometheus.io")

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
			Type:  ptr.To(monitoringv1alpha1.DNSRecordType("A")),
			Port:  ptr.To(int32(9100)),
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
	err = framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1)
	require.NoError(t, err)

	// Remove the ScrapeConfig
	err = framework.DeleteScrapeConfig(context.Background(), ns, "scrape-config")
	require.NoError(t, err)

	// Check that the targets disappeared in Prometheus
	err = framework.WaitForActiveTargets(context.Background(), ns, "prometheus-operated", 0)
	require.NoError(t, err)
}

type scrapeCRDTestCase struct {
	name             string
	scrapeConfigSpec monitoringv1alpha1.ScrapeConfigSpec
	expectedError    bool
}

func testScrapeConfigCRDValidations(t *testing.T) {
	t.Parallel()
	t.Run("ScrapeConfig", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, ScrapeConfigCRDTestCases)
	})
	t.Run("KubernetesSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, K8STestCases)
	})
	t.Run("DNSSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, DNSSDTestCases)
	})
	t.Run("EC2SD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, EC2SDTestCases)
	})
	t.Run("StaticConfig", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, staticConfigTestCases)
	})
	t.Run("ConsulSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, ConsulSDTestCases)
	})
	t.Run("FileSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, FileSDTestCases)
	})
	t.Run("HTTPSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, HTTPSDTestCases)
	})
	t.Run("DigitalOceanSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, DigitalOceanSDTestCases)
	})
	t.Run("IonosSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, IonosSDTestCases)
	})
	t.Run("LightSailSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, LightSailSDTestCases)
	})
	t.Run("AzureSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, AzureSDTestCases)
	})
	t.Run("OVHCloudSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, OVHCloudSDTestCases)
	})
	t.Run("GCESD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, GCESDTestCases)
	})
	t.Run("OpenStackSD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, OpenStackSDTestCases)
	})
	t.Run("ScalewaySD", func(t *testing.T) {
		runScrapeConfigCRDValidation(t, ScalewaySDTestCases)
	})
}

func runScrapeConfigCRDValidation(t *testing.T, testCases []scrapeCRDTestCase) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			sc := &monitoringv1alpha1.ScrapeConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   ns,
					Annotations: map[string]string{},
				},
				Spec: test.scrapeConfigSpec,
			}

			_, err := framework.MonClientV1alpha1.ScrapeConfigs(ns).Create(context.Background(), sc, metav1.CreateOptions{})
			if test.expectedError {
				require.True(t, apierrors.IsInvalid(err))
				return
			}

			require.NoError(t, err)
		})
	}
}

var ConsulSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid Server",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Server with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid missing Server",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid PathPrefix",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:     "valid-server",
					PathPrefix: ptr.To("valid-server"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid PathPrefix with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:     "valid-server",
					PathPrefix: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing PathPrefix",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Datacenter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:     "valid-server",
					Datacenter: ptr.To("valid-server"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Datacenter with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:     "valid-server",
					Datacenter: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Datacenter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Namespace",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:    "valid-server",
					Namespace: ptr.To("valid-server"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Namespace with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:    "valid-server",
					Namespace: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Namespace",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Partition",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:    "valid-server",
					Partition: ptr.To("valid-server"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Partition with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:    "valid-server",
					Partition: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Partition",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid HTTP Scheme",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
					Scheme: ptr.To("HTTP"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid HTTPS Scheme",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
					Scheme: ptr.To("HTTPS"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Scheme with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
					Scheme: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Scheme",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Services",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:   "valid-server",
					Services: []string{"foo", "bar"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Services with repeating values",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:   "valid-server",
					Services: []string{"foo", "foo"},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Services",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Services with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:   "valid-server",
					Services: []string{},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Tags",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
					Tags:   []string{"foo", "bar"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Tags with repeating values",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
					Tags:   []string{"foo", "foo"},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing Tags",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid TagSeparator",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:       "valid-server",
					TagSeparator: ptr.To(","),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid TagSeparator with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server:       "valid-server",
					TagSeparator: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid missing TagSeparator",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ConsulSDConfigs: []monitoringv1alpha1.ConsulSDConfig{
				{
					Server: "valid-server",
				},
			},
		},
		expectedError: false,
	},
}

var HTTPSDTestCases = []scrapeCRDTestCase{
	{
		name: "Invalid URL",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{
					URL: "valid-server",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid empty URL",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{
					URL: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid absent URL",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid URL with http scheme",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{
					URL: "http://valid.test",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid URL with https scheme",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			HTTPSDConfigs: []monitoringv1alpha1.HTTPSDConfig{
				{
					URL: "https://valid-url",
				},
			},
		},
		expectedError: false,
	},
}

var K8STestCases = []scrapeCRDTestCase{
	{
		name: "APIServer with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role:      "EndpointSlice",
					APIServer: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing required Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					// Role is missing
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Wrong",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Role with empty APIServer",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role:      "Pod",
					APIServer: nil,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Namespace discovery with valid namespace",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role:       "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{Names: []string{"default"}},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Selector Role missing",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							// Role is missing
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Selector Role valid",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role: "Pod",
						},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Selector Label with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Label: ptr.To(""),
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Selector Label with valid value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Label: ptr.To("node.kubernetes.io/instance-type=master"),
						},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Selector Field with empty value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Field: ptr.To(""),
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Selector Field with valid value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Field: ptr.To("metadata.name=foobar"),
						},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Selector Field with valid value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Field: ptr.To("metadata.name=foobar"),
						},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Selector Label and Field with duplicate values",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Selectors: []monitoringv1alpha1.K8SSelectorConfig{
						{
							Role:  "Pod",
							Label: ptr.To("node.kubernetes.io/instance-type=master"),
							Field: ptr.To("metadata.name=foobar"),
						},
						{
							Role:  "Pod",
							Label: ptr.To("node.kubernetes.io/instance-type=master"),
							Field: ptr.To("metadata.name=foobar"),
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "IncludeOwnNamespace set to true",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						IncludeOwnNamespace: ptr.To(true),
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "IncludeOwnNamespace set to false with empty Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						IncludeOwnNamespace: ptr.To(false),
						Names:               []string{},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "IncludeOwnNamespace unset with empty Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						Names: []string{},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Names with valid namespaces",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						Names: []string{"default", "kube-system"},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "IncludeOwnNamespace set to true with valid Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						IncludeOwnNamespace: ptr.To(true),
						Names:               []string{"default", "kube-system"},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "IncludeOwnNamespace set to true with repeated Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			KubernetesSDConfigs: []monitoringv1alpha1.KubernetesSDConfig{
				{
					Role: "Pod",
					Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
						IncludeOwnNamespace: ptr.To(true),
						Names:               []string{"default", "default"},
					},
				},
			},
		},
		expectedError: true,
	},
}

var DNSSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1", "test2"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Missing Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Empty Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Empty string in Names",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{""},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Record Type A",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Record Type AAAA",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeAAAA),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Record Type MX",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeMX),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Record Type NS",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeNS),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Record Type SRV",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeSRV),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Record Type",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Type:  ptr.To(monitoringv1alpha1.DNSRecordType("WRONG")),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port Number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Port:  ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port Number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names: []string{"test1"},
					Port:  ptr.To(int32(80809)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names:           []string{"test1"},
					RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DNSSDConfigs: []monitoringv1alpha1.DNSSDConfig{
				{
					Names:           []string{"test1"},
					RefreshInterval: ptr.To(monitoringv1.Duration("30g")),
				},
			},
		},
		expectedError: true,
	},
}

var EC2SDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid AWS Region",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To("us-west"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Absent AWS Region",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AWS Region",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AWS RoleARN",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					RoleARN: ptr.To("valid-role"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Absent AWS RoleARN",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AWS RoleARN",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					RoleARN: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port Number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Port: ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port Number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Port: ptr.To(int32(80809)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Filters",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To("us-west"),
					Filters: []monitoringv1alpha1.Filter{
						{
							Name:   "foo",
							Values: []string{"bar"},
						},
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Filters with repeat value items",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To("us-west"),
					Filters: []monitoringv1alpha1.Filter{
						{
							Name:   "foo",
							Values: []string{"bar", "bar"},
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Filters with empty values",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To("us-west"),
					Filters: []monitoringv1alpha1.Filter{
						{
							Name:   "foo",
							Values: []string{},
						},
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Filters with empty string values",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			EC2SDConfigs: []monitoringv1alpha1.EC2SDConfig{
				{
					Region: ptr.To("us-west"),
					Filters: []monitoringv1alpha1.Filter{
						{
							Name:   "foo",
							Values: []string{""},
						},
					},
				},
			},
		},
		expectedError: true,
	},
}

var ScrapeConfigCRDTestCases = []scrapeCRDTestCase{
	{
		name:             "JobName: Not Specified",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{},
		expectedError:    false,
	},
	{
		name: "JobName: Empty String",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			JobName: ptr.To(""),
		},
		expectedError: true,
	},
	{
		name: "JobName: Valid Value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			JobName: ptr.To("validJob"),
		},
		expectedError: false,
	},
	{
		name:             "Scheme: Not Specified",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{},
		expectedError:    false,
	},
	{
		name: "Scheme: Invalid Value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			Scheme: ptr.To("FTP"),
		},
		expectedError: true,
	},
	{
		name: "Scheme: Valid Value HTTP",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			Scheme: ptr.To("HTTP"),
		},
		expectedError: false,
	},
	{
		name: "Scheme: Valid Value HTTPS",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			Scheme: ptr.To("HTTPS"),
		},
		expectedError: false,
	},
	{
		name:             "ScrapeClassName: Not Specified",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{},
		expectedError:    false,
	},
	{
		name: "ScrapeClassName: Empty String",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeClassName: ptr.To(""),
		},
		expectedError: true,
	},
	{
		name: "ScrapeClassName: Valid Value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeClassName: ptr.To("default"),
		},
		expectedError: false,
	},
	{
		name:             "ScrapeProtocols: Not Specified",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{},
		expectedError:    false,
	},
	{
		name: "ScrapeProtocols: Single Valid Protocol",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"PrometheusProto",
			},
		},
		expectedError: false,
	},
	{
		name: "ScrapeProtocols: Multiple Valid Protocols",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"OpenMetricsText0.0.1",
				"OpenMetricsText1.0.0",
			},
		},
		expectedError: false,
	},
	{
		name: "ScrapeProtocols: Invalid Protocol",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"InvalidProtocol",
			},
		},
		expectedError: true,
	},
	{
		name: "ScrapeProtocols: Mixed Valid and Invalid Protocols",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"PrometheusText0.0.4",
				"InvalidProtocol",
			},
		},
		expectedError: true,
	},
	{
		name: "ScrapeProtocols: Empty List",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{},
		},
		expectedError: false,
	},
	{
		name: "ScrapeProtocols: Duplicate Valid Protocols",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"OpenMetricsText0.0.1",
				"OpenMetricsText0.0.1",
			},
		},
		expectedError: true,
	},
	{
		name: "ScrapeProtocols: All Valid Protocols",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScrapeProtocols: []monitoringv1.ScrapeProtocol{
				"PrometheusProto",
				"OpenMetricsText0.0.1",
				"OpenMetricsText1.0.0",
				"PrometheusText0.0.4",
			},
		},
		expectedError: false,
	},
	{
		name: "FallbackScrapeProtocol: Valid Value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FallbackScrapeProtocol: ptr.To(monitoringv1.OpenMetricsText0_0_1),
		},
		expectedError: false,
	},
	{
		name: "FallbackScrapeProtocol: Invalid Protocol",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FallbackScrapeProtocol: ptr.To(monitoringv1.ScrapeProtocol("InvalidProtocol")),
		},
		expectedError: true,
	},
	{
		name: "FallbackScrapeProtocol: Setting nil",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FallbackScrapeProtocol: nil,
		},
		expectedError: false,
	},
}

var staticConfigTestCases = []scrapeCRDTestCase{
	{
		name: "Valid targets",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{

			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{
					Targets: []monitoringv1alpha1.Target{"1.1.1.1:9090", "0.0.0.0:9090"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid absent targets",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{

			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid duplicate targets",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{

			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{
					Targets: []monitoringv1alpha1.Target{"1.1.1.1:9090", "1.1.1.1:9090"},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid empty targets",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{

			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{
					Targets: []monitoringv1alpha1.Target{},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid labels",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{

			StaticConfigs: []monitoringv1alpha1.StaticConfig{
				{
					Targets: []monitoringv1alpha1.Target{"1.1.1.1:9090", "0.0.0.0:9090"},
					Labels:  map[string]string{"owned-by": "prometheus"},
				},
			},
		},
		expectedError: false,
	},
}

var FileSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid files list",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
				{
					Files: []monitoringv1alpha1.SDFile{"config.yml", "config.yaml"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid duplicate files list",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
				{
					Files: []monitoringv1alpha1.SDFile{"config.yml", "config.yml"},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid absent files list",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid empty files list",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			FileSDConfigs: []monitoringv1alpha1.FileSDConfig{
				{
					Files: []monitoringv1alpha1.SDFile{},
				},
			},
		},
		expectedError: true,
	},
}

var DigitalOceanSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid Port number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
				{
					Port: ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port number exceeding the maximum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
				{
					Port: ptr.To(int32(65536)), // maximum Port number = 65535
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Port number below the minimum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
				{
					Port: ptr.To(int32(-1)), // minimum Port number = 0;
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
				{
					RefreshInterval: ptr.To(monitoringv1.Duration("60s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			DigitalOceanSDConfigs: []monitoringv1alpha1.DigitalOceanSDConfig{
				{
					RefreshInterval: ptr.To(monitoringv1.Duration("60g")),
				},
			},
		},
		expectedError: true,
	},
}

var IonosSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid DataCeneterID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{
					DataCenterID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid empty DataCenterID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{
					DataCenterID: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid missing DataCenterID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{
					DataCenterID: "11111111-1111-1111-1111-111111111111",
					Port:         ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port number exceeeding the maximum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{
					DataCenterID: "11111111-1111-1111-1111-111111111111",
					Port:         ptr.To(int32(65536)), // maximum Port number = 65535
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Port number below the minimum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			IonosSDConfigs: []monitoringv1alpha1.IonosSDConfig{
				{
					DataCenterID: "11111111-1111-1111-1111-111111111111",
					Port:         ptr.To(int32(-1)), // minimum Port number = 0
				},
			},
		},
		expectedError: true,
	},
}

var LightSailSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid RegionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Region: ptr.To("us-east-1"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid RegionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Region: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Endpoint",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Endpoint: ptr.To("https://custom-endpoint.example.com"),
				},
			},
		},
		expectedError: false,
	},

	{
		name: "Invalid Endpoint",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Endpoint: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Port: ptr.To(int32(80)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port 1",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Port: ptr.To(int32(-1)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Port 2",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			LightSailSDConfigs: []monitoringv1alpha1.LightSailSDConfig{
				{
					Port: ptr.To(int32(65536)),
				},
			},
		},
		expectedError: true,
	},
}

var GCESDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid GCESDConfig required options",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project: "devops-dev",
					Zone:    "us-west-1",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid GCESDConfig with empty Project",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid GCESDConfig with empty Zone",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Zone: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid GCESDConfig with empty required fields",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid GCESDConfig with refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project:         "devops-dev",
					Zone:            "us-west-1",
					RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid GCESDConfig with wrong refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					RefreshInterval: ptr.To(monitoringv1.Duration("30g")),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid GCESDConfig with Port",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project: "devops-dev",
					Zone:    "us-west-1",
					Port:    ptr.To(int32(65534)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid GCESDConfig with wrong Port",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Port: ptr.To(int32(-1)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid GCESDConfig with Filter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project: "devops-dev",
					Zone:    "us-west-1",
					Filter:  ptr.To("filter-expression"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid GCESDConfig with wrong Filter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Filter: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid GCESDConfig with TagSeparator",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					Project:      "devops-dev",
					Zone:         "us-west-1",
					TagSeparator: ptr.To("tag-value"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid GCESDConfig with TagSeparator",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			GCESDConfigs: []monitoringv1alpha1.GCESDConfig{
				{
					TagSeparator: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
}

var AzureSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid AzureSD With SubscriptionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD With Missing SubscriptionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					Environment:          ptr.To("AzurePublicCloud"),
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid AzureSD With Empty SubscriptionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid AzureSD With No SubscriptionID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AzureSD With Environment",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					Environment:    ptr.To("AzurePublicCloud"),
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD With Empty Environment",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					Environment:    ptr.To(""),
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AzureSD With ResourceGroup",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					ResourceGroup:  ptr.To("my-resource-group"),
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD With Empty ResourceGroup",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					ResourceGroup:  ptr.To(""),
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AzureSD With ManagedIdentity Authentication",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeManagedIdentity),
					SubscriptionID:       "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid AzureSD With SDK Authentication",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
					SubscriptionID:       "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid AzureSD With OAuth Authentication And All Required Fields",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
					TenantID:             ptr.To("22222222-2222-2222-2222-222222222222"),
					ClientID:             ptr.To("33333333-3333-3333-3333-333333333333"),
					SubscriptionID:       "11111111-1111-1111-1111-111111111111",
					ClientSecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "azure-client-secret",
						},
						Key: "clientSecret",
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD With OAuth Authentication and Empty TenantID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
					TenantID:             ptr.To(""),
					SubscriptionID:       "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid AzureSD With OAuth Authentication and Empty ClientID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
					ClientID:             ptr.To(""),
					SubscriptionID:       "11111111-1111-1111-1111-111111111111",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AzureSD With RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID:  "11111111-1111-1111-1111-111111111111",
					RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD with Invalid RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID:  "11111111-1111-1111-1111-111111111111",
					RefreshInterval: ptr.To(monitoringv1.Duration("30g")),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid AzureSD With Valid Port",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
					Port:           ptr.To(int32(65534)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid AzureSD With Invalid Port",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			AzureSDConfigs: []monitoringv1alpha1.AzureSDConfig{
				{
					SubscriptionID: "11111111-1111-1111-1111-111111111111",
					Port:           ptr.To(int32(-1)),
				},
			},
		},
		expectedError: true,
	},
}

var OVHCloudSDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid OVHCloudSDConfig",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceDedicatedServer,
					Endpoint:          ptr.To("https://api.ovh.com/endpoint"),
					RefreshInterval:   ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid OVHCloudSDConfig without optional fields",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceDedicatedServer,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid ApplicationKey",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Empty ApplicationKey",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing ApplicationKey",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Service field (VPS)",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Service field (DedicatedServer)",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceDedicatedServer,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Service type",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           "InvalidService",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Service field",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid ConsumerKey",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Empty Endpoint",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
					Endpoint:          ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Empty RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
					RefreshInterval:   ptr.To(monitoringv1.Duration("")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid RefreshInterval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OVHCloudSDConfigs: []monitoringv1alpha1.OVHCloudSDConfig{
				{
					ApplicationKey:    "valid-app-key",
					ApplicationSecret: v1.SecretKeySelector{Key: "valid-secret-key"},
					ConsumerKey:       v1.SecretKeySelector{Key: "valid-consumer-key"},
					Service:           monitoringv1alpha1.OVHServiceVPS,
					RefreshInterval:   ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
}

var OpenStackSDTestCases = []scrapeCRDTestCase{
	{
		name: "Bareconfig",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Region: "default",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Region",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role: monitoringv1alpha1.OpenStackRoleHypervisor,
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Role Hypervisor",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Role Instance",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleInstance,
					Region: "default",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRole("default"),
					Region: "default",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Role Empty",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRole(""),
					Region: "default",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Region Empty",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "",
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Role Loadbalancer",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleLoadBalancer,
					Region: "default",
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Endpoint HTTP",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:             monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:           "default",
					IdentityEndpoint: ptr.To("http://example.com"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Endpoint HTTPS",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:             monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:           "default",
					IdentityEndpoint: ptr.To("https://example.com"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Endpoint",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:             monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:           "default",
					IdentityEndpoint: ptr.To("ftp://example.com"),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Endpoint Empty",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:             monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:           "default",
					IdentityEndpoint: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Username",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:     monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:   "default",
					Username: ptr.To("admin"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Username",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:     monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:   "default",
					Username: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid User ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
					UserID: ptr.To("ac3377633149401296f6c0d92d79dc16"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid User ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
					UserID: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Domain ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:     monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:   "default",
					DomainID: ptr.To("e0353a670a9e496da891347c589539e9"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Domain ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:     monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:   "default",
					DomainID: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Domain Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:       monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:     "default",
					DomainName: ptr.To("default"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Domain Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:       monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:     "default",
					DomainName: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Project Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:        monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:      "default",
					ProjectName: ptr.To("default"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Project Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:        monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:      "default",
					ProjectName: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Project ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:      monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:    "default",
					ProjectID: ptr.To("343d245e850143a096806dfaefa9afdc"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Project ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:      monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:    "default",
					ProjectID: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Application Credential Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:                      monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:                    "default",
					ApplicationCredentialName: ptr.To("monitoring"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Application Credential Name",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:                      monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:                    "default",
					ApplicationCredentialName: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Application Credential ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:                    monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:                  "default",
					ApplicationCredentialID: ptr.To("aa809205ed614a0e854bac92c0768bb9"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Application Credential ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:                    monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:                  "default",
					ApplicationCredentialID: ptr.To(""),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "All Tenants True",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:       monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:     "default",
					AllTenants: ptr.To(true),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "All Tenants False",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:       monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:     "default",
					AllTenants: ptr.To(false),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Refresh Interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:            monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:          "default",
					RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Refresh Interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:            monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:          "default",
					RefreshInterval: ptr.To(monitoringv1.Duration("30g")),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port 8080",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
					Port:   ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port -1",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
					Port:   ptr.To(int32(-1)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Port 65537",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
					Region: "default",
					Port:   ptr.To(int32(65537)),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Availability Public",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:         monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:       "default",
					Availability: ptr.To("public"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Availability Admin",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:         monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:       "default",
					Availability: ptr.To("admin"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Availability Internal",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:         monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:       "default",
					Availability: ptr.To("internal"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Availability",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:         monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:       "default",
					Availability: ptr.To("private"),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Availability Empty Value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			OpenStackSDConfigs: []monitoringv1alpha1.OpenStackSDConfig{
				{
					Role:         monitoringv1alpha1.OpenStackRoleHypervisor,
					Region:       "default",
					Availability: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
}

var ScalewaySDTestCases = []scrapeCRDTestCase{
	{
		name: "Valid Project ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Project ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Project ID",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Access Key",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Access Key",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Access Key",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Role - Instance",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Valid Role - Baremetal",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleBaremetal,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      "Invalid Role",
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Missing Role",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Api Url",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					ApiURL: ptr.To("https://api.scaleway.com"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Api Url",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					ApiURL: ptr.To("ftp://example.com"),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid NameFilter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					NameFilter: ptr.To("my-server"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid NameFilter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					NameFilter: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid TagsFilter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					TagsFilter: []string{"do"},
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Empty tag values in TagsFilter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					TagsFilter: []string{""}, // empty tag
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Repeating tags in TagsFilter",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					TagsFilter: []string{"do", "do"}, // repeating tags
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Zone value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleBaremetal,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					Zone: ptr.To("fr-par-1"),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Zone value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					Zone: ptr.To(""),
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Port number",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					Port: ptr.To(int32(8080)),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Port number exceeding the maximum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleBaremetal,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					Port: ptr.To(int32(65536)), // maximum Port number = 65535
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Invalid Port number below the minimum value",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					Port: ptr.To(int32(-1)), // minimum Port number = 0;
				},
			},
		},
		expectedError: true,
	},
	{
		name: "Valid Refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleInstance,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					RefreshInterval: ptr.To(monitoringv1.Duration("60s")),
				},
			},
		},
		expectedError: false,
	},
	{
		name: "Invalid Refresh interval",
		scrapeConfigSpec: monitoringv1alpha1.ScrapeConfigSpec{
			ScalewaySDConfigs: []monitoringv1alpha1.ScalewaySDConfig{
				{
					ProjectID: "1",
					Role:      monitoringv1alpha1.ScalewayRoleBaremetal,
					AccessKey: "AccessKey",
					SecretKey: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key.pem",
					},
					RefreshInterval: ptr.To(monitoringv1.Duration("60g")),
				},
			},
		},
		expectedError: true,
	},
}
