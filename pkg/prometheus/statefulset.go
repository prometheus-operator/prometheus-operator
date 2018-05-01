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

	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/blang/semver"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	governingServiceName   = "prometheus-operated"
	DefaultVersion         = "v2.2.1"
	defaultRetention       = "24h"
	storageDir             = "/prometheus"
	confDir                = "/etc/prometheus/config"
	confOutDir             = "/etc/prometheus/config_out"
	rulesDir               = "/etc/prometheus/config_out/rules"
	secretsDir             = "/etc/prometheus/secrets/"
	configFilename         = "prometheus.yaml"
	configEnvsubstFilename = "prometheus.env.yaml"
	ruleConfigmapsFilename = "configmaps.json"
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
		"v1.7.2",
		"v1.8.0",
		"v2.0.0",
		"v2.2.1",
	}
)

func makeStatefulSet(p monitoringv1.Prometheus, old *appsv1.StatefulSet, config *Config, ruleConfigMaps []*v1.ConfigMap) (*appsv1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if p.Spec.BaseImage == "" {
		p.Spec.BaseImage = config.PrometheusDefaultBaseImage
	}
	if p.Spec.Version == "" {
		p.Spec.Version = DefaultVersion
	}

	versionStr := strings.TrimLeft(p.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	if p.Spec.Replicas == nil {
		p.Spec.Replicas = &minReplicas
	}
	intZero := int32(0)
	if p.Spec.Replicas != nil && *p.Spec.Replicas < 0 {
		p.Spec.Replicas = &intZero
	}
	if p.Spec.Retention == "" {
		p.Spec.Retention = defaultRetention
	}

	if p.Spec.Resources.Requests == nil {
		p.Spec.Resources.Requests = v1.ResourceList{}
	}
	_, memoryRequestFound := p.Spec.Resources.Requests[v1.ResourceMemory]
	memoryLimit, memoryLimitFound := p.Spec.Resources.Limits[v1.ResourceMemory]
	if !memoryRequestFound && version.Major == 1 {
		defaultMemoryRequest := resource.MustParse("2Gi")
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		// If limit is given and smaller or equal to 2Gi, then set memory
		// request to the given limit. This is necessary as if limit < request,
		// then a Pod is not schedulable.
		if memoryLimitFound && compareResult <= 0 {
			p.Spec.Resources.Requests[v1.ResourceMemory] = memoryLimit
		} else {
			p.Spec.Resources.Requests[v1.ResourceMemory] = defaultMemoryRequest
		}
	}

	spec, err := makeStatefulSetSpec(p, config, ruleConfigMaps)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	boolTrue := true
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(p.Name),
			Labels:      config.Labels.Merge(p.ObjectMeta.Labels),
			Annotations: p.ObjectMeta.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         p.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               p.Kind,
					Name:               p.Name,
					UID:                p.UID,
				},
			},
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
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(p.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
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

		// Updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden.
		statefulset.Spec.PodManagementPolicy = old.Spec.PodManagementPolicy
	}

	return statefulset, nil
}

func makeEmptyConfig(p *monitoringv1.Prometheus, configMaps []*v1.ConfigMap, config Config) (*v1.Secret, error) {
	s, err := makeConfigSecret(p, configMaps, config)
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

func makeConfigSecret(p *monitoringv1.Prometheus, configMaps []*v1.ConfigMap, config Config) (*v1.Secret, error) {
	b, err := makeRuleConfigMapListFile(configMaps)
	if err != nil {
		return nil, err
	}

	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   configSecretName(p.Name),
			Labels: config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         p.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               p.Kind,
					Name:               p.Name,
					UID:                p.UID,
				},
			},
		},
		Data: map[string][]byte{
			configFilename:         []byte{},
			ruleConfigmapsFilename: b,
		},
	}, nil
}

func makeStatefulSetService(p *monitoringv1.Prometheus, config Config) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			Labels: config.Labels.Merge(map[string]string{
				"operated-prometheus": "true",
			}),
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

func makeStatefulSetSpec(p monitoringv1.Prometheus, c *Config, ruleConfigMaps []*v1.ConfigMap) (*appsv1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	versionStr := strings.TrimLeft(p.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	var promArgs []string
	var securityContext *v1.PodSecurityContext

	switch version.Major {
	case 1:
		promArgs = append(promArgs,
			"-storage.local.retention="+p.Spec.Retention,
			"-storage.local.num-fingerprint-mutexes=4096",
			fmt.Sprintf("-storage.local.path=%s", storageDir),
			"-storage.local.chunk-encoding-version=2",
			fmt.Sprintf("-config.file=%s", path.Join(confOutDir, configEnvsubstFilename)),
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

		securityContext = &v1.PodSecurityContext{}
	case 2:
		promArgs = append(promArgs,
			fmt.Sprintf("-config.file=%s", path.Join(confOutDir, configEnvsubstFilename)),
			fmt.Sprintf("-storage.tsdb.path=%s", storageDir),
			"-storage.tsdb.retention="+p.Spec.Retention,
			"-web.enable-lifecycle",
			"-storage.tsdb.no-lockfile",
		)

		gid := int64(2000)
		uid := int64(1000)
		nr := true
		securityContext = &v1.PodSecurityContext{
			RunAsNonRoot: &nr,
		}
		if !c.DisableAutoUserGroup {
			securityContext.FSGroup = &gid
			securityContext.RunAsUser = &uid
		}
	default:
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}

	if p.Spec.SecurityContext != nil {
		securityContext = p.Spec.SecurityContext
	}

	if p.Spec.ExternalURL != "" {
		promArgs = append(promArgs, "-web.external-url="+p.Spec.ExternalURL)
	}

	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}
	promArgs = append(promArgs, "-web.route-prefix="+webRoutePrefix)

	if p.Spec.LogLevel != "" && p.Spec.LogLevel != "info" {
		promArgs = append(promArgs, fmt.Sprintf("-log.level=%s", p.Spec.LogLevel))
	}

	var ports []v1.ContainerPort
	if p.Spec.ListenLocal {
		promArgs = append(promArgs, "-web.listen-address=127.0.0.1:9090")
	} else {
		ports = []v1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: 9090,
				Protocol:      v1.ProtocolTCP,
			},
		}
	}

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
			Name: "config-out",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	promVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config-out",
			ReadOnly:  true,
			MountPath: confOutDir,
		},
		{
			Name:      volumeName(p.Name),
			MountPath: storageDir,
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
			MountPath: secretsDir + s,
		})
	}

	configReloadVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			MountPath: confDir,
		},
		{
			Name:      "config-out",
			MountPath: confOutDir,
		},
	}

	configReloadArgs := []string{
		fmt.Sprintf("--reload-url=%s", localReloadURL),
		fmt.Sprintf("--config-file=%s", path.Join(confDir, configFilename)),
		fmt.Sprintf("--rule-list-file=%s", path.Join(confDir, ruleConfigmapsFilename)),
		fmt.Sprintf("--config-envsubst-file=%s", path.Join(confOutDir, configEnvsubstFilename)),
		fmt.Sprintf("--rule-dir=%s", rulesDir),
	}

	var livenessProbeHandler v1.Handler
	var readinessProbeHandler v1.Handler
	var livenessFailureThreshold int32
	if (version.Major == 1 && version.Minor >= 8) || version.Major == 2 {
		livenessProbeHandler = v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path.Clean(webRoutePrefix + "/-/healthy"),
				Port: intstr.FromString("web"),
			},
		}
		readinessProbeHandler = v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path.Clean(webRoutePrefix + "/-/ready"),
				Port: intstr.FromString("web"),
			},
		}
		livenessFailureThreshold = 6
	} else {
		livenessProbeHandler = v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path.Clean(webRoutePrefix + "/status"),
				Port: intstr.FromString("web"),
			},
		}
		readinessProbeHandler = livenessProbeHandler
		// For larger servers, restoring a checkpoint on startup may take quite a bit of time.
		// Wait up to 5 minutes (60 fails * 5s per fail)
		livenessFailureThreshold = 60
	}

	var livenessProbe *v1.Probe
	var readinessProbe *v1.Probe
	if !p.Spec.ListenLocal {
		livenessProbe = &v1.Probe{
			Handler:          livenessProbeHandler,
			PeriodSeconds:    5,
			TimeoutSeconds:   probeTimeoutSeconds,
			FailureThreshold: livenessFailureThreshold,
		}
		readinessProbe = &v1.Probe{
			Handler:          readinessProbeHandler,
			TimeoutSeconds:   probeTimeoutSeconds,
			PeriodSeconds:    5,
			FailureThreshold: 120, // Allow up to 10m on startup for data recovery
		}
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	if p.Spec.PodMetadata != nil {
		if p.Spec.PodMetadata.Labels != nil {
			for k, v := range p.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if p.Spec.PodMetadata.Annotations != nil {
			for k, v := range p.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	podLabels["app"] = "prometheus"
	podLabels["prometheus"] = p.Name

	finalLabels := c.Labels.Merge(podLabels)

	return &appsv1.StatefulSetSpec{
		ServiceName:         governingServiceName,
		Replicas:            p.Spec.Replicas,
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: finalLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				Containers: append([]v1.Container{
					{
						Name:           "prometheus",
						Image:          fmt.Sprintf("%s:%s", p.Spec.BaseImage, p.Spec.Version),
						Ports:          ports,
						Args:           promArgs,
						VolumeMounts:   promVolumeMounts,
						LivenessProbe:  livenessProbe,
						ReadinessProbe: readinessProbe,
						Resources:      p.Spec.Resources,
					}, {
						Name:  "prometheus-config-reloader",
						Image: c.PrometheusConfigReloader,
						Env: []v1.EnvVar{
							{
								Name: "POD_NAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
								},
							},
						},
						Args:         configReloadArgs,
						VolumeMounts: configReloadVolumeMounts,
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("10m"),
								v1.ResourceMemory: resource.MustParse("50Mi"),
							},
						},
					},
				}, p.Spec.Containers...),
				SecurityContext:               securityContext,
				ServiceAccountName:            p.Spec.ServiceAccountName,
				NodeSelector:                  p.Spec.NodeSelector,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:     volumes,
				Tolerations: p.Spec.Tolerations,
				Affinity:    p.Spec.Affinity,
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
