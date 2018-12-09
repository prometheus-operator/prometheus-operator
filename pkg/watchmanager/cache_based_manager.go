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

package watchmanager

import (
	"sync"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetObjectTTLFunc defines a function to get value of TTL.
type GetObjectTTLFunc func() (time.Duration, bool)

// GetObjectFunc defines a function to get object with a given namespace and name.
type GetObjectFunc func(string, string, metav1.GetOptions) (runtime.Object, error)

type objectKey struct {
	namespace string
	name      string
}

// cacheBasedManager keeps a store with objects necessary
// for registered ServiceMonitors. Different implementations of the store
// may result in different semantics for freshness of objects
// (e.g. ttl-based implementation vs watch-based implementation).
type cacheBasedManager struct {
	objectStore          Store
	getReferencedObjects func(*monitoringv1.ServiceMonitor) sets.String

	lock            sync.Mutex
	registeredSMons map[objectKey]*monitoringv1.ServiceMonitor
}

func (c *cacheBasedManager) GetObject(namespace, name string) (runtime.Object, error) {
	return c.objectStore.Get(namespace, name)
}

func (c *cacheBasedManager) RegisterSMon(sMon *monitoringv1.ServiceMonitor) {
	names := c.getReferencedObjects(sMon)
	c.lock.Lock()
	defer c.lock.Unlock()
	for name := range names {
		c.objectStore.AddReference(sMon.Namespace, name)
	}
	var prev *monitoringv1.ServiceMonitor
	key := objectKey{namespace: sMon.Namespace, name: sMon.Name}
	prev = c.registeredSMons[key]
	c.registeredSMons[key] = sMon
	if prev != nil {
		for name := range c.getReferencedObjects(prev) {
			// On an update, the .Add() call above will have re-incremented the
			// ref count of any existing object, so any objects that are in both
			// names and prev need to have their ref counts decremented. Any that
			// are only in prev need to be completely removed. This unconditional
			// call takes care of both cases.
			c.objectStore.DeleteReference(prev.Namespace, name)
		}
	}
}

func (c *cacheBasedManager) UnregisterSMon(sMon *monitoringv1.ServiceMonitor) {
	var prev *monitoringv1.ServiceMonitor
	key := objectKey{namespace: sMon.Namespace, name: sMon.Name}
	c.lock.Lock()
	defer c.lock.Unlock()
	prev = c.registeredSMons[key]
	delete(c.registeredSMons, key)
	if prev != nil {
		for name := range c.getReferencedObjects(prev) {
			c.objectStore.DeleteReference(prev.Namespace, name)
		}
	}
}

// NewCacheBasedManager creates a manager that keeps a cache of all objects
// necessary for registered ServiceMonitors.
// It implements the following logic:
// - whenever a sMon is created or updated, the cached versions of all objects
//   is is referencing are invalidated
// - every GetObject() call tries to fetch the value from local cache; if it is
//   not there, invalidated or too old, we fetch it from apiserver and refresh the
//   value in cache; otherwise it is just fetched from cache
func NewCacheBasedManager(objectStore Store, getReferencedObjects func(*monitoringv1.ServiceMonitor) sets.String) Manager {
	return &cacheBasedManager{
		objectStore:          objectStore,
		getReferencedObjects: getReferencedObjects,
		registeredSMons:      make(map[objectKey]*monitoringv1.ServiceMonitor),
	}
}
