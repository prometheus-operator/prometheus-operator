// Copyright The prometheus-operator Authors
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
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	// AppGroupLabel is the value of the "group" label for the instrumented sample
	// app resources.
	AppGroupLabel     = "app"
	appSecretAuthName = "auth"
)

// DeployBasicAuthApp deploys the instrumented sample app with the given number
// of replicas, a service named "app" exposing port 8080, and a basic auth
// secret named "auth" with credentials user/pass.
func (f *Framework) DeployBasicAuthApp(ctx context.Context, ns string, replicas int32) error {
	dep, err := MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		return err
	}
	dep.Spec.Replicas = new(replicas)

	if err := f.CreateDeployment(ctx, ns, dep); err != nil {
		return err
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: AppGroupLabel,
			Labels: map[string]string{
				"group": AppGroupLabel,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: dep.Spec.Template.Labels,
			Ports: []corev1.ServicePort{
				{
					Name: "web",
					Port: 8080,
				},
			},
		},
	}
	if _, err = f.KubeClient.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSecretAuthName,
			Namespace: ns,
		},
		StringData: map[string]string{
			"user": "user",
			"pass": "pass",
		},
		Type: corev1.SecretTypeOpaque,
	}
	_, err = f.KubeClient.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// DeployAppServiceMonitor creates a ServiceMonitor for the app deployed by DeployBasicAuthApp.
func (f *Framework) DeployAppServiceMonitor(ctx context.Context, ns string) error {
	sm := f.MakeBasicServiceMonitor(AppGroupLabel)
	sm.Spec.Endpoints[0] = monitoringv1.Endpoint{
		Interval: monitoringv1.Duration("5s"),
		Port:     "web",
		HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
			HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
				HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
					BasicAuth: &monitoringv1.BasicAuth{
						Username: corev1.SecretKeySelector{
							Key:                  "user",
							LocalObjectReference: corev1.LocalObjectReference{Name: appSecretAuthName},
						},
						Password: corev1.SecretKeySelector{
							Key:                  "pass",
							LocalObjectReference: corev1.LocalObjectReference{Name: appSecretAuthName},
						},
					},
				},
			},
		},
	}
	_, err := f.MonClientV1.ServiceMonitors(ns).Create(ctx, sm, metav1.CreateOptions{})
	return err
}
