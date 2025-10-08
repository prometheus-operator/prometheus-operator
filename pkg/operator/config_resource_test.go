package operator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestConfigResStatusConditionsEqual(t *testing.T) {
	now := metav1.NewTime(time.Now())
	earlier := metav1.NewTime(time.Now().Add(-1 * time.Hour))

	tests := []struct {
		name     string
		a        []monitoringv1.ConfigResourceCondition
		b        []monitoringv1.ConfigResourceCondition
		expected bool
	}{
		{
			name: "equal conditions different order",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "True",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
					LastTransitionTime: now,
				},
				{
					Type:               "Ready",
					Status:             "False",
					Reason:             "Init",
					Message:            "initializing",
					ObservedGeneration: 1,
					LastTransitionTime: earlier,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Ready",
					Status:             "False",
					Reason:             "Init",
					Message:            "initializing",
					ObservedGeneration: 1,
					LastTransitionTime: earlier,
				},
				{
					Type:               "Accepted",
					Status:             "True",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
					LastTransitionTime: now,
				},
			},
			expected: true,
		},
		{
			name: "different status",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "True",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False", // different
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "different message",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "Issue detected", // different
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "different observed generation",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 2, // different
				},
			},
			expected: false,
		},
		{
			name: "different reason",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "Issue", // different
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "different type",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Ready", // different
					Status:             "False",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "different lengths",
			a: []monitoringv1.ConfigResourceCondition{
				{
					Type:               "Accepted",
					Status:             "True",
					Reason:             "OK",
					Message:            "all good",
					ObservedGeneration: 1,
				},
			},
			b:        []monitoringv1.ConfigResourceCondition{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equalConfigResourceConditions(tt.a, tt.b)
			require.Equal(t, tt.expected, result)
		})
	}
}
