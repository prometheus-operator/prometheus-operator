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

package alertmanager

import (
	"fmt"
	"net/url"
	"path"

	"github.com/coreos/prometheus-operator/pkg/spec"
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/apps/v1alpha1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

func makePetSet(am *spec.Alertmanager, old *v1alpha1.PetSet) *v1alpha1.PetSet {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	baseImage := am.Spec.BaseImage
	if baseImage == "" {
		baseImage = "quay.io/prometheus/alertmanager"
	}
	version := am.Spec.Version
	if version == "" {
		version = "v0.5.1"
	}
	replicas := am.Spec.Replicas
	if replicas < 1 {
		replicas = 1
	}
	image := fmt.Sprintf("%s:%s", baseImage, version)

	petset := &v1alpha1.PetSet{
		ObjectMeta: v1.ObjectMeta{
			Name: am.Name,
		},
		Spec: makePetSetSpec(am.Namespace, am.Name, image, version, am.Spec.ExternalURL, replicas),
	}
	if vc := am.Spec.Storage; vc == nil {
		petset.Spec.Template.Spec.Volumes = append(petset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: fmt.Sprintf("%s-db", am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvc := v1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name: fmt.Sprintf("%s-db", am.Name),
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources:   vc.Resources,
				Selector:    vc.Selector,
			},
		}
		if len(vc.Class) > 0 {
			pvc.ObjectMeta.Annotations = map[string]string{
				"volume.beta.kubernetes.io/storage-class": vc.Class,
			}
		}
		petset.Spec.VolumeClaimTemplates = append(petset.Spec.VolumeClaimTemplates, pvc)
	}

	if old != nil {
		petset.Annotations = old.Annotations
	}
	return petset
}

func makePetSetService(p *spec.Alertmanager) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: "alertmanager",
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9093,
					TargetPort: intstr.FromInt(9093),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "mesh",
					Port:       6783,
					TargetPort: intstr.FromInt(6783),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "alertmanager",
			},
		},
	}
	return svc
}

func makePetSetSpec(ns, name, image, version, externalURL string, replicas int32) v1alpha1.PetSetSpec {
	commands := []string{
		"/bin/alertmanager",
		fmt.Sprintf("-config.file=%s", "/etc/alertmanager/config/alertmanager.yaml"),
		fmt.Sprintf("-web.listen-address=:%d", 9093),
		fmt.Sprintf("-mesh.listen-address=:%d", 6783),
		fmt.Sprintf("-storage.path=%s", "/etc/alertmanager/data"),
	}

	for i := int32(0); i < replicas; i++ {
		commands = append(commands, fmt.Sprintf("-mesh.peer=%s-%d.%s.%s.svc", name, i, "alertmanager", ns))
	}

	webRoutePrefix := ""
	if externalURL != "" {
		commands = append(commands, "-web.external-url="+externalURL)
		extUrl, err := url.Parse(externalURL)
		if err == nil {
			webRoutePrefix = extUrl.Path
		}
	}

	localReloadURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:9093",
		Path:   path.Clean(webRoutePrefix + "/-/reload"),
	}

	if externalURL != "" {
		commands = append(commands, "-web.external-url="+externalURL)
	}

	terminationGracePeriod := int64(0)
	return v1alpha1.PetSetSpec{
		ServiceName: "alertmanager",
		Replicas:    &replicas,
		Template: v1.PodTemplateSpec{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app":          "alertmanager",
					"alertmanager": name,
				},
				Annotations: map[string]string{
					"pod.alpha.kubernetes.io/initialized": "true",
				},
			},
			Spec: v1.PodSpec{
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Containers: []v1.Container{
					{
						Command: commands,
						Name:    name,
						Image:   image,
						Ports: []v1.ContainerPort{
							{
								Name:          "web",
								ContainerPort: 9093,
								Protocol:      v1.ProtocolTCP,
							},
							{
								Name:          "mesh",
								ContainerPort: 6783,
								Protocol:      v1.ProtocolTCP,
							},
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: "/etc/alertmanager/config",
							},
							{
								Name:      fmt.Sprintf("%s-db", name),
								MountPath: "/var/alertmanager/data",
								SubPath:   "alertmanager-db",
							},
						},
					}, {
						Name:  "config-reloader",
						Image: "jimmidyson/configmap-reload",
						Args: []string{
							fmt.Sprintf("-webhook-url=%s", localReloadURL),
							"-volume-dir=/etc/alertmanager/config",
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config-volume",
								ReadOnly:  true,
								MountPath: "/etc/alertmanager/config",
							},
						},
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("5m"),
								v1.ResourceMemory: resource.MustParse("10Mi"),
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
	}
}
