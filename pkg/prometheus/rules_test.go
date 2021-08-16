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
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
)

func TestMakeRulesConfigMaps(t *testing.T) {
	t.Run("ShouldReturnAtLeastOneConfigMap", shouldReturnAtLeastOneConfigMap)
	t.Run("ShouldErrorOnTooLargeRuleFile", shouldErrorOnTooLargeRuleFile)
	t.Run("ShouldSplitUpLargeSmallIntoTwo", shouldSplitUpLargeSmallIntoTwo)
	t.Run("ShouldAcceptValidRule", shouldAcceptValidRule)
	t.Run("shouldAcceptRuleWithValidPartialResponseStrategyValue", shouldAcceptRuleWithValidPartialResponseStrategyValue)
	t.Run("shouldRejectRuleWithInvalidLabels", shouldRejectRuleWithInvalidLabels)
	t.Run("shouldRejectRuleWithInvalidExpression", shouldRejectRuleWithInvalidExpression)
	t.Run("shouldRejectRuleWithInvalidPartialResponseStrategyValue", shouldRejectRuleWithInvalidPartialResponseStrategyValue)

}

func shouldAcceptValidRule(t *testing.T) {
	rules := monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
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
	}}
	_, err := GenerateContent(rules, log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err != nil {
		t.Fatalf("expected no errors when parsing valid rule")
	}
}

func shouldAcceptRuleWithValidPartialResponseStrategyValue(t *testing.T) {
	rules := monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
		{
			Name:                    "group",
			PartialResponseStrategy: "abort",
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
	}}
	_, err := GenerateContent(rules, log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err != nil {
		t.Fatalf("expected no errors when parsing rule with valid thanos partial_response_strategy value")
	}
}

func shouldRejectRuleWithInvalidLabels(t *testing.T) {
	rules := monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
		{
			Name: "group2",
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
	}}
	_, err := GenerateContent(rules, log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err == nil {
		t.Fatalf("expected errors when parsing rule with invalid labels")
	}
}

func shouldRejectRuleWithInvalidExpression(t *testing.T) {
	rules := monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
		{
			Name: "group2",
			Rules: []monitoringv1.Rule{
				{
					Alert: "alert",
					Expr:  intstr.FromString("invalidfn(1)"),
				},
			},
		},
	}}
	_, err := GenerateContent(rules, log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err == nil {
		t.Fatalf("expected errors when parsing rule with invalid expression")
	}
}

func shouldRejectRuleWithInvalidPartialResponseStrategyValue(t *testing.T) {
	rules := monitoringv1.PrometheusRuleSpec{Groups: []monitoringv1.RuleGroup{
		{
			Name:                    "group2",
			PartialResponseStrategy: "invalid",
			Rules: []monitoringv1.Rule{
				{
					Alert: "alert",
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	}}
	_, err := GenerateContent(rules, log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err == nil {
		t.Fatalf("expected errors when parsing rule with invalid partial_response_strategy value")
	}
}

// makeRulesConfigMaps should return at least one ConfigMap even if it is empty
// when there are no rules. Otherwise adding a rule to a Prometheus without rules
// would change the statefulset definition and thereby force Prometheus to
// restart.
func shouldReturnAtLeastOneConfigMap(t *testing.T) {
	p := &monitoringv1.Prometheus{}
	ruleFiles := map[string]string{}

	configMaps, err := makeRulesConfigMaps(p, ruleFiles)
	if err != nil {
		t.Fatalf("expected no error but got: %v", err.Error())
	}

	if len(configMaps) != 1 {
		t.Fatalf("expected one ConfigMaps but got %v", len(configMaps))
	}
}

func shouldErrorOnTooLargeRuleFile(t *testing.T) {
	expectedError := "rule file 'my-rule-file' is too large for a single Kubernetes ConfigMap"
	p := &monitoringv1.Prometheus{}
	ruleFiles := map[string]string{}

	ruleFiles["my-rule-file"] = strings.Repeat("a", v1.MaxSecretSize+1)

	_, err := makeRulesConfigMaps(p, ruleFiles)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("expected makeRulesConfigMaps to return error '%v' but got '%v'", expectedError, err)
	}
}

func shouldSplitUpLargeSmallIntoTwo(t *testing.T) {
	p := &monitoringv1.Prometheus{}
	ruleFiles := map[string]string{}

	ruleFiles["first"] = strings.Repeat("a", maxConfigMapDataSize)
	ruleFiles["second"] = "a"

	configMaps, err := makeRulesConfigMaps(p, ruleFiles)
	if err != nil {
		t.Fatalf("expected no error but got: %v", err)
	}

	if len(configMaps) != 2 {
		t.Fatalf("expected rule files to be split up into two ConfigMaps, but got '%v' instead", len(configMaps))
	}

	if configMaps[0].Data["first"] != ruleFiles["first"] || configMaps[1].Data["second"] != ruleFiles["second"] {
		t.Fatal("expected ConfigMap data to match rule file content")
	}
}
