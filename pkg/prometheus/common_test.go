// Copyright The prometheus-operator Authors
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
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestStartupProbeTimeoutSeconds(t *testing.T) {
	tests := []struct {
		maximumStartupDurationSeconds   *int32
		expectedStartupPeriodSeconds    int32
		expectedStartupFailureThreshold int32
		expectedMaxStartupDuration      int32
	}{
		{
			maximumStartupDurationSeconds:   nil,
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(0)),
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(1)),
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(60)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 1,
			expectedMaxStartupDuration:      60,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(600)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 10,
			expectedMaxStartupDuration:      600,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(900)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 15,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(1200)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 20,
			expectedMaxStartupDuration:      1200,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(129)),
			expectedStartupPeriodSeconds:    43,
			expectedStartupFailureThreshold: 3,
			expectedMaxStartupDuration:      129,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(322)),
			expectedStartupPeriodSeconds:    54,
			expectedStartupFailureThreshold: 6,
			expectedMaxStartupDuration:      324,
		},
	}

	for _, test := range tests {
		startupPeriodSeconds, startupFailureThreshold := getStatupProbePeriodSecondsAndFailureThreshold(test.maximumStartupDurationSeconds)

		require.Equal(t, test.expectedStartupPeriodSeconds, startupPeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, startupFailureThreshold)
		require.Equal(t, test.expectedMaxStartupDuration, startupPeriodSeconds*startupFailureThreshold)
	}
}

func TestBuildCommonPrometheusArgsWithRemoteWriteMessageV2(t *testing.T) {
	for _, tc := range []struct {
		version        string
		messageVersion *monitoringv1.RemoteWriteMessageVersion

		expectedPresent bool
	}{
		{
			version: "v2.53.0",
		},
		{
			version:        "v2.53.0",
			messageVersion: ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
		},
		{
			version: "v2.54.0",
		},
		{
			version:        "v2.54.0",
			messageVersion: ptr.To(monitoringv1.RemoteWriteMessageVersion1_0),
		},
		{
			version:         "v2.54.0",
			messageVersion:  ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
			expectedPresent: true,
		},
		{
			version:         "v3.4.0",
			messageVersion:  ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
			expectedPresent: false,
		},
	} {
		t.Run("", func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.version,
						RemoteWrite: []monitoringv1.RemoteWriteSpec{
							{
								URL:            monitoringv1.URL("http://example.com"),
								MessageVersion: tc.messageVersion,
							},
						},
					},
				},
			}

			cg, err := NewConfigGenerator(NewLogger(), p)
			require.NoError(t, err)

			args := cg.BuildCommonPrometheusArgs()

			var found bool
			for _, arg := range args {
				if arg.Name == "enable-feature" && arg.Value == "metadata-wal-records" {
					found = true
					break
				}
			}

			require.Equal(t, tc.expectedPresent, found)
		})
	}
}

func TestBuildCommonPrometheusArgsWithOTLPReceiver(t *testing.T) {
	for _, tc := range []struct {
		version                    string
		enableOTLPReceiver         *bool
		expectedOTLPReceiverFlag   bool
		OTLPConfig                 *monitoringv1.OTLPConfig
		expectedOTLPFeatureEnabled bool
	}{
		// OTLP receiver not supported.
		{
			version:                    "2.46.0",
			enableOTLPReceiver:         ptr.To(true),
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   false,
		},
		// OTLP receiver supported starting with v2.47.0.
		{
			version:                    "2.47.0",
			enableOTLPReceiver:         ptr.To(true),
			expectedOTLPFeatureEnabled: true,
			expectedOTLPReceiverFlag:   false,
		},
		// OTLP receiver supported but not enabled.
		{
			version:                    "2.47.0",
			enableOTLPReceiver:         ptr.To(false),
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   false,
		},
		// OTLP receiver config supported but version not support
		{
			version:            "2.46.0",
			enableOTLPReceiver: ptr.To(false),
			OTLPConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{"aa", "bb"},
			},
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   false,
		},
		// OTLP receiver config supported
		{
			version:            "2.55.0",
			enableOTLPReceiver: nil,
			OTLPConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{"aa", "bb"},
			},
			expectedOTLPFeatureEnabled: true,
			expectedOTLPReceiverFlag:   false,
		},
		// OTLP receiver config supported with version 3.x
		{
			version:            "3.0.0",
			enableOTLPReceiver: nil,
			OTLPConfig: &monitoringv1.OTLPConfig{
				PromoteResourceAttributes: []string{"aa", "bb"},
			},
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   true,
		},
		// Test higher version from which enable-feature available.
		{
			version:                    "2.54.0",
			enableOTLPReceiver:         ptr.To(true),
			expectedOTLPFeatureEnabled: true,
			expectedOTLPReceiverFlag:   false,
		},
		// Test higher version from which web.enable-otlp-receiver arg available.
		{
			version:                    "3.0.0",
			enableOTLPReceiver:         ptr.To(true),
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   true,
		},
		// Test higher version but not enabled.
		{
			version:                    "3.0.0",
			enableOTLPReceiver:         ptr.To(false),
			expectedOTLPFeatureEnabled: false,
			expectedOTLPReceiverFlag:   false,
		},
	} {
		t.Run("", func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:            tc.version,
						EnableOTLPReceiver: tc.enableOTLPReceiver,
						OTLP:               tc.OTLPConfig,
					},
				},
			}

			cg, err := NewConfigGenerator(NewLogger(), p)
			require.NoError(t, err)

			args := cg.BuildCommonPrometheusArgs()

			var (
				argsEnabled    bool
				featureEnabled bool
			)
			for _, arg := range args {
				switch arg.Name {
				case "web.enable-otlp-receiver":
					argsEnabled = true
				case "enable-feature":
					feats := strings.Split(arg.Value, ",")
					if slices.Contains(feats, "otlp-write-receiver") {
						featureEnabled = true
					}
				}
			}

			require.Equal(t, tc.expectedOTLPReceiverFlag, argsEnabled)
			require.Equal(t, tc.expectedOTLPFeatureEnabled, featureEnabled)
		})
	}
}

func TestNodeSelectorWithTopologyZone(t *testing.T) {
	topologyMode := monitoringv1.TopologyShardingStrategyMode

	for _, tc := range []struct {
		name                       string
		nodeSelector               map[string]string
		shardingStrategy           *monitoringv1.ShardingStrategy
		prometheusTopologySharding bool
		shardIndex                 int32
		expectedSelector           map[string]string
	}{
		{
			name:                       "prometheusTopologySharding=false returns original selector",
			nodeSelector:               map[string]string{"foo": "bar"},
			prometheusTopologySharding: false,
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			expectedSelector: map[string]string{"foo": "bar"},
		},
		{
			name:                       "shardingStrategy=nil returns original selector",
			nodeSelector:               map[string]string{"foo": "bar"},
			prometheusTopologySharding: true,
			shardingStrategy:           nil,
			expectedSelector:           map[string]string{"foo": "bar"},
		},
		{
			// This case isn't possible in practice because the API enforces
			// the "topology can only be defined when strategy = topology
			// sharding" invariant.
			name:         "Topology mode with no values returns original selector",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(monitoringv1.AddressShardingStrategyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			expectedSelector:           map[string]string{"foo": "bar"},
		},
		{
			// This case isn't possible in practice because the API enforces
			// the "shards >= number of zones" invariant.
			name:         "Topology mode with no values returns original selector",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{}},
			},
			prometheusTopologySharding: true,
			expectedSelector:           map[string]string{"foo": "bar"},
		},
		{
			name:         "Topology mode shard #0 with 2 zones assigns first zone",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedSelector:           map[string]string{"foo": "bar", corev1.LabelTopologyZone: "zone-a"},
		},
		{
			name:         "Topology mode shard #1 with 2 zones assigns second zone",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 1,
			expectedSelector:           map[string]string{"foo": "bar", corev1.LabelTopologyZone: "zone-b"},
		},
		{
			name:         "Topology mode shard #2 with 2 zones assigns first zone",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 2,
			expectedSelector:           map[string]string{"foo": "bar", corev1.LabelTopologyZone: "zone-a"},
		},
		{
			name:         "Topology mode shard #3 with 2 zones assigns second zone",
			nodeSelector: map[string]string{"foo": "bar"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 3,
			expectedSelector:           map[string]string{"foo": "bar", corev1.LabelTopologyZone: "zone-b"},
		},
		{
			name:         "Topology mode overrides existing topology.kubernetes.io/zone",
			nodeSelector: map[string]string{"topology.kubernetes.io/zone": "will-be-replaced"},
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedSelector:           map[string]string{corev1.LabelTopologyZone: "zone-a"},
		},
		{
			name:         "Topology mode with nil nodeSelector creates new map",
			nodeSelector: nil,
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedSelector:           map[string]string{corev1.LabelTopologyZone: "zone-a"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						NodeSelector:     tc.nodeSelector,
						ShardingStrategy: tc.shardingStrategy,
					},
				},
			}
			opts := []ConfigGeneratorOption{}
			if tc.prometheusTopologySharding {
				opts = append(opts, WithPrometheusTopologySharding())
			}
			cg, err := NewConfigGenerator(nil, p, opts...)
			require.NoError(t, err)
			got := cg.NodeSelectorWithTopologyZone(tc.shardIndex)
			require.Equal(t, tc.expectedSelector, got)
		})
	}
}

func TestLabelSelectorForStatefulSets(t *testing.T) {
	for _, tc := range []struct {
		mode string
		exp  string
	}{
		{
			mode: "server",
			exp:  "managed-by in (prometheus-operator),operator.prometheus.io/shard,operator.prometheus.io/name,operator.prometheus.io/mode in (server)",
		},
		{
			mode: "agent",
			exp:  "managed-by in (prometheus-operator),operator.prometheus.io/shard,operator.prometheus.io/name,operator.prometheus.io/mode in (agent)",
		},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			ls := LabelSelectorForStatefulSets(tc.mode)
			require.Equal(t, tc.exp, ls)

			_, err := labels.Parse(ls)
			require.NoError(t, err)
		})
	}
}

func TestTopologyZoneForShard(t *testing.T) {
	topologyMode := monitoringv1.TopologyShardingStrategyMode
	addressMode := monitoringv1.AddressShardingStrategyMode

	for _, tc := range []struct {
		name                       string
		shardingStrategy           *monitoringv1.ShardingStrategy
		prometheusTopologySharding bool
		shardIndex                 int32
		expectedZone               string
	}{
		{
			name:                       "prometheusTopologySharding=false returns empty",
			prometheusTopologySharding: false,
			expectedZone:               "",
		},
		{
			name: "address mode returns empty",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode: ptr.To(addressMode),
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedZone:               "",
		},
		{
			name: "topology mode with no values returns empty",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedZone:               "",
		},
		{
			name: "shard 0 gets first zone",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 0,
			expectedZone:               "zone-a",
		},
		{
			name: "shard 1 gets second zone",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 1,
			expectedZone:               "zone-b",
		},
		{
			name: "shard 2 wraps around to first zone",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			shardIndex:                 2,
			expectedZone:               "zone-a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ShardingStrategy: tc.shardingStrategy,
					},
				},
			}
			opts := []ConfigGeneratorOption{}
			if tc.prometheusTopologySharding {
				opts = append(opts, WithPrometheusTopologySharding())
			}
			cg, err := NewConfigGenerator(nil, p, opts...)
			require.NoError(t, err)
			require.Equal(t, tc.expectedZone, cg.TopologyZoneForShard(tc.shardIndex))
		})
	}
}

func TestInzoneShardForShard(t *testing.T) {
	topologyMode := monitoringv1.TopologyShardingStrategyMode

	for _, tc := range []struct {
		name                       string
		shardingStrategy           *monitoringv1.ShardingStrategy
		prometheusTopologySharding bool
		// expected[i] is the expected inzone shard for shard index i.
		expected []int32
	}{
		{
			name:                       "topology not active returns shard index",
			prometheusTopologySharding: false,
			expected:                   []int32{0, 1, 2, 3},
		},
		{
			name: "2 shards 2 zones",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			expected:                   []int32{0, 0},
		},
		{
			name: "2 shards 1 zone",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a"}},
			},
			prometheusTopologySharding: true,
			expected:                   []int32{0, 1},
		},
		{
			name: "3 shards 2 zones",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			expected:                   []int32{0, 0, 1},
		},
		{
			name: "4 shards 2 zones",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b"}},
			},
			prometheusTopologySharding: true,
			expected:                   []int32{0, 0, 1, 1},
		},
		{
			name: "6 shards 3 zones",
			shardingStrategy: &monitoringv1.ShardingStrategy{
				Mode:     ptr.To(topologyMode),
				Topology: &monitoringv1.TopologyShardingStrategy{Values: []string{"zone-a", "zone-b", "zone-c"}},
			},
			prometheusTopologySharding: true,
			expected:                   []int32{0, 0, 0, 1, 1, 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ShardingStrategy: tc.shardingStrategy,
					},
				},
			}
			opts := []ConfigGeneratorOption{}
			if tc.prometheusTopologySharding {
				opts = append(opts, WithPrometheusTopologySharding())
			}
			cg, err := NewConfigGenerator(nil, p, opts...)
			require.NoError(t, err)

			for i, exp := range tc.expected {
				require.Equal(t, exp, cg.InzoneShardForShard(int32(i)))
			}
		})
	}
}
