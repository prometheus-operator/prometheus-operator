// Copyright 2021 The prometheus-operator Authors
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

package v1

import (
	"fmt"
	"regexp"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var durationRe = regexp.MustCompile(`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`)

// ValidateAlertmanagerConfig checks that the given resource complies with the
// semantics of the Alertmanager configuration.
// In particular, it verifies things that can't be modelized with the OpenAPI
// specification such as routes should refer to an existing receiver.
func ValidateAlertmanager(am *monitoringv1.Alertmanager) error {
	// Validate Global Config
	if err := validateGlobalConfig(*am.Spec.AlertmanagerConfiguration.Global); err != nil {
		return err
	}

	return nil
}

func validateGlobalConfig(gc monitoringv1.AlertmanagerGlobalConfig) error {

	if err := gc.HTTPConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate global 'httpConfig'")
	}

	return nil
}
