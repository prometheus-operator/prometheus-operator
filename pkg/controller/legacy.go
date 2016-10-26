package controller

import (
	apiV1 "k8s.io/client-go/1.5/pkg/api/v1"
	apiExtensions "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
)

func makeReplicaSet(name string, replicas int32) *apiExtensions.ReplicaSet {
	rs := &apiExtensions.ReplicaSet{
		ObjectMeta: apiV1.ObjectMeta{
			Name: name,
		},
		Spec: apiExtensions.ReplicaSetSpec{
			Replicas: &replicas,
			Template: apiV1.PodTemplateSpec{
				ObjectMeta: apiV1.ObjectMeta{
					Labels: map[string]string{
						"prometheus.coreos.com/name": name,
						"prometheus.coreos.com/type": "prometheus",
					},
				},
				Spec: apiV1.PodSpec{
					Containers: []apiV1.Container{
						{
							Name:  "prometheus",
							Image: "quay.io/prometheus/prometheus:v1.3.0-beta.0",
							Ports: []apiV1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9090,
									Protocol:      apiV1.ProtocolTCP,
								},
							},
							Args: []string{
								"-storage.local.retention=12h",
								"-storage.local.memory-chunks=500000",
								"-config.file=/etc/prometheus/prometheus.yaml",
							},
							VolumeMounts: []apiV1.VolumeMount{
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
							VolumeMounts: []apiV1.VolumeMount{
								{
									Name:      "config-volume",
									ReadOnly:  true,
									MountPath: "/etc/prometheus",
								},
							},
						},
					},
					Volumes: []apiV1.Volume{
						{
							Name: "config-volume",
							VolumeSource: apiV1.VolumeSource{
								ConfigMap: &apiV1.ConfigMapVolumeSource{
									LocalObjectReference: apiV1.LocalObjectReference{
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
	return rs
}
