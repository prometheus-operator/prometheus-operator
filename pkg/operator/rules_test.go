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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func TestMakeRulesConfigMaps(t *testing.T) {
	t.Run("shouldAcceptRuleWithValidPartialResponseStrategyValue", shouldAcceptRuleWithValidPartialResponseStrategyValue)
	t.Run("shouldAcceptValidRule", shouldAcceptValidRule)
	t.Run("shouldAcceptRulesWithEmptyDurations", shouldAcceptRulesWithEmptyDurations)
	t.Run("shouldAcceptRulesWithUTF8Labels", shouldAcceptRulesWithUTF8Labels)
	t.Run("shouldRejectRuleWithUTF8LabelsWithPrometheusVersionUnsupported", shouldRejectRuleWithUTF8LabelsWithPrometheusVersionUnsupported)
	t.Run("shouldRejectRuleWithInvalidExpression", shouldRejectRuleWithInvalidExpression)
	t.Run("shouldResetRuleWithPartialResponseStrategySet", shouldResetRuleWithPartialResponseStrategySet)
	t.Run("shouldAcceptRuleWithLimitPrometheus", shouldAcceptRuleWithLimitPrometheus)
	t.Run("shouldAcceptRuleWithLimitThanos", shouldAcceptRuleWithLimitThanos)
	t.Run("shouldAcceptRuleWithQueryOffsetPrometheus", shouldAcceptRuleWithQueryOffsetPrometheus)
	t.Run("shouldDropLimitFieldForUnsupportedPrometheusVersion", shouldDropLimitFieldForUnsupportedPrometheusVersion)
	t.Run("shouldDropLimitFieldForUnsupportedThanosVersion", shouldDropLimitFieldForUnsupportedThanosVersion)
	t.Run("shouldDropQueryOffsetFieldForUnsupportedPrometheusVersion", shouldDropQueryOffsetFieldForUnsupportedPrometheusVersion)
	t.Run("shouldAcceptRuleWithKeepFiringForPrometheus", shouldAcceptRuleWithKeepFiringForPrometheus)
	t.Run("shouldDropRuleFiringForThanos", shouldDropRuleFiringForThanos)
	t.Run("shouldAcceptRuleFiringForThanos", shouldAcceptRuleFiringForThanos)
	t.Run("shouldDropKeepFiringForFieldForUnsupportedPrometheusVersion", shouldDropKeepFiringForFieldForUnsupportedPrometheusVersion)
	t.Run("shouldErrorOnTooLargePrometheusRule", shouldErrorOnTooLargePrometheusRule)
	t.Run("shouldDropGroupLabelsForUnsupportedPrometheusVersion", shouldDropGroupLabelsForUnsupportedPrometheusVersion)
	t.Run("shouldAcceptRuleWithGroupLabels", shouldAcceptRuleWithGroupLabels)
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

	pr, err := NewPrometheusRuleSelector(ThanosFormat, DefaultThanosVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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
	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	_, err = pr.generateRulesConfiguration(rules)
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}

	_, err = pr.generateRulesConfiguration(rules)
	require.NoError(t, err)
}

func shouldRejectRuleWithUTF8LabelsWithPrometheusVersionUnsupported(t *testing.T) {
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
	pr, err := NewPrometheusRuleSelector(PrometheusFormat, "v2.55.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	_, err = pr.generateRulesConfiguration(rules)
	require.Error(t, err)
}

func shouldAcceptRulesWithUTF8Labels(t *testing.T) {
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
	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	_, err = pr.generateRulesConfiguration(rules)
	require.NoError(t, err)
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	_, err = pr.generateRulesConfiguration(rules)
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(ThanosFormat, "v0.33.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(ThanosFormat, DefaultThanosVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, "v2.33.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(ThanosFormat, DefaultThanosVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, "v2.30.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(ThanosFormat, "v0.23.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, "v2.32.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	content, err := pr.generateRulesConfiguration(rules)
	require.NoError(t, err)

	require.NotContains(t, content, "query_offset", "expected `query_offset` not to be present in PrometheusRule")
}

func shouldErrorOnTooLargePrometheusRule(t *testing.T) {
	ruleLbel := map[string]string{}
	ruleLbel["label"] = strings.Repeat("a", MaxConfigMapDataSize+1)

	err := ValidateRule(monitoringv1.PrometheusRuleSpec{
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
	})
	require.NotEmpty(t, err, "expected ValidateRule to return error of size limit")
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, "v2.53.0", nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
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

	pr, err := NewPrometheusRuleSelector(PrometheusFormat, DefaultPrometheusVersion, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("error in creating prometheus rule selector: %s", err)
	}
	_, err = pr.generateRulesConfiguration(rules)
	require.NoError(t, err)
}
