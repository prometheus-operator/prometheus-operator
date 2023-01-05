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
	"reflect"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	thanostypes "github.com/thanos-io/thanos/pkg/store/storepb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RuleConfigurationFormat int

const (
	PrometheusFormat RuleConfigurationFormat = iota
	ThanosFormat
)

type PrometheusRuleSelector struct {
	ruleFormat   RuleConfigurationFormat
	ruleSelector labels.Selector
	nsLabeler    *namespacelabeler.Labeler
	ruleInformer *informers.ForResource
	logger       log.Logger
}

func NewPrometheusRuleSelector(ruleFormat RuleConfigurationFormat, labelSelector *metav1.LabelSelector, nsLabeler *namespacelabeler.Labeler, ruleInformer *informers.ForResource, logger log.Logger) (*PrometheusRuleSelector, error) {
	ruleSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "convert rule label selector to selector")
	}

	return &PrometheusRuleSelector{
		ruleFormat:   ruleFormat,
		ruleSelector: ruleSelector,
		nsLabeler:    nsLabeler,
		ruleInformer: ruleInformer,
		logger:       logger,
	}, nil
}

func generateRulesConfiguration(ruleformat RuleConfigurationFormat, promRule monitoringv1.PrometheusRuleSpec, logger log.Logger) (string, error) {
	if ruleformat == PrometheusFormat {
		// Unset partialResponseStrategy field.
		for i := range promRule.Groups {
			promRule.Groups[i].PartialResponseStrategy = ""
		}
	}

	content, err := yaml.Marshal(promRule)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal content")
	}
	rValidationErrors := ValidateRule(promRule)
	if len(rValidationErrors) != 0 {
		const m = "Invalid rule"
		level.Debug(logger).Log("msg", m, "content", content)
		for _, rValidationError := range rValidationErrors {
			level.Info(logger).Log("msg", m, "field", rValidationError.Field, "err", rValidationError.Error)
		}
		return "", errors.New(m)
	}
	return string(content), nil
}

// ValidateRule takes PrometheusRuleSpec and validates it using the upstream prometheus rule validator.
func ValidateRule(promRule monitoringv1.PrometheusRuleSpec) []RuleValidationError {
	groupTag := GetJSONTagForField(monitoringv1.PrometheusRuleSpec{}, "Groups")
	groupMap := make(map[string]int, len(promRule.Groups))

	for i, group := range promRule.Groups {
		groupMap[group.Name] = i
		if group.PartialResponseStrategy == "" {
			continue
		}
		// TODO(slashpai): Remove this validation after v0.65 since this is handled at CRD level
		if _, ok := thanostypes.PartialResponseStrategy_value[strings.ToUpper(group.PartialResponseStrategy)]; !ok {
			return []RuleValidationError{{
				Field: fmt.Sprintf("%s[%d].%s", groupTag, i, GetJSONTagForField(monitoringv1.RuleGroup{}, "PartialResponseStrategy")),
				Error: fmt.Errorf("invalid partial_response_strategy %s value", group.PartialResponseStrategy),
			}}
		}

		// reset this as the upstream prometheus rule validator
		// is not aware of the partial_response_strategy field.
		promRule.Groups[i].PartialResponseStrategy = ""
	}

	content, err := yaml.Marshal(promRule)
	if err != nil {
		return []RuleValidationError{{Error: errors.Wrap(err, "failed to marshal content")}}
	}
	rValidationErrors := []RuleValidationError{}
	_, errs := rulefmt.Parse(content)

	for _, err := range errs {
		switch promErr := err.(type) {
		case *rulefmt.Error:
			rValidationError := RuleValidationError{
				// Subtract 1 to promErr.Rule because rulefmt.Parse adds 1
				// More info on https://github.com/prometheus/prometheus/pull/7495
				Field: fmt.Sprintf("%s[%d].%s[%d]", groupTag, groupMap[promErr.Group], GetJSONTagForField(monitoringv1.RuleGroup{}, "Rules"), promErr.Rule-1),
				Error: err,
			}
			rValidationErrors = append(rValidationErrors, rValidationError)
		default:
			rValidationError := RuleValidationError{
				Field: groupTag,
				Error: err,
			}
			rValidationErrors = append(rValidationErrors, rValidationError)
		}
	}

	return rValidationErrors
}

func GetJSONTagForField(i interface{}, fieldName string) string {
	if rRule, found := reflect.ValueOf(i).Type().FieldByName(fieldName); found {
		return strings.Split(rRule.Tag.Get("json"), ",")[0]
	}
	return ""
}

// Select selects PrometheusRules and translates them into native Prometheus/Thanos configurations.
// The second returned value is the number of rejected PrometheusRule objects.
func (pr *PrometheusRuleSelector) Select(namespaces []string) (map[string]string, int, error) {
	promRules := map[string]*monitoringv1.PrometheusRule{}

	for _, ns := range namespaces {
		err := pr.ruleInformer.ListAllByNamespace(ns, pr.ruleSelector, func(obj interface{}) {
			promRule := obj.(*monitoringv1.PrometheusRule).DeepCopy()
			if err := k8sutil.AddTypeInformationToObject(promRule); err != nil {
				level.Error(pr.logger).Log("msg", "failed to set rule type information", "namespace", ns, "err", err)
				return
			}

			promRules[fmt.Sprintf("%v-%v-%v.yaml", promRule.Namespace, promRule.Name, promRule.UID)] = promRule
		})
		if err != nil {
			return nil, 0, errors.Wrapf(err, "failed to list prometheus rules in namespace %s", ns)
		}
	}

	var rejected int
	rules := make(map[string]string, len(promRules))

	for ruleName, promRule := range promRules {
		var err error
		var content string
		if err := pr.nsLabeler.EnforceNamespaceLabel(promRule); err != nil {
			continue
		}

		content, err = generateRulesConfiguration(pr.ruleFormat, promRule.Spec, pr.logger)
		if err != nil {
			rejected++
			level.Warn(pr.logger).Log(
				"msg", "skipping prometheusrule",
				"error", err.Error(),
				"prometheusrule", promRule.Name,
				"namespace", promRule.Namespace,
			)
			continue
		}

		rules[ruleName] = content
	}

	ruleNames := []string{}
	for name := range rules {
		ruleNames = append(ruleNames, name)
	}

	level.Debug(pr.logger).Log(
		"msg", "selected Rules",
		"rules", strings.Join(ruleNames, ","),
	)

	return rules, rejected, nil
}
