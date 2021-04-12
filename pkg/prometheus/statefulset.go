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

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	governingServiceName            = "prometheus-operated"
	defaultRetention                = "24h"
	defaultReplicaExternalLabelName = "prometheus_replica"
	storageDir                      = "/prometheus"
	confDir                         = "/etc/prometheus/config"
	confOutDir                      = "/etc/prometheus/config_out"
	tlsAssetsDir                    = "/etc/prometheus/certs"
	rulesDir                        = "/etc/prometheus/rules"
	secretsDir                      = "/etc/prometheus/secrets/"
	configmapsDir                   = "/etc/prometheus/configmaps/"
	configFilename                  = "prometheus.yaml.gz"
	configEnvsubstFilename          = "prometheus.env.yaml"
	sSetInputHashName               = "prometheus-operator-input-hash"
	defaultPortName                 = "web"
)

var (
	minShards                   int32 = 1
	minReplicas                 int32 = 1
	defaultMaxConcurrency       int32 = 20
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
	shardLabelName                = "operator.prometheus.io/shard"
	prometheusNameLabelName       = "operator.prometheus.io/name"
	probeTimeoutSeconds     int32 = 3
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

func makeStatefulSet(
	name string,
	p monitoringv1.Prometheus,
	config *operator.Config,
	ruleConfigMapNames []string,
	inputHash string,
	shard int32,
) (*appsv1.StatefulSet, error) {
	// p is passed in by value, not by reference. But p contains references like
	// to annotation map, that do not get copied on function invocation. Ensure to
	// prevent side effects before editing p by creating a deep copy. For more
	// details see https://github.com/prometheus-operator/prometheus-operator/issues/1659.
	p = *p.DeepCopy()

	promVersion := operator.StringValOrDefault(p.Spec.Version, operator.DefaultPrometheusVersion)
	parsedVersion, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse prometheus version")
	}

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
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
	if !memoryRequestFound && parsedVersion.Major == 1 {
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

	spec, err := makeStatefulSetSpec(p, config, shard, ruleConfigMapNames, parsedVersion)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	boolTrue := true
	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	annotations := make(map[string]string)
	for key, value := range p.ObjectMeta.Annotations {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	labels := make(map[string]string)
	for key, value := range p.ObjectMeta.Labels {
		labels[key] = value
	}
	labels[shardLabelName] = fmt.Sprintf("%d", shard)
	labels[prometheusNameLabelName] = p.Name

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      config.Labels.Merge(labels),
			Annotations: annotations,
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
		pvcTemplate := operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(p.Name)
		}
		if storageSpec.VolumeClaimTemplate.Spec.AccessModes == nil {
			pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		} else {
			pvcTemplate.Spec.AccessModes = storageSpec.VolumeClaimTemplate.Spec.AccessModes
		}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, *pvcTemplate)
	}

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, p.Spec.Volumes...)

	return statefulset, nil
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

func makeStatefulSetSpec(p monitoringv1.Prometheus, c *operator.Config, shard int32, ruleConfigMapNames []string,
	version semver.Version) (*appsv1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	prometheusImagePath, err := operator.BuildImagePath(
		operator.StringPtrValOrDefault(p.Spec.Image, ""),
		operator.StringValOrDefault(p.Spec.BaseImage, c.PrometheusDefaultBaseImage),
		p.Spec.Version,
		p.Spec.Tag,
		p.Spec.SHA,
	)
	if err != nil {
		return nil, err
	}

	if version.Major != 2 {
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}

	promArgs := []string{
		"-web.console.templates=/etc/prometheus/consoles",
		"-web.console.libraries=/etc/prometheus/console_libraries",
	}

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

	if p.Spec.Web != nil && p.Spec.Web.PageTitle != nil {
		promArgs = append(promArgs,
			fmt.Sprintf("-web.page-title=%s", *p.Spec.Web.PageTitle),
		)
	}

	if p.Spec.EnableAdminAPI {
		promArgs = append(promArgs, "-web.enable-admin-api")
	}

	if len(p.Spec.EnableFeatures) > 0 {
		promArgs = append(promArgs, "-enable-feature="+strings.Join(p.Spec.EnableFeatures[:], ","))
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

	if version.GTE(semver.MustParse("2.8.0")) && p.Spec.AllowOverlappingBlocks {
		promArgs = append(promArgs, "-storage.tsdb.allow-overlapping-blocks")
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

	for i, a := range promArgs {
		promArgs[i] = "-" + a
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
			Name: "tls-assets",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: tlsAssetsSecretName(p.Name),
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
			Name:      "tls-assets",
			ReadOnly:  true,
			MountPath: tlsAssetsDir,
		},
		{
			Name:      volName,
			MountPath: storageDir,
			SubPath:   subPathForStorage(p.Spec.Storage),
		},
	}

	promVolumeMounts = append(promVolumeMounts, p.Spec.VolumeMounts...)
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

	const localProbe = `if [ -x "$(command -v curl)" ]; then exec curl %s; elif [ -x "$(command -v wget)" ]; then exec wget -q -O /dev/null %s; else exit 1; fi`

	var readinessProbeHandler v1.Handler
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

	// TODO(paulfantom): Re-add livenessProbe and add startupProbe when kubernetes 1.21 is available.
	// This would be a follow-up to https://github.com/prometheus-operator/prometheus-operator/pull/3502
	readinessProbe := &v1.Probe{
		Handler:          readinessProbeHandler,
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 120, // Allow up to 10m on startup for data recovery
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	podSelectorLabels := map[string]string{
		"app":                          "prometheus",
		"app.kubernetes.io/name":       "prometheus",
		"app.kubernetes.io/version":    version.String(),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   p.Name,
		"prometheus":                   p.Name,
		shardLabelName:                 fmt.Sprintf("%d", shard),
		prometheusNameLabelName:        p.Name,
	}
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

	for k, v := range podSelectorLabels {
		podLabels[k] = v
	}

	finalSelectorLabels := c.Labels.Merge(podSelectorLabels)
	finalLabels := c.Labels.Merge(podLabels)

	var additionalContainers []v1.Container

	disableCompaction := p.Spec.DisableCompaction
	if p.Spec.Thanos != nil {
		thanosImage, err := operator.BuildImagePath(
			operator.StringPtrValOrDefault(p.Spec.Thanos.Image, ""),
			operator.StringPtrValOrDefault(p.Spec.Thanos.BaseImage, c.ThanosDefaultBaseImage),
			operator.StringPtrValOrDefault(p.Spec.Thanos.Version, operator.DefaultThanosVersion),
			operator.StringPtrValOrDefault(p.Spec.Thanos.Tag, ""),
			operator.StringPtrValOrDefault(p.Spec.Thanos.SHA, ""),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build image path")
		}

		bindAddress := "[$(POD_IP)]"
		if p.Spec.Thanos.ListenLocal {
			bindAddress = "127.0.0.1"
		}

		thanosArgs := []string{"sidecar",
			fmt.Sprintf("--prometheus.url=http://%s:9090%s", c.LocalHost, path.Clean(webRoutePrefix)),
			fmt.Sprintf("--grpc-address=%s:10901", bindAddress),
			fmt.Sprintf("--http-address=%s:10902", bindAddress),
		}

		if p.Spec.Thanos.GRPCServerTLSConfig != nil {
			tls := p.Spec.Thanos.GRPCServerTLSConfig
			if tls.CertFile != "" {
				thanosArgs = append(thanosArgs, "--grpc-server-tls-cert="+tls.CertFile)
			}
			if tls.KeyFile != "" {
				thanosArgs = append(thanosArgs, "--grpc-server-tls-key="+tls.KeyFile)
			}
			if tls.CAFile != "" {
				thanosArgs = append(thanosArgs, "--grpc-server-tls-client-ca="+tls.CAFile)
			}
		}

		container := v1.Container{
			Name:                     "thanos-sidecar",
			Image:                    thanosImage,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			Args:                     thanosArgs,
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
			Resources: p.Spec.Thanos.Resources,
		}

		if p.Spec.Thanos.ObjectStorageConfig != nil || p.Spec.Thanos.ObjectStorageConfigFile != nil {
			if p.Spec.Thanos.ObjectStorageConfigFile != nil {
				container.Args = append(container.Args, "--objstore.config-file="+*p.Spec.Thanos.ObjectStorageConfigFile)
			} else {
				container.Args = append(container.Args, "--objstore.config=$(OBJSTORE_CONFIG)")
				container.Env = append(container.Env, v1.EnvVar{
					Name: "OBJSTORE_CONFIG",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: p.Spec.Thanos.ObjectStorageConfig,
					},
				})
			}
			container.Args = append(container.Args, fmt.Sprintf("--tsdb.path=%s", storageDir))
			container.VolumeMounts = append(
				container.VolumeMounts,
				v1.VolumeMount{
					Name:      volName,
					MountPath: storageDir,
					SubPath:   subPathForStorage(p.Spec.Storage),
				},
			)

			// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/ we have to turn off compaction of Prometheus
			// to avoid races during upload, if the uploads are configured.
			disableCompaction = true
		}

		if p.Spec.Thanos.TracingConfig != nil || len(p.Spec.Thanos.TracingConfigFile) > 0 {
			if len(p.Spec.Thanos.TracingConfigFile) > 0 {
				container.Args = append(container.Args, "--tracing.config-file="+p.Spec.Thanos.TracingConfigFile)
			} else {
				container.Args = append(container.Args, "--tracing.config=$(TRACING_CONFIG)")
				container.Env = append(container.Env, v1.EnvVar{
					Name: "TRACING_CONFIG",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: p.Spec.Thanos.TracingConfig,
					},
				})
			}
		}

		if p.Spec.Thanos.LogLevel != "" {
			container.Args = append(container.Args, "--log.level="+p.Spec.Thanos.LogLevel)
		} else if p.Spec.LogLevel != "" {
			container.Args = append(container.Args, "--log.level="+p.Spec.LogLevel)
		}
		if p.Spec.Thanos.LogFormat != "" {
			container.Args = append(container.Args, "--log.format="+p.Spec.Thanos.LogFormat)
		} else if p.Spec.LogFormat != "" {
			container.Args = append(container.Args, "--log.format="+p.Spec.LogFormat)
		}

		if p.Spec.Thanos.MinTime != "" {
			container.Args = append(container.Args, "--min-time="+p.Spec.Thanos.MinTime)
		}
		additionalContainers = append(additionalContainers, container)
	}
	if disableCompaction {
		promArgs = append(promArgs, "--storage.tsdb.max-block-duration=2h")
		promArgs = append(promArgs, "--storage.tsdb.min-block-duration=2h")
	}

	configReloaderArgs := []string{
		fmt.Sprintf("--config-file=%s", path.Join(confDir, configFilename)),
		fmt.Sprintf("--config-envsubst-file=%s", path.Join(confOutDir, configEnvsubstFilename)),
	}
	configReloaderVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			MountPath: confDir,
		},
		{
			Name:      "config-out",
			MountPath: confOutDir,
		},
	}
	if len(ruleConfigMapNames) != 0 {
		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			configReloaderArgs = append(configReloaderArgs, fmt.Sprintf("--watched-dir=%s", mountPath))
		}
	}

	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    prometheusImagePath,
			Ports:                    ports,
			Args:                     promArgs,
			VolumeMounts:             promVolumeMounts,
			ReadinessProbe:           readinessProbe,
			Resources:                p.Spec.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
		operator.CreateConfigReloader(
			c.ReloaderConfig,
			url.URL{
				Scheme: "http",
				Host:   c.LocalHost + ":9090",
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			},
			p.Spec.ListenLocal,
			c.LocalHost,
			p.Spec.LogFormat,
			p.Spec.LogLevel,
			configReloaderArgs,
			configReloaderVolumeMounts,
			shard,
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
				InitContainers:                p.Spec.InitContainers,
				SecurityContext:               p.Spec.SecurityContext,
				ServiceAccountName:            p.Spec.ServiceAccountName,
				NodeSelector:                  p.Spec.NodeSelector,
				PriorityClassName:             p.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   p.Spec.Tolerations,
				Affinity:                      p.Spec.Affinity,
				TopologySpreadConstraints:     p.Spec.TopologySpreadConstraints,
			},
		},
	}, nil
}

func configSecretName(name string) string {
	return prefixedName(name)
}

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("prometheus-%s", name)
}

func subPathForStorage(s *monitoringv1.StorageSpec) string {
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if s == nil || s.DisableMountSubPath {
		return ""
	}

	return "prometheus-db"
}
