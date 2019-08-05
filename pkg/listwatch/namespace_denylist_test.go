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

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

func TestDenylistList(t *testing.T) {
	logger := log.NewNopLogger()

	cases := []struct {
		name                   string
		listed, denylist, want []string
	}{
		{
			name:     "one entry",
			listed:   []string{"monitoring", "default", "kube-system"},
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
		},
		{
			name:     "multiple entries",
			listed:   []string{"monitoring", "default", "kube-system"},
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
		},
		{
			name:   "no entries",
			listed: []string{"monitoring", "default", "kube-system"},
			want:   []string{"monitoring", "default", "kube-system"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := metav1.List{}
			for _, listed := range tc.listed {
				l.Items = append(l.Items, runtime.RawExtension{
					Object: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"namespace": listed,
								"name":      "foo",
							},
						},
					},
				})
			}
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

				got = append(got, acc.GetNamespace())
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
		name                    string
		watched, denylist, want []string
	}{
		{
			name:     "one entry",
			watched:  []string{"monitoring", "default", "kube-system"},
			denylist: []string{"monitoring"},
			want:     []string{"default", "kube-system"},
		},
		{
			name:     "multiple entries",
			watched:  []string{"monitoring", "kube-system", "default"},
			denylist: []string{"monitoring", "kube-system"},
			want:     []string{"default"},
		},
		{
			name:     "denylist contains empty string",
			watched:  []string{"default", "kube-system"},
			denylist: []string{""},
			want:     []string{"default", "kube-system"},
		},
		{
			name:     "empty denylist",
			watched:  []string{"default", "kube-system"},
			denylist: []string{},
			want:     []string{"default", "kube-system"},
		},
		{
			name:    "nil denylist",
			watched: []string{"default", "kube-system"},
			want:    []string{"default", "kube-system"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			events := make(chan watch.Event, len(tc.watched))
			for _, w := range tc.watched {
				events <- watch.Event{
					Type: "foo",
					Object: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"namespace": w,
								"name":      "foo",
							},
						},
					},
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

				if got := acc.GetNamespace(); got != tc.want[i] {
					t.Errorf("want namespace %v, evt %v", tc.want[i], got)
				}
			}

			if evt, open := <-events; open {
				t.Errorf("expected all events to be processed, but they aren't: %v", evt.Object)
			}
		})
	}
}
