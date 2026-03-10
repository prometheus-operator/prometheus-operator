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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinalizerAddPatch(t *testing.T) {
	finalizerName := "cleanup.kubernetes.io/finalizer"
	tests := []struct {
		name          string
		finalizers    []string
		finalizerName string
		expectedPatch []map[string]any
		expectEmpty   bool
	}{
		{
			name:          "empty finalizers",
			finalizers:    []string{},
			finalizerName: finalizerName,
			expectedPatch: []map[string]any{
				{"op": "add", "path": "/metadata/finalizers", "value": []string{finalizerName}},
			},
		},
		{
			name:          "finalizer not present",
			finalizers:    []string{"a", "b"},
			finalizerName: finalizerName,
			expectedPatch: []map[string]any{
				{"op": "add", "path": "/metadata/finalizers/-", "value": finalizerName},
			},
		},
		{
			name:          "finalizer already present",
			finalizers:    []string{"a", finalizerName, "b"},
			finalizerName: finalizerName,
			expectEmpty:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := FinalizerAddPatch(tt.finalizers, tt.finalizerName)
			require.NoError(t, err)

			if tt.expectEmpty {
				require.Empty(t, patch)
			} else {
				expectedBytes, err := json.Marshal(tt.expectedPatch)
				require.NoError(t, err)
				require.JSONEq(t, string(expectedBytes), string(patch))
			}
		})
	}
}

func TestFinalizerDeletePatch(t *testing.T) {
	finalizerName := "cleanup.kubernetes.io/finalizer"
	tests := []struct {
		name          string
		finalizers    []string
		finalizerName string
		expectPatch   bool
		expectedIndex int
	}{
		{
			name:          "finalizer present at index 1",
			finalizers:    []string{"a", finalizerName, "b"},
			finalizerName: finalizerName,
			expectPatch:   true,
			expectedIndex: 1,
		},
		{
			name:          "finalizer not present",
			finalizers:    []string{"a", "b"},
			finalizerName: finalizerName,
			expectPatch:   false,
		},
		{
			name:          "empty finalizers",
			finalizers:    []string{},
			finalizerName: finalizerName,
			expectPatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := FinalizerDeletePatch(tt.finalizers, tt.finalizerName)
			require.NoError(t, err)

			if tt.expectPatch {
				// The patch should include a "test" operation before "remove" to ensure
				// the finalizer at the index matches the expected value, preventing
				// race conditions from removing the wrong finalizer.
				expected := []map[string]any{
					{"op": "test", "path": fmt.Sprintf("/metadata/finalizers/%d", tt.expectedIndex), "value": tt.finalizerName},
					{"op": "remove", "path": fmt.Sprintf("/metadata/finalizers/%d", tt.expectedIndex)},
				}
				expectedBytes, err := json.Marshal(expected)
				require.NoError(t, err)
				require.JSONEq(t, string(expectedBytes), string(patch))
			} else {
				require.Empty(t, patch)
			}
		})
	}
}
