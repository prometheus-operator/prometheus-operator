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
	"k8s.io/apimachinery/pkg/labels"
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
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusTopologyShardingFeature},
		},
	)
	require.NoError(t, err)

	name := "topology-sharding"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
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
						"operator.prometheus.io/name":  name,
						"topology.kubernetes.io/zone":  zone,
					}).String(),
				},
			)
			require.NoError(t, err)
		})
	}
}
