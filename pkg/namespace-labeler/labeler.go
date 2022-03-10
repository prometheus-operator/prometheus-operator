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
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"
	"github.com/prometheus-community/prom-label-proxy/injectproxy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Labeler enables to enforce adding namespace labels to PrometheusRules and to metrics used in them
type Labeler struct {
	enforcedNsLabel                  string
	prometheusRuleLabeler            bool
	excludeNamespaceGroupKindToNames map[namespaceGroupKind][]string
}

type namespaceGroupKind struct {
	namespace string
	groupKind string
}

// New - creates new Labeler
// enforcedNsLabel - label name to be enforced for namespace
// excludeConfig - list of ObjectReference to be excluded while enforcing adding namespace label
// prometheusRuleLabeler - whether this should apply for Prometheus or Thanos rules
func New(enforcedNsLabel string, excludeConfig []monitoringv1.ObjectReference, prometheusRuleLabeler bool) *Labeler {

	if enforcedNsLabel == "" {
		return &Labeler{} // no-op labeler
	}

	if len(excludeConfig) == 0 {
		return &Labeler{enforcedNsLabel: enforcedNsLabel}
	}

	objectsExcludeList := make(map[namespaceGroupKind][]string)
	for _, r := range excludeConfig {
		if r.Namespace == "" || r.GroupKind().Empty() || r.GroupKind().Kind == "" {
			continue
		}
		namespaceGroupKind := namespaceGroupKind{
			namespace: r.Namespace,
			groupKind: r.GroupKind().String()}
		objectsExcludeList[namespaceGroupKind] = append(objectsExcludeList[namespaceGroupKind], r.Name)
	}

	return &Labeler{
		excludeNamespaceGroupKindToNames: objectsExcludeList,
		enforcedNsLabel:                  enforcedNsLabel,
		prometheusRuleLabeler:            prometheusRuleLabeler,
	}
}

func (l *Labeler) GetEnforcedNamespaceLabel() string {
	return l.enforcedNsLabel
}

// IsExcluded returns true if the specified object is excluded from namespace enforcement,
// false otherwise
func (l *Labeler) IsExcluded(prometheusTypeMeta metav1.TypeMeta, prometheusObjectMeta metav1.ObjectMeta) bool {
	if l.enforcedNsLabel == "" {
		return true
	}
	if len(l.excludeNamespaceGroupKindToNames) == 0 {
		return false
	}
	namespaceGroupKind := namespaceGroupKind{
		namespace: prometheusObjectMeta.Namespace,
		groupKind: prometheusTypeMeta.GroupVersionKind().GroupKind().String(),
	}
	for _, name := range l.excludeNamespaceGroupKindToNames[namespaceGroupKind] {
		if name == "" || prometheusObjectMeta.Name == name {
			return true
		}
	}
	return false
}

// EnforceNamespaceLabel - adds(or modifies) namespace label to promRule labels with specified namespace
// and also adds namespace label to all the metrics used in promRule
func (l *Labeler) EnforceNamespaceLabel(rule *monitoringv1.PrometheusRule) error {

	if l.enforcedNsLabel == "" || l.IsExcluded(rule.TypeMeta, rule.ObjectMeta) {
		return nil
	}

	for gi, group := range rule.Spec.Groups {
		if l.prometheusRuleLabeler {
			group.PartialResponseStrategy = ""
		}
		for ri, r := range group.Rules {
			if len(rule.Spec.Groups[gi].Rules[ri].Labels) == 0 {
				rule.Spec.Groups[gi].Rules[ri].Labels = map[string]string{}
			}
			rule.Spec.Groups[gi].Rules[ri].Labels[l.enforcedNsLabel] = rule.Namespace

			expr := r.Expr.String()
			parsedExpr, err := parser.ParseExpr(expr)
			if err != nil {
				return errors.Wrap(err, "failed to parse promql expression")
			}
			enforcer := injectproxy.NewEnforcer(false, &labels.Matcher{
				Name:  l.enforcedNsLabel,
				Type:  labels.MatchEqual,
				Value: rule.Namespace,
			})
			err = enforcer.EnforceNode(parsedExpr)
			if err != nil {
				return errors.Wrap(err, "failed to inject labels to expression")
			}

			rule.Spec.Groups[gi].Rules[ri].Expr = intstr.FromString(parsedExpr.String())
		}
	}
	return nil
}

// GetRelabelingConfigs - append the namespace enforcement relabeling rule
func (l *Labeler) GetRelabelingConfigs(monitorTypeMeta metav1.TypeMeta, monitorObjectMeta metav1.ObjectMeta, rc []*monitoringv1.RelabelConfig) []*monitoringv1.RelabelConfig {

	if l.IsExcluded(monitorTypeMeta, monitorObjectMeta) {
		return rc
	}

	// Because of security risks, whenever enforcedNamespaceLabel is set, we want to append it to the
	// relabel configurations as the last relabeling, to ensure it overrides any other relabelings.
	return append(rc,
		&monitoringv1.RelabelConfig{
			TargetLabel: l.GetEnforcedNamespaceLabel(),
			Replacement: monitorObjectMeta.GetNamespace(),
		},
	)
}
