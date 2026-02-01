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
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResolveStuckStatefulSet addresses stuck rollouts by inspecting pods and applying the configured RepairPolicy.
func ResolveStuckStatefulSet(ctx context.Context, logger *slog.Logger, kclient kubernetes.Interface, sset *appsv1.StatefulSet, policy RepairPolicy) error {
	if policy == "" || policy == NoneRepairPolicy {
		return nil
	}

	if sset.Generation != sset.Status.ObservedGeneration {
		logger.Debug("statefulset spec not yet reconciled, skipping repair", "name", sset.Name)
		return nil
	}

	selector, err := metav1.LabelSelectorAsSelector(sset.Spec.Selector)
	if err != nil {
		return err
	}

	pods, err := kclient.CoreV1().Pods(sset.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return err
	}

	// Iterate in reverse ordinal order.
	podItems := pods.Items
	sortPodsByOrdinal(podItems)

	for i := len(podItems) - 1; i >= 0; i-- {
		pod := podItems[i]

		if isPodReady(pod) {
			continue
		}

		revision := pod.Labels[appsv1.ControllerRevisionHashLabelKey]
		// The pod needs to be repaired only if its revision matches neither the current nor the updated revision.
		if revision == sset.Status.CurrentRevision || revision == sset.Status.UpdateRevision {
			continue
		}

		logger.Info("found stuck pod during rollout, repairing", "pod", pod.Name, "policy", policy, "revision", revision, "currentRevision", sset.Status.CurrentRevision, "updateRevision", sset.Status.UpdateRevision)

		switch policy {
		case EvictRepairPolicy:
			err := kclient.CoreV1().Pods(pod.Namespace).EvictV1(ctx, &policyv1.Eviction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err)
			}
		case DeleteRepairPolicy:
			propagationPolicy := metav1.DeletePropagationBackground
			err := kclient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			})
			if err != nil {
				return fmt.Errorf("failed to delete pod %s/%s: %w", pod.Namespace, pod.Name, err)
			}
		}

		// Repair only one pod per invocation.
		return nil
	}

	return nil
}

func sortPodsByOrdinal(pods []v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		pi, _ := getOrdinal(pods[i].Name)
		pj, _ := getOrdinal(pods[j].Name)
		return pi < pj
	})
}

func getOrdinal(name string) (int, error) {
	dash := strings.LastIndex(name, "-")
	if dash == -1 {
		return 0, fmt.Errorf("no dash found in pod name %s", name)
	}
	return strconv.Atoi(name[dash+1:])
}

func isPodReady(pod v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
