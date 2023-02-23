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
	"path/filepath"
	"strings"

	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-kit/log"
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
	webConfigDir                    = "/etc/prometheus/web_config"
	tlsAssetsDir                    = "/etc/prometheus/certs"
	rulesDir                        = "/etc/prometheus/rules"
	secretsDir                      = "/etc/prometheus/secrets/"
	configmapsDir                   = "/etc/prometheus/configmaps/"
	configFilename                  = "prometheus.yaml.gz"
	configEnvsubstFilename          = "prometheus.env.yaml"
	sSetInputHashName               = "prometheus-operator-input-hash"
	defaultPortName                 = "web"
	defaultQueryLogDirectory        = "/var/log/prometheus"
	defaultQueryLogVolume           = "query-log-file"
)

var (
	minShards                   int32 = 1
	minReplicas                 int32 = 1
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
	cg *ConfigGenerator,
	ruleConfigMapNames []string,
	inputHash string,
	shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSet, error) {
	cpf := p.GetCommonPrometheusFields()
	objMeta := p.GetObjectMeta()
	typeMeta := p.GetTypeMeta()

	if cpf.PortName == "" {
		cpf.PortName = defaultPortName
	}

	if cpf.Replicas == nil {
		cpf.Replicas = &minReplicas
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
	labels[shardLabelName] = fmt.Sprintf("%d", shard)
	labels[prometheusNameLabelName] = objMeta.GetName()

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
			sSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[sSetInputHashName] = inputHash
	}

	if cpf.ImagePullSecrets != nil && len(cpf.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = cpf.ImagePullSecrets
	}
	storageSpec := cpf.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(objMeta.GetName()),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(objMeta.GetName()),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else if storageSpec.Ephemeral != nil {
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(objMeta.GetName()),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
			},
		})
	} else {
		pvcTemplate := operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(objMeta.GetName())
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
	cg *ConfigGenerator,
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

	if cg.version.Major != 2 {
		return nil, errors.Errorf("unsupported Prometheus major version %s", cg.version)
	}

	webRoutePrefix := "/"
	if cpf.RoutePrefix != "" {
		webRoutePrefix = cpf.RoutePrefix
	}
	promArgs := buildCommonPrometheusArgs(cpf, cg, webRoutePrefix)
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

	volumes, promVolumeMounts, err := buildCommonVolumes(p, tlsAssetSecrets)
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

		webConfig, err := webconfig.New(webConfigDir, webConfigSecretName(promName), fields)
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
	readyProbeHandler := probeHandler("/-/ready", cpf, webConfigGenerator, webRoutePrefix)
	startupProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}

	livenessProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/healthy", cpf, webConfigGenerator, webRoutePrefix),
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}

	podAnnotations, podLabels := buildPodMetadata(cpf, cg)
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := map[string]string{
		"app.kubernetes.io/name":       "prometheus",
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   promName,
		"prometheus":                   promName,
		shardLabelName:                 fmt.Sprintf("%d", shard),
		prometheusNameLabelName:        promName,
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
			thanosBlockDuration = string(thanos.BlockDuration)
		}
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.max-block-duration", Value: thanosBlockDuration})
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.min-block-duration", Value: thanosBlockDuration})
	}

	var watchedDirectories []string
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
			operator.ConfigFile(path.Join(confDir, configFilename)),
			operator.ConfigEnvsubstFile(path.Join(confOutDir, configEnvsubstFilename)),
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
			operator.ConfigFile(path.Join(confDir, configFilename)),
			operator.ConfigEnvsubstFile(path.Join(confOutDir, configEnvsubstFilename)),
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

func configSecretName(name string) string {
	return prefixedName(name)
}

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}

func webConfigSecretName(name string) string {
	return fmt.Sprintf("%s-web-config", prefixedName(name))
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

func usesDefaultQueryLogVolume(queryLogFile string) bool {
	return queryLogFile != "" && filepath.Dir(queryLogFile) == "."
}

func queryLogFileVolumeMount(queryLogFile string) (v1.VolumeMount, bool) {
	if !usesDefaultQueryLogVolume(queryLogFile) {
		return v1.VolumeMount{}, false
	}

	return v1.VolumeMount{
		Name:      defaultQueryLogVolume,
		ReadOnly:  false,
		MountPath: defaultQueryLogDirectory,
	}, true
}

func queryLogFileVolume(queryLogFile string) (v1.Volume, bool) {
	if !usesDefaultQueryLogVolume(queryLogFile) {
		return v1.Volume{}, false
	}

	return v1.Volume{
		Name: defaultQueryLogVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}, true
}

func queryLogFilePath(queryLogFile string) string {
	if !usesDefaultQueryLogVolume(queryLogFile) {
		return queryLogFile
	}

	return filepath.Join(defaultQueryLogDirectory, queryLogFile)
}

// buildCommonPrometheusArgs builds a slice of arguments that are common between Prometheus Server and Agent.
func buildCommonPrometheusArgs(cpf monitoringv1.CommonPrometheusFields, cg *ConfigGenerator, webRoutePrefix string) []monitoringv1.Argument {
	promArgs := []monitoringv1.Argument{
		{Name: "web.console.templates", Value: "/etc/prometheus/consoles"},
		{Name: "web.console.libraries", Value: "/etc/prometheus/console_libraries"},
		{Name: "config.file", Value: path.Join(confOutDir, configEnvsubstFilename)},
		{Name: "web.enable-lifecycle"},
	}

	if cpf.Web != nil {
		if cpf.Web.PageTitle != nil {
			promArgs = cg.WithMinimumVersion("2.6.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "web.page-title", Value: *cpf.Web.PageTitle})
		}

		if cpf.Web.MaxConnections != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "web.max-connections", Value: fmt.Sprintf("%d", *cpf.Web.MaxConnections)})
		}
	}

	if cpf.EnableRemoteWriteReceiver {
		promArgs = cg.WithMinimumVersion("2.33.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "web.enable-remote-write-receiver"})
	}

	if len(cpf.EnableFeatures) > 0 {
		promArgs = cg.WithMinimumVersion("2.25.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: strings.Join(cpf.EnableFeatures[:], ",")})
	}

	if cpf.ExternalURL != "" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.external-url", Value: cpf.ExternalURL})
	}

	promArgs = append(promArgs, monitoringv1.Argument{Name: "web.route-prefix", Value: webRoutePrefix})

	if cpf.LogLevel != "" && cpf.LogLevel != "info" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "log.level", Value: cpf.LogLevel})
	}

	if cpf.LogFormat != "" && cpf.LogFormat != "logfmt" {
		promArgs = cg.WithMinimumVersion("2.6.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "log.format", Value: cpf.LogFormat})
	}

	if cpf.WALCompression != nil {
		arg := monitoringv1.Argument{Name: "no-storage.tsdb.wal-compression"}
		if *cpf.WALCompression {
			arg.Name = "storage.tsdb.wal-compression"
		}
		promArgs = cg.WithMinimumVersion("2.11.0").AppendCommandlineArgument(promArgs, arg)
	}

	if cpf.ListenLocal {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.listen-address", Value: "127.0.0.1:9090"})
	}

	return promArgs
}

// appendServerArgs appends arguments that are only valid for the Prometheus server.
func appendServerArgs(
	promArgs []monitoringv1.Argument,
	cg *ConfigGenerator,
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
		monitoringv1.Argument{Name: "storage.tsdb.path", Value: storageDir},
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

// buildCommonVolumes returns a set of volumes to be mounted on statefulset spec that are common between Prometheus Server and Agent
func buildCommonVolumes(p monitoringv1.PrometheusInterface, tlsAssetSecrets []string) ([]v1.Volume, []v1.VolumeMount, error) {
	cpf := p.GetCommonPrometheusFields()
	promName := p.GetObjectMeta().GetName()

	assetsVolume := v1.Volume{
		Name: "tls-assets",
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				Sources: []v1.VolumeProjection{},
			},
		},
	}
	for _, assetShard := range tlsAssetSecrets {
		assetsVolume.Projected.Sources = append(assetsVolume.Projected.Sources,
			v1.VolumeProjection{
				Secret: &v1.SecretProjection{
					LocalObjectReference: v1.LocalObjectReference{Name: assetShard},
				},
			})
	}

	volumes := []v1.Volume{
		{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: configSecretName(promName),
				},
			},
		},
		assetsVolume,
		{
			Name: "config-out",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{
					// tmpfs is used here to avoid writing sensitive data into disk.
					Medium: v1.StorageMediumMemory,
				},
			},
		},
	}

	volName := volumeName(promName)
	if cpf.Storage != nil {
		if cpf.Storage.VolumeClaimTemplate.Name != "" {
			volName = cpf.Storage.VolumeClaimTemplate.Name
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
			SubPath:   subPathForStorage(cpf.Storage),
		},
	}

	promVolumeMounts = append(promVolumeMounts, cpf.VolumeMounts...)

	// Mount related secrets
	rn := k8sutil.NewResourceNamerWithPrefix("secret")
	for _, s := range cpf.Secrets {
		name, err := rn.DNS1123Label(s)
		if err != nil {
			return nil, nil, err
		}

		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: secretsDir + s,
		})
	}

	rn = k8sutil.NewResourceNamerWithPrefix("configmap")
	for _, c := range cpf.ConfigMaps {
		name, err := rn.DNS1123Label(c)
		if err != nil {
			return nil, nil, err
		}

		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c,
					},
				},
			},
		})
		promVolumeMounts = append(promVolumeMounts, v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: configmapsDir + c,
		})
	}

	return volumes, promVolumeMounts, nil
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
			MountPath: rulesDir + "/" + name,
		})
	}

	if vmount, ok := queryLogFileVolumeMount(queryLogFile); ok {
		volumeMounts = append(volumeMounts, vmount)
	}

	return volumes, volumeMounts
}

func probeHandler(probePath string, cpf monitoringv1.CommonPrometheusFields, webConfigGenerator *ConfigGenerator, webRoutePrefix string) v1.ProbeHandler {
	probePath = path.Clean(webRoutePrefix + probePath)
	handler := v1.ProbeHandler{}
	if cpf.ListenLocal {
		probeURL := url.URL{
			Scheme: "http",
			Host:   "localhost:9090",
			Path:   probePath,
		}
		handler.Exec = &v1.ExecAction{
			Command: []string{
				"sh",
				"-c",
				fmt.Sprintf(
					`if [ -x "$(command -v curl)" ]; then exec %s; elif [ -x "$(command -v wget)" ]; then exec %s; else exit 1; fi`,
					operator.CurlProber(probeURL.String()),
					operator.WgetProber(probeURL.String()),
				),
			},
		}
		return handler
	}

	handler.HTTPGet = &v1.HTTPGetAction{
		Path: probePath,
		Port: intstr.FromString(cpf.PortName),
	}
	if cpf.Web != nil && cpf.Web.TLSConfig != nil && webConfigGenerator.IsCompatible() {
		handler.HTTPGet.Scheme = v1.URISchemeHTTPS
	}
	return handler
}

func buildPodMetadata(cpf monitoringv1.CommonPrometheusFields, cg *ConfigGenerator) (map[string]string, map[string]string) {
	podAnnotations := map[string]string{
		"kubectl.kubernetes.io/default-container": "prometheus",
	}
	podLabels := map[string]string{
		"app.kubernetes.io/version": cg.version.String(),
	}

	if cpf.PodMetadata != nil {
		if cpf.PodMetadata.Labels != nil {
			for k, v := range cpf.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if cpf.PodMetadata.Annotations != nil {
			for k, v := range cpf.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	return podAnnotations, podLabels
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

			volName := volumeName(p.GetObjectMeta().GetName())
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tsdb.path", Value: storageDir})
			container.VolumeMounts = append(
				container.VolumeMounts,
				v1.VolumeMount{
					Name:      volName,
					MountPath: storageDir,
					SubPath:   subPathForStorage(cpf.Storage),
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

		containerArgs, err := operator.BuildArgs(thanosArgs, thanos.AdditionalArgs)
		if err != nil {
			return nil, err
		}
		container.Args = append([]string{"sidecar"}, containerArgs...)
	}

	return container, nil
}
