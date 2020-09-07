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

	informers "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NewKubeInformerFactories creates factories for monitoring resources
// for the given allowed, and denied namespaces these parameters being mutually exclusive.
// monitoringClient, defaultResync, and tweakListOptions are being passed to the underlying informer factory.
func NewMonitoringInformerFactories(
	allowNamespaces, denyNamespaces map[string]struct{},
	monitoringClient monitoring.Interface,
	defaultResync time.Duration,
	tweakListOptions func(*metav1.ListOptions),
) FactoriesForNamespaces {
	tweaks, namespaces := newInformerOptions(
		allowNamespaces, denyNamespaces, tweakListOptions,
	)

	opts := []informers.SharedInformerOption{informers.WithTweakListOptions(tweaks)}

	ret := monitoringInformersForNamespaces{}
	for _, namespace := range namespaces {
		opts = append(opts, informers.WithNamespace(namespace))
		ret[namespace] = informers.NewSharedInformerFactoryWithOptions(monitoringClient, defaultResync, opts...)
	}

	return ret
}

type monitoringInformersForNamespaces map[string]informers.SharedInformerFactory

func (i monitoringInformersForNamespaces) Namespaces() sets.String {
	return sets.StringKeySet(i)
}

func (i monitoringInformersForNamespaces) ForResource(namespace string, resource schema.GroupVersionResource) (InformLister, error) {
	return i[namespace].ForResource(resource)
}
