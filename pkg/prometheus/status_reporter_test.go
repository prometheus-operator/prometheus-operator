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

package prometheus

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

// mockSsetLister implements cache.GenericLister for testing.
type mockSsetLister struct {
	objects map[string]runtime.Object
}

func (m *mockSsetLister) List(_ labels.Selector) (ret []runtime.Object, err error) {
	for _, obj := range m.objects {
		ret = append(ret, obj)
	}
	return ret, nil
}

func (m *mockSsetLister) Get(name string) (runtime.Object, error) {
	if obj, ok := m.objects[name]; ok {
		return obj, nil
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "statefulsets"}, name)
}

func (m *mockSsetLister) ByNamespace(_ string) cache.GenericNamespaceLister {
	return nil
}

// mockInformLister implements informers.InformLister for creating ForResource.
type mockInformLister struct {
	lister *mockSsetLister
}

func (m *mockInformLister) Informer() cache.SharedIndexInformer {
	return nil
}

func (m *mockInformLister) Lister() cache.GenericLister {
	return m.lister
}

// mockFactoriesForNamespaces implements informers.FactoriesForNamespaces for testing.
type mockFactoriesForNamespaces struct {
	namespaces sets.Set[string]
	lister     *mockSsetLister
}

func (m *mockFactoriesForNamespaces) ForResource(_ string, _ schema.GroupVersionResource) (informers.InformLister, error) {
	return &mockInformLister{lister: m.lister}, nil
}

func (m *mockFactoriesForNamespaces) Namespaces() sets.Set[string] {
	return m.namespaces
}

// newTestForResource creates a *informers.ForResource for testing with the given objects.
func newTestForResource(objects map[string]runtime.Object) (*informers.ForResource, error) {
	factory := &mockFactoriesForNamespaces{
		namespaces: sets.New[string]("default"),
		lister:     &mockSsetLister{objects: objects},
	}
	return informers.NewInformersForResource(factory, schema.GroupVersionResource{})
}

func TestStatusReporterProcess(t *testing.T) {
	// Test cases cover multi-shard Prometheus availability status aggregation.
	// The bug being tested: before the fix, the code checked availableCondition.Status
	// (which was never updated) instead of availableStatus, causing incorrect
	// status aggregation when one shard is unavailable and another is degraded.
	testCases := []struct {
		name           string
		shards         int32
		replicas       int32
		shardStates    []shardState // state for each shard
		expectedStatus monitoringv1.ConditionStatus
		expectedReason string
	}{
		{
			name:     "one shard with zero ready pods returns Available=False",
			shards:   2,
			replicas: 2,
			shardStates: []shardState{
				{totalPods: 2, readyPods: 2}, // shard 0: all ready
				{totalPods: 2, readyPods: 0}, // shard 1: no ready pods
			},
			expectedStatus: monitoringv1.ConditionFalse,
			expectedReason: "NoPodReady",
		},
		{
			name:     "one shard degraded and another shard unavailable returns Available=False not Degraded",
			shards:   2,
			replicas: 2,
			shardStates: []shardState{
				{totalPods: 2, readyPods: 0}, // shard 0: no ready pods (unavailable)
				{totalPods: 2, readyPods: 1}, // shard 1: partially ready (degraded)
			},
			expectedStatus: monitoringv1.ConditionFalse,
			expectedReason: "NoPodReady",
		},
		{
			name:     "all shards partially degraded returns Available=Degraded",
			shards:   2,
			replicas: 2,
			shardStates: []shardState{
				{totalPods: 2, readyPods: 1}, // shard 0: partially ready
				{totalPods: 2, readyPods: 1}, // shard 1: partially ready
			},
			expectedStatus: monitoringv1.ConditionDegraded,
			expectedReason: "SomePodsNotReady",
		},
		{
			name:     "all shards healthy returns Available=True",
			shards:   2,
			replicas: 2,
			shardStates: []shardState{
				{totalPods: 2, readyPods: 2}, // shard 0: all ready
				{totalPods: 2, readyPods: 2}, // shard 1: all ready
			},
			expectedStatus: monitoringv1.ConditionTrue,
			expectedReason: "",
		},
		{
			// This is the key test case that demonstrates the bug fix.
			// Before the fix, this would return Degraded because availableCondition.Status
			// (empty string) != ConditionFalse would be true.
			name:     "unavailable shard followed by degraded shard stays False (regression test)",
			shards:   3,
			replicas: 2,
			shardStates: []shardState{
				{totalPods: 2, readyPods: 2}, // shard 0: healthy
				{totalPods: 2, readyPods: 0}, // shard 1: unavailable (sets status to False)
				{totalPods: 2, readyPods: 1}, // shard 2: degraded (should NOT override False)
			},
			expectedStatus: monitoringv1.ConditionFalse,
			expectedReason: "NoPodReady",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Create the Prometheus resource with sharding.
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Shards:   ptr.To(tc.shards),
						Replicas: ptr.To(tc.replicas),
					},
				},
			}

			// Build statefulsets and pods for each shard.
			ssets := make(map[string]runtime.Object)
			var pods []runtime.Object

			for shard, state := range tc.shardStates {
				ssetName := KeyToStatefulSetKey(p, "default/test", shard)
				sset := createStatefulSet(ssetName, "default", int(tc.replicas), state.totalPods)
				ssets[ssetName] = sset

				shardPods := createPods(sset, state.totalPods, state.readyPods)
				pods = append(pods, shardPods...)
			}

			// Create fake kubernetes client with pods.
			fakeClient := fake.NewSimpleClientset(pods...)

			// Create the ForResource informer with the statefulsets.
			ssetInfs, err := newTestForResource(ssets)
			require.NoError(t, err)

			// Create the StatusReporter with our test fixtures.
			// Note: Rr (ResourceReconciler) is nil-safe for our tests because:
			// - DeletionInProgress only accesses fields if DeletionTimestamp is set
			// - Our test StatefulSets don't have DeletionTimestamp set
			sr := &StatusReporter{
				Kclient:         fakeClient,
				SsetInfs:        ssetInfs,
				Rr:              nil,
				Reconciliations: &operator.ReconciliationTracker{},
			}

			status, err := sr.Process(ctx, p, "default/test")
			require.NoError(t, err)
			require.NotNil(t, status)

			// Find the Available condition.
			var availableCondition *monitoringv1.Condition
			for i := range status.Conditions {
				if status.Conditions[i].Type == monitoringv1.Available {
					availableCondition = &status.Conditions[i]
					break
				}
			}

			require.NotNil(t, availableCondition, "Available condition not found")
			require.Equal(t, tc.expectedStatus, availableCondition.Status,
				"Expected Available.Status=%q, got %q", tc.expectedStatus, availableCondition.Status)
			require.Equal(t, tc.expectedReason, availableCondition.Reason,
				"Expected Available.Reason=%q, got %q", tc.expectedReason, availableCondition.Reason)
		})
	}
}

// shardState describes the state of a single shard for testing.
type shardState struct {
	totalPods int
	readyPods int
}

// createStatefulSet creates a StatefulSet for testing.
func createStatefulSet(name, namespace string, replicas, currentPods int) *appsv1.StatefulSet {
	// Parse the name to get just the statefulset name (without namespace prefix).
	parts := splitKey(name)
	ssetName := parts[1]
	ns := parts[0]
	if namespace != "" {
		ns = namespace
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ssetName,
			Namespace: ns,
			UID:       "test-uid",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To(int32(replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": ssetName,
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:       int32(currentPods),
			ReadyReplicas:  int32(currentPods),
			UpdateRevision: "rev-1",
		},
	}
}

// splitKey splits a "namespace/name" key into its components.
func splitKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if key[i] == '/' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{"", key}
}

// createPods creates test pods for a StatefulSet.
func createPods(sset *appsv1.StatefulSet, total, ready int) []runtime.Object {
	pods := make([]runtime.Object, 0, total)

	for i := range total {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sset.Name + "-" + strconv.Itoa(i),
				Namespace: sset.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name":   sset.Name,
					"controller-revision-hash": sset.Status.UpdateRevision,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "StatefulSet",
						Name: sset.Name,
						UID:  sset.UID,
					},
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				Conditions: []v1.PodCondition{
					{
						Type:   v1.PodReady,
						Status: v1.ConditionFalse,
					},
				},
			},
		}

		// Mark pods as ready up to the 'ready' count.
		if i < ready {
			pod.Status.Conditions[0].Status = v1.ConditionTrue
		}

		pods = append(pods, pod)
	}

	return pods
}
