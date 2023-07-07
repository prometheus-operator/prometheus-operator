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

package k8sutil

import (
	"context"
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUniqueVolumeName(t *testing.T) {
	cases := []struct {
		prefix   string
		name     string
		expected string
		err      bool
	}{
		{
			name:     "@$!!@$%!#$%!#$%!#$!#$%%$#@!#",
			expected: "",
			err:      true,
		},
		{
			name:     "NAME",
			expected: "name-4cfd3574",
		},
		{
			name:     "foo--",
			expected: "foo-e705c7c8",
		},
		{
			name:     "foo^%#$bar",
			expected: "foo-bar-f3e212b1",
		},
		{
			name:     "fOo^%#$bar",
			expected: "foo-bar-ee5c3c18",
		},
		{
			name: strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: strings.Repeat("a", validation.DNS1123LabelMaxLength-9) +
				"-4ed69ce2",
		},
		{
			prefix:   "with-prefix",
			name:     "name",
			expected: "with-prefix-name-6c5f7b2e",
		},
		{
			prefix:   "with-prefix-",
			name:     "name",
			expected: "with-prefix-name-6c5f7b2e",
		},
		{
			prefix:   "with-prefix",
			name:     strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: "with-prefix-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-4ed69ce2",
		},
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rn := ResourceNamer{prefix: c.prefix}

			out, err := rn.UniqueDNS1123Label(c.name)
			if c.err {
				if err == nil {
					t.Errorf("expecting error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expecting no error, got %v", err)
			}

			if c.expected != out {
				t.Errorf("expected test case %d to be %q but got %q", i, c.expected, out)
			}
		})
	}
}

func TestUniqueVolumeNameCollision(t *testing.T) {
	// a<63>-foo
	foo := strings.Repeat("a", validation.DNS1123LabelMaxLength) + "foo"
	// a<63>-bar
	bar := strings.Repeat("a", validation.DNS1123LabelMaxLength) + "bar"

	rn := ResourceNamer{}

	fooSanitized, err := rn.UniqueDNS1123Label(foo)
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}

	barSanitized, err := rn.UniqueDNS1123Label(bar)
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}

	if fooSanitized == barSanitized {
		t.Fatalf("expected sanitized volume name of %q and %q to be different but got %q", foo, bar, fooSanitized)
	}
}

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

			err := UpdateStatefulSet(ctx, ssetClient, modifiedSset)
			if err != nil {
				t.Fatal(err)
			}

			updatedSset, err := ssetClient.Get(ctx, "prometheus", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expected, updatedSset.Spec.Template.Annotations) {
				t.Errorf("expected annotations %q, got %q", tc.expected, updatedSset.Spec.Template.Annotations)
			}
		})
	}
}

func TestMergeMetadata(t *testing.T) {
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

	t.Run("CreateOrUpdateService", func(t *testing.T) {
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
				for l, v := range tc.modifiedLabels {
					modifiedSvc.Labels[l] = v
				}
				for a, v := range tc.modifiedAnnotations {
					modifiedSvc.Annotations[a] = v
				}
				_, err := svcClient.Update(context.Background(), modifiedSvc, metav1.UpdateOptions{})
				if err != nil {
					t.Fatal(err)
				}

				err = CreateOrUpdateService(context.Background(), svcClient, service)
				if err != nil {
					t.Fatal(err)
				}

				updatedSvc, err := svcClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedAnnotations, updatedSvc.Annotations) {
					t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSvc.Annotations)
				}
				if !reflect.DeepEqual(tc.expectedLabels, updatedSvc.Labels) {
					t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSvc.Labels)
				}
			})
		}
	})

	t.Run("CreateOrUpdateEndpoints", func(t *testing.T) {
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
				for l, v := range tc.modifiedLabels {
					modifiedEndpoints.Labels[l] = v
				}
				for a, v := range tc.modifiedAnnotations {
					modifiedEndpoints.Annotations[a] = v
				}
				_, err := endpointsClient.Update(context.Background(), modifiedEndpoints, metav1.UpdateOptions{})
				if err != nil {
					t.Fatal(err)
				}

				err = CreateOrUpdateEndpoints(context.Background(), endpointsClient, endpoints)
				if err != nil {
					t.Fatal(err)
				}

				updatedEndpoints, err := endpointsClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedAnnotations, updatedEndpoints.Annotations) {
					t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedEndpoints.Annotations)
				}
				if !reflect.DeepEqual(tc.expectedLabels, updatedEndpoints.Labels) {
					t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedEndpoints.Labels)
				}
			})
		}
	})

	t.Run("UpdateStatefulSet", func(t *testing.T) {
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
				for l, v := range tc.modifiedLabels {
					modifiedSset.Labels[l] = v
				}
				for a, v := range tc.modifiedAnnotations {
					modifiedSset.Annotations[a] = v
				}
				_, err := ssetClient.Update(context.Background(), modifiedSset, metav1.UpdateOptions{})
				if err != nil {
					t.Fatal(err)
				}

				err = UpdateStatefulSet(context.Background(), ssetClient, sset)
				if err != nil {
					t.Fatal(err)
				}

				updatedSset, err := ssetClient.Get(context.Background(), "prometheus", metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedAnnotations, updatedSset.Annotations) {
					t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSset.Annotations)
				}
				if !reflect.DeepEqual(tc.expectedLabels, updatedSset.Labels) {
					t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSset.Labels)
				}
			})
		}
	})

	t.Run("CreateOrUpdateSecret", func(t *testing.T) {
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
				for l, v := range tc.modifiedLabels {
					modifiedSecret.Labels[l] = v
				}
				for a, v := range tc.modifiedAnnotations {
					modifiedSecret.Annotations[a] = v
				}
				_, err := sClient.Update(context.Background(), modifiedSecret, metav1.UpdateOptions{})
				if err != nil {
					t.Fatal(err)
				}

				err = CreateOrUpdateSecret(context.Background(), sClient, secret)
				if err != nil {
					t.Fatal(err)
				}

				updatedSecret, err := sClient.Get(context.Background(), "prometheus-tls-assets", metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedAnnotations, updatedSecret.Annotations) {
					t.Errorf("expected annotations %q, got %q", tc.expectedAnnotations, updatedSecret.Annotations)
				}
				if !reflect.DeepEqual(tc.expectedLabels, updatedSecret.Labels) {
					t.Errorf("expected labels %q, got %q", tc.expectedLabels, updatedSecret.Labels)
				}
			})
		}
	})
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

		if err := CreateOrUpdateService(context.TODO(), svcClient, modifiedSvc); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(service.Spec.IPFamilies, modifiedSvc.Spec.IPFamilies) {
			t.Fatalf("services Spec.IPFamilies are not equal, expected %q, got %q",
				service.Spec.IPFamilies, modifiedSvc.Spec.IPFamilies)
		}

		if !reflect.DeepEqual(service.Spec.ClusterIP, modifiedSvc.Spec.ClusterIP) {
			t.Fatalf("services Spec.ClusterIP are not equal, expected %q, got %q",
				service.Spec.ClusterIP, modifiedSvc.Spec.ClusterIP)
		}

		if !reflect.DeepEqual(service.Spec.ClusterIPs, modifiedSvc.Spec.ClusterIPs) {
			t.Fatalf("services Spec.ClusterIPs are not equal, expected %q, got %q",
				service.Spec.ClusterIPs, modifiedSvc.Spec.ClusterIPs)
		}

		if !reflect.DeepEqual(service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy) {
			t.Fatalf("services Spec.IPFamilyPolicy are not equal, expected %v, got %v",
				service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy)
		}

		if !reflect.DeepEqual(service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy) {
			t.Fatalf("services Spec.IPFamilyPolicy are not equal, expected %v, got %v",
				service.Spec.IPFamilyPolicy, modifiedSvc.Spec.IPFamilyPolicy)
		}
	})
}
