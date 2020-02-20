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

func TestDenylistList(t *testing.T) {
	logger := log.NewNopLogger()

	cases := []struct {
		name           string
		items          []runtime.RawExtension
		denylist, want []string
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
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
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
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
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
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
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
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
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
			want: []string{"monitoring", "default", "kube-system"},
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

			var got []string
			for _, obj := range objs {
				acc, err := meta.Accessor(obj)
				if err != nil {
					t.Error(err)
					return
				}
				got = append(got, getNamespace(acc))
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
		denylist, want []string
	}{
		{
			name: "deny one",
			items: []runtime.Object{
				newUnstructured("monitoring"),
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "namespaces deny one",
			items: []runtime.Object{
				newNamespace("monitoring"),
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "deny many",
			items: []runtime.Object{
				newUnstructured("monitoring"),
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
		},
		{
			name: "namespces deny many",
			items: []runtime.Object{
				newNamespace("monitoring"),
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
		},
		{
			name: "denylist contains empty string",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: []string{""},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "namespaces denylist contains empty string",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: []string{""},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "empty denylist",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			denylist: []string{},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "namespaces empty denylist",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			denylist: []string{},
			want:     []string{"default", "kube-system"},
		},
		{
			name: "nil denylist",
			items: []runtime.Object{
				newUnstructured("default"),
				newUnstructured("kube-system"),
			},
			want: []string{"default", "kube-system"},
		},
		{
			name: "namespaces nil denylist",
			items: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
			},
			want: []string{"default", "kube-system"},
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

			for i := 0; i < len(tc.want); i++ {
				evt := <-w.ResultChan()
				acc, err := meta.Accessor(evt.Object)
				if err != nil {
					t.Error(err)
					return
				}

				if got := getNamespace(acc); got != tc.want[i] {
					t.Errorf("want namespace %v, evt %v", tc.want[i], got)
				}
			}

			if evt, open := <-events; open {
				t.Errorf("expected all events to be processed, but they aren't: %v", evt.Object)
			}
		})
	}
}
