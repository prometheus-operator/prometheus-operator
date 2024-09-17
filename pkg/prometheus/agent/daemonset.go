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

	"github.com/blang/semver/v4"
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

func makeDaemonSet(
	p monitoringv1.PrometheusInterface,
	config *prompkg.Config,
	cg *prompkg.ConfigGenerator,
	tlsSecrets *operator.ShardedSecret,
) (*appsv1.DaemonSet, error) {
	cpf := p.GetCommonPrometheusFields()
	objMeta := p.GetObjectMeta()

	if cpf.PortName == "" {
		cpf.PortName = prompkg.DefaultPortName
	}

	// We need to re-set the common fields because cpf is only a copy of the original object.
	// We set some defaults if some fields are not present, and we want those fields set in the original Prometheus object before building the DaemonSetSpec.
	p.SetCommonPrometheusFields(cpf)

	spec, err := makeDaemonSetSpec(p, config, cg, tlsSecrets)
	if err != nil {
		return nil, fmt.Errorf("make DaemonSet spec: %w", err)
	}
	daemonSet := &appsv1.DaemonSet{Spec: *spec}

	operator.UpdateObject(
		daemonSet,
		operator.WithName(prompkg.PrefixedName(p)),
		operator.WithAnnotations(objMeta.GetAnnotations()),
		operator.WithAnnotations(config.Annotations),
		operator.WithLabels(objMeta.GetLabels()),
		operator.WithLabels(map[string]string{
			prompkg.PrometheusNameLabelName: objMeta.GetName(),
			prompkg.PrometheusModeLabeLName: prometheusMode,
		}),
		operator.WithLabels(config.Labels),
		operator.WithManagingOwner(p),
		operator.WithoutKubectlAnnotations(),
	)

	if len(cpf.ImagePullSecrets) > 0 {
		daemonSet.Spec.Template.Spec.ImagePullSecrets = cpf.ImagePullSecrets
	}

	if cpf.HostNetwork {
		daemonSet.Spec.Template.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}

	return daemonSet, nil
}

func makeDaemonSetSpec(
	p monitoringv1.PrometheusInterface,
	c *prompkg.Config,
	cg *prompkg.ConfigGenerator,
	tlsSecrets *operator.ShardedSecret,
) (*appsv1.DaemonSetSpec, error) {
	cpf := p.GetCommonPrometheusFields()

	promVersion := operator.StringValOrDefault(cpf.Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	pImagePath, err := operator.BuildImagePathForAgent(
		ptr.Deref(cpf.Image, ""),
		c.PrometheusDefaultBaseImage,
		promVersion,
	)
	if err != nil {
		return nil, err
	}

	if version.Major == 2 && !slices.Contains(cpf.EnableFeatures, "agent") {
		cpf.EnableFeatures = append(cpf.EnableFeatures, "agent")
	}

	promArgs := buildAgentArgs(cpf, cg)

	volumes, promVolumeMounts, err := prompkg.BuildCommonVolumes(p, tlsSecrets, false)
	if err != nil {
		return nil, err
	}

	configReloaderVolumeMounts := prompkg.CreateConfigReloaderVolumeMounts()

	var configReloaderWebConfigFile string

	confArg, configVol, configMount, err := prompkg.BuildWebconfig(cpf, p)
	if err != nil {
		return nil, err
	}

	promArgs = append(promArgs, confArg)
	volumes = append(volumes, configVol...)
	promVolumeMounts = append(promVolumeMounts, configMount...)

	configReloaderWebConfigFile = confArg.Value
	configReloaderVolumeMounts = append(configReloaderVolumeMounts, configMount...)

	startupProbe, readinessProbe, livenessProbe := prompkg.MakeProbes(cpf, cg)

	podAnnotations, podLabels := prompkg.BuildPodMetadata(cpf, cg)
	// In cases where an existing selector label is modified, or a new one is added, new daemonset cannot match existing pods.
	// We should try to avoid removing such immutable fields whenever possible since doing
	// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
	// The requirement to make a change here should be carefully evaluated.
	podSelectorLabels := makeSelectorLabels(p.GetObjectMeta().GetName())

	for k, v := range podSelectorLabels {
		podLabels[k] = v
	}

	finalSelectorLabels := c.Labels.Merge(podSelectorLabels)
	finalLabels := c.Labels.Merge(podLabels)

	var additionalContainers, operatorInitContainers []v1.Container

	var watchedDirectories []string

	var minReadySeconds int32
	if cpf.MinReadySeconds != nil {
		minReadySeconds = int32(*cpf.MinReadySeconds)
	}

	operatorInitContainers = append(operatorInitContainers,
		prompkg.BuildConfigReloader(
			p,
			c,
			true,
			configReloaderVolumeMounts,
			watchedDirectories,
			operator.WithDaemonSetMode(),
		),
	)

	initContainers, err := k8sutil.MergePatchContainers(operatorInitContainers, cpf.InitContainers)
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
			operator.WebConfigFile(configReloaderWebConfigFile),
			operator.WithDaemonSetMode(),
		),
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, cpf.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	return &appsv1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: finalSelectorLabels,
		},
		MinReadySeconds: minReadySeconds,
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				ShareProcessNamespace:        prompkg.ShareProcessNamespace(p),
				Containers:                   containers,
				InitContainers:               initContainers,
				SecurityContext:              cpf.SecurityContext,
				ServiceAccountName:           cpf.ServiceAccountName,
				AutomountServiceAccountToken: ptr.To(ptr.Deref(cpf.AutomountServiceAccountToken, true)),
				NodeSelector:                 cpf.NodeSelector,
				PriorityClassName:            cpf.PriorityClassName,
				// Prometheus may take quite long to shut down to checkpoint existing data.
				// Allow up to 10 minutes for clean termination.
				TerminationGracePeriodSeconds: ptr.To(int64(600)),
				Volumes:                       volumes,
				Tolerations:                   cpf.Tolerations,
				Affinity:                      cpf.Affinity,
				TopologySpreadConstraints:     prompkg.MakeK8sTopologySpreadConstraint(finalSelectorLabels, cpf.TopologySpreadConstraints),
				HostAliases:                   operator.MakeHostAliases(cpf.HostAliases),
				HostNetwork:                   cpf.HostNetwork,
			},
		},
	}, nil
}
