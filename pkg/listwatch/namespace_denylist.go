// Copyright 2019 The prometheus-operator Authors
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
	"fmt"
	"log/slog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// denylistListerWatcher implements cache.ListerWatcher
// which wraps a cache.ListerWatcher,
// filtering list results and watch events by denied namespaces.
type denylistListerWatcher struct {
	denylist map[string]struct{}
	next     cache.ListerWatcher
	logger   *slog.Logger
}

// newDenylistListerWatcher creates a cache.ListerWatcher
// wrapping the given next cache.ListerWatcher
// filtering lists and watch events by the given namespaces.
func newDenylistListerWatcher(l *slog.Logger, namespaces map[string]struct{}, next cache.ListerWatcher) cache.ListerWatcher {
	if len(namespaces) == 0 {
		return next
	}

	return &denylistListerWatcher{
		denylist: namespaces,
		next:     next,
		logger:   l,
	}
}

// List lists the wrapped next listerwatcher List result,
// but filtering denied namespaces from the result.
func (w *denylistListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	var (
		l = metav1.List{}
	)

	list, err := w.next.List(options)
	if err != nil {
		w.logger.Error("error listing", "err", err)
		return nil, err
	}

	objs, err := meta.ExtractList(list)
	if err != nil {
		w.logger.Error("error extracting list", "err", err)
		return nil, err
	}

	metaObj, err := meta.ListAccessor(list)
	if err != nil {
		w.logger.Error("error getting list accessor", "err", err)
		return nil, err
	}

	for _, obj := range objs {
		acc, err := meta.Accessor(obj)
		if err != nil {
			w.logger.Error("error getting meta accessor accessor", "obj", fmt.Sprintf("%v", obj), "err", err)
			return nil, err
		}

		debugDetailed := w.logger.With("selflink", acc.GetSelfLink())

		if _, denied := w.denylist[getNamespace(acc)]; denied {
			debugDetailed.Debug("denied")
			continue
		}

		debugDetailed.Debug("allowed")

		l.Items = append(l.Items, runtime.RawExtension{Object: obj.DeepCopyObject()})
	}

	l.ResourceVersion = metaObj.GetResourceVersion()
	return &l, nil
}

// Watch.
func (w *denylistListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	nextWatch, err := w.next.Watch(options)
	if err != nil {
		return nil, err
	}

	return newDenylistWatch(w.logger, w.denylist, nextWatch), nil
}

// newDenylistWatch creates a new watch.Interface,
// wrapping the given next watcher,
// and filtering watch events by the given namespaces.
//
// It starts a new goroutine until either
// a) the result channel of the wrapped next watcher is closed, or
// b) Stop() was invoked on the returned watcher.
func newDenylistWatch(l *slog.Logger, denylist map[string]struct{}, next watch.Interface) watch.Interface {
	var (
		result = make(chan watch.Event)
		proxy  = watch.NewProxyWatcher(result)
	)

	go func() {
		defer func() {
			l.Debug("stopped denylist watcher")
			// According to watch.Interface the result channel is supposed to be called
			// in case of error or if the listwach is closed, see [1].
			//
			// [1] https://github.com/kubernetes/apimachinery/blob/533d101be9a6450773bb2829bef282b6b7c4ff6d/pkg/watch/watch.go#L34-L37
			close(result)
		}()

		for {
			select {
			case event, ok := <-next.ResultChan():
				if !ok {
					l.Debug("result channel closed")
					return
				}

				acc, err := meta.Accessor(event.Object)
				if err != nil {
					// ignore this event, it doesn't implement the metav1.Object interface,
					// hence we cannot determine its namespace.
					l.Warn(fmt.Sprintf("unexpected object type in event (%T): %v", event.Object, event.Object))
					continue
				}

				debugDetailed := l.With("selflink", acc.GetSelfLink())
				if _, denied := denylist[getNamespace(acc)]; denied {
					debugDetailed.Debug("denied")
					continue
				}

				debugDetailed.Debug("allowed")

				select {
				case result <- event:
					debugDetailed.Debug("dispatched")
				case <-proxy.StopChan():
					next.Stop()
					return
				}
			case <-proxy.StopChan():
				next.Stop()
				return
			}
		}
	}()

	return proxy
}

// getNamespace returns the namespace of the given object.
// If the object is itself a namespace, it returns the object's
// name.
func getNamespace(obj metav1.Object) string {
	if _, ok := obj.(*v1.Namespace); ok {
		return obj.GetName()
	}
	return obj.GetNamespace()
}
