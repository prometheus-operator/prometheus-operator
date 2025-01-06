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

package thanos

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"

	"github.com/blang/semver/v4"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	rulesDir                  = "/etc/thanos/rules"
	configDir                 = "/etc/thanos/config"
	storageDir                = "/thanos/data"
	webConfigDir              = "/etc/thanos/web_config"
	tlsAssetsDir              = "/etc/thanos/certs"
	governingServiceName      = "thanos-ruler-operated"
	defaultPortName           = "web"
	defaultRetention          = "24h"
	defaultEvaluationInterval = "15s"
	defaultReplicaLabelName   = "thanos_ruler_replica"
)

var (
	minReplicas int32 = 1
)

func makeStatefulSet(tr *monitoringv1.ThanosRuler, config Config, ruleConfigMapNames []string, inputHash string, tlsSecrets *operator.ShardedSecret) (*appsv1.StatefulSet, error) {

	if tr.Spec.Resources.Requests == nil {
		tr.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := tr.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		tr.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("200Mi")
	}

	spec, err := makeStatefulSetSpec(tr, config, ruleConfigMapNames, tlsSecrets)
	if err != nil {
		return nil, err
	}

	statefulset := &appsv1.StatefulSet{Spec: *spec}
	operator.UpdateObject(
		statefulset,
		operator.WithName(prefixedName(tr.Name)),
		operator.WithInputHashAnnotation(inputHash),
		operator.WithAnnotations(tr.GetAnnotations()),
		operator.WithAnnotations(config.Annotations),
		operator.WithLabels(tr.GetLabels()),
		operator.WithLabels(config.Labels),
		operator.WithManagingOwner(tr),
		operator.WithoutKubectlAnnotations(),
	)

	if len(tr.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = tr.Spec.ImagePullSecrets
	}

	storageSpec := tr.Spec.Storage
	switch {
	case storageSpec == nil:
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(tr.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})

	case storageSpec.EmptyDir != nil:
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(tr.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})

	case storageSpec.Ephemeral != nil:
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(tr.Name),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
			},
		})

	default: // storageSpec.VolumeClaimTemplate
		pvcTemplate := operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(tr.Name)
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

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, tr.Spec.Volumes...)

	return statefulset, nil
}

func makeStatefulSetSpec(tr *monitoringv1.ThanosRuler, config Config, ruleConfigMapNames []string, tlsSecrets *operator.ShardedSecret) (*appsv1.StatefulSetSpec, error) {
	if tr.Spec.QueryConfig == nil && len(tr.Spec.QueryEndpoints) < 1 {
		return nil, errors.New(tr.GetName() + ": thanos ruler requires query config or at least one query endpoint to be specified")
	}

	thanosVersion := operator.StringValOrDefault(ptr.Deref(tr.Spec.Version, ""), operator.DefaultThanosVersion)

	version, err := semver.ParseTolerant(thanosVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse thanos ruler version: %w", err)
	}

	trImagePath, err := operator.BuildImagePath(
		tr.Spec.Image,
		operator.StringValOrDefault(config.ThanosDefaultBaseImage, operator.DefaultThanosBaseImage),
		thanosVersion,
		"",
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build image path: %w", err)
	}

	trCLIArgs := []monitoringv1.Argument{
		{Name: "data-dir", Value: storageDir},
		{Name: "eval-interval", Value: string(tr.Spec.EvaluationInterval)},
		{Name: "tsdb.retention", Value: string(tr.Spec.Retention)},
	}

	trEnvVars := []v1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
	}

	trVolumes := []v1.Volume{}
	trVolumeMounts := []v1.VolumeMount{}

	trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "label", Value: fmt.Sprintf(`%s="$(POD_NAME)"`, defaultReplicaLabelName)})
	labels := operator.Map(tr.Spec.Labels)
	for _, k := range labels.SortedKeys() {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "label", Value: fmt.Sprintf(`%s="%s"`, k, labels[k])})
	}

	trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alert.label-drop", Value: defaultReplicaLabelName})
	for _, lb := range tr.Spec.AlertDropLabels {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alert.label-drop", Value: lb})
	}

	ports := []v1.ContainerPort{
		{
			Name:          "grpc",
			ContainerPort: 10901,
			Protocol:      v1.ProtocolTCP,
		},
	}
	if tr.Spec.ListenLocal {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "http-address", Value: "localhost:10902"})
	} else {
		ports = append(ports,
			v1.ContainerPort{
				Name:          tr.Spec.PortName,
				ContainerPort: 10902,
				Protocol:      v1.ProtocolTCP,
			})
	}

	if tr.Spec.LogLevel != "" && tr.Spec.LogLevel != "info" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "log.level", Value: tr.Spec.LogLevel})
	}
	if tr.Spec.LogFormat != "" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "log.format", Value: tr.Spec.LogFormat})
	}

	rulePath := rulesDir + "/*/*.yaml"
	trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "rule-file", Value: rulePath})

	if tr.Spec.QueryConfig != nil {
		var fullPath string
		trVolumes, trVolumeMounts, fullPath = mountSecretKey(trVolumes, trVolumeMounts, tr.Spec.QueryConfig, "query-config")
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "query.config-file", Value: fullPath})
	} else if len(tr.Spec.QueryEndpoints) > 0 {
		for _, endpoint := range tr.Spec.QueryEndpoints {
			trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "query", Value: endpoint})
		}
	}

	if tr.Spec.AlertManagersConfig != nil {
		var fullPath string
		trVolumes, trVolumeMounts, fullPath = mountSecretKey(trVolumes, trVolumeMounts, tr.Spec.AlertManagersConfig, "alertmanager-config")
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alertmanagers.config-file", Value: fullPath})
	} else if len(tr.Spec.AlertManagersURL) > 0 {
		for _, url := range tr.Spec.AlertManagersURL {
			trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alertmanagers.url", Value: url})
		}
	}

	if tr.Spec.ObjectStorageConfigFile != nil {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "objstore.config-file", Value: *tr.Spec.ObjectStorageConfigFile})
	} else if tr.Spec.ObjectStorageConfig != nil {
		var fullPath string
		trVolumes, trVolumeMounts, fullPath = mountSecretKey(trVolumes, trVolumeMounts, tr.Spec.ObjectStorageConfig, "objstorage-config")
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "objstore.config-file", Value: fullPath})
	}

	if tr.Spec.TracingConfigFile != "" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "tracing.config-file", Value: tr.Spec.TracingConfigFile})
	} else if tr.Spec.TracingConfig != nil {
		var fullPath string
		trVolumes, trVolumeMounts, fullPath = mountSecretKey(trVolumes, trVolumeMounts, tr.Spec.TracingConfig, "tracing-config")
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "tracing.config-file", Value: fullPath})
	}

	if tr.Spec.AlertRelabelConfigFile != nil {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alert.relabel-config-file", Value: *tr.Spec.AlertRelabelConfigFile})
	} else if tr.Spec.AlertRelabelConfigs != nil {
		var fullPath string
		trVolumes, trVolumeMounts, fullPath = mountSecretKey(trVolumes, trVolumeMounts, tr.Spec.AlertRelabelConfigs, "alertrelabel-config")
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alert.relabel-config-file", Value: fullPath})
	}

	trVolumes = append(trVolumes, tlsSecrets.Volume("tls-assets"))
	trVolumeMounts = append(trVolumeMounts, v1.VolumeMount{
		Name:      "tls-assets",
		ReadOnly:  true,
		MountPath: tlsAssetsDir,
	})

	isHTTPS := tr.Spec.Web != nil && tr.Spec.Web.TLSConfig != nil && version.GTE(semver.MustParse("0.21.0"))

	thanosrulerURIScheme := "http"
	if isHTTPS {
		thanosrulerURIScheme = "https"
	}

	if tr.Spec.GRPCServerTLSConfig != nil {
		tls := tr.Spec.GRPCServerTLSConfig
		if tls.CertFile != "" {
			trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "grpc-server-tls-cert", Value: tls.CertFile})
		}
		if tls.KeyFile != "" {
			trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "grpc-server-tls-key", Value: tls.KeyFile})
		}
		if tls.CAFile != "" {
			trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "grpc-server-tls-client-ca", Value: tls.CAFile})
		}
	}

	if tr.Spec.ExternalPrefix != "" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "web.external-prefix", Value: tr.Spec.ExternalPrefix})
	}

	if tr.Spec.RoutePrefix != "" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "web.route-prefix", Value: tr.Spec.RoutePrefix})
	}

	if tr.Spec.AlertQueryURL != "" {
		trCLIArgs = append(trCLIArgs, monitoringv1.Argument{Name: "alert.query-url", Value: tr.Spec.AlertQueryURL})
	}

	containerArgs, err := operator.BuildArgs(trCLIArgs, tr.Spec.AdditionalArgs)
	if err != nil {
		return nil, err
	}

	// The first argument to thanos must be "rule" to start thanos ruler, e.g. "thanos rule --data-dir..."
	containerArgs = append([]string{"rule"}, containerArgs...)

	var configReloaderWebConfigFile string
	var additionalContainers []v1.Container
	if len(ruleConfigMapNames) != 0 {
		var (
			watchedDirectories         []string
			configReloaderVolumeMounts []v1.VolumeMount
		)

		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			watchedDirectories = append(watchedDirectories, mountPath)
		}

		if version.GTE(semver.MustParse("0.21.0")) {
			var fields monitoringv1.WebConfigFileFields
			if tr.Spec.Web != nil {
				fields = tr.Spec.Web.WebConfigFileFields
			}

			webConfig, err := webconfig.New(webConfigDir, webConfigSecretName(tr.Name), fields)
			if err != nil {
				return nil, err
			}

			confArg, configVol, configMount, err := webConfig.GetMountParameters()
			if err != nil {
				return nil, err
			}
			containerArgs = append(containerArgs, fmt.Sprintf("--http.config=%s", confArg.Value))
			trVolumes = append(trVolumes, configVol...)
			trVolumeMounts = append(trVolumeMounts, configMount...)

			configReloaderWebConfigFile = confArg.Value
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, configMount...)
		}

		additionalContainers = append(
			additionalContainers,
			operator.CreateConfigReloader(
				"config-reloader",
				operator.ReloaderConfig(config.ReloaderConfig),
				operator.WebConfigFile(configReloaderWebConfigFile),
				operator.ReloaderURL(url.URL{
					Scheme: thanosrulerURIScheme,
					Host:   config.LocalHost + ":10902",
					Path:   path.Clean(tr.Spec.RoutePrefix + "/-/reload"),
				}),
				operator.ListenLocal(tr.Spec.ListenLocal),
				operator.LocalHost(config.LocalHost),
				operator.LogFormat(tr.Spec.LogFormat),
				operator.LogLevel(tr.Spec.LogLevel),
				operator.WatchedDirectories(watchedDirectories),
				operator.VolumeMounts(configReloaderVolumeMounts),
				operator.Shard(-1),
			),
		)
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	if tr.Spec.PodMetadata != nil {
		for k, v := range tr.Spec.PodMetadata.Labels {
			podLabels[k] = v
		}
		for k, v := range tr.Spec.PodMetadata.Annotations {
			podAnnotations[k] = v
		}
	}
	// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podLabels["app.kubernetes.io/name"] = thanosRulerLabel
	podLabels["app.kubernetes.io/managed-by"] = "prometheus-operator"
	podLabels["app.kubernetes.io/instance"] = tr.Name
	podLabels[thanosRulerLabel] = tr.Name
	finalLabels := config.Labels.Merge(podLabels)

	podAnnotations["kubectl.kubernetes.io/default-container"] = "thanos-ruler"

	storageVolName := volumeName(tr.Name)
	if tr.Spec.Storage != nil {
		if tr.Spec.Storage.VolumeClaimTemplate.Name != "" {
			storageVolName = tr.Spec.Storage.VolumeClaimTemplate.Name
		}
	}
	trVolumeMounts = append(trVolumeMounts, v1.VolumeMount{
		Name:      storageVolName,
		MountPath: storageDir,
	})

	for _, name := range ruleConfigMapNames {
		trVolumes = append(trVolumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
				},
			},
		})
		trVolumeMounts = append(trVolumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: rulesDir + "/" + name,
		})
	}

	trVolumeMounts = append(trVolumeMounts, tr.Spec.VolumeMounts...)

	operatorContainers := append([]v1.Container{
		{
			Name:                     "thanos-ruler",
			Image:                    trImagePath,
			ImagePullPolicy:          tr.Spec.ImagePullPolicy,
			Args:                     containerArgs,
			Env:                      trEnvVars,
			VolumeMounts:             trVolumeMounts,
			Resources:                tr.Spec.Resources,
			Ports:                    ports,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
			SecurityContext: &v1.SecurityContext{
				AllowPrivilegeEscalation: ptr.To(false),
				ReadOnlyRootFilesystem:   ptr.To(true),
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
			},
		},
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, tr.Spec.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	var minReadySeconds int32
	if tr.Spec.MinReadySeconds != nil {
		minReadySeconds = int32(*tr.Spec.MinReadySeconds)
	}

	spec := appsv1.StatefulSetSpec{
		ServiceName:     governingServiceName,
		Replicas:        tr.Spec.Replicas,
		MinReadySeconds: minReadySeconds,
		// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
		// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
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
				NodeSelector:                  tr.Spec.NodeSelector,
				PriorityClassName:             tr.Spec.PriorityClassName,
				ServiceAccountName:            tr.Spec.ServiceAccountName,
				TerminationGracePeriodSeconds: ptr.To(int64(120)),
				Containers:                    containers,
				InitContainers:                tr.Spec.InitContainers,
				Volumes:                       trVolumes,
				SecurityContext:               tr.Spec.SecurityContext,
				Tolerations:                   tr.Spec.Tolerations,
				Affinity:                      tr.Spec.Affinity,
				TopologySpreadConstraints:     tr.Spec.TopologySpreadConstraints,
				HostAliases:                   operator.MakeHostAliases(tr.Spec.HostAliases),
			},
		},
	}

	k8sutil.UpdateDNSConfig(&spec.Template.Spec, tr.Spec.DNSConfig)
	k8sutil.UpdateDNSPolicy(&spec.Template.Spec, tr.Spec.DNSPolicy)

	return &spec, nil
}

func makeStatefulSetService(tr *monitoringv1.ThanosRuler, config Config) *v1.Service {
	if tr.Spec.PortName == "" {
		tr.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       tr.Spec.PortName,
					Port:       10902,
					TargetPort: intstr.FromString(tr.Spec.PortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "grpc",
					Port:       10901,
					TargetPort: intstr.FromString("grpc"),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": thanosRulerLabel,
			},
		},
	}

	operator.UpdateObject(
		svc,
		operator.WithName(governingServiceName),
		operator.WithAnnotations(config.Annotations),
		operator.WithLabels(map[string]string{"operated-thanos-ruler": "true"}),
		operator.WithLabels(config.Labels),
		operator.WithOwner(tr),
	)

	return svc
}

func prefixedName(name string) string {
	return fmt.Sprintf("thanos-ruler-%s", name)
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-data", prefixedName(name))
}

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}

func webConfigSecretName(name string) string {
	return fmt.Sprintf("%s-web-config", prefixedName(name))
}

// mountSecretKey adds the secret key to the mounted volumes and returns the
// full path of the file on disk.
func mountSecretKey(vols []v1.Volume, vmounts []v1.VolumeMount, secretSelector *v1.SecretKeySelector, volumeName string) ([]v1.Volume, []v1.VolumeMount, string) {
	mountpath := filepath.Join(configDir, volumeName)

	return append(vols, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secretSelector.Name,
					Items: []v1.KeyToPath{
						{
							Key:  secretSelector.Key,
							Path: secretSelector.Key,
						},
					},
				},
			},
		}),
		append(vmounts, v1.VolumeMount{
			Name:      volumeName,
			MountPath: mountpath,
		}),
		filepath.Join(mountpath, secretSelector.Key)
}
