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
	"fmt"
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestRulesGenerateContent(t *testing.T) {
	wantRuleConfig := `groups:
- name: rule1
  rules:
  - alert: test_alert1
    expr: up{namespace="foo"}
    labels:
      namespace: foo
- name: rule2
  partial_response_strategy: warn
  rules:
  - alert: test_alert2
    expr: up{namespace="foo"}
    labels:
      namespace: foo
`

	givenSpec := monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			monitoringv1.RuleGroup{
				Name: "rule1",
				Rules: []monitoringv1.Rule{
					monitoringv1.Rule{
						Alert: "test_alert1",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "up",
						},
					},
				},
			},
			monitoringv1.RuleGroup{
				Name:                    "rule2",
				PartialResponseStrategy: "warn",
				Rules: []monitoringv1.Rule{
					monitoringv1.Rule{
						Alert: "test_alert2",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "up",
						},
					},
				},
			},
		},
	}

	gotRuleConfig, err := generateContent(givenSpec, "namespace", "foo")
	require.NoError(t, err)

	fmt.Println(gotRuleConfig)
	if wantRuleConfig != gotRuleConfig {
		fmt.Println(pretty.Compare(wantRuleConfig, gotRuleConfig))
		t.Fatal("incorrect rule config generated")
	}

}
