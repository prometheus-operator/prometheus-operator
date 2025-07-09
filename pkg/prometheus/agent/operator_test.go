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

package prometheusagent

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestValidateDaemonSetModeSpec(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		spec           monitoringv1alpha1.PrometheusAgentSpec
		expectError    bool
		errorSubstring string
	}{
		{
			name: "invalid: configuring replicas in the daemonset mode",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas: ptr.To(int32(3)),
				},
			},
			expectError:    true,
			errorSubstring: "replicas cannot be set when mode is DaemonSet",
		},
		{
			name: "invalid: configuring storage in the daemonset mode",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Storage: &monitoringv1.StorageSpec{
						VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
							Spec: v1.PersistentVolumeClaimSpec{
								Resources: v1.VolumeResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
			expectError:    true,
			errorSubstring: "storage cannot be configured when mode is DaemonSet",
		},
		{
			name: "invalid: configuring shards in the daemonset mode",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Shards: ptr.To(int32(1)),
				},
			},
			expectError:    true,
			errorSubstring: "shards cannot be set when mode is DaemonSet",
		},
		{
			name: "invalid: configuring persistentVolumeClaimRetentionPolicy in the daemonset mode",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
						WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
						WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
					},
				},
			},
			expectError:    true,
			errorSubstring: "persistentVolumeClaimRetentionPolicy cannot be set when mode is DaemonSet",
		},
		{
			name: "valid daemonset configuration",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version:            "v2.45.0",
					ServiceAccountName: "prometheus",
					PodMonitorSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"group": "test",
						},
					},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser: ptr.To(int64(1000)),
					},
				},
			},
			expectError: false,
		},
		{
			name: "minimal valid spec should not error",
			spec: monitoringv1alpha1.PrometheusAgentSpec{
				Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			},
			expectError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &monitoringv1alpha1.PrometheusAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-agent",
					Namespace: "test",
				},
				Spec: tc.spec,
			}

			err := validateDaemonSetModeSpec(p)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorSubstring)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
