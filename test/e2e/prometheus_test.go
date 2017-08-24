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
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	testFramework "github.com/coreos/prometheus-operator/test/framework"
	"github.com/pkg/errors"
)

func TestPrometheusCreateDeleteCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	prometheusTPR := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusTPR.Namespace = ns

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusTPR); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusScaleUpDownCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, framework.MakeBasicPrometheus(ns, name, name, 1)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(ns, framework.MakeBasicPrometheus(ns, name, name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdatePrometheusAndWaitUntilReady(ns, framework.MakeBasicPrometheus(ns, name, name, 2)); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusVersionMigration(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	startVersion := prometheus.CompatibilityMatrix[0]
	compatibilityMatrix := prometheus.CompatibilityMatrix[1:]

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.Version = startVersion
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	for _, v := range compatibilityMatrix {
		p.Spec.Version = v
		if err := framework.UpdatePrometheusAndWaitUntilReady(ns, p); err != nil {
			t.Fatal(err)
		}
		if err := framework.WaitForPrometheusRunImageAndReady(ns, p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPrometheusResourceUpdate(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	pods, err := framework.KubeClient.Core().Pods(ns).List(prometheus.ListOptions(name))
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
	if err := framework.UpdatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	pods, err = framework.KubeClient.Core().Pods(ns).List(prometheus.ListOptions(name))
	if err != nil {
		t.Fatal(err)
	}
	res = pods.Items[0].Spec.Containers[0].Resources

	if !reflect.DeepEqual(res, p.Spec.Resources) {
		t.Fatalf("resources don't match. Has %q, want %q", res, p.Spec.Resources)
	}
}

func TestPrometheusReloadConfig(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	replicas := int32(1)
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: monitoringv1.PrometheusSpec{
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
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
		},
		Data: map[string][]byte{
			"prometheus.yaml": []byte(firstConfig),
		},
	}

	svc := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(err)
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if err := framework.WaitForTargets(ns, svc.Name, 1); err != nil {
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
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(ns, svc.Name, 2); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusReloadRules(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	ruleFileConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s-rules", name),
			Labels: map[string]string{
				"role": "rulefile",
			},
		},
		Data: map[string]string{
			"test.rules": "",
		},
	}

	_, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	ruleFileConfigMap.Data["test.rules"] = "# comment to trigger a configmap reload"
	_, err = framework.KubeClient.CoreV1().ConfigMaps(ns).Update(ruleFileConfigMap)
	if err != nil {
		t.Fatal(err)
	}

	// remounting a ConfigMap can take some time
	err = wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		logs, err := testFramework.GetLogs(framework.KubeClient, ns, fmt.Sprintf("prometheus-%s-0", name), "prometheus-config-reloader")
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
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.Version = "v1.7.1"
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ns).Get(fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = wait.Poll(time.Second, 18*time.Minute, isDiscoveryWorking(ns, svc.Name, prometheusName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func TestPrometheusAlertmanagerDiscovery(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	alertmanagerName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)
	amsvc := framework.MakeAlertmanagerService(alertmanagerName, group, v1.ServiceTypeClusterIP)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	framework.AddAlertingToPrometheus(p, ns, alertmanagerName)
	p.Spec.Version = "v1.7.1"
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: %v", err)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ns).Get(fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Generated Secret could not be retrieved: %v", err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, amsvc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Alertmanager service failed"))
	}

	err = wait.Poll(time.Second, 18*time.Minute, isAlertmanagerDiscoveryWorking(ns, svc.Name, alertmanagerName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus Alertmanager discovery failed"))
	}
}

func TestExposingPrometheusWithKubernetesAPI(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	basicPrometheus := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	service := framework.MakePrometheusService(basicPrometheus.Name, "test-group", v1.ServiceTypeClusterIP)

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	ProxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
	request := ProxyGet("", service.Name, "web", "/metrics", make(map[string]string))
	_, err := request.DoRaw()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusDiscoverTargetPort(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	if _, err := framework.MonClient.ServiceMonitors(ns).Create(&monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: prometheusName,
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": group,
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				monitoringv1.Endpoint{
					TargetPort: intstr.FromInt(9090),
					Interval:   "30s",
				},
			},
		},
	}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ns).Get(fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = wait.Poll(time.Second, 3*time.Minute, isDiscoveryWorking(ns, svc.Name, prometheusName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func isDiscoveryWorking(ns, svcName, prometheusName string) func() (bool, error) {
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

		activeTargets, err := framework.GetActiveTargets(ns, svcName)
		if err != nil {
			return false, err
		}

		if !assertExpectedTargets(activeTargets, expectedTargets) {
			return false, nil
		}

		working, err := basicQueryWorking(ns, svcName)
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

func basicQueryWorking(ns, svcName string) (bool, error) {
	response, err := framework.QueryPrometheusSVC(ns, svcName, "/api/v1/query", map[string]string{"query": "up"})
	if err != nil {
		return false, err
	}

	rq := prometheusQueryAPIResponse{}
	if err := json.NewDecoder(response).Decode(&rq); err != nil {
		return false, err
	}

	if rq.Status != "success" && rq.Data.Result[0].Value[1] == "1" {
		log.Printf("Query Response not successful.")
		return false, nil
	}

	return true, nil
}

func isAlertmanagerDiscoveryWorking(ns, promSVCName, alertmanagerName string) func() (bool, error) {
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

		response, err := framework.QueryPrometheusSVC(ns, promSVCName, "/api/v1/alertmanagers", map[string]string{})
		if err != nil {
			return false, err
		}

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(response).Decode(&ra); err != nil {
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
