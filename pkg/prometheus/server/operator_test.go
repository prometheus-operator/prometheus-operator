// Copyright 2017 The prometheus-operator Authors
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
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		require.True(t, (o.LabelSelector == "app.kubernetes.io/name=prometheus,prometheus=test" || o.LabelSelector == "prometheus=test,app.kubernetes.io/name=prometheus"), "LabelSelector not computed correctly\n\nExpected: \"app.kubernetes.io/name=prometheus,prometheus=test\"\n\nGot:      %#+v", o.LabelSelector)
	}
}

func TestCreateStatefulSetInputHash(t *testing.T) {
	falseVal := false

	for _, tc := range []struct {
		name string
		a, b monitoringv1.Prometheus

		equal bool
	}{
		{
			name: "different generations",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
		{
			// differrent resource.Quantity produce the same hash because the
			// struct contains private fields that aren't integrated into the
			// hash computation.
			name: "different specs but same hash",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
			equal: true,
		},
		{
			name: "same hash with different status",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Paused:  true,
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Paused:  true,
						Version: "v1.7.2",
					},
				},
				Status: monitoringv1.PrometheusStatus{
					Paused: true,
				},
			},

			equal: true,
		},
		{
			name: "different labels",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
		{
			name: "different annotations",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
		{
			name: "different web http2",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
						Web: &monitoringv1.PrometheusWebSpec{
							WebConfigFileFields: monitoringv1.WebConfigFileFields{
								HTTPConfig: &monitoringv1.WebHTTPConfig{
									HTTP2: &falseVal,
								},
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := prompkg.Config{}

			p1Hash, err := createSSetInputHash(tc.a, c, []string{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{})
			require.NoError(t, err)

			p2Hash, err := createSSetInputHash(tc.b, c, []string{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{})
			require.NoError(t, err)

			if !tc.equal {
				require.NotEqual(t, p1Hash, p2Hash, "expected two different Prometheus CRDs to produce different hashes but got equal hash")
				return
			}

			require.Equal(t, p1Hash, p2Hash, "expected two Prometheus CRDs to produce the same hash but got different hash")

			p2Hash, err = createSSetInputHash(tc.a, c, []string{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{Replicas: ptr.To(int32(2))})
			require.NoError(t, err)

			require.NotEqual(t, p1Hash, p2Hash, "expected same Prometheus CRDs with different statefulset specs to produce different hashes but got equal hash")
		})
	}
}

func TestCreateThanosConfigSecret(t *testing.T) {
	version := "v0.24.0"
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		spec monitoringv1.PrometheusSpec
	}{
		{
			name: "prometheus with thanos sidecar",
			spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &version,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-thanos-config-secret",
					Namespace: "test",
				},
				Spec: tc.spec,
			}
			o := Operator{kclient: fake.NewClientset()}
			err := o.createOrUpdateThanosConfigSecret(ctx, p)
			require.NoError(t, err)

			get, err := o.kclient.CoreV1().Secrets("test").Get(ctx, thanosPrometheusHTTPClientConfigSecretName(p), metav1.GetOptions{})
			require.NoError(t, err)
			require.Equal(t, "tls_config:\n  insecure_skip_verify: true\n", string(get.Data[thanosPrometheusHTTPClientConfigFileName]))
		})
	}
}

func TestShouldDelete(t *testing.T) {
	o := &Operator{}
	for _, tc := range []struct {
		name               string
		featureGateEnabled bool
		retentionPolicy    monitoringv1.WhenScaledRetentionType
		shouldDelete       bool
	}{
		{
			name:               "feature gate disabled - always delete",
			featureGateEnabled: false,
			shouldDelete:       true,
		},
		{
			name:               "delete policy",
			featureGateEnabled: true,
			retentionPolicy:    monitoringv1.WhenScaledRetentionTypeDelete,
			shouldDelete:       true,
		},
		{
			name:               "retain policy",
			featureGateEnabled: true,
			retentionPolicy:    monitoringv1.WhenScaledRetentionTypeRetain,
			shouldDelete:       false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o.retentionPoliciesEnabled = tc.featureGateEnabled
			p := monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
						WhenScaled: tc.retentionPolicy,
					},
				},
			}

			decision, err := o.shouldDelete(&p)
			require.NoError(t, err)
			require.Equal(t, tc.shouldDelete, decision)
		})
	}
}
