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

package e2e

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/wait"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	testFramework "github.com/coreos/prometheus-operator/test/e2e/framework"
)

func TestPrometheusCreateDeleteCluster(t *testing.T) {
	name := "test"

	if err := framework.CreatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 1)); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusScaleUpDownCluster(t *testing.T) {
	name := "test"

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 1)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 2)); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusVersionMigration(t *testing.T) {
	name := "test"

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	p := framework.MakeBasicPrometheus(name, name, 1)

	p.Spec.Version = "v1.5.1"
	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.6.1"
	if err := framework.UpdatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}
	if err := framework.WaitForPrometheusRunImageAndReady(p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.5.1"
	if err := framework.UpdatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}
	if err := framework.WaitForPrometheusRunImageAndReady(p); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadConfig(t *testing.T) {
	name := "test"
	replicas := int32(1)
	p := &v1alpha1.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.PrometheusSpec{
			Replicas: &replicas,
			Version:  "v1.5.0",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("400Mi"),
				},
			},
		},
	}

	firstConfig := `
global:
  scrape_interval: 1m
scrape_configs:
  - job_name: testReloadConfig
    metrics_path: /metrics
    static_configs:
      - targets:
        - 111.111.111.111:9090
`

	cfg := &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
		},
		Data: map[string][]byte{
			"prometheus.yaml": []byte(firstConfig),
		},
	}

	svc := framework.MakeBasicPrometheusNodePortService(name, "reloadconfig-group", 30900)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, svc.Name); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := framework.KubeClient.CoreV1().Secrets(framework.Namespace.Name).Create(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, svc); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(1); err != nil {
		t.Fatal(err)
	}

	secondConfig := `
global:
  scrape_interval: 1m
scrape_configs:
  - job_name: testReloadConfig
    metrics_path: /metrics
    static_configs:
      - targets:
        - 111.111.111.111:9090
        - 111.111.111.112:9090 
`

	cfg.Data["prometheus.yaml"] = []byte(secondConfig)
	if _, err := framework.KubeClient.CoreV1().Secrets(framework.Namespace.Name).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(2); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadRules(t *testing.T) {
	name := "test"

	ruleFileConfigMap := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s-rules", name),
			Labels: map[string]string{
				"role": "rulefile",
			},
		},
		Data: map[string]string{
			"test.rules": "",
		},
	}

	_, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Create(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}

		err = framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Delete(ruleFileConfigMap.Name, nil)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 1)); err != nil {
		t.Fatal(err)
	}

	ruleFileConfigMap.Data["test.rules"] = "# comment to trigger a configmap reload"
	_, err = framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Update(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	// remounting a ConfigMap can take some time
	err = wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		logs, err := testFramework.GetLogs(framework.KubeClient, framework.Namespace.Name, fmt.Sprintf("prometheus-%s-0", name), "prometheus-config-reloader")
		if err != nil {
			return false, err
		}

		if strings.Contains(logs, "ConfigMap modified") && strings.Contains(logs, "Rule files updated") && strings.Contains(logs, "Prometheus successfully reloaded") {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusDiscovery(t *testing.T) {
	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakeBasicPrometheusNodePortService(prometheusName, group, 30900)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(prometheusName); err != nil {
			t.Fatal(err)
		}
		if err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Delete(group, nil); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, svc.Name); err != nil {
			t.Fatal(err)
		}
	}()

	log.Print("Creating Prometheus ServiceMonitor")
	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(prometheusName, group, 1)
	p.Spec.Version = "v1.5.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Prometheus Service")
	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, svc); err != nil {
		t.Fatal(err)
	}

	log.Print("Validating Prometheus config Secret was created")
	_, err := framework.KubeClient.CoreV1().Secrets(framework.Namespace.Name).Get(fmt.Sprintf("prometheus-%s", prometheusName))
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	log.Print("Validating Prometheus Targets were properly discovered")
	err = wait.Poll(time.Second, 18*time.Minute, isDiscoveryWorking(prometheusName))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusAlertmanagerDiscovery(t *testing.T) {
	prometheusName := "test"
	alertmanagerName := "test"
	group := "servicediscovery-test"
	svc := framework.MakeBasicPrometheusNodePortService(prometheusName, group, 30900)
	amsvc := framework.MakeAlertmanagerNodePortService(alertmanagerName, group, 30903)

	defer func() {
		if err := framework.DeleteAlertmanagerAndWaitUntilGone(alertmanagerName); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, amsvc.Name); err != nil {
			t.Fatal(err)
		}
		if err := framework.DeletePrometheusAndWaitUntilGone(prometheusName); err != nil {
			t.Fatal(err)
		}
		if err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Delete(group, nil); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, svc.Name); err != nil {
			t.Fatal(err)
		}
	}()

	p := framework.MakeBasicPrometheus(prometheusName, group, 1)
	framework.AddAlertingToPrometheus(p, alertmanagerName)
	p.Spec.Version = "v1.5.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Prometheus Service")
	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, svc); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Prometheus ServiceMonitor")
	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: ", err)
	}

	log.Print("Validating Prometheus config Secret was created")
	_, err := framework.KubeClient.CoreV1().Secrets(framework.Namespace.Name).Get(fmt.Sprintf("prometheus-%s", prometheusName))
	if err != nil {
		t.Fatalf("Generated Secret could not be retrieved: ", err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Alertmanager Service")
	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, amsvc); err != nil {
		t.Fatal(err)
	}

	log.Print("Validating Prometheus properly discovered alertmanagers")
	err = wait.Poll(time.Second, 18*time.Minute, isAlertmanagerDiscoveryWorking(alertmanagerName))
	if err != nil {
		t.Fatal(err)
	}
}

func TestExposingPrometheusWithNodePort(t *testing.T) {
	basicPrometheus := framework.MakeBasicPrometheus("test", "test", 1)
	service := framework.MakeBasicPrometheusNodePortService(basicPrometheus.Name, "nodeport-service", 30900)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(basicPrometheus.Name); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, service.Name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:30900/metrics", framework.ClusterIP))
	if err != nil {
		t.Fatal("Retrieving prometheus metrics failed with error: ", err)
	} else if resp.StatusCode != 200 {
		t.Fatal("Retrieving prometheus metrics failed with http status code: ", resp.StatusCode)
	}
}

func TestExposingPrometheusWithKubernetesAPI(t *testing.T) {
	basicPrometheus := framework.MakeBasicPrometheus("basic-prometheus", "test-group", 1)
	service := framework.MakePrometheusService(basicPrometheus.Name, "test-group", v1.ServiceTypeClusterIP)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(basicPrometheus.Name); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, service.Name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	ProxyGet := framework.KubeClient.CoreV1().Services(framework.Namespace.Name).ProxyGet
	request := ProxyGet("", service.Name, "web", "/metrics", make(map[string]string))
	_, err := request.DoRaw()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExposingPrometheusWithIngress(t *testing.T) {
	prometheus := framework.MakeBasicPrometheus("main", "test-group", 1)
	prometheusService := framework.MakePrometheusService(prometheus.Name, "test-group", v1.ServiceTypeClusterIP)
	ingress := testFramework.MakeBasicIngress(prometheusService.Name, 9090)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(prometheus.Name); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteService(framework.KubeClient, framework.Namespace.Name, prometheusService.Name); err != nil {
			t.Fatal(err)
		}
		if err := framework.KubeClient.Extensions().Ingresses(framework.Namespace.Name).Delete(ingress.Name, nil); err != nil {
			t.Fatal(err)
		}
		if err := testFramework.DeleteNginxIngressControllerIncDefaultBackend(framework.KubeClient, framework.Namespace.Name); err != nil {
			t.Fatal(err)
		}
	}()

	err := testFramework.SetupNginxIngressControllerIncDefaultBackend(framework.KubeClient, framework.Namespace.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.CreatePrometheusAndWaitUntilReady(prometheus)
	if err != nil {
		t.Fatal(err)
	}

	if err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, framework.Namespace.Name, prometheusService); err != nil {
		t.Fatal(err)
	}

	err = testFramework.CreateIngress(framework.KubeClient, framework.Namespace.Name, ingress)
	if err != nil {
		t.Fatal(err)
	}

	ip, err := testFramework.GetIngressIP(framework.KubeClient, framework.Namespace.Name, ingress.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = testFramework.WaitForHTTPSuccessStatusCode(time.Minute, fmt.Sprintf("http://%s:/metrics", *ip))
	if err != nil {
		t.Fatal(err)
	}
}
func isDiscoveryWorking(prometheusName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(framework.Namespace.Name).List(prometheus.ListOptions(prometheusName))
		if err != nil {
			return false, err
		}
		if 1 != len(pods.Items) {
			return false, nil
		}
		podIP := pods.Items[0].Status.PodIP
		expectedTargets := []string{fmt.Sprintf("http://%s:9090/metrics", podIP)}

		activeTargets, err := framework.GetActiveTargets()
		if err != nil {
			return false, err
		}

		if !assertExpectedTargets(activeTargets, expectedTargets) {
			return false, nil
		}

		working, err := basicQueryWorking()
		if err != nil {
			return false, err
		}
		if !working {
			return false, nil
		}

		return true, nil
	}
}

type resultVector struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

type queryResult struct {
	ResultType string          `json:"resultType"`
	Result     []*resultVector `json:"result"`
}

type prometheusQueryAPIResponse struct {
	Status string       `json:"status"`
	Data   *queryResult `json:"data"`
}

func basicQueryWorking() (bool, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:30900/api/v1/query?query=up", framework.ClusterIP))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	rq := prometheusQueryAPIResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&rq); err != nil {
		return false, err
	}

	if rq.Status != "success" && rq.Data.Result[0].Value[1] == "1" {
		log.Printf("Query Response not successful.")
		return false, nil
	}

	return true, nil
}

func isAlertmanagerDiscoveryWorking(alertmanagerName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(framework.Namespace.Name).List(alertmanager.ListOptions(alertmanagerName))
		if err != nil {
			return false, err
		}
		if 3 != len(pods.Items) {
			return false, nil
		}
		expectedAlertmanagerTargets := []string{}
		for _, p := range pods.Items {
			expectedAlertmanagerTargets = append(expectedAlertmanagerTargets, fmt.Sprintf("http://%s:9093/api/v1/alerts", p.Status.PodIP))
		}

		resp, err := http.Get(fmt.Sprintf("http://%s:30900/api/v1/alertmanagers", framework.ClusterIP))
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&ra); err != nil {
			return false, err
		}

		if assertExpectedAlertmanagerTargets(ra.Data.ActiveAlertmanagers, expectedAlertmanagerTargets) {
			return true, nil
		}

		return false, nil
	}
}

func assertExpectedTargets(targets []*testFramework.Target, expectedTargets []string) bool {
	log.Printf("Expected Targets: %#+v\n", expectedTargets)

	existingTargets := []string{}

	for _, t := range targets {
		existingTargets = append(existingTargets, t.ScrapeURL)
	}

	sort.Strings(expectedTargets)
	sort.Strings(existingTargets)

	if !reflect.DeepEqual(expectedTargets, existingTargets) {
		log.Printf("Existing Targets: %#+v\n", existingTargets)
		return false
	}

	return true
}

func assertExpectedAlertmanagerTargets(ams []*alertmanagerTarget, expectedTargets []string) bool {
	log.Printf("Expected Alertmanager Targets: %#+v\n", expectedTargets)

	existingTargets := []string{}

	for _, am := range ams {
		existingTargets = append(existingTargets, am.URL)
	}

	sort.Strings(expectedTargets)
	sort.Strings(existingTargets)

	if !reflect.DeepEqual(expectedTargets, existingTargets) {
		log.Printf("Existing Alertmanager Targets: %#+v\n", existingTargets)
		return false
	}

	return true
}

type alertmanagerTarget struct {
	URL string `json:"url"`
}

type alertmanagerDiscovery struct {
	ActiveAlertmanagers []*alertmanagerTarget `json:"activeAlertmanagers"`
}

type prometheusAlertmanagerAPIResponse struct {
	Status string                 `json:"status"`
	Data   *alertmanagerDiscovery `json:"data"`
}
