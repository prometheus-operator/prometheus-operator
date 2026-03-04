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

package alertmanager

import (
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"path"
	"strings"

	"github.com/alecthomas/units"
	"github.com/blang/semver/v4"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/clustertlsconfig"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	// WARNING: Do not use directly - users might specify a different service name!
	defaultOperatedServiceName = "alertmanager-operated"

	defaultRetention = "120h"
	defaultPortName  = "web"

	tlsAssetsVolumeName                = "tls-assets"
	tlsAssetsDir                       = "/etc/alertmanager/certs"
	secretsDir                         = "/etc/alertmanager/secrets"
	configmapsDir                      = "/etc/alertmanager/configmaps"
	alertmanagerTemplatesVolumeName    = "notification-templates"
	alertmanagerTemplatesDir           = "/etc/alertmanager/templates"
	webConfigDir                       = "/etc/alertmanager/web_config"
	clusterTLSConfigDir                = "/etc/alertmanager/cluster_tls_config"
	alertmanagerConfigVolumeName       = "config-volume"
	alertmanagerConfigDir              = "/etc/alertmanager/config"
	alertmanagerConfigOutVolumeName    = "config-out"
	alertmanagerConfigOutDir           = "/etc/alertmanager/config_out"
	alertmanagerConfigFile             = "alertmanager.yaml"
	alertmanagerConfigFileCompressed   = "alertmanager.yaml.gz"
	alertmanagerConfigEnvsubstFilename = "alertmanager.env.yaml"

	alertmanagerWebPort         = 9093
	alertmanagerMeshPort        = 9094
	alertmanagerMeshUDPPortName = "mesh-udp"
	alertmanagerMeshTCPPortName = "mesh-tcp"

	alertmanagerStorageDir = "/alertmanager"

	defaultTerminationGracePeriodSeconds = int64(120)
)

var (
	minReplicas         int32 = 1
	probeTimeoutSeconds int32 = 3
)

func getServiceName(a *monitoringv1.Alertmanager) string {
	return ptr.Deref(a.Spec.ServiceName, defaultOperatedServiceName)
}

func makeStatefulSet(logger *slog.Logger, am *monitoringv1.Alertmanager, config Config, inputHash string, tlsSecrets *operator.ShardedSecret) (*appsv1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if am.Spec.PortName == "" {
		am.Spec.PortName = defaultPortName
	}
	if am.Spec.Replicas == nil {
		am.Spec.Replicas = &minReplicas
	}
	intZero := int32(0)
	if am.Spec.Replicas != nil && *am.Spec.Replicas < 0 {
		am.Spec.Replicas = &intZero
	}
	// TODO(slashpai): Remove this assignment after v0.60 since this is handled at CRD level
	if am.Spec.Retention == "" {
		am.Spec.Retention = defaultRetention
	}
	if am.Spec.Resources.Requests == nil {
		am.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := am.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		am.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("200Mi")
	}

	spec, err := makeStatefulSetSpec(logger, am, config, tlsSecrets)
	if err != nil {
		return nil, err
	}

	statefulset := &appsv1.StatefulSet{Spec: *spec}
	operator.UpdateObject(
		statefulset,
		operator.WithName(prefixedName(am.Name)),
		operator.WithAnnotations(am.GetAnnotations()),
		operator.WithAnnotations(config.Annotations),
		operator.WithInputHashAnnotation(inputHash),
		operator.WithLabels(am.GetLabels()),
		operator.WithSelectorLabels(spec.Selector),
		operator.WithLabels(config.Labels),
		operator.WithManagingOwner(am),
		operator.WithoutKubectlAnnotations(),
	)

	if len(am.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = am.Spec.ImagePullSecrets
	}

	storageSpec := am.Spec.Storage
	switch {
	case storageSpec == nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})

	case storageSpec.EmptyDir != nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: storageSpec.EmptyDir,
			},
		})

	case storageSpec.Ephemeral != nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				Ephemeral: storageSpec.Ephemeral,
			},
		})

	default: // storageSpec.VolumeClaimTemplate
		pvcTemplate := operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(am.Name)
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

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, am.Spec.Volumes...)

	if am.Spec.PersistentVolumeClaimRetentionPolicy != nil {
		statefulset.Spec.PersistentVolumeClaimRetentionPolicy = am.Spec.PersistentVolumeClaimRetentionPolicy
	}

	return statefulset, nil
}

func makeStatefulSetService(a *monitoringv1.Alertmanager, config Config) *v1.Service {
	if a.Spec.PortName == "" {
		a.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP:                v1.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Ports: []v1.ServicePort{
				{
					Name:       a.Spec.PortName,
					Port:       alertmanagerWebPort,
					TargetPort: intstr.FromString(a.Spec.PortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "tcp-mesh",
					Port:       alertmanagerMeshPort,
					TargetPort: intstr.FromString(alertmanagerMeshTCPPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "udp-mesh",
					Port:       alertmanagerMeshPort,
					TargetPort: intstr.FromString(alertmanagerMeshUDPPortName),
					Protocol:   v1.ProtocolUDP,
				},
			},
			Selector: map[string]string{
				operator.ApplicationNameLabelKey: applicationNameLabelValue,
			},
		},
	}

	operator.UpdateObject(
		svc,
		operator.WithName(defaultOperatedServiceName),
		operator.WithAnnotations(config.Annotations),
		operator.WithLabels(map[string]string{"operated-alertmanager": "true"}),
		operator.WithLabels(config.Labels),
		operator.WithOwner(a),
	)

	return svc
}

func makeStatefulSetSpec(logger *slog.Logger, a *monitoringv1.Alertmanager, config Config, tlsSecrets *operator.ShardedSecret) (*appsv1.StatefulSetSpec, error) {
	amVersion := operator.StringValOrDefault(a.Spec.Version, operator.DefaultAlertmanagerVersion)
	amImagePath, err := operator.BuildImagePath(
		ptr.Deref(a.Spec.Image, ""),
		operator.StringValOrDefault(a.Spec.BaseImage, config.AlertmanagerDefaultBaseImage),
		amVersion,
		operator.StringValOrDefault(a.Spec.Tag, ""),
		operator.StringValOrDefault(a.Spec.SHA, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build image path: %w", err)
	}

	version, err := semver.ParseTolerant(amVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alertmanager version: %w", err)
	}

	amArgs := []monitoringv1.Argument{
		{Name: "config.file", Value: path.Join(alertmanagerConfigOutDir, alertmanagerConfigEnvsubstFilename)},
		{Name: "storage.path", Value: alertmanagerStorageDir},
		{Name: "data.retention", Value: string(a.Spec.Retention)},
	}

	if *a.Spec.Replicas == 1 && !a.Spec.ForceEnableClusterMode {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.listen-address=", Value: ""})
	} else {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.listen-address", Value: "[$(POD_IP)]:9094"})
	}

	if a.Spec.ListenLocal {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "web.listen-address", Value: "127.0.0.1:9093"})
	} else {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "web.listen-address", Value: ":9093"})
	}

	if a.Spec.ExternalURL != "" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "web.external-url", Value: a.Spec.ExternalURL})
	}

	if version.GTE(semver.MustParse("0.27.0")) && len(a.Spec.EnableFeatures) > 0 {
		amArgs = append(amArgs, monitoringv1.Argument{
			Name:  "enable-feature",
			Value: strings.Join(a.Spec.EnableFeatures[:], ","),
		})
	}

	webRoutePrefix := "/"
	if a.Spec.RoutePrefix != "" {
		webRoutePrefix = a.Spec.RoutePrefix
	}

	amArgs = append(amArgs, monitoringv1.Argument{Name: "web.route-prefix", Value: webRoutePrefix})

	web := a.Spec.Web
	if version.GTE(semver.MustParse("0.17.0")) && web != nil && web.GetConcurrency != nil {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "web.get-concurrency", Value: fmt.Sprintf("%d", *web.GetConcurrency)})
	}

	if version.GTE(semver.MustParse("0.17.0")) && web != nil && web.Timeout != nil {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "web.timeout", Value: fmt.Sprintf("%d", *web.Timeout)})
	}

	limits := a.Spec.Limits
	if version.GTE(semver.MustParse("0.28.0")) && limits != nil {
		if limits.MaxSilences != nil {
			amArgs = append(amArgs, monitoringv1.Argument{Name: "silences.max-silences", Value: fmt.Sprintf("%d", *limits.MaxSilences)})
		}

		if !limits.MaxPerSilenceBytes.IsEmpty() {
			vBytes, _ := units.ParseBase2Bytes(string(*limits.MaxPerSilenceBytes))
			amArgs = append(amArgs, monitoringv1.Argument{Name: "silences.max-per-silence-bytes", Value: fmt.Sprintf("%d", int64(vBytes))})
		}

	}

	if version.GTE(semver.MustParse("0.30.0")) && a.Spec.MinReadySeconds != nil {
		startDelayArg := monitoringv1.Argument{
			Name:  "dispatch.start-delay",
			Value: fmt.Sprintf("%ds", *a.Spec.MinReadySeconds),
		}
		if i := operator.ArgumentsIntersection([]monitoringv1.Argument{startDelayArg}, a.Spec.AdditionalArgs); len(i) == 0 {
			amArgs = append(amArgs, startDelayArg)
		}
	}

	if a.Spec.LogLevel != "" && a.Spec.LogLevel != "info" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "log.level", Value: a.Spec.LogLevel})
	}

	if version.GTE(semver.MustParse("0.16.0")) {
		if a.Spec.LogFormat != "" && a.Spec.LogFormat != "logfmt" {
			amArgs = append(amArgs, monitoringv1.Argument{Name: "log.format", Value: a.Spec.LogFormat})
		}
	}

	if a.Spec.ClusterAdvertiseAddress != "" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.advertise-address", Value: a.Spec.ClusterAdvertiseAddress})
	}

	if a.Spec.ClusterGossipInterval != "" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.gossip-interval", Value: string(a.Spec.ClusterGossipInterval)})
	}

	if a.Spec.ClusterPushpullInterval != "" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.pushpull-interval", Value: string(a.Spec.ClusterPushpullInterval)})
	}

	if a.Spec.ClusterPeerTimeout != "" {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.peer-timeout", Value: string(a.Spec.ClusterPeerTimeout)})
	}

	// If multiple Alertmanager clusters are deployed on the same cluster, it can happen
	// that because pod IP addresses are recycled, an Alertmanager instance from cluster B
	// connects with cluster A.
	// --cluster.label flag was introduced in alertmanager v0.26, this helps to block
	// any traffic that is not meant for the cluster.
	if version.GTE(semver.MustParse("0.26.0")) {
		clusterLabel := fmt.Sprintf("%s/%s", a.Namespace, a.Name)
		if a.Spec.ClusterLabel != nil {
			clusterLabel = *a.Spec.ClusterLabel
		}
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.label", Value: clusterLabel})
	}

	isHTTPS := a.Spec.Web != nil && a.Spec.Web.TLSConfig != nil && version.GTE(semver.MustParse("0.22.0"))

	livenessProbeHandler := v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/-/healthy"),
			Port: intstr.FromString(a.Spec.PortName),
		},
	}

	readinessProbeHandler := v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/-/ready"),
			Port: intstr.FromString(a.Spec.PortName),
		},
	}

	var livenessProbe *v1.Probe
	var readinessProbe *v1.Probe
	if !a.Spec.ListenLocal {
		livenessProbe = &v1.Probe{
			ProbeHandler:     livenessProbeHandler,
			TimeoutSeconds:   probeTimeoutSeconds,
			FailureThreshold: 10,
		}

		readinessProbe = &v1.Probe{
			ProbeHandler:        readinessProbeHandler,
			InitialDelaySeconds: 3,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
			FailureThreshold:    10,
		}

		if isHTTPS {
			livenessProbe.HTTPGet.Scheme = v1.URISchemeHTTPS
			readinessProbe.HTTPGet.Scheme = v1.URISchemeHTTPS
		}
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{
		operator.ApplicationVersionLabelKey: version.String(),
	}
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := makeSelectorLabels(a.GetObjectMeta().GetName())
	if a.Spec.PodMetadata != nil {
		maps.Copy(podLabels, a.Spec.PodMetadata.Labels)
		maps.Copy(podAnnotations, a.Spec.PodMetadata.Annotations)
	}
	maps.Copy(podLabels, podSelectorLabels)

	podAnnotations[operator.DefaultContainerAnnotationKey] = "alertmanager"

	var operatorInitContainers []v1.Container

	var clusterPeerDomain string
	if config.ClusterDomain != "" {
		clusterPeerDomain = fmt.Sprintf("%s.%s.svc.%s.", getServiceName(a), a.Namespace, config.ClusterDomain)
	} else {
		// The default DNS search path is .svc.<cluster domain>
		clusterPeerDomain = getServiceName(a)
	}
	for i := int32(0); i < *a.Spec.Replicas; i++ {
		amArgs = append(amArgs, monitoringv1.Argument{
			Name:  "cluster.peer",
			Value: fmt.Sprintf("%s-%d.%s:9094", prefixedName(a.Name), i, clusterPeerDomain),
		})
	}

	for _, peer := range a.Spec.AdditionalPeers {
		amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.peer", Value: peer})
	}

	ports := []v1.ContainerPort{
		{
			Name:          alertmanagerMeshTCPPortName,
			ContainerPort: alertmanagerMeshPort,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          alertmanagerMeshUDPPortName,
			ContainerPort: alertmanagerMeshPort,
			Protocol:      v1.ProtocolUDP,
		},
	}
	if !a.Spec.ListenLocal {
		ports = append([]v1.ContainerPort{
			{
				Name:          a.Spec.PortName,
				ContainerPort: alertmanagerWebPort,
				Protocol:      v1.ProtocolTCP,
			},
		}, ports...)
	}

	// Override default 6h value to allow AlertManager cluster to
	// quickly remove a cluster member after its pod restarted or during a
	// regular rolling update.
	amArgs = append(amArgs, monitoringv1.Argument{Name: "cluster.reconnect-timeout", Value: "5m"})

	volumes := []v1.Volume{
		{
			Name: alertmanagerConfigVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: generatedConfigSecretName(a.Name),
				},
			},
		},
		tlsSecrets.Volume(tlsAssetsVolumeName),
		{
			Name: alertmanagerConfigOutVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{
					// tmpfs is used here to avoid writing sensitive data into disk.
					Medium: v1.StorageMediumMemory,
				},
			},
		},
	}

	volName := volumeName(a.Name)
	if a.Spec.Storage != nil {
		if a.Spec.Storage.VolumeClaimTemplate.Name != "" {
			volName = a.Spec.Storage.VolumeClaimTemplate.Name
		}
	}

	amVolumeMounts := []v1.VolumeMount{
		{
			Name:      alertmanagerConfigVolumeName,
			MountPath: alertmanagerConfigDir,
		},
		{
			Name:      alertmanagerConfigOutVolumeName,
			ReadOnly:  true,
			MountPath: alertmanagerConfigOutDir,
		},
		{
			Name:      tlsAssetsVolumeName,
			ReadOnly:  true,
			MountPath: tlsAssetsDir,
		},
		{
			Name:      volName,
			MountPath: alertmanagerStorageDir,
			SubPath:   subPathForStorage(a.Spec.Storage),
		},
	}

	var configReloaderWebConfigFile string

	watchedDirectories := []string{alertmanagerConfigDir}
	configReloaderVolumeMounts := []v1.VolumeMount{
		{
			Name:      alertmanagerConfigVolumeName,
			MountPath: alertmanagerConfigDir,
			ReadOnly:  true,
		},
		{
			Name:      alertmanagerConfigOutVolumeName,
			MountPath: alertmanagerConfigOutDir,
		},
	}

	amCfg := a.Spec.AlertmanagerConfiguration
	if amCfg != nil && len(amCfg.Templates) > 0 {
		sources := []v1.VolumeProjection{}
		keys := sets.Set[string]{}
		for _, v := range amCfg.Templates {
			if v.ConfigMap != nil {
				if keys.Has(v.ConfigMap.Key) {
					logger.Debug(fmt.Sprintf("skipping %q due to duplicate key %q", v.ConfigMap.Key, v.ConfigMap.Name))
					continue
				}
				sources = append(sources, v1.VolumeProjection{
					ConfigMap: &v1.ConfigMapProjection{
						LocalObjectReference: v1.LocalObjectReference{
							Name: v.ConfigMap.Name,
						},
						Items: []v1.KeyToPath{{
							Key:  v.ConfigMap.Key,
							Path: v.ConfigMap.Key,
						}},
					},
				})
				keys.Insert(v.ConfigMap.Key)
			}
			if v.Secret != nil {
				if keys.Has(v.Secret.Key) {
					logger.Debug(fmt.Sprintf("skipping %q due to duplicate key %q", v.Secret.Key, v.Secret.Name))
					continue
				}
				sources = append(sources, v1.VolumeProjection{
					Secret: &v1.SecretProjection{
						LocalObjectReference: v1.LocalObjectReference{
							Name: v.Secret.Name,
						},
						Items: []v1.KeyToPath{{
							Key:  v.Secret.Key,
							Path: v.Secret.Key,
						}},
					},
				})
				keys.Insert(v.Secret.Key)
			}
		}
		volumes = append(volumes, v1.Volume{
			Name: "notification-templates",
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: sources,
				},
			},
		})
		amVolumeMounts = append(amVolumeMounts, v1.VolumeMount{
			Name:      alertmanagerTemplatesVolumeName,
			ReadOnly:  true,
			MountPath: alertmanagerTemplatesDir,
		})
		configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
			Name:      alertmanagerTemplatesVolumeName,
			ReadOnly:  true,
			MountPath: alertmanagerTemplatesDir,
		})
		watchedDirectories = append(watchedDirectories, alertmanagerTemplatesDir)
	}

	rn := k8s.NewResourceNamerWithPrefix("secret")
	for _, s := range a.Spec.Secrets {
		name, err := rn.DNS1123Label(s)
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
		mountPath := path.Join(secretsDir, s)
		mount := v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: mountPath,
		}
		amVolumeMounts = append(amVolumeMounts, mount)
		configReloaderVolumeMounts = append(configReloaderVolumeMounts, mount)
		watchedDirectories = append(watchedDirectories, mountPath)
	}

	rn = k8s.NewResourceNamerWithPrefix("configmap")
	for _, c := range a.Spec.ConfigMaps {
		name, err := rn.DNS1123Label(c)
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
		mountPath := path.Join(configmapsDir, c)
		mount := v1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: mountPath,
		}
		amVolumeMounts = append(amVolumeMounts, mount)
		configReloaderVolumeMounts = append(configReloaderVolumeMounts, mount)
		watchedDirectories = append(watchedDirectories, mountPath)
	}

	amVolumeMounts = append(amVolumeMounts, a.Spec.VolumeMounts...)

	// Mount web config and web TLS credentials as volumes.
	// We always mount the web config file for versions greater than 0.22.0.
	// With this we avoid redeploying alertmanager when reconfiguring between
	// HTTP and HTTPS and vice-versa.
	if version.GTE(semver.MustParse("0.22.0")) {
		var fields monitoringv1.WebConfigFileFields
		if a.Spec.Web != nil {
			fields = a.Spec.Web.WebConfigFileFields
		}

		webConfig, err := webconfig.New(webConfigDir, webConfigSecretName(a.Name), fields)
		if err != nil {
			return nil, err
		}

		confArg, configVol, configMount, err := webConfig.GetMountParameters()
		if err != nil {
			return nil, err
		}
		amArgs = append(amArgs, monitoringv1.Argument{Name: confArg.Name, Value: confArg.Value})
		volumes = append(volumes, configVol...)
		amVolumeMounts = append(amVolumeMounts, configMount...)

		configReloaderWebConfigFile = confArg.Value
		configReloaderVolumeMounts = append(configReloaderVolumeMounts, configMount...)
	}

	if version.GTE(semver.MustParse("0.24.0")) {
		clusterTLSConfig, err := clustertlsconfig.New(clusterTLSConfigDir, a)
		if err != nil {
			return nil, fmt.Errorf("failed to create the cluster TLS configuration: %w", err)
		}

		confArg, configVol, configMount, err := clusterTLSConfig.GetMountParameters()
		if err != nil {
			return nil, fmt.Errorf("failed to get mount parameters for cluster TLS configuration: %w", err)
		}

		// confArg is nil if the Alertmanager resource doesn't configure mTLS for the cluster protocol.
		if confArg != nil {
			amArgs = append(amArgs, monitoringv1.Argument{Name: confArg.Name, Value: confArg.Value})
		}
		volumes = append(volumes, configVol...)
		amVolumeMounts = append(amVolumeMounts, configMount...)
	}

	finalSelectorLabels := config.Labels.Merge(podSelectorLabels)
	finalLabels := config.Labels.Merge(podLabels)

	alertmanagerURIScheme := "http"
	if isHTTPS {
		alertmanagerURIScheme = "https"
	}

	containerArgs, err := operator.BuildArgs(amArgs, a.Spec.AdditionalArgs)
	if err != nil {
		return nil, err
	}

	defaultContainers := []v1.Container{
		{
			Args:            containerArgs,
			Name:            "alertmanager",
			Image:           amImagePath,
			ImagePullPolicy: a.Spec.ImagePullPolicy,
			Ports:           ports,
			VolumeMounts:    amVolumeMounts,
			LivenessProbe:   livenessProbe,
			ReadinessProbe:  readinessProbe,
			Resources:       a.Spec.Resources,
			SecurityContext: &v1.SecurityContext{
				AllowPrivilegeEscalation: ptr.To(false),
				ReadOnlyRootFilesystem:   ptr.To(true),
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
			},
			Env: []v1.EnvVar{
				{
					// Necessary for '--cluster.listen-address' flag
					Name: "POD_IP",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
			},
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
		operator.CreateConfigReloader(
			"config-reloader",
			operator.ReloaderConfig(config.ReloaderConfig),
			operator.ReloaderURL(url.URL{
				Scheme: alertmanagerURIScheme,
				Host:   config.LocalHost + ":9093",
				Path:   path.Clean(webRoutePrefix + "/-/reload"),
			}),
			operator.ListenLocal(a.Spec.ListenLocal),
			operator.LocalHost(config.LocalHost),
			operator.LogFormat(a.Spec.LogFormat),
			operator.LogLevel(a.Spec.LogLevel),
			operator.WatchedDirectories(watchedDirectories),
			operator.VolumeMounts(configReloaderVolumeMounts),
			operator.Shard(-1),
			operator.WebConfigFile(configReloaderWebConfigFile),
			operator.ConfigFile(path.Join(alertmanagerConfigDir, alertmanagerConfigFileCompressed)),
			operator.ConfigEnvsubstFile(path.Join(alertmanagerConfigOutDir, alertmanagerConfigEnvsubstFilename)),
			operator.ImagePullPolicy(a.Spec.ImagePullPolicy),
		),
	}

	containers, err := k8s.MergePatchContainers(defaultContainers, a.Spec.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	operatorInitContainers = append(operatorInitContainers,
		operator.CreateConfigReloader(
			"init-config-reloader",
			operator.ReloaderConfig(config.ReloaderConfig),
			operator.InitContainer(),
			operator.LogFormat(a.Spec.LogFormat),
			operator.LogLevel(a.Spec.LogLevel),
			operator.WatchedDirectories(watchedDirectories),
			operator.VolumeMounts(configReloaderVolumeMounts),
			operator.Shard(-1),
			operator.ConfigFile(path.Join(alertmanagerConfigDir, alertmanagerConfigFileCompressed)),
			operator.ConfigEnvsubstFile(path.Join(alertmanagerConfigOutDir, alertmanagerConfigEnvsubstFilename)),
			operator.ImagePullPolicy(a.Spec.ImagePullPolicy),
		),
	)

	initContainers, err := k8s.MergePatchContainers(operatorInitContainers, a.Spec.InitContainers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge init containers spec: %w", err)
	}

	// By default, podManagementPolicy is set to Parallel to mitigate rollout
	// issues in Kubernetes (see https://github.com/kubernetes/kubernetes/issues/60164).
	// This is also mentioned as one of limitations of StatefulSets:
	// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	podManagementPolicy := ptr.Deref(a.Spec.PodManagementPolicy, monitoringv1.ParallelPodManagement)

	spec := appsv1.StatefulSetSpec{
		ServiceName:         getServiceName(a),
		Replicas:            a.Spec.Replicas,
		MinReadySeconds:     ptr.Deref(a.Spec.MinReadySeconds, 0),
		PodManagementPolicy: appsv1.PodManagementPolicyType(podManagementPolicy),
		UpdateStrategy:      operator.UpdateStrategyForStatefulSet(a.Spec.UpdateStrategy),
		Selector: &metav1.LabelSelector{
			MatchLabels: finalSelectorLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				AutomountServiceAccountToken:  a.Spec.AutomountServiceAccountToken,
				NodeSelector:                  a.Spec.NodeSelector,
				PriorityClassName:             a.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: ptr.To(ptr.Deref(a.Spec.TerminationGracePeriodSeconds, defaultTerminationGracePeriodSeconds)),
				InitContainers:                initContainers,
				Containers:                    containers,
				Volumes:                       volumes,
				ServiceAccountName:            a.Spec.ServiceAccountName,
				SecurityContext:               a.Spec.SecurityContext,
				Tolerations:                   a.Spec.Tolerations,
				Affinity:                      a.Spec.Affinity,
				TopologySpreadConstraints:     a.Spec.TopologySpreadConstraints,
				HostAliases:                   operator.MakeHostAliases(a.Spec.HostAliases),
				EnableServiceLinks:            a.Spec.EnableServiceLinks,
				HostUsers:                     a.Spec.HostUsers,
				HostNetwork:                   a.Spec.HostNetwork,
			},
		},
	}

	if a.Spec.HostNetwork {
		spec.Template.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}
	k8s.UpdateDNSPolicy(&spec.Template.Spec, a.Spec.DNSPolicy)
	k8s.UpdateDNSConfig(&spec.Template.Spec, a.Spec.DNSConfig)
	return &spec, nil
}

func defaultConfigSecretName(am *monitoringv1.Alertmanager) string {
	if am.Spec.ConfigSecret == "" {
		return prefixedName(am.Name)
	}

	return am.Spec.ConfigSecret
}

func generatedConfigSecretName(name string) string {
	return prefixedName(name) + "-generated"
}

func webConfigSecretName(name string) string {
	return fmt.Sprintf("%s-web-config", prefixedName(name))
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("alertmanager-%s", name)
}

func subPathForStorage(s *monitoringv1.StorageSpec) string {
	if s == nil {
		return ""
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if s.DisableMountSubPath {
		return ""
	}

	return "alertmanager-db"
}
