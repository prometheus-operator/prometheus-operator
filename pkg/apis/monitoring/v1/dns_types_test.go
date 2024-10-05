// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPodDNSConfigSerialization(t *testing.T) {
	dnsConfig := &PodDNSConfig{
		Nameservers: []string{"8.8.8.8", "8.8.4.4"},
		Searches:    []string{"default.svc.cluster.local", "custom.search"},
		Options: []PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: ptrTo("5"),
			},
			{
				Name:  "timeout",
				Value: ptrTo("1"),
			},
		},
	}

	data, err := json.Marshal(dnsConfig)
	require.NoError(t, err, "failed to serialize PodDNSConfig to JSON")

	var deserialized PodDNSConfig
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err, "failed to deserialize JSON to PodDNSConfig")

	require.Equal(t, dnsConfig, &deserialized, "deserialized object does not match the original")
}

func TestPodDNSConfigOption(t *testing.T) {
	option := PodDNSConfigOption{
		Name:  "rotate",
		Value: ptrTo("true"),
	}

	data, err := json.Marshal(option)
	require.NoError(t, err, "failed to serialize PodDNSConfigOption to JSON")

	var deserialized PodDNSConfigOption
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err, "failed to deserialize JSON to PodDNSConfigOption")

	require.Equal(t, option, deserialized, "deserialized object does not match the original")
}

// ptrTo is a helper function to get a pointer to a string value
func ptrTo(val string) *string {
	return &val
}
