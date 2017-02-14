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
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
)

func (f *Framework) MakeBasicPrometheus(name, group string, replicas int32) *v1alpha1.Prometheus {
	return &v1alpha1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.PrometheusSpec{
			Replicas: replicas,
			Version:  "v1.4.0",
			ServiceMonitorSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": group,
				},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("400Mi"),
				},
			},
		},
	}
}

func (f *Framework) AddAlertingToPrometheus(p *v1alpha1.Prometheus, name string) {
	p.Spec.Alerting = v1alpha1.AlertingSpec{
		Alertmanagers: []v1alpha1.AlertmanagerEndpoints{
			v1alpha1.AlertmanagerEndpoints{
				Namespace: f.Namespace.Name,
				Name:      fmt.Sprintf("alertmanager-%s", name),
				Port:      intstr.FromString("web"),
			},
		},
	}
}

func (f *Framework) MakeBasicServiceMonitor(name string) *v1alpha1.ServiceMonitor {
	return &v1alpha1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1alpha1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": name,
				},
			},
			Endpoints: []v1alpha1.Endpoint{
				v1alpha1.Endpoint{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}
}

func (f *Framework) MakeBasicPrometheusNodePortService(name, group string, nodePort int32) *v1.Service {
	pService := f.MakePrometheusService(name, group, v1.ServiceTypeNodePort)
	pService.Spec.Ports[0].NodePort = nodePort
	return pService
}

func (f *Framework) MakePrometheusService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"prometheus": name,
			},
		},
	}
	return service
}

func (f *Framework) CreatePrometheusAndWaitUntilReady(p *v1alpha1.Prometheus) error {
	log.Printf("Creating Prometheus (%s/%s)", f.Namespace.Name, p.Name)
	_, err := f.MonClient.Prometheuses(f.Namespace.Name).Create(p)
	if err != nil {
		return err
	}

	_, err = f.WaitForPodsReady(time.Minute*6, int(p.Spec.Replicas), promImage(p.Spec.Version), prometheus.ListOptions(p.Name))
	if err != nil {
		return fmt.Errorf("failed to create %d Prometheus instances (%s): %v", p.Spec.Replicas, p.Name, err)
	}

	return nil
}

func (f *Framework) UpdatePrometheusAndWaitUntilReady(p *v1alpha1.Prometheus) error {
	log.Printf("Updating Prometheus (%s/%s)", f.Namespace.Name, p.Name)
	_, err := f.MonClient.Prometheuses(f.Namespace.Name).Update(p)
	if err != nil {
		return err
	}

	_, err = f.WaitForPodsReady(time.Minute*6, int(p.Spec.Replicas), promImage(p.Spec.Version), prometheus.ListOptions(p.Name))
	if err != nil {
		return fmt.Errorf("failed to update %d Prometheus instances (%s): %v", p.Spec.Replicas, p.Name, err)
	}

	return nil
}

func (f *Framework) DeletePrometheusAndWaitUntilGone(name string) error {
	log.Printf("Deleting Prometheus (%s/%s)", f.Namespace.Name, name)
	p, err := f.MonClient.Prometheuses(f.Namespace.Name).Get(name)
	if err != nil {
		return err
	}

	if err := f.MonClient.Prometheuses(f.Namespace.Name).Delete(name, nil); err != nil {
		return err
	}

	if _, err := f.WaitForPodsReady(time.Minute*6, 0, promImage(p.Spec.Version), prometheus.ListOptions(name)); err != nil {
		return fmt.Errorf("failed to teardown Prometheus instances (%s): %v", name, err)
	}

	return nil
}

func promImage(version string) string {
	return fmt.Sprintf("quay.io/prometheus/prometheus:%s", version)
}
