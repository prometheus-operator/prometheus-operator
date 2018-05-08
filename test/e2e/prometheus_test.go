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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	testFramework "github.com/coreos/prometheus-operator/test/framework"

	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
)

func TestPrometheusCreateDeleteCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusCRD.Namespace = ns

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheusCRD); err != nil {
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
		t.Fatalf("resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)
	}

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("200Mi"),
		},
	}
	_, err = framework.MonClientV1.Prometheuses(ns).Update(p)
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.Core().Pods(ns).List(prometheus.ListOptions(name))
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

func TestPrometheusReloadConfig(t *testing.T) {
	t.Parallel()

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
			"configmaps.json": []byte("{}"),
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

	cfg, err := framework.KubeClient.CoreV1().Secrets(ns).Get(cfg.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(errors.Wrap(err, "could not retrieve previous secret"))
	}

	cfg.Data["prometheus.yaml"] = []byte(secondConfig)
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForTargets(ns, svc.Name, 2); err != nil {
		t.Fatal(err)
	}
}

func TestPrometheusAdditionalScrapeConfig(t *testing.T) {
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
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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

func TestPrometheusAdditionalAlertManagerConfig(t *testing.T) {
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
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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
		response, err := framework.QueryPrometheusSVC(ns, svc.Name, "/api/v1/alertmanagers", map[string]string{})
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

func TestPrometheusReloadRules(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	firtAlertName := "firstAlert"
	secondAlertName := "secondAlert"

	ruleFile, err := framework.MakeAndCreateFiringRuleFile(ns, name, firtAlertName)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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
		monitoringv1.RuleGroup{
			Name: "my-alerting-group",
			Rules: []monitoringv1.Rule{
				monitoringv1.Rule{
					Alert: secondAlertName,
					Expr:  "vector(1)",
				},
			},
		},
	}
	err = framework.UpdateRuleFile(ns, ruleFile)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusFiringAlert(p.Namespace, pSVC.Name, secondAlertName)
	if err != nil {
		t.Fatal(err)
	}
}

// With Prometheus Operator v0.20.0 the 'RuleSelector' field in the Prometheus
// CRD Spec is deprecated. We need to ensure to still support it until the field
// is removed. Any value in 'RuleSelector' should just be copied to the new
// field 'RuleFileSelector'.
func TestPrometheusDeprecatedRuleSelectorField(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	firtAlertName := "firstAlert"

	_, err := framework.MakeAndCreateFiringRuleFile(ns, name, firtAlertName)
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	// Reset new 'RuleFileSelector' field
	p.Spec.RuleFileSelector = nil
	// Specify old 'RuleFile' field
	p.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"role": "rulefile",
		},
	}
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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
}

func TestPrometheusMultipleRuleFilesSameNS(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	alertNames := []string{"first-alert", "second-alert"}

	for _, alertName := range alertNames {
		_, err := framework.MakeAndCreateFiringRuleFile(ns, alertName, alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
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

func TestPrometheusMultipleRuleFilesDifferentNS(t *testing.T) {
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
		testFramework.AddLabelsToNamespace(framework.KubeClient, file.ns, ruleFilesNamespaceSelector)
	}

	for _, file := range ruleFiles {
		_, err := framework.MakeAndCreateFiringRuleFile(file.ns, file.alertName, file.alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(rootNS, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p.Spec.RuleFileNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: ruleFilesNamespaceSelector,
	}
	if err := framework.CreatePrometheusAndWaitUntilReady(rootNS, p); err != nil {
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

// Make sure the Prometheus operator only updates the Prometheus config secret
// and the Prometheus rules configmap on relevant changes
func TestPrometheusOnlyUpdatedOnRelevantChanges(t *testing.T) {
	t.Parallel()

	testCTX := framework.NewTestCtx(t)
	defer testCTX.Cleanup(t)
	ns := testCTX.CreateNamespace(t, framework.KubeClient)
	testCTX.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)

	ctx, cancel := context.WithCancel(context.Background())

	type versionedResource interface {
		GetResourceVersion() string
	}

	// TODO: rename resouceDefinitions
	resourceDefinitions := []struct {
		Name            string
		Getter          func(prometheusName string) (versionedResource, error)
		Versions        map[string]interface{}
		ExpectedChanges int
	}{
		{
			Name: "crd",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					MonClientV1.
					Prometheuses(ns).
					Get(prometheusName, metav1.GetOptions{})
			},
			ExpectedChanges: 1,
		},
		{
			Name: "rulesConfigMap",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					ConfigMaps(ns).
					Get("prometheus-"+prometheusName+"-rules", metav1.GetOptions{})
			},
			ExpectedChanges: 1,
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
			ExpectedChanges: 1,
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
			// First is the creation of the StatefulSet itself, second is the
			// update of the ReadyReplicas status field
			ExpectedChanges: 2,
		},
		{
			Name: "service",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Services(ns).
					Get("prometheus-operated", metav1.GetOptions{})
			},
			ExpectedChanges: 1,
		},
	}

	// Init Versions maps
	for i, _ := range resourceDefinitions {
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

	if err := framework.CreatePrometheusAndWaitUntilReady(ns, prometheus); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}

	cancel()

	for _, resource := range resourceDefinitions {
		if len(resource.Versions) != resource.ExpectedChanges {
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
				resource.ExpectedChanges,
				len(resource.Versions),
			)
		}
	}
}

func TestPrometheusWhenDeleteCRDCleanUpViaOwnerReference(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	if err := framework.CreatePrometheusAndWaitUntilReady(ns, p); err != nil {
		t.Fatal(err)
	}

	configMapName := fmt.Sprintf("prometheus-%v-rules", p.Name)

	_, err := framework.WaitForConfigMapExist(ns, configMapName)
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

func TestPrometheusDiscovery(t *testing.T) {
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(s); err != nil {
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

func TestPromOpMatchPromAndServMonInDiffNSs(t *testing.T) {
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
	if err := framework.CreatePrometheusAndWaitUntilReady(prometheusNSName, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, prometheusNSName, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	resp, err := framework.QueryPrometheusSVC(prometheusNSName, svc.Name, "/api/v1/status/config", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(resp), prometheusJobName) != 1 {
		t.Fatalf("expected Prometheus operator to configure Prometheus in ns '%v' to scrape the service monitor in ns '%v'", prometheusNSName, serviceMonitorNSName)
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
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&rq); err != nil {
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
		if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&ra); err != nil {
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
