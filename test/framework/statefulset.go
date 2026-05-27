// Copyright The prometheus-operator Authors
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
	"maps"
	"slices"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// RemoveAllLabelsFromStatefulSet removes all labels from a StatefulSet using JSON Patch.
func (f *Framework) RemoveAllLabelsFromStatefulSet(ctx context.Context, name, namespace string) error {
	sts, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get StatefulSet: %w", err)
	}

	if len(sts.Labels) == 0 {
		return nil
	}

	b, err := removeLabelsPatch(slices.Sorted(maps.Keys(sts.GetLabels()))...)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	updatedSts, err := f.KubeClient.AppsV1().StatefulSets(namespace).Patch(ctx, name, types.JSONPatchType, b, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch StatefulSet: %w", err)
	}

	if len(updatedSts.Labels) != 0 {
		return fmt.Errorf("expected all labels to be removed from StatefulSet, but got %d labels: %v", len(updatedSts.Labels), updatedSts.Labels)
	}

	return nil
}

// WaitForStatefulSetReplicas polls until the number of StatefulSets in the
// given namespace matching listOpts equals expected.
func (f *Framework) WaitForStatefulSetReplicas(ctx context.Context, ns string, listOpts metav1.ListOptions, expected int) error {
	var loopErr error
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		sts, listErr := f.KubeClient.AppsV1().StatefulSets(ns).List(ctx, listOpts)
		if listErr != nil {
			loopErr = fmt.Errorf("failed to list StatefulSets in namespace %s: %w", ns, listErr)
			return false, nil
		}
		if len(sts.Items) != expected {
			loopErr = fmt.Errorf("expected %d StatefulSets in namespace %s, got %d", expected, ns, len(sts.Items))
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("%v: %w", loopErr, err)
	}
	return nil
}
