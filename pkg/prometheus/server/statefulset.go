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

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	defaultRetention      = "24h"
	defaultQueryLogVolume = "query-log-file"
	prometheusMode        = "server"
	governingServiceName  = "prometheus-operated"
)

// TODO(ArthurSens): generalize it enough to be used by both server and agent.
func makeStatefulSetService(p *monitoringv1.Prometheus, config operator.Config) *v1.Service {
	p = p.DeepCopy()

	if p.Spec.PortName == "" {
		p.Spec.PortName = prompkg.DefaultPortName
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

func makeStatefulSet(
	logger log.Logger,
	name string,
	p monitoringv1.PrometheusInterface,
	baseImage, tag, sha string,
	retention monitoringv1.Duration,
	retentionSize monitoringv1.ByteSize,
	rules monitoringv1.Rules,
	query *monitoringv1.QuerySpec,
	allowOverlappingBlocks bool,
	enableAdminAPI bool,
	queryLogFile string,
	thanos *monitoringv1.ThanosSpec,
	disableCompaction bool,
	config *operator.Config,
	cg *prompkg.ConfigGenerator,
	ruleConfigMapNames []string,
	inputHash string,
	shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSet, error) {
	cpf := p.GetCommonPrometheusFields()
	objMeta := p.GetObjectMeta()
	typeMeta := p.GetTypeMeta()

	if cpf.PortName == "" {
		cpf.PortName = prompkg.DefaultPortName
	}

	if cpf.Replicas == nil {
		cpf.Replicas = &prompkg.MinReplicas
	}
	intZero := int32(0)
	if cpf.Replicas != nil && *cpf.Replicas < 0 {
		cpf.Replicas = &intZero
	}

	// We need to re-set the common fields because cpf is only a copy of the original object.
	// We set some defaults if some fields are not present, and we want those fields set in the original Prometheus object before building the StatefulSetSpec.
	p.SetCommonPrometheusFields(cpf)
	spec, err := makeStatefulSetSpec(logger, baseImage, tag, sha, retention, retentionSize, rules, query, allowOverlappingBlocks, enableAdminAPI, queryLogFile, thanos, disableCompaction, p, config, cg, shard, ruleConfigMapNames, tlsAssetSecrets)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	boolTrue := true
	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	annotations := make(map[string]string)
	for key, value := range objMeta.GetAnnotations() {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	labels := make(map[string]string)
	for key, value := range objMeta.GetLabels() {
		labels[key] = value
	}
	labels[prompkg.ShardLabelName] = fmt.Sprintf("%d", shard)
	labels[prompkg.PrometheusNameLabelName] = objMeta.GetName()
	labels[prompkg.PrometheusModeLabeLName] = prometheusMode

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      config.Labels.Merge(labels),
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         typeMeta.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               typeMeta.Kind,
					Name:               objMeta.GetName(),
					UID:                objMeta.GetUID(),
				},
			},
		},
		Spec: *spec,
	}

	if statefulset.ObjectMeta.Annotations == nil {
		statefulset.ObjectMeta.Annotations = map[string]string{
			prompkg.SSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[prompkg.SSetInputHashName] = inputHash
	}

	if cpf.ImagePullSecrets != nil && len(cpf.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = cpf.ImagePullSecrets
	}
	storageSpec := cpf.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else if storageSpec.Ephemeral != nil {
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: prompkg.VolumeName(p),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
			},
		})
	} else {
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

	if cpf.HostNetwork {
		statefulset.Spec.Template.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}

	return statefulset, nil
}

func makeStatefulSetSpec(
	logger log.Logger,
	baseImage, tag, sha string,
	retention monitoringv1.Duration,
	retentionSize monitoringv1.ByteSize,
	rules monitoringv1.Rules,
	query *monitoringv1.QuerySpec,
	allowOverlappingBlocks bool,
	enableAdminAPI bool,
	queryLogFile string,
	thanos *monitoringv1.ThanosSpec,
	disableCompaction bool,
	p monitoringv1.PrometheusInterface,
	c *operator.Config,
	cg *prompkg.ConfigGenerator,
	shard int32,
	ruleConfigMapNames []string,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)
	cpf := p.GetCommonPrometheusFields()
	promName := p.GetObjectMeta().GetName()

	pImagePath, err := operator.BuildImagePath(
		operator.StringPtrValOrDefault(cpf.Image, ""),
		operator.StringValOrDefault(baseImage, c.PrometheusDefaultBaseImage),
		operator.StringValOrDefault(cpf.Version, operator.DefaultPrometheusVersion),
		operator.StringValOrDefault(tag, ""),
		operator.StringValOrDefault(sha, ""),
	)
	if err != nil {
		return nil, err
	}

	webRoutePrefix := "/"
	if cpf.RoutePrefix != "" {
		webRoutePrefix = cpf.RoutePrefix
	}
	promArgs := prompkg.BuildCommonPrometheusArgs(cpf, cg, webRoutePrefix)
	promArgs = appendServerArgs(promArgs, cg, retention, retentionSize, rules, query, allowOverlappingBlocks, enableAdminAPI)

	var ports []v1.ContainerPort
	if !cpf.ListenLocal {
		ports = []v1.ContainerPort{
			{
				Name:          cpf.PortName,
				ContainerPort: 9090,
				Protocol:      v1.ProtocolTCP,
			},
		}
	}

	volumes, promVolumeMounts, err := prompkg.BuildCommonVolumes(p, tlsAssetSecrets)
	if err != nil {
		return nil, err
	}
	volumes, promVolumeMounts = appendServerVolumes(volumes, promVolumeMounts, queryLogFile, ruleConfigMapNames)

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 2.24.0.
	// With this we avoid redeploying prometheus when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	webConfigGenerator := cg.WithMinimumVersion("2.24.0")
	if webConfigGenerator.IsCompatible() {
		var fields monitoringv1.WebConfigFileFields
		if cpf.Web != nil {
			fields = cpf.Web.WebConfigFileFields
		}

		webConfig, err := webconfig.New(prompkg.WebConfigDir, prompkg.WebConfigSecretName(p), fields)
		if err != nil {
			return nil, err
		}

		confArg, configVol, configMount, err := webConfig.GetMountParameters()
		if err != nil {
			return nil, err
		}
		promArgs = append(promArgs, confArg)
		volumes = append(volumes, configVol...)
		promVolumeMounts = append(promVolumeMounts, configMount...)
	} else if cpf.Web != nil {
		webConfigGenerator.Warn("web.config.file")
	}

	// The /-/ready handler returns OK only after the TSDB initialization has
	// completed. The WAL replay can take a significant time for large setups
	// hence we enable the startup probe with a generous failure threshold (15
	// minutes) to ensure that the readiness probe only comes into effect once
	// Prometheus is effectively ready.
	// We don't want to use the /-/healthy handler here because it returns OK as
	// soon as the web server is started (irrespective of the WAL replay).
	readyProbeHandler := prompkg.ProbeHandler("/-/ready", cpf, webConfigGenerator, webRoutePrefix)
	startupProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   prompkg.ProbeTimeoutSeconds,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   prompkg.ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}

	livenessProbe := &v1.Probe{
		ProbeHandler:     prompkg.ProbeHandler("/-/healthy", cpf, webConfigGenerator, webRoutePrefix),
		TimeoutSeconds:   prompkg.ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}

	podAnnotations, podLabels := prompkg.BuildPodMetadata(cpf, cg)
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := map[string]string{
		"app.kubernetes.io/name":        "prometheus",
		"app.kubernetes.io/managed-by":  "prometheus-operator",
		"app.kubernetes.io/instance":    promName,
		"prometheus":                    promName,
		prompkg.ShardLabelName:          fmt.Sprintf("%d", shard),
		prompkg.PrometheusNameLabelName: promName,
	}

	for k, v := range podSelectorLabels {
		podLabels[k] = v
	}

	finalSelectorLabels := c.Labels.Merge(podSelectorLabels)
	finalLabels := c.Labels.Merge(podLabels)

	var additionalContainers, operatorInitContainers []v1.Container

	prometheusURIScheme := "http"
	if cpf.Web != nil && cpf.Web.TLSConfig != nil {
		prometheusURIScheme = "https"
	}

	thanosContainer, err := createThanosContainer(&disableCompaction, p, thanos, c, prometheusURIScheme, webRoutePrefix)
	if err != nil {
		return nil, err
	}
	if thanosContainer != nil {
		additionalContainers = append(additionalContainers, *thanosContainer)
	}

	if disableCompaction {
		thanosBlockDuration := "2h"
		if thanos != nil {
			thanosBlockDuration = operator.StringValOrDefault(string(thanos.BlockDuration), thanosBlockDuration)
		}
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.max-block-duration", Value: thanosBlockDuration})
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.min-block-duration", Value: thanosBlockDuration})
	}

	var watchedDirectories []string
	configReloaderVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			MountPath: prompkg.ConfDir,
		},
		{
			Name:      "config-out",
			MountPath: prompkg.ConfOutDir,
		},
	}

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

	var minReadySeconds int32
	if cpf.MinReadySeconds != nil {
		minReadySeconds = int32(*cpf.MinReadySeconds)
	}

	operatorInitContainers = append(operatorInitContainers,
		operator.CreateConfigReloader(
			"init-config-reloader",
			operator.ReloaderConfig(c.ReloaderConfig),
			operator.ReloaderRunOnce(),
			operator.LogFormat(cpf.LogFormat),
			operator.LogLevel(cpf.LogLevel),
			operator.VolumeMounts(configReloaderVolumeMounts),
			operator.ConfigFile(path.Join(prompkg.ConfDir, prompkg.ConfigFilename)),
			operator.ConfigEnvsubstFile(path.Join(prompkg.ConfOutDir, prompkg.ConfigEnvsubstFilename)),
			operator.WatchedDirectories(watchedDirectories),
			operator.Shard(shard),
			operator.ImagePullPolicy(cpf.ImagePullPolicy),
		),
	)

	initContainers, err := k8sutil.MergePatchContainers(operatorInitContainers, cpf.InitContainers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge init containers spec")
	}

	containerArgs, err := operator.BuildArgs(promArgs, cpf.AdditionalArgs)

	if err != nil {
		return nil, err
	}

	boolFalse := false
	boolTrue := true
	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    pImagePath,
			ImagePullPolicy:          cpf.ImagePullPolicy,
			Ports:                    ports,
			Args:                     containerArgs,
			VolumeMounts:             promVolumeMounts,
			StartupProbe:             startupProbe,
			LivenessProbe:            livenessProbe,
			ReadinessProbe:           readinessProbe,
			Resources:                cpf.Resources,
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
			operator.ReloaderConfig(c.ReloaderConfig),
			operator.ReloaderURL(url.URL{
				Scheme: prometheusURIScheme,
				Host:   c.LocalHost + ":9090",
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			}),
			operator.ListenLocal(cpf.ListenLocal),
			operator.LocalHost(c.LocalHost),
			operator.LogFormat(cpf.LogFormat),
			operator.LogLevel(cpf.LogLevel),
			operator.ConfigFile(path.Join(prompkg.ConfDir, prompkg.ConfigFilename)),
			operator.ConfigEnvsubstFile(path.Join(prompkg.ConfOutDir, prompkg.ConfigEnvsubstFilename)),
			operator.WatchedDirectories(watchedDirectories), operator.VolumeMounts(configReloaderVolumeMounts),
			operator.Shard(shard),
			operator.ImagePullPolicy(cpf.ImagePullPolicy),
		),
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, cpf.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		ServiceName:         governingServiceName,
		Replicas:            cpf.Replicas,
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
				SecurityContext:               cpf.SecurityContext,
				ServiceAccountName:            cpf.ServiceAccountName,
				AutomountServiceAccountToken:  &boolTrue,
				NodeSelector:                  cpf.NodeSelector,
				PriorityClassName:             cpf.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   cpf.Tolerations,
				Affinity:                      cpf.Affinity,
				TopologySpreadConstraints:     cpf.TopologySpreadConstraints,
				HostAliases:                   operator.MakeHostAliases(cpf.HostAliases),
				HostNetwork:                   cpf.HostNetwork,
			},
		},
	}, nil
}

// appendServerArgs appends arguments that are only valid for the Prometheus server.
func appendServerArgs(
	promArgs []monitoringv1.Argument,
	cg *prompkg.ConfigGenerator,
	retention monitoringv1.Duration,
	retentionSize monitoringv1.ByteSize,
	rules monitoringv1.Rules,
	query *monitoringv1.QuerySpec,
	allowOverlappingBlocks, enableAdminAPI bool) []monitoringv1.Argument {
	var (
		retentionTimeFlagName  = "storage.tsdb.retention.time"
		retentionTimeFlagValue = string(retention)
	)
	if cg.WithMaximumVersion("2.7.0").IsCompatible() {
		retentionTimeFlagName = "storage.tsdb.retention"
		if retention == "" {
			retentionTimeFlagValue = defaultRetention
		}
	} else if retention == "" && retentionSize == "" {
		retentionTimeFlagValue = defaultRetention
	}

	if retentionTimeFlagValue != "" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: retentionTimeFlagName, Value: retentionTimeFlagValue})
	}
	if retentionSize != "" {
		retentionSizeFlag := monitoringv1.Argument{Name: "storage.tsdb.retention.size", Value: string(retentionSize)}
		promArgs = cg.WithMinimumVersion("2.7.0").AppendCommandlineArgument(promArgs, retentionSizeFlag)
	}

	promArgs = append(promArgs,
		monitoringv1.Argument{Name: "storage.tsdb.path", Value: prompkg.StorageDir},
	)

	if enableAdminAPI {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.enable-admin-api"})
	}

	if rules.Alert.ForOutageTolerance != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.for-outage-tolerance", Value: rules.Alert.ForOutageTolerance})
	}
	if rules.Alert.ForGracePeriod != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.for-grace-period", Value: rules.Alert.ForGracePeriod})
	}
	if rules.Alert.ResendDelay != "" {
		promArgs = cg.WithMinimumVersion("2.4.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "rules.alert.resend-delay", Value: rules.Alert.ResendDelay})
	}

	if query != nil {
		if query.LookbackDelta != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.lookback-delta", Value: *query.LookbackDelta})
		}

		if query.MaxSamples != nil && *query.MaxSamples > 0 {
			promArgs = cg.WithMinimumVersion("2.5.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "query.max-samples", Value: fmt.Sprintf("%d", *query.MaxSamples)})
		}

		if query.MaxConcurrency != nil && *query.MaxConcurrency > 1 {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.max-concurrency", Value: fmt.Sprintf("%d", *query.MaxConcurrency)})
		}

		if query.Timeout != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.timeout", Value: string(*query.Timeout)})
		}
	}

	if allowOverlappingBlocks {
		promArgs = cg.WithMinimumVersion("2.11.0").WithMaximumVersion("2.39.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "storage.tsdb.allow-overlapping-blocks"})
	}
	return promArgs
}

// appendServerVolumes returns a set of volumes to be mounted on the statefulset spec that are specific to Prometheus Server
func appendServerVolumes(volumes []v1.Volume, volumeMounts []v1.VolumeMount, queryLogFile string, ruleConfigMapNames []string) ([]v1.Volume, []v1.VolumeMount) {
	if volume, ok := queryLogFileVolume(queryLogFile); ok {
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
				},
			},
		})
	}

	for _, name := range ruleConfigMapNames {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: prompkg.RulesDir + "/" + name,
		})
	}

	if vmount, ok := queryLogFileVolumeMount(queryLogFile); ok {
		volumeMounts = append(volumeMounts, vmount)
	}

	return volumes, volumeMounts
}

func createThanosContainer(
	disableCompaction *bool,
	p monitoringv1.PrometheusInterface,
	thanos *monitoringv1.ThanosSpec,
	c *operator.Config,
	prometheusURIScheme, webRoutePrefix string) (*v1.Container, error) {

	var container *v1.Container
	cpf := p.GetCommonPrometheusFields()

	if thanos != nil {
		thanosImage, err := operator.BuildImagePath(
			operator.StringPtrValOrDefault(thanos.Image, ""),
			operator.StringPtrValOrDefault(thanos.BaseImage, c.ThanosDefaultBaseImage),
			operator.StringPtrValOrDefault(thanos.Version, operator.DefaultThanosVersion),
			operator.StringPtrValOrDefault(thanos.Tag, ""),
			operator.StringPtrValOrDefault(thanos.SHA, ""),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build image path")
		}

		var grpcBindAddress, httpBindAddress string
		if thanos.ListenLocal || thanos.GRPCListenLocal {
			grpcBindAddress = "127.0.0.1"
		}

		if thanos.ListenLocal || thanos.HTTPListenLocal {
			httpBindAddress = "127.0.0.1"
		}

		thanosArgs := []monitoringv1.Argument{
			{Name: "prometheus.url", Value: fmt.Sprintf("%s://%s:9090%s", prometheusURIScheme, c.LocalHost, path.Clean(webRoutePrefix))},
			{Name: "prometheus.http-client", Value: `{"tls_config": {"insecure_skip_verify":true}}`},
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

		boolFalse := false
		boolTrue := true
		container = &v1.Container{
			Name:                     "thanos-sidecar",
			Image:                    thanosImage,
			ImagePullPolicy:          cpf.ImagePullPolicy,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			SecurityContext: &v1.SecurityContext{
				AllowPrivilegeEscalation: &boolFalse,
				ReadOnlyRootFilesystem:   &boolTrue,
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

			volName := prompkg.VolumeName(p)
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tsdb.path", Value: prompkg.StorageDir})
			container.VolumeMounts = append(
				container.VolumeMounts,
				v1.VolumeMount{
					Name:      volName,
					MountPath: prompkg.StorageDir,
					SubPath:   prompkg.SubPathForStorage(cpf.Storage),
				},
			)

			// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/ we have to turn off compaction of Prometheus
			// to avoid races during upload, if the uploads are configured.
			*disableCompaction = true
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

		thanosVersion, err := semver.ParseTolerant(operator.StringPtrValOrDefault(thanos.Version, operator.DefaultThanosVersion))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Thanos version")
		}

		if thanos.GetConfigTimeout != "" && thanosVersion.GTE(semver.MustParse("0.29.0")) {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.get_config_timeout", Value: string(thanos.GetConfigTimeout)})
		}
		if thanos.GetConfigInterval != "" && thanosVersion.GTE(semver.MustParse("0.29.0")) {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.get_config_interval", Value: string(thanos.GetConfigInterval)})
		}

		containerArgs, err := operator.BuildArgs(thanosArgs, thanos.AdditionalArgs)
		if err != nil {
			return nil, err
		}
		container.Args = append([]string{"sidecar"}, containerArgs...)
	}

	return container, nil
}

func queryLogFileVolumeMount(queryLogFile string) (v1.VolumeMount, bool) {
	if !prompkg.UsesDefaultQueryLogVolume(queryLogFile) {
		return v1.VolumeMount{}, false
	}

	return v1.VolumeMount{
		Name:      defaultQueryLogVolume,
		ReadOnly:  false,
		MountPath: prompkg.DefaultQueryLogDirectory,
	}, true
}

func queryLogFileVolume(queryLogFile string) (v1.Volume, bool) {
	if !prompkg.UsesDefaultQueryLogVolume(queryLogFile) {
		return v1.Volume{}, false
	}

	return v1.Volume{
		Name: defaultQueryLogVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}, true
}
