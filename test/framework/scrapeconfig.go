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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func (f *Framework) MakeBasicScrapeConfig(ns, name string) *monitoringv1alpha1.ScrapeConfig {
	return &monitoringv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"role": "scrapeconfig",
			},
		},
		Spec: monitoringv1alpha1.ScrapeConfigSpec{},
	}
}

func (f *Framework) CreateScrapeConfig(ctx context.Context, ns string, ar *monitoringv1alpha1.ScrapeConfig) (*monitoringv1alpha1.ScrapeConfig, error) {
	var (
		scrapeConfig *monitoringv1alpha1.ScrapeConfig
		err          error
	)

	err = wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		scrapeConfig, err = f.MonClientV1alpha1.ScrapeConfigs(ns).Create(ctx, ar, metav1.CreateOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	})

	return scrapeConfig, err
}

func (f *Framework) GetScrapeConfig(ctx context.Context, ns, name string) (*monitoringv1alpha1.ScrapeConfig, error) {
	result, err := f.MonClientV1alpha1.ScrapeConfigs(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting %v ScrapeConfig failed: %v", name, err)
	}

	return result, nil
}

func (f *Framework) UpdateScrapeConfig(ctx context.Context, ns string, ar *monitoringv1alpha1.ScrapeConfig) (*monitoringv1alpha1.ScrapeConfig, error) {
	var (
		scrapeConfig *monitoringv1alpha1.ScrapeConfig
		err          error
	)

	err = wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		scrapeConfig, err = f.MonClientV1alpha1.ScrapeConfigs(ns).Update(ctx, ar, metav1.UpdateOptions{})
		if err != nil {
			return false, fmt.Errorf("updating %v ScrapeConfig failed: %v", ar.Name, err)
		}
		return true, nil
	})

	return scrapeConfig, err
}

func (f *Framework) DeleteScrapeConfig(ctx context.Context, ns string, r string) error {
	err := f.MonClientV1alpha1.ScrapeConfigs(ns).Delete(ctx, r, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting %v ScrapeConfig rule in namespace %v failed: %v", r, ns, err.Error())
	}

	return nil
}

func (f *Framework) WaitForScrapeConfigCondition(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig, workload metav1.Object, resource string, conditionType monitoringv1.ConditionType, conditionStatus monitoringv1.ConditionStatus, timeout time.Duration) (*monitoringv1alpha1.ScrapeConfig, error) {
	var current *monitoringv1alpha1.ScrapeConfig

	if err := f.WaitForConfigResourceCondition(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1alpha1.ScrapeConfigs(sc.Namespace).Get(ctx, sc.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return current.Status.Bindings, nil
		},
		workload,
		resource,
		conditionType,
		conditionStatus,
		timeout,
	); err != nil {
		return nil, fmt.Errorf("scrapeConfig status %v/%v failed to reach expected condition: %w", sc.Namespace, sc.Name, err)
	}
	return current, nil
}

func (f *Framework) WaitForScrapeConfigWorkloadBindingCleanup(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig, workload metav1.Object, resource string, timeout time.Duration) (*monitoringv1alpha1.ScrapeConfig, error) {
	var current *monitoringv1alpha1.ScrapeConfig

	if err := f.WaitForConfigResWorkloadBindingCleanup(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1alpha1.ScrapeConfigs(sc.Namespace).Get(ctx, sc.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return current.Status.Bindings, nil
		},
		workload,
		resource,
		timeout,
	); err != nil {
		return nil, fmt.Errorf("scrapeConfig status %v/%v failed to reach expected condition: %w", sc.Namespace, sc.Name, err)
	}
	return current, nil
}
