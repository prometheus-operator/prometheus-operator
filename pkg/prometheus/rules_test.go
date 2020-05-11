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
	"strings"
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestEnforcedNamespaceLabelRule(t *testing.T) {
	type testCase struct {
		Name                           string
		PromRuleSpec                   monitoringv1.PrometheusRuleSpec
		PromSpecEnforcedNamespaceLabel string
		PromRuleNamespace              string

		Expected monitoringv1.PrometheusRuleSpec
	}

	testcases := []testCase{
		{
			Name: "recordingrule-enforced",
			PromRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Record: "rule",
								Expr:   intstr.FromString("rate(requests_total{job=\"myjob\"}[5m])"),
							},
						},
					},
				},
			},
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromRuleNamespace:              "bar",

			Expected: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Record: "rule",
								Expr:   intstr.FromString("rate(requests_total{job=\"myjob\",namespace=\"bar\"}[5m])"),
								Labels: map[string]string{"namespace": "bar"},
							},
						},
					},
				},
			},
		},
		{
			Name: "alertname-enforced",
			PromRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Expr:  intstr.FromString("node_cpu_seconds_total{job=\"node-exporter\",mode=\"idle\"}"),
							},
						},
					},
				},
			},
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromRuleNamespace:              "bar",

			Expected: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert:  "alert",
								Expr:   intstr.FromString("node_cpu_seconds_total{job=\"node-exporter\",mode=\"idle\",namespace=\"bar\"}"),
								Labels: map[string]string{"namespace": "bar"},
							},
						},
					},
				},
			},
		},
		{
			Name: "alertname-enforced-removed-ns",
			PromRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Expr:  intstr.FromString("node_cpu_seconds_total{namespace=\"foo-bar\",mode=\"idle\"}"),
							},
						},
					},
				},
			},
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromRuleNamespace:              "bar",

			Expected: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert:  "alert",
								Expr:   intstr.FromString("node_cpu_seconds_total{mode=\"idle\",namespace=\"bar\"}"),
								Labels: map[string]string{"namespace": "bar"},
							},
						},
					},
				},
			},
		},
		{
			Name: "alertname-enforced-no-labels",
			PromRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Labels: map[string]string{
									"foo": "bar",
								},
								Expr: intstr.FromString("http_requests_total"),
							},
						},
					},
				},
			},
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromRuleNamespace:              "default",

			Expected: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Labels: map[string]string{
									"foo":       "bar",
									"namespace": "default",
								},
								Expr: intstr.FromString("http_requests_total{namespace=\"default\"}"),
							},
						},
					},
				},
			},
		},
		{
			Name: "alertname-not-enforced",
			PromRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Expr:  intstr.FromString("vector(1)"),
							},
						},
					},
				},
			},
			PromSpecEnforcedNamespaceLabel: "",
			PromRuleNamespace:              "default",

			Expected: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Rules: []monitoringv1.Rule{
							{
								Alert: "alert",
								Expr:  intstr.FromString("vector(1)"),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name,
			func(t *testing.T) {
				if tc.PromSpecEnforcedNamespaceLabel != "" {
					err := injectNamespaceLabel(&tc.PromRuleSpec, tc.PromSpecEnforcedNamespaceLabel, tc.PromRuleNamespace)
					if err != nil {
						t.Error(err)
					}
				}
				if diff := cmp.Diff(tc.Expected, tc.PromRuleSpec); diff != "" {
					t.Errorf("Unexpected result (-want +got):\n%s", diff)
				}
			},
		)
	}
}

func TestPrometheusRuleNamespaceLabelExclude(t *testing.T) {
	type testCase struct {
		Name     string
		PromSpec monitoringv1.PrometheusSpec
		Expected nsLabelEnforcementExcludeList
	}

	testcases := []testCase{
		{
			Name: "valid excludes",
			PromSpec: monitoringv1.PrometheusSpec{
				PrometheusRulesExcludedFromEnforce: []monitoringv1.PrometheusRuleExcludeConfig{
					{RuleName: "RuleSet1", RuleNamespace: "system"},
					{RuleName: "RuleSet2", RuleNamespace: "system"},
					{RuleName: "commonAlerts", RuleNamespace: "monitoring"},
				},
			},
			Expected: nsLabelEnforcementExcludeList{
				"system":     map[string]struct{}{"RuleSet1": {}, "RuleSet2": {}},
				"monitoring": map[string]struct{}{"commonAlerts": {}},
			},
		},
		{
			Name: "invalid excludes",
			PromSpec: monitoringv1.PrometheusSpec{
				PrometheusRulesExcludedFromEnforce: []monitoringv1.PrometheusRuleExcludeConfig{
					{RuleName: "", RuleNamespace: "system"},
					{RuleName: "RuleSet2", RuleNamespace: ""},
				},
			},
			Expected: nsLabelEnforcementExcludeList{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			got := newNSLabelEnforcementExcludeList(tc.PromSpec.PrometheusRulesExcludedFromEnforce)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrometheusExcludeListContains(t *testing.T) {
	l := nsLabelEnforcementExcludeList{
		"system":     map[string]struct{}{"RuleSet1": {}, "RuleSet2": {}},
		"monitoring": map[string]struct{}{"commonAlerts": {}},
	}

	testcases := []struct {
		ruleNamespace string
		ruleName      string
		expected      bool
	}{
		{"system", "RuleSet1", true},
		{"system", "RuleSet4", false},
		{"monitoring", "RuleSet1", false},
		{"monitoring", "commonAlerts", true},
	}
	for _, tc := range testcases {
		if got := l.Contains(tc.ruleNamespace, tc.ruleName); got != tc.expected {
			t.Errorf("%s/%s want %t - got %t", tc.ruleNamespace, tc.ruleName, tc.expected, got)
		}
	}
}

func TestMakeRulesConfigMaps(t *testing.T) {
	t.Run("ShouldReturnAtLeastOneConfigMap", shouldReturnAtLeastOneConfigMap)
	t.Run("ShouldErrorOnTooLargeRuleFile", shouldErrorOnTooLargeRuleFile)
	t.Run("ShouldSplitUpLargeSmallIntoTwo", shouldSplitUpLargeSmallIntoTwo)
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
