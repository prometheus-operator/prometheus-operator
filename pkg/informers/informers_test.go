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
	"context"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/metadata/fake"
	kubetesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type mockFactory struct {
	namespaces sets.Set[string]
	objects    map[string]runtime.Object
}

func (m *mockFactory) List(_ labels.Selector) (ret []runtime.Object, err error) {
	panic("implement me")
}

func (m *mockFactory) Get(name string) (runtime.Object, error) {
	if obj, ok := m.objects[name]; ok {
		return obj, nil
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{}, name)
}

func (m *mockFactory) ByNamespace(_ string) cache.GenericNamespaceLister {
	panic("not implemented")
}

func (m *mockFactory) Informer() cache.SharedIndexInformer {
	panic("not implemented")
}

func (m *mockFactory) Lister() cache.GenericLister {
	return m
}

func (m *mockFactory) ForResource(_ string, _ schema.GroupVersionResource) (InformLister, error) {
	return m, nil
}

func (m *mockFactory) Namespaces() sets.Set[string] {
	return m.namespaces
}

func TestInformers(t *testing.T) {
	t.Run("TestGet", func(t *testing.T) {
		ifs, err := NewInformersForResource(
			&mockFactory{
				namespaces: sets.New[string]("foo", "bar"),
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
		if !apierrors.IsNotFound(err) {
			t.Errorf("expected IsNotFound error, got %v", err)
			return
		}
	})
}

func TestNewInformerOptions(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		allowedNamespaces, deniedNamespaces map[string]struct{}
		tweaks                              func(*metav1.ListOptions)

		expectedOptions    metav1.ListOptions
		expectedNamespaces []string
	}{
		{
			name:               "all unset",
			expectedOptions:    metav1.ListOptions{},
			expectedNamespaces: nil,
		},
		{
			name: "allowed namespaces",
			allowedNamespaces: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			expectedOptions: metav1.ListOptions{},
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
			tweaks: func(options *metav1.ListOptions) {
				options.FieldSelector = "metadata.name=foo"
			},

			expectedOptions: metav1.ListOptions{
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

			expectedOptions: metav1.ListOptions{},
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

			expectedOptions: metav1.ListOptions{},
			expectedNamespaces: []string{
				"foo",
			},
		},
		{
			name: "all allowed namespaces denying namespaces",
			allowedNamespaces: map[string]struct{}{
				metav1.NamespaceAll: {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied2": {},
				"denied1": {},
			},

			expectedNamespaces: []string{
				metav1.NamespaceAll,
			},
			expectedOptions: metav1.ListOptions{
				FieldSelector: "metadata.namespace!=denied1,metadata.namespace!=denied2",
			},
		},
		{
			name: "denied namespaces with tweak",
			allowedNamespaces: map[string]struct{}{
				metav1.NamespaceAll: {},
			},
			deniedNamespaces: map[string]struct{}{
				"denied2": {},
				"denied1": {},
			},
			tweaks: func(options *metav1.ListOptions) {
				options.FieldSelector = "metadata.name=foo"
			},

			expectedNamespaces: []string{
				metav1.NamespaceAll,
			},
			expectedOptions: metav1.ListOptions{
				FieldSelector: "metadata.name=foo,metadata.namespace!=denied1,metadata.namespace!=denied2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tweaks, namespaces := newInformerOptions(tc.allowedNamespaces, tc.deniedNamespaces, tc.tweaks)
			opts := metav1.ListOptions{}
			tweaks(&opts)

			// sort the field selector as entries are in non-deterministic order
			sortFieldSelector := func(opts *metav1.ListOptions) {
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

// TestPartialObjectMetadataStripOnDeletedFinalStateUnknown makes sure
// that PartialObjectMetadataStrip doesn't fail on DeletedFinalStateUnknown.
func TestPartialObjectMetadataStripOnDeletedFinalStateUnknown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Mock the following scenario:
	// 1. the informer lists the secrets and the API returns 1 item.
	// 2. the informer watches the secrets resource and the API returns a watch error.
	// 3. the informer lists again the secrets and the API returns no item.
	//
	// After the third step, the informer should send a delete event with a
	// "cache.DeletedFinalStateUnknown" object.
	fakeClient := fake.NewSimpleMetadataClient(fake.NewTestScheme())
	listCalls, watchCalls := &atomic.Uint64{}, &atomic.Uint64{}
	fakeClient.PrependReactor("list", "secrets", func(_ kubetesting.Action) (bool, runtime.Object, error) {
		objects := &metav1.List{
			Items: []runtime.RawExtension{},
		}

		// The first call to list returns 1 item. Subsequent calls returns an empty list.
		if listCalls.Load() == 0 {
			objects.Items = []runtime.RawExtension{
				{Object: &metav1.PartialObjectMetadata{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "777",
					},
				}},
			}
		}
		listCalls.Add(1)

		return true, objects, nil
	})

	fakeClient.PrependWatchReactor("secrets", func(_ kubetesting.Action) (handled bool, ret watch.Interface, err error) {
		w := watch.NewRaceFreeFake()

		// Trigger a watch error after the first list operation.
		if listCalls.Load() == 1 {
			w.Error(&apierrors.NewResourceExpired("expired").ErrStatus)
		}

		watchCalls.Add(1)
		return true, w, nil
	})

	infs, err := NewInformersForResourceWithTransform(
		NewMetadataInformerFactory(
			map[string]struct{}{"bar": {}},
			map[string]struct{}{},
			fakeClient,
			time.Second,
			nil,
		),
		appsv1.SchemeGroupVersion.WithResource("secrets"),
		PartialObjectMetadataStrip(schema.GroupVersionKind{Version: "v1", Kind: "Secret"}),
	)
	require.NoError(t, err)

	var (
		addCount    = &atomic.Uint64{}
		delReceived = make(chan struct{})
	)
	infs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			t.Logf("added object %T", obj)
			addCount.Add(1)
		},
		DeleteFunc: func(obj any) {
			t.Logf("deleted object %T", obj)
			close(delReceived)
		},
	})

	errCh := make(chan error, 1)
	for _, inf := range infs.informers {
		inf.Informer().SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
			errCh <- err
		})
	}

	go infs.Start(ctx.Done())

	select {
	case <-delReceived:
	case err = <-errCh:
		require.NoError(t, err)
	case <-ctx.Done():
		require.FailNow(t, "timeout waiting for the delete event")
	}

	// List should be called twice.
	require.Equal(t, uint64(2), listCalls.Load())

	// Watch should be called at least once.
	require.GreaterOrEqual(t, watchCalls.Load(), uint64(1))
	// 1 object should have been added.
	require.Equal(t, uint64(1), addCount.Load())
}
