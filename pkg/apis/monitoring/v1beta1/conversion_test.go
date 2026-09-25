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

package v1beta1

import (
	"testing"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestConvertRouteLabels(t *testing.T) {
	tests := []struct {
		name           string
		inputLabels    []KeyValue
		expectedLabels []KeyValue
	}{
		{
			name:           "nil labels",
			inputLabels:    nil,
			expectedLabels: nil,
		},
		{
			name:           "empty labels",
			inputLabels:    []KeyValue{},
			expectedLabels: []KeyValue{},
		},
		{
			name: "simple labels",
			inputLabels: []KeyValue{
				{Key: "severity", Value: "critical"},
				{Key: "team", Value: "platform"},
			},
			expectedLabels: []KeyValue{
				{Key: "severity", Value: "critical"},
				{Key: "team", Value: "platform"},
			},
		},
		{
			name: "labels with template expressions",
			inputLabels: []KeyValue{
				{Key: "environment", Value: "{{ .Labels.env }}"},
				{Key: "region", Value: "{{ .ExternalLabels.region }}"},
			},
			expectedLabels: []KeyValue{
				{Key: "environment", Value: "{{ .Labels.env }}"},
				{Key: "region", Value: "{{ .ExternalLabels.region }}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("convertRouteFrom preserves labels", func(t *testing.T) {
				var alphaLabels []v1alpha1.KeyValue
				if tc.inputLabels != nil {
					alphaLabels = make([]v1alpha1.KeyValue, len(tc.inputLabels))
					for i, kv := range tc.inputLabels {
						alphaLabels[i] = v1alpha1.KeyValue{Key: kv.Key, Value: kv.Value}
					}
				}

				alphaRoute := &v1alpha1.Route{
					Receiver: "test",
					Labels:   alphaLabels,
				}

				betaRoute, err := convertRouteFrom(alphaRoute)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tc.inputLabels == nil {
					if betaRoute.Labels != nil {
						t.Fatalf("expected nil labels, got %v", betaRoute.Labels)
					}
					return
				}

				if len(betaRoute.Labels) != len(tc.expectedLabels) {
					t.Fatalf("expected %d labels, got %d", len(tc.expectedLabels), len(betaRoute.Labels))
				}

				for i, expected := range tc.expectedLabels {
					if betaRoute.Labels[i].Key != expected.Key || betaRoute.Labels[i].Value != expected.Value {
						t.Errorf("expected label[%d] = {%s, %s}, got {%s, %s}",
							i, expected.Key, expected.Value,
							betaRoute.Labels[i].Key, betaRoute.Labels[i].Value)
					}
				}
			})

			t.Run("convertRouteTo preserves labels", func(t *testing.T) {
				betaRoute := &Route{
					Receiver: "test",
					Labels:   tc.inputLabels,
				}

				alphaRoute, err := convertRouteTo(betaRoute)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tc.inputLabels == nil {
					if alphaRoute.Labels != nil {
						t.Fatalf("expected nil labels, got %v", alphaRoute.Labels)
					}
					return
				}

				if len(alphaRoute.Labels) != len(tc.expectedLabels) {
					t.Fatalf("expected %d labels, got %d", len(tc.expectedLabels), len(alphaRoute.Labels))
				}

				for i, expected := range tc.expectedLabels {
					if alphaRoute.Labels[i].Key != expected.Key || alphaRoute.Labels[i].Value != expected.Value {
						t.Errorf("expected label[%d] = {%s, %s}, got {%s, %s}",
							i, expected.Key, expected.Value,
							alphaRoute.Labels[i].Key, alphaRoute.Labels[i].Value)
					}
				}
			})
		})
	}
}
