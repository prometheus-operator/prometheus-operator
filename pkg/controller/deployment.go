package controller

import (
	"fmt"

	"github.com/coreos/kube-prometheus-controller/pkg/spec"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

func makeDeployment(p *spec.Prometheus, old *v1beta1.Deployment) *v1beta1.Deployment {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	baseImage := p.Spec.BaseImage
	if baseImage == "" {
		baseImage = "quay.io/prometheus/prometheus"
	}
	version := p.Spec.Version
	if version == "" {
		version = "v1.3.0-beta.0"
	}
	replicas := p.Spec.Replicas
	if replicas < 1 {
		replicas = 1
	}
	image := fmt.Sprintf("%s:%s", baseImage, version)

	depl := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: p.Name,
		},
		Spec: makeDeploymentSpec(p.Name, image, replicas),
	}
	if old != nil {
		depl.Annotations = old.Annotations
	}
	return depl
}

func makeDeploymentSpec(name, image string, replicas int32) v1beta1.DeploymentSpec {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	return v1beta1.DeploymentSpec{
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
						Image: image,
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
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/status",
									Port: intstr.FromString("web"),
								},
							},
							InitialDelaySeconds: 1,
							TimeoutSeconds:      3,
							PeriodSeconds:       5,
							// For larger servers, restoring a checkpoint on startup may take quite a bit of time.
							// Wait up to 5 minutes.
							FailureThreshold: 100,
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
				TerminationGracePeriodSeconds: &terminationGracePeriod,
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
	}
}
