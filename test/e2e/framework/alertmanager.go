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
	"encoding/json"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/pkg/util/yaml"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/pkg/errors"
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
			Replicas: &replicas,
		},
	}
}

func (f *Framework) MakeAlertmanagerNodePortService(name, group string, nodePort int32) *v1.Service {
	aMService := f.MakeAlertmanagerService(name, group, v1.ServiceTypeNodePort)
	aMService.Spec.Ports[0].NodePort = nodePort
	return aMService
}

func (f *Framework) MakeAlertmanagerService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", name),
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "web",
					Port:       9093,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"alertmanager": name,
			},
		},
	}

	return service
}

func (f *Framework) SecretFromYaml(filepath string) (*v1.Secret, error) {
	manifest, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	s := v1.Secret{}
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (f *Framework) AlertmanagerConfigSecret(name string) (*v1.Secret, error) {
	s, err := f.SecretFromYaml("../../contrib/kube-prometheus/manifests/alertmanager/alertmanager-config.yaml")
	if err != nil {
		return nil, err
	}

	s.Name = name
	return s, nil
}

func (f *Framework) CreateAlertmanagerAndWaitUntilReady(ns string, a *v1alpha1.Alertmanager) error {
	amConfigSecretName := fmt.Sprintf("alertmanager-%s", a.Name)
	s, err := f.AlertmanagerConfigSecret(amConfigSecretName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("making alertmanager config secret %v failed", amConfigSecretName))
	}
	_, err = f.KubeClient.CoreV1().Secrets(ns).Create(s)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating alertmanager config secret %v failed", s.Name))
	}

	_, err = f.MonClient.Alertmanagers(ns).Create(a)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating alertmanager %v failed", a.Name))
	}

	err = WaitForPodsReady(
		f.KubeClient,
		ns,
		5*time.Minute,
		int(*a.Spec.Replicas),
		alertmanager.ListOptions(a.Name),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create an Alertmanager cluster (%s) with %d instances", a.Name, a.Spec.Replicas))
	}
	return nil
}

func (f *Framework) UpdateAlertmanagerAndWaitUntilReady(ns string, a *v1alpha1.Alertmanager) error {
	_, err := f.MonClient.Alertmanagers(ns).Update(a)
	if err != nil {
		return err
	}

	err = WaitForPodsReady(
		f.KubeClient,
		ns,
		5*time.Minute,
		int(*a.Spec.Replicas),
		alertmanager.ListOptions(a.Name),
	)
	if err != nil {
		return fmt.Errorf("failed to update %d Alertmanager instances (%s): %v", a.Spec.Replicas, a.Name, err)
	}

	return nil
}

func (f *Framework) DeleteAlertmanagerAndWaitUntilGone(ns, name string) error {
	_, err := f.MonClient.Alertmanagers(ns).Get(name)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting Alertmanager tpr %v failed", name))
	}

	if err := f.MonClient.Alertmanagers(ns).Delete(name, nil); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting Alertmanager tpr %v failed", name))
	}

	if err := WaitForPodsReady(
		f.KubeClient,
		ns,
		f.DefaultTimeout,
		0,
		alertmanager.ListOptions(name),
	); err != nil {
		return errors.Wrap(err, fmt.Sprintf("waiting for Alertmanager tpr (%s) to vanish timed out", name))
	}

	return f.KubeClient.CoreV1().Secrets(ns).Delete(fmt.Sprintf("alertmanager-%s", name), nil)
}

func amImage(version string) string {
	return fmt.Sprintf("quay.io/prometheus/alertmanager:%s", version)
}

func (f *Framework) WaitForAlertmanagerInitializedMesh(ns, name string, amountPeers int) error {
	return wait.Poll(time.Second, time.Second*20, func() (bool, error) {
		amStatus, err := f.GetAlertmanagerConfig(ns, name)
		if err != nil {
			return false, err
		}
		if len(amStatus.Data.MeshStatus.Peers) == amountPeers {
			return true, nil
		}

		return false, nil
	})
}

func (f *Framework) GetAlertmanagerConfig(ns, n string) (alertmanagerStatus, error) {
	var amStatus alertmanagerStatus
	request := ProxyGetPod(f.KubeClient, ns, n, "9093", "/api/v1/status")
	resp, err := request.DoRaw()
	if err != nil {
		return amStatus, err
	}

	if err := json.Unmarshal(resp, &amStatus); err != nil {
		return amStatus, err
	}

	return amStatus, nil
}

func (f *Framework) WaitForSpecificAlertmanagerConfig(ns, amName string, expectedConfig string) error {
	return wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		config, err := f.GetAlertmanagerConfig(ns, "alertmanager-"+amName+"-0")
		if err != nil {
			return false, err
		}

		if config.Data.Config == expectedConfig {
			return true, nil
		}

		return false, nil
	})
}

type alertmanagerStatus struct {
	Data alertmanagerStatusData `json:"data"`
}

type alertmanagerStatusData struct {
	MeshStatus meshStatus `json:"meshStatus"`
	Config     string     `json:"config"`
}

type meshStatus struct {
	Peers []interface{} `json:"peers"`
}
