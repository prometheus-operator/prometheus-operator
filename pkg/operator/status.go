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
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	// NoSelectedResourcesReason is used in status conditions to indicate that
	// a workload resource selected no configuration resources.
	NoSelectedResourcesReason = "NoSelectedResources"

	// DeprecatedFieldsInUseReason is used in status conditions to indicate that
	// the resource uses deprecated fields.
	DeprecatedFieldsInUseReason = "DeprecatedFieldsInUse"
)

// StatusGetter represents a workload resource implementing the interface
// required by StatusPoller.
type StatusGetter interface {
	metav1.Object
	GetConditions() []monitoringv1.Condition
	ExpectedReplicas() int
	GetUpdatedReplicas() int
	GetAvailableReplicas() int
}

// StatusReconciler can walk through all workload resources (Iterate) and
// trigger a status reconciliation (RefreshStatusFor).
type StatusReconciler interface {
	Iterate(func(StatusGetter))
	RefreshStatusFor(metav1.Object)
}

// StatusPoller refreshes regularly the workload resources for which:
//   - the Available condition isn't True.
//   - the number of updated and available replicas don't match the expected
//     replica number.
//
// It ensures that the status subresource gets eventually reconciled. For
// instance when a new version of the statefulset is rolled out and the updated
// pod has non-ready containers, the statefulset status won't see any update
// because the number of ready/updated replicas doesn't change. Without the
// periodic refresh, the object's status would report "containers with
// incomplete status: [init-config-reloader]" forever.
// It can also be that the updated/available replica fields aren't updated as
// they should be due to races in the controller logic.
func StatusPoller(ctx context.Context, sr StatusReconciler) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sr.Iterate(func(resource StatusGetter) {
				replicas := resource.ExpectedReplicas()
				if replicas != resource.GetUpdatedReplicas() || replicas != resource.GetAvailableReplicas() {
					sr.RefreshStatusFor(resource)
				}

				for _, cond := range resource.GetConditions() {
					if cond.Type == monitoringv1.Available && cond.Status != monitoringv1.ConditionTrue {
						sr.RefreshStatusFor(resource)
						return
					}
				}
			})
		}
	}
}
