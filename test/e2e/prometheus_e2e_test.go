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

	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
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

	p.Spec.Version = "v1.4.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.4.1"
	if err := framework.UpdatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.4.0"
	if err := framework.UpdatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadConfig(t *testing.T) {
	name := "test"

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	cfg := &v1.ConfigMap{
		ObjectMeta: apimetav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
		},
		Data: map[string]string{
			"prometheus.yaml": "",
		},
	}

	if _, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Create(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 1)); err != nil {
		t.Fatal(err)
	}

	cfg.Data["prometheus.yaml"] = "global:\n  scrape_interval: 1m"
	if _, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Update(cfg); err != nil {
		t.Fatal(err)
	}

	// remounting a ConfigMap can take some time
	err := framework.Poll(time.Minute*5, time.Second, func() (bool, error) {
		logs, err := framework.GetLogs(fmt.Sprintf("prometheus-%s-0", name), "config-reloader")
		if err != nil {
			return false, err
		}

		if strings.Contains(logs, "config map updated") && strings.Contains(logs, "successfully triggered reload") {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadRules(t *testing.T) {
	name := "test"

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(framework.MakeBasicPrometheus(name, name, 1)); err != nil {
		t.Fatal(err)
	}

	_, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Update(&v1.ConfigMap{
		ObjectMeta: apimetav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s-rules", name),
		},
		Data: map[string]string{
			"test.rules": "",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// remounting a ConfigMap can take some time
	err = framework.Poll(time.Minute*5, time.Second, func() (bool, error) {
		logs, err := framework.GetLogs(fmt.Sprintf("prometheus-%s-0", name), "config-reloader")
		if err != nil {
			return false, err
		}

		if strings.Contains(logs, "config map updated") && strings.Contains(logs, "successfully triggered reload") {
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
		if err := framework.DeleteService(svc.Name); err != nil {
			t.Fatal(err)
		}
	}()

	log.Print("Creating Prometheus ServiceMonitor")
	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(prometheusName, group, 1)
	p.Spec.Version = "v1.5.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(p); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Prometheus Service")
	if err := framework.CreateServiceAndWaitUntilReady(svc); err != nil {
		t.Fatal(err)
	}

	log.Print("Validating Prometheus ConfigMap was created")
	_, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Get(fmt.Sprintf("prometheus-%s", prometheusName), apimetav1.GetOptions{})
	if err != nil {
		t.Fatalf("Generated ConfigMap could not be retrieved: ", err)
	}

	log.Print("Validating Prometheus Targets were properly discovered")
	err = framework.Poll(18*time.Minute, time.Second, isDiscoveryWorking(prometheusName))
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
		if err := framework.DeleteService(amsvc.Name); err != nil {
			t.Fatal(err)
		}
		if err := framework.DeletePrometheusAndWaitUntilGone(prometheusName); err != nil {
			t.Fatal(err)
		}
		if err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Delete(group, nil); err != nil {
			t.Fatal(err)
		}
		if err := framework.DeleteService(svc.Name); err != nil {
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
	if err := framework.CreateServiceAndWaitUntilReady(svc); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Prometheus ServiceMonitor")
	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(framework.Namespace.Name).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: ", err)
	}

	log.Print("Validating Prometheus ConfigMap was created")
	_, err := framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Get(fmt.Sprintf("prometheus-%s", prometheusName), apimetav1.GetOptions{})
	if err != nil {
		t.Fatalf("Generated ConfigMap could not be retrieved: ", err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
		t.Fatal(err)
	}

	log.Print("Creating Alertmanager Service")
	if err := framework.CreateServiceAndWaitUntilReady(amsvc); err != nil {
		t.Fatal(err)
	}

	log.Print("Validating Prometheus properly discovered alertmanagers")
	err = framework.Poll(18*time.Minute, time.Second, isAlertmanagerDiscoveryWorking(alertmanagerName))
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
		if err := framework.DeleteService(service.Name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if err := framework.CreateServiceAndWaitUntilReady(service); err != nil {
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
		if err := framework.DeleteService(service.Name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreatePrometheusAndWaitUntilReady(basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if err := framework.CreateServiceAndWaitUntilReady(service); err != nil {
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
	ingress := framework.MakeBasicIngress(prometheusService.Name, 9090)

	defer func() {
		if err := framework.DeletePrometheusAndWaitUntilGone(prometheus.Name); err != nil {
			t.Fatal(err)
		}
		if err := framework.DeleteService(prometheusService.Name); err != nil {
			t.Fatal(err)
		}
		if err := framework.KubeClient.Extensions().Ingresses(framework.Namespace.Name).Delete(ingress.Name, nil); err != nil {
			t.Fatal(err)
		}
		if err := framework.DeleteNginxIngressControllerIncDefaultBackend(); err != nil {
			t.Fatal(err)
		}
	}()

	err := framework.SetupNginxIngressControllerIncDefaultBackend()
	if err != nil {
		t.Fatal(err)
	}

	err = framework.CreatePrometheusAndWaitUntilReady(prometheus)
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateServiceAndWaitUntilReady(prometheusService); err != nil {
		t.Fatal(err)
	}

	err = framework.CreateIngress(ingress)
	if err != nil {
		t.Fatal(err)
	}

	ip, err := framework.GetIngressIP(ingress.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForHTTPSuccessStatusCode(time.Second*30, fmt.Sprintf("http://%s:/metrics", *ip))
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

		resp, err := http.Get(fmt.Sprintf("http://%s:30900/api/v1/targets", framework.ClusterIP))
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		rt := prometheusTargetAPIResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
			return false, err
		}

		if !assertExpectedTargets(rt.Data.ActiveTargets, expectedTargets) {
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

func assertExpectedTargets(targets []*target, expectedTargets []string) bool {
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

type target struct {
	ScrapeURL string `json:"scrapeUrl"`
}

type targetDiscovery struct {
	ActiveTargets []*target `json:"activeTargets"`
}

type prometheusTargetAPIResponse struct {
	Status string           `json:"status"`
	Data   *targetDiscovery `json:"data"`
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
