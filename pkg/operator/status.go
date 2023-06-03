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

type StatusReconciler interface {
	Iterate(func(metav1.Object, []monitoringv1.Condition))
	RefreshStatusFor(metav1.Object)
}

// StatusPoller refreshes regularly the objects for which the Available
// condition isn't True. It ensures that the status subresource eventually
// reflects the pods conditions.
// For instance when a new version of the statefulset is rolled out and the
// updated pod has non-ready containers, the statefulset status won't see
// any update because the number of ready/updated replicas doesn't change.
// Without the periodic refresh, the object's status would report "containers
// with incomplete status: [init-config-reloader]" forever.
func StatusPoller(ctx context.Context, sr StatusReconciler) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sr.Iterate(func(meta metav1.Object, conditions []monitoringv1.Condition) {
				for _, cond := range conditions {
					if cond.Type == monitoringv1.Available && cond.Status != monitoringv1.ConditionTrue {
						sr.RefreshStatusFor(meta)
						break
					}
				}
			})
		}
	}
}
