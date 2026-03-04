// Copyright 2026 The prometheus-operator Authors
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
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusCleanupFinalizerName is the name of the finalizer used to garbage
// collect status bindings on configuration resources.
const StatusCleanupFinalizerName = "monitoring.coreos.com/status-cleanup"

// FinalizerAddPatch generates the JSON patch payload which adds the finalizer to the object's metadata.
// If the finalizer is already present, it returns an empty []byte slice.
func FinalizerAddPatch(finalizers []string, finalizerName string) ([]byte, error) {
	if slices.Contains(finalizers, finalizerName) {
		return []byte{}, nil
	}
	if len(finalizers) == 0 {
		patch := []map[string]any{
			{
				"op":    "add",
				"path":  "/metadata/finalizers",
				"value": []string{finalizerName},
			},
		}
		return json.Marshal(patch)
	}
	patch := []map[string]any{
		{
			"op":    "add",
			"path":  "/metadata/finalizers/-",
			"value": finalizerName,
		},
	}
	return json.Marshal(patch)
}

// FinalizerDeletePatch generates a JSON Patch payload to remove the specified
// finalizer from an object's metadata.
//
// If the finalizer is not present, the function returns nil.
//
// The patch includes a "test" operation before "remove" to ensure the value at
// the computed index matches the expected finalizer. This prevents race
// conditions when finalizers are modified concurrently.
func FinalizerDeletePatch(finalizers []string, finalizerName string) ([]byte, error) {
	for i, f := range finalizers {
		if f == finalizerName {
			patch := []map[string]any{
				{
					"op":    "test",
					"path":  fmt.Sprintf("/metadata/finalizers/%d", i),
					"value": finalizerName,
				},
				{
					"op":   "remove",
					"path": fmt.Sprintf("/metadata/finalizers/%d", i),
				},
			}
			return json.Marshal(patch)
		}
	}
	return nil, nil
}

func HasStatusCleanupFinalizer(obj metav1.Object) bool {
	return slices.Contains(obj.GetFinalizers(), StatusCleanupFinalizerName)
}
