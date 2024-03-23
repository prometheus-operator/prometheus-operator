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
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
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

func testAgentCheckStorageClass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)

	prometheusAgentCRD, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, prometheusAgentCRD)
	if err != nil {
		t.Fatal(err)
	}

	// Invalid storageclass e2e test

	_, err = framework.PatchPrometheusAgent(
		context.Background(),
		prometheusAgentCRD.Name,
		ns,
		monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
						Spec: v1.PersistentVolumeClaimSpec{
							StorageClassName: ptr.To("unknown-storage-class"),
							Resources: v1.VolumeResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceStorage: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var loopError error
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1alpha1.PrometheusAgents(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err == nil {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		t.Fatalf("%v: %v", err, loopError)
	}
}

func testPrometheusAgentStatusScale(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	pAgent := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	pAgent.Spec.CommonPrometheusFields.Shards = proto.Int32(1)

	pAgent, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, pAgent)
	if err != nil {
		t.Fatal(err)
	}

	if pAgent.Status.Shards != 1 {
		t.Fatalf("expected scale of 1 shard, got %d", pAgent.Status.Shards)
	}

	pAgent, err = framework.ScalePrometheusAgentAndWaitUntilReady(ctx, name, ns, 2)
	if err != nil {
		t.Fatal(err)
	}

	if pAgent.Status.Shards != 2 {
		t.Fatalf("expected scale of 2 shards, got %d", pAgent.Status.Shards)
	}
}

func testPrometheusAgentSecretUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	matchLabels := map[string]string{
		"tc": ns,
	}

	err := framework.AddLabelsToNamespace(context.Background(), ns, matchLabels)
	if err != nil {
		t.Fatal(err)
	}

	simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateDeployment(context.Background(), ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name: "web",
					Port: 8080,
				},
			},
			Selector: map[string]string{
				"group": name,
			},
		},
	}
	if finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(err)
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"bearertoken": []byte(""),
		},
	}
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     "web",
			Interval: "1s",
			Scheme:   "http",
			Path:     "bearer-metrics",
			BearerTokenSecret: &v1.SecretKeySelector{ //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
				LocalObjectReference: v1.LocalObjectReference{
					Name: name,
				},
				Key: "bearertoken",
			},
		},
	}

	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	pAgent := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	pAgent.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: matchLabels,
	}
	pAgent.Spec.ScrapeInterval = "1s"
	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, pAgent)
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-agent-operated", 1); err == nil {
		t.Fatal("All targets should be down")
	}

	secret.Data["bearertoken"] = []byte("abc")
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(context.Background(), secret, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Minute * 2)

	if err := framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-agent-operated", 1); err != nil {
		t.Fatal(err)
	}

}
