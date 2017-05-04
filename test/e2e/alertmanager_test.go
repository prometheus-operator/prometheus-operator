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
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	testFramework "github.com/coreos/prometheus-operator/test/e2e/framework"
)

func TestAlertmanagerCreateDeleteCluster(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeleteAlertmanagerAndWaitUntilGone(ctx.Id, name); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerScaling(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ctx.Id, framework.MakeBasicAlertmanager(name, 5)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ctx.Id, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerVersionMigration(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	name := "test"

	am := framework.MakeBasicAlertmanager(name, 1)
	am.Spec.Version = "v0.6.0"
	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.6.1"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ctx.Id, am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.6.0"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ctx.Id, am); err != nil {
		t.Fatal(err)
	}
}

func TestExposingAlertmanagerWithNodePort(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("test-alertmanager", 1)
	alertmanagerService := framework.MakeAlertmanagerNodePortService(alertmanager.Name, "nodeport-service", 30903)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, alertmanager); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, alertmanagerService); err != nil {
		t.Fatal(err)
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:30903/", framework.ClusterIP))
	if err != nil {
		t.Fatal("Retrieving alertmanager landing page failed with error: ", err)
	} else if resp.StatusCode != 200 {
		t.Fatal("Retrieving alertmanager landing page failed with http status code: ", resp.StatusCode)
	}
}

func TestExposingAlertmanagerWithKubernetesAPI(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("test-alertmanager", 1)
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	proxyGet := framework.KubeClient.CoreV1().Services(ctx.Id).ProxyGet
	request := proxyGet("", alertmanagerService.Name, "web", "/", make(map[string]string))
	_, err := request.DoRaw()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExposingAlertmanagerWithIngress(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("main", 1)
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "test-group", v1.ServiceTypeClusterIP)
	ingress := testFramework.MakeBasicIngress(alertmanagerService.Name, 9093)

	if err := testFramework.SetupNginxIngressControllerIncDefaultBackend(framework.KubeClient, ctx.Id); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	if err := testFramework.CreateIngress(framework.KubeClient, ctx.Id, ingress); err != nil {
		t.Fatal(err)
	}

	ip, err := testFramework.GetIngressIP(framework.KubeClient, ctx.Id, ingress.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = testFramework.WaitForHTTPSuccessStatusCode(time.Minute, fmt.Sprintf("http://%s/metrics", *ip))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMeshInitialization(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	var amountAlertmanagers int32 = 3
	alertmanager := &v1alpha1.Alertmanager{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.AlertmanagerSpec{
			Replicas: &amountAlertmanagers,
			Version:  "master",
		},
	}

	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ctx.Id, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < int(amountAlertmanagers); i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitializedMesh(ctx.Id, name, int(amountAlertmanagers)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAlertmanagerReloadConfig(t *testing.T) {
	ctx := testFramework.NewTestCtx(t)
	defer ctx.CleanUp(t)
	ctx.BasicSetup(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("reload-config", 1)

	firstConfig := `
global:
  resolve_timeout: 6m
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://alertmanagerwh:30500/'
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
  - url: 'http://alertmanagerwh:30500/'
`

	cfg := &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", alertmanager.Name),
		},
		Data: map[string][]byte{
			"alertmanager.yaml": []byte(firstConfig),
		},
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ctx.Id, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForSpecificAlertmanagerConfig(ctx.Id, alertmanager.Name, firstConfig); err != nil {
		t.Fatal(err)
	}

	cfg.Data["alertmanager.yaml"] = []byte(secondConfig)

	if _, err := framework.KubeClient.CoreV1().Secrets(ctx.Id).Update(cfg); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForSpecificAlertmanagerConfig(ctx.Id, alertmanager.Name, secondConfig); err != nil {
		t.Fatal(err)
	}
}
