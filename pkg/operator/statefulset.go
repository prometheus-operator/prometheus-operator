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

package operator

import (
	appsv1 "k8s.io/api/apps/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// UpdateStrategyForStatefulSet converts a monitoring update strategy to a statefulset update strategy.
func UpdateStrategyForStatefulSet(updateStrategy *monitoringv1.StatefulSetUpdateStrategy) appsv1.StatefulSetUpdateStrategy {
	if updateStrategy == nil {
		return appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		}
	}

	converted := appsv1.StatefulSetUpdateStrategy{
		Type: appsv1.StatefulSetUpdateStrategyType(updateStrategy.Type),
	}
	if updateStrategy.RollingUpdate != nil {
		converted.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{
			MaxUnavailable: updateStrategy.RollingUpdate.MaxUnavailable,
		}
	}

	return converted
}
