package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	files := os.Args[1:]

	for _, filename := range files {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		var meta metav1.TypeMeta

		err = yaml.Unmarshal(content, &meta)
		if err != nil {
			log.Fatal(err)
		}

		switch meta.Kind {
		case v1.AlertmanagersKind:
			var alertmanager v1.Alertmanager
			err := yaml.Unmarshal(content, &alertmanager)
			if err != nil {
				log.Fatal(fmt.Errorf("alertmanager is invalid: %v", err))
			}
		case v1.PrometheusesKind:
			var prometheus v1.Prometheus
			err := yaml.Unmarshal(content, &prometheus)
			if err != nil {
				log.Fatal(fmt.Errorf("prometheus is invalid: %v", err))
			}
		case v1.PrometheusRuleKind:
			var rule v1.PrometheusRule
			err := yaml.Unmarshal(content, &rule)
			if err != nil {
				log.Fatal(fmt.Errorf("prometheus rule is invalid: %v", err))
			}
		case v1.ServiceMonitorsKind:
			var serviceMonitor v1.ServiceMonitor
			err := yaml.Unmarshal(content, &serviceMonitor)
			if err != nil {
				log.Fatal(fmt.Errorf("serviceMonitor is invalid: %v", err))
			}
		default:
			log.Fatal("MetaType is unknown to linter. Not in Alertmanager, Prometheus, PrometheusRule, ServiceMonitor")
		}
	}
}
