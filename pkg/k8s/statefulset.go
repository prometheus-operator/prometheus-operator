// Copyright 2026 The prometheus-operator Authors
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

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
)

// CreateStatefulSetOrPatchLabels creates a StatefulSet resource.
// If the StatefulSet already exists, it patches the labels from the input StatefulSet.
func CreateStatefulSetOrPatchLabels(ctx context.Context, ssetClient clientappsv1.StatefulSetInterface, sset *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	created, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{})
	if err == nil {
		return created, nil
	}

	if !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	// StatefulSet already exists, patch the labels
	patchData, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"labels": sset.Labels,
		},
	})
	if err != nil {
		return nil, err
	}

	return ssetClient.Patch(
		ctx,
		sset.Name,
		types.StrategicMergePatchType,
		patchData,
		metav1.PatchOptions{FieldManager: PrometheusOperatorFieldManager},
	)
}

// ForceUpdateStatefulSet updates a StatefulSet resource preserving custom
// labels and annotations. But when the update operation tries to update
// immutable fields for example, `.spec.selector`), the function will delete
// the statefulset (relying on the higher-level controller to re-create the
// resource during the next reconciliation).
//
// It calls onDeleteFunc when the deletion of the resource is required. The
// function is given a string explaining the reason why the update was not
// possible.
func ForceUpdateStatefulSet(ctx context.Context, ssetClient clientappsv1.StatefulSetInterface, sset *appsv1.StatefulSet, onDeleteFunc func(string)) error {
	err := updateStatefulSet(ctx, ssetClient, sset)
	if err == nil {
		return err
	}

	// When trying to update immutable fields, the API server returns a 422 status code.
	sErr, ok := err.(*apierrors.StatusError)
	if !ok || (sErr.ErrStatus.Code != 422 || sErr.ErrStatus.Reason != metav1.StatusReasonInvalid) {
		return fmt.Errorf("failed to update StatefulSet: %w", err)
	}

	// Gather the reason(s) why the update failed.
	failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
	for i, cause := range sErr.ErrStatus.Details.Causes {
		failMsg[i] = cause.Message
	}
	if onDeleteFunc != nil {
		onDeleteFunc(strings.Join(failMsg, ", "))
	}

	return ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationForeground)})
}

// updateStatefulSet updates a StatefulSet resource preserving custom labels and annotations from the current resource.
func updateStatefulSet(ctx context.Context, sstClient clientappsv1.StatefulSetInterface, sset *appsv1.StatefulSet) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existingSset, err := sstClient.Get(ctx, sset.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		mergeMetadata(&sset.ObjectMeta, existingSset.ObjectMeta)
		// Propagate annotations set by kubectl on spec.template.annotations. e.g performing a rolling restart.
		mergeKubectlAnnotations(&existingSset.Spec.Template.ObjectMeta, sset.Spec.Template.ObjectMeta)

		_, err = sstClient.Update(ctx, sset, metav1.UpdateOptions{})
		return err
	})
}
