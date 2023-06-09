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

import "github.com/prometheus/common/version"

const (
	// DefaultAlertmanagerVersion is a default image tag for the prometheus alertmanager
	DefaultAlertmanagerVersion = "v0.25.0"
	// DefaultAlertmanagerBaseImage is a base container registry address for the prometheus alertmanager
	DefaultAlertmanagerBaseImage = "quay.io/prometheus/alertmanager"
	// DefaultAlertmanagerImage is a default image pulling address for the prometheus alertmanager
	DefaultAlertmanagerImage = DefaultAlertmanagerBaseImage + ":" + DefaultAlertmanagerVersion

	// DefaultThanosVersion is a default image tag for the Thanos long-term prometheus storage collector
	DefaultThanosVersion = "v0.31.0"
	// DefaultThanosBaseImage is a base container registry address for the Thanos long-term prometheus
	// storage collector
	DefaultThanosBaseImage = "quay.io/thanos/thanos"
	// DefaultThanosImage is a default image pulling address for the Thanos long-term prometheus storage collector
	DefaultThanosImage = DefaultThanosBaseImage + ":" + DefaultThanosVersion
)

var (
	// DefaultPrometheusVersion is a default image tag for the prometheus
	DefaultPrometheusVersion = PrometheusCompatibilityMatrix[len(PrometheusCompatibilityMatrix)-1]
	// DefaultPrometheusBaseImage is a base container registry address for the prometheus
	DefaultPrometheusBaseImage = "quay.io/prometheus/prometheus"
	// DefaultPrometheusImage is a default image pulling address for the prometheus
	DefaultPrometheusImage = DefaultPrometheusBaseImage + ":" + DefaultPrometheusVersion

	// DefaultPrometheusConfigReloaderImage is an image that will be used as a sidecar to provide dynamic prometheus
	// configuration reloading
	DefaultPrometheusConfigReloaderImage = "quay.io/prometheus-operator/prometheus-config-reloader:v" + version.Version

	// PrometheusCompatibilityMatrix is a list of supported prometheus versions.
	// prometheus-operator end-to-end tests verify that the operator can deploy from LTS n-1 to the latest stable.
	// This list should be updated every time a new LTS is released.
	PrometheusCompatibilityMatrix = []string{
		"v2.37.0",
		"v2.37.1",
		"v2.37.2",
		"v2.37.3",
		"v2.37.4",
		"v2.37.5",
		"v2.37.6",
		"v2.37.7",
		"v2.37.8",
		"v2.38.0",
		"v2.39.0",
		"v2.39.1",
		"v2.39.2",
		"v2.40.0",
		"v2.40.1",
		"v2.40.2",
		"v2.40.3",
		"v2.40.4",
		"v2.40.5",
		"v2.40.6",
		"v2.40.7",
		"v2.41.0",
		"v2.42.0",
		"v2.43.0",
		"v2.43.1",
		"v2.44.0",
	}
)
