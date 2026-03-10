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

package k8s

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelSelectionHasChanged(t *testing.T) {
	for _, tc := range []struct {
		name string

		old      map[string]string
		current  map[string]string
		selector *metav1.LabelSelector

		expected    bool
		expectedErr bool
	}{
		{
			name: "no label change",
			old: map[string]string{
				"app": "foo",
			},
			current: map[string]string{
				"app": "foo",
			},
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "foo"},
			},
		},
		{
			name: "old matches and current doesn't match",
			old: map[string]string{
				"app": "foo",
			},
			current: map[string]string{
				"app": "bar",
			},
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "foo"},
			},
			expected: true,
		},
		{
			name: "old doesn't match and current matches",
			old: map[string]string{
				"app": "bar",
			},
			current: map[string]string{
				"app": "foo",
			},
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "foo"},
			},
			expected: true,
		},
		{
			name: "match-all label selector",
			old: map[string]string{
				"app": "foo",
			},
			current: map[string]string{
				"app": "bar",
			},
			selector: &metav1.LabelSelector{},
		},
		{
			name: "match-nothing label selector",
			old: map[string]string{
				"app": "foo",
			},
			current: map[string]string{
				"app": "bar",
			},
			selector: nil,
		},
		{
			name: "invalid label selector",
			old: map[string]string{
				"app": "foo",
			},
			current: map[string]string{
				"app": "bar",
			},
			selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "foo",
						Operator: metav1.LabelSelectorOperator("invalid"),
					},
				},
			},
			expectedErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed, err := LabelSelectionHasChanged(tc.old, tc.current, tc.selector)

			if tc.expectedErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tc.expected != changed {
				t.Errorf("expected %v, got %v", tc.expected, changed)
			}
		})
	}
}
