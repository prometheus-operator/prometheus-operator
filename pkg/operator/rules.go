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

	"github.com/blang/semver/v4"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RuleConfigurationFormat int

const (
	PrometheusFormat RuleConfigurationFormat = iota
	ThanosFormat
)

type PrometheusRuleSelector struct {
	ruleFormat   RuleConfigurationFormat
	version      semver.Version
	ruleSelector labels.Selector
	nsLabeler    *namespacelabeler.Labeler
	ruleInformer *informers.ForResource
	logger       log.Logger
}

func NewPrometheusRuleSelector(ruleFormat RuleConfigurationFormat, version string, labelSelector *metav1.LabelSelector, nsLabeler *namespacelabeler.Labeler, ruleInformer *informers.ForResource, logger log.Logger) (*PrometheusRuleSelector, error) {
	componentVersion, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	ruleSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "convert rule label selector to selector")
	}

	return &PrometheusRuleSelector{
		ruleFormat:   ruleFormat,
		version:      componentVersion,
		ruleSelector: ruleSelector,
		nsLabeler:    nsLabeler,
		ruleInformer: ruleInformer,
		logger:       logger,
	}, nil
}

func (prs *PrometheusRuleSelector) generateRulesConfiguration(promRule *monitoringv1.PrometheusRule) (string, error) {
	logger := log.With(prs.logger, "prometheusrule", promRule.Name, "prometheusrule-namespace", promRule.Namespace)
	promRuleSpec := promRule.Spec
	minVersionLimits := semver.MustParse("2.31.0")
	component := "Prometheus"

	if prs.ruleFormat == ThanosFormat {
		minVersionLimits = semver.MustParse("0.24.0")
		component = "Thanos"
	}

	for i := range promRuleSpec.Groups {
		if promRuleSpec.Groups[i].Limit != nil && prs.version.LT(minVersionLimits) {
			promRuleSpec.Groups[i].Limit = nil
			level.Warn(logger).Log("msg", fmt.Sprintf("`limit` is supported only from %s version %q", component, minVersionLimits))
		}
		if prs.ruleFormat == PrometheusFormat {
			// Unset partialResponseStrategy field.
			promRuleSpec.Groups[i].PartialResponseStrategy = ""
		}
	}

	content, err := yaml.Marshal(promRuleSpec)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal content")
	}
	errs := ValidateRule(promRuleSpec)
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
	for i := range promRule.Groups {
		// The upstream Prometheus rule validator doesn't support the
		// partial_response_strategy field.
		promRule.Groups[i].PartialResponseStrategy = ""
	}

	content, err := yaml.Marshal(promRule)
	if err != nil {
		return []error{errors.Wrap(err, "failed to marshal content")}
	}
	_, errs := rulefmt.Parse(content)
	return errs
}

// Select selects PrometheusRules and translates them into native Prometheus/Thanos configurations.
// The second returned value is the number of rejected PrometheusRule objects.
func (prs *PrometheusRuleSelector) Select(namespaces []string) (map[string]string, int, error) {
	promRules := map[string]*monitoringv1.PrometheusRule{}

	for _, ns := range namespaces {
		err := prs.ruleInformer.ListAllByNamespace(ns, prs.ruleSelector, func(obj interface{}) {
			promRule := obj.(*monitoringv1.PrometheusRule).DeepCopy()
			if err := k8sutil.AddTypeInformationToObject(promRule); err != nil {
				level.Error(prs.logger).Log("msg", "failed to set rule type information", "namespace", ns, "err", err)
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
		if err := prs.nsLabeler.EnforceNamespaceLabel(promRule); err != nil {
			continue
		}

		content, err = prs.generateRulesConfiguration(promRule)
		if err != nil {
			rejected++
			level.Warn(prs.logger).Log(
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

	level.Debug(prs.logger).Log(
		"msg", "selected Rules",
		"rules", strings.Join(ruleNames, ","),
	)

	return rules, rejected, nil
}
