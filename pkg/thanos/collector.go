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

package thanos

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

var (
	descThanosSpecReplicas = prometheus.NewDesc(
		"prometheus_operator_spec_replicas",
		"Number of expected replicas for the object.",
		[]string{
			"namespace",
			"name",
		}, nil,
	)
)

type thanosRulerCollector struct {
	stores []cache.Store
}

// newThanosRulerCollectorForStores creates a thanosRulerCollector initialized with the given cache store
func newThanosRulerCollectorForStores(s ...cache.Store) *thanosRulerCollector {
	return &thanosRulerCollector{stores: s}
}

// Describe implements the prometheus.Collector interface.
func (c *thanosRulerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descThanosSpecReplicas
}

// Collect implements the prometheus.Collector interface.
func (c *thanosRulerCollector) Collect(ch chan<- prometheus.Metric) {
	for _, s := range c.stores {
		for _, tr := range s.List() {
			c.collectThanos(ch, tr.(*v1.ThanosRuler))
		}
	}
}

func (c *thanosRulerCollector) collectThanos(ch chan<- prometheus.Metric, tr *v1.ThanosRuler) {
	replicas := float64(minReplicas)
	if tr.Spec.Replicas != nil {
		replicas = float64(*tr.Spec.Replicas)
	}
	ch <- prometheus.MustNewConstMetric(descThanosSpecReplicas, prometheus.GaugeValue, replicas, tr.Namespace, tr.Name)
}
