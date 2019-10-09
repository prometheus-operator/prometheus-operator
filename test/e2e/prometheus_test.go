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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	testFramework "github.com/coreos/prometheus-operator/test/framework"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
)

func testPromCreateDeleteCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusCRD.Namespace = ns

	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func testPromScaleUpDownCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Replicas = proto.Int32(3)
	p, err = framework.UpdatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Replicas = proto.Int32(2)
	p, err = framework.UpdatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}
}

func testPromNoServiceMonitorSelector(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ServiceMonitorSelector = nil
	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}
}

func testPromVersionMigration(t *testing.T) {
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
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range compatibilityMatrix {
		p.Spec.Version = v
		p, err = framework.UpdatePrometheusAndWaitUntilReady(ns, p)
		if err != nil {
			t.Fatal(err)
		}
		if err := framework.WaitForPrometheusRunImageAndReady(ns, p); err != nil {
			t.Fatal(err)
		}
	}
}

func testPromResourceUpdate(t *testing.T) {
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
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	pods, err := framework.KubeClient.CoreV1().Pods(ns).List(prometheus.ListOptions(name))
	if err != nil {
		t.Fatal(err)
	}
	res := pods.Items[0].Spec.Containers[0].Resources

	if !reflect.DeepEqual(res, p.Spec.Resources) {
		t.Fatalf("resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)
	}

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("200Mi"),
		},
	}
	p, err = framework.MonClientV1.Prometheuses(ns).Update(p)
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(prometheus.ListOptions(name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		res = pods.Items[0].Spec.Containers[0].Resources
		if !reflect.DeepEqual(res, p.Spec.Resources) {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testPromStorageUpdate(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Storage = &monitoringv1.StorageSpec{
		VolumeClaimTemplate: v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Now(),
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}
	p, err = framework.MonClientV1.Prometheuses(ns).Update(p)
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(prometheus.ListOptions(name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		for _, volume := range pods.Items[0].Spec.Volumes {
			if volume.Name == "prometheus-"+name+"-db" && volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testPromReloadConfig(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ServiceMonitorSelector = nil

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

	var bufOne bytes.Buffer
	if err := gzipConfig(&bufOne, []byte(firstConfig)); err != nil {
		t.Fatal(err)
	}
	firstConfigCompressed := bufOne.Bytes()

	cfg := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
		},
		Data: map[string][]byte{
			"prometheus.yaml.gz": firstConfigCompressed,
			"configmaps.json":    []byte("{}"),
		},
	}

	svc := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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

	var bufTwo bytes.Buffer
	if err := gzipConfig(&bufTwo, []byte(secondConfig)); err != nil {
		t.Fatal(err)
	}
	secondConfigCompressed := bufTwo.Bytes()

	cfg, err := framework.KubeClient.CoreV1().Secrets(ns).Get(cfg.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(errors.Wrap(err, "could not retrieve previous secret"))
	}

	cfg.Data["prometheus.yaml.gz"] = secondConfigCompressed
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(ns, svc.Name, 2); err != nil {
		t.Fatal(err)
	}
}

func testPromAdditionalScrapeConfig(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "additional-config-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	additionalConfig := `
- job_name: "prometheus"
  static_configs:
  - targets: ["localhost:9090"]
`
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "additional-scrape-configs",
		},
		Data: map[string][]byte{
			"prometheus-additional.yaml": []byte(additionalConfig),
		},
	}
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(&secret)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalScrapeConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-scrape-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	// Wait for ServiceMonitor target, as well as additional-config target
	if err := framework.WaitForTargets(ns, svc.Name, 2); err != nil {
		t.Fatal(err)
	}
}

func testPromAdditionalAlertManagerConfig(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "additional-alert-config-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	additionalConfig := `
- path_prefix: /
  scheme: http
  static_configs:
  - targets: ["localhost:9093"]
`
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "additional-alert-configs",
		},
		Data: map[string][]byte{
			"prometheus-additional.yaml": []byte(additionalConfig),
		},
	}
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(&secret)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalAlertManagerConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-alert-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	// Wait for ServiceMonitor target
	if err := framework.WaitForTargets(ns, svc.Name, 1); err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(time.Second, 5*time.Minute, func() (done bool, err error) {
		response, err := framework.PrometheusSVCGetRequest(ns, svc.Name, "/api/v1/alertmanagers", map[string]string{})
		if err != nil {
			return true, err
		}

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&ra); err != nil {
			return true, err
		}

		if ra.Status == "success" && len(ra.Data.ActiveAlertmanagers) == 1 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus Alertmanager configuration failed"))
	}
}

func testPromReloadRules(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	firtAlertName := "firstAlert"
	secondAlertName := "secondAlert"

	ruleFile, err := framework.MakeAndCreateFiringRule(ns, name, firtAlertName)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err = framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	err = framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, firtAlertName)
	if err != nil {
		t.Fatal(err)
	}

	ruleFile.Spec.Groups = []monitoringv1.RuleGroup{
		{
			Name: "my-alerting-group",
			Rules: []monitoringv1.Rule{
				{
					Alert: secondAlertName,
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	}
	ruleFile, err = framework.UpdateRule(ns, ruleFile)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, secondAlertName)
	if err != nil {
		t.Fatal(err)
	}
}

func testPromMultiplePrometheusRulesSameNS(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	alertNames := []string{"first-alert", "second-alert"}

	for _, alertName := range alertNames {
		_, err := framework.MakeAndCreateFiringRule(ns, alertName, alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	for _, alertName := range alertNames {
		err := framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, alertName)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testPromMultiplePrometheusRulesDifferentNS(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	rootNS := ctx.CreateNamespace(t, framework.KubeClient)
	alertNSOne := ctx.CreateNamespace(t, framework.KubeClient)
	alertNSTwo := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, rootNS, framework.KubeClient)

	name := "test"
	ruleFiles := []struct {
		alertName string
		ns        string
	}{{"first-alert", alertNSOne}, {"second-alert", alertNSTwo}}

	ruleFilesNamespaceSelector := map[string]string{"prometheus": rootNS}

	for _, file := range ruleFiles {
		err := testFramework.AddLabelsToNamespace(framework.KubeClient, file.ns, ruleFilesNamespaceSelector)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range ruleFiles {
		_, err := framework.MakeAndCreateFiringRule(file.ns, file.alertName, file.alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(rootNS, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: ruleFilesNamespaceSelector,
	}
	p, err := framework.CreatePrometheusAndWaitUntilReady(rootNS, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, rootNS, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	for _, file := range ruleFiles {
		err := framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, file.alertName)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testPromRulesExceedingConfigMapLimit(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusRules := []*monitoringv1.PrometheusRule{}
	for i := 0; i < 2; i++ {
		rule := generateHugePrometheusRule(ns, strconv.Itoa(i))
		rule, err := framework.CreateRule(ns, rule)
		if err != nil {
			t.Fatal(err)
		}
		prometheusRules = append(prometheusRules, rule)
	}

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if t.Failed() {
			if err := framework.PrintPodLogs(ns, "prometheus-"+p.Name+"-0"); err != nil {
				t.Fatal(err)
			}
		}
	}()

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	for i := range prometheusRules {
		_, err := framework.WaitForConfigMapExist(ns, "prometheus-"+p.Name+"-rulefiles-"+strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Make sure both rule files ended up in the Prometheus Pod
	for i := range prometheusRules {
		err := framework.WaitForPrometheusFiringAlert(ns, pSVC.Name, "my-alert-"+strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	err = framework.DeleteRule(ns, prometheusRules[1].Name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = framework.WaitForConfigMapExist(ns, "prometheus-"+p.Name+"-rulefiles-0")
	if err != nil {
		t.Fatal(err)
	}
	err = framework.WaitForConfigMapNotExist(ns, "prometheus-"+p.Name+"-rulefiles-1")
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusFiringAlert(ns, pSVC.Name, "my-alert-0")
	if err != nil {
		t.Fatal(err)
	}
}

func testPromRulesMustBeAnnotated(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "admission"
	admissionAlert := "admissionAlert"

	_, err := framework.MakeAndCreateFiringRule(ns, name, admissionAlert)
	if err != nil {
		t.Fatal(err)
	}

	rule, err := framework.GetRule(ns, name)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := rule.Annotations["prometheus-operator-validated"]
	if !ok {
		t.Fatal("Expected prometheusrule to be annotated")
	}
	if val != "true" {
		t.Fatal("Expected prometheusrule annotation to be 'true'")
	}
}

func testInvalidRulesAreRejected(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "admission"
	admissionAlert := "admissionAlert"

	_, err := framework.MakeAndCreateInvalidRule(ns, name, admissionAlert)
	if err == nil {
		t.Fatal("Expected invalid prometheusrule to be rejected")
	}
}

// generateHugePrometheusRule returns a Prometheus rule instance that would fill
// more than half of the space of a Kubernetes ConfigMap.
func generateHugePrometheusRule(ns, identifier string) *monitoringv1.PrometheusRule {
	alertName := "my-alert"
	groups := []monitoringv1.RuleGroup{
		{
			Name:  alertName,
			Rules: []monitoringv1.Rule{},
		},
	}
	// One rule marshaled as yaml is ~34 bytes long, the max is ~524288 bytes.
	for i := 0; i < 12000; i++ {
		groups[0].Rules = append(groups[0].Rules, monitoringv1.Rule{
			Alert: alertName + "-" + identifier,
			Expr:  intstr.FromString("vector(1)"),
		})
	}
	rule := framework.MakeBasicRule(ns, "prometheus-rule-"+identifier, groups)

	return rule
}

// Make sure the Prometheus operator only updates the Prometheus config secret
// and the Prometheus rules configmap on relevant changes
func testPromOnlyUpdatedOnRelevantChanges(t *testing.T) {
	t.Parallel()

	testCTX := framework.NewTestCtx(t)
	defer testCTX.Cleanup(t)
	ns := testCTX.CreateNamespace(t, framework.KubeClient)
	testCTX.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)

	// Adding an annotation to Prometheus lead to high CPU usage in the past
	// updating the Prometheus StatefulSet in a loop (See
	// https://github.com/coreos/prometheus-operator/issues/1659). Added here to
	// prevent a regression.
	prometheus.Annotations["test-annotation"] = "test-value"

	ctx, cancel := context.WithCancel(context.Background())

	type versionedResource interface {
		GetResourceVersion() string
	}

	resourceDefinitions := []struct {
		Name               string
		Getter             func(prometheusName string) (versionedResource, error)
		Versions           map[string]interface{}
		MaxExpectedChanges int
	}{
		{
			Name: "prometheus",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					MonClientV1.
					Prometheuses(ns).
					Get(prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
		{
			Name: "rulesConfigMap",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					ConfigMaps(ns).
					Get("prometheus-"+prometheusName+"-rulefiles-0", metav1.GetOptions{})
			},
			// The Prometheus Operator first creates the ConfigMap for the
			// given Prometheus stateful set and then updates it with the matching
			// Prometheus rules.
			MaxExpectedChanges: 2,
		},
		{
			Name: "configurationSecret",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Secrets(ns).
					Get("prometheus-"+prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 2,
		},
		{
			Name: "tlsAssetSecret",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Secrets(ns).
					Get("prometheus-"+prometheusName+"-tls-assets", metav1.GetOptions{})
			},
			MaxExpectedChanges: 2,
		},
		{
			Name: "statefulset",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					AppsV1().
					StatefulSets(ns).
					Get("prometheus-"+prometheusName, metav1.GetOptions{})
			},
			// First is the creation of the StatefulSet itself, following is the
			// update of e.g. the ReadyReplicas status field
			MaxExpectedChanges: 3,
		},
		{
			Name: "service-operated",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Services(ns).
					Get("prometheus-operated", metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
		{
			Name: "serviceMonitor",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					MonClientV1.
					ServiceMonitors(ns).
					Get(prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
	}

	// Init Versions maps
	for i := range resourceDefinitions {
		resourceDefinitions[i].Versions = map[string]interface{}{}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(10 * time.Millisecond)

				for i, resourceDef := range resourceDefinitions {
					resource, err := resourceDef.Getter(prometheus.Name)
					if apierrors.IsNotFound(err) {
						continue
					}
					if err != nil {
						cancel()
						t.Fatal(err)
					}

					resourceDefinitions[i].Versions[resource.GetResourceVersion()] = resource
				}
			}
		}
	}()

	alertName := "my-alert"
	if _, err := framework.MakeAndCreateFiringRule(ns, "my-prometheus-rule", alertName); err != nil {
		t.Fatal(err)
	}

	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheus)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCTX.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(name)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	err = framework.WaitForPrometheusFiringAlert(prometheus.Namespace, pSVC.Name, alertName)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForDiscoveryWorking(ns, pSVC.Name, prometheus.Name)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}

	cancel()

	for _, resource := range resourceDefinitions {
		if len(resource.Versions) > resource.MaxExpectedChanges || len(resource.Versions) < 1 {
			var previous interface{}
			for _, version := range resource.Versions {
				if previous == nil {
					previous = version
					continue
				}
				fmt.Println(pretty.Compare(previous, version))
				previous = version
			}

			t.Fatalf(
				"expected resource %v to be created/updated %v times, but saw %v instead",
				resource.Name,
				resource.MaxExpectedChanges,
				len(resource.Versions),
			)
		}
	}
}

func testPromWhenDeleteCRDCleanUpViaOwnerRef(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	configMapName := fmt.Sprintf("prometheus-%v-rulefiles-0", p.Name)

	_, err = framework.WaitForConfigMapExist(ns, configMapName)
	if err != nil {
		t.Fatal(err)
	}

	// Waits for Prometheus pods to vanish
	err = framework.DeletePrometheusAndWaitUntilGone(ns, p.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForConfigMapNotExist(ns, configMapName)
	if err != nil {
		t.Fatal(err)
	}
}

func testPromDiscovery(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = framework.WaitForDiscoveryWorking(ns, svc.Name, prometheusName)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func testPromAlertmanagerDiscovery(t *testing.T) {
	t.Parallel()

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
	p, err := framework.CreatePrometheusAndWaitUntilReady(ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: %v", err)
	}

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Generated Secret could not be retrieved: %v", err)
	}

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
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

func testPromExposingWithKubernetesAPI(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	basicPrometheus := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	service := framework.MakePrometheusService(basicPrometheus.Name, "test-group", v1.ServiceTypeClusterIP)

	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, basicPrometheus); err != nil {
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

func testPromDiscoverTargetPort(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	targetPort := intstr.FromInt(9090)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(&monitoringv1.ServiceMonitor{
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
				{
					TargetPort: &targetPort,
					Interval:   "30s",
				},
			},
		},
	}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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

	err = framework.WaitForDiscoveryWorking(ns, svc.Name, prometheusName)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func testPromOpMatchPromAndServMonInDiffNSs(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	prometheusNSName := ctx.CreateNamespace(t, framework.KubeClient)
	serviceMonitorNSName := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, prometheusNSName, framework.KubeClient)

	if err := testFramework.AddLabelsToNamespace(
		framework.KubeClient,
		serviceMonitorNSName,
		map[string]string{"team": "frontend"},
	); err != nil {
		t.Fatal(err)
	}

	group := "sample-app"

	prometheusJobName := serviceMonitorNSName + "/" + group

	prometheusName := "test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)

	if _, err := framework.MonClientV1.ServiceMonitors(serviceMonitorNSName).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(prometheusNSName, prometheusName, group, 1)
	p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"team": "frontend",
		},
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(prometheusNSName, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, prometheusNSName, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	resp, err := framework.PrometheusSVCGetRequest(prometheusNSName, svc.Name, "/api/v1/status/config", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(resp), prometheusJobName) != 1 {
		t.Fatalf("expected Prometheus operator to configure Prometheus in ns '%v' to scrape the service monitor in ns '%v'", prometheusNSName, serviceMonitorNSName)
	}
}

func testThanos(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	version := prometheus.DefaultThanosVersion

	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	prom.Spec.Replicas = proto.Int32(2)
	prom.Spec.Thanos = &monitoringv1.ThanosSpec{
		Version: &version,
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, prom); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	promSvc := framework.MakePrometheusService(prom.Name, "test-group", v1.ServiceTypeClusterIP)
	if _, err := framework.KubeClient.CoreV1().Services(ns).Create(promSvc); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	svcMon := framework.MakeBasicServiceMonitor("test-group")
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(svcMon); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	qryDep, err := testFramework.MakeDeployment("../../example/thanos/query-deployment.yaml")
	if err != nil {
		t.Fatal("Making thanos query deployment failed: ", err)
	}
	// override image
	qryImage := "quay.io/thanos/thanos:" + version
	t.Log("setting up query with image: ", qryImage)
	qryDep.Spec.Template.Spec.Containers[0].Image = qryImage
	// override args
	qryArgs := []string{
		"query",
		"--log.level=debug",
		"--query.replica-label=prometheus_replica",
		fmt.Sprintf("--store=dnssrv+_grpc._tcp.prometheus-operated.%s.svc.cluster.local", ns),
	}
	t.Log("setting up query with args: ", qryArgs)
	qryDep.Spec.Template.Spec.Containers[0].Args = qryArgs
	if err := testFramework.CreateDeployment(framework.KubeClient, ns, qryDep); err != nil {
		t.Fatal("Creating Thanos query deployment failed: ", err)
	}

	qrySvc := framework.MakeThanosQuerierService(qryDep.Name)
	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, qrySvc); err != nil {
		t.Fatal("Creating Thanos query service failed: ", err)
	}

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (bool, error) {
		proxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
		request := proxyGet("http", qrySvc.Name, "http-query", "/api/v1/query", map[string]string{
			"query": "prometheus_build_info",
			"dedup": "false",
		})
		b, err := request.DoRaw()
		if err != nil {
			t.Logf("Error performing request against Thanos query: %v\n\nretrying...", err)
			return false, nil
		}

		d := struct {
			Data struct {
				Result []map[string]interface{} `json:"result"`
			} `json:"data"`
		}{}

		err = json.Unmarshal(b, &d)
		if err != nil {
			return false, err
		}

		result := len(d.Data.Result)
		// We're expecting 4 results as we are requesting the
		// `prometheus_build_info` metric, which is collected for both
		// Prometheus replicas by both replicas.
		expected := 4
		if result != expected {
			t.Logf("Unexpected number of results from query. Got %d, expected %d. retrying...", result, expected)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatal("Failed to get correct result from Thanos query: ", err)
	}
}

func testPromGetAuthSecret(t *testing.T) {
	t.Parallel()
	name := "test"

	tests := []struct {
		name           string
		secret         *v1.Secret
		serviceMonitor func() *monitoringv1.ServiceMonitor
	}{
		{
			name: "basic-auth",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"user":     []byte("user"),
					"password": []byte("pass"),
				},
			},
			serviceMonitor: func() *monitoringv1.ServiceMonitor {
				sm := framework.MakeBasicServiceMonitor(name)
				sm.Spec.Endpoints[0].BasicAuth = &monitoringv1.BasicAuth{
					Username: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "user",
					},
					Password: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "password",
					},
				}

				return sm
			},
		},
		{
			name: "bearer-token",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"bearertoken": []byte("abc"),
				},
			},
			serviceMonitor: func() *monitoringv1.ServiceMonitor {
				sm := framework.MakeBasicServiceMonitor(name)
				sm.Spec.Endpoints[0].BearerTokenSecret = v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearertoken",
				}
				sm.Spec.Endpoints[0].Path = "/bearer-token"
				return sm
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := framework.NewTestCtx(t)
			defer ctx.Cleanup(t)
			ns := ctx.CreateNamespace(t, framework.KubeClient)
			ctx.SetupPrometheusRBACGlobal(t, ns, framework.KubeClient)

			maptest := make(map[string]string)
			maptest["tc"] = ns
			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
				MatchLabels: maptest,
			}
			prometheusCRD.Spec.ScrapeInterval = "1s"

			if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusCRD); err != nil {
				t.Fatal(err)
			}
			testNamespace := ctx.CreateNamespace(t, framework.KubeClient)

			err := testFramework.AddLabelsToNamespace(framework.KubeClient, testNamespace, maptest)
			if err != nil {
				t.Fatal(err)
			}

			simple, err := testFramework.MakeDeployment("../../test/framework/ressources/basic-auth-app-deployment.yaml")
			if err != nil {
				t.Fatal(err)
			}

			if err := testFramework.CreateDeployment(framework.KubeClient, testNamespace, simple); err != nil {
				t.Fatal("Creating simple basic auth app failed: ", err)
			}

			authSecret := test.secret
			if _, err := framework.KubeClient.CoreV1().Secrets(testNamespace).Create(authSecret); err != nil {
				t.Fatal(err)
			}

			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"group": name,
					},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Name: "web",
							Port: 8080,
						},
					},
					Selector: map[string]string{
						"group": name,
					},
				},
			}

			sm := test.serviceMonitor()
			if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, testNamespace, svc); err != nil {
				t.Fatal(err)
			} else {
				ctx.AddFinalizerFn(finalizerFn)
			}

			if _, err := framework.MonClientV1.ServiceMonitors(testNamespace).Create(sm); err != nil {
				t.Fatal("Creating ServiceMonitor failed: ", err)
			}

			if err := framework.WaitForTargets(ns, "prometheus-operated", 1); err != nil {
				t.Fatal(err)
			}
		})
	}

}

// testOperatorNSScope tests the multi namespace feature of the Prometheus Operator.
// It checks whether it ignores rules that are not in the watched namespaces of the
// Prometheus Operator. The Prometheus Operator internally treats watching a
// single namespace different than watching multiple namespaces, hence the two
// sub-tests.
func testOperatorNSScope(t *testing.T) {
	name := "test"
	firtAlertName := "firstAlert"
	secondAlertName := "secondAlert"

	t.Run("SingleNS", func(t *testing.T) {
		ctx := framework.NewTestCtx(t)
		defer ctx.Cleanup(t)

		operatorNS := ctx.CreateNamespace(t, framework.KubeClient)
		mainNS := ctx.CreateNamespace(t, framework.KubeClient)
		arbitraryNS := ctx.CreateNamespace(t, framework.KubeClient)

		ctx.SetupPrometheusRBAC(t, mainNS, framework.KubeClient)

		prometheusNamespaceSelector := map[string]string{"prometheus": mainNS}

		// Add labels to namespaces for Prometheus RuleNamespaceSelector.
		for _, ns := range []string{mainNS, arbitraryNS} {
			err := testFramework.AddLabelsToNamespace(framework.KubeClient, ns, prometheusNamespaceSelector)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Prometheus Operator only watches single namespace mainNS, not arbitraryNS.
		_, err := framework.CreatePrometheusOperator(operatorNS, *opImage, []string{mainNS}, nil, nil, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {mainNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(r.NSName, name, r.AlertName)
			if err != nil {
				t.Fatal(err)
			}
		}

		p := framework.MakeBasicPrometheus(mainNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(mainNS, p)
		if err != nil {
			t.Fatal(err)
		}

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, mainNS, pSVC); err != nil {
			t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
		} else {
			ctx.AddFinalizerFn(finalizerFn)
		}

		err = framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, firtAlertName)
		if err != nil {
			t.Fatal(err)
		}

		firing, err := framework.CheckPrometheusFiringAlert(p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		if firing {
			t.Fatalf("expected alert %q not to fire", secondAlertName)
		}
	})

	t.Run("MultiNS", func(t *testing.T) {
		ctx := framework.NewTestCtx(t)
		defer ctx.Cleanup(t)

		operatorNS := ctx.CreateNamespace(t, framework.KubeClient)
		prometheusNS := ctx.CreateNamespace(t, framework.KubeClient)
		ruleNS := ctx.CreateNamespace(t, framework.KubeClient)
		arbitraryNS := ctx.CreateNamespace(t, framework.KubeClient)

		ctx.SetupPrometheusRBAC(t, prometheusNS, framework.KubeClient)

		prometheusNamespaceSelector := map[string]string{"prometheus": prometheusNS}

		for _, ns := range []string{ruleNS, arbitraryNS} {
			err := testFramework.AddLabelsToNamespace(framework.KubeClient, ns, prometheusNamespaceSelector)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Prometheus Operator only watches prometheusNS and ruleNS, not arbitraryNS.
		_, err := framework.CreatePrometheusOperator(operatorNS, *opImage, []string{prometheusNS, ruleNS}, nil, nil, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {ruleNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(r.NSName, name, r.AlertName)
			if err != nil {
				t.Fatal(err)
			}
		}

		p := framework.MakeBasicPrometheus(prometheusNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(prometheusNS, p)
		if err != nil {
			t.Fatal(err)
		}

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, prometheusNS, pSVC); err != nil {
			t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
		} else {
			ctx.AddFinalizerFn(finalizerFn)
		}

		err = framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, firtAlertName)
		if err != nil {
			t.Fatal(err)
		}

		firing, err := framework.CheckPrometheusFiringAlert(p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		if firing {
			t.Fatalf("expected alert %q not to fire", secondAlertName)
		}
	})
}

// testPromArbitraryFSAcc tests the
// github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1.PrometheusSpec.ArbitraryFSAccessThroughSMs
// configuration with the service monitor bearer token and tls assets option.
func testPromArbitraryFSAcc(t *testing.T) {
	t.Parallel()

	name := "test"

	tests := []struct {
		name                              string
		arbitraryFSAccessThroughSMsConfig monitoringv1.ArbitraryFSAccessThroughSMsConfig
		endpoint                          monitoringv1.Endpoint
		expectTargets                     bool
	}{
		//
		// Bearer tokens:
		//
		{
			name: "allowed-bearer-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port:            "web",
				BearerTokenFile: "abc",
			},
			expectTargets: true,
		},
		{
			name: "denied-bearer-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port:            "web",
				BearerTokenFile: "abc",
			},
			expectTargets: false,
		},
		{
			name: "allowed-bearer-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				BearerTokenSecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearer-token",
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-bearer-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				BearerTokenSecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearer-token",
				},
			},
			expectTargets: true,
		},
		//
		// TLS assets:
		//
		{
			name: "allowed-tls-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					CAFile:   "/etc/ca-certificates/example-ca.pem",
					CertFile: "/etc/ca-certificates/example-cert.pem",
					KeyFile:  "/etc/ca-certificates/example-key.pem",
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-tls-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					CAFile:   "/etc/ca-certificates/example-ca.pem",
					CertFile: "/etc/ca-certificates/example-cert.pem",
					KeyFile:  "/etc/ca-certificates/example-key.pem",
				},
			},
			expectTargets: false,
		},
		{
			name: "allowed-tls-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					InsecureSkipVerify: true,
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-tls-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					InsecureSkipVerify: true,
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
			},
			expectTargets: false,
		},
		{
			name: "allowed-tls-configmap",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					InsecureSkipVerify: true,
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-tls-configmap",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					InsecureSkipVerify: true,
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
			},
			expectTargets: false,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := framework.NewTestCtx(t)
			defer ctx.Cleanup(t)
			ns := ctx.CreateNamespace(t, framework.KubeClient)
			ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

			// Create secret either used by bearer token secret key ref, tls
			// asset key ref or tls configmap key ref.
			cert, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
			if err != nil {
				t.Fatalf("failed to load cert.pem: %v", err)
			}

			key, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
			if err != nil {
				t.Fatalf("failed to load key.pem: %v", err)
			}

			tlsCertsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"cert.pem":     cert,
					"key.pem":      key,
					"bearer-token": []byte("abc"),
				},
			}

			if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(tlsCertsSecret); err != nil {
				t.Fatal(err)
			}

			tlsCertsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"cert.pem": string(cert),
				},
			}

			if _, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(tlsCertsConfigMap); err != nil {
				t.Fatal(err)
			}

			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Namespace = ns
			prometheusCRD.Spec.ArbitraryFSAccessThroughSMs = test.arbitraryFSAccessThroughSMsConfig

			if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusCRD); err != nil {
				t.Fatal(err)
			}

			svc := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)
			if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
				t.Fatal(err)
			}

			s := framework.MakeBasicServiceMonitor(name)
			s.Spec.Endpoints[0] = test.endpoint
			if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
				t.Fatal("creating ServiceMonitor failed: ", err)
			}

			if test.expectTargets {
				if err := framework.WaitForTargets(ns, svc.Name, 1); err != nil {
					t.Fatal(err)
				}

				return
			}

			// Make sure Prometheus has enough time to reload.
			time.Sleep(2 * time.Minute)
			if err := framework.WaitForTargets(ns, svc.Name, 0); err != nil {
				t.Fatal(err)
			}
		})
	}

}

// testPromTLSConfigViaSecret tests the service monitor endpoint option to load
// certificate assets via Kubernetes secrets into the Prometheus container.
func testPromTLSConfigViaSecret(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)
	name := "test"

	//
	// Setup sample app.
	//

	cert, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
	if err != nil {
		t.Fatalf("failed to load cert.pem: %v", err)
	}

	key, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
	if err != nil {
		t.Fatalf("failed to load key.pem: %v", err)
	}

	tlsCertsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"cert.pem": cert,
			"key.pem":  key,
		},
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(tlsCertsSecret); err != nil {
		t.Fatal(err)
	}

	simple, err := testFramework.MakeDeployment("../../test/framework/ressources/basic-auth-app-deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

	simple.Spec.Template.Spec.Containers[0].Args = []string{"--cert-path=/etc/certs"}

	simple.Spec.Template.Spec.Volumes = []v1.Volume{
		v1.Volume{
			Name: "tls-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: tlsCertsSecret.Name,
				},
			},
		},
	}

	simple.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      simple.Spec.Template.Spec.Volumes[0].Name,
			MountPath: "/etc/certs",
		},
	}

	if err := testFramework.CreateDeployment(framework.KubeClient, ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "web",
					Port: 8080,
				},
				v1.ServicePort{
					Name: "mtls",
					Port: 8081,
				},
			},
			Selector: map[string]string{
				"group": name,
			},
		},
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, svc); err != nil {
		t.Fatal(err)
	}

	//
	// Setup monitoring.
	//

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     "mtls",
			Interval: "30s",
			Scheme:   "https",
			TLSConfig: &monitoringv1.TLSConfig{
				InsecureSkipVerify: true,
				CA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: tlsCertsSecret.Name,
						},
						Key: "cert.pem",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: tlsCertsSecret.Name,
						},
						Key: "cert.pem",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: tlsCertsSecret.Name,
					},
					Key: "key.pem",
				},
			},
		},
	}

	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(sm); err != nil {
		t.Fatal("creating ServiceMonitor failed: ", err)
	}

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	promSVC := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, promSVC); err != nil {
		t.Fatal(err)
	}

	//
	// Check for proper scraping.
	//

	if err := framework.WaitForTargets(ns, promSVC.Name, 1); err != nil {
		t.Fatal(err)
	}

	// TODO: Do a poll instead, should speed up things.
	time.Sleep(30 * time.Second)

	response, err := framework.PrometheusSVCGetRequest(
		ns,
		promSVC.Name,
		"/api/v1/query",
		map[string]string{"query": fmt.Sprintf(`up{job="%v",endpoint="%v"}`, name, sm.Spec.Endpoints[0].Port)},
	)
	if err != nil {
		t.Fatal(err)
	}

	q := testFramework.PrometheusQueryAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&q); err != nil {
		t.Fatal(err)
	}

	if q.Status != "success" {
		t.Fatalf("expected query status to be 'success' but got %v", q.Status)
	}

	if q.Data.Result[0].Value[1] != "1" {
		t.Fatalf("expected query result to be '1' but got %v", q.Data.Result[0].Value[1])
	}
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

		response, err := framework.PrometheusSVCGetRequest(ns, promSVCName, "/api/v1/alertmanagers", map[string]string{})
		if err != nil {
			return false, err
		}

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&ra); err != nil {
			return false, err
		}

		if assertExpectedAlertmanagerTargets(ra.Data.ActiveAlertmanagers, expectedAlertmanagerTargets) {
			return true, nil
		}

		return false, nil
	}
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

func gzipConfig(buf *bytes.Buffer, conf []byte) error {
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(conf); err != nil {
		return err
	}
	return nil
}
