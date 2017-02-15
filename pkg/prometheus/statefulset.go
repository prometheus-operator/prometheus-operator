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

package prometheus

import (
	"fmt"
	"net/url"
	"path"

	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

const (
	governingServiceName = "prometheus-operated"
	defaultBaseImage     = "quay.io/prometheus/prometheus"
	defaultVersion       = "v1.5.2"
	minReplicas          = 1
	defaultRetention     = "24h"
)

func makeStatefulSet(p v1alpha1.Prometheus, old *v1beta1.StatefulSet, config *Config) *v1beta1.StatefulSet {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if p.Spec.BaseImage == "" {
		p.Spec.BaseImage = defaultBaseImage
	}
	if p.Spec.Version == "" {
		p.Spec.Version = defaultVersion
	}
	if p.Spec.Replicas < minReplicas {
		p.Spec.Replicas = minReplicas
	}
	if p.Spec.Retention == "" {
		p.Spec.Retention = defaultRetention
	}

	if p.Spec.Resources.Requests == nil {
		p.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := p.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		p.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("2Gi")
	}

	statefulset := &v1beta1.StatefulSet{
		ObjectMeta: apimetav1.ObjectMeta{
			Name:        prefixedName(p.Name),
			Labels:      p.ObjectMeta.Labels,
			Annotations: p.ObjectMeta.Annotations,
		},
		Spec: makeStatefulSetSpec(p, config),
	}
	if vc := p.Spec.Storage; vc == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(p.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvc := v1.PersistentVolumeClaim{
			ObjectMeta: apimetav1.ObjectMeta{
				Name: volumeName(p.Name),
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
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvc)
	}

	if old != nil {
		statefulset.Annotations = old.Annotations
	}
	return statefulset
}

func makeEmptyConfig(name string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: apimetav1.ObjectMeta{
			Name: configConfigMapName(name),
		},
		Data: map[string]string{
			"prometheus.yaml": "",
		},
	}
}

func makeEmptyRules(name string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: apimetav1.ObjectMeta{
			Name: rulesConfigMapName(name),
		},
	}
}

func makeStatefulSetService(p *v1alpha1.Prometheus) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: apimetav1.ObjectMeta{
			Name: governingServiceName,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"app": "prometheus",
			},
		},
	}
	return svc
}

func makeStatefulSetSpec(p v1alpha1.Prometheus, c *Config) v1beta1.StatefulSetSpec {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	// We attempt to specify decent storage tuning flags based on how much the
	// requested memory can fit. The user has to specify an appropriate buffering
	// in memory limits to catch increased memory usage during query bursts.
	// More info: https://prometheus.io/docs/operating/storage/.
	reqMem := p.Spec.Resources.Requests[v1.ResourceMemory]
	// 1024 byte is the fixed chunk size. With increasing number of chunks actually
	// in memory, overhead owed to their management, higher ingestion buffers, etc.
	// increases.
	// We are conservative for now an assume this to be 80% as the Kubernetes environment
	// generally has a very high time series churn.
	memChunks := reqMem.Value() / 1024 / 5

	promArgs := []string{
		"-storage.local.retention=" + p.Spec.Retention,
		"-storage.local.memory-chunks=" + fmt.Sprintf("%d", memChunks),
		"-storage.local.max-chunks-to-persist=" + fmt.Sprintf("%d", memChunks/2),
		"-storage.local.num-fingerprint-mutexes=4096",
		"-storage.local.path=/var/prometheus/data",
		"-config.file=/etc/prometheus/config/prometheus.yaml",
	}

	webRoutePrefix := ""
	if p.Spec.ExternalURL != "" {
		promArgs = append(promArgs, "-web.external-url="+p.Spec.ExternalURL)
		extUrl, err := url.Parse(p.Spec.ExternalURL)
		if err == nil {
			webRoutePrefix = extUrl.Path
		}
	}

	if p.Spec.RoutePrefix != "" {
		promArgs = append(promArgs, "-web.route-prefix="+p.Spec.RoutePrefix)
		webRoutePrefix = p.Spec.RoutePrefix
	}

	localReloadURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:9090",
		Path:   path.Clean(webRoutePrefix + "/-/reload"),
	}

	return v1beta1.StatefulSetSpec{
		ServiceName: "prometheus",
		Replicas:    &p.Spec.Replicas,
		Template: v1.PodTemplateSpec{
			ObjectMeta: apimetav1.ObjectMeta{
				Labels: map[string]string{
					"app":        "prometheus",
					"prometheus": p.Name,
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "prometheus",
						Image: fmt.Sprintf("%s:%s", p.Spec.BaseImage, p.Spec.Version),
						Ports: []v1.ContainerPort{
							{
								Name:          "web",
								ContainerPort: 9090,
								Protocol:      v1.ProtocolTCP,
							},
						},
						Args: promArgs,
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config",
								ReadOnly:  true,
								MountPath: "/etc/prometheus/config",
							},
							{
								Name:      "rules",
								ReadOnly:  true,
								MountPath: "/etc/prometheus/rules",
							},
							{
								Name:      volumeName(p.Name),
								MountPath: "/var/prometheus/data",
								SubPath:   subPathForStorage(p.Spec.Storage),
							},
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: path.Clean(webRoutePrefix + "/status"),
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
						Resources: p.Spec.Resources,
					}, {
						Name:  "config-reloader",
						Image: c.ConfigReloaderImage,
						Args: []string{
							fmt.Sprintf("-webhook-url=%s", localReloadURL),
							"-volume-dir=/etc/prometheus/config",
							"-volume-dir=/etc/prometheus/rules/",
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config",
								ReadOnly:  true,
								MountPath: "/etc/prometheus/config",
							},
							{
								Name:      "rules",
								ReadOnly:  true,
								MountPath: "/etc/prometheus/rules",
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
				NodeSelector:                  p.Spec.NodeSelector,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes: []v1.Volume{
					{
						Name: "config",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: configConfigMapName(p.Name),
								},
							},
						},
					},
					{
						Name: "rules",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: rulesConfigMapName(p.Name),
								},
							},
						},
					},
				},
			},
		},
	}
}

func configConfigMapName(name string) string {
	return prefixedName(name)
}

func rulesConfigMapName(name string) string {
	return fmt.Sprintf("%s-rules", prefixedName(name))
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("prometheus-%s", name)
}

func subPathForStorage(s *v1alpha1.StorageSpec) string {
	if s == nil {
		return ""
	}

	return "prometheus-db"
}
