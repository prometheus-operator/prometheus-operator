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
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type mockListerWatcher struct {
	listResult runtime.Object
	evCh       chan watch.Event
	stopped    bool
}

func (m *mockListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	return m.listResult, nil
}

func (m *mockListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return m, nil
}

func (m *mockListerWatcher) Stop() {
	m.stopped = true
}

func (m *mockListerWatcher) ResultChan() <-chan watch.Event {
	return m.evCh
}

func newUnstructured(namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      "foo",
			},
		}}
}

func newNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func namespaces(ns ...string) map[string]struct{} {
	namespaces := map[string]struct{}{}
	for _, n := range ns {
		namespaces[n] = struct{}{}
	}
	return namespaces
}

func TestDenylistList(t *testing.T) {
	logger := log.NewNopLogger()

	cases := []struct {
		name           string
		items          []runtime.RawExtension
		denylist, want map[string]struct{}
	}{
		{
			name: "deny one",
			items: []runtime.RawExtension{
				{
					Object: newUnstructured("monitoring"),
				},
				{
					Object: newUnstructured("default"),
				},
				{
					Object: newUnstructured("kube-system"),
				},
			},
			denylist: namespaces("monitoring"),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "namespaces deny one",
			items: []runtime.RawExtension{
				{
					Object: newNamespace("monitoring"),
				},
				{
					Object: newNamespace("default"),
				},
				{
					Object: newNamespace("kube-system"),
				},
			},
			denylist: namespaces("monitoring"),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "deny many",
			items: []runtime.RawExtension{
				{
					Object: newUnstructured("monitoring"),
				},
				{
					Object: newUnstructured("default"),
				},
				{
					Object: newUnstructured("kube-system"),
				},
			},
			denylist: namespaces("monitoring", "kube-system"),
			want:     namespaces("default"),
		},
		{
			name: "namespaces deny many",
			items: []runtime.RawExtension{
				{
					Object: newNamespace("monitoring"),
				},
				{
					Object: newNamespace("default"),
				},
				{
					Object: newNamespace("kube-system"),
				},
			},
			denylist: namespaces("monitoring", "kube-system"),
			want:     namespaces("default"),
		},
		{
			name: "deny none",
			items: []runtime.RawExtension{
				{
					Object: newUnstructured("monitoring"),
				},
				{
					Object: newUnstructured("default"),
				},
				{
					Object: newUnstructured("kube-system"),
				},
			},
			want: namespaces("monitoring", "default", "kube-system"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := metav1.List{}
			l.Items = tc.items
			mock := &mockListerWatcher{
				listResult: &l,
			}

			lw := newDenylistListerWatcher(logger, tc.denylist, mock)
			result, err := lw.List(metav1.ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			objs, err := meta.ExtractList(result)
			if err != nil {
				t.Error(err)
				return
			}

			got := map[string]struct{}{}
			for _, obj := range objs {
				acc, err := meta.Accessor(obj)
				if err != nil {
					t.Error(err)
					return
				}
				got[getNamespace(acc)] = struct{}{}
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("want listed namespaces to be %q, got %q", tc.want, got)
			}
		})
	}
}

func TestDenylistWatch(t *testing.T) {
	logger := log.NewNopLogger()

	cases := []struct {
		name           string
		items          []runtime.Object
		denylist, want map[string]struct{}
	}{
		{
			name: "deny one",
			items: []runtime.Object{
				newUnstructured("monitoring"),
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: namespaces("monitoring"),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "namespaces deny one",
			items: []runtime.Object{
				newNamespace("monitoring"),
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: namespaces("monitoring"),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "deny many",
			items: []runtime.Object{
				newUnstructured("monitoring"),
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: namespaces("monitoring", "kube-system"),
			want:     namespaces("default"),
		},
		{
			name: "namespces deny many",
			items: []runtime.Object{
				newNamespace("monitoring"),
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: namespaces("monitoring", "kube-system"),
			want:     namespaces("default"),
		},
		{
			name: "denylist contains empty string",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: namespaces(""),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "namespaces denylist contains empty string",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: namespaces(""),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "empty denylist",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: namespaces(),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "namespaces empty denylist",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: namespaces(),
			want:     namespaces("default", "kube-system"),
		},
		{
			name: "nil denylist",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			want: namespaces("default", "kube-system"),
		},
		{
			name: "namespaces nil denylist",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			want: namespaces("default", "kube-system"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			events := make(chan watch.Event, len(tc.items))
			for i := range tc.items {
				events <- watch.Event{
					Type:   "foo",
					Object: tc.items[i],
				}
			}
			close(events)
			mock := &mockListerWatcher{
				evCh: events,
			}
			lw := newDenylistListerWatcher(logger, tc.denylist, mock)
			w, err := lw.Watch(metav1.ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}

			for evt := range w.ResultChan() {
				acc, err := meta.Accessor(evt.Object)
				if err != nil {
					t.Error(err)
					return
				}
				got := getNamespace(acc)
				if _, ok := tc.want[got]; !ok {
					t.Errorf("unexpected namespace %v, should have been denied", got)
				}
				delete(tc.want, got)
			}
			if len(tc.want) != 0 {
				t.Errorf("namespace(s) not used %v", tc.want)
			}

			if evt, open := <-events; open {
				t.Errorf("expected all events to be processed, but they aren't: %v", evt.Object)
			}
		})
	}
}
