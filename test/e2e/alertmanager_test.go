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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	testFramework "github.com/coreos/prometheus-operator/test/e2e/framework"
)

func TestAlertmanagerCreateDeleteCluster(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	name := "test"

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeleteAlertmanagerAndWaitUntilGone(ns, name); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerScaling(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	name := "test"

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 5)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerVersionMigration(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	name := "test"

	am := framework.MakeBasicAlertmanager(name, 1)
	am.Spec.Version = "v0.7.0"
	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.7.1"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.7.0"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(ns, am); err != nil {
		t.Fatal(err)
	}
}

func TestExposingAlertmanagerWithNodePort(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("test-alertmanager", 1)
	alertmanagerService := framework.MakeAlertmanagerNodePortService(alertmanager.Name, "nodeport-service", 30903)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
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
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("test-alertmanager", 1)
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	proxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
	request := proxyGet("", alertmanagerService.Name, "web", "/", make(map[string]string))
	_, err := request.DoRaw()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExposingAlertmanagerWithIngress(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	alertmanager := framework.MakeBasicAlertmanager("main", 1)
	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "test-group", v1.ServiceTypeClusterIP)
	ingress := testFramework.MakeBasicIngress(alertmanagerService.Name, 9093)

	if err := testFramework.SetupNginxIngressControllerIncDefaultBackend(framework.KubeClient, ns); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	if err := testFramework.CreateIngress(framework.KubeClient, ns, ingress); err != nil {
		t.Fatal(err)
	}

	ip, err := testFramework.GetIngressIP(framework.KubeClient, ns, ingress.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = testFramework.WaitForHTTPSuccessStatusCode(time.Minute, fmt.Sprintf("http://%s/metrics", *ip))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMeshInitialization(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

	var amountAlertmanagers int32 = 3
	alertmanager := &v1alpha1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.AlertmanagerSpec{
			Replicas: &amountAlertmanagers,
			Version:  "v0.7.1",
		},
	}

	alertmanagerService := framework.MakeAlertmanagerService(alertmanager.Name, "alertmanager-service", v1.ServiceTypeClusterIP)

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := testFramework.CreateServiceAndWaitUntilReady(framework.KubeClient, ns, alertmanagerService); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < int(amountAlertmanagers); i++ {
		name := "alertmanager-" + alertmanager.Name + "-" + strconv.Itoa(i)
		if err := framework.WaitForAlertmanagerInitializedMesh(ns, name, int(amountAlertmanagers)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAlertmanagerReloadConfig(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)

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
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", alertmanager.Name),
		},
		Data: map[string][]byte{
			"alertmanager.yaml": []byte(firstConfig),
		},
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns, alertmanager); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(cfg); err != nil {
		t.Fatal(err)
	}

	firstExpectedConfig := "global:\n  resolve_timeout: 6m\n  smtp_from: \"\"\n  smtp_smarthost: \"\"\n  smtp_auth_username: \"\"\n  smtp_auth_password: null\n  smtp_auth_secret: null\n  smtp_auth_identity: \"\"\n  smtp_require_tls: true\n  slack_api_url: null\n  pagerduty_url: https://events.pagerduty.com/generic/2010-04-15/create_event.json\n  hipchat_url: https://api.hipchat.com/\n  hipchat_auth_token: null\n  opsgenie_api_host: https://api.opsgenie.com/\n  victorops_api_url: https://alert.victorops.com/integrations/generic/20131114/alert/\nroute:\n  receiver: webhook\n  group_by:\n  - job\n  group_wait: 30s\n  group_interval: 5m\n  repeat_interval: 12h\nreceivers:\n- name: webhook\n  webhook_configs:\n  - send_resolved: true\n    url: http://alertmanagerwh:30500/\ntemplates: []\n"
	if err := framework.WaitForSpecificAlertmanagerConfig(ns, alertmanager.Name, firstExpectedConfig); err != nil {
		t.Fatal(err)
	}

	cfg.Data["alertmanager.yaml"] = []byte(secondConfig)

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(cfg); err != nil {
		t.Fatal(err)
	}

	secondExpectedConfig := "global:\n  resolve_timeout: 5m\n  smtp_from: \"\"\n  smtp_smarthost: \"\"\n  smtp_auth_username: \"\"\n  smtp_auth_password: null\n  smtp_auth_secret: null\n  smtp_auth_identity: \"\"\n  smtp_require_tls: true\n  slack_api_url: null\n  pagerduty_url: https://events.pagerduty.com/generic/2010-04-15/create_event.json\n  hipchat_url: https://api.hipchat.com/\n  hipchat_auth_token: null\n  opsgenie_api_host: https://api.opsgenie.com/\n  victorops_api_url: https://alert.victorops.com/integrations/generic/20131114/alert/\nroute:\n  receiver: webhook\n  group_by:\n  - job\n  group_wait: 30s\n  group_interval: 5m\n  repeat_interval: 12h\nreceivers:\n- name: webhook\n  webhook_configs:\n  - send_resolved: true\n    url: http://alertmanagerwh:30500/\ntemplates: []\n"
	if err := framework.WaitForSpecificAlertmanagerConfig(ns, alertmanager.Name, secondExpectedConfig); err != nil {
		t.Fatal(err)
	}
}
