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

func (c *Operator) migrateRuleConfigMapsToRuleFileCRDs(p *monitoringv1.Prometheus) error {
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

	ruleFiles := []monitoringv1.RuleFile{}
	for _, cm := range configMaps {
		files, err := cmToRuleFiles(cm)
		if err != nil {
			return err
		}
		ruleFiles = append(ruleFiles, files...)
	}

	ruleFileNames := []string{}
	for _, file := range configMaps {
		ruleFileNames = append(ruleFileNames, file.Name)
	}
	level.Debug(c.logger).Log(
		"msg", "rule files to be created",
		"rulefiles", strings.Join(ruleFileNames, ","),
		"namespace", p.Namespace,
		"prometheus", p.Name,
	)

	for _, ruleFile := range ruleFiles {
		_, err := c.mclient.MonitoringV1().RuleFiles(p.Namespace).Create(&ruleFile)
		if apierrors.IsAlreadyExists(err) {
			level.Debug(c.logger).Log(
				"msg", "rule file already exists for configmap key",
				"rulefilename", ruleFile.Name,
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

func cmToRuleFiles(cm *v1.ConfigMap) ([]monitoringv1.RuleFile, error) {
	ruleFiles := []monitoringv1.RuleFile{}

	for name, content := range cm.Data {
		ruleFileSpec := monitoringv1.RuleFileSpec{}

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(content), 1000).Decode(&ruleFileSpec); err != nil {
			return []monitoringv1.RuleFile{}, errors.Wrapf(
				err,
				"unmarshal rules file %v in  configmap '%v' in namespace '%v'",
				name, cm.Name, cm.Namespace,
			)
		}

		ruleFile := monitoringv1.RuleFile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cm.Name + "-" + name,
				Namespace: cm.Namespace,
				Labels:    cm.Labels,
			},
			Spec: ruleFileSpec,
		}

		ruleFiles = append(ruleFiles, ruleFile)
	}

	return ruleFiles, nil
}
