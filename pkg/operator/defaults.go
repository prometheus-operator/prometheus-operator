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

package operator

import (
	"strings"

	"github.com/prometheus/common/version"
)

const (
	// DefaultAlertmanagerVersion is a default image tag for the prometheus alertmanager.
	DefaultAlertmanagerVersion = "v0.31.1"
	// DefaultAlertmanagerBaseImage is a base container registry address for the prometheus alertmanager.
	DefaultAlertmanagerBaseImage = "quay.io/prometheus/alertmanager"
	// DefaultAlertmanagerImage is a default image pulling address for the prometheus alertmanager.
	DefaultAlertmanagerImage = DefaultAlertmanagerBaseImage + ":" + DefaultAlertmanagerVersion

	// DefaultThanosVersion is a default image tag for the Thanos long-term prometheus storage collector.
	DefaultThanosVersion = "v0.41.0"
	// DefaultThanosBaseImage is a base container registry address for the Thanos long-term prometheus
	// storage collector.
	DefaultThanosBaseImage = "quay.io/thanos/thanos"
	// DefaultThanosImage is a default image pulling address for the Thanos long-term prometheus storage collector.
	DefaultThanosImage = DefaultThanosBaseImage + ":" + DefaultThanosVersion
)

var (
	// DefaultPrometheusVersion is a default image tag for the prometheus.
	DefaultPrometheusVersion = PrometheusCompatibilityMatrix[len(PrometheusCompatibilityMatrix)-1]
	// DefaultPrometheusV2 is latest version of Prometheus v2.
	DefaultPrometheusV2 = getLatestPrometheusV2()
	// DefaultPrometheusBaseImage is a base container registry address for the prometheus.
	DefaultPrometheusBaseImage = "quay.io/prometheus/prometheus"
	// DefaultPrometheusImage is a default image pulling address for the prometheus.
	DefaultPrometheusImage = DefaultPrometheusBaseImage + ":" + DefaultPrometheusVersion

	// DefaultPrometheusConfigReloaderImage is an image that will be used as a sidecar to provide dynamic prometheus
	// configuration reloading.
	DefaultPrometheusConfigReloaderImage = "quay.io/prometheus-operator/prometheus-config-reloader:v" + version.Version

	// PrometheusCompatibilityMatrix is a list of supported prometheus versions.
	// prometheus-operator end-to-end tests verify that the operator can deploy from the current LTS version to the latest stable release.
	// This list should be updated every time a new LTS is released.
	PrometheusCompatibilityMatrix = []string{
		"v2.45.0",
		"v2.46.0",
		"v2.47.0",
		"v2.47.1",
		"v2.47.2",
		"v2.48.0",
		"v2.48.1",
		"v2.49.0",
		"v2.49.1",
		"v2.50.0",
		"v2.50.1",
		"v2.51.0",
		"v2.51.1",
		"v2.51.2",
		"v2.52.0",
		// The v2.52.1 image tag is missing from docker.io and quay.io registries.
		"v2.53.0",
		"v2.53.1",
		"v2.53.2",
		"v2.53.3",
		"v2.54.0",
		"v2.54.1",
		"v2.55.0",
		"v2.55.1",
		"v3.0.0",
		"v3.0.1",
		"v3.1.0",
		"v3.2.0",
		"v3.2.1",
		"v3.3.0",
		"v3.3.1",
		"v3.4.0",
		"v3.4.1",
		"v3.4.2",
		"v3.5.0",
		"v3.6.0",
		"v3.7.0",
		"v3.7.1",
		"v3.7.2",
		"v3.7.3",
		"v3.8.0",
		"v3.8.1",
		"v3.9.0",
		"v3.9.1",
	}
)

func getLatestPrometheusV2() string {
	for i, version := range PrometheusCompatibilityMatrix {
		// Since last v2 version would be one just before the first v3 version
		if strings.HasPrefix(version, "v3") {
			return PrometheusCompatibilityMatrix[i-1]
		}
	}
	panic("failed to find a v2.x entry in the compatibility matrix")
}
