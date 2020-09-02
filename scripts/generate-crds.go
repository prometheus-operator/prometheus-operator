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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type crdName struct {
	Singular string
	Plural   string
}

type crdGenerator struct {
	ControllerGenOpts string
	YAMLDir           string
	JsonnetDir        string
}

var (
	controllergen string
	gojsontoyaml  string

	crdAPIGroup    = "monitoring.coreos.com"
	controllerPath = "./pkg/apis/monitoring/v1"

	crdNames = []crdName{
		{"alertmanager", "alertmanagers"},
		{"podmonitor", "podmonitors"},
		{"probe", "probes"},
		{"prometheus", "prometheuses"},
		{"prometheusrule", "prometheusrules"},
		{"servicemonitor", "servicemonitors"},
		{"thanosruler", "thanosrulers"},
	}

	crdGenerators = []crdGenerator{
		{
			ControllerGenOpts: "crd:crdVersions=v1",
			YAMLDir:           "./example/prometheus-operator-crd",
			JsonnetDir:        "./jsonnet/prometheus-operator",
		},
	}
)

func (generator crdGenerator) generateYAMLManifests() error {
	outputDir, err := filepath.Abs(generator.YAMLDir)
	if err != nil {
		return errors.Wrapf(err, "absolute CRD output path %s", generator.YAMLDir)
	}
	cmd := exec.Command(controllergen,
		generator.ControllerGenOpts,
		"paths=.",
		"output:crd:dir="+outputDir,
	)
	cmd.Dir, err = filepath.Abs(controllerPath)
	if err != nil {
		return errors.Wrapf(err, "absolute controller path %s", controllerPath)
	}
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running %s", cmd)
	}

	prometheusManifest := fmt.Sprintf("%s/%s_%s.yaml", generator.YAMLDir, crdAPIGroup, "prometheus")
	data, err := ioutil.ReadFile(prometheusManifest)
	if err != nil {
		return errors.Wrapf(err, "reading %s", prometheusManifest)
	}
	data = bytes.ReplaceAll(data, []byte("plural: prometheus"), []byte("plural: prometheuses"))
	data = bytes.ReplaceAll(data, []byte("prometheus.monitoring.coreos.com"), []byte("prometheuses.monitoring.coreos.com"))

	prometheusesManifest := fmt.Sprintf("%s/%s_%s.yaml", generator.YAMLDir, crdAPIGroup, "prometheuses")
	err = ioutil.WriteFile(prometheusesManifest, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "generating %s", prometheusesManifest)
	}

	err = os.Remove(prometheusManifest)
	if err != nil {
		return errors.Wrapf(err, "removing %s", prometheusManifest)
	}
	return nil
}

func (generator crdGenerator) generateJsonnetManifests() error {
	for _, crdName := range crdNames {
		yamlFile := fmt.Sprintf("%s/%s_%s.yaml", generator.YAMLDir, crdAPIGroup, crdName.Plural)
		yamlData, err := ioutil.ReadFile(yamlFile)
		if err != nil {
			return errors.Wrapf(err, "reading %s", yamlFile)
		}

		cmd := exec.Command(gojsontoyaml, "-yamltojson")
		cmd.Stdin = strings.NewReader(string(yamlData))
		var jsonnetData bytes.Buffer
		cmd.Stdout = &jsonnetData
		err = cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "running %s on %s", cmd, yamlFile)
		}

		jsonnetFile := fmt.Sprintf("%s/%s-crd.libsonnet", generator.JsonnetDir, crdName.Singular)
		err = ioutil.WriteFile(jsonnetFile, jsonnetData.Bytes(), 0644)
		if err != nil {
			return errors.Wrapf(err, "generating %s", jsonnetFile)
		}
	}
	return nil
}

func main() {
	flag.StringVar(&controllergen, "controller-gen", "controller-gen", "controller-gen binary path")
	flag.StringVar(&gojsontoyaml, "gojsontoyaml", "gojsontoyaml", "gojsontoyaml binary path")
	flag.Parse()

	for _, generator := range crdGenerators {
		err := generator.generateYAMLManifests()
		if err != nil {
			log.Fatalf("generating YAML manifests: %v", err)
		}
		err = generator.generateJsonnetManifests()
		if err != nil {
			log.Fatalf("generating Jsonnet manifests: %v", err)
		}
	}
}
