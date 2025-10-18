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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func ValidateAlertmanager(am *monitoringv1.Alertmanager) error {
	if am.Spec.AlertmanagerConfiguration != nil {
		config := am.Spec.AlertmanagerConfiguration
		if config.Global != nil {
			if err := ValidateAlertmanagerGlobalConfig(config.Global); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateAlertmanagerGlobalConfig(gc *monitoringv1.AlertmanagerGlobalConfig) error {
	if gc == nil {
		return nil
	}

	if err := gc.HTTPConfig.Validate(); err != nil {
		return fmt.Errorf("httpConfig: %w", err)
	}

	return nil
}
