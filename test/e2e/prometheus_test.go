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
	"github.com/pkg/errors"
)

func TestPrometheusCreateDeleteCluster(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	prometheusTPR := framework.MakeBasicPrometheus(ctx.Id, name, name, 1)
	prometheusTPR.Namespace = ctx.Id

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, prometheusTPR); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(ctx.Id, name); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusScaleUpDownCluster(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, framework.MakeBasicPrometheus(ctx.Id, name, name, 1)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(ctx.Id, framework.MakeBasicPrometheus(ctx.Id, name, name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(ctx.Id, framework.MakeBasicPrometheus(ctx.Id, name, name, 2)); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusVersionMigration(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ctx.Id, name, name, 1)

	p.Spec.Version = "v1.5.1"
	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.6.1"
	if err := framework.UpdatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}
	if err := framework.WaitForPrometheusRunImageAndReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	p.Spec.Version = "v1.5.1"
	if err := framework.UpdatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}
	if err := framework.WaitForPrometheusRunImageAndReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusResourceUpdate(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ctx.Id, name, name, 1)

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	pods, err := framework.KubeClient.Core().Pods(ctx.Id).List(prometheus.ListOptions(name))
	if err != nil {
		t.Fatal(err)
	}
	res := pods.Items[0].Spec.Containers[0].Resources

	if !reflect.DeepEqual(res, p.Spec.Resources) {
		t.Fatalf("resources don't match. Has %q, want %q", res, p.Spec.Resources)
	}

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("200Mi"),
		},
	}
	if err := framework.UpdatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	pods, err = framework.KubeClient.Core().Pods(ctx.Id).List(prometheus.ListOptions(name))
	if err != nil {
		t.Fatal(err)
	}
	res = pods.Items[0].Spec.Containers[0].Resources

	if !reflect.DeepEqual(res, p.Spec.Resources) {
		t.Fatalf("resources don't match. Has %q, want %q", res, p.Spec.Resources)
	}
}

func TestPrometheusReloadConfig(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"
	replicas := int32(1)
	p := &v1alpha1.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ctx.Id,
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

	if _, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Create(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, svc); err != nil {
		t.Fatal(err)
	} else {
		ctx.AddFinalizerFn(finalizerFn)
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
	if _, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(2); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadRules(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

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

	_, err := framework.KubeClient.CoreV1().ConfigMaps(ctx.Id).Create(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, framework.MakeBasicPrometheus(ctx.Id, name, name, 1)); err != nil {
		t.Fatal(err)
	}

	ruleFileConfigMap.Data["test.rules"] = "# comment to trigger a configmap reload"
	_, err = framework.KubeClient.CoreV1().ConfigMaps(ctx.Id).Update(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	// remounting a ConfigMap can take some time
	err = wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		logs, err := testFramework.GetLogs(framework.KubeClient, ctx.Id, fmt.Sprintf("prometheus-%s-0", name), "prometheus-config-reloader")
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
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakeBasicPrometheusNodePortService(prometheusName, group, 30900)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(ctx.Id).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ctx.Id, prometheusName, group, 1)
	p.Spec.Version = "v1.5.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Get(fmt.Sprintf("prometheus-%s", prometheusName))
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = wait.Poll(time.Second, 18*time.Minute, isDiscoveryWorking(ctx.Id, prometheusName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func TestPrometheusAlertmanagerDiscovery(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	prometheusName := "test"
	alertmanagerName := "test"
	group := "servicediscovery-test"
	svc := framework.MakeBasicPrometheusNodePortService(prometheusName, group, 30900)
	amsvc := framework.MakeAlertmanagerNodePortService(alertmanagerName, group, 30903)

	p := framework.MakeBasicPrometheus(ctx.Id, prometheusName, group, 1)
	framework.AddAlertingToPrometheus(p, ctx.Id, alertmanagerName)
	p.Spec.Version = "v1.5.0"
	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(ctx.Id).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: %v", err)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Get(fmt.Sprintf("prometheus-%s", prometheusName))
	if err != nil {
		t.Fatalf("Generated Secret could not be retrieved: %v", err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, amsvc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Alertmanager service failed"))
	}

	err = wait.Poll(time.Second, 18*time.Minute, isAlertmanagerDiscoveryWorking(ctx.Id, alertmanagerName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus Alertmanager discovery failed"))
	}
}

func TestExposingPrometheusWithNodePort(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	basicPrometheus := framework.MakeBasicPrometheus(ctx.Id, "test", "test", 1)
	service := framework.MakeBasicPrometheusNodePortService(basicPrometheus.Name, "nodeport-service", 30900)

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:30900/metrics", framework.ClusterIP))
	if err != nil {
		t.Fatal("Retrieving prometheus metrics failed with error: ", err)
	} else if resp.StatusCode != 200 {
		t.Fatal("Retrieving prometheus metrics failed with http status code: ", resp.StatusCode)
	}
}

func TestExposingPrometheusWithKubernetesAPI(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	basicPrometheus := framework.MakeBasicPrometheus(ctx.Id, "basic-prometheus", "test-group", 1)
	service := framework.MakePrometheusService(basicPrometheus.Name, "test-group", v1.ServiceTypeClusterIP)

	if err := framework.CreatePrometheusAndWaitUntilReady(ctx.Id, basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	ProxyGet := framework.KubeClient.CoreV1().Services(ctx.Id).ProxyGet
	request := ProxyGet("", service.Name, "web", "/metrics", make(map[string]string))
	_, err := request.DoRaw()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExposingPrometheusWithIngress(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	prometheus := framework.MakeBasicPrometheus(ctx.Id, "main", "test-group", 1)
	prometheusService := framework.MakePrometheusService(prometheus.Name, "test-group", v1.ServiceTypeClusterIP)
	ingress := testFramework.MakeBasicIngress(prometheusService.Name, 9090)

	err := testFramework.SetupNginxIngressControllerIncDefaultBackend(framework.KubeClient, ctx.Id)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.CreatePrometheusAndWaitUntilReady(ctx.Id, prometheus)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, prometheusService); err != nil {
		t.Fatal(err)
	}

	err = testFramework.CreateIngress(framework.KubeClient, ctx.Id, ingress)
	if err != nil {
		t.Fatal(err)
	}

	ip, err := testFramework.GetIngressIP(framework.KubeClient, ctx.Id, ingress.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = testFramework.WaitForHTTPSuccessStatusCode(time.Minute, fmt.Sprintf("http://%s:/metrics", *ip))
	if err != nil {
		t.Fatal(err)
	}
}
func isDiscoveryWorking(ns, prometheusName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(prometheus.ListOptions(prometheusName))
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

func isAlertmanagerDiscoveryWorking(ns, alertmanagerName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(alertmanager.ListOptions(alertmanagerName))
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
