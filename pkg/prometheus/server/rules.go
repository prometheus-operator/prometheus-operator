// Copyright 2016 The prometheus-operator Authors
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

package prometheus

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

func (c *Operator) selectPrometheusRules(p *monitoringv1.Prometheus, logger *slog.Logger) (operator.PrometheusRuleSelection, error) {
	namespaces, err := operator.SelectNamespacesFromCache(p, p.Spec.RuleNamespaceSelector, c.nsMonInf)
	var rules operator.PrometheusRuleSelection
	if err != nil {
		return rules, err
	}
	logger.Debug("selected RuleNamespaces", "namespaces", strings.Join(namespaces, ","))

	excludedFromEnforcement := p.Spec.ExcludedFromEnforcement
	// append the deprecated PrometheusRulesExcludedFromEnforce
	for _, rule := range p.Spec.PrometheusRulesExcludedFromEnforce {
		excludedFromEnforcement = append(excludedFromEnforcement,
			monitoringv1.ObjectReference{
				Namespace: rule.RuleNamespace,
				Group:     monitoring.GroupName,
				Resource:  monitoringv1.PrometheusRuleName,
				Name:      rule.RuleName,
			})
	}

	var (
		nsLabeler   = namespacelabeler.New(p.Spec.EnforcedNamespaceLabel, excludedFromEnforcement, true)
		promVersion = operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)
	)

	// Select and filter PrometheusRule resources.
	promRuleSelector, err := operator.NewPrometheusRuleSelector(
		operator.PrometheusFormat,
		promVersion,
		p.Spec.RuleSelector,
		nsLabeler,
		c.ruleInfs,
		c.newEventRecorder(p),
		logger,
	)
	if err != nil {
		return rules, fmt.Errorf("initializing PrometheusRules failed: %w", err)
	}

	rules, err = promRuleSelector.Select(namespaces)
	if err != nil {
		return rules, fmt.Errorf("selecting PrometheusRules failed: %w", err)
	}

	if pKey, ok := c.accessor.MetaNamespaceKey(p); ok {
		c.metrics.SetSelectedResources(pKey, monitoringv1.PrometheusRuleKind, rules.SelectedLen())
		c.metrics.SetRejectedResources(pKey, monitoringv1.PrometheusRuleKind, rules.RejectedLen())
	}
	return rules, nil
}

func (c *Operator) createOrUpdateRuleConfigMaps(ctx context.Context, p *monitoringv1.Prometheus, rules operator.PrometheusRuleSelection, logger *slog.Logger) ([]string, error) {

	// Update the corresponding ConfigMap resources.
	prs := operator.NewPrometheusRuleSyncer(
		logger,
		fmt.Sprintf("prometheus-%s", p.Name),
		c.kclient.CoreV1().ConfigMaps(p.Namespace),
		labels.Set{prompkg.LabelPrometheusName: p.Name},
		[]operator.ObjectOption{
			operator.WithAnnotations(c.config.Annotations),
			operator.WithLabels(c.config.Labels),
			operator.WithManagingOwner(p),
		},
	)

	configMapNames, err := prs.Sync(ctx, rules.RuleFiles())
	if err != nil {
		return nil, fmt.Errorf("synchronizing PrometheusRules failed: %w", err)
	}

	return prs.AppendConfigMapNames(configMapNames, 3), nil
}
