// Copyright 2020 The prometheus-operator Authors
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
	"strings"

	"github.com/blang/semver/v4"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

var (
	boolFalse = false
	boolTrue  = true
	int32Zero int32
)

// PrometheusType defines interface for both modes Prometheus is operating
type PrometheusType interface {
	// GetNomenclator returns object used for naming other objects unified way
	GetNomenclator() *Nomenclator
	// Returns common fields of Prometheus resources
	GetCommonFields() *monitoringv1.CommonPrometheusFields
	// MakeCommandArgs returns map of command line arguments for object's container, slice of warnings raised during generation process and error
	MakeCommandArgs() (map[string]string, []string, error)
}

//---------------------------- CLI argument helpers ----------------------------
// ProcessArgs merge additional arguments to operator argument map and transformt to slice of CLI arguments
func ProcessCommandArgs(opArgs map[string]string, addArgs []monitoringv1.Argument) ([]string, error) {
	// merge operator args with additional args
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

	out := make([]string, 0, len(opArgs))
	if len(invalid) > 0 {
		return out, errors.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(invalid, ","))
	}

	// transform arg map to list of cli args
	for key, value := range opArgs {
		item := fmt.Sprintf("--%s", key)
		if value != "" {
			item = fmt.Sprintf("%s=%s", item, value)
		}
		out = append(out, item)
	}
	return out, nil
}

// MakePrometheusCommandArgs returns slice of Prometheus command arguments for either Prometheus server or Prometheus agent
func MakePrometheusCommandArgs(pt PrometheusType) (out, warns []string, err error) {
	args, warns, err := pt.MakeCommandArgs()
	if err != nil {
		return out, warns, err
	}

	pcf := pt.GetCommonFields()
	out, err = ProcessCommandArgs(args, pcf.AdditionalArgs)
	if err != nil {
		return out, warns, err
	}
	return out, warns, nil
}

//---------------------------- StatefulSet helpers -----------------------------

// PrometheusVolumeFactory can be used for creating and mounting additional volumes to main prometheus container
type PrometheusVolumeFactory func() ([]v1.Volume, []v1.VolumeMount, error)

// PrometheusContainerFactory can be used for adding additional containers to StatefulSet
type PrometheusContainerFactory func() ([]v1.Container, error)

// PrometheusStatefulSetGenerator objects generates StatefulSet from the given resource
type PrometheusStatefulSetGenerator struct {
	config *Config
	logger log.Logger

	volumeFactory    PrometheusVolumeFactory
	containerFactory PrometheusContainerFactory
}

func (pssg *PrometheusStatefulSetGenerator) MakePrometheusStatefulSetSpec(pt PrometheusType,
	imagePath string, shard int32, tlsAssetSecrets []string) (*appsv1.StatefulSetSpec, error) {

	nc := pt.GetNomenclator()
	pcf := pt.GetCommonFields()

	// version checks
	version, err := ParseVersion(pcf.Version)
	if err != nil {
		return nil, err
	}
	if version.Major != 2 {
		return nil, errors.Errorf("unsupported Prometheus major version %s", version)
	}
	if pcf.EnableRemoteWriteReceiver && !version.GTE(semver.MustParse("2.33.0")) {
		level.Warn(pssg.logger).Log("msg", "ignoring 'enableRemoteWriteReceiver' not supported by Prometheus", "version", version, "minimum_version", "2.33.0")
	}

	var ports []v1.ContainerPort
	if !pcf.ListenLocal {
		ports = []v1.ContainerPort{
			{
				Name:          pcf.PortName,
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
	factVolumes, factVolMounts, err := pssg.volumeFactory()
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, factVolumes...)

	volName := nc.VolumeName()
	if pcf.Storage != nil && pcf.Storage.VolumeClaimTemplate.Name != "" {
		volName = pcf.Storage.VolumeClaimTemplate.Name
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
			SubPath:   nc.SubPathForStorage(pcf.Storage),
		},
	}
	mounts = append(append(mounts, factVolMounts...), pcf.VolumeMounts...)

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 2.24.0.
	// With this we avoid redeploying prometheus when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	if version.GTE(semver.MustParse("2.24.0")) {
		var fields monitoringv1.WebConfigFileFields
		if pcf.Web != nil {
			fields = pcf.Web.WebConfigFileFields
		}

		webConfig, err := webconfig.New(WebConfigDir, nc.WebConfigSecretName(), fields)
		if err != nil {
			return nil, err
		}

		_, configVol, configMount := webConfig.GetMountParameters()
		volumes = append(volumes, configVol...)
		mounts = append(mounts, configMount...)
	}

	// Mount related secrets
	rn := k8sutil.NewResourceNamerWithPrefix("secret")
	for _, s := range pcf.Secrets {
		name, err := rn.UniqueVolumeName(s)
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
		mounts = append(mounts, v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: path.Join(PrometheusSecretsDir, s),
		})
	}

	rn = k8sutil.NewResourceNamerWithPrefix("configmap")
	for _, c := range pcf.ConfigMaps {
		name, err := rn.UniqueVolumeName(c)
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
		mounts = append(mounts, v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: path.Join(PrometheusConfigmapsDir, c),
		})
	}

	webRoutePrefix := "/"
	if pcf.RoutePrefix != "" {
		webRoutePrefix = pcf.RoutePrefix
	}

	probeHandler := func(probePath string) v1.ProbeHandler {
		probePath = path.Clean(webRoutePrefix + probePath)
		handler := v1.ProbeHandler{}
		if pcf.ListenLocal {
			probeURL := url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("localhost:%s", DefaultPrometheusContainerPort),
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
			Port: intstr.FromString(pcf.PortName),
		}
		if pcf.Web != nil && pcf.Web.TLSConfig != nil && version.GTE(semver.MustParse("2.24.0")) {
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
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}

	livenessProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/healthy"),
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler:     probeHandler("/-/ready"),
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{
		"app.kubernetes.io/version": version.String(),
	}
	if pcf.PodMetadata != nil {
		if pcf.PodMetadata.Labels != nil {
			for k, v := range pcf.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if pcf.PodMetadata.Annotations != nil {
			for k, v := range pcf.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := map[string]string{
		"app.kubernetes.io/name":       nc.Prefix(),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   nc.BaseName(),
		nc.Prefix():                    nc.BaseName(),
		shardLabelName:                 fmt.Sprintf("%d", shard),
		nc.NameLabelName():             nc.BaseName(),
	}
	for k, v := range podSelectorLabels {
		podLabels[k] = v
	}

	podAnnotations["kubectl.kubernetes.io/default-container"] = nc.Prefix()

	finalSelectorLabels := pssg.config.Labels.Merge(podSelectorLabels)
	finalLabels := pssg.config.Labels.Merge(podLabels)

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

	var minReadySeconds int32
	if pcf.MinReadySeconds != nil {
		minReadySeconds = int32(*pcf.MinReadySeconds)
	}

	// command line arguments
	cmdArgs, warns, err := MakePrometheusCommandArgs(pt)
	if err != nil {
		return nil, err
	}
	for _, msg := range warns {
		level.Warn(pssg.logger).Log("msg", msg)
	}

	additionalContainers, err := pssg.containerFactory()
	if err != nil {
		return nil, err
	}

	operatorContainers := append([]v1.Container{
		{
			Name:                     nc.Prefix(),
			Image:                    imagePath,
			Ports:                    ports,
			Args:                     cmdArgs,
			VolumeMounts:             mounts,
			StartupProbe:             startupProbe,
			LivenessProbe:            livenessProbe,
			ReadinessProbe:           readinessProbe,
			Resources:                pcf.Resources,
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
			ReloaderResources(pssg.config.ReloaderConfig),
			ReloaderURL(url.URL{
				Scheme: GetURIScheme(pcf.Web),
				Host:   fmt.Sprintf("%s:9090", pssg.config.LocalHost),
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			}),
			ListenLocal(pcf.ListenLocal),
			LocalHost(pssg.config.LocalHost),
			LogFormat(pcf.LogFormat),
			LogLevel(pcf.LogLevel),
			ConfigFile(path.Join(PrometheusConfDir, PrometheusConfFilename)),
			ConfigEnvsubstFile(path.Join(PrometheusConfOutDir, PrometheusConfEnvSubstFilename)),
			WatchedDirectories(watchedDirectories), VolumeMounts(configReloaderVolumeMounts),
			Shard(shard),
		),
	}, additionalContainers...)
	containers, err := k8sutil.MergePatchContainers(operatorContainers, pcf.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	operatorInitContainers := []v1.Container{
		CreateConfigReloader(
			"init-config-reloader",
			ReloaderResources(pssg.config.ReloaderConfig),
			ReloaderRunOnce(),
			LogFormat(pcf.LogFormat),
			LogLevel(pcf.LogLevel),
			VolumeMounts(configReloaderVolumeMounts),
			ConfigFile(path.Join(PrometheusConfDir, PrometheusConfFilename)),
			ConfigEnvsubstFile(path.Join(PrometheusConfOutDir, PrometheusConfEnvSubstFilename)),
			WatchedDirectories(watchedDirectories),
			Shard(shard),
		),
	}
	initContainers, err := k8sutil.MergePatchContainers(operatorInitContainers, pcf.InitContainers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge init containers spec")
	}

	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		ServiceName:         GoverningServiceName,
		Replicas:            pcf.Replicas,
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
				SecurityContext:               pcf.SecurityContext,
				ServiceAccountName:            pcf.ServiceAccountName,
				AutomountServiceAccountToken:  &boolTrue,
				NodeSelector:                  pcf.NodeSelector,
				PriorityClassName:             pcf.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   pcf.Tolerations,
				Affinity:                      pcf.Affinity,
				TopologySpreadConstraints:     pcf.TopologySpreadConstraints,
				HostAliases:                   MakeHostAliases(pcf.HostAliases),
			},
		},
	}, nil
}

func ParseVersion(verStr string) (semver.Version, error) {
	version, err := semver.ParseTolerant(StringValOrDefault(verStr, DefaultPrometheusVersion))
	if err != nil {
		return version, errors.Wrap(err, "failed to parse prometheus version")
	}
	return version, nil
}

func GetURIScheme(webSpec *monitoringv1.PrometheusWebSpec) string {
	URIScheme := "http"
	if webSpec != nil && webSpec.TLSConfig != nil {
		URIScheme = "https"
	}
	return URIScheme
}

func MakeVolumeClaimTemplate(e monitoringv1.EmbeddedPersistentVolumeClaim) *v1.PersistentVolumeClaim {
	pvc := v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: e.APIVersion,
			Kind:       e.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.Name,
			Labels:      e.Labels,
			Annotations: e.Annotations,
		},
		Spec:   e.Spec,
		Status: e.Status,
	}
	return &pvc
}

// MakeHostAliases converts array of monitoringv1 HostAlias to array of corev1 HostAlias
func MakeHostAliases(input []monitoringv1.HostAlias) []v1.HostAlias {
	if len(input) == 0 {
		return nil
	}

	output := make([]v1.HostAlias, len(input))

	for i, in := range input {
		output[i].Hostnames = in.Hostnames
		output[i].IP = in.IP
	}

	return output
}
