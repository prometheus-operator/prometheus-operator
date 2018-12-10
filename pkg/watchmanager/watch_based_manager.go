/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// TODO: We did some scalability tests and using watchBasedManager
// seems to help with apiserver performance at scale visibly.
// No issues we also observed at the scale of ~200k watchers with a
// single apiserver.
// However, we need to perform more extensive testing before we
// enable this in production setups.

package watchmanager

import (
	"fmt"
	"sync"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/client-go/tools/cache"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
)

// GetObjectFunc defines a function to get object with a given namespace and name.
type GetObjectFunc func(string, string, metav1.GetOptions) (runtime.Object, error)

type objectKey struct {
	namespace string
	name      string
}

type listObjectFunc func(string, metav1.ListOptions) (runtime.Object, error)
type watchObjectFunc func(string, metav1.ListOptions) (watch.Interface, error)
type newObjectFunc func() runtime.Object

// objectCacheItem is a single item stored in objectCache.
type objectCacheItem struct {
	refCount  int
	store     cache.Store
	hasSynced func() (bool, error)
	stopCh    chan struct{}
}

// objectCache is a local cache of objects propagated via
// individual watches.
type objectCache struct {
	listObject    listObjectFunc
	watchObject   watchObjectFunc
	newObject     newObjectFunc
	groupResource schema.GroupResource

	lock            sync.Mutex
	items           map[objectKey]*objectCacheItem
	registeredSMons map[objectKey]*monitoringv1.ServiceMonitor

	getReferencedObjects func(*monitoringv1.ServiceMonitor) sets.String
	resourceEventHandler cache.ResourceEventHandler
}

// GetReferencedObjectsFunc returns names  resources given SMon is referencing
type GetReferencedObjectsFunc func(m *monitoringv1.ServiceMonitor) sets.String

// NewObjectCache creates a manager that keeps a cache of all objects
// necessary for registered pods.
// It implements the following logic:
// - whenever a pod is created or updated, we start individual watches for all
//   referenced objects that aren't referenced from other registered pods
// - every GetObject() returns a value from local cache propagated via watches
func NewObjectCache(listObject listObjectFunc, watchObject watchObjectFunc, newObject newObjectFunc, groupResource schema.GroupResource, h cache.ResourceEventHandler, getReferencedObjects GetReferencedObjectsFunc) Manager {
	return &objectCache{
		listObject:           listObject,
		watchObject:          watchObject,
		newObject:            newObject,
		groupResource:        groupResource,
		items:                make(map[objectKey]*objectCacheItem),
		registeredSMons:      make(map[objectKey]*monitoringv1.ServiceMonitor),
		getReferencedObjects: getReferencedObjects,
		resourceEventHandler: h,
	}
}

func (c *objectCache) newInformer(namespace, name string) *objectCacheItem {
	fieldSelector := fields.Set{"metadata.name": name}.AsSelector().String()
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		options.FieldSelector = fieldSelector
		return c.listObject(namespace, options)
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		options.FieldSelector = fieldSelector
		return c.watchObject(namespace, options)
	}
	store, informer := cache.NewInformer(
		&cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc},
		c.newObject(), 0, c.resourceEventHandler)
	stopCh := make(chan struct{})
	go informer.Run(stopCh)
	return &objectCacheItem{
		refCount:  0,
		store:     store,
		hasSynced: func() (bool, error) { return informer.HasSynced(), nil },
		stopCh:    stopCh,
	}
}

func (c *objectCache) addReference(namespace, name string) {
	key := objectKey{namespace: namespace, name: name}

	item, exists := c.items[key]
	if !exists {
		item = c.newInformer(namespace, name)
		c.items[key] = item
	}
	item.refCount++
}

func (c *objectCache) deleteReference(namespace, name string) {
	key := objectKey{namespace: namespace, name: name}

	if item, ok := c.items[key]; ok {
		item.refCount--
		if item.refCount == 0 {
			// Stop the underlying reflector.
			close(item.stopCh)
			delete(c.items, key)
		}
	}
}

func (c *objectCache) GetObject(namespace, name string) (runtime.Object, error) {
	return c.Get(namespace, name)
}

func (c *objectCache) RegisterSMon(sMon *monitoringv1.ServiceMonitor) {
	names := c.getReferencedObjects(sMon)
	c.lock.Lock()
	defer c.lock.Unlock()

	for name := range names {
		c.addReference(sMon.Namespace, name)
	}
	key := objectKey{namespace: sMon.Namespace, name: sMon.Name}
	prev := c.registeredSMons[key]
	c.registeredSMons[key] = sMon
	if prev != nil {
		for name := range c.getReferencedObjects(prev) {
			// On an update, the .Add() call above will have re-incremented the
			// ref count of any existing object, so any objects that are in both
			// names and prev need to have their ref counts decremented. Any that
			// are only in prev need to be completely removed. This unconditional
			// call takes care of both cases.
			c.deleteReference(prev.Namespace, name)
		}
	}
}

func (c *objectCache) UnregisterSMon(sMon *monitoringv1.ServiceMonitor) {
	key := objectKey{namespace: sMon.Namespace, name: sMon.Name}
	c.lock.Lock()
	defer c.lock.Unlock()
	prev := c.registeredSMons[key]
	delete(c.registeredSMons, key)
	if prev != nil {
		for name := range c.getReferencedObjects(prev) {
			c.deleteReference(prev.Namespace, name)
		}
	}
}

// key returns key of an object with a given name and namespace.
// This has to be in-sync with cache.MetaNamespaceKeyFunc.
func (c *objectCache) key(namespace, name string) string {
	if len(namespace) > 0 {
		return namespace + "/" + name
	}
	return name
}

func (c *objectCache) Get(namespace, name string) (runtime.Object, error) {
	key := objectKey{namespace: namespace, name: name}

	c.lock.Lock()
	item, exists := c.items[key]
	c.lock.Unlock()

	if !exists {
		return nil, fmt.Errorf("object %q/%q not registered", namespace, name)
	}
	if err := wait.PollImmediate(10*time.Millisecond, time.Second, item.hasSynced); err != nil {
		return nil, fmt.Errorf("couldn't propagate object cache: %v", err)
	}

	obj, exists, err := item.store.GetByKey(c.key(namespace, name))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apierrors.NewNotFound(c.groupResource, name)
	}
	if object, ok := obj.(runtime.Object); ok {
		return object, nil
	}
	return nil, fmt.Errorf("unexpected object type: %v", obj)
}
