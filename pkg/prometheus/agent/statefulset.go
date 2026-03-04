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
	"maps"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

const (
	prometheusMode       = "agent"
	governingServiceName = "prometheus-agent-operated"
)

func makeStatefulSet(
	name string,
	p *monitoringv1alpha1.PrometheusAgent,
	config prompkg.Config,
	cg *prompkg.ConfigGenerator,
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
	spec, err := makeStatefulSetSpec(p, config, cg, shard, tlsSecrets)
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
	p *monitoringv1alpha1.PrometheusAgent,
	c prompkg.Config,
	cg *prompkg.ConfigGenerator,
	shard int32,
	tlsSecrets *operator.ShardedSecret,
) (*appsv1.StatefulSetSpec, error) {
	cpf := p.GetCommonPrometheusFields()

	pImagePath, err := operator.BuildImagePathForAgent(
		ptr.Deref(cpf.Image, ""),
		c.PrometheusDefaultBaseImage,
		"v"+cg.Version().String(),
	)
	if err != nil {
		return nil, err
	}

	promArgs := buildAgentArgs(cg, cpf.WALCompression)

	volumes, promVolumeMounts, err := prompkg.BuildCommonVolumes(p, tlsSecrets, true)
	if err != nil {
		return nil, err
	}

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

	var watchedDirectories []string

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

	operatorContainers := append([]v1.Container{
		{
			Name:                     "prometheus",
			Image:                    pImagePath,
			ImagePullPolicy:          cpf.ImagePullPolicy,
			Ports:                    prompkg.MakeContainerPorts(cpf),
			Args:                     containerArgs,
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

	spec := v1.PodSpec{
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
	}

	if cpf.HostNetwork {
		spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}
	k8s.UpdateDNSPolicy(&spec, cpf.DNSPolicy)
	k8s.UpdateDNSConfig(&spec, cpf.DNSConfig)

	// By default, podManagementPolicy is set to Parallel to mitigate rollout
	// issues in Kubernetes (see https://github.com/kubernetes/kubernetes/issues/60164).
	// This is also mentioned as one of limitations of StatefulSets:
	// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	podManagementPolicy := ptr.Deref(cpf.PodManagementPolicy, monitoringv1.ParallelPodManagement)

	return &appsv1.StatefulSetSpec{
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
			Spec: spec,
		},
	}, nil
}

// buildAgentArgs returns the CLI arguments that are only valid for the Prometheus agent.
func buildAgentArgs(cg *prompkg.ConfigGenerator, walCompression *bool) []monitoringv1.Argument {
	promArgs := cg.BuildCommonPrometheusArgs()

	switch cg.Version().Major {
	case 2:
		promArgs = append(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: "agent"})
	case 3:
		promArgs = append(promArgs, monitoringv1.Argument{Name: "agent"})
	}

	promArgs = append(promArgs,
		monitoringv1.Argument{Name: "storage.agent.path", Value: prompkg.StorageDir},
	)

	if walCompression != nil {
		arg := monitoringv1.Argument{Name: "no-storage.agent.wal-compression"}
		if *walCompression {
			arg.Name = "storage.agent.wal-compression"
		}
		promArgs = cg.AppendCommandlineArgument(promArgs, arg)
	}

	return promArgs
}
