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
	"bytes"
	"strings"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/cache"

	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

func (c *Operator) migrateRuleConfigMapsToRuleCRDs(p *monitoringv1.Prometheus) error {
	configMaps, err := c.getRuleCMs(p.Namespace, p.Spec.RuleSelector)
	if err != nil {
		return err
	}

	configMapNames := []string{}
	for _, cm := range configMaps {
		configMapNames = append(configMapNames, cm.Name)
	}
	level.Debug(c.logger).Log(
		"msg", "selected rule configmaps for migration",
		"configmaps", strings.Join(configMapNames, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	rules := []monitoringv1.PrometheusRule{}
	for _, cm := range configMaps {
		cmRules, err := CMToRule(cm)
		if err != nil {
			return err
		}
		rules = append(rules, cmRules...)
	}

	ruleNames := []string{}
	for _, rule := range configMaps {
		ruleNames = append(ruleNames, rule.Name)
	}
	level.Debug(c.logger).Log(
		"msg", "rule files to be created",
		"rules", strings.Join(ruleNames, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	for _, rule := range rules {
		_, err := c.mclient.MonitoringV1().PrometheusRules(p.Namespace).Create(&rule)
		if apierrors.IsAlreadyExists(err) {
			level.Debug(c.logger).Log(
				"msg", "rule file already exists for configmap key",
				"rule", rule.Name,
				"namespace", p.Namespace,
				"prometheus", p.Name,
			)
		} else if err != nil {
			return err
		}
	}

	level.Debug(c.logger).Log(
		"msg", "rule files created successfully",
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return nil
}

func (c *Operator) getRuleCMs(ns string, cmLabelSelector *metav1.LabelSelector) ([]*v1.ConfigMap, error) {
	cmSelector, err := metav1.LabelSelectorAsSelector(cmLabelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "convert rule file label selector to selector")
	}

	configMaps := []*v1.ConfigMap{}

	err = cache.ListAllByNamespace(c.cmapInf.GetIndexer(), ns, cmSelector, func(obj interface{}) {
		configMaps = append(configMaps, obj.(*v1.ConfigMap))
	})

	return configMaps, nil
}

// CMToRule takes a rule config map and transforms it to possibly multiple
// rule file crds. It is used in `cmd/po-rule-cm-to-rule-file-crds`. Thereby it
// needs to be public.
func CMToRule(cm *v1.ConfigMap) ([]monitoringv1.PrometheusRule, error) {
	rules := []monitoringv1.PrometheusRule{}

	for name, content := range cm.Data {
		ruleSpec := monitoringv1.PrometheusRuleSpec{}

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(content), 1000).Decode(&ruleSpec); err != nil {
			return []monitoringv1.PrometheusRule{}, errors.Wrapf(
				err,
				"unmarshal rules file %v in  configmap '%v' in namespace '%v'",
				name, cm.Name, cm.Namespace,
			)
		}

		rule := monitoringv1.PrometheusRule{
			TypeMeta: metav1.TypeMeta{
				Kind:       monitoringv1.PrometheusRuleKind,
				APIVersion: monitoringv1.Group + "/" + monitoringv1.Version,
			},

			ObjectMeta: metav1.ObjectMeta{
				Name:      cm.Name + "-" + name,
				Namespace: cm.Namespace,
				Labels:    cm.Labels,
			},
			Spec: ruleSpec,
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
