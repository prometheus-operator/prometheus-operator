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
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// Pod is an alias for the Kubernetes v1.Pod type.
type Pod v1.Pod

// Ready returns true if the pod is ready.
func (p *Pod) Ready() bool {
	if p.Status.Phase != v1.PodRunning {
		return false
	}

	for _, cond := range p.Status.Conditions {
		if cond.Type != v1.PodReady {
			continue
		}
		return cond.Status == v1.ConditionTrue
	}

	return false
}

// Message returns a human-readable and terse message about the state of the pod.
func (p *Pod) Message() string {
	for _, condType := range []v1.PodConditionType{
		v1.PodScheduled,    // Check first that the pod is scheduled.
		v1.PodInitialized,  // Then that init containers have been started successfully.
		v1.ContainersReady, // Then that all containers are ready.
		v1.PodReady,        // And finally that the pod is ready.
	} {
		for _, cond := range p.Status.Conditions {
			if cond.Type == condType && cond.Status == v1.ConditionFalse {
				return cond.Message
			}
		}
	}

	return ""
}

type StatefulSetReporter struct {
	Pods []*Pod
	sset *appsv1.StatefulSet
}

// UpdatedPods returns the list of pods that match with the statefulset's revision.
func (sr *StatefulSetReporter) UpdatedPods() []*Pod {
	return sr.filterPods(func(p *Pod) bool {
		return sr.IsUpdated(p)
	})
}

// IsUpdated returns true if the given pod matches with the statefulset's revision.
func (sr *StatefulSetReporter) IsUpdated(p *Pod) bool {
	return sr.sset.Status.UpdateRevision == p.Labels["controller-revision-hash"]
}

// ReadyPods returns the list of pods that are ready.
func (sr *StatefulSetReporter) ReadyPods() []*Pod {
	return sr.filterPods(func(p *Pod) bool {
		return p.Ready()
	})
}

func (sr *StatefulSetReporter) filterPods(f func(*Pod) bool) []*Pod {
	pods := make([]*Pod, 0, len(sr.Pods))

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
		ready     = len(sr.ReadyPods())
		updated   = len(sr.UpdatedPods())
		available = len(sr.ReadyPods())
	)
	gObj.SetReplicas(len(sr.Pods))
	gObj.SetUpdatedReplicas(updated)
	gObj.SetAvailableReplicas(ready)
	gObj.SetUnavailableReplicas(len(sr.Pods) - ready)

	if ready < gObj.ExpectedReplicas() {
		if available == 0 {
			condition.Reason = "NoPodReady"
			condition.Status = monitoringv1.ConditionFalse
		} else {
			condition.Reason = "SomePodsNotReady"
			condition.Status = monitoringv1.ConditionDegraded
		}
	}

	var messages []string
	for _, p := range sr.Pods {
		if m := p.Message(); m != "" {
			messages = append(messages, fmt.Sprintf("pod %s: %s", p.Name, m))
		}
	}

	condition.Message = strings.Join(messages, "\n")

	return condition
}

// NewStatefulSetReporter returns a statefulset's reporter.
func NewStatefulSetReporter(ctx context.Context, kclient kubernetes.Interface, sset *appsv1.StatefulSet) (*StatefulSetReporter, error) {
	ls, err := metav1.LabelSelectorAsSelector(sset.Spec.Selector)
	if err != nil {
		// Something is really broken if the statefulset's selector isn't valid.
		panic(err)
	}

	pods, err := kclient.CoreV1().Pods(sset.Namespace).List(ctx, metav1.ListOptions{LabelSelector: ls.String()})
	if err != nil {
		return nil, err
	}

	stsReporter := &StatefulSetReporter{
		sset: sset,
		Pods: make([]*Pod, 0, len(pods.Items)),
	}
	for _, p := range pods.Items {
		var found bool
		for _, owner := range p.ObjectMeta.OwnerReferences {
			if owner.Kind == "StatefulSet" && owner.Name == sset.Name {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		stsReporter.Pods = append(stsReporter.Pods, (func(p Pod) *Pod { return &p })(Pod(p)))
	}

	return stsReporter, nil
}
