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

package prometheus

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestResolveStuckStatefulSet_Delete(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	logger := slog.Default()

	op := &Operator{
		kclient: client,
		logger:  logger,
	}

	repairPolicy := monitoringv1.DeleteNotReadyPodsRepairPolicy
	p := &monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				UpdateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
					RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
						RepairPolicy: &repairPolicy,
					},
				},
			},
		},
	}

	sset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		Status: appsv1.StatefulSetStatus{
			UpdateRevision: "revision-2",
		},
	}

	// Create stuck pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Labels: map[string]string{
				"app":                                 "test",
				appsv1.ControllerRevisionHashLabelKey: "revision-1",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionFalse},
			},
		},
	}
	_, err := client.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err)

	err = op.resolveStuckStatefulSet(ctx, logger, p, sset)
	require.NoError(t, err)

	// Verify pod is deleted
	_, err = client.CoreV1().Pods("default").Get(ctx, "test-pod-0", metav1.GetOptions{})
	require.Error(t, err)
	require.True(t, apierrors.IsNotFound(err))
}

func TestResolveStuckStatefulSet_IgnoreReady(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	logger := slog.Default()

	op := &Operator{
		kclient: client,
		logger:  logger,
	}

	repairPolicy := monitoringv1.DeleteNotReadyPodsRepairPolicy
	p := &monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				UpdateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
					RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
						RepairPolicy: &repairPolicy,
					},
				},
			},
		},
	}

	sset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		Status: appsv1.StatefulSetStatus{
			UpdateRevision: "revision-2",
		},
	}

	// Create ready pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Labels: map[string]string{
				"app":                                 "test",
				appsv1.ControllerRevisionHashLabelKey: "revision-1",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionTrue},
			},
		},
	}
	_, err := client.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err)

	err = op.resolveStuckStatefulSet(ctx, logger, p, sset)
	require.NoError(t, err)

	// Verify pod is NOT deleted
	_, err = client.CoreV1().Pods("default").Get(ctx, "test-pod-0", metav1.GetOptions{})
	require.NoError(t, err)
}

func TestResolveStuckStatefulSet_IgnoreCurrentRevision(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	logger := slog.Default()

	op := &Operator{
		kclient: client,
		logger:  logger,
	}

	repairPolicy := monitoringv1.DeleteNotReadyPodsRepairPolicy
	p := &monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				UpdateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
					RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
						RepairPolicy: &repairPolicy,
					},
				},
			},
		},
	}

	sset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		Status: appsv1.StatefulSetStatus{
			UpdateRevision: "revision-2",
		},
	}

	// Create pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-0",
			Namespace: "default",
			Labels: map[string]string{
				"app":                                 "test",
				appsv1.ControllerRevisionHashLabelKey: "revision-2",
			},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionFalse},
			},
		},
	}
	_, err := client.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err)

	err = op.resolveStuckStatefulSet(ctx, logger, p, sset)
	require.NoError(t, err)

	// Verify pod is NOT deleted
	_, err = client.CoreV1().Pods("default").Get(ctx, "test-pod-0", metav1.GetOptions{})
	require.NoError(t, err)
}
