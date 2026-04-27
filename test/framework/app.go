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
	"k8s.io/utils/ptr"
)

// DeployBasicAuthApp deploys the instrumented sample app with the given number
// of replicas, a service named "app" exposing port 8080, and a basic auth
// secret named "auth" with credentials user/pass.
func (f *Framework) DeployBasicAuthApp(ctx context.Context, ns string, replicas int32) error {
	dep, err := MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		return err
	}
	dep.Spec.Replicas = ptr.To(replicas)

	if err := f.CreateDeployment(ctx, ns, dep); err != nil {
		return err
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app",
			Labels: map[string]string{
				"group": "app",
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
			Name:      "auth",
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
