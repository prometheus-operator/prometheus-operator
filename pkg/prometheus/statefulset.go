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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"

	"strings"

	"github.com/blang/semver"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/pkg/errors"
)

const (
	governingServiceName = "prometheus-operated"
	defaultBaseImage     = "quay.io/prometheus/prometheus"
	defaultVersion       = "v1.6.1"
	defaultRetention     = "24h"
)

var (
	minReplicas             int32 = 1
	managedByOperatorLabels       = map[string]string{
		"managed-by": "prometheus-operator",
	}
)

func makeStatefulSet(p v1alpha1.Prometheus, old *v1beta1.StatefulSet, config *Config, ruleConfigMaps []*v1.ConfigMap) (*v1beta1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if p.Spec.BaseImage == "" {
		p.Spec.BaseImage = defaultBaseImage
	}
	if p.Spec.Version == "" {
		p.Spec.Version = defaultVersion
	}
	if p.Spec.Replicas != nil && *p.Spec.Replicas < minReplicas {
		p.Spec.Replicas = &minReplicas
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

	spec, err := makeStatefulSetSpec(p, config, ruleConfigMaps)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	statefulset := &v1beta1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:        prefixedName(p.Name),
			Labels:      p.ObjectMeta.Labels,
			Annotations: p.ObjectMeta.Annotations,
		},
		Spec: *spec,
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
			ObjectMeta: v1.ObjectMeta{
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

		// mounted volumes are not reconciled as StatefulSets do not allow
		// modification of the PodTemplate.
		// TODO(brancz): remove this once StatefulSets allow modification of the
		// PodTemplate.
		statefulset.Spec.Template.Spec.Containers[0].VolumeMounts = old.Spec.Template.Spec.Containers[0].VolumeMounts
		statefulset.Spec.Template.Spec.Volumes = old.Spec.Template.Spec.Volumes
	}
	return statefulset, nil
}

func makeEmptyConfig(name string, configMaps []*v1.ConfigMap) (*v1.Secret, error) {
	s, err := makeConfigSecret(name, configMaps)
	if err != nil {
		return nil, err
	}

	s.ObjectMeta.Annotations = map[string]string{
		"empty": "true",
	}

	return s, nil
}

type ConfigMapReference struct {
	Key      string `json:"key"`
	Checksum string `json:"checksum"`
}

type ConfigMapReferenceList struct {
	Items []*ConfigMapReference `json:"items"`
}

func makeRuleConfigMap(cm *v1.ConfigMap) (*ConfigMapReference, error) {
	hash := sha256.New()
	err := json.NewEncoder(hash).Encode(cm)
	if err != nil {
		return nil, err
	}

	return &ConfigMapReference{
		Key:      cm.Namespace + "/" + cm.Name,
		Checksum: fmt.Sprintf("%x", hash.Sum(nil)),
	}, nil
}

func makeRuleConfigMapListFile(configMaps []*v1.ConfigMap) ([]byte, error) {
	cml := &ConfigMapReferenceList{}

	for _, cm := range configMaps {
		configmap, err := makeRuleConfigMap(cm)
		if err != nil {
			return nil, err
		}
		cml.Items = append(cml.Items, configmap)
	}

	return json.Marshal(cml)
}

func makeConfigSecret(name string, configMaps []*v1.ConfigMap) (*v1.Secret, error) {
	b, err := makeRuleConfigMapListFile(configMaps)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:   configSecretName(name),
			Labels: managedByOperatorLabels,
		},
		Data: map[string][]byte{
			"prometheus.yaml": []byte{},
			"configmaps.json": b,
		},
	}, nil
}

func makeStatefulSetService(p *v1alpha1.Prometheus) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: governingServiceName,
			Labels: map[string]string{
				"operated-prometheus": "true",
			},
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

func makeStatefulSetSpec(p v1alpha1.Prometheus, c *Config, ruleConfigMaps []*v1.ConfigMap) (*v1beta1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	versionStr := strings.TrimLeft(p.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	var promArgs []string

	switch version.Major {
	case 1:
		promArgs = append(promArgs,
			"-storage.local.retention="+p.Spec.Retention,
			"-storage.local.num-fingerprint-mutexes=4096",
			"-storage.local.path=/var/prometheus/data",
			"-storage.local.chunk-encoding-version=2",
			"-config.file=/etc/prometheus/config/prometheus.yaml",
		)
		// We attempt to specify decent storage tuning flags based on how much the
		// requested memory can fit. The user has to specify an appropriate buffering
		// in memory limits to catch increased memory usage during query bursts.
		// More info: https://prometheus.io/docs/operating/storage/.
		reqMem := p.Spec.Resources.Requests[v1.ResourceMemory]

		if version.Minor < 6 {
			// 1024 byte is the fixed chunk size. With increasing number of chunks actually
			// in memory, overhead owed to their management, higher ingestion buffers, etc.
			// increases.
			// We are conservative for now an assume this to be 80% as the Kubernetes environment
			// generally has a very high time series churn.
			memChunks := reqMem.Value() / 1024 / 5

			promArgs = append(promArgs,
				"-storage.local.memory-chunks="+fmt.Sprintf("%d", memChunks),
				"-storage.local.max-chunks-to-persist="+fmt.Sprintf("%d", memChunks/2),
			)
		} else {
			// Leave 1/3 head room for other overhead.
			promArgs = append(promArgs,
				"-storage.local.target-heap-size="+fmt.Sprintf("%d", reqMem.Value()/3*2),
			)
		}
	default:
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}

	webRoutePrefix := ""

	if p.Spec.ExternalURL != "" {
		extUrl, err := url.Parse(p.Spec.ExternalURL)
		if err != nil {
			return nil, errors.Errorf("invalid external URL %s", p.Spec.ExternalURL)
		}
		webRoutePrefix = extUrl.Path
		promArgs = append(promArgs, "-web.external-url="+p.Spec.ExternalURL)
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

	volumes := []v1.Volume{
		{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: configSecretName(p.Name),
				},
			},
		},
		{
			Name: "rules",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	promVolumeMounts := []v1.VolumeMount{
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
	}

	for _, s := range p.Spec.Secrets {
		volumes = append(volumes, v1.Volume{
			Name: "secret-" + s,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      "secret-" + s,
			ReadOnly:  true,
			MountPath: "/etc/prometheus/secrets/" + s,
		})
	}

	configReloadVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			ReadOnly:  true,
			MountPath: "/etc/prometheus/config",
		},
		{
			Name:      "rules",
			MountPath: "/etc/prometheus/rules",
		},
	}

	configReloadArgs := []string{
		fmt.Sprintf("-reload-url=%s", localReloadURL),
		"-config-volume-dir=/etc/prometheus/config",
		"-rule-volume-dir=/etc/prometheus/rules",
	}

	return &v1beta1.StatefulSetSpec{
		ServiceName: governingServiceName,
		Replicas:    p.Spec.Replicas,
		Template: v1.PodTemplateSpec{
			ObjectMeta: v1.ObjectMeta{
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
						Args:         promArgs,
						VolumeMounts: promVolumeMounts,
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
						Name:         "prometheus-config-reloader",
						Image:        c.PrometheusConfigReloader,
						Args:         configReloadArgs,
						VolumeMounts: configReloadVolumeMounts,
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("5m"),
								v1.ResourceMemory: resource.MustParse("10Mi"),
							},
						},
					},
				},
				ServiceAccountName:            p.Spec.ServiceAccountName,
				NodeSelector:                  p.Spec.NodeSelector,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes: volumes,
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
	return fmt.Sprintf("prometheus-%s", name)
}

func subPathForStorage(s *v1alpha1.StorageSpec) string {
	if s == nil {
		return ""
	}

	return "prometheus-db"
}
