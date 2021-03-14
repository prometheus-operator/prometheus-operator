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
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testAMCreateDeleteCluster(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeleteAlertmanagerAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func testAMScaling(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	a, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 3))
	if err != nil {
		t.Fatal(err)
	}

	a.Spec.Replicas = proto.Int32(5)
	a, err = framework.UpdateAlertmanagerAndWaitUntilReady(ns, a)
	if err != nil {
		t.Fatal(err)
	}

	a.Spec.Replicas = proto.Int32(3)
	if _, err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, a); err != nil {
		t.Fatal(err)
	}
}

func testAMVersionMigration(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	am := framework.MakeBasicAlertmanager(name, 1)
	am.Spec.Version = "v0.16.2"
	am, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, am)
	if err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.17.0"
	am, err = framework.UpdateAlertmanagerAndWaitUntilReady(ns, am)
	if err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.16.2"
	_, err = framework.UpdateAlertmanagerAndWaitUntilReady(ns, am)
	if err != nil {
		t.Fatal(err)
	}
}

func testAMStorageUpdate(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	name := "test"

	am := framework.MakeBasicAlertmanager(name, 1)

	am, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, am)
	if err != nil {
		t.Fatal(err)
	}

	am.Spec.Storage = &monitoringv1.StorageSpec{
		VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
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

	_, err = framework.MonClientV1.Alertmanagers(ns).Update(context.TODO(), am, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.TODO(), alertmanager.ListOptions(name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		for _, volume := range pods.Items[0].Spec.Volumes {
			if volume.Name == "alertmanager-"+name+"-db" && volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testAMExposingWithKubernetesAPI(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("test-alertmanager", 1)
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	proxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
	request := proxyGet("", alertmanagerService.Name, "web", "/", make(map[string]string))
	_, err := request.DoRaw(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
}

func testAMClusterInitialization(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	amClusterSize := 3
	alertmanager := framework.MakeBasicAlertmanager("test", int32(amClusterSize))
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	// Print Alertmanager logs on failure.
	defer func() {
		if !t.Failed() {
			return
		}

		for i := 0; i < amClusterSize; i++ {
			err := framework.PrintPodLogs(ns, fmt.Sprintf("alertmanager-test-%v", strconv.Itoa(i)))
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < amClusterSize; i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitialized(ns, name, amClusterSize, alertmanager.Spec.ForceEnableClusterMode); err != nil {
			t.Fatal(err)
		}
	}
}

// testAMClusterAfterRollingUpdate tests whether all Alertmanager instances join
// the cluster after a rolling update, even though DNS records will probably be
// outdated at startup time. See
// https://github.com/prometheus/alertmanager/pull/1428 for more details.
func testAMClusterAfterRollingUpdate(t *testing.T) {
	var err error

	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	amClusterSize := 3

	alertmanager := framework.MakeBasicAlertmanager("test", int32(amClusterSize))

	if alertmanager, err = framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < amClusterSize; i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitialized(ns, name, amClusterSize, alertmanager.Spec.ForceEnableClusterMode); err != nil {
			t.Fatal(err)
		}
	}

	// We need to force a rolling update, e.g. by changing one of the command
	// line flags via the Retention.
	alertmanager.Spec.Retention = "1h"

	if _, err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < amClusterSize; i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitialized(ns, name, amClusterSize, alertmanager.Spec.ForceEnableClusterMode); err != nil {
			t.Fatal(err)
		}
	}
}

func testAMClusterGossipSilences(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	amClusterSize := 3
	alertmanager := framework.MakeBasicAlertmanager("test", int32(amClusterSize))

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < amClusterSize; i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitialized(ns, name, amClusterSize, alertmanager.Spec.ForceEnableClusterMode); err != nil {
			t.Fatal(err)
		}
	}

	silID, err := framework.CreateSilence(ns, "alertmanager-test-0")
	if err != nil {
		t.Fatalf("failed to create silence: %v", err)
	}

	for i := 0; i < amClusterSize; i++ {
		err = wait.Poll(time.Second, framework.DefaultTimeout, func() (bool, error) {
			silences, err := framework.GetSilences(ns, "alertmanager-"+alertmanager.Name+"-"+strconv.Itoa(i))
			if err != nil {
				return false, err
			}

			if len(silences) != 1 {
				return false, nil
			}

			if *silences[0].ID != silID {
				return false, errors.Errorf("expected silence id on alertmanager %v to match id of created silence '%v' but got %v", i, silID, *silences[0].ID)
			}
			return true, nil
		})
		if err != nil {
			t.Fatalf("could not retrieve created silence on alertmanager %v: %v", i, err)
		}
	}
}

func testAMReloadConfig(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("reload-config", 1)
	templateResourceName := fmt.Sprintf("alertmanager-templates-%s", alertmanager.Name)
	alertmanager.Spec.ConfigMaps = []string{templateResourceName}
	alertmanager.Spec.Secrets = []string{templateResourceName}

	firstConfig := `
global:
  resolve_timeout: 5m
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://firstConfigWebHook:30500/'
`
	secondConfig := `
global:
  resolve_timeout: 5m
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://secondConfigWebHook:30500/'
`
	template := `
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
  <meta name="viewport" content="width=device-width" />
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <title>An Alert</title>
  <style>
  </style>
</head>
`

	secondTemplate := `
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
  <meta name="viewport" content="width=device-width" />
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <title>An Alert</title>
  <style>
  </style>
</head>

<body>
An Alert test
</body>
`

	cfg := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", alertmanager.Name),
		},
		Data: map[string][]byte{
			"alertmanager.yaml": []byte(firstConfig),
		},
	}

	templateFileKey := "test-emails.tmpl"
	templateSecretFileKey := "test-emails-secret.tmpl"
	templateCfg := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateResourceName,
		},
		Data: map[string]string{
			templateFileKey: template,
		},
	}
	templateSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateResourceName,
		},
		Data: map[string][]byte{
			templateSecretFileKey: []byte(template),
		},
	}

	if _, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.TODO(), templateCfg, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), templateSecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(context.TODO(), cfg, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	firstExpectedString := "firstConfigWebHook"
	if err := framework.WaitForAlertmanagerConfigToContainString(ns, alertmanager.Name, firstExpectedString); err != nil {
		t.Fatal(errors.Wrap(err, "failed to wait for first expected config"))
	}
	cfg.Data["alertmanager.yaml"] = []byte(secondConfig)

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(context.TODO(), cfg, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	secondExpectedString := "secondConfigWebHook"

	if err := framework.WaitForAlertmanagerConfigToContainString(ns, alertmanager.Name, secondExpectedString); err != nil {
		t.Fatal(errors.Wrap(err, "failed to wait for second expected config"))
	}

	priorToReloadTime := time.Now()
	templateCfg.Data[templateFileKey] = secondTemplate
	if _, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Update(context.TODO(), templateCfg, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForAlertmanagerConfigToBeReloaded(ns, alertmanager.Name, priorToReloadTime); err != nil {
		t.Fatal(errors.Wrap(err, "failed to wait for additional configMaps reload"))
	}

	priorToReloadTime = time.Now()
	templateSecret.Data[templateSecretFileKey] = []byte(secondTemplate)
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(context.TODO(), templateSecret, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForAlertmanagerConfigToBeReloaded(ns, alertmanager.Name, priorToReloadTime); err != nil {
		t.Fatal(errors.Wrap(err, "failed to wait for additional secrets reload"))
	}
}

func testAMZeroDowntimeRollingDeployment(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	whReplicas := int32(1)
	whdpl := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "alertmanager-webhook",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &whReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "alertmanager-webhook",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "alertmanager-webhook",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "webhook-server",
							Image: "quay.io/coreos/prometheus-alertmanager-test-webhook",
							Ports: []v1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 5001,
								},
							},
						},
					},
				},
			},
		},
	}
	whsvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "alertmanager-webhook",
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       5001,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"app": "alertmanager-webhook",
			},
		},
	}
	if err := testFramework.CreateDeployment(framework.KubeClient, ns, whdpl); err != nil {
		t.Fatal(err)
	}
	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, whsvc); err != nil {
		t.Fatal(err)
	}
	err := testFramework.WaitForPodsReady(framework.KubeClient, ns, time.Minute*5, 1,
		metav1.ListOptions{
			LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
				"app": "alertmanager-webhook",
			})).String(),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	alertmanager := framework.MakeBasicAlertmanager("rolling-deploy", 3)
	amsvc := framework.MakeAlertmanagerService(alertmanager.Name, "test", v1.ServiceTypeClusterIP)
	amcfg := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", alertmanager.Name),
		},
		Data: map[string][]byte{
			"alertmanager.yaml": []byte(fmt.Sprintf(`
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://%s.%s.svc:5001/'
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
`, whsvc.Name, ns)),
		},
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), amcfg, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	alertmanager, err = framework.MonClientV1.Alertmanagers(ns).Create(context.TODO(), alertmanager, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForAlertmanagerReady(ns, alertmanager.Name, int(*alertmanager.Spec.Replicas), alertmanager.Spec.ForceEnableClusterMode); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, amsvc); err != nil {
		t.Fatal(err)
	}

	// Send alert to each Alertmanager
	for i := 0; i < int(*alertmanager.Spec.Replicas); i++ {
		replica := i
		done := make(chan struct{})
		errc := make(chan error, 1)

		defer func() {
			close(done)
			select {
			case err := <-errc:
				t.Fatal(errors.Wrapf(err, "sending alert to alertmanager %v", replica))
			default:
				return
			}
		}()

		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			failures := 0
			for {
				select {
				case <-ticker.C:
					err := framework.SendAlertToAlertmanager(
						ns,
						"alertmanager-rolling-deploy-"+strconv.Itoa(replica),
					)
					if err != nil {
						failures++
						// Allow 50 (~5 Seconds) failures during Alertmanager rolling update.
						if failures > 50 {
							errc <- err
							return
						}
					}
				case <-done:
					return
				}

			}
		}()
	}

	// Wait for alert to propagate
	time.Sleep(30 * time.Second)

	opts := metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app": "alertmanager-webhook",
		})).String(),
	}
	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(pl.Items) != 1 {
		t.Fatalf("Expected one webhook pod, but got %d", len(pl.Items))
	}

	podName := pl.Items[0].Name
	logs, err := testFramework.GetLogs(framework.KubeClient, ns, podName, "webhook-server")
	if err != nil {
		t.Fatal(err)
	}

	c := strings.Count(logs, "Alertmanager Notification Payload Received")
	if c != 1 {
		t.Fatalf("One notification expected, but %d received.\n\n%s", c, logs)
	}

	// We need to force a rolling update, e.g. by changing one of the command
	// line flags via the Retention.
	alertmanager.Spec.Retention = "1h"
	if _, err := framework.MonClientV1.Alertmanagers(ns).Update(context.TODO(), alertmanager, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	// Wait for the change above to take effect.
	time.Sleep(time.Minute)

	if err := framework.WaitForAlertmanagerReady(ns, alertmanager.Name, int(*alertmanager.Spec.Replicas), alertmanager.Spec.ForceEnableClusterMode); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Minute)

	logs, err = testFramework.GetLogs(framework.KubeClient, ns, podName, "webhook-server")
	if err != nil {
		t.Fatal(err)
	}

	c = strings.Count(logs, "Alertmanager Notification Payload Received")
	if c != 1 {
		t.Fatalf("Only one notification expected, but %d received after rolling update of Alertmanager cluster.\n\n%s", c, logs)
	}
}

func testAlertmanagerConfigCRD(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("amconfig-crd", 1)
	alertmanager.Spec.AlertmanagerConfigSelector = &metav1.LabelSelector{}
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	// reuse the secret for pagerduty and wechat
	testingSecret := "testing-secret"
	testingSecretKey := "testing-secret-key"
	testingKeySecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testingSecret,
		},
		Data: map[string][]byte{
			testingSecretKey: []byte("1234abc"),
		},
	}
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), testingKeySecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	apiKeySecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "og-receiver-api-key",
		},
		Data: map[string][]byte{
			"api-key": []byte("1234abc"),
		},
	}
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), apiKeySecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	slackAPIURLSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "s-receiver-api-url",
		},
		Data: map[string][]byte{
			"api-url": []byte("http://slack.example.com"),
		},
	}
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), slackAPIURLSecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// A valid AlertmanagerConfig resource with many receivers.
	configCR := &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-amconfig-many-receivers",
			Namespace: ns,
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "e2e",
				Matchers: []monitoringv1alpha1.Matcher{},
				Continue: true,
			},
			Receivers: []monitoringv1alpha1.Receiver{{
				Name: "e2e",
				OpsGenieConfigs: []monitoringv1alpha1.OpsGenieConfig{{
					APIKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "og-receiver-api-key",
						},
						Key: "api-key",
					},
				}},
				PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
					RoutingKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
				}},
				SlackConfigs: []monitoringv1alpha1.SlackConfig{{
					APIURL: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "s-receiver-api-url",
						},
						Key: "api-url",
					},
					Actions: []monitoringv1alpha1.SlackAction{
						{
							Type: "type",
							Text: "text",
							Name: "my-action",
							ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
								Text: "text",
							},
						},
					},
					Fields: []monitoringv1alpha1.SlackField{
						{
							Title: "title",
							Value: "value",
						},
					},
				}},
				WebhookConfigs: []monitoringv1alpha1.WebhookConfig{{
					URL: func(s string) *string {
						return &s
					}("http://test.url"),
				}},
				WeChatConfigs: []monitoringv1alpha1.WeChatConfig{{
					APISecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
					CorpID: "testingCorpID",
				}},
				EmailConfigs: []monitoringv1alpha1.EmailConfig{{
					SendResolved: func(b bool) *bool {
						return &b
					}(true),
					To: "test@example.com",
					AuthPassword: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
					AuthSecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
					Headers: []monitoringv1alpha1.KeyValue{
						{Key: "Subject", Value: "subject"},
						{Key: "Comment", Value: "comment"},
					},
				}},
				VictorOpsConfigs: []monitoringv1alpha1.VictorOpsConfig{{
					APIKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
					RoutingKey: "abc",
				}},
				PushoverConfigs: []monitoringv1alpha1.PushoverConfig{{
					UserKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
					Token: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: testingSecretKey,
					},
				}},
			}},
		},
	}

	if _, err := framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(context.TODO(), configCR, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// Another AlertmanagerConfig object with nested routes.
	configCR = &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-amconfig-sub-routes",
			Namespace: ns,
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "e2e",
				Matchers: []monitoringv1alpha1.Matcher{
					{Name: "service", Value: "webapp"},
				},
				Routes: []apiextensionsv1.JSON{
					{Raw: []byte(`
{
  "receiver": "e2e",
  "groupBy": ["env", "instance"],
  "matchers": [
    {
      "name": "job",
      "value": "db"
    }
  ],
  "routes": [
    {
      "receiver": "e2e",
      "matchers": [
        {
          "name": "alertname",
          "value": "TargetDown"
        }
      ]
    },
    {
      "receiver": "e2e",
      "matchers": [
        {
          "name": "severity",
          "value": "critical|warning",
          "regex": true
        }
      ]
    }
  ]
}
					`)},
				},
			},
			Receivers: []monitoringv1alpha1.Receiver{{
				Name: "e2e",
				WebhookConfigs: []monitoringv1alpha1.WebhookConfig{{
					URL: func(s string) *string {
						return &s
					}("http://test.url"),
				}},
			}},
		},
	}

	if _, err := framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(context.TODO(), configCR, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// An AlertmanagerConfig resource that references a missing secret key, it
	// should be rejected by the operator.
	configCR = &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-amconfig-missing-secret",
			Namespace: ns,
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "e2e",
				Matchers: []monitoringv1alpha1.Matcher{},
			},
			Receivers: []monitoringv1alpha1.Receiver{{
				Name: "e2e",
				PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
					RoutingKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: "non-existing-key",
					},
				}},
			}},
		},
	}

	if _, err := framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(context.TODO(), configCR, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// An AlertmanagerConfig resource that contains an invalid sub-route.
	// It should be rejected by the operator.
	configCR = &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-amconfig-invalid-route",
			Namespace: ns,
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			Route: &monitoringv1alpha1.Route{
				Receiver: "e2e",
				Matchers: []monitoringv1alpha1.Matcher{},
				Routes: []apiextensionsv1.JSON{
					{Raw: []byte(`"invalid"`)},
				},
			},
			Receivers: []monitoringv1alpha1.Receiver{{
				Name: "e2e",
				PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
					RoutingKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testingSecret,
						},
						Key: "non-existing-key",
					},
				}},
			}},
		},
	}

	if _, err := framework.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(context.TODO(), configCR, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// Wait for the change above to take effect.
	var lastErr error
	err := wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		cfgSecret, err := framework.KubeClient.CoreV1().Secrets(ns).Get(context.TODO(), "alertmanager-amconfig-crd-generated", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			lastErr = errors.New("Generated configuration secret not found")
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if cfgSecret.Data["alertmanager.yaml"] == nil {
			lastErr = errors.New("'alertmanager.yaml' key is missing")
			return false, nil
		}

		expected := fmt.Sprintf(`global:
  resolve_timeout: 5m
route:
  receiver: "null"
  group_by:
  - job
  routes:
  - receiver: %s-e2e-test-amconfig-many-receivers-e2e
    match:
      namespace: %s
    continue: true
  - receiver: %s-e2e-test-amconfig-sub-routes-e2e
    match:
      namespace: %s
      service: webapp
    continue: true
    routes:
    - receiver: %s-e2e-test-amconfig-sub-routes-e2e
      group_by:
      - env
      - instance
      match:
        job: db
      routes:
      - receiver: %s-e2e-test-amconfig-sub-routes-e2e
        match:
          alertname: TargetDown
      - receiver: %s-e2e-test-amconfig-sub-routes-e2e
        match_re:
          severity: critical|warning
  - receiver: "null"
    match:
      alertname: DeadMansSwitch
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
receivers:
- name: "null"
- name: %v-e2e-test-amconfig-many-receivers-e2e
  opsgenie_configs:
  - api_key: 1234abc
  pagerduty_configs:
  - routing_key: 1234abc
  slack_configs:
  - api_url: http://slack.example.com
    fields:
    - title: title
      value: value
    actions:
    - type: type
      text: text
      name: my-action
      confirm:
        text: text
  webhook_configs:
  - url: http://test.url
  wechat_configs:
  - api_secret: 1234abc
    corp_id: testingCorpID
  email_configs:
  - send_resolved: true
    to: test@example.com
    auth_password: 1234abc
    auth_secret: 1234abc
    headers:
      Comment: comment
      Subject: subject
  pushover_configs:
  - user_key: 1234abc
    token: 1234abc
  victorops_configs:
  - api_key: 1234abc
    routing_key: abc
- name: %s-e2e-test-amconfig-sub-routes-e2e
  webhook_configs:
  - url: http://test.url
templates: []
`, ns, ns, ns, ns, ns, ns, ns, ns, ns)

		if diff := cmp.Diff(string(cfgSecret.Data["alertmanager.yaml"]), expected); diff != "" {
			t.Log("got(-), want(+):\n" + diff)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		t.Fatalf("%v: %v", err, lastErr)
	}
}

func testUserDefinedAlertmanagerConfig(t *testing.T) {
	// Don't run Alertmanager tests in parallel. See
	// https://github.com/prometheus/alertmanager/issues/1835 for details.

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	yamlConfig := `route:
  receiver: "void"
receivers:
- name: "void"
`
	amConfig := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amconfig",
		},
		Data: map[string][]byte{
			"alertmanager.yaml": []byte(yamlConfig),
			"template1.tmpl":    []byte(`template1`),
		},
	}
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.TODO(), amConfig, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	alertmanager := framework.MakeBasicAlertmanager("user-amconfig", 1)
	alertmanager.Spec.ConfigSecret = "amconfig"
	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	// Wait for the change above to take effect.
	var lastErr error
	err := wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		cfgSecret, err := framework.KubeClient.CoreV1().Secrets(ns).Get(context.TODO(), "alertmanager-user-amconfig-generated", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			lastErr = err
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if cfgSecret.Data["template1.tmpl"] == nil {
			lastErr = errors.New("'template1.yaml' key is missing")
			return false, nil
		}

		if cfgSecret.Data["alertmanager.yaml"] == nil {
			lastErr = errors.New("'alertmanager.yaml' key is missing")
			return false, nil
		}

		if string(cfgSecret.Data["alertmanager.yaml"]) != yamlConfig {
			lastErr = errors.Errorf("expected Alertmanager configuration %q, got %q", yamlConfig, cfgSecret.Data["alertmanager.yaml"])
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		t.Fatalf("%v: %v", err, lastErr)
	}
}

func testAMPreserveUserAddedMetadata(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns, framework.KubeClient)

	name := "test"

	alertManager := framework.MakeBasicAlertmanager(name, 3)

	alertManager, err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertManager)
	if err != nil {
		t.Fatal(err)
	}

	updatedLabels := map[string]string{
		"user-defined-label": "custom-label-value",
	}
	updatedAnnotations := map[string]string{
		"user-defined-annotation": "custom-annotation-val",
	}

	svcClient := framework.KubeClient.CoreV1().Services(ns)
	ssetClient := framework.KubeClient.AppsV1().StatefulSets(ns)
	secretClient := framework.KubeClient.CoreV1().Secrets(ns)

	resourceConfigs := []struct {
		name   string
		get    func() (metav1.Object, error)
		update func(object metav1.Object) (metav1.Object, error)
	}{
		{
			name: "alertmanager-operated service",
			get: func() (metav1.Object, error) {
				return svcClient.Get(context.TODO(), "alertmanager-operated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return svcClient.Update(context.TODO(), asService(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "alertmanager stateful set",
			get: func() (metav1.Object, error) {
				return ssetClient.Get(context.TODO(), "alertmanager-test", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return ssetClient.Update(context.TODO(), asStatefulSet(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "alertmanager secret",
			get: func() (metav1.Object, error) {
				return secretClient.Get(context.TODO(), "alertmanager-test-generated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return secretClient.Update(context.TODO(), asSecret(t, object), metav1.UpdateOptions{})
			},
		},
	}

	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		if err != nil {
			t.Fatal(err)
		}

		updateObjectLabels(res, updatedLabels)
		updateObjectAnnotations(res, updatedAnnotations)

		_, err = rConf.update(res)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Ensure resource reconciles
	alertManager.Spec.Replicas = proto.Int32(2)
	_, err = framework.UpdateAlertmanagerAndWaitUntilReady(ns, alertManager)
	if err != nil {
		t.Fatal(err)
	}

	// Assert labels preserved
	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		if err != nil {
			t.Fatal(err)
		}

		labels := res.GetLabels()
		if !containsValues(labels, updatedLabels) {
			t.Errorf("%s: labels do not contain updated labels, found: %q, should contain: %q", rConf.name, labels, updatedLabels)
		}

		annotations := res.GetAnnotations()
		if !containsValues(annotations, updatedAnnotations) {
			t.Fatalf("%s: annotations do not contain updated annotations, found: %q, should contain: %q", rConf.name, annotations, updatedAnnotations)
		}
	}

	if err := framework.DeleteAlertmanagerAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}
