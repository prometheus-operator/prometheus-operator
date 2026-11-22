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

package operator

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
)

func TestPrometheusVersionForThanos(t *testing.T) {
	for _, tc := range []struct {
		thanos     string
		prometheus string
	}{
		{thanos: "0.24.0", prometheus: "2.32.0"},
		{thanos: "0.27.0", prometheus: "2.32.0"},
		{thanos: "0.28.0", prometheus: "2.38.0"},
		{thanos: "0.30.0", prometheus: "2.40.7"},
		{thanos: "0.31.0", prometheus: "2.42.0"},
		{thanos: "0.32.0", prometheus: "2.48.0"},
		{thanos: "0.35.1", prometheus: "2.48.0"},
		{thanos: "0.36.0", prometheus: "2.52.2"},
		{thanos: "0.37.0", prometheus: "2.55.1"},
		{thanos: "0.37.9", prometheus: "2.55.1"},
		{thanos: "0.38.0", prometheus: "3.1.0"},
		{thanos: "0.39.2", prometheus: "3.1.0"},
		{thanos: "0.40.0", prometheus: "3.5.1"},
		{thanos: "0.42.5", prometheus: "3.5.1"},
		{thanos: "0.23.0", prometheus: "2.32.0"},
	} {
		t.Run(tc.thanos, func(t *testing.T) {
			require.Equal(
				t,
				semver.MustParse(tc.prometheus),
				PrometheusVersionForThanos(semver.MustParse(tc.thanos)),
			)
		})
	}
}

func TestThanosPrometheusEquivalentsOrder(t *testing.T) {
	for i := 1; i < len(thanosPrometheusEquivalents); i++ {
		require.Truef(
			t,
			thanosPrometheusEquivalents[i-1].thanos.GT(thanosPrometheusEquivalents[i].thanos),
			"thanosPrometheusEquivalents must be sorted by descending Thanos version, got %s before %s",
			thanosPrometheusEquivalents[i-1].thanos,
			thanosPrometheusEquivalents[i].thanos,
		)
	}
}
