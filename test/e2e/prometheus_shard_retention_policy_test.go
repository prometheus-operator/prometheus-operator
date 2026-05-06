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
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

// testPrometheusTargetDistributionOnResharding verifies that targets are
// correctly distributed across active shard(s) when their number scales up
// and down and the Retain policy is used. It also ensures that targets setting
// the __tmp_disable_sharding label are scraped by all active shards.
//
// After a scale-down, the "inactive" shards shouldn't scrape any target.
func testPrometheusTargetDistributionOnResharding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	const (
		prometheusName       = "shard-retention"
		prometheusGroupLabel = "test"
	)
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

	// Deploy an application with 10 replicas to ensure that targets will be
	// spread across the 2 shards.
	err = framework.DeployBasicAuthApp(ctx, ns, 10)
	require.NoError(t, err)

	// Create a service monitor for the app deployment.
	err = framework.DeployAppServiceMonitor(ctx, ns)
	require.NoError(t, err)

	// Create 1 service monitor for the Prometheus service.
	sm := framework.MakeBasicServiceMonitor(prometheusGroupLabel)
	sm.Spec.Endpoints[0].RelabelConfigs = []monitoringv1.RelabelConfig{
		{
			TargetLabel: "__tmp_disable_sharding",
			Action:      "Replace",
			Replacement: new("true"),
		},
	}
	sm, err = framework.MonClientV1.ServiceMonitors(ns).Create(ctx, sm, metav1.CreateOptions{})
	require.NoError(t, err)

	// Deploy a Prometheus resource with 1 shard and ensure that it discovers
	// 10 targets for the app service and 1 target for the test service.
	// We test only against the Retain strategy. There's no need to verify with
	// the Delete strategy because the second shard will not exist anymore in
	// case of scale down.
	prom := framework.MakeBasicPrometheus(ns, prometheusName, prometheusGroupLabel, 1)
	prom.Spec.Shards = new(int32(1))
	prom.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "group",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{testFramework.AppGroupLabel, prometheusGroupLabel},
			},
		},
	}

	prom.Spec.ShardRetentionPolicy = &monitoringv1.ShardRetentionPolicy{
		WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
	}

	shardServices := make([]*corev1.Service, 2)
	for i := range shardServices {
		svc := framework.MakePrometheusService(prometheusName, prometheusGroupLabel, corev1.ServiceTypeClusterIP)
		svc.Name += "-" + strconv.Itoa(i)
		svc.Spec.Selector["operator.prometheus.io/shard"] = strconv.Itoa(i)

		svc, err = framework.KubeClient.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
		require.NoError(t, err)
		shardServices[i] = svc
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
	require.NoError(t, err)

	// Expecting 10 app pods + 1 prometheus pod.
	err = framework.WaitForHealthyTargets(context.Background(), ns, shardServices[0].Name, 11)
	require.NoError(t, err)

	// Scale up the number of shards to 2 and ensure that
	// * Each shard discovers more than 2 targets (at least 1 app pod + 2 prometheus pods).
	// * The sum of targets is 10 app pods + 2*2 prometheus pods = 14.
	_, err = framework.ScalePrometheusAndWaitUntilReady(ctx, prometheusName, ns, 2)
	require.NoError(t, err)

	t.Run("2 active shards", func(t *testing.T) {
		var pollErr error
		err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
			var total int
			for _, svc := range shardServices {
				err := framework.WaitForHealthyTargetsWithCondition(
					context.Background(),
					ns,
					svc.Name,
					func(targets []*testFramework.Target) error {
						if len(targets) < 2 {
							return errors.New("expected more than 2 targets")
						}

						total += len(targets)
						return nil
					},
				)
				if err != nil {
					pollErr = fmt.Errorf("%s: %w", svc.Name, err)
					return false, nil
				}
			}

			if total != 14 {
				pollErr = fmt.Errorf("expected 14 targets, got %d", total)
				return false, nil
			}
			return true, nil
		})
		require.NoError(t, err, fmt.Sprintf("%s: %s", err, pollErr))
	})

	// Scale down the number of shards to 1 and ensure that all targets are
	// reaffected to the first shard.
	_, err = framework.ScalePrometheusAndWaitUntilReady(ctx, prometheusName, ns, 1)
	require.NoError(t, err)

	t.Run("1 active shard", func(t *testing.T) {
		for i, svc := range shardServices {
			t.Run(svc.Name, func(t *testing.T) {
				t.Parallel()

				var targets int
				if i == 0 {
					// 10 app pods + 2 prometheus pods.
					targets = 12
				}
				err := framework.WaitForHealthyTargets(context.Background(), ns, svc.Name, targets)
				require.NoError(t, err)
			})
		}
	})
}

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
			whenScaledDown: new(monitoringv1.DeleteWhenScaledRetentionType),
		},
		{
			name:           "retain policy",
			whenScaledDown: new(monitoringv1.RetainWhenScaledRetentionType),
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name := "shard-retention-" + strconv.Itoa(i)
			p := framework.MakeBasicPrometheus(ns, name, name, 1)
			p.Spec.ShardRetentionPolicy = &monitoringv1.ShardRetentionPolicy{
				WhenScaled: tc.whenScaledDown,
			}
			p.Spec.Shards = new(int32(2))
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
				// Scenario with Delete when scaling down stops here.
				return
			}

			// Ensure that the deadline annotation is defined.
			for _, sts := range sts.Items {
				deadlineAnnotation, found := sts.Annotations["operator.prometheus.io/deletion-deadline"]
				switch shard := sts.Labels["operator.prometheus.io/shard"]; shard {
				case "0":
					require.False(t, found)
				case "1":
					require.True(t, found)
					require.NotEmpty(t, deadlineAnnotation)
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

			t.Log("patching Prometheus and scaling down the number of shards to 1 again")
			// Set a very low retention period to ensure that the operator
			// deletes the inactive shard.
			_, err = framework.PatchPrometheus(
				context.Background(),
				p.Name,
				ns,
				monitoringv1.PrometheusSpec{
					ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
						Retain: &monitoringv1.RetainConfig{
							RetentionPeriod: "1s",
						},
					},
				},
			)
			require.NoError(t, err)

			p, err = framework.ScalePrometheusAndWaitUntilReady(ctx, name, ns, 1)
			require.NoError(t, err)
			require.Equal(t, int32(1), p.Status.Shards)

			sts, err = framework.KubeClient.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{LabelSelector: p.Status.Selector})
			require.NoError(t, err)
			require.Len(t, sts.Items, 1)
		})
	}
}
