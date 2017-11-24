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
	"strings"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/blang/semver"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/pkg/errors"
)

const (
	governingServiceName   = "alertmanager-operated"
	defaultVersion         = "v0.9.1"
	alertmanagerConfDir    = "/etc/alertmanager/config"
	alertmanagerConfFile   = alertmanagerConfDir + "/alertmanager.yaml"
	alertmanagerStorageDir = "/var/alertmanager/data"
)

var (
	minReplicas         int32 = 1
	probeTimeoutSeconds int32 = 3
)

func makeStatefulSet(am *monitoringv1.Alertmanager, old *v1beta1.StatefulSet, config Config) (*v1beta1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if am.Spec.BaseImage == "" {
		am.Spec.BaseImage = config.AlertmanagerDefaultBaseImage
	}
	if am.Spec.Version == "" {
		am.Spec.Version = defaultVersion
	}
	if am.Spec.Replicas == nil {
		am.Spec.Replicas = &minReplicas
	}
	intZero := int32(0)
	if am.Spec.Replicas != nil && *am.Spec.Replicas < 0 {
		am.Spec.Replicas = &intZero
	}
	if am.Spec.Resources.Requests == nil {
		am.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := am.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		am.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("200Mi")
	}

	spec, err := makeStatefulSetSpec(am, config)
	if err != nil {
		return nil, err
	}

	boolTrue := true
	statefulset := &v1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(am.Name),
			Labels:      config.Labels.Merge(am.ObjectMeta.Labels),
			Annotations: am.ObjectMeta.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         am.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               am.Kind,
					Name:               am.Name,
					UID:                am.UID,
				},
			},
		},
		Spec: *spec,
	}

	if am.Spec.ImagePullSecrets != nil && len(am.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = am.Spec.ImagePullSecrets
	}

	storageSpec := am.Spec.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvcTemplate := storageSpec.VolumeClaimTemplate
		pvcTemplate.Name = volumeName(am.Name)
		pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvcTemplate)
	}

	if old != nil {
		statefulset.Annotations = old.Annotations
	}

	return statefulset, nil
}

func makeStatefulSetService(p *monitoringv1.Alertmanager, config Config) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			Labels: config.Labels.Merge(map[string]string{
				"operated-alertmanager": "true",
			}),
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

func makeStatefulSetSpec(a *monitoringv1.Alertmanager, config Config) (*v1beta1.StatefulSetSpec, error) {
	image := fmt.Sprintf("%s:%s", a.Spec.BaseImage, a.Spec.Version)
	versionStr := strings.TrimLeft(a.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	amArgs := []string{
		fmt.Sprintf("-config.file=%s", alertmanagerConfFile),
		fmt.Sprintf("-web.listen-address=:%d", 9093),
		fmt.Sprintf("-mesh.listen-address=:%d", 6783),
		fmt.Sprintf("-storage.path=%s", alertmanagerStorageDir),
	}

	if a.Spec.ExternalURL != "" {
		amArgs = append(amArgs, "-web.external-url="+a.Spec.ExternalURL)
	}

	webRoutePrefix := "/"
	if a.Spec.RoutePrefix != "" {
		webRoutePrefix = a.Spec.RoutePrefix
	}

	switch version.Major {
	case 0:
		if version.Minor >= 7 {
			amArgs = append(amArgs, "-web.route-prefix="+webRoutePrefix)
		}
	default:
		return nil, errors.Errorf("unsupported Alertmanager major version %s", version)
	}

	localReloadURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:9093",
		Path:   path.Clean(webRoutePrefix + "/-/reload"),
	}

	probeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/api/v1/status"),
			Port: intstr.FromString("web"),
		},
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	if a.Spec.PodMetadata != nil {
		if a.Spec.PodMetadata.Labels != nil {
			for k, v := range a.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if a.Spec.PodMetadata.Annotations != nil {
			for k, v := range a.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}
	podLabels["app"] = "alertmanager"
	podLabels["alertmanager"] = a.Name

	for i := int32(0); i < *a.Spec.Replicas; i++ {
		amArgs = append(amArgs, fmt.Sprintf("-mesh.peer=%s-%d.%s.%s.svc", prefixedName(a.Name), i, governingServiceName, a.Namespace))
	}

	terminationGracePeriod := int64(0)
	return &v1beta1.StatefulSetSpec{
		ServiceName: governingServiceName,
		Replicas:    a.Spec.Replicas,
		UpdateStrategy: v1beta1.StatefulSetUpdateStrategy{
			Type: v1beta1.RollingUpdateStatefulSetStrategyType,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      config.Labels.Merge(podLabels),
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				NodeSelector:                  a.Spec.NodeSelector,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Containers: []v1.Container{
					{
						Args:  amArgs,
						Name:  "alertmanager",
						Image: image,
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
								MountPath: alertmanagerConfDir,
							},
							{
								Name:      volumeName(a.Name),
								MountPath: alertmanagerStorageDir,
								SubPath:   subPathForStorage(a.Spec.Storage),
							},
						},
						LivenessProbe: &v1.Probe{
							Handler:          probeHandler,
							TimeoutSeconds:   probeTimeoutSeconds,
							FailureThreshold: 10,
						},
						ReadinessProbe: &v1.Probe{
							Handler:             probeHandler,
							InitialDelaySeconds: 3,
							TimeoutSeconds:      3,
							PeriodSeconds:       5,
							FailureThreshold:    10,
						},
						Resources: a.Spec.Resources,
					}, {
						Name:  "config-reloader",
						Image: config.ConfigReloaderImage,
						Args: []string{
							fmt.Sprintf("-webhook-url=%s", localReloadURL),
							fmt.Sprintf("-volume-dir=%s", alertmanagerConfDir),
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config-volume",
								ReadOnly:  true,
								MountPath: alertmanagerConfDir,
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
							Secret: &v1.SecretVolumeSource{
								SecretName: configSecretName(a.Name),
							},
						},
					},
				},
				Tolerations: a.Spec.Tolerations,
				Affinity:    a.Spec.Affinity,
			},
		},
	}, nil
}

func configSecretName(name string) string {
	return prefixedName(name)
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("alertmanager-%s", name)
}

func subPathForStorage(s *monitoringv1.StorageSpec) string {
	if s == nil {
		return ""
	}

	return "alertmanager-db"
}
