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
	DefaultAlertmanagerVersion = "v0.24.0"
	// DefaultAlertmanagerBaseImage is a base container registry address for the prometheus alertmanager
	DefaultAlertmanagerBaseImage = "quay.io/prometheus/alertmanager"
	// DefaultAlertmanagerImage is a default image pulling address for the prometheus alertmanager
	DefaultAlertmanagerImage = DefaultAlertmanagerBaseImage + ":" + DefaultAlertmanagerVersion

	// DefaultThanosVersion is a default image tag for the Thanos long-term prometheus storage collector
	DefaultThanosVersion = "v0.25.2"
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

	// PrometheusCompatibilityMatrix is a list of supported prometheus version
	PrometheusCompatibilityMatrix = []string{
		"v2.0.0",
		"v2.2.1",
		"v2.3.1",
		"v2.3.2",
		"v2.4.0",
		"v2.4.1",
		"v2.4.2",
		"v2.4.3",
		"v2.5.0",
		"v2.6.0",
		"v2.6.1",
		"v2.7.0",
		"v2.7.1",
		"v2.7.2",
		"v2.8.1",
		"v2.9.2",
		"v2.10.0",
		"v2.11.0",
		"v2.14.0",
		"v2.15.2",
		"v2.16.0",
		"v2.17.2",
		"v2.18.0",
		"v2.18.1",
		"v2.18.2",
		"v2.19.0",
		"v2.19.1",
		"v2.19.2",
		"v2.19.3",
		"v2.20.0",
		"v2.20.1",
		"v2.21.0",
		"v2.22.0",
		"v2.22.1",
		"v2.22.2",
		"v2.23.0",
		"v2.24.0",
		"v2.24.1",
		"v2.25.0",
		"v2.25.1",
		"v2.25.2",
		"v2.26.0",
		"v2.26.1",
		"v2.27.0",
		"v2.27.1",
		"v2.28.0",
		"v2.28.1",
		"v2.29.0",
		"v2.29.1",
		"v2.30.0",
		"v2.30.1",
		"v2.30.2",
		"v2.30.3",
		"v2.31.0",
		"v2.31.1",
		"v2.32.0",
		"v2.32.1",
		"v2.33.0",
		"v2.33.1",
		"v2.33.2",
		"v2.33.3",
		"v2.33.4",
		"v2.33.5",
		"v2.34.0",
	}
)
