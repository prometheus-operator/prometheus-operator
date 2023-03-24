// Copyright 2023 The prometheus-operator Authors
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

package prometheusagent

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-kit/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	prometheusMode       = "agent"
	governingServiceName = "prometheus-agent-operated"
)

func makeStatefulSet(
	logger log.Logger,
	name string,
	p monitoringv1.PrometheusInterface,
	config *operator.Config,
	cg *prompkg.ConfigGenerator,
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
	spec, err := makeStatefulSetSpec(logger, p, config, cg, shard, tlsAssetSecrets)
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
	p monitoringv1.PrometheusInterface,
	c *operator.Config,
	cg *prompkg.ConfigGenerator,
	shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSetSpec, error) {
	// Prometheus may take quite long to shut down to checkpoint existing data.
	// Allow up to 10 minutes for clean termination.
	terminationGracePeriod := int64(600)
	cpf := p.GetCommonPrometheusFields()
	promName := p.GetObjectMeta().GetName()

	pImagePath, err := operator.BuildImagePath(
		operator.StringPtrValOrDefault(cpf.Image, ""),
		operator.StringValOrDefault("", c.PrometheusDefaultBaseImage),
		operator.StringValOrDefault(cpf.Version, operator.DefaultPrometheusVersion),
		"",
		"",
	)
	if err != nil {
		return nil, err
	}

	webRoutePrefix := "/"
	if cpf.RoutePrefix != "" {
		webRoutePrefix = cpf.RoutePrefix
	}

	cpf.EnableFeatures = append(cpf.EnableFeatures, "agent")
	promArgs := prompkg.BuildCommonPrometheusArgs(cpf, cg, webRoutePrefix)
	promArgs = appendAgentArgs(promArgs, cg)

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
		"app.kubernetes.io/name":        "prometheus-agent",
		"app.kubernetes.io/managed-by":  "prometheus-operator",
		"app.kubernetes.io/instance":    promName,
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

func makeStatefulSetService(p *monitoringv1alpha1.PrometheusAgent, config operator.Config) *v1.Service {
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
				"app.kubernetes.io/name": "prometheus-agent",
			},
		},
	}

	return svc
}

// appendAgentArgs appends arguments that are only valid for the Prometheus agent.
func appendAgentArgs(promArgs []monitoringv1.Argument, cg *prompkg.ConfigGenerator) []monitoringv1.Argument {

	promArgs = append(promArgs,
		monitoringv1.Argument{Name: "storage.agent.path", Value: prompkg.StorageDir},
	)

	return promArgs
}
