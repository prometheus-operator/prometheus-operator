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

	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func (c *Operator) createOrUpdateRuleFileConfigMap(p *monitoringv1.Prometheus) error {
	cClient := c.kclient.CoreV1().ConfigMaps(p.Namespace)

	namespaces, err := c.selectRuleFileNamespaces(p)
	if err != nil {
		return err
	}

	ruleFiles, err := c.selectRuleFiles(p, namespaces)
	if err != nil {
		return err
	}

	newConfigMap := c.makeRulesConfigMap(p, ruleFiles)

	currentConfigMap, err := cClient.Get(prometheusRuleFilesConfigMapName(p.Name), metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	isNotFound := false
	if apierrors.IsNotFound(err) {
		level.Debug(c.logger).Log(
			"msg", "no RuleFiles configmap created yet",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		isNotFound = true
	}

	newChecksum := checksumRuleFiles(ruleFiles)
	currentChecksum := checksumRuleFiles(currentConfigMap.Data)

	if newChecksum == currentChecksum && !isNotFound {
		level.Debug(c.logger).Log(
			"msg", "no RuleFile changes",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		return nil
	}

	if isNotFound {
		level.Debug(c.logger).Log(
			"msg", "no RuleFile found, creating new one",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		_, err = cClient.Create(newConfigMap)
	} else {
		level.Debug(c.logger).Log(
			"msg", "updating RuleFile",
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

func (c *Operator) selectRuleFileNamespaces(p *monitoringv1.Prometheus) ([]string, error) {
	namespaces := []string{}

	// If 'RuleFilesNamespaceSelector' is nil, only check own namespace.
	if p.Spec.RuleFileNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		ruleFileNamespaceSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleFileNamespaceSelector)
		if err != nil {
			return namespaces, errors.Wrap(err, "convert rule file namespace label selector to selector")
		}

		cache.ListAll(c.nsInf.GetStore(), ruleFileNamespaceSelector, func(obj interface{}) {
			namespaces = append(namespaces, obj.(*v1.Namespace).Name)
		})
	}

	level.Debug(c.logger).Log(
		"msg", "selected RuleFileNamespaces",
		"namespaces", strings.Join(namespaces, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return namespaces, nil
}

func (c *Operator) selectRuleFiles(p *monitoringv1.Prometheus, namespaces []string) (map[string]string, error) {
	ruleFiles := map[string]string{}

	// With Prometheus Operator v0.20.0 the 'RuleSelector' field in the Prometheus
	// CRD Spec is deprecated. Any value in 'RuleSelector' is just copied to the new
	// field 'RuleFileSelector'.
	if p.Spec.RuleFileSelector == nil && p.Spec.RuleSelector != nil {
		p.Spec.RuleFileSelector = p.Spec.RuleSelector
	}

	fileSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleFileSelector)
	if err != nil {
		return ruleFiles, errors.Wrap(err, "convert rule file label selector to selector")
	}

	for _, ns := range namespaces {
		var marshalErr error
		err := cache.ListAllByNamespace(c.ruleFileInf.GetIndexer(), ns, fileSelector, func(obj interface{}) {
			file := obj.(*monitoringv1.RuleFile)
			content, err := yaml.Marshal(file.Spec)
			if err != nil {
				marshalErr = err
				return
			}
			ruleFiles[fmt.Sprintf("%v-%v.yaml", file.Namespace, file.Name)] = string(content)
		})
		if err != nil {
			return nil, err
		}
		if marshalErr != nil {
			return nil, marshalErr
		}
	}

	// sort ruleFiles map
	filenames := []string{}
	for k, _ := range ruleFiles {
		filenames = append(filenames, k)
	}
	sort.Strings(filenames)
	sortedRuleFiles := map[string]string{}
	for _, name := range filenames {
		sortedRuleFiles[name] = ruleFiles[name]
	}

	level.Debug(c.logger).Log(
		"msg", "selected RuleFiles",
		"rulefiles", strings.Join(filenames, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return sortedRuleFiles, nil
}

func (c *Operator) makeRulesConfigMap(p *monitoringv1.Prometheus, ruleFiles map[string]string) *v1.ConfigMap {
	boolTrue := true
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   prometheusRuleFilesConfigMapName(p.Name),
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

func checksumRuleFiles(files map[string]string) string {
	var sum string
	for name, value := range files {
		sum = sum + name + value
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(sum)))
}

func prometheusRuleFilesConfigMapName(prometheusName string) string {
	return "prometheus-" + prometheusName + "-rules"
}
