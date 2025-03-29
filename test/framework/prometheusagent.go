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

package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	wait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheusagent "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/agent"
)

func (f *Framework) MakeBasicPrometheusAgent(ns, name, group string, replicas int32) *monitoringv1alpha1.PrometheusAgent {
	return &monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: map[string]string{},
		},
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: &replicas,
				Version:  operator.DefaultPrometheusVersion,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": group,
					},
				},
				PodMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": group,
					},
				},
				ServiceAccountName: "prometheus",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
			},
		},
	}
}

func (f *Framework) MakeBasicPrometheusAgentDaemonSet(ns, name string) *monitoringv1alpha1.PrometheusAgent {
	return &monitoringv1alpha1.PrometheusAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: map[string]string{},
		},
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			Mode: ptr.To(monitoringv1alpha1.DaemonSetPrometheusAgentMode),
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Version:            operator.DefaultPrometheusVersion,
				ServiceAccountName: "prometheus",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
				PodMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": name,
					},
				},
			},
		},
	}
}

func (f *Framework) CreatePrometheusAgentAndWaitUntilReady(ctx context.Context, ns string, p *monitoringv1alpha1.PrometheusAgent) (*monitoringv1alpha1.PrometheusAgent, error) {
	result, err := f.MonClientV1alpha1.PrometheusAgents(ns).Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v prometheus-agent instances failed (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	if ptr.Deref(p.Spec.Mode, "StatefulSet") == "DaemonSet" {
		err = f.WaitForPrometheusAgentDSReady(ctx, ns, p)
		if err != nil {
			return nil, fmt.Errorf("waiting for prometheus-agent DaemonSet timed out (%v): %v", p.Name, err)
		}
	} else {
		result, err = f.WaitForPrometheusAgentReady(ctx, result, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("waiting for %v prometheus-agent instances timed out (%v): %v", p.Spec.Replicas, p.Name, err)
		}
	}

	return result, nil
}

func (f *Framework) WaitForPrometheusAgentReady(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, timeout time.Duration) (*monitoringv1alpha1.PrometheusAgent, error) {
	expected := *p.Spec.Replicas
	if p.Spec.Shards != nil && *p.Spec.Shards > 0 {
		expected = expected * *p.Spec.Shards
	}

	var current *monitoringv1alpha1.PrometheusAgent
	var getErr error
	if err := f.WaitForResourceAvailable(
		ctx,
		func(context.Context) (resourceStatus, error) {
			current, getErr = f.MonClientV1alpha1.PrometheusAgents(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
			if getErr != nil {
				return resourceStatus{}, getErr
			}
			return resourceStatus{
				expectedReplicas: expected,
				generation:       current.Generation,
				replicas:         current.Status.UpdatedReplicas,
				conditions:       current.Status.Conditions,
			}, nil
		},
		timeout,
	); err != nil {
		return nil, fmt.Errorf("prometheus-agent %v/%v failed to become available: %w", p.Namespace, p.Name, err)
	}

	return current, nil
}

func (f *Framework) WaitForPrometheusAgentDSReady(ctx context.Context, ns string, p *monitoringv1alpha1.PrometheusAgent) error {
	var pollErr error
	if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		name := fmt.Sprintf("prom-agent-%s", p.Name)
		// TODO: Implement UpdateStatus() for DaemonSet and check status instead of using Get().
		dms, err := f.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			pollErr = fmt.Errorf("failed to get Prometheus Agent DaemonSet: %w", err)
			return false, nil
		}

		if dms.DeletionTimestamp != nil {
			pollErr = fmt.Errorf("prometheus Agent DaemonSet deletion in progress")
			return false, nil
		}

		if dms.Status.NumberUnavailable > 0 {
			pollErr = fmt.Errorf("prometheus Agent DaemonSet is not available")
			return false, nil
		}

		if dms.Status.NumberReady == 0 {
			pollErr = fmt.Errorf("prometheus Agent DaemonSet is not ready")
			return false, nil
		}

		return true, nil
	}); err != nil {
		return fmt.Errorf("%v: %w", pollErr, err)
	}

	return nil
}

func (f *Framework) DeletePrometheusAgentAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1alpha1.PrometheusAgents(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("requesting PrometheusAgent custom resource %v failed: %w", name, err)
	}

	if err := f.MonClientV1alpha1.PrometheusAgents(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("deleting PrometheusAgent custom resource %v failed: %w", name, err)
	}

	if err := f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		0,
		prometheusagent.ListOptions(name),
	); err != nil {
		return fmt.Errorf("waiting for PrometheusAgent custom resource (%s) to vanish timed out: %w", name, err)
	}

	return nil
}

func (f *Framework) DeletePrometheusAgentDSAndWaitUntilGone(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, ns, name string) error {
	if err := f.MonClientV1alpha1.PrometheusAgents(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("deleting PrometheusAgent custom resource %v failed: %w", name, err)
	}

	var pollErr error
	if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		dmsName := fmt.Sprintf("prom-agent-%s", p.Name)
		dms, _ := f.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, dmsName, metav1.GetOptions{})
		if dms.Status.NumberAvailable != 0 {
			pollErr = fmt.Errorf("prometheus Agent DaemonSet still exists after deleting")
			return false, nil
		}

		return true, nil
	}); err != nil {
		return fmt.Errorf("%v: %w", pollErr, err)
	}

	return nil
}

func (f *Framework) PatchPrometheusAgent(ctx context.Context, name, ns string, spec monitoringv1alpha1.PrometheusAgentSpec) (*monitoringv1alpha1.PrometheusAgent, error) {
	b, err := json.Marshal(
		&monitoringv1alpha1.PrometheusAgent{
			TypeMeta: metav1.TypeMeta{
				Kind:       monitoringv1alpha1.PrometheusAgentsKind,
				APIVersion: schema.GroupVersion{Group: monitoring.GroupName, Version: monitoringv1alpha1.Version}.String(),
			},
			Spec: spec,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(err.Error(), "failed to marshal PrometheusAgent spec")
	}

	p, err := f.MonClientV1alpha1.PrometheusAgents(ns).Patch(
		ctx,
		name,
		types.ApplyPatchType,
		b,
		metav1.PatchOptions{
			Force:        ptr.To(true),
			FieldManager: "e2e-test",
		},
	)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (f *Framework) ScalePrometheusAgentAndWaitUntilReady(ctx context.Context, name, ns string, shards int32) (*monitoringv1alpha1.PrometheusAgent, error) {
	pAgentClient := f.MonClientV1alpha1.PrometheusAgents(ns)
	scale, err := pAgentClient.GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get prometheus agent %s/%s scale: %w", ns, name, err)
	}
	scale.Spec.Replicas = shards

	_, err = pAgentClient.UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to scale prometheus agent %s/%s: %w", ns, name, err)
	}
	p, err := pAgentClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get prometheus agent %s/%s: %w", ns, name, err)
	}
	return f.WaitForPrometheusAgentReady(ctx, p, 5*time.Minute)
}

func (f *Framework) PatchPrometheusAgentAndWaitUntilReady(ctx context.Context, name, ns string, spec monitoringv1alpha1.PrometheusAgentSpec) (*monitoringv1alpha1.PrometheusAgent, error) {
	p, err := f.PatchPrometheusAgent(ctx, name, ns, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to patch prometheus agent %s/%s: %w", ns, name, err)
	}

	if ptr.Deref(p.Spec.Mode, "StatefulSet") == "DaemonSet" {
		err = f.WaitForPrometheusAgentDSReady(ctx, ns, p)
	} else {
		p, err = f.WaitForPrometheusAgentReady(ctx, p, 5*time.Minute)
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}
