// Copyright 2020 The prometheus-operator Authors
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

package informers

import (
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
)

// InformLister is the interface that both exposes a shared index informer
// and a generic lister.
// Usually generated clients declare this interface as "GenericInformer".
type InformLister interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

// FactoriesForNamespaces is a way to combine several shared informers into a single struct with unified listing power.
type FactoriesForNamespaces interface {
	ForResource(namespace string, resource schema.GroupVersionResource) (InformLister, error)
	Namespaces() sets.Set[string]
}

// ForResource contains a slice of InformLister for a concrete resource type,
// one per namespace.
type ForResource struct {
	gr        schema.GroupResource
	informers []InformLister
}

// NewInformersForResource returns a composite informer exposing the most basic set of operations
// needed from informers and listers. It does not implement a formal interface,
// but exposes a minimal set of methods from underlying slice of cache.SharedIndexInformers and cache.GenericListers.
//
// It takes a namespace aware informer factory, wrapped in a FactoriesForNamespaces interface
// that is able to instantiate an informer for a given namespace.
func NewInformersForResource(ifs FactoriesForNamespaces, resource schema.GroupVersionResource) (*ForResource, error) {
	return NewInformersForResourceWithTransform(ifs, resource, nil)
}

func NewInformersForResourceWithTransform(ifs FactoriesForNamespaces, resource schema.GroupVersionResource, handler cache.TransformFunc) (*ForResource, error) {
	namespaces := ifs.Namespaces().UnsortedList()
	slices.Sort(namespaces)

	informers := make([]InformLister, 0, len(namespaces))

	for _, ns := range namespaces {
		informer, err := ifs.ForResource(ns, resource)
		if err != nil {
			return nil, fmt.Errorf("error getting informer in namespace %q for resource %v: %w", ns, resource, err)
		}
		if handler != nil {
			if err := informer.Informer().SetTransform(handler); err != nil {
				return nil, fmt.Errorf("error setting transform in namespace %q for resource %v: %w", ns, resource, err)
			}
		}
		informers = append(informers, informer)
	}

	return &ForResource{
		gr:        resource.GroupResource(),
		informers: informers,
	}, nil
}

func partialObjectMetadataStrip(obj any) (*v1.PartialObjectMetadata, error) {
	partialMeta, ok := obj.(*v1.PartialObjectMetadata)
	if !ok {
		// Don't do anything if the cast isn't successful.
		// The object might be of type "cache.DeletedFinalStateUnknown".
		return nil, fmt.Errorf("invalid object type: %T", obj)
	}

	partialMeta.Annotations = nil
	partialMeta.Labels = nil
	partialMeta.ManagedFields = nil
	partialMeta.Finalizers = nil
	partialMeta.OwnerReferences = nil

	return partialMeta, nil
}

// PartialObjectMetadataStrip removes the following fields from PartialObjectMetadata objects:
// * Annotations
// * Labels
// * ManagedFields
// * Finalizers
// * OwnerReferences.
//
// It also sets the TypeMeta field on the PartialObjectMetadata objects so
// consumers can introspect the object's type.
//
// If the passed object isn't of type *v1.PartialObjectMetadata, it is returned unmodified.
//
// It matches the cache.TransformFunc type and can be used by informers
// watching PartialObjectMetadata objects to reduce memory consumption.
// See https://pkg.go.dev/k8s.io/client-go@v0.29.1/tools/cache#TransformFunc for details.
func PartialObjectMetadataStrip(gvk schema.GroupVersionKind) cache.TransformFunc {
	return func(obj any) (any, error) {
		partialMeta, err := partialObjectMetadataStrip(obj)
		if err != nil {
			return obj, nil
		}

		partialMeta.TypeMeta = v1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		}

		return partialMeta, nil
	}
}

// Start starts all underlying informers, passing the given stop channel to each of them.
func (w *ForResource) Start(stopCh <-chan struct{}) {
	for _, i := range w.informers {
		go i.Informer().Run(stopCh)
	}
}

// GetInformers returns all wrapped informers.
func (w *ForResource) GetInformers() []InformLister {
	return w.informers
}

// AddEventHandler registers the given handler to all wrapped informers.
func (w *ForResource) AddEventHandler(handler cache.ResourceEventHandler) {
	for _, i := range w.informers {
		_, _ = i.Informer().AddEventHandler(handler)
	}
}

// HasSynced returns true if all underlying informers have synced, else false.
func (w *ForResource) HasSynced() bool {
	for _, i := range w.informers {
		if !i.Informer().HasSynced() {
			return false
		}
	}

	return true
}

// ListAll invokes the ListAll method for all wrapped informers passing the
// same selector and appendFn.
func (w *ForResource) ListAll(selector labels.Selector, appendFn cache.AppendFunc) error {
	for _, inf := range w.informers {
		err := cache.ListAll(inf.Informer().GetIndexer(), selector, appendFn)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListAllByNamespace invokes all wrapped informers passing the same appendFn.
// While wrapped informers are usually namespace aware, it is still important to iterate over all of them
// as some informers might wrap k8s.io/apimachinery/pkg/apis/meta/v1.NamespaceAll.
func (w *ForResource) ListAllByNamespace(namespace string, selector labels.Selector, appendFn cache.AppendFunc) error {
	for _, inf := range w.informers {
		err := cache.ListAllByNamespace(inf.Informer().GetIndexer(), namespace, selector, appendFn)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get invokes all wrapped informers and returns the first found runtime object.
// It returns a NotFound error if the object isn't found in any informer.
func (w *ForResource) Get(name string) (runtime.Object, error) {
	for _, inf := range w.informers {
		ret, err := inf.Lister().Get(name)
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		return ret, nil
	}

	return nil, apierrors.NewNotFound(w.gr, name)
}

// newInformerOptions returns a list option tweak function and a list of namespaces
// based on the given allowed and denied namespaces.
//
// If allowedNamespaces contains one only entry equal to k8s.io/apimachinery/pkg/apis/meta/v1.NamespaceAll
// then it returns it and a tweak function filtering denied namespaces using a field selector.
//
// Else, denied namespaces are ignored and just the set of allowed namespaces is returned.
func newInformerOptions(allowedNamespaces, deniedNamespaces map[string]struct{}, tweaks func(*v1.ListOptions)) (func(*v1.ListOptions), []string) {
	if tweaks == nil {
		tweaks = func(*v1.ListOptions) {} // nop
	}

	var namespaces []string

	if listwatch.IsAllNamespaces(allowedNamespaces) {
		return func(options *v1.ListOptions) {
			tweaks(options)
			listwatch.DenyTweak(options, "metadata.namespace", deniedNamespaces)
		}, []string{v1.NamespaceAll}
	}

	for ns := range allowedNamespaces {
		namespaces = append(namespaces, ns)
	}

	return tweaks, namespaces
}
