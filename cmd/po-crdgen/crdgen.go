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

package main

import (
	"encoding/json"
	"fmt"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	k8sutil "github.com/coreos/prometheus-operator/pkg/k8sutil"
	prometheus "github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/ghodss/yaml"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"os"
)

// Config stores the user configuration input
type Config struct {
	EnableValidation bool
	OutputFormat     string
	Namespace        string
	Labels           prometheus.Labels
	CrdGroup         string
	CrdKinds         monitoringv1.CrdKinds
}

func marshallCrd(crd *extensionsobj.CustomResourceDefinition, outputFormat string) {
	jsonBytes, err := json.MarshalIndent(crd, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}

	if outputFormat == "json" {
		os.Stdout.Write(jsonBytes)
	} else {
		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		if err != nil {
			fmt.Println("error:", err)
		}
		os.Stdout.Write([]byte("---\n"))
		os.Stdout.Write(yamlBytes)
	}
}

// PrintCrd write on stdout the Prometheus CRD files as json or yaml
func PrintCrd(conf Config) {
	promCrd := k8sutil.NewPrometheusCustomResourceDefinition(conf.CrdKinds.Prometheus, conf.CrdGroup, conf.Labels.LabelsMap, conf.EnableValidation)
	serviceMonitorCrd := k8sutil.NewServiceMonitorCustomResourceDefinition(conf.CrdKinds.ServiceMonitor, conf.CrdGroup, conf.Labels.LabelsMap, conf.EnableValidation)
	alertManagerCrd := k8sutil.NewAlertmanagerCustomResourceDefinition(conf.CrdKinds.Alertmanager, conf.CrdGroup, conf.Labels.LabelsMap, conf.EnableValidation)

	marshallCrd(promCrd, conf.OutputFormat)
	marshallCrd(serviceMonitorCrd, conf.OutputFormat)
	marshallCrd(alertManagerCrd, conf.OutputFormat)

}
