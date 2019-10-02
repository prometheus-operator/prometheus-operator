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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/blang/semver"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
)

const (
	governingServiceName            = "prometheus-operated"
	DefaultPrometheusVersion        = "v2.7.1"
	DefaultThanosVersion            = "v0.7.0"
	defaultRetention                = "24h"
	defaultReplicaExternalLabelName = "prometheus_replica"
	storageDir                      = "/prometheus"
	confDir                         = "/etc/prometheus/config"
	confOutDir                      = "/etc/prometheus/config_out"
	rulesDir                        = "/etc/prometheus/rules"
	secretsDir                      = "/etc/prometheus/secrets/"
	configmapsDir                   = "/etc/prometheus/configmaps/"
	configFilename                  = "prometheus.yaml.gz"
	configEnvsubstFilename          = "prometheus.env.yaml"
	sSetInputHashName               = "prometheus-operator-input-hash"
	defaultPortName                 = "web"
)

var (
	minReplicas                 int32 = 1
	defaultMaxConcurrency       int32 = 20
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
		"v2.3.1",
		"v2.3.2",
		"v2.4.0",
		"v2.4.1",
		"v2.4.2",
		"v2.4.3",
		"v2.5.0",
		"v2.6.0",
		"v2.6.1",
		"v2.7.0",
		"v2.7.1",
		"v2.7.2",
		"v2.8.1",
		"v2.9.2",
		"v2.10.0",
	}
)

func makeStatefulSet(
	p monitoringv1.Prometheus,
	config *Config,
	ruleConfigMapNames []string,
	inputHash string,
) (*appsv1.StatefulSet, error) {
	// p is passed in by value, not by reference. But p contains references like
	// to annotation map, that do not get copied on function invocation. Ensure to
	// prevent side effects before editing p by creating a deep copy. For more
	// details see https://github.com/coreos/prometheus-operator/issues/1659.
	p = *p.DeepCopy()

	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if p.Spec.BaseImage == "" {
		p.Spec.BaseImage = config.PrometheusDefaultBaseImage
	}
	if p.Spec.Version == "" {
		p.Spec.Version = DefaultPrometheusVersion
	}
	if p.Spec.Thanos != nil && p.Spec.Thanos.Version == nil {
		v := DefaultThanosVersion
		p.Spec.Thanos.Version = &v
	}

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
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

	spec, err := makeStatefulSetSpec(p, config, ruleConfigMapNames)
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

	if statefulset.ObjectMeta.Annotations == nil {
		statefulset.ObjectMeta.Annotations = map[string]string{
			sSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[sSetInputHashName] = inputHash
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
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(p.Name)
		}
		pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvcTemplate)
	}

	for _, volume := range p.Spec.Volumes {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, volume)
	}

	return statefulset, nil
}

func makeEmptyConfigurationSecret(p *monitoringv1.Prometheus, config Config) (*v1.Secret, error) {
	s := makeConfigSecret(p, config)

	s.ObjectMeta.Annotations = map[string]string{
		"empty": "true",
	}

	return s, nil
}

func makeConfigSecret(p *monitoringv1.Prometheus, config Config) *v1.Secret {
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
			configFilename: {},
		},
	}
}

func makeStatefulSetService(p *monitoringv1.Prometheus, config Config) *v1.Service {
	p = p.DeepCopy()

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					Name:       p.GetName(),
					Kind:       p.Kind,
					APIVersion: p.APIVersion,
					UID:        p.GetUID(),
				},
			},
			Labels: config.Labels.Merge(map[string]string{
				"operated-prometheus": "true",
			}),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       p.Spec.PortName,
					Port:       9090,
					TargetPort: intstr.FromString(p.Spec.PortName),
				},
			},
			Selector: map[string]string{
				"app": "prometheus",
			},
		},
	}

	if p.Spec.Thanos != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       "grpc",
			Port:       10901,
			TargetPort: intstr.FromString("grpc"),
		})
	}

	return svc
}

func makeStatefulSetSpec(p monitoringv1.Prometheus, c *Config, ruleConfigMapNames []string) (*appsv1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	versionStr := strings.TrimLeft(p.Spec.Version, "v")

	version, err := semver.Parse(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	promArgs := []string{
		"-web.console.templates=/etc/prometheus/consoles",
		"-web.console.libraries=/etc/prometheus/console_libraries",
	}

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
	case 2:
		retentionTimeFlag := "-storage.tsdb.retention="
		if version.Minor >= 7 {
			retentionTimeFlag = "-storage.tsdb.retention.time="
			if p.Spec.RetentionSize != "" {
				promArgs = append(promArgs,
					fmt.Sprintf("-storage.tsdb.retention.size=%s", p.Spec.RetentionSize),
				)
			}
		}
		promArgs = append(promArgs,
			fmt.Sprintf("-config.file=%s", path.Join(confOutDir, configEnvsubstFilename)),
			fmt.Sprintf("-storage.tsdb.path=%s", storageDir),
			retentionTimeFlag+p.Spec.Retention,
			"-web.enable-lifecycle",
			"-storage.tsdb.no-lockfile",
		)

		if p.Spec.Query != nil && p.Spec.Query.LookbackDelta != nil {
			promArgs = append(promArgs,
				fmt.Sprintf("-query.lookback-delta=%s", *p.Spec.Query.LookbackDelta),
			)
		}

		if version.Minor >= 4 {
			if p.Spec.Rules.Alert.ForOutageTolerance != "" {
				promArgs = append(promArgs, "-rules.alert.for-outage-tolerance="+p.Spec.Rules.Alert.ForOutageTolerance)
			}
			if p.Spec.Rules.Alert.ForGracePeriod != "" {
				promArgs = append(promArgs, "-rules.alert.for-grace-period="+p.Spec.Rules.Alert.ForGracePeriod)
			}
			if p.Spec.Rules.Alert.ResendDelay != "" {
				promArgs = append(promArgs, "-rules.alert.resend-delay="+p.Spec.Rules.Alert.ResendDelay)
			}
		}

		if version.Minor >= 5 {
			if p.Spec.Query != nil && p.Spec.Query.MaxSamples != nil {
				promArgs = append(promArgs,
					fmt.Sprintf("-query.max-samples=%d", *p.Spec.Query.MaxSamples),
				)
			}
		}
	default:
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}

	if p.Spec.Query != nil {
		if p.Spec.Query.MaxConcurrency != nil {
			if *p.Spec.Query.MaxConcurrency < 1 {
				p.Spec.Query.MaxConcurrency = &defaultMaxConcurrency
			}
			promArgs = append(promArgs,
				fmt.Sprintf("-query.max-concurrency=%d", *p.Spec.Query.MaxConcurrency),
			)
		}
		if p.Spec.Query.Timeout != nil {
			promArgs = append(promArgs,
				fmt.Sprintf("-query.timeout=%s", *p.Spec.Query.Timeout),
			)
		}
	}

	var securityContext *v1.PodSecurityContext = nil
	if p.Spec.SecurityContext != nil {
		securityContext = p.Spec.SecurityContext
	}

	if p.Spec.EnableAdminAPI {
		promArgs = append(promArgs, "-web.enable-admin-api")
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
	if version.GTE(semver.MustParse("2.6.0")) {
		if p.Spec.LogFormat != "" && p.Spec.LogFormat != "logfmt" {
			promArgs = append(promArgs, fmt.Sprintf("-log.format=%s", p.Spec.LogFormat))
		}
	}

	if version.GTE(semver.MustParse("2.11.0")) && p.Spec.WALCompression != nil {
		if *p.Spec.WALCompression {
			promArgs = append(promArgs, "-storage.tsdb.wal-compression")
		} else {
			promArgs = append(promArgs, "-no-storage.tsdb.wal-compression")
		}
	}

	var ports []v1.ContainerPort
	if p.Spec.ListenLocal {
		promArgs = append(promArgs, "-web.listen-address=127.0.0.1:9090")
	} else {
		ports = []v1.ContainerPort{
			{
				Name:          p.Spec.PortName,
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
		Host:   c.LocalHost + ":9090",
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

	for _, name := range ruleConfigMapNames {
		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
				},
			},
		})
	}

	volName := volumeName(p.Name)
	if p.Spec.Storage != nil {
		if p.Spec.Storage.VolumeClaimTemplate.Name != "" {
			volName = p.Spec.Storage.VolumeClaimTemplate.Name
		}
	}

	promVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config-out",
			ReadOnly:  true,
			MountPath: confOutDir,
		},
		{
			Name:      volName,
			MountPath: storageDir,
			SubPath:   subPathForStorage(p.Spec.Storage),
		},
	}

	for _, name := range ruleConfigMapNames {
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: rulesDir + "/" + name,
		})
	}

	for _, s := range p.Spec.Secrets {
		volumes = append(volumes, v1.Volume{
			Name: k8sutil.SanitizeVolumeName("secret-" + s),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("secret-" + s),
			ReadOnly:  true,
			MountPath: secretsDir + s,
		})
	}

	for _, c := range p.Spec.ConfigMaps {
		volumes = append(volumes, v1.Volume{
			Name: k8sutil.SanitizeVolumeName("configmap-" + c),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c,
					},
				},
			},
		})
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("configmap-" + c),
			ReadOnly:  true,
			MountPath: configmapsDir + c,
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
		fmt.Sprintf("--log-format=%s", c.LogFormat),
		fmt.Sprintf("--reload-url=%s", localReloadURL),
		fmt.Sprintf("--config-file=%s", path.Join(confDir, configFilename)),
		fmt.Sprintf("--config-envsubst-file=%s", path.Join(confOutDir, configEnvsubstFilename)),
	}

	const localProbe = `if [ -x "$(command -v curl)" ]; then curl %s; elif [ -x "$(command -v wget)" ]; then wget -q %s; else exit 1; fi`

	var livenessProbeHandler v1.Handler
	var readinessProbeHandler v1.Handler
	var livenessFailureThreshold int32
	if (version.Major == 1 && version.Minor >= 8) || version.Major == 2 {
		{
			healthyPath := path.Clean(webRoutePrefix + "/-/healthy")
			if p.Spec.ListenLocal {
				localHealthyPath := fmt.Sprintf("http://localhost:9090%s", healthyPath)
				livenessProbeHandler.Exec = &v1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						fmt.Sprintf(localProbe, localHealthyPath, localHealthyPath),
					},
				}
			} else {
				livenessProbeHandler.HTTPGet = &v1.HTTPGetAction{
					Path: healthyPath,
					Port: intstr.FromString(p.Spec.PortName),
				}
			}
		}
		{
			readyPath := path.Clean(webRoutePrefix + "/-/ready")
			if p.Spec.ListenLocal {
				localReadyPath := fmt.Sprintf("http://localhost:9090%s", readyPath)
				readinessProbeHandler.Exec = &v1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						fmt.Sprintf(localProbe, localReadyPath, localReadyPath),
					},
				}

			} else {
				readinessProbeHandler.HTTPGet = &v1.HTTPGetAction{
					Path: readyPath,
					Port: intstr.FromString(p.Spec.PortName),
				}
			}
		}

		livenessFailureThreshold = 6

	} else {
		livenessProbeHandler = v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path.Clean(webRoutePrefix + "/status"),
				Port: intstr.FromString(p.Spec.PortName),
			},
		}
		readinessProbeHandler = livenessProbeHandler
		// For larger servers, restoring a checkpoint on startup may take quite a bit of time.
		// Wait up to 5 minutes (60 fails * 5s per fail)
		livenessFailureThreshold = 60
	}

	livenessProbe := &v1.Probe{
		Handler:          livenessProbeHandler,
		PeriodSeconds:    5,
		TimeoutSeconds:   probeTimeoutSeconds,
		FailureThreshold: livenessFailureThreshold,
	}
	readinessProbe := &v1.Probe{
		Handler:          readinessProbeHandler,
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 120, // Allow up to 10m on startup for data recovery
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

	var additionalContainers []v1.Container

	if len(ruleConfigMapNames) != 0 {
		container := v1.Container{
			Name:  "rules-configmap-reloader",
			Image: c.ConfigReloaderImage,
			Args: []string{
				fmt.Sprintf("--webhook-url=%s", localReloadURL),
			},
			VolumeMounts: []v1.VolumeMount{},
			Resources:    v1.ResourceRequirements{Limits: v1.ResourceList{}},
		}

		if c.ConfigReloaderCPU != "0" {
			container.Resources.Limits[v1.ResourceCPU] = resource.MustParse(c.ConfigReloaderCPU)
		}
		if c.ConfigReloaderMemory != "0" {
			container.Resources.Limits[v1.ResourceMemory] = resource.MustParse(c.ConfigReloaderMemory)
		}

		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			container.Args = append(container.Args, fmt.Sprintf("--volume-dir=%s", mountPath))
		}

		additionalContainers = append(additionalContainers, container)
	}

	if p.Spec.Thanos != nil {
		// Version is used by default.
		// If the tag is specified, we use the tag to identify the container image.
		// If the sha is specified, we use the sha to identify the container image,
		// as it has even stronger immutable guarantees to identify the image.
		thanosBaseImage := c.ThanosDefaultBaseImage
		if p.Spec.Thanos.BaseImage != nil {
			thanosBaseImage = *p.Spec.Thanos.BaseImage
		}
		thanosImage := fmt.Sprintf("%s:%s", thanosBaseImage, *p.Spec.Thanos.Version)
		if p.Spec.Thanos.Tag != nil {
			thanosImage = fmt.Sprintf("%s:%s", thanosBaseImage, *p.Spec.Thanos.Tag)
		}
		if p.Spec.Thanos.SHA != nil {
			thanosImage = fmt.Sprintf("%s@sha256:%s", thanosBaseImage, *p.Spec.Thanos.SHA)
		}
		if p.Spec.Thanos.Image != nil && *p.Spec.Thanos.Image != "" {
			thanosImage = *p.Spec.Thanos.Image
		}
		bindAddress := "[$(POD_IP)]"
		if p.Spec.Thanos.ListenLocal {
			bindAddress = "127.0.0.1"
		}

		container := v1.Container{
			Name:  "thanos-sidecar",
			Image: thanosImage,
			Args: []string{
				"sidecar",
				fmt.Sprintf("--prometheus.url=http://%s:9090%s", c.LocalHost, path.Clean(webRoutePrefix)),
				fmt.Sprintf("--tsdb.path=%s", storageDir),
				fmt.Sprintf("--grpc-address=%s:10901", bindAddress),
				fmt.Sprintf("--http-address=%s:10902", bindAddress),
			},
			Env: []v1.EnvVar{
				{
					Name: "POD_IP",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 10902,
				},
				{
					Name:          "grpc",
					ContainerPort: 10901,
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      volName,
					MountPath: storageDir,
					SubPath:   subPathForStorage(p.Spec.Storage),
				},
			},
			Resources: p.Spec.Thanos.Resources,
		}

		if p.Spec.Thanos.ObjectStorageConfig != nil {
			container.Args = append(container.Args, "--objstore.config=$(OBJSTORE_CONFIG)")
			container.Env = append(container.Env, v1.EnvVar{
				Name: "OBJSTORE_CONFIG",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: p.Spec.Thanos.ObjectStorageConfig,
				},
			})
		}

		if p.Spec.LogLevel != "" {
			container.Args = append(container.Args, fmt.Sprintf("--log.level=%s", p.Spec.LogLevel))
		}
		if p.Spec.LogFormat != "" {
			container.Args = append(container.Args, fmt.Sprintf("--log.format=%s", p.Spec.LogFormat))
		}

		additionalContainers = append(additionalContainers, container)
		promArgs = append(promArgs, "--storage.tsdb.min-block-duration=2h", "--storage.tsdb.max-block-duration=2h")
	}

	// Version is used by default.
	// If the tag is specified, we use the tag to identify the container image.
	// If the sha is specified, we use the sha to identify the container image,
	// as it has even stronger immutable guarantees to identify the image.
	prometheusImage := fmt.Sprintf("%s:%s", p.Spec.BaseImage, p.Spec.Version)
	if p.Spec.Tag != "" {
		prometheusImage = fmt.Sprintf("%s:%s", p.Spec.BaseImage, p.Spec.Tag)
	}
	if p.Spec.SHA != "" {
		prometheusImage = fmt.Sprintf("%s@sha256:%s", p.Spec.BaseImage, p.Spec.SHA)
	}
	if p.Spec.Image != nil && *p.Spec.Image != "" {
		prometheusImage = *p.Spec.Image
	}

	prometheusConfigReloaderResources := v1.ResourceRequirements{Limits: v1.ResourceList{}}
	if c.ConfigReloaderCPU != "0" {
		prometheusConfigReloaderResources.Limits[v1.ResourceCPU] = resource.MustParse(c.ConfigReloaderCPU)
	}
	if c.ConfigReloaderMemory != "0" {
		prometheusConfigReloaderResources.Limits[v1.ResourceMemory] = resource.MustParse(c.ConfigReloaderMemory)
	}

	operatorContainers := append([]v1.Container{
		{
			Name:           "prometheus",
			Image:          prometheusImage,
			Ports:          ports,
			Args:           promArgs,
			VolumeMounts:   promVolumeMounts,
			LivenessProbe:  livenessProbe,
			ReadinessProbe: readinessProbe,
			Resources:      p.Spec.Resources,
		}, {
			Name:  "prometheus-config-reloader",
			Image: c.PrometheusConfigReloaderImage,
			Env: []v1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
					},
				},
			},
			Command:      []string{"/bin/prometheus-config-reloader"},
			Args:         configReloadArgs,
			VolumeMounts: configReloadVolumeMounts,
			Resources:    prometheusConfigReloaderResources,
		},
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, p.Spec.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}
	// PodManagementPolicy is set to Parallel to mitigate issues in kuberentes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
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
				Containers:                    containers,
				InitContainers:                p.Spec.InitContainers,
				SecurityContext:               securityContext,
				ServiceAccountName:            p.Spec.ServiceAccountName,
				NodeSelector:                  p.Spec.NodeSelector,
				PriorityClassName:             p.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   p.Spec.Tolerations,
				Affinity:                      p.Spec.Affinity,
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
