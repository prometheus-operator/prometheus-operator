// Copyright 2016 The prometheus-operator Authors
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

package listwatch

import (
	"context"

	"github.com/go-kit/kit/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
)

// NewUnprivilegedNamespaceListWatchFromClient mimics
// cache.NewListWatchFromClient.
// It allows for the creation of a cache.ListWatch for namespaces from a client
// that does not have `List` privileges. If the slice of namespaces contains
// only v1.NamespaceAll, then this func assumes that the client has List and
// Watch privileges and returns a regular cache.ListWatch, since there is no
// other way to get all namespaces.
//
// The allowed namespaces and denied namespaces are mutually exclusive.
// See NewFilteredUnprivilegedNamespaceListWatchFromClient for a description on how they are applied.
func NewUnprivilegedNamespaceListWatchFromClient(l log.Logger, c cache.Getter, allowedNamespaces, deniedNamespaces map[string]struct{}, fieldSelector fields.Selector) cache.ListerWatcher {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fieldSelector.String()
	}
	return NewFilteredUnprivilegedNamespaceListWatchFromClient(l, c, allowedNamespaces, deniedNamespaces, optionsModifier)
}

// NewFilteredUnprivilegedNamespaceListWatchFromClient mimics
// cache.NewUnprivilegedNamespaceListWatchFromClient.
// It allows for the creation of a cache.ListWatch for allowed or denied namespaces
// from a client that does not have `List` privileges.
//
// If the given allowed namespaces contain only v1.NamespaceAll,
// then this function assumes that the client has List and
// Watch privileges and returns a regular cache.ListWatch, since there is no
// other way to get all namespaces.
//
// The given allowed and denied namespaces are mutually exclusive.
// If allowed namespaces contain multiple items, the given denied namespaces have no effect.
// If the allowed namespaces includes exactly one entry with the value v1.NamespaceAll (empty string),
// the given denied namespaces are applied.
func NewFilteredUnprivilegedNamespaceListWatchFromClient(l log.Logger, c cache.Getter, allowedNamespaces, deniedNamespaces map[string]struct{}, optionsModifier func(options *metav1.ListOptions)) cache.ListerWatcher {
	// If the only namespace given is `v1.NamespaceAll`, then this
	// cache.ListWatch must be privileged. In this case, return a regular
	// cache.ListWatch decorated with a denylist watcher
	// filtering the given denied namespaces.
	if IsAllNamespaces(allowedNamespaces) {
		return newAllowDenylistListerWatcher(l, nil, deniedNamespaces, cache.NewFilteredListWatchFromClient(c, "namespaces", metav1.NamespaceAll, optionsModifier))
	}
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		optionsModifier(&options)
		list := &v1.NamespaceList{}
		for name := range allowedNamespaces {
			result := &v1.Namespace{}
			err := c.Get().
				Resource("namespaces").
				Name(name).
				VersionedParams(&options, scheme.ParameterCodec).
				Do(context.TODO()).
				Into(result)
			if err != nil {
				return nil, err
			}
			list.Items = append(list.Items, *result)
		}
		return list, nil
	}
	watchFunc := func(_ metav1.ListOptions) (watch.Interface, error) {
		// Since the client does not have Watch privileges, do not
		// actually watch anything. Use a watch.FakeWatcher here to
		// implement watch.Interface but not send any events.
		return watch.NewFake(), nil
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

// MultiNamespaceListerWatcher takes allowed and denied namespaces and a
// cache.ListerWatcher generator func and returns a single cache.ListerWatcher
// capable of operating on multiple namespaces.
//
// Allowed namespaces and denied namespaces are mutually exclusive.
// If allowed namespaces contain multiple items, the given denied namespaces have no effect.
// If the allowed namespaces includes exactly one entry with the value v1.NamespaceAll (empty string),
// the given denied namespaces are applied.
func MultiNamespaceListerWatcher(l log.Logger, allowedNamespaces, deniedNamespaces map[string]struct{}, f func(string) cache.ListerWatcher) cache.ListerWatcher {
	// If there is only one namespace then there is no need to create a
	// multi lister watcher proxy.
	if IsAllNamespaces(allowedNamespaces) {
		return newAllowDenylistListerWatcher(l, nil, deniedNamespaces, f(v1.NamespaceAll))
	}
	if len(allowedNamespaces) == 1 {
		for n := range allowedNamespaces {
			return f(n)
		}
	}

	return newAllowDenylistListerWatcher(l, allowedNamespaces, nil, f(v1.NamespaceAll))
}

// IsAllNamespaces checks if the given map of namespaces
// contains only v1.NamespaceAll.
func IsAllNamespaces(namespaces map[string]struct{}) bool {
	_, ok := namespaces[v1.NamespaceAll]
	return ok && len(namespaces) == 1
}

// IdenticalNamespaces returns true if a and b are identical.
func IdenticalNamespaces(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}

	return true
}
