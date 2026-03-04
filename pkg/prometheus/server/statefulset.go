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
	"maps"
	"path"
	"path/filepath"

	"github.com/blang/semver/v4"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

const (
	defaultRetention                     = "24h"
	prometheusMode                       = "server"
	governingServiceName                 = "prometheus-operated"
	thanosSupportedVersionHTTPClientFlag = "0.24.0"
)

func makeStatefulSet(
	name string,
	p *monitoringv1.Prometheus,
	config prompkg.Config,
	cg *prompkg.ConfigGenerator,
	ruleConfigMapNames []string,
	inputHash string,
	shard int32,
	tlsSecrets *operator.ShardedSecret,
) (*appsv1.StatefulSet, error) {
	cpf := p.GetCommonPrometheusFields()
	objMeta := p.GetObjectMeta()

	if cpf.PortName == "" {
		cpf.PortName = prompkg.DefaultPortName
	}

	cpf.Replicas = prompkg.ReplicasNumberPtr(p)

	// We need to re-set the common fields because cpf is only a copy of the original object.
	// We set some defaults if some fields are not present, and we want those fields set in the original Prometheus object before building the StatefulSetSpec.
	p.SetCommonPrometheusFields(cpf)
	spec, err := makeStatefulSetSpec(p, config, cg, shard, ruleConfigMapNames, tlsSecrets)
	if err != nil {
		return nil, fmt.Errorf("make StatefulSet spec: %w", err)
	}

	statefulset := &appsv1.StatefulSet{Spec: *spec}

	operator.UpdateObject(
		statefulset,
		operator.WithName(name),
		operator.WithAnnotations(objMeta.GetAnnotations()),
		operator.WithAnnotations(config.Annotations),
		operator.WithInputHashAnnotation(inputHash),
		operator.WithLabels(objMeta.GetLabels()),
		operator.WithLabels(map[string]string{
			prompkg.PrometheusModeLabelName: prometheusMode,
		}),
		operator.WithSelectorLabels(spec.Selector),
		operator.WithLabels(config.Labels),
		operator.WithManagingOwner(p),
		operator.WithoutKubectlAnnotations(),
	)

	if len(cpf.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = cpf.ImagePullSecrets
	}

	storageSpec := cpf.Storage
	switch {
	case storageSpec == nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})

	case storageSpec.EmptyDir != nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				EmptyDir: storageSpec.EmptyDir,
			},
		})

	case storageSpec.Ephemeral != nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				Ephemeral: storageSpec.Ephemeral,
			},
		})

	default: // storageSpec.VolumeClaimTemplate
		pvcTemplate := operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = prompkg.VolumeName(p)
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

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, cpf.Volumes...)

	if cpf.PersistentVolumeClaimRetentionPolicy != nil {
		statefulset.Spec.PersistentVolumeClaimRetentionPolicy = cpf.PersistentVolumeClaimRetentionPolicy
	}

	return statefulset, nil
}

func makeStatefulSetSpec(
	p *monitoringv1.Prometheus,
	c prompkg.Config,
	cg *prompkg.ConfigGenerator,
	shard int32,
	ruleConfigMapNames []string,
	tlsSecrets *operator.ShardedSecret,
) (*appsv1.StatefulSetSpec, error) {
	cpf := p.GetCommonPrometheusFields()

	pImagePath, err := operator.BuildImagePath(
		ptr.Deref(cpf.Image, ""),
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		operator.StringValOrDefault(p.Spec.BaseImage, c.PrometheusDefaultBaseImage),
		"v"+cg.Version().String(),
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		operator.StringValOrDefault(p.Spec.Tag, ""),
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		operator.StringValOrDefault(p.Spec.SHA, ""),
	)
	if err != nil {
		return nil, err
	}

	promArgs := buildServerArgs(cg, p)

	volumes, promVolumeMounts, err := prompkg.BuildCommonVolumes(p, tlsSecrets, true)
	if err != nil {
		return nil, err
	}

	volumes, promVolumeMounts = appendServerVolumes(p, volumes, promVolumeMounts, ruleConfigMapNames)

	configReloaderVolumeMounts := prompkg.CreateConfigReloaderVolumeMounts()

	var configReloaderWebConfigFile string

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 2.24.0.
	// With this we avoid redeploying prometheus when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	webConfigGenerator := cg.WithMinimumVersion("2.24.0")
	if webConfigGenerator.IsCompatible() {
		confArg, configVol, configMount, err := prompkg.BuildWebconfig(cpf, p)
		if err != nil {
			return nil, err
		}

		promArgs = append(promArgs, confArg)
		volumes = append(volumes, configVol...)
		promVolumeMounts = append(promVolumeMounts, configMount...)

		// To avoid breaking users deploying an old version of the config-reloader image.
		// TODO: remove the if condition after v0.72.0.
		if cpf.Web != nil {
			configReloaderWebConfigFile = confArg.Value
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, configMount...)
		}
	} else if cpf.Web != nil {
		webConfigGenerator.Warn("web.config.file")
	}

	startupProbe, readinessProbe, livenessProbe := cg.BuildProbes()

	podAnnotations, podLabels := cg.BuildPodMetadata()
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := makeSelectorLabels(p.GetObjectMeta().GetName())
	podSelectorLabels[prompkg.ShardLabelName] = fmt.Sprintf("%d", shard)

	maps.Copy(podLabels, podSelectorLabels)

	finalSelectorLabels := c.Labels.Merge(podSelectorLabels)
	finalLabels := c.Labels.Merge(podLabels)

	var additionalContainers, operatorInitContainers []v1.Container

	thanosContainer, thanosVolumes, err := createThanosContainer(p, c)
	if err != nil {
		return nil, err
	}

	if thanosContainer != nil {
		additionalContainers = append(additionalContainers, *thanosContainer)
		volumes = append(volumes, thanosVolumes...)
	}

	if compactionDisabled(p) {
		thanosBlockDuration := "2h"
		if p.Spec.Thanos != nil {
			thanosBlockDuration = operator.StringValOrDefault(string(p.Spec.Thanos.BlockDuration), thanosBlockDuration)
		}
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.max-block-duration", Value: thanosBlockDuration})
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.min-block-duration", Value: thanosBlockDuration})
	}

	// ref: https://github.com/prometheus-operator/prometheus-operator/issues/6829
	// automatically set --no-storage.tsdb.allow-overlapping-compaction when all the conditions are met:
	//   1. Prometheus >= v2.55.0
	//   2. Thanos sidecar configured for uploading blocks to object storage
	//   3. out-of-order window is > 0
	if cpf.TSDB != nil && cpf.TSDB.OutOfOrderTimeWindow != nil &&
		compactionDisabled(p) &&
		cg.WithMinimumVersion("2.55.0").IsCompatible() {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "no-storage.tsdb.allow-overlapping-compaction"})
	}

	var watchedDirectories []string

	if len(ruleConfigMapNames) != 0 {
		for _, name := range ruleConfigMapNames {
			mountPath := prompkg.RulesDir + "/" + name
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			watchedDirectories = append(watchedDirectories, mountPath)
		}
	}

	operatorInitContainers = append(operatorInitContainers,
		prompkg.BuildConfigReloader(
			p,
			c,
			true,
			configReloaderVolumeMounts,
			watchedDirectories,
			operator.Shard(shard),
		),
	)

	initContainers, err := k8s.MergePatchContainers(operatorInitContainers, cpf.InitContainers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge init containers spec: %w", err)
	}

	containerArgs, err := operator.BuildArgs(promArgs, cpf.AdditionalArgs)
	if err != nil {
		return nil, err
	}

	var envVars []v1.EnvVar
	// For higher Prometheus version its set with runtime field in configuration
	if p.Spec.Runtime != nil && p.Spec.Runtime.GoGC != nil && !cg.WithMinimumVersion("2.53.0").IsCompatible() {
		envVars = append(envVars, v1.EnvVar{Name: "GOGC", Value: fmt.Sprintf("%d", *p.Spec.Runtime.GoGC)})
	}

	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    pImagePath,
			ImagePullPolicy:          cpf.ImagePullPolicy,
			Ports:                    prompkg.MakeContainerPorts(cpf),
			Args:                     containerArgs,
			Env:                      envVars,
			VolumeMounts:             promVolumeMounts,
			StartupProbe:             startupProbe,
			LivenessProbe:            livenessProbe,
			ReadinessProbe:           readinessProbe,
			Resources:                cpf.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			SecurityContext: &v1.SecurityContext{
				ReadOnlyRootFilesystem:   ptr.To(true),
				AllowPrivilegeEscalation: ptr.To(false),
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
			},
		},
		prompkg.BuildConfigReloader(
			p,
			c,
			false,
			configReloaderVolumeMounts,
			watchedDirectories,
			operator.Shard(shard),
			operator.WebConfigFile(configReloaderWebConfigFile),
		),
	}, additionalContainers...)

	containers, err := k8s.MergePatchContainers(operatorContainers, cpf.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	// By default, podManagementPolicy is set to Parallel to mitigate rollout
	// issues in Kubernetes (see https://github.com/kubernetes/kubernetes/issues/60164).
	// This is also mentioned as one of limitations of StatefulSets:
	// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	podManagementPolicy := ptr.Deref(cpf.PodManagementPolicy, monitoringv1.ParallelPodManagement)

	spec := appsv1.StatefulSetSpec{
		ServiceName:         ptr.Deref(cpf.ServiceName, governingServiceName),
		Replicas:            cpf.Replicas,
		PodManagementPolicy: appsv1.PodManagementPolicyType(podManagementPolicy),
		UpdateStrategy:      operator.UpdateStrategyForStatefulSet(cpf.UpdateStrategy),
		MinReadySeconds:     ptr.Deref(p.Spec.MinReadySeconds, 0),
		Selector: &metav1.LabelSelector{
			MatchLabels: finalSelectorLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				ShareProcessNamespace:         prompkg.ShareProcessNamespace(p),
				Containers:                    containers,
				InitContainers:                initContainers,
				SecurityContext:               cpf.SecurityContext,
				ServiceAccountName:            cpf.ServiceAccountName,
				AutomountServiceAccountToken:  ptr.To(ptr.Deref(cpf.AutomountServiceAccountToken, true)),
				NodeSelector:                  cpf.NodeSelector,
				PriorityClassName:             cpf.PriorityClassName,
				TerminationGracePeriodSeconds: ptr.To(ptr.Deref(cpf.TerminationGracePeriodSeconds, prompkg.DefaultTerminationGracePeriodSeconds)),
				Volumes:                       volumes,
				Tolerations:                   cpf.Tolerations,
				Affinity:                      cpf.Affinity,
				TopologySpreadConstraints:     prompkg.MakeK8sTopologySpreadConstraint(finalSelectorLabels, cpf.TopologySpreadConstraints),
				HostAliases:                   operator.MakeHostAliases(cpf.HostAliases),
				HostNetwork:                   cpf.HostNetwork,
				EnableServiceLinks:            cpf.EnableServiceLinks,
				HostUsers:                     cpf.HostUsers,
			},
		},
	}

	if cpf.HostNetwork {
		spec.Template.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}
	k8s.UpdateDNSPolicy(&spec.Template.Spec, cpf.DNSPolicy)
	k8s.UpdateDNSConfig(&spec.Template.Spec, cpf.DNSConfig)

	return &spec, nil
}

// buildServerArgs returns the CLI arguments that are only valid for the Prometheus server.
func buildServerArgs(cg *prompkg.ConfigGenerator, p *monitoringv1.Prometheus) []monitoringv1.Argument {
	var (
		promArgs               = cg.BuildCommonPrometheusArgs()
		retentionTimeFlagName  = "storage.tsdb.retention.time"
		retentionTimeFlagValue = string(p.Spec.Retention)
	)

	if cg.WithMaximumVersion("2.7.0").IsCompatible() {
		retentionTimeFlagName = "storage.tsdb.retention"
		if p.Spec.Retention == "" {
			retentionTimeFlagValue = defaultRetention
		}
	} else if p.Spec.Retention == "" && p.Spec.RetentionSize == "" {
		retentionTimeFlagValue = defaultRetention
	}

	if retentionTimeFlagValue != "" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: retentionTimeFlagName, Value: retentionTimeFlagValue})
	}

	if p.Spec.RetentionSize != "" {
		retentionSizeFlag := monitoringv1.Argument{Name: "storage.tsdb.retention.size", Value: string(p.Spec.RetentionSize)}
		promArgs = cg.WithMinimumVersion("2.7.0").AppendCommandlineArgument(promArgs, retentionSizeFlag)
	}

	promArgs = append(promArgs,
		monitoringv1.Argument{Name: "storage.tsdb.path", Value: prompkg.StorageDir},
	)

	if p.Spec.EnableAdminAPI {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.enable-admin-api"})
	}

	rules := p.Spec.Rules
	if rules.Alert.ForOutageTolerance != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.for-outage-tolerance", Value: rules.Alert.ForOutageTolerance})
	}
	if rules.Alert.ForGracePeriod != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.for-grace-period", Value: rules.Alert.ForGracePeriod})
	}
	if rules.Alert.ResendDelay != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.resend-delay", Value: rules.Alert.ResendDelay})
	}

	query := p.Spec.Query
	if query != nil {
		if query.LookbackDelta != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.lookback-delta", Value: *query.LookbackDelta})
		}

		if query.MaxSamples != nil && *query.MaxSamples > 0 {
			promArgs = cg.WithMinimumVersion("2.5.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "query.max-samples", Value: fmt.Sprintf("%d", *query.MaxSamples)})
		}

		if ptr.Deref(query.MaxConcurrency, 0) > 0 {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.max-concurrency", Value: fmt.Sprintf("%d", *query.MaxConcurrency)})
		}

		if query.Timeout != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.timeout", Value: string(*query.Timeout)})
		}
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if p.Spec.AllowOverlappingBlocks {
		promArgs = cg.WithMinimumVersion("2.11.0").WithMaximumVersion("2.39.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "storage.tsdb.allow-overlapping-blocks"})
	}

	if p.Spec.WALCompression != nil {
		arg := monitoringv1.Argument{Name: "no-storage.tsdb.wal-compression"}
		if *p.Spec.WALCompression {
			arg.Name = "storage.tsdb.wal-compression"
		}
		promArgs = cg.WithMinimumVersion("2.11.0").AppendCommandlineArgument(promArgs, arg)
	}

	return promArgs
}

// appendServerVolumes returns a set of volumes to be mounted on the statefulset spec that are specific to Prometheus Server.
func appendServerVolumes(p *monitoringv1.Prometheus, volumes []v1.Volume, volumeMounts []v1.VolumeMount, ruleConfigMapNames []string) ([]v1.Volume, []v1.VolumeMount) {
	// not mount 2 emptyDir volumes at the same mountpath
	if volume, ok := queryLogFileVolume(p.Spec.QueryLogFile); ok && p.Spec.ScrapeFailureLogFile == nil {
		volumes = append(volumes, volume)
	}

	for _, name := range ruleConfigMapNames {
		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Optional: ptr.To(true),
				},
			},
		})
	}

	for _, name := range ruleConfigMapNames {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: prompkg.RulesDir + "/" + name,
			ReadOnly:  true,
		})
	}

	// Prevent mounting 2 emptyDir volumes at the same mountpath
	if vmount, ok := queryLogFileVolumeMount(p.Spec.QueryLogFile); ok && p.Spec.ScrapeFailureLogFile == nil {
		volumeMounts = append(volumeMounts, vmount)
	}

	return volumes, volumeMounts
}

func createThanosContainer(p *monitoringv1.Prometheus, c prompkg.Config) (*v1.Container, []v1.Volume, error) {
	if p.Spec.Thanos == nil {
		return nil, nil, nil
	}

	var (
		container *v1.Container
		cpf       = p.GetCommonPrometheusFields()
		thanos    = p.Spec.Thanos
	)

	thanosImage, err := operator.BuildImagePath(
		ptr.Deref(thanos.Image, ""),
		ptr.Deref(thanos.BaseImage, c.ThanosDefaultBaseImage),
		ptr.Deref(thanos.Version, operator.DefaultThanosVersion),
		ptr.Deref(thanos.Tag, ""),
		ptr.Deref(thanos.SHA, ""),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build image path: %w", err)
	}

	var grpcBindAddress, httpBindAddress string
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if thanos.ListenLocal || thanos.GRPCListenLocal {
		grpcBindAddress = "127.0.0.1"
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if thanos.ListenLocal || thanos.HTTPListenLocal {
		httpBindAddress = "127.0.0.1"
	}

	thanosArgs := []monitoringv1.Argument{
		{Name: "prometheus.url", Value: fmt.Sprintf("%s://%s:9090%s", cpf.PrometheusURIScheme(), c.LocalHost, path.Clean(cpf.WebRoutePrefix()))},
		{Name: "grpc-address", Value: fmt.Sprintf("%s:10901", grpcBindAddress)},
		{Name: "http-address", Value: fmt.Sprintf("%s:10902", httpBindAddress)},
	}

	if thanos.GRPCServerTLSConfig != nil {
		tls := thanos.GRPCServerTLSConfig
		if tls.CertFile != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "grpc-server-tls-cert", Value: tls.CertFile})
		}
		if tls.KeyFile != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "grpc-server-tls-key", Value: tls.KeyFile})
		}
		if tls.CAFile != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "grpc-server-tls-client-ca", Value: tls.CAFile})
		}
	}

	container = &v1.Container{
		Name:                     "thanos-sidecar",
		Image:                    thanosImage,
		ImagePullPolicy:          cpf.ImagePullPolicy,
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: ptr.To(false),
			ReadOnlyRootFilesystem:   ptr.To(true),
			Capabilities: &v1.Capabilities{
				Drop: []v1.Capability{"ALL"},
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
		Resources: thanos.Resources,
	}

	for _, thanosSideCarVM := range thanos.VolumeMounts {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      thanosSideCarVM.Name,
			MountPath: thanosSideCarVM.MountPath,
		})
	}

	if thanos.ObjectStorageConfig != nil || thanos.ObjectStorageConfigFile != nil {
		if thanos.ObjectStorageConfigFile != nil {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "objstore.config-file", Value: *thanos.ObjectStorageConfigFile})
		} else {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "objstore.config", Value: "$(OBJSTORE_CONFIG)"})
			container.Env = append(container.Env, v1.EnvVar{
				Name: "OBJSTORE_CONFIG",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: thanos.ObjectStorageConfig,
				},
			})
		}

		volName := prompkg.VolumeClaimName(p, cpf)
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tsdb.path", Value: prompkg.StorageDir})
		container.VolumeMounts = append(
			container.VolumeMounts,
			v1.VolumeMount{
				Name:      volName,
				MountPath: prompkg.StorageDir,
				SubPath:   prompkg.SubPathForStorage(cpf.Storage),
			},
		)
	}

	if thanos.TracingConfig != nil || len(thanos.TracingConfigFile) > 0 {
		if len(thanos.TracingConfigFile) > 0 {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tracing.config-file", Value: thanos.TracingConfigFile})
		} else {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tracing.config", Value: "$(TRACING_CONFIG)"})
			container.Env = append(container.Env, v1.EnvVar{
				Name: "TRACING_CONFIG",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: thanos.TracingConfig,
				},
			})
		}
	}

	if thanos.LogLevel != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.level", Value: thanos.LogLevel})
	} else if cpf.LogLevel != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.level", Value: cpf.LogLevel})
	}
	if thanos.LogFormat != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.format", Value: thanos.LogFormat})
	} else if cpf.LogFormat != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.format", Value: cpf.LogFormat})
	}

	if thanos.MinTime != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "min-time", Value: thanos.MinTime})
	}

	if thanos.ReadyTimeout != "" {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.ready_timeout", Value: string(thanos.ReadyTimeout)})
	}

	thanosVersion, err := semver.ParseTolerant(ptr.Deref(thanos.Version, operator.DefaultThanosVersion))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Thanos version: %w", err)
	}

	if thanos.GetConfigTimeout != "" && thanosVersion.GTE(semver.MustParse("0.29.0")) {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.get_config_timeout", Value: string(thanos.GetConfigTimeout)})
	}
	if thanos.GetConfigInterval != "" && thanosVersion.GTE(semver.MustParse("0.29.0")) {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.get_config_interval", Value: string(thanos.GetConfigInterval)})
	}

	// set prometheus.http-client-config
	// ref: https://thanos.io/tip/components/sidecar.md/#prometheus-http-client
	var volumes []v1.Volume
	if thanosVersion.GTE(semver.MustParse(thanosSupportedVersionHTTPClientFlag)) {
		thanosArgs = append(thanosArgs, monitoringv1.Argument{
			Name:  "prometheus.http-client-file",
			Value: filepath.Join(thanosConfigDir, thanosPrometheusHTTPClientConfigFileName),
		})
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      thanosPrometheusHTTPClientConfigSecretNameSuffix,
			MountPath: thanosConfigDir,
		})
		volumes = append(volumes, v1.Volume{
			Name: thanosPrometheusHTTPClientConfigSecretNameSuffix,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: thanosPrometheusHTTPClientConfigSecretName(p),
				},
			},
		})
	}

	containerArgs, err := operator.BuildArgs(thanosArgs, thanos.AdditionalArgs)
	if err != nil {
		return nil, nil, err
	}
	container.Args = append([]string{"sidecar"}, containerArgs...)

	return container, volumes, nil
}

func queryLogFileVolumeMount(queryLogFile string) (v1.VolumeMount, bool) {
	if !prompkg.UsesDefaultFileVolume(queryLogFile) {
		return v1.VolumeMount{}, false
	}

	return v1.VolumeMount{
		Name:      prompkg.DefaultLogFileVolume,
		ReadOnly:  false,
		MountPath: prompkg.DefaultLogDirectory,
	}, true
}

func queryLogFileVolume(queryLogFile string) (v1.Volume, bool) {
	if !prompkg.UsesDefaultFileVolume(queryLogFile) {
		return v1.Volume{}, false
	}

	return v1.Volume{
		Name: prompkg.DefaultLogFileVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}, true
}

func compactionDisabled(p *monitoringv1.Prometheus) bool {
	// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/
	// we have to turn off compaction of Prometheus if export to object
	// storage is configured to avoid races during uploads.
	return p.Spec.DisableCompaction ||
		(p.Spec.Thanos != nil &&
			(p.Spec.Thanos.ObjectStorageConfig != nil ||
				p.Spec.Thanos.ObjectStorageConfigFile != nil))
}
