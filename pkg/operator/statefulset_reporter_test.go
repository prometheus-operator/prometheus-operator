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

package operator

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func makePod(name, namespace, revisionHash string, ready bool) *Pod {
	phase := v1.PodRunning
	condStatus := v1.ConditionTrue
	if !ready {
		phase = v1.PodPending
		condStatus = v1.ConditionFalse
	}
	return &Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"controller-revision-hash": revisionHash,
			},
		},
		Status: v1.PodStatus{
			Phase: phase,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: condStatus,
				},
			},
		},
	}
}

func makeStatefulSetReporter(updateRevision string, pods ...*Pod) *StatefulSetReporter {
	return &StatefulSetReporter{
		sset: &appsv1.StatefulSet{
			Status: appsv1.StatefulSetStatus{
				UpdateRevision: updateRevision,
			},
		},
		Pods: pods,
	}
}

func TestStuckPods(t *testing.T) {
	tests := []struct {
		name          string
		reporter      *StatefulSetReporter
		expectedStuck []string
	}{
		{
			name:          "nil statefulset returns nil",
			reporter:      &StatefulSetReporter{sset: nil, Pods: []*Pod{}},
			expectedStuck: nil,
		},
		{
			name: "all pods on current revision",
			reporter: makeStatefulSetReporter("rev-2",
				makePod("pod-0", "ns", "rev-2", true),
				makePod("pod-1", "ns", "rev-2", true),
			),
			expectedStuck: nil,
		},
		{
			name: "pod on old revision but ready is not stuck",
			reporter: makeStatefulSetReporter("rev-2",
				makePod("pod-0", "ns", "rev-2", true),
				makePod("pod-1", "ns", "rev-1", true),
			),
			expectedStuck: nil,
		},
		{
			name: "pod on current revision but not ready is not stuck",
			reporter: makeStatefulSetReporter("rev-2",
				makePod("pod-0", "ns", "rev-2", true),
				makePod("pod-1", "ns", "rev-2", false),
			),
			expectedStuck: nil,
		},
		{
			name: "pod on old revision and not ready is stuck",
			reporter: makeStatefulSetReporter("rev-2",
				makePod("pod-0", "ns", "rev-2", true),
				makePod("pod-1", "ns", "rev-1", false),
			),
			expectedStuck: []string{"pod-1"},
		},
		{
			name: "multiple stuck pods",
			reporter: makeStatefulSetReporter("rev-3",
				makePod("pod-0", "ns", "rev-2", false),
				makePod("pod-1", "ns", "rev-1", false),
				makePod("pod-2", "ns", "rev-3", true),
			),
			expectedStuck: []string{"pod-0", "pod-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stuck := tc.reporter.StuckPods()
			var names []string
			for _, p := range stuck {
				names = append(names, p.Name)
			}
			if tc.expectedStuck == nil {
				require.Empty(t, names)
			} else {
				require.Equal(t, tc.expectedStuck, names)
			}
		})
	}
}

func TestRepairStuckPods(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	stuckPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "ns",
			Labels:    map[string]string{"controller-revision-hash": "rev-1"},
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionFalse},
			},
		},
	}

	reporter := makeStatefulSetReporter("rev-2",
		makePod("pod-0", "ns", "rev-2", true),
		makePod("pod-1", "ns", "rev-1", false),
	)

	tests := []struct {
		name           string
		policy         monitoringv1.RepairPolicyType
		expectEviction bool
		expectDeletion bool
		expectNoAction bool
	}{
		{
			name:           "None policy takes no action",
			policy:         monitoringv1.RepairPolicyNone,
			expectNoAction: true,
		},
		{
			name:           "empty policy takes no action",
			policy:         "",
			expectNoAction: true,
		},
		{
			name:           "EvictNotReadyPods evicts stuck pods",
			policy:         monitoringv1.RepairPolicyEvictNotReadyPods,
			expectEviction: true,
		},
		{
			name:           "DeleteNotReadyPods deletes stuck pods",
			policy:         monitoringv1.RepairPolicyDeleteNotReadyPods,
			expectDeletion: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kclient := fake.NewClientset(&stuckPod)

			var evictions, deletions int
			kclient.PrependReactor("create", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
				if action.GetSubresource() == "eviction" {
					evictions++
					return true, nil, nil
				}
				return false, nil, nil
			})
			kclient.PrependReactor("delete", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
				deletions++
				return true, nil, nil
			})

			err := RepairStuckPods(ctx, logger, kclient, tc.policy, reporter)
			require.NoError(t, err)

			if tc.expectNoAction {
				require.Equal(t, 0, evictions, "expected no evictions")
				require.Equal(t, 0, deletions, "expected no deletions")
			}
			if tc.expectEviction {
				require.Equal(t, 1, evictions, "expected 1 eviction")
				require.Equal(t, 0, deletions, "expected no deletions")
			}
			if tc.expectDeletion {
				require.Equal(t, 0, evictions, "expected no evictions")
				require.Equal(t, 1, deletions, "expected 1 deletion")
			}
		})
	}
}

func TestRepairStuckPodsNoStuckPods(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()
	kclient := fake.NewClientset()

	reporter := makeStatefulSetReporter("rev-2",
		makePod("pod-0", "ns", "rev-2", true),
		makePod("pod-1", "ns", "rev-2", true),
	)

	var actions int
	kclient.PrependReactor("*", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		actions++
		return false, nil, nil
	})

	err := RepairStuckPods(ctx, logger, kclient, monitoringv1.RepairPolicyEvictNotReadyPods, reporter)
	require.NoError(t, err)
	require.Equal(t, 0, actions, "expected no API calls when no pods are stuck")
}
