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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

const (
	// managedByOperatorLabel is the legacy label key.
	// It is redundant with ManagedByLabelKey but preserved for historical
	// reasons.
	managedByOperatorLabel = "managed-by"
)

// ManagedByOperatorLabelSelector returns a label selector which selects
// objects managed by this operator.
func ManagedByOperatorLabelSelector() string {
	return fmt.Sprintf("%s in (%s)", managedByOperatorLabel, ManagedByLabelValue)
}

type ObjectOption func(metav1.Object)

type Owner interface {
	metav1.ObjectMetaAccessor
	schema.ObjectKind
}

// WithOwner adds the given object to the list of owner references.
func WithOwner(owner Owner) ObjectOption {
	return func(o metav1.Object) {
		o.SetOwnerReferences(
			append(
				o.GetOwnerReferences(),
				metav1.OwnerReference{
					APIVersion: owner.GroupVersionKind().GroupVersion().String(),
					Kind:       owner.GroupVersionKind().Kind,
					Name:       owner.GetObjectMeta().GetName(),
					UID:        owner.GetObjectMeta().GetUID(),
				},
			),
		)
	}
}

// WithManagingOwner adds the given object as the managing object.
func WithManagingOwner(owner Owner) ObjectOption {
	return func(o metav1.Object) {
		o.SetOwnerReferences(
			append(
				o.GetOwnerReferences(),
				metav1.OwnerReference{
					APIVersion:         owner.GroupVersionKind().GroupVersion().String(),
					BlockOwnerDeletion: ptr.To(true),
					Controller:         ptr.To(true),
					Kind:               owner.GroupVersionKind().Kind,
					Name:               owner.GetObjectMeta().GetName(),
					UID:                owner.GetObjectMeta().GetUID(),
				},
			),
		)
	}
}

// WithName updates the name of the object.
func WithName(name string) ObjectOption {
	return func(o metav1.Object) {
		o.SetName(name)
	}
}

// WithNamespace updates the namespace of the object.
func WithNamespace(namespace string) ObjectOption {
	return func(o metav1.Object) {
		o.SetNamespace(namespace)
	}
}

// WithLabels merges the given labels with the existing object's labels.
// The given labels take precedence over the existing ones.
func WithLabels(labels map[string]string) ObjectOption {
	return func(o metav1.Object) {
		l := Map{}
		l = l.Merge(labels)
		l = l.Merge(o.GetLabels())

		o.SetLabels(l)
	}
}

// WithSelectorLabels merges the labels from the selector with the existing
// object's labels.
// The selector's labels take precedence over the existing ones.
func WithSelectorLabels(selector *metav1.LabelSelector) ObjectOption {
	return func(o metav1.Object) {
		if selector == nil {
			return
		}

		l := Map{}
		l = l.Merge(selector.MatchLabels)
		l = l.Merge(o.GetLabels())

		o.SetLabels(l)
	}
}

// WithAnnotations merges the given annotations with the existing object's annotations.
// The given annotations take precedence over the existing ones.
func WithAnnotations(annotations map[string]string) ObjectOption {
	return func(o metav1.Object) {
		a := Map{}
		a = a.Merge(annotations)
		a = a.Merge(o.GetAnnotations())

		o.SetAnnotations(a)
	}
}

// WithInputHashAnnotation records the given hash string in the object's
// annotations.
func WithInputHashAnnotation(h string) ObjectOption {
	return func(o metav1.Object) {
		a := o.GetAnnotations()
		if a == nil {
			a = map[string]string{}
		}
		a[InputHashAnnotationKey] = h
		o.SetAnnotations(a)
	}
}

// WithoutKubectlAnnotations removes kubectl annotations inherited from the
// governing object. Otherwise the managed object might be deleted when
// "kubectl apply --prune" is run against the governing object.
func WithoutKubectlAnnotations() ObjectOption {
	return func(o metav1.Object) {
		a := make(map[string]string, len(o.GetAnnotations()))
		for k, v := range o.GetAnnotations() {
			if !strings.HasPrefix(k, "kubectl.kubernetes.io/") {
				a[k] = v
			}
		}

		o.SetAnnotations(a)
	}
}

// UpdateObject updates the object's metadata with the provided options.
// It automatically injects the "managed-by" label which identifies the
// operator as the managing entity.
func UpdateObject(o metav1.Object, opts ...ObjectOption) {
	WithLabels(map[string]string{
		managedByOperatorLabel: ManagedByLabelValue,
	})(o)

	for _, opt := range opts {
		opt(o)
	}
}
