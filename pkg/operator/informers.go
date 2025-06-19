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
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// GetObjectFromKey retrieves an object from the informer cache using the provided key.
// Returns nil,nil if the object is not found, or an error if there is a problem retrieving it.
func GetObjectFromKey[T runtime.Object](infs *informers.ForResource, key string) (T, error) {
	obj, err := infs.Get(key)
	var zero T

	if err != nil {
		if apierrors.IsNotFound(err) {
			return zero, nil
		}
		return zero, fmt.Errorf("failed to retrieve object from informer: %w", err)
	}

	copy, ok := obj.DeepCopyObject().(T)
	if !ok {
		return zero, fmt.Errorf("object %T is not of type %T", copy, zero)
	}

	if err = k8sutil.AddTypeInformationToObject(obj); err != nil {
		return zero, err
	}
	return copy, nil
}
