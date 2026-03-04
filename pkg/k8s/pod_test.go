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

package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestConvertToK8sDNSConfig(t *testing.T) {
	monitoringDNSConfig := &monitoringv1.PodDNSConfig{
		Nameservers: []string{"8.8.8.8", "8.8.4.4"},
		Searches:    []string{"custom.search"},
		Options: []monitoringv1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: ptr.To("5"),
			},
			{
				Name:  "timeout",
				Value: ptr.To("1"),
			},
		},
	}

	var spec v1.PodSpec
	UpdateDNSConfig(&spec, monitoringDNSConfig)

	// Verify the conversion matches the original content
	require.Equal(t, monitoringDNSConfig.Nameservers, spec.DNSConfig.Nameservers, "expected nameservers to match")
	require.Equal(t, monitoringDNSConfig.Searches, spec.DNSConfig.Searches, "expected searches to match")

	// Check if DNSConfig options match
	require.Equal(t, len(monitoringDNSConfig.Options), len(spec.DNSConfig.Options), "expected options length to match")
	for i, opt := range monitoringDNSConfig.Options {
		require.Equal(t, opt.Name, spec.DNSConfig.Options[i].Name, "expected option names to match")
		require.Equal(t, opt.Value, spec.DNSConfig.Options[i].Value, "expected option values to match")
	}
}
