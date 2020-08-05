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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestEnforceNamespaceLabel(t *testing.T) {

	type testCase struct {
		Name                           string
		PromRule                       monitoringv1.PrometheusRule
		PromSpecEnforcedNamespaceLabel string
		PromSpecExcludedRules          []monitoringv1.PrometheusRuleExcludeConfig

		Expected monitoringv1.PrometheusRule
	}

	testcases := []testCase{
		{
			Name: "rule-ns-enforced-add",
			PromRule: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{
					{
						Record: "rule",
						Expr:   intstr.FromString("rate(requests_total{job=\"myjob\"}[5m])"),
					},
					{
						Alert: "alert",
						Expr:  intstr.FromString("node_cpu_seconds_total{job=\"node-exporter\",mode=\"idle\"}"),
					},
					{
						Record: "rule-no-labels",
						Expr:   intstr.FromString("rate(http_requests_total[5m])"),
					},
				},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			Expected: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{
					{
						Record: "rule",
						Expr:   intstr.FromString("rate(requests_total{job=\"myjob\",namespace=\"bar\"}[5m])"),
						Labels: map[string]string{"namespace": "bar"},
					},
					{
						Alert:  "alert",
						Expr:   intstr.FromString("node_cpu_seconds_total{job=\"node-exporter\",mode=\"idle\",namespace=\"bar\"}"),
						Labels: map[string]string{"namespace": "bar"},
					},
					{
						Record: "rule-no-labels",
						Expr:   intstr.FromString("rate(http_requests_total{namespace=\"bar\"}[5m])"),
						Labels: map[string]string{"namespace": "bar"},
					},
				},
			}),
		},
		{
			Name: "rule-ns-enforced-replace",
			PromRule: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert:  "alert",
					Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"dummy\",mode=\"idle\"}"),
					Labels: map[string]string{"namespace": "dummy"},
				}},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",

			Expected: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert:  "alert",
					Expr:   intstr.FromString("node_cpu_seconds_total{mode=\"idle\",namespace=\"bar\"}"),
					Labels: map[string]string{"namespace": "bar"},
				}},
			}),
		},
		{
			Name: "namespace-not-enforced",
			PromRule: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert: "alert",
					Expr:  intstr.FromString("node_cpu_seconds_total"),
				}},
			}),
			PromSpecEnforcedNamespaceLabel: "",

			Expected: expand(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert: "alert",
					Expr:  intstr.FromString("node_cpu_seconds_total"),
				}},
			}),
		},
		{
			Name: "excludeList-exist-but-no-match",
			PromRule: expand(&promRuleFlat{
				Name:      "alert",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert: "alert",
					Expr:  intstr.FromString("node_cpu_seconds_total"),
				}},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromSpecExcludedRules: []monitoringv1.PrometheusRuleExcludeConfig{
				{
					RuleName:      "excluded",
					RuleNamespace: "bar",
				},
			},

			Expected: expand(&promRuleFlat{
				Name:      "alert",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert:  "alert",
					Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"bar\"}"),
					Labels: map[string]string{"namespace": "bar"},
				}},
			}),
		},
		{
			Name: "excludeList-exist-match",
			PromRule: expand(&promRuleFlat{
				Name:      "rule-to-exclude",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{
					{
						Alert:  "alert",
						Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"DUMMY1\"}"),
						Labels: map[string]string{"namespace": "DUMMY2"},
					},
				},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			PromSpecExcludedRules: []monitoringv1.PrometheusRuleExcludeConfig{
				{
					RuleName:      "rule-to-exclude",
					RuleNamespace: "bar",
				},
			},

			Expected: expand(&promRuleFlat{
				Name:      "rule-to-exclude",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{
					{
						Alert:  "alert",
						Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"DUMMY1\"}"),
						Labels: map[string]string{"namespace": "DUMMY2"},
					},
				},
			}),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name,
			func(t *testing.T) {

				nsLabeler := New(tc.PromSpecEnforcedNamespaceLabel, tc.PromSpecExcludedRules, true)

				if err := nsLabeler.EnforceNamespaceLabel(&tc.PromRule); err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(tc.Expected, tc.PromRule); diff != "" {
					t.Errorf("Unexpected result (-want +got):\n%s", diff)
				}
			},
		)
	}
}

type promRuleFlat struct {
	Name      string
	Namespace string
	Rules     []monitoringv1.Rule
}

func expand(r *promRuleFlat) monitoringv1.PrometheusRule {
	return monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{Rules: r.Rules},
			},
		},
	}
}
