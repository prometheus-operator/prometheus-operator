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
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

func (c *Operator) createOrUpdateRuleConfigMap(p *monitoringv1.Prometheus) error {
	cClient := c.kclient.CoreV1().ConfigMaps(p.Namespace)

	namespaces, err := c.selectRuleNamespaces(p)
	if err != nil {
		return err
	}

	rules, err := c.selectRules(p, namespaces)
	if err != nil {
		return err
	}

	newConfigMap := c.makeRulesConfigMap(p, rules)

	currentConfigMap, err := cClient.Get(prometheusRuleConfigMapName(p.Name), metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	isNotFound := false
	if apierrors.IsNotFound(err) {
		level.Debug(c.logger).Log(
			"msg", "no PrometheusRule configmap created yet",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		isNotFound = true
	}

	newChecksum := checksumRules(rules)
	currentChecksum := checksumRules(currentConfigMap.Data)

	if newChecksum == currentChecksum && !isNotFound {
		level.Debug(c.logger).Log(
			"msg", "no PrometheusRule changes",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		return nil
	}

	if isNotFound {
		level.Debug(c.logger).Log(
			"msg", "no PrometheusRule found, creating new one",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		_, err = cClient.Create(newConfigMap)
	} else {
		level.Debug(c.logger).Log(
			"msg", "updating PrometheusRule",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		_, err = cClient.Update(newConfigMap)
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Operator) selectRuleNamespaces(p *monitoringv1.Prometheus) ([]string, error) {
	namespaces := []string{}

	// If 'RuleNamespaceSelector' is nil, only check own namespace.
	if p.Spec.RuleNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		ruleNamespaceSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleNamespaceSelector)
		if err != nil {
			return namespaces, errors.Wrap(err, "convert rule namespace label selector to selector")
		}

		cache.ListAll(c.nsInf.GetStore(), ruleNamespaceSelector, func(obj interface{}) {
			namespaces = append(namespaces, obj.(*v1.Namespace).Name)
		})
	}

	level.Debug(c.logger).Log(
		"msg", "selected RuleNamespaces",
		"namespaces", strings.Join(namespaces, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return namespaces, nil
}

func (c *Operator) selectRules(p *monitoringv1.Prometheus, namespaces []string) (map[string]string, error) {
	rules := map[string]string{}

	ruleSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleSelector)
	if err != nil {
		return rules, errors.Wrap(err, "convert rule label selector to selector")
	}

	for _, ns := range namespaces {
		var marshalErr error
		err := cache.ListAllByNamespace(c.ruleInf.GetIndexer(), ns, ruleSelector, func(obj interface{}) {
			rule := obj.(*monitoringv1.PrometheusRule)
			content, err := yaml.Marshal(rule.Spec)
			if err != nil {
				marshalErr = err
				return
			}
			rules[fmt.Sprintf("%v-%v.yaml", rule.Namespace, rule.Name)] = string(content)
		})
		if err != nil {
			return nil, err
		}
		if marshalErr != nil {
			return nil, marshalErr
		}
	}

	// sort rules map
	rulenames := []string{}
	for k := range rules {
		rulenames = append(rulenames, k)
	}
	sort.Strings(rulenames)
	sortedRules := map[string]string{}
	for _, name := range rulenames {
		sortedRules[name] = rules[name]
	}

	level.Debug(c.logger).Log(
		"msg", "selected Rules",
		"rules", strings.Join(rulenames, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return sortedRules, nil
}

func (c *Operator) makeRulesConfigMap(p *monitoringv1.Prometheus, ruleFiles map[string]string) *v1.ConfigMap {
	boolTrue := true
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   prometheusRuleConfigMapName(p.Name),
			Labels: managedByOperatorLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         p.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               p.Kind,
					Name:               p.Name,
					UID:                p.UID,
				},
			},
		},
		Data: ruleFiles,
	}
}

func checksumRules(files map[string]string) string {
	var sum string
	for name, value := range files {
		sum = sum + name + value
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(sum)))
}

func prometheusRuleConfigMapName(prometheusName string) string {
	return "prometheus-" + prometheusName + "-rulefiles"
}
