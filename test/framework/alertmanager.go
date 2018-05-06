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
	"log"
	"os"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
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

func (f *Framework) MakeBasicAlertmanager(name string, replicas int32) *monitoringv1.Alertmanager {
	return &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Replicas: &replicas,
		},
	}
}

func (f *Framework) MakeBasicAlertmanagerV1alpha1(name string, replicas int32) *v1alpha1.Alertmanager {
	return &v1alpha1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.AlertmanagerSpec{
			Replicas: &replicas,
		},
	}
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

func (f *Framework) AlertmanagerConfigSecret(ns, name string) (*v1.Secret, error) {
	s, err := f.SecretFromYaml("../../test/framework/ressources/alertmanager-main-secret.yaml")
	if err != nil {
		return nil, err
	}

	s.Name = name
	s.Namespace = ns
	return s, nil
}

func (f *Framework) CreateAlertmanagerAndWaitUntilReady(ns string, a *monitoringv1.Alertmanager) error {
	amConfigSecretName := fmt.Sprintf("alertmanager-%s", a.Name)
	s, err := f.AlertmanagerConfigSecret(ns, amConfigSecretName)
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

	return f.WaitForAlertmanagerReady(ns, a.Name, int(*a.Spec.Replicas))
}

func (f *Framework) WaitForAlertmanagerReady(ns, name string, replicas int) error {
	err := WaitForPodsReady(
		f.KubeClient,
		ns,
		5*time.Minute,
		replicas,
		alertmanager.ListOptions(name),
	)

	return errors.Wrap(err, fmt.Sprintf("failed to create an Alertmanager cluster (%s) with %d instances", name, replicas))
}

func (f *Framework) UpdateAlertmanagerAndWaitUntilReady(ns string, a *monitoringv1.Alertmanager) error {
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
	_, err := f.MonClient.Alertmanagers(ns).Get(name, metav1.GetOptions{})
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
	return wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		amStatus, err := f.GetAlertmanagerConfig(ns, name)
		if err != nil {
			return false, err
		}
		if amStatus.Data.getAmountPeers() == amountPeers {
			return true, nil
		}

		return false, nil
	})
}

func (f *Framework) GetAlertmanagerConfig(ns, n string) (amAPIStatusResp, error) {
	var amStatus amAPIStatusResp
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

func (f *Framework) CreateSilence(ns, n string) (string, error) {
	var createSilenceResponse amAPICreateSilResp

	request := ProxyPostPod(
		f.KubeClient, ns, n,
		"9093", "/api/v1/silences",
		"{\"id\":\"\",\"createdBy\":\"Max Mustermann\",\"comment\":\"1234\",\"startsAt\":\"2030-04-09T09:16:15.114Z\",\"endsAt\":\"2031-04-09T11:16:15.114Z\",\"matchers\":[{\"name\":\"test\",\"value\":\"123\",\"isRegex\":false}]}",
	)
	resp, err := request.DoRaw()
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(resp, &createSilenceResponse); err != nil {
		return "", err
	}

	if createSilenceResponse.Status != "success" {
		return "", errors.Errorf(
			"expected Alertmanager to return 'success', but got '%v' instead",
			createSilenceResponse.Status,
		)
	}

	return createSilenceResponse.Data.SilenceID, nil
}

func (f *Framework) GetSilences(ns, n string) ([]amAPISil, error) {
	var getSilencesResponse amAPIGetSilResp

	request := ProxyGetPod(f.KubeClient, ns, n, "9093", "/api/v1/silences")
	resp, err := request.DoRaw()
	if err != nil {
		return getSilencesResponse.Data, err
	}

	if err := json.Unmarshal(resp, &getSilencesResponse); err != nil {
		return getSilencesResponse.Data, err
	}

	if getSilencesResponse.Status != "success" {
		return getSilencesResponse.Data, errors.Errorf(
			"expected Alertmanager to return 'success', but got '%v' instead",
			getSilencesResponse.Status,
		)
	}

	return getSilencesResponse.Data, nil
}

func (f *Framework) WaitForSpecificAlertmanagerConfig(ns, amName string, expectedConfig string) error {
	return wait.Poll(10*time.Second, time.Minute*5, func() (bool, error) {
		config, err := f.GetAlertmanagerConfig(ns, "alertmanager-"+amName+"-0")
		if err != nil {
			return false, err
		}

		if config.Data.ConfigYAML == expectedConfig {
			return true, nil
		}

		log.Printf("\n\nFound:\n\n%#+v\n\nExpected:\n\n%#+v\n\n", config.Data.ConfigYAML, expectedConfig)

		return false, nil
	})
}

type amAPICreateSilResp struct {
	Status string             `json:"status"`
	Data   amAPICreateSilData `json:"data"`
}

type amAPICreateSilData struct {
	SilenceID string `json:"silenceId"`
}

type amAPIGetSilResp struct {
	Status string     `json:"status"`
	Data   []amAPISil `json:"data"`
}

type amAPISil struct {
	ID        string `json:"id"`
	CreatedBy string `json:"createdBy"`
}

type amAPIStatusResp struct {
	Data amAPIStatusData `json:"data"`
}

type amAPIStatusData struct {
	ClusterStatus *clusterStatus `json:"clusterStatus,omitempty"`
	MeshStatus    *clusterStatus `json:"meshStatus,omitempty"`
	ConfigYAML    string         `json:"configYAML"`
}

// Starting from AM v0.15.0 'MeshStatus' is called 'ClusterStatus'
func (s *amAPIStatusData) getAmountPeers() int {
	if s.MeshStatus != nil {
		return len(s.MeshStatus.Peers)
	} else {
		return len(s.ClusterStatus.Peers)
	}
}

type clusterStatus struct {
	Peers []interface{} `json:"peers"`
}
