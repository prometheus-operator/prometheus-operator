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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testPrometheusTopologySharding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	// Prometheus needs read permissions on Nodes to attach node metadata to
	// the discovered pods.
	framework.SetupPrometheusRBACGlobal(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusTopologyShardingFeature},
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

	const (
		prometheusName = "topology-sharding"
	)
	p := framework.MakeBasicPrometheus(ns, prometheusName, testFramework.AppGroupLabel, 1)
	p.Spec.ShardingStrategy = &monitoringv1.ShardingStrategy{
		Mode: ptr.To(monitoringv1.TopologyShardingStrategyMode),
		Topology: &monitoringv1.TopologyShardingStrategy{
			// Zone values are defined test/e2e/kind-conf.yaml
			// The expected mapping is
			// Shard 0 => zone-a
			// Shard 1 => zone-b
			Values: []string{"zone-a", "zone-b"},
		},
	}
	p.Spec.Shards = ptr.To(int32(2))

	shardServices := make([]*corev1.Service, 2)
	for i := range shardServices {
		svc := framework.MakePrometheusService(prometheusName, prometheusName, corev1.ServiceTypeClusterIP)
		svc.Name += "-" + strconv.Itoa(i)
		svc.Spec.Selector["operator.prometheus.io/shard"] = strconv.Itoa(i)
		svc, err = framework.KubeClient.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
		require.NoError(t, err)
		shardServices[i] = svc
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	// Ensure that each shard's pod is scheduled to the expected availability zone.
	for i, zone := range []string{"zone-a", "zone-b"} {
		t.Run(zone, func(t *testing.T) {
			err = framework.WaitForPodsReady(
				ctx,
				ns,
				10*time.Second,
				1,
				metav1.ListOptions{
					LabelSelector: labels.Set(map[string]string{
						"operator.prometheus.io/shard": strconv.Itoa(i),
						"operator.prometheus.io/name":  prometheusName,
						"topology.kubernetes.io/zone":  zone,
					}).String(),
				},
			)
			require.NoError(t, err)
		})
	}

	// Ensure that each shard scrapes a non-zero number of targets and that
	// the total across all shards equals the number of app replicas (10).
	t.Run("target distribution", func(t *testing.T) {
		var pollErr error
		err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
			total := 0
			for _, svc := range shardServices {
				err := framework.WaitForHealthyTargetsWithCondition(
					context.Background(),
					ns,
					svc.Name,
					func(targets []*testFramework.Target) error {
						if len(targets) == 0 {
							return errors.New("expected non-zero targets")
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
			if total != 10 {
				pollErr = fmt.Errorf("expected 10 total targets, got %d", total)
				return false, nil
			}
			return true, nil
		})
		require.NoError(t, err, fmt.Sprintf("%s: %s", err, pollErr))
	})
}
