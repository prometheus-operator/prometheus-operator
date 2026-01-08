// Copyright 2016 The prometheus-operator Authors
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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func (f *Framework) WaitForAlertmanagerConfigCondition(ctx context.Context, alc *monitoringv1alpha1.AlertmanagerConfig, workload metav1.Object, resource string, conditionType monitoringv1.ConditionType, conditionStatus monitoringv1.ConditionStatus, timeout time.Duration) (*monitoringv1alpha1.AlertmanagerConfig, error) {
	var current *monitoringv1alpha1.AlertmanagerConfig

	if err := f.WaitForConfigResourceCondition(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1alpha1.AlertmanagerConfigs(alc.Namespace).Get(ctx, alc.Name, metav1.GetOptions{})
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
		return nil, fmt.Errorf("alertmanagerConfig status %v/%v failed to reach expected condition: %w", alc.Namespace, alc.Name, err)
	}
	return current, nil
}

func (f *Framework) WaitForAlertmanagerConfigWorkloadBindingCleanup(ctx context.Context, alc *monitoringv1alpha1.AlertmanagerConfig, workload metav1.Object, resource string, timeout time.Duration) (*monitoringv1alpha1.AlertmanagerConfig, error) {
	var current *monitoringv1alpha1.AlertmanagerConfig

	if err := f.WaitForConfigResWorkloadBindingCleanup(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1alpha1.AlertmanagerConfigs(alc.Namespace).Get(ctx, alc.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return current.Status.Bindings, nil
		},
		workload,
		resource,
		timeout,
	); err != nil {
		return nil, fmt.Errorf("alertmanagerConfig status %v/%v failed to reach expected condition: %w", alc.Namespace, alc.Name, err)
	}
	return current, nil
}
