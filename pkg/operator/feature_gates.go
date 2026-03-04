// Copyright 2024 The prometheus-operator Authors
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

package operator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// PrometheusAgentDaemonSetFeature enables the DaemonSet mode for PrometheusAgent.
	PrometheusAgentDaemonSetFeature FeatureGateName = "PrometheusAgentDaemonSet"

	// PrometheusTopologyShardingFeature enables the zone-aware sharding for Prometheus.
	PrometheusTopologyShardingFeature FeatureGateName = "PrometheusTopologySharding"

	// PrometheusShardRetentionPolicyFeature enables the shard retention policy for Prometheus.
	PrometheusShardRetentionPolicyFeature FeatureGateName = "PrometheusShardRetentionPolicy"

	// StatusForConfigurationResourcesFeature enables the status subresource for Prometheus-Operator Config Objects.
	StatusForConfigurationResourcesFeature FeatureGateName = "StatusForConfigurationResources"

	// RemoteWriteCustomResourceDefinitionFeature enables the RemoteWrite CRD support.
	RemoteWriteCustomResourceDefinitionFeature FeatureGateName = "RemoteWriteCustomResourceDefinition"
)

type FeatureGateName string

type FeatureGates map[FeatureGateName]FeatureGate

type FeatureGate struct {
	description string
	enabled     bool
}

func (fg *FeatureGates) Enabled(name FeatureGateName) bool {
	return (*fg)[name].enabled
}

// UpdateFeatureGates merges the current feature gate values with
// the values provided by the user.
func (fg *FeatureGates) UpdateFeatureGates(flags map[string]bool) error {
	for k := range flags {
		f, found := (*fg)[FeatureGateName(k)]
		if !found {
			return fmt.Errorf("feature gate %q is unknown (supported feature gates: %s)", k, fg.String())
		}
		f.enabled = flags[k]
		(*fg)[FeatureGateName(k)] = f
	}

	return nil
}

func (fg *FeatureGates) keyValuePairs() ([]FeatureGateName, []FeatureGate) {
	if fg == nil {
		return nil, nil
	}

	var (
		names = make([]FeatureGateName, 0, len(*fg))
		gates = make([]FeatureGate, 0, len(*fg))
	)
	for k := range *fg {
		names = append(names, k)
	}
	slices.Sort(names)

	for _, v := range names {
		gates = append(gates, (*fg)[v])
	}

	return names, gates
}

func (fg *FeatureGates) Descriptions() []string {
	var (
		names, gates = fg.keyValuePairs()
		desc         = make([]string, 0, len(names))
	)

	for i := range names {
		desc = append(desc, fmt.Sprintf("%s: %s (enabled: %t)", names[i], gates[i].description, gates[i].enabled))
	}

	return desc
}

func (fg *FeatureGates) String() string {
	names, gates := fg.keyValuePairs()

	s := make([]string, len(names))
	for i := range names {
		s[i] = fmt.Sprintf("%s=%t", names[i], gates[i].enabled)
	}

	return strings.Join(s, ",")
}

var featureGateInfoDesc = prometheus.NewDesc(
	"prometheus_operator_feature_gate",
	"Reports about the Prometheus operator feature gates. A value of 1 means that the feature gate is enabled. Otherwise the value is 0.",
	[]string{"name"},
	nil,
)

// Describe implements the prometheus.Collector interface.
func (fg *FeatureGates) Describe(ch chan<- *prometheus.Desc) {
	ch <- featureGateInfoDesc
}

// Collect implements the prometheus.Collector interface.
func (fg *FeatureGates) Collect(ch chan<- prometheus.Metric) {
	names, gates := fg.keyValuePairs()

	for i, v := range names {
		var val float64
		if gates[i].enabled {
			val = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			featureGateInfoDesc,
			prometheus.GaugeValue,
			val,
			string(v),
		)
	}
}
