// Copyright 2016 The prometheus-operator Authors
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

package k8s

import (
	"context"
	"maps"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPropagateKubectlTemplateAnnotations(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		existing map[string]string
		new      map[string]string
		expected map[string]string
	}{
		{
			name:     "no annotations",
			expected: nil,
		},
		{
			name: "add owned annotation",
			new: map[string]string{
				"test-key": "test-value",
			},
			expected: map[string]string{
				"test-key": "test-value",
			},
		},
		{
			name: "change owned annotation",
			existing: map[string]string{
				"test-key": "test-value",
			},
			new: map[string]string{
				"test-key": "modified-test-value",
			},
			expected: map[string]string{
				"test-key": "modified-test-value",
			},
		},
		{
			name: "remove owned annotation",
			existing: map[string]string{
				"test-key": "test-value",
			},
			new:      map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "add kubectl annotation",
			existing: map[string]string{
				"test-key": "test-value",
			},
			new: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "now",
			},
			expected: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "now",
			},
		},
		{
			name: "modify kubectl annotation",
			existing: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "yesterday",
			},
			new: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "now",
			},
			expected: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "yesterday",
			},
		},
		{
			name: "remove kubectl annotation",
			existing: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "now",
			},
			new: map[string]string{},
			expected: map[string]string{
				"kubectl.kubernetes.io/restartedAt": "now",
			},
		},
	}

	namespace := "ns-1"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sset := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prometheus",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: tc.existing,
						},
					},
				},
			}

			ssetClient := fake.NewSimpleClientset(sset).AppsV1().StatefulSets(namespace)

			modifiedSset := sset.DeepCopy()
			modifiedSset.Spec.Template.Annotations = tc.new

			err := updateStatefulSet(ctx, ssetClient, modifiedSset)
			require.NoError(t, err)

			updatedSset, err := ssetClient.Get(ctx, "prometheus", metav1.GetOptions{})
			require.NoError(t, err)

			if !reflect.DeepEqual(tc.expected, updatedSset.Spec.Template.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expected, updatedSset.Spec.Template.Annotations)
			}
		})
	}
}

func TestMergeMetadata_UpdateStatefulSet(t *testing.T) {
	testCases := []struct {
		name                string
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
		modifiedLabels      map[string]string
		modifiedAnnotations map[string]string
	}{
		{
			name: "no change",
			expectedLabels: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
			},
			expectedAnnotations: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
			},
		},
		{
			name: "added label and annotation",
			expectedLabels: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
				"label":                  "value",
			},
			modifiedLabels: map[string]string{
				"label": "value",
			},
			expectedAnnotations: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
				"annotation":             "value",
			},
			modifiedAnnotations: map[string]string{
				"annotation": "value",
			},
		},
		{
			name: "overridden label amd annotation",
			expectedLabels: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
			},
			modifiedLabels: map[string]string{
				"app.kubernetes.io/name": "overridden-value",
			},
			expectedAnnotations: map[string]string{
				"app.kubernetes.io/name": "kube-state-metrics",
			},
			modifiedAnnotations: map[string]string{
				"app.kubernetes.io/name": "overridden-value",
			},
		},
	}

	namespace := "ns-1"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sset := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "prometheus",
					Namespace:   namespace,
					Labels:      map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
					Annotations: map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
				},
			}

			ssetClient := fake.NewSimpleClientset(sset).AppsV1().StatefulSets(namespace)

			modifiedSset := sset.DeepCopy()
			maps.Copy(modifiedSset.Labels, tc.modifiedLabels)
			maps.Copy(modifiedSset.Annotations, tc.modifiedAnnotations)
			_, err := ssetClient.Update(context.Background(), modifiedSset, metav1.UpdateOptions{})
			require.NoError(t, err)

			err = updateStatefulSet(context.Background(), ssetClient, sset)
			require.NoError(t, err)

			updatedSset, err := ssetClient.Get(context.Background(), "prometheus", metav1.GetOptions{})
			require.NoError(t, err)

			if !reflect.DeepEqual(tc.expectedAnnotations, updatedSset.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSset.Annotations)
			}
			if !reflect.DeepEqual(tc.expectedLabels, updatedSset.Labels) {
				t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSset.Labels)
			}
		})
	}
}

func TestCreateStatefulSetOrPatchLabels(t *testing.T) {
	testCases := []struct {
		name                string
		existingStatefulSet *appsv1.StatefulSet
		newStatefulSet      *appsv1.StatefulSet
		expectedLabels      map[string]string
	}{
		{
			name: "create new statefulset successfully",
			newStatefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prometheus",
					Namespace: "default",
					Labels: map[string]string{
						"app": "prometheus",
						"env": "prod",
					},
				},
			},
			expectedLabels: map[string]string{
				"app": "prometheus",
				"env": "prod",
			},
		},
		{
			name: "statefulset already exists - patch labels",
			existingStatefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prometheus",
					Namespace: "default",
					Labels: map[string]string{
						"app": "prometheus",
						"env": "dev",
					},
				},
			},
			newStatefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prometheus",
					Namespace: "default",
					Labels: map[string]string{
						"app":     "prometheus",
						"env":     "prod",
						"version": "v2.0",
					},
				},
			},
			expectedLabels: map[string]string{
				"app":     "prometheus",
				"env":     "prod",
				"version": "v2.0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			var clientSet *fake.Clientset
			if tc.existingStatefulSet != nil {
				clientSet = fake.NewClientset(tc.existingStatefulSet)
			} else {
				clientSet = fake.NewClientset()
			}

			ssetClient := clientSet.AppsV1().StatefulSets(tc.newStatefulSet.Namespace)

			_, err := CreateStatefulSetOrPatchLabels(ctx, ssetClient, tc.newStatefulSet)
			require.NoError(t, err)

			// Verify the statefulset in the cluster has the expected labels
			result, err := ssetClient.Get(ctx, tc.newStatefulSet.Name, metav1.GetOptions{})
			require.NoError(t, err)
			require.Equal(t, tc.expectedLabels, result.Labels)
		})
	}
}
