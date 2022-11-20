// Copyright 2022 The prometheus-operator Authors
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
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	thanostypes "github.com/thanos-io/thanos/pkg/store/storepb"
)

func GenerateRulesConfiguration(promRule monitoringv1.PrometheusRuleSpec, logger log.Logger) (string, error) {
	content, err := yaml.Marshal(promRule)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal content")
	}
	errs := ValidateRule(promRule)
	if len(errs) != 0 {
		const m = "Invalid rule"
		level.Debug(logger).Log("msg", m, "content", content)
		for _, err := range errs {
			level.Info(logger).Log("msg", m, "err", err)
		}
		return "", errors.New(m)
	}
	return string(content), nil
}

// ValidateRule takes PrometheusRuleSpec and validates it using the upstream prometheus rule validator.
func ValidateRule(promRule monitoringv1.PrometheusRuleSpec) []error {
	for i, group := range promRule.Groups {
		if group.PartialResponseStrategy == "" {
			continue
		}
		// TODO(slashpai): Remove this validation after v0.65 since this is handled at CRD level
		if _, ok := thanostypes.PartialResponseStrategy_value[strings.ToUpper(group.PartialResponseStrategy)]; !ok {
			return []error{
				fmt.Errorf("invalid partial_response_strategy %s value", group.PartialResponseStrategy),
			}
		}

		// reset this as the upstream prometheus rule validator
		// is not aware of the partial_response_strategy field.
		promRule.Groups[i].PartialResponseStrategy = ""
	}
	content, err := yaml.Marshal(promRule)
	if err != nil {
		return []error{errors.Wrap(err, "failed to marshal content")}
	}
	_, errs := rulefmt.Parse(content)
	return errs
}
