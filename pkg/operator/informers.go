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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

type InformerGetter interface {
	GetInformers() []informers.InformLister
}

type RefResolver interface {
	HasRefTo(namespacedName string, obj runtime.Object) bool
}

// HasReferenceFunc returns a function which takes a object (secret or
// configmap) as input parameter and returns true if at least one workload
// watched by the controller has a reference to this object or is in the same
// namespace.
//
// The object is expected to be a [*metav1.PartialObjectMetadata].
func HasReferenceFunc(
	informerGetter InformerGetter,
	refResolver RefResolver,
) FilterFunc {
	return func(ep EventPayload) bool {
		partialObjMeta, ok := ep.Current.(*metav1.PartialObjectMetadata)
		if !ok {
			return false
		}

		for _, informer := range informerGetter.GetInformers() {
			workloads, err := informer.Lister().List(labels.Everything())
			if err != nil {
				continue
			}

			for _, workload := range workloads {
				workloadMeta, ok := workload.(metav1.Object)
				if !ok {
					continue
				}

				// Check if the object is located in the same namespace as the
				// workload resource. In this case, the operator should always
				// trigger a reconciliation because the workload might use the
				// object for its configuration or an external actor may have
				// altered an object owned by the operator.
				if workloadMeta.GetNamespace() == partialObjMeta.GetNamespace() {
					return true
				}

				if refResolver.HasRefTo(fmt.Sprintf("%s/%s", workloadMeta.GetNamespace(), workloadMeta.GetName()), partialObjMeta) {
					return true
				}
			}
		}

		return false
	}
}

// GetObjectFromKey retrieves an object from the informer cache using the
// provided key. It returns a nil value and no error if the object is not
// found.
//
// The function will panic if the caller provides an informer which doesn't
// reference objects of type T.
func GetObjectFromKey[T runtime.Object](infs *informers.ForResource, key string) (T, error) {
	var zero T

	obj, err := infs.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return zero, nil
		}

		return zero, fmt.Errorf("failed to retrieve object from informer: %w", err)
	}

	obj = obj.DeepCopyObject()
	if err = k8s.AddTypeInformationToObject(obj); err != nil {
		return zero, err
	}

	return obj.(T), nil
}
