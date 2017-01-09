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

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
)

func (f *Framework) MakeBasicPrometheus(name string, replicas int32) *v1alpha1.Prometheus {
	return &v1alpha1.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.PrometheusSpec{
			Replicas: replicas,
		},
	}
}

func (f *Framework) CreatePrometheusAndWaitUntilReady(p *v1alpha1.Prometheus) error {
	_, err := f.MonClient.Prometheuses(f.Namespace.Name).Create(p)
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
	if err := f.MonClient.Prometheuses(f.Namespace.Name).Delete(name, nil); err != nil {
		return err
	}

	if _, err := f.WaitForPodsReady(time.Minute*2, 0, prometheus.ListOptions(name)); err != nil {
		return fmt.Errorf("failed to teardown Prometheus instances (%s): %v", name, err)
	}

	return nil
}
