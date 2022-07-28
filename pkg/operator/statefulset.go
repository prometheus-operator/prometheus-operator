// Copyright 2022 The prometheus-operator Authors
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

package operator

import (
	"fmt"
	"net/url"
	"path"
	//"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

var (
	boolFalse bool = false
	boolTrue  bool = true

	int32Zero int32 = 0
	//minShards             int32 = 1
	minReplicas           int32 = 1
	defaultMaxConcurrency int32 = 20
	probeTimeoutSeconds   int32 = 3
)

const (
	DefaultRetention                = "24h"
	DefaultMemoryRequestValue       = "2Gi"
	DefaultPrometheusContainerPort  = 9090
	DefaultThanosGRPCPort           = 10901
	DefaultThanosHTTPPort           = 10902
	DefaultPrometheusQueryLogVolume = "query-log-file"
	DefaultPrometheusQueryLogDir    = "/var/log/prometheus"
	DefaultPrometheusPortName       = "web"

	GoverningServiceName     = "prometheus-operated"
	StatefulSetInputHashName = "prometheus-operator-input-hash"

	PrometheusConfDir              = "/etc/prometheus/config"
	PrometheusConfOutDir           = "/etc/prometheus/config_out"
	PrometheusConfEnvSubstFilename = "prometheus.env.yaml"
	PrometheusConfFilename         = "prometheus.yaml.gz"
	PrometheusStorageDir           = "/prometheus"
	PrometheusTLSAssetsDir         = "/etc/prometheus/certs"
	PrometheusRulesDir             = "/etc/prometheus/rules"
	PrometheusSecretsDir           = "/etc/prometheus/secrets/"
	PrometheusConfigmapsDir        = "/etc/prometheus/configmaps/"

	ThanosObjStoreEnvVar    = "OBJSTORE_CONFIG"
	ThanosTraceConfigEnvVar = "TRACING_CONFIG"

	prometheusNameLabelName = "prometheus.io/name"
	shardLabelName          = "prometheus.io/shard"

	WebConfigDir           = "/etc/prometheus/web_config"
	WebConfigFilename      = "web-config.yaml"
	WebConsoleLibraryDir   = "/etc/prometheus/console_libraries"
	WebConsoleTemplatesDir = "/etc/prometheus/consoles"
)

// MakePrometheusCommandArgs returns slice of Prometheus command arguments for either Prometheus server or Proemtheus agent
func MakePrometheusCommandArgs(pt PrometheusType) (out, warns []string, err error) {
	warns = []string{}

	args, warns, err := pt.MakeCommandArgs()
	if err != nil {
		return out, warns, err
	}

	args, err = processAdditionalArgs(args, pt.GetAdditionalArgs())
	if err != nil {
		return out, warns, err
	}

	out = transformArgs(args)
	return out, warns, nil
}

func MakeThanosCommandArgs(pt PrometheusType, c *Config) (out, warns []string, err error) {
	promArgs, warns, err := pt.MakeCommandArgs()
	if err != nil {
		return out, warns, err
	}

	thanos := pt.GetThanosSpec()
	uriScheme := getURIScheme(pt.GetWebSpec())
	prefix := getWebRoutePrefix(promArgs)
	bindAddress := "" // Listen to all available IP addresses by default
	if thanos.ListenLocal {
		bindAddress = "127.0.0.1"
	}

	args := map[string]string{
		"prometheus.url": fmt.Sprintf("%s://%s:9090%s", uriScheme, c.LocalHost, path.Clean(prefix)),
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

	args, err = processAdditionalArgs(args, thanos.AdditionalArgs)
	if err != nil {
		return out, warns, err
	}

	out = append([]string{"sidecar"}, transformArgs(args)...)
	return out, warns, err
}

// --------------------- *StatefulSet* helpers ----------------------
func subPathForStorage(nc *Nomenclator, s *promv1.StorageSpec) string {
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if s == nil || s.DisableMountSubPath {
		return ""
	}

	return nc.VolumeName()
}

func getURIScheme(webSpec *promv1.PrometheusWebSpec) string {
	URIScheme := "http"
	if webSpec != nil && webSpec.TLSConfig != nil {
		URIScheme = "https"
	}
	return URIScheme
}

func getWebRoutePrefix(args map[string]string) string {
	if prefix, ok := args["web.route-prefix"]; ok {
		return prefix
	}
	return "/"
}

func processAdditionalArgs(opArgs map[string]string, addArgs []promv1.Argument) (map[string]string, error) {
	invalid := []string{}
	for _, addArg := range addArgs {
		// test regular arg occurence
		found := false
		if _, ok := opArgs[addArg.Name]; ok {
			found = true
		}
		// test negated arg occurence
		if !strings.HasPrefix(addArg.Name, "no-") {
			if _, ok := opArgs[fmt.Sprintf("no-%s", addArg.Name)]; ok {
				found = true
			}
		} else {
			// test removed negation occurence
			if _, ok := opArgs[addArg.Name[3:]]; ok {
				found = true
			}
		}

		if found {
			invalid = append(invalid, addArg.Name)
		} else {
			opArgs[addArg.Name] = addArg.Value
		}
	}

	if len(invalid) > 0 {
		return opArgs, errors.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(invalid, ","))
	}
	return opArgs, nil
}

func transformArgs(args map[string]string) []string {
	out := make([]string, 0, len(args))
	for key, value := range args {
		item := fmt.Sprintf("--%s", key)
		if value != "" {
			item = fmt.Sprintf("%s=%s", item, value)
		}
		out = append(out, item)
	}
	return out
}

// ------------------------------------------------------------------

func MakePrometheusStatefulsetSpec(pt PrometheusType, logger log.Logger,
	c *Config, shard int32, ruleConfigMapNames []string,
	tlsAssetSecrets []string) (*appsv1.StatefulSetSpec, error) {
	// naming objects
	nc := pt.GetNomenclator()

	// version checks
	version, err := pt.GetVersion(DefaultPrometheusVersion)
	if err != nil {
		return nil, err
	}
	switch pt.(type) {
	case PrometheusServer:
		if version.Major != 2 {
			return nil, errors.Errorf("unsupported Prometheus major version %s", version)
		}
	//case PrometheusAgent:
	//	if !version.GTE(semver.MustParse("2.32.0")) {
	//		return nil, errors.Errorf("unsupported Prometheus version %s", version)
	//	}
	default:
		return nil, errors.Errorf("unsupported object type")
	}

	baseImage, tag, sha := pt.GetDeprecatedImageInfo()
	specVersion, err := pt.GetVersion("")
	if err != nil {
		return nil, err
	}
	imageVersion := ""
	if specVersion != nil {
		specVersion.String()
	}
	imagePath, err := BuildImagePath(
		StringPtrValOrDefault(pt.GetImage(), ""),
		StringValOrDefault(baseImage, c.PrometheusDefaultBaseImage),
		imageVersion, tag, sha)
	if err != nil {
		return nil, err
	}

	query := pt.GetQuery()
	if query != nil && query.MaxConcurrency != nil && *query.MaxConcurrency < 1 {
		query.MaxConcurrency = &defaultMaxConcurrency
	}

	var ports []v1.ContainerPort
	if portName := pt.ListensOn(); portName != "localhost" {
		ports = []v1.ContainerPort{
			{
				Name:          portName,
				ContainerPort: DefaultPrometheusContainerPort,
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
					SecretName: nc.ConfigSecretName(),
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

	if pt.UsesDefaultQueryLogVolume() {
		volumes = append(volumes, v1.Volume{
			Name: DefaultPrometheusQueryLogVolume,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
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

	storage := pt.GetStorageSpec()
	volName := nc.VolumeName()
	if storage != nil {
		if storage.VolumeClaimTemplate.Name != "" {
			volName = storage.VolumeClaimTemplate.Name
		}
	}
	mounts := []v1.VolumeMount{
		{
			Name:      "config-out",
			ReadOnly:  true,
			MountPath: PrometheusConfOutDir,
		},
		{
			Name:      "tls-assets",
			ReadOnly:  true,
			MountPath: PrometheusTLSAssetsDir,
		},
		{
			Name:      volName,
			MountPath: PrometheusStorageDir,
			SubPath:   subPathForStorage(nc, storage),
		},
	}

	mounts = append(mounts, pt.GetVolumeMounts()...)
	for _, name := range ruleConfigMapNames {
		mounts = append(mounts, v1.VolumeMount{
			Name:      name,
			MountPath: path.Join(PrometheusRulesDir, name),
		})
	}

	if pt.UsesDefaultQueryLogVolume() {
		mounts = append(mounts, v1.VolumeMount{
			Name:      DefaultPrometheusQueryLogVolume,
			ReadOnly:  false,
			MountPath: DefaultPrometheusQueryLogDir,
		})
	}

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 2.24.0.
	// With this we avoid redeploying prometheus when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	webSpec := pt.GetWebSpec()
	var fields promv1.WebConfigFileFields
	if webSpec != nil {
		fields = webSpec.WebConfigFileFields
	}

	if version.GTE(semver.MustParse("2.24.0")) {
		webConfig, err := webconfig.New(WebConfigDir, nc.WebConfigSecretName(), fields)
		if err != nil {
			return nil, err
		}

		configVol, configMount := webConfig.GetMountParameters()
		volumes = append(volumes, configVol...)
		mounts = append(mounts, configMount...)
	}

	// Mount related secrets
	for _, s := range pt.GetSecrets() {
		volumes = append(volumes, v1.Volume{
			Name: k8sutil.SanitizeVolumeName("secret-" + s),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		mounts = append(mounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("secret-" + s),
			ReadOnly:  true,
			MountPath: path.Join(PrometheusSecretsDir, s),
		})
	}

	for _, c := range pt.GetConfigMaps() {
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
		mounts = append(mounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("configmap-" + c),
			ReadOnly:  true,
			MountPath: path.Join(PrometheusConfigmapsDir, c),
		})
	}

	argMap, _, err := pt.MakeCommandArgs()
	webRoutePrefix := getWebRoutePrefix(argMap)

	probeHandler := func(probePath string) v1.ProbeHandler {
		probePath = path.Clean(webRoutePrefix + probePath)
		handler := v1.ProbeHandler{}
		if pt.ListensOn() == "localhost" {
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
						CurlProber(probeURL.String()),
						WgetProber(probeURL.String()),
					),
				},
			}
			return handler
		}

		handler.HTTPGet = &v1.HTTPGetAction{
			Path: probePath,
			Port: intstr.FromString(pt.ListensOn()),
		}
		if webSpec != nil && webSpec.TLSConfig != nil && version.GTE(semver.MustParse("2.24.0")) {
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
		"app.kubernetes.io/instance":   nc.BaseName(),
		"prometheus":                   nc.BaseName(),
		shardLabelName:                 fmt.Sprintf("%d", shard),
		prometheusNameLabelName:        nc.BaseName(),
	}
	metadata := pt.GetPodMetadata()
	if metadata != nil {
		if metadata.Labels != nil {
			for k, v := range metadata.Labels {
				podLabels[k] = v
			}
		}
		if metadata.Annotations != nil {
			for k, v := range metadata.Annotations {
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

	var additionalContainers []v1.Container
	thanos := pt.GetThanosSpec()
	if thanos != nil {
		thanosImage, err := BuildImagePath(
			StringPtrValOrDefault(thanos.Image, ""),
			StringPtrValOrDefault(thanos.BaseImage, c.ThanosDefaultBaseImage),
			StringPtrValOrDefault(thanos.Version, DefaultThanosVersion),
			StringPtrValOrDefault(thanos.Tag, ""),
			StringPtrValOrDefault(thanos.SHA, ""),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build Thanos image path")
		}

		thanosArgs, warns, err := MakeThanosCommandArgs(pt, c)
		if err != nil {
			return nil, err
		}
		for _, msg := range warns {
			level.Warn(logger).Log("msg", msg)
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
				{
					Name:          "http",
					ContainerPort: DefaultThanosHTTPPort,
				},
				{
					Name:          "grpc",
					ContainerPort: DefaultThanosGRPCPort,
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
					MountPath: PrometheusStorageDir,
					SubPath:   subPathForStorage(nc, storage),
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
		additionalContainers = append(additionalContainers, container)
	}

	var watchedDirectories []string
	configReloaderVolumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			MountPath: PrometheusConfDir,
		},
		{
			Name:      "config-out",
			MountPath: PrometheusConfOutDir,
		},
	}

	for _, name := range ruleConfigMapNames {
		mountPath := path.Join(PrometheusRulesDir, name)
		configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: mountPath,
		})
		watchedDirectories = append(watchedDirectories, mountPath)
	}

	var minReadySeconds int32
	if pt.GetMinReadySeconds() != nil {
		minReadySeconds = int32(*pt.GetMinReadySeconds())
	}

	logLevel, logFormat := pt.GetLoggerInfo()
	operatorInitContainers := []v1.Container{
		CreateConfigReloader(
			"init-config-reloader",
			ReloaderResources(c.ReloaderConfig),
			ReloaderRunOnce(),
			LogFormat(logFormat),
			LogLevel(logLevel),
			VolumeMounts(configReloaderVolumeMounts),
			ConfigFile(path.Join(PrometheusConfDir, PrometheusConfFilename)),
			ConfigEnvsubstFile(path.Join(PrometheusConfOutDir, PrometheusConfEnvSubstFilename)),
			WatchedDirectories(watchedDirectories),
			Shard(shard),
		),
	}
	initContainers, err := k8sutil.MergePatchContainers(operatorInitContainers, pt.GetInitContainers())
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge init containers spec")
	}

	// command line arguments
	cmdArgs, warns, err := MakePrometheusCommandArgs(pt)
	if err != nil {
		return nil, err
	}
	for _, msg := range warns {
		level.Warn(logger).Log("msg", msg)
	}
	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    imagePath,
			Ports:                    ports,
			Args:                     cmdArgs,
			VolumeMounts:             mounts,
			StartupProbe:             startupProbe,
			LivenessProbe:            livenessProbe,
			ReadinessProbe:           readinessProbe,
			Resources:                pt.GetResources(),
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			SecurityContext: &v1.SecurityContext{
				ReadOnlyRootFilesystem:   &boolTrue,
				AllowPrivilegeEscalation: &boolFalse,
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
			},
		},
		CreateConfigReloader(
			"config-reloader",
			ReloaderResources(c.ReloaderConfig),
			ReloaderURL(url.URL{
				Scheme: getURIScheme(webSpec),
				Host:   fmt.Sprintf("%s:9090", c.LocalHost),
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			}),
			ListenLocal(pt.ListensOn() == "localhost"),
			LocalHost(c.LocalHost),
			LogFormat(logFormat),
			LogLevel(logLevel),
			ConfigFile(path.Join(PrometheusConfDir, PrometheusConfFilename)),
			ConfigEnvsubstFile(path.Join(PrometheusConfOutDir, PrometheusConfEnvSubstFilename)),
			WatchedDirectories(watchedDirectories), VolumeMounts(configReloaderVolumeMounts),
			Shard(shard),
		),
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, pt.GetContainers())
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		ServiceName:         GoverningServiceName,
		Replicas:            pt.GetReplicas(),
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
			Spec: pt.MakePodSpec(containers, initContainers, volumes),
		},
	}, nil
}

func MakePrometheusStatefulSet(logger log.Logger, name string, pt PrometheusType,
	config *Config, ruleConfigMapNames []string, inputHash string, shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSet, error) {
	// p is passed in by value, not by reference. But p contains references like
	// to annotation map, that do not get copied on function invocation. Ensure to
	// prevent side effects before editing p by creating a deep copy. For more
	// details see https://github.com/prometheus-operator/prometheus-operator/issues/1659.
	pt = pt.Duplicate()
	nc := pt.GetNomenclator()

	version, err := pt.GetVersion(DefaultPrometheusVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse prometheus version")
	}

	pt.SetDefaultPortname(DefaultPrometheusPortName)

	replicas := pt.GetReplicas()
	if replicas == nil {
		pt.SetDefaultReplicas(&minReplicas)
	} else if *replicas < 0 {
		pt.SetDefaultReplicas(&int32Zero)
	}

	resources := pt.GetResources()
	requests := v1.ResourceList{}
	if resources.Requests != nil {
		requests = resources.Requests
	}
	_, memoryRequestFound := requests[v1.ResourceMemory]
	memoryLimit, memoryLimitFound := resources.Limits[v1.ResourceMemory]
	if !memoryRequestFound && version.Major == 1 {
		defaultMemoryRequest := resource.MustParse(DefaultMemoryRequestValue)
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		// If limit is given and smaller or equal to 2Gi, then set memory
		// request to the given limit. This is necessary as if limit < request,
		// then a Pod is not schedulable.
		if memoryLimitFound && compareResult <= 0 {
			requests[v1.ResourceMemory] = memoryLimit
		} else {
			requests[v1.ResourceMemory] = defaultMemoryRequest
		}
	}
	pt.SetResourceRequests(requests)

	spec, err := MakePrometheusStatefulsetSpec(pt, logger, config, shard, ruleConfigMapNames, tlsAssetSecrets)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	objectMeta := pt.GetObjectMeta()
	annotations := make(map[string]string)
	for key, value := range objectMeta.Annotations {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	labels := make(map[string]string)
	for key, value := range objectMeta.Labels {
		labels[key] = value
	}
	labels[shardLabelName] = fmt.Sprintf("%d", shard)
	labels[prometheusNameLabelName] = nc.BaseName()

	ownr := pt.GetOwnerReference()
	ownr.BlockOwnerDeletion = &boolTrue
	ownr.Controller = &boolTrue
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          config.Labels.Merge(labels),
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{ownr},
		},
		Spec: *spec,
	}

	if statefulset.ObjectMeta.Annotations == nil {
		statefulset.ObjectMeta.Annotations = map[string]string{
			StatefulSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[StatefulSetInputHashName] = inputHash
	}

	imagePullSecrets := pt.GetImagePullSecrets()
	if imagePullSecrets != nil && len(imagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = imagePullSecrets
	}
	storageSpec := pt.GetStorageSpec()
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else if storageSpec.Ephemeral != nil {
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
			},
		})
	} else {
		pvcTemplate := MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = nc.VolumeName()
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

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, pt.GetVolumes()...)

	return statefulset, nil
}
