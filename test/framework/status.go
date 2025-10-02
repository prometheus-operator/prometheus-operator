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
)

type resourceStatus struct {
	expectedReplicas int32
	generation       int64
	replicas         int32
	conditions       []monitoringv1.Condition
}

func (f *Framework) AssertCondition(conds []monitoringv1.Condition, expectedType monitoringv1.ConditionType, expectedStatus monitoringv1.ConditionStatus) error {
	for _, c := range conds {
		if c.Type != expectedType {
			continue
		}

		if c.Status != expectedStatus {
			return fmt.Errorf("expected condition %q to be %q but got %q", c.Type, expectedStatus, c.Status)
		}

		return nil
	}

	return fmt.Errorf("condition %q not found", expectedType)
}

// WaitForResourceAvailable waits for a monitoring resource to report itself as being reconciled & available.
// If the resource isn't available within the given timeout, it returns an error.
func (f *Framework) WaitForResourceAvailable(ctx context.Context, getResourceStatus func(context.Context) (resourceStatus, error), timeout time.Duration) error {
	var pollErr error
	if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		var status resourceStatus
		status, pollErr = getResourceStatus(ctx)
		if pollErr != nil {
			return false, nil
		}

		if status.replicas != status.expectedReplicas {
			pollErr = fmt.Errorf("expected %d replicas, got %d", status.expectedReplicas, status.replicas)
			return false, nil
		}

		var reconciled, available *monitoringv1.Condition
		for _, cond := range status.conditions {
			// We need to create a new address for 'cond' inside the loop, otherwise we change the value of
			// 'available' and 'reconciled' will have their pointer changed on every loop iteration.
			// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
			if cond.Type == monitoringv1.Available {
				available = &cond
			}
			if cond.Type == monitoringv1.Reconciled {
				reconciled = &cond
			}
			if cond.ObservedGeneration != status.generation {
				pollErr = fmt.Errorf("observed generation %d for condition %s isn't equal to the state generation %d",
					cond.ObservedGeneration,
					cond.Type,
					status.generation)
				return false, nil
			}
		}

		if reconciled == nil {
			pollErr = fmt.Errorf("failed to find Reconciled condition in status subresource")
			return false, nil
		}

		if reconciled.Status != monitoringv1.ConditionTrue {
			pollErr = fmt.Errorf(
				"expected Reconciled condition to be 'True', got %q (reason %s, %q)",
				reconciled.Status,
				reconciled.Reason,
				reconciled.Message,
			)
			return false, nil
		}

		if available == nil {
			pollErr = fmt.Errorf("failed to find Available condition in status subresource")
			return false, nil
		}

		if available.Status != monitoringv1.ConditionTrue {
			pollErr = fmt.Errorf(
				"expected Available condition to be 'True', got %q (reason %s, %q)",
				available.Status,
				available.Reason,
				available.Message,
			)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("%v: %w", pollErr, err)
	}

	return nil
}

// GetWorkloadBinding returns the binding matching the workload.
func (f *Framework) GetWorkloadBinding(bindings []monitoringv1.WorkloadBinding, workload metav1.Object, resource string) (monitoringv1.WorkloadBinding, error) {
	for _, binding := range bindings {
		if binding.Resource == resource && binding.Name == workload.GetName() && binding.Namespace == workload.GetNamespace() {
			return binding, nil
		}
	}

	return monitoringv1.WorkloadBinding{}, fmt.Errorf("binding not found")
}

// GetConfigResourceCondition returns the condition type.
func (f *Framework) GetConfigResourceCondition(conditions []monitoringv1.ConfigResourceCondition, conditionType monitoringv1.ConditionType) (monitoringv1.ConfigResourceCondition, error) {
	for _, cond := range conditions {
		if cond.Type == conditionType {
			return cond, nil
		}
	}

	return monitoringv1.ConfigResourceCondition{}, fmt.Errorf("condition %q not found", conditionType)
}

// WaitForConfigResourceCondition waits for a configuration resource
// (serviceMonitor, podMonitor, scrapeConfig and probes) to meet the expected
// condition.
// If the condition isn't met within the given timeout, it returns an error.
func (f *Framework) WaitForConfigResourceCondition(ctx context.Context, getConfigResourceStatus func(context.Context) ([]monitoringv1.WorkloadBinding, error), workload metav1.Object, resource string, conditionType monitoringv1.ConditionType, conditionStatus monitoringv1.ConditionStatus, timeout time.Duration) error {
	var pollErr error
	if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		var bindings []monitoringv1.WorkloadBinding
		bindings, pollErr = getConfigResourceStatus(ctx)
		if pollErr != nil {
			return false, nil
		}

		var binding monitoringv1.WorkloadBinding
		binding, pollErr = f.GetWorkloadBinding(bindings, workload, resource)
		if pollErr != nil {
			return false, nil
		}

		var cond monitoringv1.ConfigResourceCondition
		cond, pollErr = f.GetConfigResourceCondition(binding.Conditions, conditionType)
		if pollErr != nil {
			return false, nil
		}

		if cond.Status != conditionStatus {
			pollErr = fmt.Errorf(
				"expected condition %q with status %q, got %q",
				conditionType, conditionStatus, cond.Status)
			return false, nil
		}

		return true, nil
	}); err != nil {
		return fmt.Errorf("%v: resource %q %s/%s: %w", err, resource, workload.GetNamespace(), workload.GetName(), pollErr)
	}

	return nil
}

// WaitForConfigResWorkloadBindingCleanup waits for a configuration resource (serviceMonitor, podMonitor, scrapeConfig and probes) to remove the expected workload binding from the status.
// If the binding isn't removed within the given timeout, it returns an error.
func (f *Framework) WaitForConfigResWorkloadBindingCleanup(ctx context.Context, getConfigResourceStatus func(context.Context) ([]monitoringv1.WorkloadBinding, error), workload metav1.Object, resource string, timeout time.Duration) error {
	var pollErr error
	if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		var bindings []monitoringv1.WorkloadBinding
		bindings, pollErr = getConfigResourceStatus(ctx)
		if pollErr != nil {
			return false, nil
		}

		for _, binding := range bindings {
			if binding.Resource == resource && binding.Name == workload.GetName() && binding.Namespace == workload.GetNamespace() {

				pollErr = fmt.Errorf("binding for resource %q with name %q in namespace %q still exists", resource, workload.GetName(), workload.GetNamespace())
				return false, nil
			}
		}

		return true, nil
	}); err != nil {
		return fmt.Errorf("%v: %w", err, pollErr)
	}

	return nil
}
