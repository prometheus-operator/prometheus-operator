// Copyright 2025 The prometheus-operator Authors
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

package telemetry

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestSpanNameFormatting(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		expectedSpan string
	}{
		{
			name:         "namespaced resource with specific names",
			method:       "DELETE",
			path:         "/api/v1/namespaces/alertmanager/secrets/alertmanager-shadow-alertmanager-tls-assets-1",
			expectedSpan: "DELETE /api/v1/namespaces/{namespace}/secrets/{name}",
		},
		{
			name:         "namespaced statefulsets",
			method:       "PUT",
			path:         "/api/v1/namespaces/default/statefulsets/prometheus-test",
			expectedSpan: "PUT /api/v1/namespaces/{namespace}/statefulsets/{name}",
		},
		{
			name:         "cluster resource",
			method:       "GET",
			path:         "/api/v1/nodes/my-node-name",
			expectedSpan: "GET /api/v1/nodes/{name}",
		},
		{
			name:         "list resources without specific name",
			method:       "GET",
			path:         "/api/v1/namespaces/default/pods",
			expectedSpan: "GET /api/v1/namespaces/{namespace}/pods",
		},
		{
			name:         "custom resource with api group",
			method:       "POST",
			path:         "/apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/test-prometheus",
			expectedSpan: "POST /apis/monitoring.coreos.com/v1/namespaces/{namespace}/prometheuses/{name}",
		},
		{
			name:         "resource with subresource",
			method:       "PATCH",
			path:         "/api/v1/namespaces/default/pods/test-pod/status",
			expectedSpan: "PATCH /api/v1/namespaces/{namespace}/pods/{name}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Method: tt.method,
				URL: &url.URL{
					Path: tt.path,
				},
			}

			// Create a simple formatter that mimics what we do in WrapRoundTripper
			formatter := func(operation string, r *http.Request) string {
				path := r.URL.Path

				// Replace specific namespace and resource names with placeholders to reduce cardinality
				if matches := namespacedResourcePathRegex.FindStringSubmatch(path); len(matches) >= 5 {
					// matches[1] = api version part, matches[2] = namespace, matches[3] = resource type, matches[4] = resource name
					path = fmt.Sprintf("%s/namespaces/{namespace}/%s/{name}", matches[1], matches[3])
				} else if matches := namespacedCollectionPathRegex.FindStringSubmatch(path); len(matches) >= 4 {
					// matches[1] = api version part, matches[2] = namespace, matches[3] = resource type
					path = fmt.Sprintf("%s/namespaces/{namespace}/%s", matches[1], matches[3])
				} else if matches := clusterResourcePathRegex.FindStringSubmatch(path); len(matches) >= 4 {
					// matches[1] = api version part, matches[2] = resource type, matches[3] = resource name
					// Exclude namespace-related paths which should be handled by the first regex
					if matches[2] != "namespaces" {
						path = fmt.Sprintf("%s/%s/{name}", matches[1], matches[2])
					}
				}

				return fmt.Sprintf("%s %s", r.Method, path)
			}

			result := formatter("", req)
			if result != tt.expectedSpan {
				t.Errorf("Expected span name %q, got %q", tt.expectedSpan, result)
			}
		})
	}
}
