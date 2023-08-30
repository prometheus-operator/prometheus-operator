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
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

// The maximum `Data` size of a ConfigMap seems to differ between
// environments. This is probably due to different meta data sizes which count
// into the overall maximum size of a ConfigMap. Thereby lets leave a
// large buffer.
var maxConfigMapDataSize = int(float64(v1.MaxSecretSize) * 0.5)

func (c *Operator) createOrUpdateRuleConfigMaps(ctx context.Context, p *monitoringv1.Prometheus) ([]string, error) {
	cClient := c.kclient.CoreV1().ConfigMaps(p.Namespace)

	namespaces, err := c.selectRuleNamespaces(p)
	if err != nil {
		return nil, err
	}

	excludedFromEnforcement := p.Spec.ExcludedFromEnforcement
	// append the deprecated PrometheusRulesExcludedFromEnforce
	for _, rule := range p.Spec.PrometheusRulesExcludedFromEnforce {
		excludedFromEnforcement = append(excludedFromEnforcement,
			monitoringv1.ObjectReference{
				Namespace: rule.RuleNamespace,
				Group:     monitoring.GroupName,
				Resource:  monitoringv1.PrometheusRuleName,
				Name:      rule.RuleName,
			})
	}
	nsLabeler := namespacelabeler.New(
		p.Spec.EnforcedNamespaceLabel,
		excludedFromEnforcement,
		true,
	)

	logger := log.With(c.logger, "prometheus", p.Name, "namespace", p.Namespace)
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)

	promRuleSelector, err := operator.NewPrometheusRuleSelector(operator.PrometheusFormat, promVersion, p.Spec.RuleSelector, nsLabeler, c.ruleInfs, logger)
	if err != nil {
		return nil, errors.Wrap(err, "initializing PrometheusRules failed")
	}

	newRules, rejected, err := promRuleSelector.Select(namespaces)
	if err != nil {
		return nil, errors.Wrap(err, "selecting PrometheusRules failed")
	}

	if pKey, ok := c.accessor.MetaNamespaceKey(p); ok {
		c.metrics.SetSelectedResources(pKey, monitoringv1.PrometheusRuleKind, len(newRules))
		c.metrics.SetRejectedResources(pKey, monitoringv1.PrometheusRuleKind, rejected)
	}

	currentConfigMapList, err := cClient.List(ctx, prometheusRulesConfigMapSelector(p.Name))
	if err != nil {
		return nil, err
	}
	currentConfigMaps := currentConfigMapList.Items

	currentRules := map[string]string{}
	for _, cm := range currentConfigMaps {
		for ruleFileName, ruleFile := range cm.Data {
			currentRules[ruleFileName] = ruleFile
		}
	}

	equal := reflect.DeepEqual(newRules, currentRules)
	if equal && len(currentConfigMaps) != 0 {
		level.Debug(c.logger).Log(
			"msg", "no PrometheusRule changes",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		currentConfigMapNames := []string{}
		for _, cm := range currentConfigMaps {
			currentConfigMapNames = append(currentConfigMapNames, cm.Name)
		}
		return currentConfigMapNames, nil
	}

	newConfigMaps, err := makeRulesConfigMaps(p, newRules)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make rules ConfigMaps")
	}

	newConfigMapNames := []string{}
	for _, cm := range newConfigMaps {
		newConfigMapNames = append(newConfigMapNames, cm.Name)
	}

	if len(currentConfigMaps) == 0 {
		level.Debug(c.logger).Log(
			"msg", "no PrometheusRule configmap found, creating new one",
			"namespace", p.Namespace,
			"prometheus", p.Name,
		)
		for _, cm := range newConfigMaps {
			_, err = cClient.Create(ctx, &cm, metav1.CreateOptions{})
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create ConfigMap '%v'", cm.Name)
			}
		}
		return newConfigMapNames, nil
	}

	// Simply deleting old ConfigMaps and creating new ones for now. Could be
	// replaced by logic that only deletes obsolete ConfigMaps in the future.
	for _, cm := range currentConfigMaps {
		err := cClient.Delete(ctx, cm.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to delete current ConfigMap '%v'", cm.Name)
		}
	}

	level.Debug(c.logger).Log(
		"msg", "updating PrometheusRule",
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)
	for _, cm := range newConfigMaps {
		_, err = cClient.Create(ctx, &cm, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create new ConfigMap '%v'", cm.Name)
		}
	}

	return newConfigMapNames, nil
}

func prometheusRulesConfigMapSelector(prometheusName string) metav1.ListOptions {
	return metav1.ListOptions{LabelSelector: fmt.Sprintf("%v=%v", prompkg.LabelPrometheusName, prometheusName)}
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

		namespaces, err = operator.ListMatchingNamespaces(ruleNamespaceSelector, c.nsMonInf)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(c.logger).Log(
		"msg", "selected RuleNamespaces",
		"namespaces", strings.Join(namespaces, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	return namespaces, nil
}

// makeRulesConfigMaps takes a Prometheus configuration and rule files and
// returns a list of Kubernetes ConfigMaps to be later on mounted into the
// Prometheus instance.
// If the total size of rule files exceeds the Kubernetes ConfigMap limit,
// they are split up via the simple first-fit [1] bin packing algorithm. In the
// future this can be replaced by a more sophisticated algorithm, but for now
// simplicity should be sufficient.
// [1] https://en.wikipedia.org/wiki/Bin_packing_problem#First-fit_algorithm
func makeRulesConfigMaps(p *monitoringv1.Prometheus, ruleFiles map[string]string) ([]v1.ConfigMap, error) {
	//check if none of the rule files is too large for a single ConfigMap
	for filename, file := range ruleFiles {
		if len(file) > maxConfigMapDataSize {
			return nil, errors.Errorf(
				"rule file '%v' is too large for a single Kubernetes ConfigMap",
				filename,
			)
		}
	}

	buckets := []map[string]string{
		{},
	}
	currBucketIndex := 0

	// To make bin packing algorithm deterministic, sort ruleFiles filenames and
	// iterate over filenames instead of ruleFiles map (not deterministic).
	fileNames := []string{}
	for n := range ruleFiles {
		fileNames = append(fileNames, n)
	}
	sort.Strings(fileNames)

	for _, filename := range fileNames {
		// If rule file doesn't fit into current bucket, create new bucket.
		if bucketSize(buckets[currBucketIndex])+len(ruleFiles[filename]) > maxConfigMapDataSize {
			buckets = append(buckets, map[string]string{})
			currBucketIndex++
		}
		buckets[currBucketIndex][filename] = ruleFiles[filename]
	}

	ruleFileConfigMaps := []v1.ConfigMap{}
	for i, bucket := range buckets {
		cm := makeRulesConfigMap(p, bucket)
		cm.Name = cm.Name + "-" + strconv.Itoa(i)
		ruleFileConfigMaps = append(ruleFileConfigMaps, cm)
	}

	return ruleFileConfigMaps, nil
}

func bucketSize(bucket map[string]string) int {
	totalSize := 0
	for _, v := range bucket {
		totalSize += len(v)
	}

	return totalSize
}

func makeRulesConfigMap(p *monitoringv1.Prometheus, ruleFiles map[string]string) v1.ConfigMap {
	boolTrue := true

	labels := map[string]string{prompkg.LabelPrometheusName: p.Name}
	for k, v := range prompkg.ManagedByOperatorLabels {
		labels[k] = v
	}

	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   prometheusRuleConfigMapName(p.Name),
			Labels: labels,
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

func prometheusRuleConfigMapName(prometheusName string) string {
	return "prometheus-" + prometheusName + "-rulefiles"
}
