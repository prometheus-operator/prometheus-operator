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
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testRepairPolicy(t *testing.T) {
	for _, tc := range []struct {
		policy string
	}{
		{
			policy: string(operator.EvictRepairPolicy),
		},
		{
			policy: string(operator.DeleteRepairPolicy),
		},
	} {
		t.Run(tc.policy, func(t *testing.T) {
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)

			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			if tc.policy == string(operator.EvictRepairPolicy) {
				// Grant evict permission to the prometheus-operator service account.
				// TODO: remove when the permission gets included into the
				// default set of permissions for the service account.
				evictRole := &rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name: "evict",
					},
					Rules: []rbacv1.PolicyRule{{
						Verbs:     []string{"create"},
						Resources: []string{"pods/eviction"},
						APIGroups: []string{""},
					}},
				}
				evictRole, err := framework.KubeClient.RbacV1().Roles(ns).Create(context.Background(), evictRole, metav1.CreateOptions{})
				require.NoError(t, err)

				roleBinding := &rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: evictRole.Name,
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     "Role",
						Name:     evictRole.Name,
					},
					Subjects: []rbacv1.Subject{{
						Kind:      "ServiceAccount",
						Name:      "prometheus-operator",
						Namespace: ns,
					}},
				}
				_, err = framework.KubeClient.RbacV1().RoleBindings(ns).Create(context.Background(), roleBinding, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
				context.Background(),
				operatorFramework.PrometheusOperatorOpts{
					Namespace:         ns,
					AllowedNamespaces: []string{ns},
					AdditionalArgs: []string{
						"--repair-policy-for-statefulsets=" + tc.policy,
					},
				},
			)
			require.NoError(t, err)

			t.Run("workload", func(t *testing.T) {
				t.Run("Prometheus", repairPrometheus(ns))
				t.Run("Alertmanager", repairAlertmanager(ns))
				t.Run("ThanosRuler", repairThanosRuler(ns))
			})
		})
	}
}

const badImage = "quay.io/prometheus-operator/foo:bar"

func repairPrometheus(ns string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		prom := framework.MakeBasicPrometheus(ns, "repair", "test", 2)
		prom.Spec.PodManagementPolicy = ptr.To(monitoringv1.OrderedReadyPodManagement)
		prom, err := framework.CreatePrometheusAndWaitUntilReady(
			context.Background(),
			ns,
			prom,
		)
		require.NoError(t, err)

		prom, err = framework.PatchPrometheus(
			context.Background(),
			prom.Name,
			prom.Namespace,
			monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: new(badImage),
				},
			},
		)
		require.NoError(t, err)

		// The rollout should start from the highest pod ordinal.
		err = framework.WaitForContainerInErrPullImage(context.Background(), prom.Namespace, "prometheus-"+prom.Name+"-1", badImage)
		require.NoError(t, err)

		// Fix the bad image location and ensure that the resource goes back to ready.
		_, err = framework.PatchPrometheusAndWaitUntilReady(
			context.Background(),
			prom.Name,
			prom.Namespace,
			monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: new(operator.DefaultPrometheusImage),
				},
			},
		)
		require.NoError(t, err)
	}
}

func repairAlertmanager(ns string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		am := framework.MakeBasicAlertmanager(ns, "repair", 2)
		am.Spec.PodManagementPolicy = ptr.To(monitoringv1.OrderedReadyPodManagement)
		am, err := framework.CreateAlertmanagerAndWaitUntilReady(
			context.Background(),
			am,
		)
		require.NoError(t, err)

		am, err = framework.PatchAlertmanager(
			context.Background(),
			am.Name,
			am.Namespace,
			monitoringv1.AlertmanagerSpec{
				Image: new(badImage),
			},
		)
		require.NoError(t, err)

		// The rollout should start from the highest pod ordinal.
		err = framework.WaitForContainerInErrPullImage(context.Background(), am.Namespace, "alertmanager-"+am.Name+"-1", badImage)
		require.NoError(t, err)

		// Fix the bad image location and ensure that the resource goes back to ready.
		_, err = framework.PatchAlertmanagerAndWaitUntilReady(
			context.Background(),
			am.Name,
			am.Namespace,
			monitoringv1.AlertmanagerSpec{
				Image: ptr.To(operator.DefaultAlertmanagerImage),
			},
		)
		require.NoError(t, err)
	}
}

func repairThanosRuler(ns string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		tr := framework.MakeBasicThanosRuler("repair", 2, "http://example.com")
		tr.Spec.PodManagementPolicy = ptr.To(monitoringv1.OrderedReadyPodManagement)
		tr, err := framework.CreateThanosRulerAndWaitUntilReady(
			context.Background(),
			ns,
			tr,
		)
		require.NoError(t, err)

		tr, err = framework.PatchThanosRuler(
			context.Background(),
			tr.Name,
			tr.Namespace,
			monitoringv1.ThanosRulerSpec{
				Image: badImage,
			},
		)
		require.NoError(t, err)

		// The rollout should start from the highest pod ordinal.
		err = framework.WaitForContainerInErrPullImage(context.Background(), tr.Namespace, "thanos-ruler-"+tr.Name+"-1", badImage)
		require.NoError(t, err)

		// Fix the bad image location and ensure that the resource goes back to ready.
		_, err = framework.PatchThanosRulerAndWaitUntilReady(
			context.Background(),
			tr.Name,
			tr.Namespace,
			monitoringv1.ThanosRulerSpec{
				Image: operator.DefaultThanosImage,
			},
		)
		require.NoError(t, err)
	}
}
