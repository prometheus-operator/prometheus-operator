// Copyright 2025 The prometheus-operator Authors
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestStatefulSetRepair(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	namespace := "default"
	stsName := "test-sts"

	createPod := func(name, revision string, ready bool) corev1.Pod {
		readyStatus := corev1.ConditionFalse
		if ready {
			readyStatus = corev1.ConditionTrue
		}
		return corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					"app":                                 stsName,
					appsv1.ControllerRevisionHashLabelKey: revision,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "StatefulSet",
						Name: stsName,
					},
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: readyStatus,
					},
				},
			},
		}
	}

	sset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       stsName,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": stsName},
			},
		},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 1,
			CurrentRevision:    "rev-0",
			UpdateRevision:     "rev-2",
		},
	}

	t.Run("none policy does nothing", func(t *testing.T) {
		kclient := fake.NewSimpleClientset()
		fixer, err := NewStatefulSetReporter(ctx, kclient, sset)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, NoneRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("returns no error if statefulset is nil", func(t *testing.T) {
		kclient := fake.NewSimpleClientset()
		fixer, err := NewStatefulSetReporter(ctx, kclient, nil)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("skips if generation not reconciled", func(t *testing.T) {
		unreconciledSts := sset.DeepCopy()
		unreconciledSts.Status.ObservedGeneration = 0
		kclient := fake.NewSimpleClientset()
		fixer, err := NewStatefulSetReporter(ctx, kclient, unreconciledSts)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("repairs only one pod at a time in reverse order", func(t *testing.T) {
		pod0 := createPod("test-sts-0", "rev-0", false) // not ready with current revision.
		pod1 := createPod("test-sts-1", "rev-1", false) // not ready with neither current nor updated revision.
		pod2 := createPod("test-sts-2", "rev-1", true)  // ready with neither current nor updated revision.

		kclient := fake.NewSimpleClientset(&pod0, &pod1, &pod2)

		// Repair test-sts-2 pod.
		fixer, err := NewStatefulSetReporter(ctx, kclient, sset)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 1, len(actions))
		deleteAction, ok := actions[0].(clienttesting.DeleteAction)
		require.True(t, ok)
		assert.Equal(t, "delete", deleteAction.GetVerb())
		assert.Equal(t, "test-sts-2", deleteAction.GetName())

		// The next call shouldn't delete any pod because there's at most 1
		// repair operation allowed.
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		kclient.ClearActions()
		actions = kclient.Actions()
		assert.Equal(t, 0, len(actions))

		// Simulate new ready pod for test-sts-2.
		pod2 = createPod("test-sts-2", "rev-2", true)
		kclient.Tracker().Add(&pod2)

		// Repair test-sts-1 pod.
		fixer, err = NewStatefulSetReporter(ctx, kclient, sset)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions = kclient.Actions()
		assert.Equal(t, 1, len(actions))
		deleteAction, ok = actions[0].(clienttesting.DeleteAction)
		require.True(t, ok)
		assert.Equal(t, "delete", deleteAction.GetVerb())
		assert.Equal(t, "test-sts-1", deleteAction.GetName())

		// Simulate new ready pod for test-sts-1.
		pod1 = createPod("test-sts-1", "rev-2", true)
		kclient.Tracker().Add(&pod1)

		// No more pods to repair.
		fixer, err = NewStatefulSetReporter(ctx, kclient, sset)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions = kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("skips current or update revision", func(t *testing.T) {
		pod0 := createPod("test-sts-0", "rev-0", false) // not ready at current revision.
		pod1 := createPod("test-sts-1", "rev-2", false) // not ready at updated revision.

		kclient := fake.NewSimpleClientset(&pod0, &pod1)

		fixer, err := NewStatefulSetReporter(ctx, kclient, sset)
		require.NoError(t, err)

		kclient.ClearActions()
		err = fixer.Repair(ctx, logger, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})
}
