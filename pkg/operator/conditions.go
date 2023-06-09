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

package operator

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// FindStatusCondition returns the condition matching the given type.
// If the condition isn't present, it returns nil.
func FindStatusCondition(conditions []monitoringv1.Condition, conditionType monitoringv1.ConditionType) *monitoringv1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

// UpdateConditions merges the existing conditions with newConditions.
func UpdateConditions(conditions []monitoringv1.Condition, newConditions ...monitoringv1.Condition) []monitoringv1.Condition {
	ret := make([]monitoringv1.Condition, 0, len(conditions))

	for _, nc := range newConditions {
		c := FindStatusCondition(conditions, nc.Type)
		if c == nil {
			ret = append(ret, nc)
			continue
		}

		if nc.Status == c.Status {
			// Retain the last transition time if the status of the condition hasn't changed.
			nc.LastTransitionTime = c.LastTransitionTime
		}
		ret = append(ret, nc)
	}

	return ret
}
