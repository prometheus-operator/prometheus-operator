// Copyright 2020 The prometheus-operator Authors
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
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *Framework) MakeBlackBoxExporterService(ns, name string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": "blackbox-exporter",
			},
			Ports: []v1.ServicePort{
				{
					Port:       9115,
					TargetPort: intstr.FromInt(9115),
				},
			},
		},
	}
}

func (f *Framework) createBlackBoxExporterConfigMapAndWaitExists(ns, name string) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string]string{
			"blackbox.yml": `modules:
  http_2xx:
    http:
      no_follow_redirects: false
      preferred_ip_protocol: ip4
      valid_http_versions:
      - HTTP/1.1
      - HTTP/2
    prober: http
`,
		},
	}
	ctx := context.TODO()
	if _, err := f.KubeClient.CoreV1().ConfigMaps(ns).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
		return err
	}

	if _, err := f.WaitForConfigMapExist(ns, name); err != nil {
		return err
	}
	return nil
}

func (f *Framework) createBlackBoxExporterDeploymentAndWaitReady(ns, name string, replicas int32) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "blackbox-exporter",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "blackbox-exporter",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "blackbox-exporter",
							Image: "prom/blackbox-exporter:v0.17.0",
							Args: []string{
								"--config.file=/config/blackbox.yml",
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9115,
									Protocol:      v1.ProtocolTCP,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/config",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ctx := context.TODO()
	deploymentInterface := f.KubeClient.AppsV1().Deployments(ns)
	if _, err := deploymentInterface.Create(ctx, deploy, metav1.CreateOptions{}); err != nil {
		return err
	}

	return wait.Poll(2*time.Second, f.DefaultTimeout, func() (bool, error) {
		blackbox, err := deploymentInterface.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if blackbox.Status.ReadyReplicas != *blackbox.Spec.Replicas {
			return false, nil
		}

		return true, nil
	})
}

func (f *Framework) CreateBlackBoxExporterAndWaitUntilReady(ns, name string) error {
	if err := f.createBlackBoxExporterConfigMapAndWaitExists(ns, name); err != nil {
		return err
	}

	return f.createBlackBoxExporterDeploymentAndWaitReady(ns, name, 1)
}

func (f *Framework) MakeBasicStaticProbe(name, url string, targets []string) *monitoringv1.Probe {
	return &monitoringv1.Probe{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: monitoringv1.ProbeSpec{
			Interval: "15s",
			Module:   "http_2xx",
			ProberSpec: monitoringv1.ProberSpec{
				URL: url,
			},
			Targets: monitoringv1.ProbeTargets{
				StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
					Targets: targets,
				},
			},
		},
	}
}
