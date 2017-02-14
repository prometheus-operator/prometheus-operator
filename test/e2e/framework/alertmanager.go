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
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"

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
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.AlertmanagerSpec{
			Replicas: replicas,
			Version:  "v0.5.0",
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
		ObjectMeta: metav1.ObjectMeta{
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

func (f *Framework) CreateAlertmanagerAndWaitUntilReady(a *v1alpha1.Alertmanager) error {
	log.Printf("Creating Alertmanager (%s/%s)", f.Namespace.Name, a.Name)
	_, err := f.KubeClient.CoreV1().ConfigMaps(f.Namespace.Name).Create(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("alertmanager-%s", a.Name),
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

	_, err = f.WaitForPodsReady(time.Minute*6, int(a.Spec.Replicas), amImage(a.Spec.Version), alertmanager.ListOptions(a.Name))
	if err != nil {
		return fmt.Errorf("failed to create an Alertmanager cluster (%s) with %d instances: %v", a.Name, a.Spec.Replicas, err)
	}
	return nil
}

func (f *Framework) UpdateAlertmanagerAndWaitUntilReady(a *v1alpha1.Alertmanager) error {
	log.Printf("Updating Alertmanager (%s/%s)", f.Namespace.Name, a.Name)
	_, err := f.MonClient.Alertmanagers(f.Namespace.Name).Update(a)
	if err != nil {
		return err
	}

	_, err = f.WaitForPodsReady(time.Minute*6, int(a.Spec.Replicas), amImage(a.Spec.Version), alertmanager.ListOptions(a.Name))
	if err != nil {
		return fmt.Errorf("failed to update %d Alertmanager instances (%s): %v", a.Spec.Replicas, a.Name, err)
	}

	return nil
}

func (f *Framework) DeleteAlertmanagerAndWaitUntilGone(name string) error {
	log.Printf("Deleting Alertmanager (%s/%s)", f.Namespace.Name, name)
	a, err := f.MonClient.Alertmanagers(f.Namespace.Name).Get(name)
	if err != nil {
		return err
	}

	if err := f.MonClient.Alertmanagers(f.Namespace.Name).Delete(name, nil); err != nil {
		return err
	}

	if _, err := f.WaitForPodsReady(time.Minute*6, 0, amImage(a.Spec.Version), alertmanager.ListOptions(name)); err != nil {
		return fmt.Errorf("failed to teardown Alertmanager (%s) instances: %v", name, err)
	}

	return f.KubeClient.CoreV1().ConfigMaps(f.Namespace.Name).Delete(fmt.Sprintf("alertmanager-%s", name), nil)
}

func amImage(version string) string {
	return fmt.Sprintf("quay.io/prometheus/alertmanager:%s", version)
}
