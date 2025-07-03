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

	t.Run("ReplicasValidation", testValidateDaemonSetModeSpec_Replicas)
	t.Run("StorageValidation", testValidateDaemonSetModeSpec_Storage)
	t.Run("ShardsValidation", testValidateDaemonSetModeSpec_Shards)
	t.Run("PVCRetentionValidation", testValidateDaemonSetModeSpec_PVCRetention)
	t.Run("ValidSpecNoErrors", testValidateDaemonSetModeSpec_ValidSpec)
}

func testValidateDaemonSetModeSpec_Replicas(t *testing.T) {
	t.Parallel()

	// setting a replica
	p := &monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: ptr.To(int32(3)),
			},
		},
	}

	err := validateDaemonSetModeSpec(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "replicas cannot be set when mode is DaemonSet")

	p.Spec.Replicas = nil
	err = validateDaemonSetModeSpec(p)
	require.NoError(t, err)
}

func testValidateDaemonSetModeSpec_Storage(t *testing.T) {
	t.Parallel()

	// setting a storage field
	p := &monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
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
	}

	err := validateDaemonSetModeSpec(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "storage cannot be configured when mode is DaemonSet")

	p.Spec.Storage = nil
	err = validateDaemonSetModeSpec(p)
	require.NoError(t, err)
}

func testValidateDaemonSetModeSpec_Shards(t *testing.T) {
	t.Parallel()

	// setting a shards field >1
	p := &monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Shards: ptr.To(int32(2)),
			},
		},
	}

	err := validateDaemonSetModeSpec(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "shards cannot be greater than 1 when mode is DaemonSet")

	// setting a shards field = 1
	p.Spec.Shards = ptr.To(int32(1))
	err = validateDaemonSetModeSpec(p)
	require.NoError(t, err)

	p.Spec.Shards = nil
	err = validateDaemonSetModeSpec(p)
	require.NoError(t, err)
}

func testValidateDaemonSetModeSpec_PVCRetention(t *testing.T) {
	t.Parallel()

	// setting a PVC retention policy
	p := &monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
					WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
					WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
				},
			},
		},
	}

	err := validateDaemonSetModeSpec(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "persistentVolumeClaimRetentionPolicy cannot be set when mode is DaemonSet")

	p.Spec.PersistentVolumeClaimRetentionPolicy = nil
	err = validateDaemonSetModeSpec(p)
	require.NoError(t, err)
}

func testValidateDaemonSetModeSpec_ValidSpec(t *testing.T) {
	t.Parallel()

	// setting a valid spec
	p := &monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "test",
		},
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
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
	}

	err := validateDaemonSetModeSpec(p)
	require.NoError(t, err)
}

func testValidateDaemonSetModeSpec_MultipleInvalidFields(t *testing.T) {
	t.Parallel()

	// setting multiple invalid fields
	p := &monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: ptr.To(int32(3)),
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
				Shards: ptr.To(int32(2)),
			},
		},
	}

	err := validateDaemonSetModeSpec(p)
	require.Error(t, err)
	// Should return the first validation error (replicas)
	require.Contains(t, err.Error(), "replicas cannot be set when mode is DaemonSet")
}

func TestValidateDaemonSetModeSpec_MultipleInvalidFields(t *testing.T) {
	testValidateDaemonSetModeSpec_MultipleInvalidFields(t)
}
