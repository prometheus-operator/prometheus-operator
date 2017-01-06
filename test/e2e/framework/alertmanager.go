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

package framework

import (
	"fmt"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

var ValidAlertmanagerConfig = `global:
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

func (f *Framework) MakeBasicAlertmanager(name string, replicas int32) *v1alpha1.Alertmanager {
	return &v1alpha1.Alertmanager{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.AlertmanagerSpec{
			Replicas: replicas,
		},
	}
}

func (f *Framework) CreateAlertmanagerAndWaitUntilReady(a *v1alpha1.Alertmanager) error {
	_, err := f.KubeClient.CoreV1().ConfigMaps(f.Namespace.Name).Create(
		&v1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name: a.Name,
			},
			Data: map[string]string{
				"alertmanager.yaml": ValidAlertmanagerConfig,
			},
		},
	)
	if err != nil {
		return err
	}

	_, err = f.MonClient.Alertmanagers(f.Namespace.Name).Create(a)
	if err != nil {
		return err
	}

	_, err = f.WaitForPodsReady(time.Minute*2, int(a.Spec.Replicas), alertmanager.ListOptions(a.Name))
	if err != nil {
		return fmt.Errorf("failed to create an Alertmanager cluster (%s) with %d instances: %v", a.Name, a.Spec.Replicas, err)
	}
	return nil
}

func (f *Framework) DeleteAlertmanagerAndWaitUntilGone(name string) error {
	if err := f.MonClient.Alertmanagers(f.Namespace.Name).Delete(name, nil); err != nil {
		return err
	}

	if _, err := f.WaitForPodsReady(time.Minute*2, 0, alertmanager.ListOptions(name)); err != nil {
		return fmt.Errorf("failed to teardown Alertmanager (%s) instances: %v", name, err)
	}

	return f.KubeClient.CoreV1().ConfigMaps(f.Namespace.Name).Delete(name, nil)
}
