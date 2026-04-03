// Copyright The prometheus-operator Authors
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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfigurationsappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

// testPrometheusRetentionPolicies tests the shard retention policies for Prometheus.
// ShardRetentionPolicy requires the ShardRetention feature gate to be enabled,
// therefore, it runs in the feature-gated test suite.
func testPrometheusRetentionPolicies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusShardRetentionPolicyFeature},
		},
	)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		whenScaledDown *monitoringv1.WhenScaledRetentionType
	}{
		{
			name:           "delete policy",
			whenScaledDown: ptr.To(monitoringv1.DeleteWhenScaledRetentionType),
		},
		{
			name:           "retain policy",
			whenScaledDown: ptr.To(monitoringv1.RetainWhenScaledRetentionType),
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name := "shard-retention-" + strconv.Itoa(i)
			p := framework.MakeBasicPrometheus(ns, name, name, 1)
			p.Spec.ShardRetentionPolicy = &monitoringv1.ShardRetentionPolicy{
				WhenScaled: tc.whenScaledDown,
			}
			p.Spec.Shards = ptr.To(int32(2))
			_, err := framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
			require.NoError(t, err)

			t.Log("scaling down the number of shards to 1")
			p, err = framework.ScalePrometheusAndWaitUntilReady(ctx, name, ns, 1)
			require.NoError(t, err)
			require.Equal(t, int32(1), p.Status.Shards)

			sts, err := framework.KubeClient.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{LabelSelector: p.Status.Selector})
			require.NoError(t, err)

			expectedRemaining := 2
			if *tc.whenScaledDown == monitoringv1.DeleteWhenScaledRetentionType {
				expectedRemaining = 1
			}
			require.Len(t, sts.Items, expectedRemaining)

			if expectedRemaining == 1 {
				return
			}

			var shard1 string
			for _, sts := range sts.Items {
				deadlineAnnotation, found := sts.Annotations["operator.prometheus.io/deletion-deadline"]
				switch shard := sts.Labels["operator.prometheus.io/shard"]; shard {
				case "0":
					require.False(t, found)
				case "1":
					require.True(t, found)
					require.NotEmpty(t, deadlineAnnotation)
					shard1 = sts.Name
				default:
					t.Fatalf("unexpected shard label: %s", shard)
				}
			}

			t.Log("scaling up the number of shards to 2")
			p, err = framework.ScalePrometheusAndWaitUntilReady(ctx, name, ns, 2)
			require.NoError(t, err)
			require.Equal(t, int32(2), p.Status.Shards)

			sts, err = framework.KubeClient.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{LabelSelector: p.Status.Selector})
			require.NoError(t, err)
			require.Len(t, sts.Items, 2)

			for _, sts := range sts.Items {
				_, found := sts.Annotations["operator.prometheus.io/deletion-deadline"]
				switch shard := sts.Labels["operator.prometheus.io/shard"]; shard {
				case "0":
				case "1":
					require.False(t, found)
				default:
					t.Fatalf("unexpected shard label: %s", shard)
				}
			}

			t.Log("scaling down the number of shards to 1 again")
			p, err = framework.ScalePrometheusAndWaitUntilReady(ctx, name, ns, 1)
			require.NoError(t, err)
			require.Equal(t, int32(1), p.Status.Shards)

			// Update the deadline annotation to trigger a deletion of the statefulset.
			_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Apply(
				ctx,
				applyconfigurationsappsv1.StatefulSet(shard1, ns).WithAnnotations(map[string]string{
					"operator.prometheus.io/deletion-deadline": time.Now().UTC().Format(time.RFC3339),
				}),
				metav1.ApplyOptions{FieldManager: "e2e-test", Force: true},
			)
			require.NoError(t, err)

			err = framework.WaitForPodsReady(ctx, ns, 2*time.Minute, 1, metav1.ListOptions{LabelSelector: p.Status.Selector})
			require.NoError(t, err)
		})
	}
}
