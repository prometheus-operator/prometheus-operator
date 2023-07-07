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
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/prometheus/agent"
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

func (f *Framework) CreatePrometheusAgentAndWaitUntilReady(ctx context.Context, ns string, p *monitoringv1alpha1.PrometheusAgent) (*monitoringv1alpha1.PrometheusAgent, error) {
	result, err := f.MonClientV1alpha1.PrometheusAgents(ns).Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v prometheus-agent instances failed (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	if err := f.WaitForPrometheusAgentReady(ctx, result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("waiting for %v prometheus-agent instances timed out (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	return result, nil
}

func (f *Framework) WaitForPrometheusAgentReady(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, timeout time.Duration) error {
	expected := *p.Spec.Replicas
	if p.Spec.Shards != nil && *p.Spec.Shards > 0 {
		expected = expected * *p.Spec.Shards
	}

	if err := f.WaitForResourceAvailable(
		ctx,
		func(context.Context) (resourceStatus, error) {
			current, err := f.MonClientV1alpha1.PrometheusAgents(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
			if err != nil {
				return resourceStatus{}, err
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
		return errors.Wrapf(err, "prometheus-agent %v/%v failed to become available", p.Namespace, p.Name)
	}

	return nil
}

func (f *Framework) DeletePrometheusAgentAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1alpha1.PrometheusAgents(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting PrometheusAgent custom resource %v failed", name))
	}

	if err := f.MonClientV1alpha1.PrometheusAgents(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting PrometheusAgent custom resource %v failed", name))
	}

	if err := f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		0,
		prometheusagent.ListOptions(name),
	); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("waiting for PrometheusAgent custom resource (%s) to vanish timed out", name),
		)
	}

	return nil
}
