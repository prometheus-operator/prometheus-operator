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

package operator

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func CheckStorageClass(ctx context.Context, canReadStorageClass bool, kclient kubernetes.Interface, storage *monitoringv1.StorageSpec) error {
	// Check the existence of the storage class if not empty and if the operator has enough permissions.
	if !canReadStorageClass || storage == nil {
		return nil
	}

	storageClassName := StringPtrValOrDefault(storage.VolumeClaimTemplate.Spec.StorageClassName, "")
	if storageClassName == "" {
		return nil
	}

	_, err := kclient.StorageV1().StorageClasses().Get(ctx, storageClassName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("storage class %q does not exist", storageClassName)
		}
		return fmt.Errorf("cannot get %q storageclass: %w", storageClassName, err)
	}

	return nil
}
