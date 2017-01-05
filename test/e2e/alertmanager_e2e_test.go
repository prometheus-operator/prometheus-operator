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
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/spec"
)

var validAlertmanagerConfig = `global:
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

func TestAlertmanagerCreateCluster(t *testing.T) {
	name := "alertmanager-test"

	framework.KubeClient.CoreV1().ConfigMaps(framework.Namespace.Name).Create(
		&v1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
			Data: map[string]string{
				"alertmanager.yaml": validAlertmanagerConfig,
			},
		},
	)

	spec := &spec.Alertmanager{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: spec.AlertmanagerSpec{
			Replicas: 3,
		},
	}

	_, err := framework.CreateAlertmanager(spec)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := framework.DeleteAlertmanager(name); err != nil {
			t.Fatal(err)
		}

		if _, err := framework.WaitForPodsReady(time.Minute*2, 0, alertmanager.ListOptions(name)); err != nil {
			t.Fatalf("failed to teardown Alertmanager instances: %v", err)
		}
	}()

	if _, err := framework.WaitForPodsReady(time.Minute*2, 3, alertmanager.ListOptions(name)); err != nil {
		t.Fatalf("failed to create an Alertmanager cluster with 3 instances: %v", err)
	}
}
