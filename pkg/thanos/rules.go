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

package thanos

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const labelThanosRulerName = "thanos-ruler-name"

func (o *Operator) createOrUpdateRuleConfigMaps(ctx context.Context, t *monitoringv1.ThanosRuler) ([]string, error) {
	logger := o.logger.With("thanos", t.Name, "namespace", t.Namespace)

	namespaces, err := operator.SelectNamespacesFromCache(t, t.Spec.RuleNamespaceSelector, o.nsRuleInf)
	if err != nil {
		return nil, err
	}
	logger.Debug("selected RuleNamespaces", "namespaces", strings.Join(namespaces, ","))

	excludedFromEnforcement := t.Spec.ExcludedFromEnforcement
	// append the deprecated PrometheusRulesExcludedFromEnforce
	for _, rule := range t.Spec.PrometheusRulesExcludedFromEnforce {
		excludedFromEnforcement = append(excludedFromEnforcement,
			monitoringv1.ObjectReference{
				Namespace: rule.RuleNamespace,
				Group:     monitoring.GroupName,
				Resource:  monitoringv1.PrometheusRuleName,
				Name:      rule.RuleName,
			})
	}

	var (
		nsLabeler     = namespacelabeler.New(t.Spec.EnforcedNamespaceLabel, excludedFromEnforcement, false)
		thanosVersion = operator.StringValOrDefault(ptr.Deref(t.Spec.Version, ""), operator.DefaultThanosVersion)
	)

	promRuleSelector, err := operator.NewPrometheusRuleSelector(
		operator.ThanosFormat,
		thanosVersion,
		t.Spec.RuleSelector,
		nsLabeler,
		o.ruleInfs,
		o.newEventRecorder(t),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("initializing PrometheusRules failed: %w", err)
	}

	rules, rejected, err := promRuleSelector.Select(namespaces)
	if err != nil {
		return nil, fmt.Errorf("selecting PrometheusRules failed: %w", err)
	}

	if tKey, ok := o.accessor.MetaNamespaceKey(t); ok {
		o.metrics.SetSelectedResources(tKey, monitoringv1.PrometheusRuleKind, len(rules))
		o.metrics.SetRejectedResources(tKey, monitoringv1.PrometheusRuleKind, rejected)
	}

	// Update the corresponding ConfigMap resources.
	prs := operator.NewPrometheusRuleSyncer(
		logger,
		o.kclient.CoreV1().ConfigMaps(t.Namespace),
		labels.Set{labelThanosRulerName: t.Name},
		[]operator.ObjectOption{
			operator.WithAnnotations(o.config.Annotations),
			operator.WithLabels(o.config.Labels),
			operator.WithManagingOwner(t),
			operator.WithName(fmt.Sprintf("thanos-ruler-%s", t.Name)),
		},
	)
	return prs.Sync(ctx, rules.ValidMarshalledResources())
}
