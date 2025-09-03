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
)

func (f *Framework) WaitForServiceMonitorCondition(ctx context.Context, sm *monitoringv1.ServiceMonitor, workload metav1.Object, resource string, conditionType monitoringv1.ConditionType, conditionStatus monitoringv1.ConditionStatus, timeout time.Duration) (*monitoringv1.ServiceMonitor, error) {
	var current *monitoringv1.ServiceMonitor

	if err := f.WaitForConfigResourceCondition(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1.ServiceMonitors(sm.Namespace).Get(ctx, sm.Name, metav1.GetOptions{})
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
		return nil, fmt.Errorf("serviceMonitor status %v/%v failed to reach expected condition: %w", sm.Namespace, sm.Name, err)
	}
	return current, nil
}

func (f *Framework) WaitForServiceMonitorWorkloadBindingCleanup(ctx context.Context, sm *monitoringv1.ServiceMonitor, workload metav1.Object, resource string, timeout time.Duration) (*monitoringv1.ServiceMonitor, error) {
	var current *monitoringv1.ServiceMonitor

	if err := f.WaitForConfigResWorkloadBindingCleanup(
		ctx,
		func(ctx context.Context) ([]monitoringv1.WorkloadBinding, error) {
			var err error
			current, err = f.MonClientV1.ServiceMonitors(sm.Namespace).Get(ctx, sm.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return current.Status.Bindings, nil
		},
		workload,
		resource,
		timeout,
	); err != nil {
		return nil, fmt.Errorf("serviceMonitor status %v/%v failed to reach expected condition: %w", sm.Namespace, sm.Name, err)
	}
	return current, nil
}
