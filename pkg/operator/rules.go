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
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
)

func init() {
	// For now, the operator only supports legacy label names.
	// Eventually the operator should support UTF-8 label names too and the
	// issue is tracked by
	// https://github.com/prometheus-operator/prometheus-operator/issues/7362
	// nolint:staticcheck
	model.NameValidationScheme = model.LegacyValidation
}

type RuleConfigurationFormat int

const (
	PrometheusFormat RuleConfigurationFormat = iota
	ThanosFormat
)

// The maximum `Data` size of a ConfigMap seems to differ between
// environments. This is probably due to different meta data sizes which count
// into the overall maximum size of a ConfigMap. Thereby lets leave a
// large buffer.
var MaxConfigMapDataSize = int(float64(v1.MaxSecretSize) * 0.5)

type PrometheusRuleSelector struct {
	ruleFormat   RuleConfigurationFormat
	version      semver.Version
	ruleSelector labels.Selector
	nsLabeler    *namespacelabeler.Labeler
	ruleInformer *informers.ForResource

	eventRecorder record.EventRecorder

	logger *slog.Logger
}

func NewPrometheusRuleSelector(ruleFormat RuleConfigurationFormat, version string, labelSelector *metav1.LabelSelector, nsLabeler *namespacelabeler.Labeler, ruleInformer *informers.ForResource, eventRecorder record.EventRecorder, logger *slog.Logger) (*PrometheusRuleSelector, error) {
	componentVersion, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	ruleSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("convert rule label selector to selector: %w", err)
	}

	return &PrometheusRuleSelector{
		ruleFormat:    ruleFormat,
		version:       componentVersion,
		ruleSelector:  ruleSelector,
		nsLabeler:     nsLabeler,
		ruleInformer:  ruleInformer,
		eventRecorder: eventRecorder,
		logger:        logger,
	}, nil
}

func (prs *PrometheusRuleSelector) generateRulesConfiguration(promRule *monitoringv1.PrometheusRule) (string, error) {
	logger := prs.logger.With("prometheusrule", promRule.Name, "prometheusrule-namespace", promRule.Namespace)
	promRuleSpec := promRule.Spec

	promRuleSpec = prs.sanitizePrometheusRulesSpec(promRuleSpec, logger)

	content, err := yaml.Marshal(promRuleSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	errs := ValidateRule(promRuleSpec)
	if len(errs) != 0 {
		const m = "invalid rule"
		logger.Debug(m, "content", content)
		for _, err := range errs {
			logger.Info(m, "err", err)
		}
		return "", errors.New(m)
	}

	return string(content), nil
}

// sanitizePrometheusRulesSpec sanitizes the PrometheusRules spec depending on the Prometheus/Thanos version.
func (prs *PrometheusRuleSelector) sanitizePrometheusRulesSpec(promRuleSpec monitoringv1.PrometheusRuleSpec, logger *slog.Logger) monitoringv1.PrometheusRuleSpec {
	minVersionKeepFiringFor := semver.MustParse("2.42.0")
	minVersionLimits := semver.MustParse("2.31.0")
	minVersionQueryOffset := semver.MustParse("2.53.0")
	minVersionRuleGroupLabels := semver.MustParse("3.0.0")
	component := "Prometheus"

	if prs.ruleFormat == ThanosFormat {
		minVersionKeepFiringFor = semver.MustParse("0.34.0")
		minVersionLimits = semver.MustParse("0.24.0")
		minVersionQueryOffset = semver.MustParse("100.0.0")     // Arbitrary very high major version because it's not yet supported by Thanos.
		minVersionRuleGroupLabels = semver.MustParse("100.0.0") // Arbitrary very high major version because it's not yet supported by Thanos.
		component = "Thanos"
	}

	for i := range promRuleSpec.Groups {
		if promRuleSpec.Groups[i].Limit != nil && prs.version.LT(minVersionLimits) {
			promRuleSpec.Groups[i].Limit = nil
			logger.Warn(fmt.Sprintf("ignoring `limit` not supported by %s", component), "minimum_version", minVersionLimits)
		}

		if promRuleSpec.Groups[i].QueryOffset != nil && prs.version.LT(minVersionQueryOffset) {
			promRuleSpec.Groups[i].QueryOffset = nil
			logger.Warn(fmt.Sprintf("ignoring `query_offset` not supported by %s", component), "minimum_version", minVersionQueryOffset)
		}

		if prs.ruleFormat == PrometheusFormat {
			// Unset partialResponseStrategy field.
			promRuleSpec.Groups[i].PartialResponseStrategy = ""
		}

		if len(promRuleSpec.Groups[i].Labels) > 0 && prs.version.LT(minVersionRuleGroupLabels) {
			promRuleSpec.Groups[i].Labels = nil
			logger.Warn(fmt.Sprintf("ignoring group labels since not supported by %s", component), "minimum_version", minVersionRuleGroupLabels)
		}

		for j := range promRuleSpec.Groups[i].Rules {
			if promRuleSpec.Groups[i].Rules[j].KeepFiringFor != nil && prs.version.LT(minVersionKeepFiringFor) {
				promRuleSpec.Groups[i].Rules[j].KeepFiringFor = nil
				logger.Warn(fmt.Sprintf("ignoring 'keep_firing_for' not supported by %s", component), "minimum_version", minVersionKeepFiringFor)
			}
		}
	}

	return promRuleSpec
}

// ValidateRule takes PrometheusRuleSpec and validates it using the upstream prometheus rule validator.
func ValidateRule(promRuleSpec monitoringv1.PrometheusRuleSpec) []error {
	for i := range promRuleSpec.Groups {
		// The upstream Prometheus rule validator doesn't support the
		// partial_response_strategy field.
		promRuleSpec.Groups[i].PartialResponseStrategy = ""

		// Empty durations need to be translated to nil to be omitted from the
		// YAML output otherwise the generated configuration will not be valid.
		if promRuleSpec.Groups[i].Interval != nil && *promRuleSpec.Groups[i].Interval == "" {
			promRuleSpec.Groups[i].Interval = nil
		}

		for j := range promRuleSpec.Groups[i].Rules {
			if ptr.Deref(promRuleSpec.Groups[i].Rules[j].For, "") == "" {
				promRuleSpec.Groups[i].Rules[j].For = nil
			}
		}
	}

	content, err := yaml.Marshal(promRuleSpec)
	if err != nil {
		return []error{fmt.Errorf("failed to marshal content: %w", err)}
	}

	// Check if the serialized rules exceed our internal limit.
	promRuleSize := len(content)
	if promRuleSize > MaxConfigMapDataSize {
		return []error{fmt.Errorf("the length of rendered Prometheus Rule is %d bytes which is above the maximum limit of %d bytes", promRuleSize, MaxConfigMapDataSize)}
	}

	_, errs := rulefmt.Parse(content, false)
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
				prs.logger.Error("failed to set rule type information", "namespace", ns, "err", err)
				return
			}

			promRules[fmt.Sprintf("%v-%v-%v.yaml", promRule.Namespace, promRule.Name, promRule.UID)] = promRule
		})
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list prometheus rules in namespace %s: %w", ns, err)
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
			prs.logger.Warn(
				"skipping prometheusrule",
				"error", err.Error(),
				"prometheusrule", promRule.Name,
				"namespace", promRule.Namespace,
			)
			prs.eventRecorder.Eventf(promRule, v1.EventTypeWarning, "InvalidConfiguration", "PrometheusRule %s was rejected due to invalid configuration: %v", promRule.Name, err)
			continue
		}

		rules[ruleName] = content
	}

	ruleNames := []string{}
	for name := range rules {
		ruleNames = append(ruleNames, name)
	}

	prs.logger.Debug(
		"selected Rules",
		"rules", strings.Join(ruleNames, ","),
	)

	return rules, rejected, nil
}
