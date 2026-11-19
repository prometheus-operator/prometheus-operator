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
	"slices"

	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
)

// ValidationSchemeForPrometheus returns the appropriate validation scheme based on Prometheus version.
func ValidationSchemeForPrometheus(version semver.Version) model.ValidationScheme {
	if version.GTE(semver.MustParse("3.0.0")) {
		return model.UTF8Validation
	}
	return model.LegacyValidation
}

// ValidationSchemeForThanos returns the appropriate validation scheme based on Thanos version.
func ValidationSchemeForThanos(version semver.Version) model.ValidationScheme {
	if version.GTE(semver.MustParse("0.38.0")) {
		return model.UTF8Validation
	}
	return model.LegacyValidation
}

// thanosPrometheusEquivalence associates a Thanos version with its
// equivalent Prometheus version.
type thanosPrometheusEquivalence struct {
	thanos     semver.Version
	prometheus semver.Version
}

// thanosPrometheusEquivalents maps Thanos versions to the equivalent
// Prometheus version, sorted by descending Thanos version.
var thanosPrometheusEquivalents = func() []thanosPrometheusEquivalence {
	versions := []thanosPrometheusEquivalence{
		{semver.MustParse("0.40.0"), semver.MustParse("3.5.1")},
		{semver.MustParse("0.38.0"), semver.MustParse("3.1.0")},
		{semver.MustParse("0.37.0"), semver.MustParse("2.55.1")},
		{semver.MustParse("0.36.0"), semver.MustParse("2.52.2")},
		{semver.MustParse("0.32.0"), semver.MustParse("2.48.0")},
		{semver.MustParse("0.31.0"), semver.MustParse("2.42.0")},
		{semver.MustParse("0.30.0"), semver.MustParse("2.40.7")},
		{semver.MustParse("0.28.0"), semver.MustParse("2.38.0")},
		{semver.MustParse("0.24.0"), semver.MustParse("2.32.0")},
	}

	slices.SortFunc(versions, func(a, b thanosPrometheusEquivalence) int {
		return b.thanos.Compare(a.thanos)
	})

	return versions
}()

// PrometheusVersionForThanos returns the Prometheus version equivalent to the
// given Thanos version.
func PrometheusVersionForThanos(version semver.Version) semver.Version {
	for _, e := range thanosPrometheusEquivalents {
		if version.GTE(e.thanos) {
			return e.prometheus
		}
	}

	return thanosPrometheusEquivalents[len(thanosPrometheusEquivalents)-1].prometheus
}
