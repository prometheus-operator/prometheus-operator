// Copyright 2018
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

package crdvalidation

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Config stores the user configuration input
type Config struct {
	SpecDefinitionName    string
	EnableValidation      bool
	OutputFormat          string
	Labels                Labels
	Annotations           Labels
	ResourceScope         string
	Group                 string
	Kind                  string
	Version               string
	Plural                string
	Categories            []string
	ShortNames            []string
	GetOpenAPIDefinitions GetAPIDefinitions
}

type Labels struct {
	LabelsString string
	LabelsMap    map[string]string
}

// Implement the flag.Value interface
func (labels *Labels) String() string {
	return labels.LabelsString
}

// Merge labels create a new map with labels merged.
func (labels *Labels) Merge(otherLabels map[string]string) map[string]string {
	mergedLabels := map[string]string{}

	for key, value := range otherLabels {
		mergedLabels[key] = value
	}

	for key, value := range labels.LabelsMap {
		mergedLabels[key] = value
	}
	return mergedLabels
}

// Implement the flag.Set interface
func (labels *Labels) Set(value string) error {
	m := map[string]string{}
	if value != "" {
		splited := strings.Split(value, ",")
		for _, pair := range splited {
			sp := strings.Split(pair, "=")
			m[sp[0]] = sp[1]
		}
	}
	(*labels).LabelsMap = m
	(*labels).LabelsString = value
	return nil
}

func NewCustomResourceDefinition(config Config) *extensionsobj.CustomResourceDefinition {
	crd := &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.Plural + "." + config.Group,
			Labels:      config.Labels.LabelsMap,
			Annotations: config.Annotations.LabelsMap,
		},
		TypeMeta: CustomResourceDefinitionTypeMeta,
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   config.Group,
			Version: config.Version,
			Scope:   extensionsobj.ResourceScope(config.ResourceScope),
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     config.Plural,
				Kind:       config.Kind,
				Categories: config.Categories,
				ShortNames: config.ShortNames,
			},
		},
	}

	if config.SpecDefinitionName != "" && config.EnableValidation == true {
		crd.Spec.Validation = GetCustomResourceValidation(config.SpecDefinitionName, config.GetOpenAPIDefinitions)
	}

	return crd
}

func MarshallCrd(crd *extensionsobj.CustomResourceDefinition, outputFormat string) error {
	jsonBytes, err := json.Marshal(crd)
	if err != nil {
		return err
	}

	var r unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &r.Object); err != nil {
		return err
	}

	unstructured.RemoveNestedField(r.Object, "status")

	jsonBytes, err = json.MarshalIndent(r.Object, "", "    ")
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		_, err = os.Stdout.Write(jsonBytes)
		if err != nil {
			return err
		}
	} else {
		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write([]byte("---\n"))
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(yamlBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// InitFlags prepares command line flags parser
func InitFlags(cfg *Config, flagset *flag.FlagSet) *flag.FlagSet {
	flagset.Var(&cfg.Labels, "labels", "Labels")
	flagset.Var(&cfg.Annotations, "annotations", "Annotations")
	flagset.BoolVar(&cfg.EnableValidation, "with-validation", true, "Add CRD validation field, default: true")
	flagset.StringVar(&cfg.Group, "apigroup", "custom.example.com", "CRD api group")
	flagset.StringVar(&cfg.SpecDefinitionName, "spec-name", "", "CRD spec definition name")
	flagset.StringVar(&cfg.OutputFormat, "output", "yaml", "output format: json|yaml")
	flagset.StringVar(&cfg.Kind, "kind", "", "CRD Kind")
	flagset.StringVar(&cfg.ResourceScope, "scope", string(extensionsobj.NamespaceScoped), "CRD scope: 'Namespaced' | 'Cluster'.  Default: Namespaced")
	flagset.StringVar(&cfg.Version, "version", "v1", "CRD version, default: 'v1'")
	flagset.StringVar(&cfg.Plural, "plural", "", "CRD plural name")
	return flagset
}
