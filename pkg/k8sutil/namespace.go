// Copyright 2024 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sutil

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// GetNamespaceSelector converts a label selector into a labels.Selector that can be used for filtering namespaces.
// It returns nil for nil or empty selectors, which indicates matching only the current namespace.
func GetNamespaceSelector(nsSelector *metav1.LabelSelector) (labels.Selector, error) {
	if nsSelector == nil {
		// A nil namespace selector is equivalent to selecting only the current namespace
		return nil, nil
	}

	// An empty namespace selector (not nil, but with no requirements) means matching all namespaces
	if len(nsSelector.MatchLabels) == 0 && len(nsSelector.MatchExpressions) == 0 {
		return labels.Everything(), nil
	}

	// Convert the label selector to a labels.Selector
	selector, err := metav1.LabelSelectorAsSelector(nsSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector to labels.Selector: %w", err)
	}

	return selector, nil
}
