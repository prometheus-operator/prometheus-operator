// Copyright 2021 The prometheus-operator Authors
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

package operator

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

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



type mockBindingRemover struct {
	removeBindingFunc func(context.Context, ConfigurationObject) error
}

func (m *mockBindingRemover) RemoveBinding(ctx context.Context, co ConfigurationObject) error {
	return m.removeBindingFunc(ctx, co)
}

func TestPruneOrphanedBindings(t *testing.T) {
	gvr := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheuses",
	}

	tests := []struct {
		name            string
		bindings        []monitoringv1.WorkloadBinding
		workloads       []string
		expectedRemoved []string
	}{
		{
			name: "remove orphaned binding",
			bindings: []monitoringv1.WorkloadBinding{
				{
					Group:     gvr.Group,
					Resource:  gvr.Resource,
					Namespace: "default",
					Name:      "prom-1",
				},
			},
			workloads:       []string{},
			expectedRemoved: []string{"default/prom-1"},
		},
		{
			name: "keep existing binding",
			bindings: []monitoringv1.WorkloadBinding{
				{
					Group:     gvr.Group,
					Resource:  gvr.Resource,
					Namespace: "default",
					Name:      "prom-1",
				},
			},
			workloads:       []string{"default/prom-1"},
			expectedRemoved: []string{},
		},
		{
			name: "ignore other resources",
			bindings: []monitoringv1.WorkloadBinding{
				{
					Group:     "other.group",
					Resource:  "other",
					Namespace: "default",
					Name:      "other-1",
				},
			},
			workloads:       []string{},
			expectedRemoved: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &monitoringv1.ServiceMonitor{}
			cr.Status.Bindings = tt.bindings
			cr.SetNamespace("default")
			cr.SetName("test-config")
			
			// AddTypeInformationToObject called in PruneOrphanedBindings - if it fails, test fails.
			// assuming it works
			cr.SetGroupVersionKind(monitoringv1.SchemeGroupVersion.WithKind(monitoringv1.ServiceMonitorsKind))

			listerFunc := func(selector labels.Selector, appendFunc cache.AppendFunc) error {
				appendFunc(cr)
				return nil
			}

			workloadChecker := func(ns, name string) (bool, error) {
				return slices.Contains(tt.workloads, ns+"/"+name), nil
			}

			removedBindings := []string{}
			factory := func(workload RuntimeObject) BindingRemover {
				return &mockBindingRemover{
					removeBindingFunc: func(ctx context.Context, co ConfigurationObject) error {
						if co == cr {
							removedBindings = append(removedBindings, workload.GetNamespace()+"/"+workload.GetName())
						}
						return nil
					},
				}
			}

			err := PruneOrphanedBindings[*monitoringv1.ServiceMonitor](
				context.Background(),
				listerFunc,
				gvr,
				workloadChecker,
				factory,
			)

			require.NoError(t, err)
			require.ElementsMatch(t, tt.expectedRemoved, removedBindings)
		})
	}
}

