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
	"github.com/stretchr/testify/require"
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

	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

}

func testCreatePrometheusAgentDaemonSet(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ctx := context.Background()

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	name := "test"
	prometheusAgentDSCRD := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentDSCRD)
	require.NoError(t, err)
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

	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD)
	require.NoError(t, err)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)
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

	require.NoError(t, err, "%v: %v", err, loopError)
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
	require.NoError(t, err)

	require.Equal(t, int32(1), pAgent.Status.Shards)

	pAgent, err = framework.ScalePrometheusAgentAndWaitUntilReady(ctx, name, ns, 2)
	require.NoError(t, err)

	require.Equal(t, int32(2), pAgent.Status.Shards)
}
