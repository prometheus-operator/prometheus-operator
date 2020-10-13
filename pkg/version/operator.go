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

// Holds the operator related version functions.
package version

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// NewCollector returns a collector that exports metrics about current version
// information.
func NewCollector(program string) prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: program,
			Name:      "build_info",
			Help: fmt.Sprintf(
				"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
				program,
			),
			ConstLabels: prometheus.Labels{
				"version":   Version,
				"revision":  Revision,
				"branch":    Branch,
				"goversion": GoVersion,
			},
		},
		func() float64 { return 1 },
	)
}

// Info returns version, branch and revision information.
func Info() string {
	return fmt.Sprintf("(version=%s, branch=%s, revision=%s)", Version, Branch, Revision)
}

// BuildContext returns goVersion, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(go=%s, user=%s, date=%s)", GoVersion, BuildUser, BuildDate)
}
