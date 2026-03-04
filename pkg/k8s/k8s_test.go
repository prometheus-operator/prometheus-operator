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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMergeMetadata_CreateOrUpdateSecret(t *testing.T) {
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
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "prometheus-tls-assets",
					Namespace:   namespace,
					Labels:      map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
					Annotations: map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
				},
			}

			sClient := fake.NewSimpleClientset(secret).CoreV1().Secrets(namespace)

			modifiedSecret := secret.DeepCopy()
			maps.Copy(modifiedSecret.Labels, tc.modifiedLabels)
			maps.Copy(modifiedSecret.Annotations, tc.modifiedAnnotations)
			_, err := sClient.Update(context.Background(), modifiedSecret, metav1.UpdateOptions{})
			require.NoError(t, err)

			err = CreateOrUpdateSecret(context.Background(), sClient, secret)
			require.NoError(t, err)

			updatedSecret, err := sClient.Get(context.Background(), "prometheus-tls-assets", metav1.GetOptions{})
			require.NoError(t, err)

			if !reflect.DeepEqual(tc.expectedAnnotations, updatedSecret.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSecret.Annotations)
			}
			if !reflect.DeepEqual(tc.expectedLabels, updatedSecret.Labels) {
				t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSecret.Labels)
			}
		})
	}
}
