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

package alertmanager

import (
	"context"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func (c *Operator) resolveStuckStatefulSet(ctx context.Context, logger *slog.Logger, am *monitoringv1.Alertmanager, sset *appsv1.StatefulSet) error {
	if am.Spec.UpdateStrategy == nil || am.Spec.UpdateStrategy.RollingUpdate == nil || am.Spec.UpdateStrategy.RollingUpdate.RepairPolicy == nil {
		return nil
	}

	policy := *am.Spec.UpdateStrategy.RollingUpdate.RepairPolicy
	if policy == monitoringv1.NoneRepairPolicy {
		return nil
	}

	selector, err := metav1.LabelSelectorAsSelector(sset.Spec.Selector)
	if err != nil {
		return err
	}

	pods, err := c.kclient.CoreV1().Pods(sset.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if isPodReady(pod) {
			continue
		}

		if pod.Labels[appsv1.ControllerRevisionHashLabelKey] == sset.Status.UpdateRevision {
			continue
		}

		logger.Info("found not ready pod during rollout", "pod", pod.Name, "policy", policy)

		switch policy {
		case monitoringv1.EvictNotReadyPodsRepairPolicy:
			err := c.kclient.CoreV1().Pods(pod.Namespace).EvictV1(ctx, &policyv1.Eviction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				},
			})
			if err != nil {
				logger.Error("failed to evict pod", "pod", pod.Name, "err", err)
			}
		case monitoringv1.DeleteNotReadyPodsRepairPolicy:
			// Set propagation policy to Background to delete the pod immediately.
			propagationPolicy := metav1.DeletePropagationBackground
			err := c.kclient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			})
			if err != nil {
				logger.Error("failed to delete pod", "pod", pod.Name, "err", err)
			}
		}
	}

	return nil
}

func isPodReady(pod v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
