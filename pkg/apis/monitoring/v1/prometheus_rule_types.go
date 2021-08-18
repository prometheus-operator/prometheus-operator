// Copyright 2018 The prometheus-operator Authors
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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PrometheusRuleKind    = "PrometheusRule"
	PrometheusRuleName    = "prometheusrules"
	PrometheusRuleKindKey = "prometheusrule"
)

// PrometheusRuleList is a list of PrometheusRules.
// +k8s:openapi-gen=true
type PrometheusRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Rules
	Items []*PrometheusRule `json:"items"`
}

// PrometheusRule defines recording and alerting rules for a Prometheus instance
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator"
type PrometheusRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired alerting rule definitions for Prometheus.
	Spec PrometheusRuleSpec `json:"spec"`
}

// PrometheusRuleSpec contains specification parameters for a Rule.
// +k8s:openapi-gen=true
type PrometheusRuleSpec struct {
	// Content of Prometheus rule file
	Groups []RuleGroup `json:"groups,omitempty"`
}

// RuleGroup and Rule are copied instead of vendored because the
// upstream Prometheus struct definitions don't have json struct tags.

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
// Note: PartialResponseStrategy is only used by ThanosRuler and will
// be ignored by Prometheus instances.  Valid values for this field are 'warn'
// or 'abort'.  More info: https://github.com/thanos-io/thanos/blob/master/docs/components/rule.md#partial-response
// +k8s:openapi-gen=true
type RuleGroup struct {
	Name                    string `json:"name"`
	Interval                string `json:"interval,omitempty"`
	Rules                   []Rule `json:"rules"`
	PartialResponseStrategy string `json:"partial_response_strategy,omitempty"`
}

// Rule describes an alerting or recording rule
// See Prometheus documentation: [alerting](https://www.prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) or [recording](https://www.prometheus.io/docs/prometheus/latest/configuration/recording_rules/#recording-rules) rule
// +k8s:openapi-gen=true
type Rule struct {
	Record      string             `json:"record,omitempty"`
	Alert       string             `json:"alert,omitempty"`
	Expr        intstr.IntOrString `json:"expr"`
	For         string             `json:"for,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (f *PrometheusRule) DeepCopyObject() runtime.Object {
	return f.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusRuleList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
