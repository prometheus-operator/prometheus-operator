// Copyright The prometheus-operator Authors
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

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1ac "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
)

func TestPrometheusRuleCRDValidation(t *testing.T) {
	keepFiringFor := monitoringv1.NonEmptyDuration("5m")
	emptyDuration := monitoringv1.NonEmptyDuration("")

	skipPrometheusTests(t)
	t.Parallel()

	tests := []struct {
		name          string
		promRuleSpec  monitoringv1.PrometheusRuleSpec
		expectedError bool
	}{
		{
			name: "duplicate-rule-name",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:  "rule1",
						Rules: []monitoringv1.Rule{},
					},
					{
						Name:  "rule1",
						Rules: []monitoringv1.Rule{},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid-partial-rsp",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "invalid",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "valid-rule-names",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:  "rule1",
						Rules: []monitoringv1.Rule{},
					},
					{
						Name:  "rule2",
						Rules: []monitoringv1.Rule{},
					},
				},
			},
		},
		{
			name: "empty-rule",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name: "empty",
					},
				},
			},
		},
		{
			name: "valid-partial-rsp-1",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "abort",
					},
				},
			},
		},
		{
			name: "valid-partial-rsp-2",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "ABORT",
					},
				},
			},
		},
		{
			name: "valid-partial-rsp-3",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "Warn",
					},
				},
			},
		},
		{
			name: "valid-partial-rsp-4",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "WaRn",
					},
				},
			},
		},
		{
			name: "valid-partial-rsp-5",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:                    "test",
						Rules:                   []monitoringv1.Rule{},
						PartialResponseStrategy: "",
					},
				},
			},
		},
		{
			name: "valid-keep-firing-for",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name: "test",
						Rules: []monitoringv1.Rule{
							{
								Record:        "rule1",
								Expr:          intstr.FromString("vector(0)"),
								KeepFiringFor: &keepFiringFor,
							},
						},
					},
				},
			},
		},
		{
			name: "invalid-keep-firing-for",
			promRuleSpec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name: "test",
						Rules: []monitoringv1.Rule{
							{
								Record:        "rule1",
								Expr:          intstr.FromString("vector(0)"),
								KeepFiringFor: &emptyDuration,
							},
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)

			promRule := framework.MakeBasicRule(ns, "prometheus-rule", test.promRuleSpec.Groups)
			_, err := framework.MonClientV1.PrometheusRules(ns).Create(context.Background(), promRule, metav1.CreateOptions{})

			if err == nil {
				if test.expectedError {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if !apierrors.IsInvalid(err) {
				t.Fatalf("expected Invalid error but got %v", err)
			}
		})
	}
}

func TestPrometheusRuleApply(t *testing.T) {
	skipPrometheusTests(t)
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	const (
		firstManager  = "first"
		secondManager = "second"
	)

	for _, tc := range []struct {
		name         string
		fieldManager string
		force        bool
		groups       []*monitoringv1ac.RuleGroupApplyConfiguration
		expected     []monitoringv1.RuleGroup
		expectedErr  bool
	}{
		{
			name:         "initial apply by firstManager",
			fieldManager: firstManager,
			groups: []*monitoringv1ac.RuleGroupApplyConfiguration{
				monitoringv1ac.RuleGroup().WithName("firstGroup").WithRules(
					monitoringv1ac.Rule().WithRecord("firstRule").WithExpr(intstr.FromString("vector(0)")),
				),
			},
			expected: []monitoringv1.RuleGroup{
				{
					Name: "firstGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRule",
						Expr:   intstr.FromString("vector(0)"),
					}},
				},
			},
		},
		{
			name:         "apply by secondManager",
			fieldManager: secondManager,
			groups: []*monitoringv1ac.RuleGroupApplyConfiguration{
				monitoringv1ac.RuleGroup().WithName("secondGroup").WithRules(
					monitoringv1ac.Rule().WithRecord("firstRule").WithExpr(intstr.FromString("vector(1)")),
				),
				monitoringv1ac.RuleGroup().WithName("thirdGroup").WithRules(
					monitoringv1ac.Rule().WithRecord("firstRule").WithExpr(intstr.FromString("vector(2)")),
				),
			},
			expected: []monitoringv1.RuleGroup{
				{
					Name: "firstGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRule",
						Expr:   intstr.FromString("vector(0)"),
					}},
				},
				{
					Name: "secondGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRule",
						Expr:   intstr.FromString("vector(1)"),
					}},
				},
				{
					Name: "thirdGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRule",
						Expr:   intstr.FromString("vector(2)"),
					}},
				},
			},
		},
		{
			name:         "apply by firstManager with conflict",
			fieldManager: firstManager,
			groups: []*monitoringv1ac.RuleGroupApplyConfiguration{
				monitoringv1ac.RuleGroup().WithName("secondGroup").WithRules(
					monitoringv1ac.Rule().WithRecord("firstRuleModified").WithExpr(intstr.FromString("vector(0)")),
				),
			},
			expectedErr: true,
		},
		{
			name:         "apply by firstManager with force",
			fieldManager: firstManager,
			groups: []*monitoringv1ac.RuleGroupApplyConfiguration{
				monitoringv1ac.RuleGroup().WithName("secondGroup").WithRules(
					monitoringv1ac.Rule().WithRecord("firstRuleModified").WithExpr(intstr.FromString("vector(0)")),
				),
			},
			force: true,
			expected: []monitoringv1.RuleGroup{
				{
					Name: "secondGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRuleModified",
						Expr:   intstr.FromString("vector(0)"),
					}},
				},
				{
					Name: "thirdGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRule",
						Expr:   intstr.FromString("vector(2)"),
					}},
				},
			},
		},
		{
			name:         "remove all groups managed by secondManager",
			fieldManager: secondManager,
			groups:       []*monitoringv1ac.RuleGroupApplyConfiguration{},
			expected: []monitoringv1.RuleGroup{
				{
					Name: "secondGroup",
					Rules: []monitoringv1.Rule{{
						Record: "firstRuleModified",
						Expr:   intstr.FromString("vector(0)"),
					}},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prac := monitoringv1ac.PrometheusRule("rule", ns).WithSpec(
				monitoringv1ac.PrometheusRuleSpec().WithGroups(tc.groups...),
			)
			res, err := framework.MonClientV1.PrometheusRules(ns).Apply(
				context.Background(),
				prac,
				metav1.ApplyOptions{FieldManager: tc.fieldManager, Force: tc.force},
			)

			if tc.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, res.Spec.Groups)
		})
	}
}

func testPrometheusRuleWithParserOptions(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	// Skip the admission webhook because the test exercices PromQL features
	// which aren't enabled by default.
	ruleNamespaceSelector := map[string]string{"excludeFromWebhook": "true"}
	err := framework.AddLabelsToNamespace(context.Background(), ns, ruleNamespaceSelector)
	require.NoError(t, err)

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	// Create a PrometheusRule with one group per parser option and an always-firing alert.
	// The rules exercise syntax that would be rejected without the corresponding feature flag.
	rule := framework.MakeBasicRule(ns, "parser-options-rule", []monitoringv1.RuleGroup{
		{
			Name: "always-firing",
			Rules: []monitoringv1.Rule{
				{
					Alert: "AlwaysFiring",
					Expr:  intstr.FromString("vector(1) > 0"),
				},
			},
		},
		{
			Name: "experimental-functions",
			Rules: []monitoringv1.Rule{
				{
					Record: "test:limitk",
					Expr:   intstr.FromString("limitk(5, up)"),
				},
			},
		},
		{
			Name: "duration-expr",
			Rules: []monitoringv1.Rule{
				{
					Record: "test:duration_expr",
					Expr:   intstr.FromString("rate(up[2*1m])"),
				},
			},
		},
		{
			Name: "extended-range-selectors",
			Rules: []monitoringv1.Rule{
				{
					Record: "test:smoothed",
					Expr:   intstr.FromString("rate(up[1m] smoothed)"),
				},
			},
		},
		{
			Name: "binop-fill-modifiers",
			Rules: []monitoringv1.Rule{
				{
					Record: "test:fill",
					Expr:   intstr.FromString("up + fill(0) up"),
				},
			},
		},
	})
	_, err = framework.CreateRule(context.Background(), ns, rule)
	require.NoError(t, err)

	// Deploy Prometheus with all promql-* features enabled so that the rules above are accepted.
	p := framework.MakeBasicPrometheus(ns, "test", "test", 1)
	p.Spec.EvaluationInterval = "1s"
	p.Spec.EnableFeatures = []monitoringv1.EnableFeature{
		"promql-experimental-functions",
		"promql-duration-expr",
		"promql-extended-range-selectors",
		"promql-binop-fill-modifiers",
	}
	p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	pSVC := framework.MakePrometheusService(p.Name, p.Name, corev1.ServiceTypeClusterIP)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
	require.NoError(t, err)

	// Firing alert confirms the PrometheusRule was accepted and evaluated.
	err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, "AlwaysFiring")
	require.NoError(t, err)
}
