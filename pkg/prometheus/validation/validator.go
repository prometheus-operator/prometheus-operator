// Copyright 2025 The prometheus-operator Authors
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

package validation

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/prometheus/prometheus/model/relabel"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

// LabelConfigValidator validates relabel configurations.
type LabelConfigValidator struct {
	v semver.Version
}

// NewLabelConfigValidator creates a new LabelConfigValidator.
func NewLabelConfigValidator(p monitoringv1.PrometheusInterface) (*LabelConfigValidator, error) {
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)
	v, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	return NewLabelConfigValidatorFromVersion(v)
}

// NewLabelConfigValidatorFromVersion creates a new LabelConfigValidator from a semver.Version.
func NewLabelConfigValidatorFromVersion(v semver.Version) (*LabelConfigValidator, error) {
	return &LabelConfigValidator{
		v: v,
	}, nil
}

// Validate validates a list of relabel configurations.
func (lcv *LabelConfigValidator) Validate(rcs []monitoringv1.RelabelConfig) error {
	for i, rc := range rcs {
		if err := lcv.ValidateRelabelConfig(rc); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

// From https://github.com/prometheus/prometheus/blob/747c5ee2b19a9e6a51acfafae9fa2c77e224803d/model/relabel/relabel.go#L378-L380
func varInRegexTemplate(template string) bool {
	return strings.Contains(template, "$")
}

func (lcv *LabelConfigValidator) isValidLabelName(labelName string) bool {
	validationScheme := operator.ValidationSchemeForPrometheus(lcv.v)
	return validationScheme.IsValidLabelName(labelName)
}

// ValidateRelabelConfig validates a single relabel configuration.
func (lcv *LabelConfigValidator) ValidateRelabelConfig(rc monitoringv1.RelabelConfig) error {
	minimumVersionCaseActions := lcv.v.GTE(semver.MustParse("2.36.0"))
	minimumVersionEqualActions := lcv.v.GTE(semver.MustParse("2.41.0"))
	if rc.Action == "" {
		rc.Action = string(relabel.Replace)
	}
	action := strings.ToLower(rc.Action)

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase)) && !minimumVersionCaseActions {
		return fmt.Errorf("%s relabel action is only supported from Prometheus version 2.36.0", rc.Action)
	}

	if (action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && !minimumVersionEqualActions {
		return fmt.Errorf("%s relabel action is only supported from Prometheus version 2.41.0", rc.Action)
	}

	if _, err := relabel.NewRegexp(rc.Regex); err != nil {
		return fmt.Errorf("invalid regex %s for relabel configuration: %w", rc.Regex, err)
	}

	if rc.Modulus == 0 && action == string(relabel.HashMod) {
		return fmt.Errorf("relabel configuration for hashmod requires non-zero modulus")
	}

	if (action == string(relabel.Replace) || action == string(relabel.HashMod) || action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && rc.TargetLabel == "" {
		return fmt.Errorf("relabel configuration for %s action needs targetLabel value", rc.Action)
	}

	if (action == string(relabel.Replace)) && !varInRegexTemplate(rc.TargetLabel) && !lcv.isValidLabelName(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (action == string(relabel.Replace)) && varInRegexTemplate(rc.TargetLabel) && !lcv.isValidLabelName(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && !lcv.isValidLabelName(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && (rc.Replacement != nil && *rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
		return fmt.Errorf("'replacement' can not be set for %s action", rc.Action)
	}

	if action == string(relabel.LabelMap) && (rc.Replacement != nil) && !lcv.isValidLabelName(*rc.Replacement) {
		return fmt.Errorf("%q is invalid 'replacement' for %s action", *rc.Replacement, rc.Action)
	}

	if action == string(relabel.HashMod) && !lcv.isValidLabelName(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if action == string(relabel.KeepEqual) || action == string(relabel.DropEqual) {
		if (rc.Regex != "" && rc.Regex != relabel.DefaultRelabelConfig.Regex.String()) ||
			(rc.Modulus != uint64(0) &&
				rc.Modulus != relabel.DefaultRelabelConfig.Modulus) ||
			(rc.Separator != nil &&
				*rc.Separator != relabel.DefaultRelabelConfig.Separator) ||
			(rc.Replacement != nil && *rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
			return fmt.Errorf("%s action requires only 'source_labels' and `target_label`, and no other fields", rc.Action)
		}
	}

	if action == string(relabel.LabelDrop) || action == string(relabel.LabelKeep) {
		if len(rc.SourceLabels) != 0 ||
			(rc.TargetLabel != "" &&
				rc.TargetLabel != relabel.DefaultRelabelConfig.TargetLabel) ||
			(rc.Modulus != uint64(0) &&
				rc.Modulus != relabel.DefaultRelabelConfig.Modulus) ||
			(rc.Separator != nil &&
				*rc.Separator != relabel.DefaultRelabelConfig.Separator) ||
			(rc.Replacement != nil &&
				*rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
			return fmt.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
		}
	}
	return nil
}
