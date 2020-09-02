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
	"reflect"
	"sort"
	"strings"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
)

type mockFactory struct {
	namespaces sets.String
	objects    map[string]runtime.Object
}

func (m *mockFactory) List(selector labels.Selector) (ret []runtime.Object, err error) {
	panic("implement me")
}

func (m *mockFactory) Get(name string) (runtime.Object, error) {
	if obj, ok := m.objects[name]; ok {
		return obj, nil
	}

	return nil, errors.NewNotFound(schema.GroupResource{}, name)
}

func (m *mockFactory) ByNamespace(namespace string) cache.GenericNamespaceLister {
	panic("not implemented")
}

func (m *mockFactory) Informer() cache.SharedIndexInformer {
	panic("not implemented")
}

func (m *mockFactory) Lister() cache.GenericLister {
	return m
}

func (m *mockFactory) ForResource(namespace string, resource schema.GroupVersionResource) (InformLister, error) {
	return m, nil
}

func (m *mockFactory) Namespaces() sets.String {
	return m.namespaces
}

func TestInformers(t *testing.T) {
	t.Run("TestGet", func(t *testing.T) {
		ifs, err := NewInformersForResource(
			&mockFactory{
				namespaces: sets.NewString("foo", "bar"),
				objects: map[string]runtime.Object{
					"foo": &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
			},
			schema.GroupVersionResource{},
		)
		if err != nil {
			t.Error(err)
			return
		}

		_, err = ifs.Get("foo")
		if err != nil {
			t.Error(err)
			return
		}

		_, err = ifs.Get("bar")
		if !errors.IsNotFound(err) {
			t.Errorf("expected IsNotFound error, got %v", err)
			return
		}
	})
}

func TestNewInformerOptions(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		allowedNamespaces, deniedNamespaces map[string]struct{}
		tweaks                              func(*v1.ListOptions)

		expectedOptions    v1.ListOptions
		expectedNamespaces []string
	}{
		{
			name:               "all unset",
			expectedOptions:    v1.ListOptions{},
			expectedNamespaces: nil,
		},
		{
			name: "allowed namespaces",
			allowedNamespaces: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			expectedOptions: v1.ListOptions{},
			expectedNamespaces: []string{
				"foo",
				"bar",
			},
		},
		{
			name: "allowed namespaces with a tweak",
			allowedNamespaces: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			tweaks: func(options *v1.ListOptions) {
				options.FieldSelector = "metadata.name=foo"
			},

			expectedOptions: v1.ListOptions{
				FieldSelector: "metadata.name=foo",
			},
			expectedNamespaces: []string{
				"foo",
				"bar",
			},
		},
		{
			name: "allowed and ignored denied namespaces",
			allowedNamespaces: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied1": {},
				"denied2": {},
			},

			expectedOptions: v1.ListOptions{},
			expectedNamespaces: []string{
				"foo",
				"bar",
			},
		},
		{
			name: "one allowed namespace and ignored denied namespaces",
			allowedNamespaces: map[string]struct{}{
				"foo": {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied1": {},
				"denied2": {},
			},

			expectedOptions: v1.ListOptions{},
			expectedNamespaces: []string{
				"foo",
			},
		},
		{
			name: "all allowed namespaces denying namespaces",
			allowedNamespaces: map[string]struct{}{
				v1.NamespaceAll: {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied2": {},
				"denied1": {},
			},

			expectedNamespaces: []string{
				v1.NamespaceAll,
			},
			expectedOptions: v1.ListOptions{
				FieldSelector: "metadata.namespace!=denied1,metadata.namespace!=denied2",
			},
		},
		{
			name: "denied namespaces with tweak",
			allowedNamespaces: map[string]struct{}{
				v1.NamespaceAll: {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied2": {},
				"denied1": {},
			},
			tweaks: func(options *v1.ListOptions) {
				options.FieldSelector = "metadata.name=foo"
			},

			expectedNamespaces: []string{
				v1.NamespaceAll,
			},
			expectedOptions: v1.ListOptions{
				FieldSelector: "metadata.name=foo,metadata.namespace!=denied1,metadata.namespace!=denied2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tweaks, namespaces := newInformerOptions(tc.allowedNamespaces, tc.deniedNamespaces, tc.tweaks)
			opts := v1.ListOptions{}
			tweaks(&opts)

			// sort the field selector as entries are in non-deterministic order
			sortFieldSelector := func(opts *v1.ListOptions) {
				fs := strings.Split(opts.FieldSelector, ",")
				sort.Strings(fs)
				opts.FieldSelector = strings.Join(fs, ",")
			}
			sortFieldSelector(&opts)
			sortFieldSelector(&tc.expectedOptions)

			if !reflect.DeepEqual(tc.expectedOptions, opts) {
				t.Errorf("expected list options %v, got %v", tc.expectedOptions, opts)
			}

			// sort namespaces as entries are in non-deterministic order
			sort.Strings(namespaces)
			sort.Strings(tc.expectedNamespaces)

			if !reflect.DeepEqual(tc.expectedNamespaces, namespaces) {
				t.Errorf("expected namespaces %v, got %v", tc.expectedNamespaces, namespaces)
			}
		})
	}
}
