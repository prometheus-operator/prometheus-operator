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
	"github.com/prometheus-operator/prometheus-operator/pkg/version"
)

const (
	DefaultAlertmanagerVersion    = "v0.21.0"
	DefaultAlertmanagerBaseImage  = "quay.io/prometheus/alertmanager"
	DefaultAlertmanagerImage      = DefaultAlertmanagerBaseImage + ":" + DefaultAlertmanagerVersion
	DefaultThanosVersion          = "v0.14.0"
	DefaultThanosBaseImage        = "quay.io/thanos/thanos"
	DefaultThanosImage            = DefaultThanosBaseImage + ":" + DefaultThanosVersion
	DefaultConfigMapReloaderImage = "jimmidyson/configmap-reload:v0.4.0"
)

var (
	DefaultPrometheusConfigReloaderImage = "quay.io/prometheus-operator/prometheus-config-reloader:v" + version.Version

	PrometheusCompatibilityMatrix = []string{
		"v1.4.0",
		"v1.4.1",
		"v1.5.0",
		"v1.5.1",
		"v1.5.2",
		"v1.5.3",
		"v1.6.0",
		"v1.6.1",
		"v1.6.2",
		"v1.6.3",
		"v1.7.0",
		"v1.7.1",
		"v1.7.2",
		"v1.8.0",
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
		"v2.20.0",
	}
	DefaultPrometheusVersion   = PrometheusCompatibilityMatrix[len(PrometheusCompatibilityMatrix)-1]
	DefaultPrometheusBaseImage = "quay.io/prometheus/prometheus"
	DefaultPrometheusImage     = DefaultPrometheusBaseImage + ":" + DefaultPrometheusVersion
)
