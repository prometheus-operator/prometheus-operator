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
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMergeMetadata_CreateOrUpdateService(t *testing.T) {
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
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "prometheus-operated",
					Namespace:   namespace,
					Labels:      map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
					Annotations: map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
				},
				Spec:   corev1.ServiceSpec{},
				Status: corev1.ServiceStatus{},
			}

			svcClient := fake.NewSimpleClientset(service).CoreV1().Services(namespace)

			modifiedSvc := service.DeepCopy()
			maps.Copy(modifiedSvc.Labels, tc.modifiedLabels)
			maps.Copy(modifiedSvc.Annotations, tc.modifiedAnnotations)
			_, err := svcClient.Update(context.Background(), modifiedSvc, metav1.UpdateOptions{})
			require.NoError(t, err)

			_, err = CreateOrUpdateService(context.Background(), svcClient, service)
			require.NoError(t, err)

			updatedSvc, err := svcClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
			require.NoError(t, err)

			if !reflect.DeepEqual(tc.expectedAnnotations, updatedSvc.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSvc.Annotations)
			}
			if !reflect.DeepEqual(tc.expectedLabels, updatedSvc.Labels) {
				t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSvc.Labels)
			}
		})
	}
}

func TestMergeMetadata_CreateOrUpdateEndpoints(t *testing.T) {
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
			endpoints := &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "prometheus-operated",
					Namespace:   namespace,
					Labels:      map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
					Annotations: map[string]string{"app.kubernetes.io/name": "kube-state-metrics"},
				},
			}

			endpointsClient := fake.NewSimpleClientset(endpoints).CoreV1().Endpoints(namespace)

			modifiedEndpoints := endpoints.DeepCopy()
			maps.Copy(modifiedEndpoints.Labels, tc.modifiedLabels)
			maps.Copy(modifiedEndpoints.Annotations, tc.modifiedAnnotations)
			_, err := endpointsClient.Update(context.Background(), modifiedEndpoints, metav1.UpdateOptions{})
			require.NoError(t, err)

			err = CreateOrUpdateEndpoints(context.Background(), endpointsClient, endpoints)
			require.NoError(t, err)

			updatedEndpoints, err := endpointsClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
			require.NoError(t, err)

			if !reflect.DeepEqual(tc.expectedAnnotations, updatedEndpoints.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedEndpoints.Annotations)
			}
			if !reflect.DeepEqual(tc.expectedLabels, updatedEndpoints.Labels) {
				t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedEndpoints.Labels)
			}
		})
	}
}

func TestCreateOrUpdateImmutableFields(t *testing.T) {
	namespace := "default"
	policy := corev1.IPFamilyPolicyRequireDualStack

	t.Run("CreateOrUpdateService with immutable fields", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-operated-test",
				Namespace: namespace,
			},
			Spec: corev1.ServiceSpec{
				ClusterIP: "127.0.0.1",
				ClusterIPs: []string{
					"127.0.0.1",
					"192.168.0.159",
				},
				IPFamilyPolicy: &policy,
				IPFamilies: []corev1.IPFamily{
					corev1.IPv6Protocol,
				},
				Ports: []corev1.ServicePort{
					{
						Name: "https-metrics",
						Port: 10250,
					},
					{
						Name: "http-metrics",
						Port: 10255,
					},
				},
			},
			Status: corev1.ServiceStatus{},
		}

		svcClient := fake.NewSimpleClientset(service).CoreV1().Services(namespace)

		modifiedSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-operated-test",
				Namespace: namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name: "https-metrics",
						Port: 10250,
					},
				},
			},
			Status: corev1.ServiceStatus{},
		}

		_, err := CreateOrUpdateService(context.TODO(), svcClient, modifiedSvc)
		require.NoError(t, err)

		require.Equal(t, service.Spec.IPFamilies, modifiedSvc.Spec.IPFamilies, "services Spec.IPFamilies are not equal, expected %q, got %q",
			service.Spec.IPFamilies, modifiedSvc.Spec.IPFamilies)

		require.Equal(t, service.Spec.ClusterIP, modifiedSvc.Spec.ClusterIP, "services Spec.ClusterIP are not equal, expected %q, got %q",
			service.Spec.ClusterIP, modifiedSvc.Spec.ClusterIP)

		require.Equal(t, service.Spec.ClusterIPs, modifiedSvc.Spec.ClusterIPs, "services Spec.ClusterIPs are not equal, expected %q, got %q",
			service.Spec.ClusterIPs, modifiedSvc.Spec.ClusterIPs)

		require.Equal(t, service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy, "services Spec.IPFamilyPolicy are not equal, expected %v, got %v",
			service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy)
	})
}

func TestEnsureCustomGoverningService(t *testing.T) {
	name := "test-k8sutil"
	serviceName := "test-svc"
	ns := "test-ns"
	testcases := []struct {
		name           string
		service        corev1.Service
		selectorLabels map[string]string
		expectedErr    bool
	}{
		{
			name: "custom service selects k8sutil",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"k8sutil":                      name,
						"app.kubernetes.io/name":       "k8sutil",
						"app.kubernetes.io/instance":   name,
						"app.kubernetes.io/managed-by": "prometheus-operator",
					},
				},
			},
			selectorLabels: map[string]string{
				"k8sutil":                      name,
				"app.kubernetes.io/name":       "k8sutil",
				"app.kubernetes.io/instance":   name,
				"app.kubernetes.io/managed-by": "prometheus-operator",
			},
			expectedErr: false,
		},
		{
			name: "custom service does not select k8sutil",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"k8sutil":                      "different-name",
						"app.kubernetes.io/name":       "k8sutil",
						"app.kubernetes.io/instance":   "different-name",
						"app.kubernetes.io/managed-by": "prometheus-operator",
					},
				},
			},
			selectorLabels: map[string]string{
				"k8sutil":                      name,
				"app.kubernetes.io/name":       "k8sutil",
				"app.kubernetes.io/instance":   name,
				"app.kubernetes.io/managed-by": "prometheus-operator",
			},
			expectedErr: true,
		},
		{
			name: "custom service selects k8sutil but in different ns",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "wrong-ns",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"k8sutil":                      name,
						"app.kubernetes.io/name":       "k8sutil",
						"app.kubernetes.io/instance":   name,
						"app.kubernetes.io/managed-by": "prometheus-operator",
					},
				},
			},
			selectorLabels: map[string]string{
				"k8sutil":                      name,
				"app.kubernetes.io/name":       "k8sutil",
				"app.kubernetes.io/instance":   name,
				"app.kubernetes.io/managed-by": "prometheus-operator",
			},
			expectedErr: true,
		},
		{
			name: "custom svc doesn't exist",
			selectorLabels: map[string]string{
				"k8sutil":                      name,
				"app.kubernetes.io/name":       "k8sutil",
				"app.kubernetes.io/instance":   name,
				"app.kubernetes.io/managed-by": "prometheus-operator",
			},
			expectedErr: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := makeBarebonesPrometheus(name, ns)
			p.Spec.ServiceName = &serviceName

			clientSet := fake.NewSimpleClientset(&tc.service)
			svcClient := clientSet.CoreV1().Services(ns)

			err := EnsureCustomGoverningService(context.Background(), p.Namespace, *p.Spec.ServiceName, svcClient, tc.selectorLabels)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func makeBarebonesPrometheus(name, ns string) *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: map[string]string{},
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: ptr.To(int32(1)),
			},
		},
	}
}
