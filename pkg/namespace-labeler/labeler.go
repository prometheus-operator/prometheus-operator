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

package namespacelabeler

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"
	"github.com/prometheus-community/prom-label-proxy/injectproxy"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Labeler enables to enforce adding namespace labels to PrometheusRules and to metrics used in them
type Labeler struct {
	enforcedNsLabel       string
	prometheusRuleLabeler bool
	excludeList           map[string]map[string]struct{}
}

// New - creates new Labeler
// enforcedNsLabel - label name to be enforced for namespace
// excludeConfig - list of namespace + PrometheusRule names to be excluded while enforcing adding namespace label
func New(enforcedNsLabel string, excludeConfig []monitoringv1.PrometheusRuleExcludeConfig, prometheusRuleLabeler bool) *Labeler {

	if enforcedNsLabel == "" {
		return &Labeler{} // no-op labeler
	}

	if len(excludeConfig) == 0 {
		return &Labeler{enforcedNsLabel: enforcedNsLabel}
	}

	ruleExcludeList := make(map[string]map[string]struct{})

	for _, r := range excludeConfig {
		if r.RuleNamespace == "" || r.RuleName == "" {
			continue
		}
		if _, ok := ruleExcludeList[r.RuleNamespace]; !ok {
			ruleExcludeList[r.RuleNamespace] = make(map[string]struct{})
		}
		ruleExcludeList[r.RuleNamespace][r.RuleName] = struct{}{}
	}

	return &Labeler{
		excludeList:           ruleExcludeList,
		enforcedNsLabel:       enforcedNsLabel,
		prometheusRuleLabeler: prometheusRuleLabeler,
	}
}

func (l *Labeler) isExcludedRule(namespace, name string) bool {
	if l.excludeList == nil {
		return false
	}
	nsRules, ok := l.excludeList[namespace]
	if !ok {
		return false
	}
	_, ok = nsRules[name]
	return ok
}

// EnforceNamespaceLabel - adds(or modifies) namespace label to promRule labels with specified namespace
// and also adds namespace label to all the metrics used in promRule
func (l *Labeler) EnforceNamespaceLabel(rule *monitoringv1.PrometheusRule) error {

	if l.enforcedNsLabel == "" || l.isExcludedRule(rule.Namespace, rule.Name) {
		return nil
	}

	for gi, group := range rule.Spec.Groups {
		if l.prometheusRuleLabeler {
			group.PartialResponseStrategy = ""
		}
		for ri, r := range group.Rules {
			if len(rule.Spec.Groups[gi].Rules[ri].Labels) == 0 {
				rule.Spec.Groups[gi].Rules[ri].Labels = map[string]string{}
			}
			rule.Spec.Groups[gi].Rules[ri].Labels[l.enforcedNsLabel] = rule.Namespace

			expr := r.Expr.String()
			parsedExpr, err := parser.ParseExpr(expr)
			if err != nil {
				return errors.Wrap(err, "failed to parse promql expression")
			}
			enforcer := injectproxy.NewEnforcer(&labels.Matcher{
				Name:  l.enforcedNsLabel,
				Type:  labels.MatchEqual,
				Value: rule.Namespace,
			})
			err = enforcer.EnforceNode(parsedExpr)
			if err != nil {
				return errors.Wrap(err, "failed to inject labels to expression")
			}

			rule.Spec.Groups[gi].Rules[ri].Expr = intstr.FromString(parsedExpr.String())
		}
	}
	return nil
}
