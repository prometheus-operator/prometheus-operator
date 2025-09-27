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
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMakeRulesConfigMaps(t *testing.T) {
	// Basic validation
	t.Run("shouldAcceptValidRule", shouldAcceptValidRule)
	t.Run("shouldRejectRuleWithInvalidLabels", shouldRejectRuleWithInvalidLabels)
	t.Run("shouldRejectRuleWithInvalidExpression", shouldRejectRuleWithInvalidExpression)
	t.Run("shouldAcceptRulesWithEmptyDurations", shouldAcceptRulesWithEmptyDurations)
	t.Run("shouldErrorOnTooLargePrometheusRule", shouldErrorOnTooLargePrometheusRule)

	// Prometheus features
	t.Run("shouldAcceptRuleWithLimitPrometheus", shouldAcceptRuleWithLimitPrometheus)
	t.Run("shouldDropLimitFieldForUnsupportedPrometheusVersion", shouldDropLimitFieldForUnsupportedPrometheusVersion)
	t.Run("shouldAcceptRuleWithQueryOffsetPrometheus", shouldAcceptRuleWithQueryOffsetPrometheus)
	t.Run("shouldDropQueryOffsetFieldForUnsupportedPrometheusVersion", shouldDropQueryOffsetFieldForUnsupportedPrometheusVersion)
	t.Run("shouldAcceptRuleWithKeepFiringForPrometheus", shouldAcceptRuleWithKeepFiringForPrometheus)
	t.Run("shouldDropKeepFiringForFieldForUnsupportedPrometheusVersion", shouldDropKeepFiringForFieldForUnsupportedPrometheusVersion)
	t.Run("shouldDropGroupLabelsForUnsupportedPrometheusVersion", shouldDropGroupLabelsForUnsupportedPrometheusVersion)
	t.Run("shouldAcceptRuleWithGroupLabels", shouldAcceptRuleWithGroupLabels)

	// Thanos features
	t.Run("shouldAcceptRuleWithValidPartialResponseStrategyValue", shouldAcceptRuleWithValidPartialResponseStrategyValue)
	t.Run("shouldResetRuleWithPartialResponseStrategySet", shouldResetRuleWithPartialResponseStrategySet)
	t.Run("shouldAcceptRuleWithLimitThanos", shouldAcceptRuleWithLimitThanos)
	t.Run("shouldDropLimitFieldForUnsupportedThanosVersion", shouldDropLimitFieldForUnsupportedThanosVersion)
	t.Run("shouldDropRuleFiringForThanos", shouldDropRuleFiringForThanos)
	t.Run("shouldAcceptRuleFiringForThanos", shouldAcceptRuleFiringForThanos)

	// UTF-8 validation
	t.Run("UTF8Validation", TestUTF8Validation)
}

func newRuleSelectorForConfigGeneration(ruleFormat RuleConfigurationFormat, version semver.Version) PrometheusRuleSelector {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return PrometheusRuleSelector{
		ruleFormat: ruleFormat,
		version:    version,
		logger:     logger,
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
	require.Contains(t, content, "partial_response_strategy: warn", "expected `partial_response_strategy` to be set in PrometheusRule as `warn`")
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
	require.NoError(t, err)
}

func shouldAcceptRulesWithEmptyDurations(t *testing.T) {
	durationPtr := func(d string) *monitoringv1.Duration {
		v := monitoringv1.Duration(d)
		return &v
	}

	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name:     "group",
				Interval: durationPtr(""),
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
						Labels: map[string]string{
							"valid_label": "valid_value",
						},
						For: durationPtr(""),
					},
				},
			},
		}},
	}
	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	_, err := pr.generateRulesConfiguration(rules)
	require.NoError(t, err)
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
	promVersion, err := semver.ParseTolerant("2.55.0")
	require.NoError(t, err)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	_, err = pr.generateRulesConfiguration(rules)
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.NotContains(t, content, "partial_response_strategy", "expected `partial_response_strategy` removed from PrometheusRule")
}

func shouldAcceptRuleWithKeepFiringForPrometheus(t *testing.T) {
	duration := monitoringv1.NonEmptyDuration("5m")
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert:         "alert",
						Expr:          intstr.FromString("vector(1)"),
						KeepFiringFor: &duration,
					},
				},
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant(DefaultPrometheusVersion)
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	require.Contains(t, content, "keep_firing_for", "expected `keep_firing_for` to be present in PrometheusRule")
}

func shouldDropRuleFiringForThanos(t *testing.T) {
	duration := monitoringv1.NonEmptyDuration("5m")
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert:         "alert",
						Expr:          intstr.FromString("vector(1)"),
						KeepFiringFor: &duration,
					},
				},
			},
		}},
	}

	thanosVersion, _ := semver.ParseTolerant("v0.33.0")
	pr := newRuleSelectorForConfigGeneration(ThanosFormat, thanosVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	require.NotContains(t, content, "keep_firing_for", "expected `keep_firing_for` not to be present in PrometheusRule")
}

func shouldAcceptRuleFiringForThanos(t *testing.T) {
	duration := monitoringv1.NonEmptyDuration("5m")
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert:         "alert",
						Expr:          intstr.FromString("vector(1)"),
						KeepFiringFor: &duration,
					},
				},
			},
		}},
	}

	thanosVersion, _ := semver.ParseTolerant(DefaultThanosVersion)
	pr := newRuleSelectorForConfigGeneration(ThanosFormat, thanosVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	require.Contains(t, content, "keep_firing_for", "expected `keep_firing_for` to be present in PrometheusRule")
}

func shouldDropKeepFiringForFieldForUnsupportedPrometheusVersion(t *testing.T) {
	duration := monitoringv1.NonEmptyDuration("5m")
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert:         "alert",
						Expr:          intstr.FromString("vector(1)"),
						KeepFiringFor: &duration,
					},
				},
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant("v2.30.0")
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)
	require.NotContains(t, content, "keep_firing_for", "expected `keep_firing_for` not to be present in PrometheusRule")
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
	require.Contains(t, content, "limit", "expected `limit` to be present in PrometheusRule")
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
	require.Contains(t, content, "limit", "expected `limit` to be present in PrometheusRule")
}

func shouldAcceptRuleWithQueryOffsetPrometheus(t *testing.T) {
	var queryOffset monitoringv1.Duration = "30s"
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
				QueryOffset: &queryOffset,
			},
		}},
	}

	promVersion, err := semver.ParseTolerant(DefaultPrometheusVersion)
	require.NoError(t, err)

	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, err := pr.generateRulesConfiguration(rules)
	require.NoError(t, err)

	require.Contains(t, content, "query_offset", "expected `query_offset` to be present in PrometheusRule")
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
	require.NotContains(t, content, "limit", "expected `limit` not to be present in PrometheusRule")
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
	require.NotContains(t, content, "limit", "expected `limit` not to be present in PrometheusRule")
}

func shouldDropQueryOffsetFieldForUnsupportedPrometheusVersion(t *testing.T) {
	var queryOffset monitoringv1.Duration = "30s"
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
				QueryOffset: &queryOffset,
			},
		}},
	}

	promVersion, err := semver.ParseTolerant("v2.52.0")
	require.NoError(t, err)

	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, err := pr.generateRulesConfiguration(rules)
	require.NoError(t, err)

	require.NotContains(t, content, "query_offset", "expected `query_offset` not to be present in PrometheusRule")
}

func shouldErrorOnTooLargePrometheusRule(t *testing.T) {
	ruleLbel := map[string]string{}
	ruleLbel["label"] = strings.Repeat("a", MaxConfigMapDataSize+1)

	ruleSpec := monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Record: "record",
						Expr:   intstr.FromString("vector(1)"),
						Alert:  "alert",
						Labels: ruleLbel,
					},
				},
			},
		},
	}

	err := ValidateRule(ruleSpec, model.UTF8Validation)
	require.NotEmpty(t, err, "expected ValidateRule to return error of size limit with UTF8Validation")

	err = ValidateRule(ruleSpec, model.LegacyValidation)
	require.NotEmpty(t, err, "expected ValidateRule to return error of size limit with LegacyValidation")
}

func shouldDropGroupLabelsForUnsupportedPrometheusVersion(t *testing.T) {
	labels := map[string]string{
		"key": "value",
	}
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name:   "group",
				Labels: labels,
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
					},
				},
			},
		}},
	}

	promVersion, _ := semver.ParseTolerant("2.53.0")
	pr := newRuleSelectorForConfigGeneration(PrometheusFormat, promVersion)
	content, _ := pr.generateRulesConfiguration(rules)

	require.NotContains(t, content, "key", "expected group labels not to be present in PrometheusRule")
	require.NotContains(t, content, "value", "expected group labels not to be present in PrometheusRule")
}

func shouldAcceptRuleWithGroupLabels(t *testing.T) {
	labels := map[string]string{
		"key": "value",
	}
	rules := &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name:   "group",
				Labels: labels,
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
	_, err := pr.generateRulesConfiguration(rules)
	require.NoError(t, err)
}

func TestUTF8Validation(t *testing.T) {
	rule := createUTF8Rule()

	tests := []struct {
		name          string
		format        RuleConfigurationFormat
		version       string
		shouldSucceed bool
	}{
		{"Prometheus 3.0.0 accepts UTF-8", PrometheusFormat, "3.0.0", true},
		{"Prometheus 2.55.0 rejects UTF-8", PrometheusFormat, "2.55.0", false},
		{"Thanos 0.38.0 accepts UTF-8", ThanosFormat, "0.38.0", true},
		{"Thanos 0.37.0 rejects UTF-8", ThanosFormat, "0.37.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := semver.ParseTolerant(tt.version)
			require.NoError(t, err)

			pr := newRuleSelectorForConfigGeneration(tt.format, version)
			_, err = pr.generateRulesConfiguration(rule)

			if tt.shouldSucceed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func createUTF8Rule() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		Spec: monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
			{
				Name: "group",
				Rules: []monitoringv1.Rule{
					{
						Alert: "alert",
						Expr:  intstr.FromString("vector(1)"),
						Labels: map[string]string{
							"unicode_测试": "utf8_value",
						},
					},
				},
			},
		}},
	}
}
