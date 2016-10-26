package controller

import (
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
)

func makeDeployment(name string, replicas int32) *v1beta1.Deployment {
	depl := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"prometheus.coreos.com/name": name,
						"prometheus.coreos.com/type": "prometheus",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "prometheus",
							Image: "quay.io/prometheus/prometheus:v1.3.0-beta.0",
							Ports: []v1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9090,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Args: []string{
								"-storage.local.retention=12h",
								"-storage.local.memory-chunks=500000",
								"-config.file=/etc/prometheus/prometheus.yaml",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config-volume",
									ReadOnly:  true,
									MountPath: "/etc/prometheus",
								},
							},
						}, {
							Name:  "reloader",
							Image: "jimmidyson/configmap-reload",
							Args: []string{
								"-webhook-url=http://localhost:9090/-/reload",
								"-volume-dir=/etc/prometheus/",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config-volume",
									ReadOnly:  true,
									MountPath: "/etc/prometheus",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config-volume",
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
	return depl
}
