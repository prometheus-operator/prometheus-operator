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

package prometheus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

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
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("200Mi"),
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
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("100Mi"),
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
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("100Mi"),
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

			p2Hash, err = createSSetInputHash(tc.a, c, []string{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{Replicas: new(int32(2))})
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

func TestProcessShardRetention(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		retentionPoliciesEnabled bool
		spec                     monitoringv1.PrometheusSpec
		annotations              map[string]string
		injectPatchError         bool

		expectedDelete           bool
		expectedErr              bool
		expectedPatch            bool
		expectedDeadlineIsZero   bool
		expectedDeadlineDuration time.Duration
	}{
		{
			name:                     "feature gate disabled",
			retentionPoliciesEnabled: false,
			spec:                     monitoringv1.PrometheusSpec{},
			expectedDelete:           true,
		},
		{
			// Regression test: should not panic when ShardRetentionPolicy is nil
			name:                     "feature gate enabled but ShardRetentionPolicy is nil",
			retentionPoliciesEnabled: true,
			spec:                     monitoringv1.PrometheusSpec{},
			expectedDelete:           true,
		},
		{
			name:                     "feature gate enabled with empty ShardRetentionPolicy",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{},
			},
			expectedDelete: true,
		},
		{
			name:                     "feature gate enabled with WhenScaled set to Delete",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.DeleteWhenScaledRetentionType),
				},
			},
			expectedDelete: true,
		},
		{
			name:                     "WhenScaled set to Retain and no annotation",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDelete:           false,
			expectedPatch:            true,
			expectedDeadlineDuration: 24 * time.Hour,
		},
		{
			name:                     "WhenScaled set to Retain and deadline in the future",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			annotations:    map[string]string{deletionDeadlineAnnotation: time.Now().UTC().Add(24 * time.Hour).Format(annotationTimeFormat)},
			expectedDelete: false,
			expectedPatch:  false,
		},
		{
			name:                     "WhenScaled set to Retain with size-only retention",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				RetentionSize: "10Gi",
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDelete:         false,
			expectedPatch:          true,
			expectedDeadlineIsZero: true,
		},
		{
			name:                     "patch failure returns error",
			retentionPoliciesEnabled: true,
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			injectPatchError: true,
			expectedErr:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "prometheus-example",
					Namespace:   "test",
					Annotations: tc.annotations,
				},
			}
			if sset.Annotations == nil {
				sset.Annotations = map[string]string{}
			}
			kclient := fake.NewSimpleClientset(sset)
			if tc.injectPatchError {
				kclient.Fake.PrependReactor("patch", "statefulsets", func(_ clienttesting.Action) (bool, k8sruntime.Object, error) {
					return true, nil, errors.New("patch failed")
				})
			}
			o := &Operator{
				retentionPoliciesEnabled: tc.retentionPoliciesEnabled,
				kclient:                  kclient,
			}

			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "test",
				},
				Spec: tc.spec,
			}

			shouldDelete, err := o.processShardRetention(context.Background(), p, sset)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.expectedDelete {
				require.True(t, shouldDelete)
			} else {
				require.False(t, shouldDelete)
			}

			if tc.expectedPatch {
				actions := kclient.Actions()
				require.Equal(t, 1, len(actions))
				patchAction, ok := actions[0].(clienttesting.PatchAction)
				require.True(t, ok)

				if tc.expectedDeadlineIsZero || tc.expectedDeadlineDuration > 0 {
					var body struct {
						Metadata struct {
							Annotations map[string]string `json:"annotations"`
						} `json:"metadata"`
					}
					require.NoError(t, json.Unmarshal(patchAction.GetPatch(), &body))
					deadline, err := time.Parse(annotationTimeFormat, body.Metadata.Annotations[deletionDeadlineAnnotation])
					require.NoError(t, err)
					if tc.expectedDeadlineIsZero {
						require.True(t, deadline.IsZero(), "expected zero deadline, got %s", deadline)
					} else {
						require.WithinDuration(t, time.Now().Add(tc.expectedDeadlineDuration), deadline, 5*time.Second)
					}
				}
			} else {
				require.Equal(t, 0, len(kclient.Actions()))
			}
		})
	}
}

func TestGracePeriodForPrometheusStorage(t *testing.T) {
	for _, tc := range []struct {
		name             string
		spec             monitoringv1.PrometheusSpec
		expectedDuration time.Duration
		expectedErr      bool
	}{
		{
			name: "empty retention uses default (24h)",
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDuration: 24 * time.Hour,
		},
		{
			name: "explicit retain retention duration",
			spec: monitoringv1.PrometheusSpec{
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
					Retain: &monitoringv1.RetainConfig{
						RetentionPeriod: monitoringv1.Duration("15d"),
					},
				},
			},
			expectedDuration: 15 * 24 * time.Hour,
		},
		{
			name: "explicit retention duration",
			spec: monitoringv1.PrometheusSpec{
				Retention: "15d",
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDuration: 15 * 24 * time.Hour,
		},
		{
			name: "size-only retention returns zero duration",
			spec: monitoringv1.PrometheusSpec{
				RetentionSize: "10Gi",
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDuration: 0,
		},
		{
			name: "size and time retention uses time-based value",
			spec: monitoringv1.PrometheusSpec{
				Retention:     "7d",
				RetentionSize: "10Gi",
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedDuration: 7 * 24 * time.Hour,
		},
		{
			name: "invalid retention returns error",
			spec: monitoringv1.PrometheusSpec{
				Retention: "invalid",
				ShardRetentionPolicy: &monitoringv1.ShardRetentionPolicy{
					WhenScaled: new(monitoringv1.RetainWhenScaledRetentionType),
				},
			},
			expectedErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{Spec: tc.spec}
			d, err := gracePeriodForPrometheusStorage(p)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedDuration, d)
		})
	}
}
