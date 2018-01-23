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
	Namespace        string
	Labels           prometheus.Labels
	CrdGroup         string
	CrdKinds         monitoringv1.CrdKinds
}

func marshallCrd(crd *extensionsobj.CustomResourceDefinition, encode string) {
	jsonBytes, err := json.MarshalIndent(crd, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}

	if encode == "json" {
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
	promCrd := k8sutil.NewPrometheusCustomResourceDefinition(conf.CrdKinds.Prometheus, conf.CrdGroup, conf.Labels.LabelsMap)
	serviceMonitorCrd := k8sutil.NewServiceMonitorCustomResourceDefinition(conf.CrdKinds.ServiceMonitor, conf.CrdGroup, conf.Labels.LabelsMap)
	alertManagerCrd := k8sutil.NewAlertmanagerCustomResourceDefinition(conf.CrdKinds.Alertmanager, conf.CrdGroup, conf.Labels.LabelsMap)

	marshallCrd(promCrd, "yaml")
	marshallCrd(serviceMonitorCrd, "yaml")
	marshallCrd(alertManagerCrd, "yaml")

}
