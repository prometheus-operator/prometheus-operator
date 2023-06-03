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
	"os"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMakeRulesConfigMaps(t *testing.T) {
	t.Run("shouldAcceptRuleWithValidPartialResponseStrategyValue", shouldAcceptRuleWithValidPartialResponseStrategyValue)
	t.Run("shouldAcceptValidRule", shouldAcceptValidRule)
	t.Run("shouldRejectRuleWithInvalidLabels", shouldRejectRuleWithInvalidLabels)
	t.Run("shouldRejectRuleWithInvalidExpression", shouldRejectRuleWithInvalidExpression)
	t.Run("shouldResetRuleWithPartialResponseStrategySet", shouldResetRuleWithPartialResponseStrategySet)
	t.Run("shouldAcceptRuleWithLimitPrometheus", shouldAcceptRuleWithLimitPrometheus)
	t.Run("shouldAcceptRuleWithLimitThanos", shouldAcceptRuleWithLimitThanos)
	t.Run("shouldDropLimitFieldForUnsupportedPrometheusVersion", shouldDropLimitFieldForUnsupportedPrometheusVersion)
	t.Run("shouldDropLimitFieldForUnsupportedThanosVersion", shouldDropLimitFieldForUnsupportedThanosVersion)
}

func newRuleSelectorForConfigGeneration(ruleFormat RuleConfigurationFormat, version semver.Version) PrometheusRuleSelector {
	return PrometheusRuleSelector{
		ruleFormat: ruleFormat,
		version:    version,
		logger:     log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
	}
}

func shouldAcceptRuleWithValidPartialResponseStrategyValue(t *testing.T) {
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name:                    "group",
				PartialResponseStrategy: "warn",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
			},
		}},
	}

	thanosVersion, _ := semver.ParseTolerant(DefaultThanosVersion)
	pr := newRuleSelectorForConfigGeneration(ThanosFormat, thanosVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if !strings.Contains(content, "partial_response_strategy: warn") {
		t.Fatalf("expected `partial_response_strategy` to be set in PrometheusRule as `warn`")

	}
}

func shouldAcceptValidRule(t *testing.T) {
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
						Labels: map[string]string{
							"valid_label": "valid_value",
						},
					},
				},
			},
		}},
	}
	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	_, err := pr.generateRulesConfiguration(rules)
	if err != nil {
		t.Fatalf("expected no errors when parsing valid rule")
	}
}

func shouldRejectRuleWithInvalidLabels(t *testing.T) {
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
						Labels: map[string]string{
							"invalid/label": "value",
						},
					},
				},
			},
		}},
	}
	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	_, err := pr.generateRulesConfiguration(rules)
	if err == nil {
		t.Fatalf("expected errors when parsing rule with invalid labels")
	}
}

func shouldRejectRuleWithInvalidExpression(t *testing.T) {
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("invalidfn(1)"),
					},
				},
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	_, err := pr.generateRulesConfiguration(rules)
	if err == nil {
		t.Fatalf("expected errors when parsing rule with invalid expression")
	}
}

func shouldResetRuleWithPartialResponseStrategySet(t *testing.T) {
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name:                    "group",
				PartialResponseStrategy: "warn",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
			},
		}},
	}
	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if strings.Contains(content, "partial_response_strategy") {
		t.Fatalf("expected `partial_response_strategy` removed from PrometheusRule")
	}
}

func shouldAcceptRuleWithLimitPrometheus(t *testing.T) {
	limit := 50
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
				Limit: &limit,
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if !strings.Contains(content, "limit") {
		t.Fatalf("expected `limit` to be present in PrometheusRule")
	}
}

func shouldAcceptRuleWithLimitThanos(t *testing.T) {
	limit := 50
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
				Limit: &limit,
			},
		}},
	}

	thanosVersion, _ := semver.ParseTolerant(DefaultThanosVersion)
	pr := newRuleSelectorForConfigGeneration(ThanosFormat, thanosVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if !strings.Contains(content, "limit") {
		t.Fatalf("expected `limit` to be present in PrometheusRule")
	}
}

func shouldDropLimitFieldForUnsupportedPrometheusVersion(t *testing.T) {
	limit := 50
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
				Limit: &limit,
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant("v2.30.0")
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if strings.Contains(content, "limit") {
		t.Fatalf("expected `limit` not to be present in PrometheusRule")
	}
}

func shouldDropLimitFieldForUnsupportedThanosVersion(t *testing.T) {
	limit := 50
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
				Limit: &limit,
			},
		}},
	}

	thanosVersion, _ := semver.ParseTolerant("v0.23.0")
	pr := newRuleSelectorForConfigGeneration(ThanosFormat, thanosVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	if strings.Contains(content, "limit") {
		t.Fatalf("expected `limit` not to be present in PrometheusRule")
	}
}
