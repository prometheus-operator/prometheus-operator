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
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/listwatch"

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
)

type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

// InformerFactoriesForNamespaces is a simple way to combine several shared informers into a single struct with unified listing power
type InformerFactoriesForNamespaces interface {
	ForResource(namespace string, resource schema.GroupVersionResource) (GenericInformer, error)
	Namespaces() sets.String
}

type InformersForResource struct {
	informers []GenericInformer
}

func NewInformersForResource(ifs InformerFactoriesForNamespaces, resource schema.GroupVersionResource) (*InformersForResource, error) {
	namespaces := ifs.Namespaces().List()
	var informers []GenericInformer

	for _, ns := range namespaces {
		informer, err := ifs.ForResource(ns, resource)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting informer for resource %v", resource)
		}
		informers = append(informers, informer)
	}

	return &InformersForResource{
		informers: informers,
	}, nil
}

func (w *InformersForResource) Start(stopCh <-chan struct{}) {
	for _, i := range w.informers {
		go i.Informer().Run(stopCh)
	}
}

func (w *InformersForResource) GetInformers() []GenericInformer {
	return w.informers
}

func (w *InformersForResource) AddEventHandler(handler cache.ResourceEventHandler) {
	for _, i := range w.informers {
		i.Informer().AddEventHandler(handler)
	}
}

func (w *InformersForResource) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	for _, i := range w.informers {
		i.Informer().AddEventHandlerWithResyncPeriod(handler, resyncPeriod)
	}
}

func (w *InformersForResource) HasSynced() bool {
	for _, i := range w.informers {
		if !i.Informer().HasSynced() {
			return false
		}
	}

	return true
}

func (w *InformersForResource) List(selector labels.Selector) ([]runtime.Object, error) {
	var ret []runtime.Object

	for _, inf := range w.informers {
		objs, err := inf.Lister().List(selector)
		if err != nil {
			return nil, err
		}
		ret = append(ret, objs...)
	}

	return ret, nil
}

func (w *InformersForResource) ListAllByNamespace(namespace string, selector labels.Selector, appendFn cache.AppendFunc) error {
	for _, inf := range w.informers {
		err := cache.ListAllByNamespace(inf.Informer().GetIndexer(), namespace, selector, appendFn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *InformersForResource) Get(name string) (runtime.Object, error) {
	var err error

	for _, inf := range w.informers {
		var ret runtime.Object
		ret, err = inf.Lister().Get(name)

		if apierrors.IsNotFound(err) {
			continue
		}

		if err != nil {
			return nil, err
		}

		return ret, nil
	}

	return nil, err
}

func newInformerOptions(allowedNamespaces, deniedNamespaces map[string]struct{}, tweaks func(*v1.ListOptions)) (func(*v1.ListOptions), []string) {
	if tweaks == nil {
		tweaks = func(*v1.ListOptions) {} // nop
	}

	var namespaces []string

	if listwatch.IsAllNamespaces(allowedNamespaces) {
		namespaces = append(namespaces, v1.NamespaceAll)

		return func(options *v1.ListOptions) {
			tweaks(options)
			denyListTweak(options, deniedNamespaces)
		}, namespaces
	}

	for ns, _ := range allowedNamespaces {
		namespaces = append(namespaces, ns)
	}

	return tweaks, namespaces
}

func denyListTweak(options *metav1.ListOptions, namespaces map[string]struct{}) {
	if len(namespaces) == 0 {
		return
	}

	var denied []string

	for ns, _ := range namespaces {
		denied = append(denied, "metadata.namespace!="+ns)
	}

	options.FieldSelector = strings.Join(denied, ",")
}
