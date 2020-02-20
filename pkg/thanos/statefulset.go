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
	"fmt"
	"net/url"
	"path"
	"strings"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	DefaultThanosVersion      = "v0.10.1"
	rulesDir                  = "/etc/thanos/rules"
	storageDir                = "/thanos/data"
	governingServiceName      = "thanos-ruler-operated"
	defaultPortName           = "web"
	defaultRetention          = "24h"
	defaultEvaluationInterval = "15s"
	defaultReplicaLabelName   = "thanos_ruler_replica"
)

var (
	minReplicas                 int32 = 1
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
)

func makeStatefulSet(tr *monitoringv1.ThanosRuler, old *appsv1.StatefulSet, config Config, ruleConfigMapNames []string) (*appsv1.StatefulSet, error) {

	if tr.Spec.Image == "" {
		tr.Spec.Image = config.ThanosDefaultBaseImage
	}
	if !strings.Contains(tr.Spec.Image, ":") {
		tr.Spec.Image = tr.Spec.Image + ":" + DefaultThanosVersion
	}
	if tr.Spec.Resources.Requests == nil {
		tr.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := tr.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		tr.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("200Mi")
	}

	spec, err := makeStatefulSetSpec(tr, config, ruleConfigMapNames)
	if err != nil {
		return nil, err
	}

	boolTrue := true
	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	annotations := make(map[string]string)
	for key, value := range tr.ObjectMeta.Annotations {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(tr.Name),
			Labels:      config.Labels.Merge(tr.ObjectMeta.Labels),
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         tr.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               tr.Kind,
					Name:               tr.Name,
					UID:                tr.UID,
				},
			},
		},
		Spec: *spec,
	}

	if tr.Spec.ImagePullSecrets != nil && len(tr.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = tr.Spec.ImagePullSecrets
	}

	if old != nil {
		statefulset.Annotations = old.Annotations
	}

	storageSpec := tr.Spec.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(tr.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(tr.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else {
		pvcTemplate := storageSpec.VolumeClaimTemplate
		pvcTemplate.CreationTimestamp = metav1.Time{}
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
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvcTemplate)
	}

	for _, volume := range tr.Spec.Volumes {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, volume)
	}

	return statefulset, nil
}

func makeStatefulSetSpec(tr *monitoringv1.ThanosRuler, config Config, ruleConfigMapNames []string) (*appsv1.StatefulSetSpec, error) {
	// Before editing 'tr' create deep copy, to prevent side effects. For more
	// details see https://github.com/coreos/prometheus-operator/issues/1659
	tr = tr.DeepCopy()

	if len(tr.Spec.QueryEndpoints) < 1 {
		return nil, errors.New(tr.GetName() + ": thanos ruler requires at least one query endpoint")
	}

	if tr.Spec.EvaluationInterval == "" {
		tr.Spec.EvaluationInterval = defaultEvaluationInterval
	}
	if tr.Spec.Retention == "" {
		tr.Spec.Retention = defaultRetention
	}

	trCLIArgs := []string{
		"rule",
		fmt.Sprintf("--data-dir=%s", storageDir),
		fmt.Sprintf("--eval-interval=%s", tr.Spec.EvaluationInterval),
		fmt.Sprintf("--tsdb.retention=%s", tr.Spec.Retention),
	}
	trEnvVars := []v1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
	}

	if len(tr.Spec.Labels) == 0 {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf(`--label=%s="$(POD_NAME)"`, defaultReplicaLabelName))
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--alert.label-drop=%s", defaultReplicaLabelName))
	} else {
		for k, v := range tr.Spec.Labels {
			trCLIArgs = append(trCLIArgs, fmt.Sprintf(`--label=%s="%s"`, k, v))
		}
	}

	for _, lb := range tr.Spec.AlertDropLabels {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--alert.label-drop=%s", lb))
	}

	if tr.Spec.ListenLocal {
		trCLIArgs = append(trCLIArgs, "--http-address=localhost:10902")
	}
	if tr.Spec.LogLevel != "" && tr.Spec.LogLevel != "info" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--log.level=%s", tr.Spec.LogLevel))
	}
	if tr.Spec.LogFormat != "" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--log.format=%s", tr.Spec.LogFormat))
	}
	for _, endpoint := range tr.Spec.QueryEndpoints {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--query=%s", endpoint))
	}
	for _, ruleConfigMapName := range ruleConfigMapNames {
		rulePath := rulesDir + "/" + ruleConfigMapName + "/*.yaml"
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--rule-file=%s", rulePath))
	}

	if tr.Spec.AlertManagersConfig != nil {
		trCLIArgs = append(trCLIArgs, "--alertmanagers.config=$(ALERTMANAGERS_CONFIG)")
		trEnvVars = append(trEnvVars, v1.EnvVar{
			Name: "ALERTMANAGERS_CONFIG",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: tr.Spec.AlertManagersConfig,
			},
		})
	} else if tr.Spec.AlertManagersURL != "" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--alertmanagers.url=%s", tr.Spec.AlertManagersURL))
	}

	if tr.Spec.ObjectStorageConfig != nil {
		trCLIArgs = append(trCLIArgs, "--objstore.config=$(OBJSTORE_CONFIG)")
		trEnvVars = append(trEnvVars, v1.EnvVar{
			Name: "OBJSTORE_CONFIG",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: tr.Spec.ObjectStorageConfig,
			},
		})
	}

	if tr.Spec.TracingConfig != nil {
		trCLIArgs = append(trCLIArgs, "--tracing.config=$(TRACING_CONFIG)")
		trEnvVars = append(trEnvVars, v1.EnvVar{
			Name: "TRACING_CONFIG",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: tr.Spec.TracingConfig,
			},
		})
	}

	localReloadURL := &url.URL{
		Scheme: "http",
		Host:   config.LocalHost + ":10902",
		// TODO: add web prefix
		Path: path.Clean("/-/reload"),
	}

	additionalContainers := []v1.Container{}
	if len(ruleConfigMapNames) != 0 {
		reloader := v1.Container{
			Name:  "rules-configmap-reloader",
			Image: config.ConfigReloaderImage,
			Args: []string{
				fmt.Sprintf("--webhook-url=%s", localReloadURL),
			},
			VolumeMounts: []v1.VolumeMount{},
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{}, Requests: v1.ResourceList{}},
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		}

		if config.ConfigReloaderCPU != "0" {
			reloader.Resources.Limits[v1.ResourceCPU] = resource.MustParse(config.ConfigReloaderCPU)
			reloader.Resources.Requests[v1.ResourceCPU] = resource.MustParse(config.ConfigReloaderCPU)
		}
		if config.ConfigReloaderMemory != "0" {
			reloader.Resources.Limits[v1.ResourceMemory] = resource.MustParse(config.ConfigReloaderMemory)
			reloader.Resources.Requests[v1.ResourceMemory] = resource.MustParse(config.ConfigReloaderMemory)
		}

		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			reloader.VolumeMounts = append(reloader.VolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			reloader.Args = append(reloader.Args, fmt.Sprintf("--volume-dir=%s", mountPath))
		}
		additionalContainers = append(additionalContainers, reloader)
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	if tr.Spec.PodMetadata != nil {
		if tr.Spec.PodMetadata.Labels != nil {
			for k, v := range tr.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if tr.Spec.PodMetadata.Annotations != nil {
			for k, v := range tr.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}
	podLabels["app"] = "thanos-ruler"
	podLabels["thanos"] = tr.Name
	finalLabels := config.Labels.Merge(podLabels)

	storageVolName := volumeName(tr.Name)
	if tr.Spec.Storage != nil {
		if tr.Spec.Storage.VolumeClaimTemplate.Name != "" {
			storageVolName = tr.Spec.Storage.VolumeClaimTemplate.Name
		}
	}
	trVolumes := []v1.Volume{}
	trVolumeMounts := []v1.VolumeMount{
		{
			Name:      storageVolName,
			MountPath: storageDir,
		},
	}

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

	operatorContainers := append([]v1.Container{
		{
			Name:         "thanos-ruler",
			Image:        tr.Spec.Image,
			Args:         trCLIArgs,
			Env:          trEnvVars,
			VolumeMounts: trVolumeMounts,
			Resources:    tr.Spec.Resources,
			Ports: []v1.ContainerPort{
				{
					Name:          "grpc",
					ContainerPort: 10901,
					Protocol:      v1.ProtocolTCP,
				},
				{
					Name:          "http",
					ContainerPort: 10902,
					Protocol:      v1.ProtocolTCP,
				},
			},
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, tr.Spec.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		Replicas:            tr.Spec.Replicas,
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
				Containers:     containers,
				InitContainers: tr.Spec.InitContainers,
				Volumes:        trVolumes,
			},
		},
	}, nil
}

func makeStatefulSetService(tr *monitoringv1.ThanosRuler, config Config) *v1.Service {

	if tr.Spec.PortName == "" {
		tr.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			Labels: config.Labels.Merge(map[string]string{
				"operated-thanos-ruler": "true",
			}),
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					Name:       tr.GetName(),
					Kind:       tr.Kind,
					APIVersion: tr.APIVersion,
					UID:        tr.GetUID(),
				},
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       tr.Spec.PortName,
					Port:       10902,
					TargetPort: intstr.FromInt(10902),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "grpc",
					Port:       10901,
					TargetPort: intstr.FromInt(10901),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "thanos-ruler",
			},
		},
	}
	return svc
}

func prefixedName(name string) string {
	return fmt.Sprintf("thanos-ruler-%s", name)
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-data", prefixedName(name))
}
