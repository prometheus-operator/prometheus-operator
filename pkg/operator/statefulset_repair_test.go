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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestResolveStuckStatefulSet(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	namespace := "default"
	stsName := "test-sts"

	createPod := func(name, revision string, ready bool) v1.Pod {
		readyStatus := v1.ConditionFalse
		if ready {
			readyStatus = v1.ConditionTrue
		}
		return v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					"app":                                 stsName,
					appsv1.ControllerRevisionHashLabelKey: revision,
				},
			},
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{
						Type:   v1.PodReady,
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
			CurrentRevision:    "rev-1",
			UpdateRevision:     "rev-2",
		},
	}

	t.Run("none policy does nothing", func(t *testing.T) {
		kclient := fake.NewSimpleClientset(
			&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-sts-0", Namespace: namespace, Labels: map[string]string{"app": stsName}}},
		)
		err := ResolveStuckStatefulSet(ctx, logger, kclient, sset, NoneRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("skips if generation not reconciled", func(t *testing.T) {
		unreconciledSts := sset.DeepCopy()
		unreconciledSts.Status.ObservedGeneration = 0
		kclient := fake.NewSimpleClientset()
		err := ResolveStuckStatefulSet(ctx, logger, kclient, unreconciledSts, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 0, len(actions))
	})

	t.Run("repairs only one pod in reverse order", func(t *testing.T) {
		pod0 := createPod("test-sts-0", "rev-0", false)
		pod1 := createPod("test-sts-1", "rev-0", false)
		pod2 := createPod("test-sts-2", "rev-2", true)

		kclient := fake.NewSimpleClientset(&pod0, &pod1, &pod2)

		err := ResolveStuckStatefulSet(ctx, logger, kclient, sset, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 2, len(actions))

		deleteAction := actions[1].(clienttesting.DeleteAction)
		assert.Equal(t, "delete", deleteAction.GetVerb())
		assert.Equal(t, "test-sts-1", deleteAction.GetName())
	})

	t.Run("skips current or update revision", func(t *testing.T) {
		pod0 := createPod("test-sts-0", "rev-1", false)
		pod1 := createPod("test-sts-1", "rev-2", false)

		kclient := fake.NewSimpleClientset(&pod0, &pod1)
		err := ResolveStuckStatefulSet(ctx, logger, kclient, sset, DeleteRepairPolicy)
		require.NoError(t, err)

		actions := kclient.Actions()
		assert.Equal(t, 1, len(actions))
	})
}
