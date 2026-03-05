// Copyright 2022 The prometheus-operator Authors
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
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type RepairPolicy string

const (
	NoneRepairPolicy   RepairPolicy = "none"
	EvictRepairPolicy  RepairPolicy = "evict"
	DeleteRepairPolicy RepairPolicy = "delete"
)

// Set implements the flag.Value interface.
func (p *RepairPolicy) Set(value string) error {
	*p = RepairPolicy(value)

	switch *p {
	case NoneRepairPolicy:
	case EvictRepairPolicy:
	case DeleteRepairPolicy:
	default:
		return fmt.Errorf("invalid value: %s", value)
	}

	return nil
}

func (p *RepairPolicy) String() string { return string(*p) }

// Pod is an alias for the Kubernetes v1.Pod type.
type Pod corev1.Pod

// Ready returns true if the pod is ready.
func (p *Pod) Ready() bool {
	if p.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, cond := range p.Status.Conditions {
		if cond.Type != corev1.PodReady {
			continue
		}
		return cond.Status == corev1.ConditionTrue
	}

	return false
}

// Message returns a human-readable and terse message about the state of the pod.
// If the pod is ready, it returns an empty string.
func (p *Pod) Message() string {
	for _, condType := range []corev1.PodConditionType{
		corev1.PodScheduled,    // Check first that the pod is scheduled.
		corev1.PodInitialized,  // Then that init containers have been started successfully.
		corev1.ContainersReady, // Then that all containers are ready.
		corev1.PodReady,        // And finally that the pod is ready.
	} {
		for _, cond := range p.Status.Conditions {
			if cond.Type == condType && cond.Status == corev1.ConditionFalse {
				return cond.Message
			}
		}
	}

	return ""
}

type StatefulSetReporter struct {
	kclient        kubernetes.Interface
	Pods           []Pod
	sset           *appsv1.StatefulSet
	allowedRepairs int
}

// NewStatefulSetReporter returns a statefulset's reporter.
func NewStatefulSetReporter(ctx context.Context, kclient kubernetes.Interface, sset *appsv1.StatefulSet) (*StatefulSetReporter, error) {
	if sset == nil {
		// sset is nil when the controller couldn't create the statefulset
		// (incompatible spec fields for instance).
		return &StatefulSetReporter{}, nil
	}

	ls, err := metav1.LabelSelectorAsSelector(sset.Spec.Selector)
	if err != nil {
		// Something is really broken if the statefulset's selector isn't valid.
		panic(err)
	}

	podList, err := kclient.CoreV1().Pods(sset.Namespace).List(ctx, metav1.ListOptions{LabelSelector: ls.String()})
	if err != nil {
		return nil, err
	}

	pods := make([]Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		var found bool
		for _, owner := range p.OwnerReferences {
			if owner.Kind == "StatefulSet" && owner.Name == sset.Name {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		pods = append(pods, Pod(p))
	}

	if err := sortByOrdinalAsc(pods); err != nil {
		return nil, err
	}

	return &StatefulSetReporter{
		kclient: kclient,
		sset:    sset,
		Pods:    pods,
		// Allow at most 1 repair operation per reporter.
		// The expectation is that if additional repairs are needed, they
		// happen in a future reconciliation loop which will create a new
		// reporter (otherwise the information about statefulset and pods will
		// be stale).
		allowedRepairs: 1,
	}, nil
}

func sortByOrdinalAsc(pods []Pod) error {
	var sortErr error
	slices.SortStableFunc(pods, func(i, j Pod) int {
		oi, err := i.getOrdinal()
		if sortErr == nil {
			sortErr = err
		}

		oj, err := j.getOrdinal()
		if sortErr == nil {
			sortErr = err
		}

		return cmp.Compare(oi, oj)
	})

	return sortErr
}

func (p *Pod) getOrdinal() (int, error) {
	// Preferred solution if the PodIndexLabel feature gate is enabled (GA
	// since 1.32).
	if idx, found := p.Labels[appsv1.PodIndexLabel]; found {
		return strconv.Atoi(idx)
	}

	// Otherwise try to guess from the pod name.
	dash := strings.LastIndex(p.Name, "-")
	if dash == -1 {
		return 0, fmt.Errorf("no dash found in pod name %s", p.Name)
	}

	return strconv.Atoi(p.Name[dash+1:])
}

// UpdatedPods returns the list of pods that match with the statefulset's revision.
func (sr *StatefulSetReporter) UpdatedPods() []Pod {
	return sr.filterPods(func(p Pod) bool {
		return sr.IsUpdated(p)
	})
}

// IsUpdated returns true if the given pod matches with the statefulset's revision.
func (sr *StatefulSetReporter) IsUpdated(p Pod) bool {
	return sr.sset.Status.UpdateRevision == p.Labels["controller-revision-hash"]
}

// ReadyPods returns the list of pods that are ready.
func (sr *StatefulSetReporter) ReadyPods() []Pod {
	return sr.filterPods(func(p Pod) bool {
		return p.Ready()
	})
}

func (sr *StatefulSetReporter) filterPods(f func(Pod) bool) []Pod {
	pods := make([]Pod, 0, len(sr.Pods))

	for _, p := range sr.Pods {
		if f(p) {
			pods = append(pods, p)
		}
	}

	return pods
}

type GoverningObject interface {
	metav1.Object
	ExpectedReplicas() int
	SetReplicas(int)
	SetUpdatedReplicas(int)
	SetAvailableReplicas(int)
	SetUnavailableReplicas(int)
}

// Update updates the status replica fields of the resource governing the
// statefulset (e.g.  the Prometheus operator's workload resource) and returns
// the Available status condition.
func (sr *StatefulSetReporter) Update(gObj GoverningObject) monitoringv1.Condition {
	condition := monitoringv1.Condition{
		Type:   monitoringv1.Available,
		Status: monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: gObj.GetGeneration(),
	}

	var (
		ready   = len(sr.ReadyPods())
		updated = len(sr.UpdatedPods())
	)
	gObj.SetReplicas(len(sr.Pods))
	gObj.SetUpdatedReplicas(updated)
	gObj.SetAvailableReplicas(ready)
	gObj.SetUnavailableReplicas(len(sr.Pods) - ready)

	condition.Status, condition.Reason = sr.StatusAndReasonForAvailableCondition(gObj.ExpectedReplicas())

	var messages []string
	for _, p := range sr.Pods {
		if m := p.Message(); m != "" {
			messages = append(messages, fmt.Sprintf("pod %s: %s", p.Name, m))
		}
	}
	condition.Message = strings.Join(messages, "\n")

	return condition
}

// StatusAndReasonForAvailableCondition computes the status and reason for the
// resource governing the statefulset based on the expected number of replicas
// and the state of the pods.
func (sr *StatefulSetReporter) StatusAndReasonForAvailableCondition(expectedReplicas int) (monitoringv1.ConditionStatus, string) {
	var (
		status = monitoringv1.ConditionTrue
		reason string
		ready  = len(sr.ReadyPods())
	)
	switch {
	case sr.sset == nil:
		reason = "StatefulSetNotFound"
		status = monitoringv1.ConditionFalse
	case ready < expectedReplicas:
		switch ready {
		case 0:
			reason = "NoPodReady"
			status = monitoringv1.ConditionFalse
		default:
			reason = "SomePodsNotReady"
			status = monitoringv1.ConditionDegraded
		}
	}

	return status, reason
}

// Repair checks if the statefulset is stuck and if yes, evicts/deletes the
// first pod which isn't ready.
// The function will update at most one pod to avoid further disruption.
func (sr *StatefulSetReporter) Repair(ctx context.Context, logger *slog.Logger, policy RepairPolicy) error {
	sset := sr.sset
	logger = logger.With("policy", policy)
	if sset == nil {
		logger.Warn("skipping because the statefulset couldn't be found")
		return nil
	}

	if sr.allowedRepairs <= 0 {
		logger.Warn("skipping because no more repairs are allowed")
		return nil
	}

	logger = logger.With("statefulset", fmt.Sprintf("%s/%s", sset.Namespace, sset.Name))

	if policy == NoneRepairPolicy {
		logger.Debug("skipping because repair policy is none")
		return nil
	}

	if sset.Generation != sset.Status.ObservedGeneration {
		logger.Info("skipping because statefulset is not yet reconciled")
		return nil
	}

	// Iterate pods in reverse ordinal order.
	for _, pod := range slices.Backward(sr.Pods) {
		if pod.Ready() {
			logger.Debug("pod is ready", "pod", pod.Name)
			continue
		}

		revision := pod.Labels[appsv1.ControllerRevisionHashLabelKey]
		// The pod should not be repaired if its revision matches either the
		// current or the updated revision.
		if revision == sset.Status.CurrentRevision {
			logger.Debug("pod revision == current revision", "pod", pod.Name)
			continue
		}

		if revision == sset.Status.UpdateRevision {
			logger.Debug("pod revision == update revision", "pod", pod.Name)
			continue
		}

		if pod.DeletionTimestamp != nil {
			logger.Debug("pod deletion in progress", "pod", pod.Name)
			continue
		}

		logger.Info("found pod which needs to be repaired, applying repair policy", "pod", pod.Name)
		deleteOpts := metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationBackground)}
		switch policy {
		case EvictRepairPolicy:
			err := sr.kclient.CoreV1().Pods(pod.Namespace).EvictV1(ctx, &policyv1.Eviction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				},
				DeleteOptions: &deleteOpts,
			})
			if err != nil {
				return fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err)
			}
		case DeleteRepairPolicy:
			err := sr.kclient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, deleteOpts)
			if err != nil {
				return fmt.Errorf("failed to delete pod %s/%s: %w", pod.Namespace, pod.Name, err)
			}
		}

		// Repair only one pod per invocation.
		sr.allowedRepairs--
		return nil
	}

	logger.Debug("all pods are ok")
	return nil
}
