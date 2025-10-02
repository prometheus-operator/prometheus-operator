// Copyright 2025 The prometheus-operator Authors
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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *Framework) WaitForBoundPVC(ctx context.Context, ns string, labelSelector string, expected int) error {
	var pollErr error
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		pvcs, err := f.KubeClient.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			pollErr = err
			return false, nil
		}

		if len(pvcs.Items) != expected {
			pollErr = fmt.Errorf("expecting %d pvcs, got %d", expected, len(pvcs.Items))
			return false, nil
		}

		for _, pvc := range pvcs.Items {
			if pvc.Status.Phase != v1.ClaimBound {
				pollErr = fmt.Errorf("expecting PVC %s to have Bound phase, got %q", pvc.Name, pvc.Status.Phase)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("%w: %w", err, pollErr)
	}

	return nil
}
