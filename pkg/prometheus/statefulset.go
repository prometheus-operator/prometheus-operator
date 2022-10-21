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
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	governingServiceName            = "prometheus-operated"
	defaultReplicaExternalLabelName = "prometheus_replica"

	configFilename           = "prometheus.yaml.gz"
	defaultPortName          = "web"
	defaultQueryLogDirectory = "/var/log/prometheus"
)

var (
	minShards                   int32 = 1
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}

	prometheusNameLabelName = "operator.prometheus.io/name"
)

func expectedStatefulSetShardNames(
	p *monitoringv1.Prometheus,
) []string {
	res := []string{}
	shards := minShards
	if p.Spec.Shards != nil && *p.Spec.Shards > 1 {
		shards = *p.Spec.Shards
	}

	for i := int32(0); i < shards; i++ {
		res = append(res, prometheusNameByShard(p.Name, i))
	}

	return res
}

func prometheusNameByShard(name string, shard int32) string {
	base := prefixedName(name)
	if shard == 0 {
		return base
	}
	return fmt.Sprintf("%s-shard-%d", base, shard)
}

func makeEmptyConfigurationSecret(p *monitoringv1.Prometheus, config operator.Config) (*v1.Secret, error) {
	s := makeConfigSecret(p, config)

	s.ObjectMeta.Annotations = map[string]string{
		"empty": "true",
	}

	return s, nil
}

func makeConfigSecret(p *monitoringv1.Prometheus, config operator.Config) *v1.Secret {
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

func makeStatefulSetService(p *monitoringv1.Prometheus, config operator.Config) *v1.Service {
	p = p.DeepCopy()

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			OwnerReferences: []metav1.OwnerReference{
				{
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
				"app.kubernetes.io/name": "prometheus",
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

func configSecretName(name string) string {
	return prefixedName(name)
}

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}

func webConfigSecretName(name string) string {
	return fmt.Sprintf("%s-web-config", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("prometheus-%s", name)
}

func usesDefaultQueryLogVolume(p *monitoringv1.Prometheus) bool {
	return p.Spec.QueryLogFile != "" && filepath.Dir(p.Spec.QueryLogFile) == "."
}

func queryLogFilePath(p *monitoringv1.Prometheus) string {
	if !usesDefaultQueryLogVolume(p) {
		return p.Spec.QueryLogFile
	}

	return filepath.Join(defaultQueryLogDirectory, p.Spec.QueryLogFile)
}

// makeThanosCommandArgs returns slice of Thanos command arguments for Thanos sidecar
func makeThanosCommandArgs(thanos monitoringv1.ThanosSpec, c *operator.Config, webRoutePrefix, uriScheme string) (out []string, err error) {
	bindAddress := "" // Listen to all available IP addresses by default
	if thanos.ListenLocal {
		bindAddress = "127.0.0.1"
	}

	args := map[string]string{
		"prometheus.url": fmt.Sprintf("%s://%s:9090%s", uriScheme, c.LocalHost, path.Clean(webRoutePrefix)),
		"grpc-address":   fmt.Sprintf("%s:%d", bindAddress, DefaultThanosGRPCPort),
		"http-address":   fmt.Sprintf("%s:%d", bindAddress, DefaultThanosHTTPPort),
	}

	if thanos.GRPCServerTLSConfig != nil {
		tls := thanos.GRPCServerTLSConfig
		if tls.CertFile != "" {
			args["grpc-server-tls-cert"] = tls.CertFile
		}
		if tls.KeyFile != "" {
			args["grpc-server-tls-key"] = tls.KeyFile
		}
		if tls.CAFile != "" {
			args["grpc-server-tls-client-ca"] = tls.CAFile
		}
	}

	if thanos.ObjectStorageConfig != nil || thanos.ObjectStorageConfigFile != nil {
		if thanos.ObjectStorageConfigFile != nil {
			args["objstore.config-file"] = *thanos.ObjectStorageConfigFile
		} else {
			args["objstore.config"] = fmt.Sprintf("$(%s)", ThanosObjStoreEnvVar)
		}
		args["tsdb.path"] = PrometheusStorageDir
	}

	if thanos.TracingConfig != nil || len(thanos.TracingConfigFile) > 0 {
		traceConfig := fmt.Sprintf("$(%s)", ThanosTraceConfigEnvVar)
		if len(thanos.TracingConfigFile) > 0 {
			traceConfig = thanos.TracingConfigFile
		}
		args["tracing.config"] = traceConfig
	}

	logLevel, logFormat := pt.GetLoggerInfo()
	if thanos.LogLevel != "" {
		logLevel = thanos.LogLevel
	}
	if logLevel != "" {
		args["log.level"] = logLevel
	}
	if thanos.LogFormat != "" {
		logFormat = thanos.LogFormat
	}
	if logFormat != "" {
		args["log.format"] = logFormat
	}

	if thanos.MinTime != "" {
		args["min-time"] = thanos.MinTime
	}

	if thanos.ReadyTimeout != "" {
		args["prometheus.ready_timeout"] = string(thanos.ReadyTimeout)
	}

	out, err := operator.ProcessCommandArgs(args, thanos.AdditionalArgs)
	if err != nil {
		return out, err
	}
	return append([]string{"sidecar"}, out...), err
}

func makeStatefulSetSpec(
	logger log.Logger,
	p monitoringv1.Prometheus,
	c *operator.Config,
	shard int32,
	ruleConfigMapNames []string,
	tlsAssetSecrets []string,
	version semver.Version,
) (*appsv1.StatefulSetSpec, error) {
	imagePath, err := operator.BuildImagePath(
		operator.StringPtrValOrDefault(p.Spec.Image, ""),
		operator.StringValOrDefault(p.Spec.BaseImage, c.PrometheusDefaultBaseImage),
		p.Spec.Version,
		p.Spec.Tag,
		p.Spec.SHA,
	)
	if err != nil {
		return nil, err
	}

	uriScheme := operator.GetURIScheme(p.Spec.Web)
	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}

	// volume factory
	func() ([]v1.Volume, []v1.VolumeMount, error) {
		addVolumes := []v1.Volume{}
		// query log volume
		if volume, ok := queryLogFileVolume(&p); ok {
			addVolumes = append(addVolumes, volume)
		}
		// rule config maps volumes
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

		addMounts := []v1.VolumeMount{}
		for _, name := range ruleConfigMapNames {
			addMounts = append(addMounts, v1.VolumeMount{
				Name:      name,
				MountPath: path.Join(PrometheusRulesDir, name),
			})
		}
		if vmount, ok := queryLogFileVolumeMount(&p); ok {
			addMounts = append(addMounts, vmount)
		}

		return addVolumes, addMounts, nil
	}

	// container factory
	func() ([]v1.Container, error) {
		addContainers := []v1.Container
		if p.Spec.Thanos != nil {
			thanosImage, err := operator.BuildImagePath(
				operator.StringPtrValOrDefault(p.Spec.Thanos.Image, ""),
				operator.StringPtrValOrDefault(p.Spec.Thanos.BaseImage, c.ThanosDefaultBaseImage),
				operator.StringPtrValOrDefault(p.Spec.Thanos.Version, operator.DefaultThanosVersion),
				operator.StringPtrValOrDefault(p.Spec.Thanos.Tag, ""),
				operator.StringPtrValOrDefault(p.Spec.Thanos.SHA, ""),
			)
			if err != nil {
				return addContainers, errors.Wrap(err, "failed to build image path")
			}

			thanosArgs, err := makeThanosCommandArgs(p.Spec.Thanos, c, webRoutePrefix, uriScheme)
			if err != nil {
				return addContainers, err
			}
      container := v1.Container{
      	Name:                     "thanos-sidecar",
        Image:                    thanosImage,
        TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
        Args:                     thanosArgs,
        SecurityContext: &v1.SecurityContext{
        	AllowPrivilegeEscalation: &boolFalse,
          ReadOnlyRootFilesystem:   &boolTrue,
          Capabilities: &v1.Capabilities{
            Drop: []v1.Capability{"ALL"},
          },
        },
        Ports: []v1.ContainerPort{
          {Name: "http", ContainerPort: DefaultThanosHTTPPort},
          {Name: "grpc", ContainerPort: DefaultThanosGRPCPort},
        },
        Resources: thanos.Resources,
      }

			for _, thanosSideCarVM := range thanos.VolumeMounts {
				container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
					Name:      thanosSideCarVM.Name,
					MountPath: thanosSideCarVM.MountPath,
				})
			}

			if thanos.ObjectStorageConfig != nil || thanos.ObjectStorageConfigFile != nil {
				if thanos.ObjectStorageConfigFile == nil {
					container.Env = append(container.Env, v1.EnvVar{
						Name: ThanosObjStoreEnvVar,
						ValueFrom: &v1.EnvVarSource{
							SecretKeyRef: thanos.ObjectStorageConfig,
						},
					})
				}
				container.VolumeMounts = append(
					container.VolumeMounts,
					v1.VolumeMount{
						Name:      volName,
						MountPath: operator.PrometheusStorageDir,
						SubPath:   nc.SubPathForStorage(storage),
					},
				)
			}

			if thanos.TracingConfig != nil && len(thanos.TracingConfigFile) == 0 {
      	container.Env = append(container.Env, v1.EnvVar{
	      	Name: ThanosTraceConfigEnvVar,
	      	ValueFrom: &v1.EnvVarSource{
	          SecretKeyRef: thanos.TracingConfig,
	       	},
				})
     	}

			addContainers = append(addContainers, container)
		}

		return addContainers, nil
	}

	if len(ruleConfigMapNames) != 0 {
		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			watchedDirectories = append(watchedDirectories, mountPath)
		}
	}

	var minReadySeconds int32
	if p.Spec.MinReadySeconds != nil {
		minReadySeconds = int32(*p.Spec.MinReadySeconds)
	}

	operatorInitContainers = append(operatorInitContainers,
		operator.CreateConfigReloader(
			"init-config-reloader",
			operator.ReloaderResources(c.ReloaderConfig),
			operator.ReloaderRunOnce(),
			operator.LogFormat(p.Spec.LogFormat),
			operator.LogLevel(p.Spec.LogLevel),
			operator.VolumeMounts(configReloaderVolumeMounts),
			operator.ConfigFile(path.Join(confDir, configFilename)),
			operator.ConfigEnvsubstFile(path.Join(confOutDir, configEnvsubstFilename)),
			operator.WatchedDirectories(watchedDirectories),
			operator.Shard(shard),
		),
	)

	initContainers, err := k8sutil.MergePatchContainers(operatorInitContainers, p.Spec.InitContainers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge init containers spec")
	}

	containerArgs, err := buildArgs(promArgs, p.Spec.AdditionalArgs)

	if err != nil {
		return nil, err
	}

	boolFalse := false
	boolTrue := true
	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    prometheusImagePath,
			Ports:                    ports,
			Args:                     containerArgs,
			VolumeMounts:             promVolumeMounts,
			StartupProbe:             startupProbe,
			LivenessProbe:            livenessProbe,
			ReadinessProbe:           readinessProbe,
			Resources:                p.Spec.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			SecurityContext: &v1.SecurityContext{
				ReadOnlyRootFilesystem:   &boolTrue,
				AllowPrivilegeEscalation: &boolFalse,
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
			},
		},
		operator.CreateConfigReloader(
			"config-reloader",
			operator.ReloaderResources(c.ReloaderConfig),
			operator.ReloaderURL(url.URL{
				Scheme: prometheusURIScheme,
				Host:   c.LocalHost + ":9090",
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			}),
			operator.ListenLocal(p.Spec.ListenLocal),
			operator.LocalHost(c.LocalHost),
			operator.LogFormat(p.Spec.LogFormat),
			operator.LogLevel(p.Spec.LogLevel),
			operator.ConfigFile(path.Join(confDir, configFilename)),
			operator.ConfigEnvsubstFile(path.Join(confOutDir, configEnvsubstFilename)),
			operator.WatchedDirectories(watchedDirectories), operator.VolumeMounts(configReloaderVolumeMounts),
			operator.Shard(shard),
		),
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, p.Spec.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		ServiceName:         governingServiceName,
		Replicas:            p.Spec.Replicas,
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		MinReadySeconds: minReadySeconds,
		Selector: &metav1.LabelSelector{
			MatchLabels: finalSelectorLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				Containers:                    containers,
				InitContainers:                initContainers,
				SecurityContext:               p.Spec.SecurityContext,
				ServiceAccountName:            p.Spec.ServiceAccountName,
				AutomountServiceAccountToken:  &boolTrue,
				NodeSelector:                  p.Spec.NodeSelector,
				PriorityClassName:             p.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   p.Spec.Tolerations,
				Affinity:                      p.Spec.Affinity,
				TopologySpreadConstraints:     p.Spec.TopologySpreadConstraints,
				HostAliases:                   operator.MakeHostAliases(p.Spec.HostAliases),
			},
		},
	}, nil
}
