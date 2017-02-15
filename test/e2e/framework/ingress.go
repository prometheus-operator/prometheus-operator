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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
	"os"
	"time"
)

func (f *Framework) MakeBasicIngress(serviceName string, servicePort int) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "monitoring",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								v1beta1.HTTPIngressPath{
									Backend: v1beta1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: intstr.FromInt(servicePort),
									},
									Path: "/metrics",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (f *Framework) CreateIngress(i *v1beta1.Ingress) error {
	_, err := f.KubeClient.Extensions().Ingresses(f.Namespace.Name).Create(i)
	return err
}

func (f *Framework) SetupNginxIngressControllerIncDefaultBackend() error {
	// Create Nginx Ingress Replication Controller
	if err := createReplicationControllerViaYml("./framework/ressources/nxginx-ingress-controller.yml", f); err != nil {
		return err
	}

	// Create Default HTTP Backend Replication Controller
	if err := createReplicationControllerViaYml("./framework/ressources/default-http-backend.yml", f); err != nil {
		return err
	}

	// Create Default HTTP Backend Service
	manifest, err := os.Open("./framework/ressources/default-http-backend-service.yml")
	if err != nil {
		return err
	}

	service := v1.Service{}
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&service)
	if err != nil {
		return err
	}

	_, err = f.KubeClient.CoreV1().Services(f.Namespace.Name).Create(&service)
	if err != nil {
		return err
	}
	if err := f.WaitForServiceReady(service.Name); err != nil {
		return err
	}

	return nil
}

func (f *Framework) DeleteNginxIngressControllerIncDefaultBackend() error {
	// Delete Nginx Ingress Replication Controller
	if err := deleteReplicationControllerViaYml("./framework/ressources/nxginx-ingress-controller.yml", f); err != nil {
		return err
	}

	// Delete Default HTTP Backend Replication Controller
	if err := deleteReplicationControllerViaYml("./framework/ressources/default-http-backend.yml", f); err != nil {
		return err
	}

	// Delete Default HTTP Backend Service
	manifest, err := os.Open("./framework/ressources/default-http-backend-service.yml")
	if err != nil {
		return err
	}

	service := v1.Service{}
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&service)
	if err != nil {
		return err
	}

	if err := f.KubeClient.CoreV1().Services(f.Namespace.Name).Delete(service.Name, nil); err != nil {
		return err
	}

	return nil
}

func (f *Framework) GetIngressIP(ingressName string) (*string, error) {
	var ingress *v1beta1.Ingress
	err := f.Poll(time.Minute*5, time.Millisecond*500, func() (bool, error) {
		var err error
		ingress, err = f.KubeClient.Extensions().Ingresses(f.Namespace.Name).Get(ingressName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		ingresses := ingress.Status.LoadBalancer.Ingress
		if len(ingresses) != 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &ingress.Status.LoadBalancer.Ingress[0].IP, nil
}
