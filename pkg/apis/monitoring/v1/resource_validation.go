// Copyright 2024 The prometheus-operator Authors
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

package v1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// ValidateResourceRequirements validates that resource requests do not exceed limits.
func ValidateResourceRequirements(resources v1.ResourceRequirements) error {
	return validateResourceRequirements(resources, "")
}

// validateResourceRequirements validates that resource requests do not exceed limits.
// The prefix parameter is used to provide context in error messages.
func validateResourceRequirements(resources v1.ResourceRequirements, prefix string) error {
	if len(resources.Requests) == 0 || len(resources.Limits) == 0 {
		return nil // No validation needed if either requests or limits are empty
	}

	// Check CPU requests vs limits
	if cpuRequest, hasRequest := resources.Requests[v1.ResourceCPU]; hasRequest {
		if cpuLimit, hasLimit := resources.Limits[v1.ResourceCPU]; hasLimit {
			if cpuRequest.Cmp(cpuLimit) > 0 {
				msg := "CPU requests must not exceed CPU limits"
				if prefix != "" {
					msg = fmt.Sprintf("%s: %s", prefix, msg)
				}
				return fmt.Errorf("%s (request: %s, limit: %s)", msg, cpuRequest.String(), cpuLimit.String())
			}
		}
	}

	// Check Memory requests vs limits
	if memRequest, hasRequest := resources.Requests[v1.ResourceMemory]; hasRequest {
		if memLimit, hasLimit := resources.Limits[v1.ResourceMemory]; hasLimit {
			if memRequest.Cmp(memLimit) > 0 {
				msg := "Memory requests must not exceed memory limits"
				if prefix != "" {
					msg = fmt.Sprintf("%s: %s", prefix, msg)
				}
				return fmt.Errorf("%s (request: %s, limit: %s)", msg, memRequest.String(), memLimit.String())
			}
		}
	}

	return nil
}
