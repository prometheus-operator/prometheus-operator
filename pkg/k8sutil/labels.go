// Copyright 2021 The prometheus-operator Authors
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

package k8sutil

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelSelectionHasChanged returns true if the selector doesn't yield the same results
// for the old and current labels.
func LabelSelectionHasChanged(old, current map[string]string, selector *monitoringv1.ValidatedLabelSelector) (bool, error) {
	// If the labels haven't changed, the selector won't return different results.
	if reflect.DeepEqual(old, current) {
		return false, nil
	}

	sel, err := selector.AsSelector()
	if err != nil {
		return false, fmt.Errorf("failed to convert selector %q: %w", selector.String(), err)
	}

	// The selector doesn't restrict the selection thus old and current labels always match.
	if sel.Empty() {
		return false, nil
	}

	return sel.Matches(labels.Set(old)) != sel.Matches(labels.Set(current)), nil
}

// ValidateLabelSelector validates a LabelSelector using Kubernetes validation functions.
func ValidateLabelSelector(selector *monitoringv1.ValidatedLabelSelector) error {
	if selector == nil {
		return nil
	}

	// Validate matchLabels keys and values.
	for key, value := range selector.MatchLabels {
		if errs := validation.IsQualifiedName(key); len(errs) > 0 {
			return fmt.Errorf("key: Invalid value: %q: %v", key, errs)
		}
		if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
			return fmt.Errorf("value: Invalid value: %q: %v", value, errs)
		}
	}

	// Validate matchExpressions keys and values.
	for _, expr := range selector.MatchExpressions {
		if errs := validation.IsQualifiedName(expr.Key); len(errs) > 0 {
			return fmt.Errorf("key: Invalid value: %q: %v", expr.Key, errs)
		}

		for _, value := range expr.Values {
			if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
				return fmt.Errorf("value: Invalid value: %q: %v", value, errs)
			}
		}

		// Validate operators.
		validOperators := map[metav1.LabelSelectorOperator]bool{
			metav1.LabelSelectorOpIn:           true,
			metav1.LabelSelectorOpNotIn:        true,
			metav1.LabelSelectorOpExists:       true,
			metav1.LabelSelectorOpDoesNotExist: true,
		}
		if !validOperators[expr.Operator] {
			return fmt.Errorf("operator: Invalid value: %q: invalid operator", expr.Operator)
		}
	}

	return nil
}
