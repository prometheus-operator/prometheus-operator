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

package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	descPrometheusSpecReplicas = prometheus.NewDesc(
		"prometheus_operator_spec_replicas",
		"Number of expected replicas for the object.",
		[]string{
			"namespace",
			"name",
		}, nil,
	)
	descPrometheusSpecShards = prometheus.NewDesc(
		"prometheus_operator_spec_shards",
		"Number of expected shards for the object.",
		[]string{
			"namespace",
			"name",
		}, nil,
	)
	descPrometheusEnforcedSampleLimit = prometheus.NewDesc(
		"prometheus_operator_prometheus_enforced_sample_limit",
		"Global limit on the number of scraped samples per scrape target.",
		[]string{
			"namespace",
			"name",
		}, nil,
	)
)

type Collector struct {
	stores []cache.Store
}

func NewCollectorForStores(s ...cache.Store) *Collector {
	return &Collector{stores: s}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descPrometheusSpecReplicas
	ch <- descPrometheusEnforcedSampleLimit
	ch <- descPrometheusSpecShards
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for _, s := range c.stores {
		for _, p := range s.List() {
			c.collectPrometheus(ch, p.(v1.PrometheusInterface))
		}
	}
}

func (c *Collector) collectPrometheus(ch chan<- prometheus.Metric, p v1.PrometheusInterface) {
	replicas := float64(MinReplicas)
	cpf := p.GetCommonPrometheusFields()
	namespace := p.GetObjectMeta().GetNamespace()
	name := p.GetObjectMeta().GetName()

	if cpf.Replicas != nil {
		replicas = float64(*cpf.Replicas)
	}
	ch <- prometheus.MustNewConstMetric(descPrometheusSpecReplicas, prometheus.GaugeValue, replicas, namespace, name)
	// Include EnforcedSampleLimit in metrics if set in Prometheus object.
	if cpf.EnforcedSampleLimit != nil {
		ch <- prometheus.MustNewConstMetric(descPrometheusEnforcedSampleLimit, prometheus.GaugeValue, float64(*cpf.EnforcedSampleLimit), namespace, name)
	}
	if cpf.Shards != nil {
		ch <- prometheus.MustNewConstMetric(descPrometheusSpecShards, prometheus.GaugeValue, float64(*cpf.Shards), namespace, name)
	}
}
