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
	"reflect"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// LabelSelectionHasChanged returns true if the selector doesn't yield the same results
// for the old and current labels.
func LabelSelectionHasChanged(old, current map[string]string, selector *metav1.LabelSelector) (bool, error) {
	// If the labels haven't changed, the selector won't return different results.
	if reflect.DeepEqual(old, current) {
		return false, nil
	}

	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false, errors.Wrapf(err, "failed to convert selector %q", selector.String())
	}

	// The selector doesn't restrict the selection thus old and current labels always match.
	if sel.Empty() {
		return false, nil
	}

	return sel.Matches(labels.Set(old)) != sel.Matches(labels.Set(current)), nil
}
