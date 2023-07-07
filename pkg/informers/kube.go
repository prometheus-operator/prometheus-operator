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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
)

// NewKubeInformerFactories creates factories for kube resources
// for the given allowed, and denied namespaces these parameters being mutually exclusive.
// kubeClient, defaultResync, and tweakListOptions are being passed to the underlying informer factory.
func NewKubeInformerFactories(
	allowNamespaces, denyNamespaces map[string]struct{},
	kubeClient kubernetes.Interface,
	defaultResync time.Duration,
	tweakListOptions func(*metav1.ListOptions),
) FactoriesForNamespaces {
	tweaks, namespaces := newInformerOptions(
		allowNamespaces, denyNamespaces, tweakListOptions,
	)

	opts := []informers.SharedInformerOption{informers.WithTweakListOptions(tweaks)}

	ret := kubeInformersForNamespaces{}
	for _, namespace := range namespaces {
		opts = append(opts, informers.WithNamespace(namespace))
		ret[namespace] = informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResync, opts...)
	}
	return ret
}

// NewMetadataInformerFactory creates metadatainformer factory for kube resources
// for the given allowed, and denied namespaces (these parameters being mutually exclusive).
// mdClient, defaultResync, and tweakListOptions are  passed to the underlying informer factory.
// factories
func NewMetadataInformerFactory(
	allowNamespaces, denyNamespaces map[string]struct{},
	mdClient metadata.Interface,
	defaultResync time.Duration,
	tweakListOptions func(*metav1.ListOptions),
) FactoriesForNamespaces {

	tweaks, namespaces := newInformerOptions(allowNamespaces, denyNamespaces, tweakListOptions)

	ret := metadataInformersForNamespace{}
	for _, namespace := range namespaces {
		ret[namespace] = metadatainformer.NewFilteredSharedInformerFactory(mdClient, defaultResync, namespace, tweaks)
	}

	return ret
}

type kubeInformersForNamespaces map[string]informers.SharedInformerFactory

func (i kubeInformersForNamespaces) Namespaces() sets.Set[string] {
	return sets.KeySet(i)
}

func (i kubeInformersForNamespaces) ForResource(namespace string, resource schema.GroupVersionResource) (InformLister, error) {
	return i[namespace].ForResource(resource)
}

type metadataInformersForNamespace map[string]metadatainformer.SharedInformerFactory

func (i metadataInformersForNamespace) Namespaces() sets.Set[string] {
	return sets.KeySet(i)
}

func (i metadataInformersForNamespace) ForResource(namespace string, resource schema.GroupVersionResource) (InformLister, error) {
	return i[namespace].ForResource(resource), nil
}
