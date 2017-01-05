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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/yaml"

	"github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/coreos/prometheus-operator/pkg/spec"
)

func (f *Framework) CreatePrometheus(e *spec.Prometheus) (*spec.Prometheus, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	resp, err := f.HTTPClient.Post(
		fmt.Sprintf("%s/apis/monitoring.coreos.com/v1alpha1/namespaces/%s/prometheuses", f.MasterHost, f.Namespace.Name),
		"application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}
	decoder := yaml.NewYAMLOrJSONDecoder(resp.Body, 100)
	res := &spec.Prometheus{}
	if err := decoder.Decode(res); err != nil {
		return nil, err
	}
	return res, nil
}

func (f *Framework) DeletePrometheus(name string) error {
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/apis/monitoring.coreos.com/v1alpha1/namespaces/%s/prometheuses/%s", f.MasterHost, f.Namespace.Name, name), nil)
	if err != nil {
		return err
	}
	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %v", resp.Status)
	}
	return nil
}

func (f *Framework) MakeBasicPrometheus(name string, replicas int32) *spec.Prometheus {
	return &spec.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: spec.PrometheusSpec{
			Replicas: replicas,
		},
	}
}

func (f *Framework) CreatePrometheusAndWaitUntilReady(p *spec.Prometheus) error {
	_, err := f.CreatePrometheus(p)
	if err != nil {
		return err
	}

	_, err = f.WaitForPodsReady(time.Minute*2, int(p.Spec.Replicas), prometheus.ListOptions(p.Name))
	if err != nil {
		return fmt.Errorf("failed to create %d Prometheus instances (%s): %v", p.Spec.Replicas, p.Name, err)
	}

	return nil
}

func (f *Framework) DeletePrometheusAndWaitUntilGone(name string) error {
	if err := f.DeletePrometheus(name); err != nil {
		return err
	}

	if _, err := f.WaitForPodsReady(time.Minute*2, 0, prometheus.ListOptions(name)); err != nil {
		return fmt.Errorf("failed to teardown Prometheus instances (%s): %v", name, err)
	}

	return nil
}
