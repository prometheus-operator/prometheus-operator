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
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/blang/semver"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	governingServiceName = "prometheus-operated"
	DefaultVersion       = "v1.7.1"
	defaultRetention     = "24h"

	configMapsFilename = "configmaps.json"
)

var (
	minReplicas                 int32 = 1
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
	probeTimeoutSeconds int32 = 3

	CompatibilityMatrix = []string{
		"v1.4.0",
		"v1.4.1",
		"v1.5.0",
		"v1.5.1",
		"v1.5.2",
		"v1.5.3",
		"v1.6.0",
		"v1.6.1",
		"v1.6.2",
		"v1.6.3",
		"v1.7.0",
		"v1.7.1",
		"v2.0.0-beta.0",
	}
)

func makeStatefulSet(p monitoringv1.Prometheus, old *v1beta1.StatefulSet, config *Config, ruleConfigMaps []*v1.ConfigMap) (*v1beta1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if p.Spec.BaseImage == "" {
		p.Spec.BaseImage = config.PrometheusDefaultBaseImage
	}
	if p.Spec.Version == "" {
		p.Spec.Version = DefaultVersion
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
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(p.Name),
			Labels:      p.ObjectMeta.Labels,
			Annotations: p.ObjectMeta.Annotations,
		},
		Spec: *spec,
	}

	if p.Spec.ImagePullSecrets != nil && len(p.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = p.Spec.ImagePullSecrets
	}
	storageSpec := p.Spec.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(p.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvcTemplate := storageSpec.VolumeClaimTemplate
		pvcTemplate.Name = volumeName(p.Name)
		pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvcTemplate)
	}

	if old != nil {
		statefulset.Annotations = old.Annotations
	}

	if !config.StatefulSetUpdatesAvailable {
		statefulset.Spec.UpdateStrategy = v1beta1.StatefulSetUpdateStrategy{}

		if old != nil {
			// Mounted volumes are not reconciled as StatefulSets do not allow
			// modification of the PodTemplate.
			//
			// TODO(brancz): remove this when dropping 1.6 compatibility.
			statefulset.Spec.Template.Spec.Containers[0].VolumeMounts = old.Spec.Template.Spec.Containers[0].VolumeMounts
			statefulset.Spec.Template.Spec.Volumes = old.Spec.Template.Spec.Volumes
		}
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

func (l *ConfigMapReferenceList) Len() int {
	return len(l.Items)
}

func (l *ConfigMapReferenceList) Less(i, j int) bool {
	return l.Items[i].Key < l.Items[j].Key
}

func (l *ConfigMapReferenceList) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

func makeRuleConfigMap(cm *v1.ConfigMap) (*ConfigMapReference, error) {
	keys := []string{}
	for k, _ := range cm.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	m := yaml.MapSlice{}
	for _, k := range keys {
		m = append(m, yaml.MapItem{Key: k, Value: cm.Data[k]})
	}

	b, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &ConfigMapReference{
		Key:      cm.Namespace + "/" + cm.Name,
		Checksum: fmt.Sprintf("%x", sha256.Sum256(b)),
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

	sort.Sort(cml)
	return json.Marshal(cml)
}

func makeConfigSecret(name string, configMaps []*v1.ConfigMap) (*v1.Secret, error) {
	b, err := makeRuleConfigMapListFile(configMaps)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   configSecretName(name),
			Labels: managedByOperatorLabels,
		},
		Data: map[string][]byte{
			configFilename:     []byte{},
			configMapsFilename: b,
		},
	}, nil
}

func makeStatefulSetService(p *monitoringv1.Prometheus) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
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

func makeStatefulSetSpec(p monitoringv1.Prometheus, c *Config, ruleConfigMaps []*v1.ConfigMap) (*v1beta1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	versionStr := strings.TrimLeft(p.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	var promArgs []string
	var securityContext v1.PodSecurityContext

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

		securityContext = v1.PodSecurityContext{}
	case 2:

		// Prometheus 2.0 is in alpha and is highly experimental, and therefore
		// flags and other things may change for the final release of 2.0. This
		// section is also regarded as experimental until a Prometheus 2.0 stable
		// has been released. These flags will be updated to work with every new
		// 2.0 release until a stable release. These flags are taregeted at version
		// v2.0.0-alpha.3, there is no guarantee that these flags will continue to
		// work for any further version, this feature is experimental and developed
		// on a best effort basis.

		promArgs = append(promArgs,
			"-config.file=/etc/prometheus/config/prometheus.yaml",
			"-storage.tsdb.path=/var/prometheus/data",
			"-storage.tsdb.retention="+p.Spec.Retention,
			"-web.enable-lifecycle",
		)

		gid := int64(2000)
		uid := int64(1000)
		nr := true
		securityContext = v1.PodSecurityContext{
			FSGroup: &gid,
			RunAsNonRoot: &nr,
			RunAsUser: &uid,
		}
	default:
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}

	if p.Spec.ExternalURL != "" {
		promArgs = append(promArgs, "-web.external-url="+p.Spec.ExternalURL)
	}

	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}
	promArgs = append(promArgs, "-web.route-prefix="+webRoutePrefix)

	if version.Major == 2 {
		for i, a := range promArgs {
			promArgs[i] = "-" + a
		}
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

	probeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/status"),
			Port: intstr.FromString("web"),
		},
	}

	return &v1beta1.StatefulSetSpec{
		ServiceName: governingServiceName,
		Replicas:    p.Spec.Replicas,
		UpdateStrategy: v1beta1.StatefulSetUpdateStrategy{
			Type: v1beta1.RollingUpdateStatefulSetStrategyType,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
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
						LivenessProbe: &v1.Probe{
							Handler: probeHandler,
							// For larger servers, restoring a checkpoint on startup may take quite a bit of time.
							// Wait up to 5 minutes.
							InitialDelaySeconds: 300,
							PeriodSeconds:       5,
							TimeoutSeconds:      probeTimeoutSeconds,
							FailureThreshold:    10,
						},
						ReadinessProbe: &v1.Probe{
							Handler:          probeHandler,
							TimeoutSeconds:   probeTimeoutSeconds,
							PeriodSeconds:    5,
							FailureThreshold: 6,
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
				SecurityContext:               &securityContext,
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

func subPathForStorage(s *monitoringv1.StorageSpec) string {
	if s == nil {
		return ""
	}

	return "prometheus-db"
}
