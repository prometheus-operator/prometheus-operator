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

	k8sflag "k8s.io/component-base/cli/flag"
)

// At the moment, the are no feature gates available.
var defaultFeatureGates = map[string]bool{}

// ValidateFeatureGates merges the feature gate default values with
// the values provided by the user.
func ValidateFeatureGates(flags *k8sflag.MapStringBool) (string, error) {
	gates := defaultFeatureGates
	if flags.Empty() {
		return mapToString(gates), nil
	}

	imgs := *flags.Map
	for k, v := range imgs {
		if _, ok := gates[k]; !ok {
			return "", fmt.Errorf("feature gate %v is unknown", k)
		}
		gates[k] = v
	}
	return mapToString(gates), nil
}

func AvailableFeatureGates() []string {
	i := 0
	gates := make([]string, len(defaultFeatureGates))
	for k := range defaultFeatureGates {
		gates[i] = k
		i++
	}
	slices.Sort(gates)
	return gates
}

func mapToString(m map[string]bool) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%t", k, v))
	}
	return strings.Join(s, ",")
}

func HasFeatureGate(feat string) bool {
	return defaultFeatureGates[feat]
}
