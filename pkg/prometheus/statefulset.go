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

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	p monitoringv1.Prometheus,
	config *operator.Config,
	ruleConfigMapNames []string,
	inputHash string,
	shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSet, error) {
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

	spec, err := makeStatefulSetSpec(logger, p, config, shard, ruleConfigMapNames, tlsAssetSecrets, parsedVersion)
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
	} else if storageSpec.Ephemeral != nil {
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(p.Name),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
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
	p monitoringv1.Prometheus,
	c *operator.Config,
	shard int32,
	ruleConfigMapNames []string,
	tlsAssetSecrets []string,
	version semver.Version,
) (*appsv1.StatefulSetSpec, error) {
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

	promArgs := []monitoringv1.Argument{
		{Name: "web.console.templates", Value: "/etc/prometheus/consoles"},
		{Name: "web.console.libraries", Value: "/etc/prometheus/console_libraries"},
	}

	// TODO(simonpasquier): log a warning message if the Prometheus version
	// doesn't support the flag (do it everywhere it needs to be, not only for
	// this block).
	retentionTimeFlag := monitoringv1.Argument{Name: "storage.tsdb.retention"}
	if version.GTE(semver.MustParse("2.7.0")) {
		retentionTimeFlag = monitoringv1.Argument{Name: "storage.tsdb.retention.time"}
		if p.Spec.Retention == "" && p.Spec.RetentionSize == "" {
			retentionTimeFlag.Value = defaultRetention
			promArgs = append(promArgs, retentionTimeFlag)
		} else {
			if p.Spec.Retention != "" {
				retentionTimeFlag.Value = string(p.Spec.Retention)
				promArgs = append(promArgs, retentionTimeFlag)
			}

			if p.Spec.RetentionSize != "" {
				retentionSizeFlag := monitoringv1.Argument{Name: "storage.tsdb.retention.size", Value: string(p.Spec.RetentionSize)}
				promArgs = append(promArgs, retentionSizeFlag)
			}
		}
	} else {
		if p.Spec.Retention == "" {
			retentionTimeFlag.Value = defaultRetention
			promArgs = append(promArgs, retentionTimeFlag)
		} else {
			retentionTimeFlag.Value = string(p.Spec.Retention)
			promArgs = append(promArgs, retentionTimeFlag)
		}
	}

	promArgs = append(promArgs,
		monitoringv1.Argument{Name: "config.file", Value: path.Join(confOutDir, configEnvsubstFilename)},
		monitoringv1.Argument{Name: "storage.tsdb.path", Value: storageDir},
		monitoringv1.Argument{Name: "web.enable-lifecycle"},
	)

	if version.Minor >= 4 {
		if p.Spec.Rules.Alert.ForOutageTolerance != "" {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "rules.alert.for-outage-tolerance", Value: p.Spec.Rules.Alert.ForOutageTolerance})
		}
		if p.Spec.Rules.Alert.ForGracePeriod != "" {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "rules.alert.for-grace-period", Value: p.Spec.Rules.Alert.ForGracePeriod})
		}
		if p.Spec.Rules.Alert.ResendDelay != "" {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "rules.alert.resend-delay", Value: p.Spec.Rules.Alert.ResendDelay})
		}
	}

	if p.Spec.Query != nil {
		if p.Spec.Query.LookbackDelta != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.lookback-delta", Value: *p.Spec.Query.LookbackDelta})
		}

		if version.Minor >= 5 {
			if p.Spec.Query.MaxSamples != nil && *p.Spec.Query.MaxSamples > 0 {
				promArgs = append(promArgs, monitoringv1.Argument{Name: "query.max-samples", Value: fmt.Sprintf("%d", *p.Spec.Query.MaxSamples)})
			}
		}

		if p.Spec.Query.MaxConcurrency != nil && *p.Spec.Query.MaxConcurrency > 1 {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.max-concurrency", Value: fmt.Sprintf("%d", *p.Spec.Query.MaxConcurrency)})
		}

		if p.Spec.Query.Timeout != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "query.timeout", Value: string(*p.Spec.Query.Timeout)})
		}
	}

	// TODO(simonpasquier): check that the Prometheus version supports the flag.
	if p.Spec.Web != nil && p.Spec.Web.PageTitle != nil {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.page-title", Value: *p.Spec.Web.PageTitle})
	}

	if p.Spec.EnableAdminAPI {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.enable-admin-api"})
	}

	if p.Spec.EnableRemoteWriteReceiver {
		if version.GTE(semver.MustParse("2.33.0")) {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "web.enable-remote-write-receiver"})
		} else {
			level.Warn(logger).Log("msg", "ignoring 'enableRemoteWriteReceiver' not supported by Prometheus", "version", version, "minimum_version", "2.33.0")
		}
	}

	if len(p.Spec.EnableFeatures) > 0 {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: strings.Join(p.Spec.EnableFeatures[:], ",")})
	}

	if p.Spec.ExternalURL != "" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.external-url", Value: p.Spec.ExternalURL})
	}

	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}
	promArgs = append(promArgs, monitoringv1.Argument{Name: "web.route-prefix", Value: webRoutePrefix})

	if p.Spec.LogLevel != "" && p.Spec.LogLevel != "info" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "log.level", Value: p.Spec.LogLevel})
	}
	if version.GTE(semver.MustParse("2.6.0")) {
		if p.Spec.LogFormat != "" && p.Spec.LogFormat != "logfmt" {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "log.format", Value: p.Spec.LogFormat})
		}
	}

	if version.GTE(semver.MustParse("2.11.0")) && p.Spec.WALCompression != nil {
		if *p.Spec.WALCompression {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.wal-compression"})
		} else {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "no-storage.tsdb.wal-compression"})
		}
	}

	if version.GTE(semver.MustParse("2.8.0")) && p.Spec.AllowOverlappingBlocks {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.allow-overlapping-blocks"})
	}

	var ports []v1.ContainerPort
	if p.Spec.ListenLocal {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.listen-address", Value: "127.0.0.1:9090"})
	} else {
		ports = []v1.ContainerPort{
			{
				Name:          p.Spec.PortName,
				ContainerPort: 9090,
				Protocol:      v1.ProtocolTCP,
			},
		}
	}

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
					SecretName: configSecretName(p.Name),
				},
			},
		},
		assetsVolume,
		{
			Name: "config-out",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	if volume, ok := queryLogFileVolume(&p); ok {
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

	if vmount, ok := queryLogFileVolumeMount(&p); ok {
		promVolumeMounts = append(promVolumeMounts, vmount)
	}

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 2.24.0.
	// With this we avoid redeploying prometheus when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	if version.GTE(semver.MustParse("2.24.0")) {
		var fields monitoringv1.WebConfigFileFields
		if p.Spec.Web != nil {
			fields = p.Spec.Web.WebConfigFileFields
		}

		webConfig, err := webconfig.New(webConfigDir, webConfigSecretName(p.Name), fields)
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
	}

	// Mount related secrets

	rn := k8sutil.NewResourceNamerWithPrefix("secret")
	for _, s := range p.Spec.Secrets {
		name, err := rn.VolumeName(s)
		if err != nil {
			return nil, err
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
	for _, c := range p.Spec.ConfigMaps {
		name, err := rn.VolumeName(c)
		if err != nil {
			return nil, err
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

	probeHandler := func(probePath string) v1.ProbeHandler {
		probePath = path.Clean(webRoutePrefix + probePath)
		handler := v1.ProbeHandler{}
		if p.Spec.ListenLocal {
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
			Port: intstr.FromString(p.Spec.PortName),
		}
		if p.Spec.Web != nil && p.Spec.Web.TLSConfig != nil && version.GTE(semver.MustParse("2.24.0")) {
			handler.HTTPGet.Scheme = v1.URISchemeHTTPS
		}
		return handler
	}

	// The /-/ready handler returns OK only after the TSDB initialization has
	// completed. The WAL replay can take a significant time for large setups
	// hence we enable the startup probe with a generous failure threshold (15
	// minutes) to ensure that the readiness probe only comes into effect once
	// Prometheus is effectively ready.
	// We don't want to use the /-/healthy handler here because it returns OK as
	// soon as the web server is started (irrespective of the WAL replay).
	startupProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/ready"),
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}

	livenessProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/healthy"),
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/ready"),
		TimeoutSeconds:   probeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{
		"app.kubernetes.io/version": version.String(),
	}
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := map[string]string{
		"app.kubernetes.io/name":       "prometheus",
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

	podAnnotations["kubectl.kubernetes.io/default-container"] = "prometheus"

	finalSelectorLabels := c.Labels.Merge(podSelectorLabels)
	finalLabels := c.Labels.Merge(podLabels)

	var additionalContainers, operatorInitContainers []v1.Container

	disableCompaction := p.Spec.DisableCompaction
	prometheusURIScheme := "http"
	if p.Spec.Web != nil && p.Spec.Web.TLSConfig != nil {
		prometheusURIScheme = "https"
	}
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

		bindAddress := "" // Listen to all available IP addresses by default
		if p.Spec.Thanos.ListenLocal {
			bindAddress = "127.0.0.1"
		}

		thanosArgs := []monitoringv1.Argument{
			{Name: "prometheus.url", Value: fmt.Sprintf("%s://%s:9090%s", prometheusURIScheme, c.LocalHost, path.Clean(webRoutePrefix))},
			{Name: "prometheus.http-client", Value: `{"tls_config": {"insecure_skip_verify":true}}`},
			{Name: "grpc-address", Value: fmt.Sprintf("%s:10901", bindAddress)},
			{Name: "http-address", Value: fmt.Sprintf("%s:10902", bindAddress)},
		}

		if p.Spec.Thanos.GRPCServerTLSConfig != nil {
			tls := p.Spec.Thanos.GRPCServerTLSConfig
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
		container := v1.Container{
			Name:                     "thanos-sidecar",
			Image:                    thanosImage,
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
			Resources: p.Spec.Thanos.Resources,
		}

		for _, thanosSideCarVM := range p.Spec.Thanos.VolumeMounts {
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      thanosSideCarVM.Name,
				MountPath: thanosSideCarVM.MountPath,
			})
		}

		if p.Spec.Thanos.ObjectStorageConfig != nil || p.Spec.Thanos.ObjectStorageConfigFile != nil {
			if p.Spec.Thanos.ObjectStorageConfigFile != nil {
				thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "objstore.config-file", Value: *p.Spec.Thanos.ObjectStorageConfigFile})
			} else {
				thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "objstore.config", Value: "$(OBJSTORE_CONFIG)"})
				container.Env = append(container.Env, v1.EnvVar{
					Name: "OBJSTORE_CONFIG",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: p.Spec.Thanos.ObjectStorageConfig,
					},
				})
			}
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tsdb.path", Value: storageDir})
			container.VolumeMounts = append(
				container.VolumeMounts,
				v1.VolumeMount{
					Name:      volName,
					MountPath: storageDir,
					SubPath:   subPathForStorage(p.Spec.Storage),
				},
			)

			// The Thanos sidecar needs the CAP_FOWNER capability because it links block files as hard link.
			container.SecurityContext.Capabilities.Add = append(container.SecurityContext.Capabilities.Add, "CAP_FOWNER")

			// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/ we have to turn off compaction of Prometheus
			// to avoid races during upload, if the uploads are configured.
			disableCompaction = true
		}

		if p.Spec.Thanos.TracingConfig != nil || len(p.Spec.Thanos.TracingConfigFile) > 0 {
			if len(p.Spec.Thanos.TracingConfigFile) > 0 {
				thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tracing.config-file", Value: p.Spec.Thanos.TracingConfigFile})
			} else {
				thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "tracing.config", Value: "$(TRACING_CONFIG)"})
				container.Env = append(container.Env, v1.EnvVar{
					Name: "TRACING_CONFIG",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: p.Spec.Thanos.TracingConfig,
					},
				})
			}
		}

		if p.Spec.Thanos.LogLevel != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.level", Value: p.Spec.Thanos.LogLevel})
		} else if p.Spec.LogLevel != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.level", Value: p.Spec.LogLevel})
		}
		if p.Spec.Thanos.LogFormat != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.format", Value: p.Spec.Thanos.LogFormat})
		} else if p.Spec.LogFormat != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "log.format", Value: p.Spec.LogFormat})
		}

		if p.Spec.Thanos.MinTime != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "min-time", Value: p.Spec.Thanos.MinTime})
		}

		if p.Spec.Thanos.ReadyTimeout != "" {
			thanosArgs = append(thanosArgs, monitoringv1.Argument{Name: "prometheus.ready_timeout", Value: string(p.Spec.Thanos.ReadyTimeout)})
		}

		containerArgs, err := buildArgs(thanosArgs, p.Spec.Thanos.AdditionalArgs)
		if err != nil {
			return nil, err
		}
		container.Args = append([]string{"sidecar"}, containerArgs...)

		additionalContainers = append(additionalContainers, container)
	}
	if disableCompaction {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.max-block-duration", Value: "2h"})
		promArgs = append(promArgs, monitoringv1.Argument{Name: "storage.tsdb.min-block-duration", Value: "2h"})
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

func usesDefaultQueryLogVolume(p *monitoringv1.Prometheus) bool {
	return p.Spec.QueryLogFile != "" && filepath.Dir(p.Spec.QueryLogFile) == "."
}

func queryLogFileVolumeMount(p *monitoringv1.Prometheus) (v1.VolumeMount, bool) {
	if !usesDefaultQueryLogVolume(p) {
		return v1.VolumeMount{}, false
	}

	return v1.VolumeMount{
		Name:      defaultQueryLogVolume,
		ReadOnly:  false,
		MountPath: defaultQueryLogDirectory,
	}, true
}

func queryLogFileVolume(p *monitoringv1.Prometheus) (v1.Volume, bool) {
	if !usesDefaultQueryLogVolume(p) {
		return v1.Volume{}, false
	}

	return v1.Volume{
		Name: defaultQueryLogVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}, true
}

func queryLogFilePath(p *monitoringv1.Prometheus) string {
	if !usesDefaultQueryLogVolume(p) {
		return p.Spec.QueryLogFile
	}

	return filepath.Join(defaultQueryLogDirectory, p.Spec.QueryLogFile)
}

func intersection(a, b []string) (i []string) {
	m := make(map[string]struct{})

	for _, item := range a {
		m[item] = struct{}{}
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			i = append(i, item)
		}

		negatedItem := strings.TrimPrefix(item, "no-")
		if item == negatedItem {
			negatedItem = fmt.Sprintf("no-%s", item)
		}

		if _, ok := m[negatedItem]; ok {
			i = append(i, item)
		}
	}
	return i
}

func extractArgKeys(args []monitoringv1.Argument) []string {
	var k []string
	for _, arg := range args {
		key := arg.Name
		k = append(k, key)
	}

	return k
}

func buildArgs(args []monitoringv1.Argument, additionalArgs []monitoringv1.Argument) ([]string, error) {
	var containerArgs []string

	argKeys := extractArgKeys(args)
	additionalArgKeys := extractArgKeys(additionalArgs)

	i := intersection(argKeys, additionalArgKeys)
	if len(i) > 0 {
		return nil, errors.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(i, ","))
	}

	args = append(args, additionalArgs...)

	for _, arg := range args {
		if arg.Value != "" {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s=%s", arg.Name, arg.Value))
		} else {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s", arg.Name))

		}
	}

	return containerArgs, nil
}
