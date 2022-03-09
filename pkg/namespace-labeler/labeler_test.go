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

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestEnforceNamespaceLabelOnPrometheusRules(t *testing.T) {

	type testCase struct {
		Name                           string
		PromRule                       monitoringv1.PrometheusRule
		PromSpecEnforcedNamespaceLabel string
		ExcludedFromEnforcement        []monitoringv1.ObjectReference
		PromSpecExcludedRules          []monitoringv1.PrometheusRuleExcludeConfig

		Expected monitoringv1.PrometheusRule
	}

	testcases := []testCase{
		{
			Name: "rule-ns-enforced-add",
			PromRule: expandPromRule(&promRuleFlat{
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
			Expected: expandPromRule(&promRuleFlat{
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
			PromRule: expandPromRule(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert:  "alert",
					Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"dummy\",mode=\"idle\"}"),
					Labels: map[string]string{"namespace": "dummy"},
				}},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",

			Expected: expandPromRule(&promRuleFlat{
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
			PromRule: expandPromRule(&promRuleFlat{
				Name:      "foo",
				Namespace: "bar",
				Rules: []monitoringv1.Rule{{
					Alert: "alert",
					Expr:  intstr.FromString("node_cpu_seconds_total"),
				}},
			}),
			PromSpecEnforcedNamespaceLabel: "",

			Expected: expandPromRule(&promRuleFlat{
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
			PromRule: expandPromRule(&promRuleFlat{
				Name:      "alert",
				Namespace: "bar",
				Labels: map[string]string{
					"group": "group1",
				},
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
			ExcludedFromEnforcement: []monitoringv1.ObjectReference{
				{
					Namespace: "foo",
					Group:     "monitoring.coreos.com",
					Resource:  monitoringv1.PrometheusRuleName,
				},
			},

			Expected: expandPromRule(&promRuleFlat{
				Name:      "alert",
				Namespace: "bar",
				Labels: map[string]string{
					"group": "group1",
				},
				Rules: []monitoringv1.Rule{{
					Alert:  "alert",
					Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"bar\"}"),
					Labels: map[string]string{"namespace": "bar"},
				}},
			}),
		},
		{
			Name: "excludeList-exist-match",
			PromRule: expandPromRule(&promRuleFlat{
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

			Expected: expandPromRule(&promRuleFlat{
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
		{
			Name: "excludeList-exist-match-selector",
			PromRule: expandPromRule(&promRuleFlat{
				Name:      "rule-to-exclude",
				Namespace: "bar",
				Labels: map[string]string{
					"id": "rule-to-exclude",
				},
				Rules: []monitoringv1.Rule{
					{
						Alert:  "alert",
						Expr:   intstr.FromString("node_cpu_seconds_total{namespace=\"DUMMY1\"}"),
						Labels: map[string]string{"namespace": "DUMMY2"},
					},
				},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			ExcludedFromEnforcement: []monitoringv1.ObjectReference{
				{
					Namespace: "bar",
					Group:     monitoring.GroupName,
					Resource:  monitoringv1.PrometheusRuleName,
				},
			},

			Expected: expandPromRule(&promRuleFlat{
				Name:      "rule-to-exclude",
				Namespace: "bar",
				Labels: map[string]string{
					"id": "rule-to-exclude",
				},
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

				excludedFromEnforcement := tc.ExcludedFromEnforcement
				// append the deprecated PrometheusRulesExcludedFromEnforce
				for _, rule := range tc.PromSpecExcludedRules {
					excludedFromEnforcement = append(excludedFromEnforcement,
						monitoringv1.ObjectReference{
							Namespace: rule.RuleNamespace,
							Group:     monitoring.GroupName,
							Resource:  monitoringv1.PrometheusRuleName,
							Name:      rule.RuleName,
						})
				}
				nsLabeler := New(tc.PromSpecEnforcedNamespaceLabel, excludedFromEnforcement, true)

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

func TestEnforceNamespaceLabelOnPrometheusMonitors(t *testing.T) {

	type testCase struct {
		Name                           string
		ServiceMonitor                 monitoringv1.ServiceMonitor
		PromSpecEnforcedNamespaceLabel string
		ExcludedFromEnforcement        []monitoringv1.ObjectReference

		Expected monitoringv1.ServiceMonitor
	}

	testcases := []testCase{
		{
			Name: "servicemonitor-ns-enforced-add",
			ServiceMonitor: expandServiceMonitor(&promServiceMonitorFlat{
				Name:                 "foo",
				Namespace:            "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{},
				RelabelConfigs:       []*monitoringv1.RelabelConfig{},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			Expected: expandServiceMonitor(&promServiceMonitorFlat{
				Name:      "foo",
				Namespace: "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
					{
						TargetLabel: "namespace",
						Replacement: "bar",
					},
				},
				RelabelConfigs: []*monitoringv1.RelabelConfig{
					{
						TargetLabel: "namespace",
						Replacement: "bar",
					},
				},
			}),
		},
		{
			Name: "servicemonitor-ns-enforced-exclude-by-name",
			ServiceMonitor: expandServiceMonitor(&promServiceMonitorFlat{
				Name:                 "exclude-me",
				Namespace:            "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{},
				RelabelConfigs:       []*monitoringv1.RelabelConfig{},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			ExcludedFromEnforcement: []monitoringv1.ObjectReference{
				{
					Namespace: "bar",
					Group:     monitoring.GroupName,
					Resource:  monitoringv1.ServiceMonitorName,
					Name:      "exclude-me",
				},
			},
			Expected: expandServiceMonitor(&promServiceMonitorFlat{
				Name:                 "exclude-me",
				Namespace:            "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{},
				RelabelConfigs:       []*monitoringv1.RelabelConfig{},
			}),
		},
		{
			Name: "servicemonitor-ns-enforced-exclude-all-by-namespace",
			ServiceMonitor: expandServiceMonitor(&promServiceMonitorFlat{
				Name:                 "exclude-me",
				Namespace:            "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{},
				RelabelConfigs:       []*monitoringv1.RelabelConfig{},
			}),
			PromSpecEnforcedNamespaceLabel: "namespace",
			ExcludedFromEnforcement: []monitoringv1.ObjectReference{
				{
					Namespace: "bar",
					Group:     monitoring.GroupName,
					Resource:  monitoringv1.ServiceMonitorName,
				},
			},
			Expected: expandServiceMonitor(&promServiceMonitorFlat{
				Name:                 "exclude-me",
				Namespace:            "bar",
				MetricRelabelConfigs: []*monitoringv1.RelabelConfig{},
				RelabelConfigs:       []*monitoringv1.RelabelConfig{},
			}),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name,
			func(t *testing.T) {
				nsLabeler := New(tc.PromSpecEnforcedNamespaceLabel, tc.ExcludedFromEnforcement, true)
				tc.ServiceMonitor.Spec.Endpoints[0].MetricRelabelConfigs = nsLabeler.GetRelabelingConfigs(tc.ServiceMonitor.TypeMeta, tc.ServiceMonitor.ObjectMeta, tc.ServiceMonitor.Spec.Endpoints[0].MetricRelabelConfigs)
				tc.ServiceMonitor.Spec.Endpoints[0].RelabelConfigs = nsLabeler.GetRelabelingConfigs(tc.ServiceMonitor.TypeMeta, tc.ServiceMonitor.ObjectMeta, tc.ServiceMonitor.Spec.Endpoints[0].RelabelConfigs)
				if diff := cmp.Diff(tc.Expected, tc.ServiceMonitor); diff != "" {
					t.Errorf("Unexpected result (-want +got):\n%s", diff)
				}
			},
		)
	}
}

type promRuleFlat struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Rules     []monitoringv1.Rule
}

func expandPromRule(r *promRuleFlat) monitoringv1.PrometheusRule {
	return monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{Rules: r.Rules},
			},
		},
	}
}

type promServiceMonitorFlat struct {
	Name                 string
	Namespace            string
	Labels               map[string]string
	MetricRelabelConfigs []*monitoringv1.RelabelConfig
	RelabelConfigs       []*monitoringv1.RelabelConfig
}

func expandServiceMonitor(r *promServiceMonitorFlat) monitoringv1.ServiceMonitor {
	return monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.ServiceMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					MetricRelabelConfigs: r.MetricRelabelConfigs,
					RelabelConfigs:       r.RelabelConfigs,
				},
			},
		},
	}
}
