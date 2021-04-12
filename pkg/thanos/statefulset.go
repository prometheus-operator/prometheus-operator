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

	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	rulesDir                  = "/etc/thanos/rules"
	storageDir                = "/thanos/data"
	governingServiceName      = "thanos-ruler-operated"
	defaultPortName           = "web"
	defaultRetention          = "24h"
	defaultEvaluationInterval = "15s"
	defaultReplicaLabelName   = "thanos_ruler_replica"
	sSetInputHashName         = "prometheus-operator-input-hash"
)

var (
	minReplicas                 int32 = 1
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
)

func makeStatefulSet(tr *monitoringv1.ThanosRuler, config Config, ruleConfigMapNames []string, inputHash string) (*appsv1.StatefulSet, error) {

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

	if statefulset.ObjectMeta.Annotations == nil {
		statefulset.ObjectMeta.Annotations = map[string]string{
			sSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[sSetInputHashName] = inputHash
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

func makeStatefulSetSpec(tr *monitoringv1.ThanosRuler, config Config, ruleConfigMapNames []string) (*appsv1.StatefulSetSpec, error) {
	// Before editing 'tr' create deep copy, to prevent side effects. For more
	// details see https://github.com/prometheus-operator/prometheus-operator/issues/1659
	tr = tr.DeepCopy()

	if tr.Spec.QueryConfig == nil && len(tr.Spec.QueryEndpoints) < 1 {
		return nil, errors.New(tr.GetName() + ": thanos ruler requires query config or at least one query endpoint to be specified")
	}

	trImagePath, err := operator.BuildImagePath(
		tr.Spec.Image,
		operator.StringValOrDefault(config.ThanosDefaultBaseImage, operator.DefaultThanosBaseImage),
		operator.DefaultThanosVersion,
		"",
		"",
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build image path")
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

	ports := []v1.ContainerPort{
		{
			Name:          "grpc",
			ContainerPort: 10901,
			Protocol:      v1.ProtocolTCP,
		},
	}
	if tr.Spec.ListenLocal {
		trCLIArgs = append(trCLIArgs, "--http-address=localhost:10902")
	} else {
		ports = append(ports,
			v1.ContainerPort{
				Name:          tr.Spec.PortName,
				ContainerPort: 10902,
				Protocol:      v1.ProtocolTCP,
			})
	}

	if tr.Spec.LogLevel != "" && tr.Spec.LogLevel != "info" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--log.level=%s", tr.Spec.LogLevel))
	}
	if tr.Spec.LogFormat != "" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--log.format=%s", tr.Spec.LogFormat))
	}

	if tr.Spec.QueryConfig != nil {
		trCLIArgs = append(trCLIArgs, "--query.config=$(QUERY_CONFIG)")
		trEnvVars = append(trEnvVars, v1.EnvVar{
			Name: "QUERY_CONFIG",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: tr.Spec.QueryConfig,
			},
		})
	} else if len(tr.Spec.QueryEndpoints) > 0 {
		for _, endpoint := range tr.Spec.QueryEndpoints {
			trCLIArgs = append(trCLIArgs, fmt.Sprintf("--query=%s", endpoint))
		}
	}

	rulePath := rulesDir + "/*/*.yaml"
	trCLIArgs = append(trCLIArgs, fmt.Sprintf("--rule-file=%s", rulePath))

	if tr.Spec.AlertManagersConfig != nil {
		trCLIArgs = append(trCLIArgs, "--alertmanagers.config=$(ALERTMANAGERS_CONFIG)")
		trEnvVars = append(trEnvVars, v1.EnvVar{
			Name: "ALERTMANAGERS_CONFIG",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: tr.Spec.AlertManagersConfig,
			},
		})
	} else if len(tr.Spec.AlertManagersURL) > 0 {
		for _, url := range tr.Spec.AlertManagersURL {
			trCLIArgs = append(trCLIArgs, fmt.Sprintf("--alertmanagers.url=%s", url))
		}
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

	if tr.Spec.ObjectStorageConfigFile != nil {
		trCLIArgs = append(trCLIArgs, "--objstore.config-file="+*tr.Spec.ObjectStorageConfigFile)
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

	if tr.Spec.GRPCServerTLSConfig != nil {
		tls := tr.Spec.GRPCServerTLSConfig
		if tls.CertFile != "" {
			trCLIArgs = append(trCLIArgs, "--grpc-server-tls-cert="+tls.CertFile)
		}
		if tls.KeyFile != "" {
			trCLIArgs = append(trCLIArgs, "--grpc-server-tls-key="+tls.KeyFile)
		}
		if tls.CAFile != "" {
			trCLIArgs = append(trCLIArgs, "--grpc-server-tls-client-ca="+tls.CAFile)
		}
	}

	if tr.Spec.ExternalPrefix != "" {
		trCLIArgs = append(trCLIArgs, "--web.external-prefix="+tr.Spec.ExternalPrefix)
	}

	if tr.Spec.RoutePrefix != "" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--web.route-prefix=%s", tr.Spec.RoutePrefix))
	}

	if tr.Spec.AlertQueryURL != "" {
		trCLIArgs = append(trCLIArgs, fmt.Sprintf("--alert.query-url=%s", tr.Spec.AlertQueryURL))
	}

	var additionalContainers []v1.Container
	if len(ruleConfigMapNames) != 0 {
		var (
			configReloaderArgs         []string
			configReloaderVolumeMounts []v1.VolumeMount
		)

		for _, name := range ruleConfigMapNames {
			mountPath := rulesDir + "/" + name
			configReloaderVolumeMounts = append(configReloaderVolumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: mountPath,
			})
			configReloaderArgs = append(configReloaderArgs, fmt.Sprintf("--watched-dir=%s", mountPath))
		}

		additionalContainers = append(
			additionalContainers,
			operator.CreateConfigReloader(
				config.ReloaderConfig,
				url.URL{
					Scheme: "http",
					Host:   config.LocalHost + ":10902",
					Path:   path.Clean(tr.Spec.RoutePrefix + "/-/reload"),
				},
				tr.Spec.ListenLocal,
				config.LocalHost,
				tr.Spec.LogFormat,
				tr.Spec.LogLevel,
				configReloaderArgs,
				configReloaderVolumeMounts,
				-1,
			),
		)
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
	podLabels["app"] = thanosRulerLabel
	podLabels["app.kubernetes.io/name"] = thanosRulerLabel
	podLabels["app.kubernetes.io/managed-by"] = "prometheus-operator"
	podLabels["app.kubernetes.io/instance"] = tr.Name
	podLabels[thanosRulerLabel] = tr.Name
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
			Name:                     "thanos-ruler",
			Image:                    trImagePath,
			Args:                     trCLIArgs,
			Env:                      trEnvVars,
			VolumeMounts:             trVolumeMounts,
			Resources:                tr.Spec.Resources,
			Ports:                    ports,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}, additionalContainers...)

	containers, err := k8sutil.MergePatchContainers(operatorContainers, tr.Spec.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	terminationGracePeriod := int64(120)

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
				NodeSelector:                  tr.Spec.NodeSelector,
				PriorityClassName:             tr.Spec.PriorityClassName,
				ServiceAccountName:            tr.Spec.ServiceAccountName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Containers:                    containers,
				InitContainers:                tr.Spec.InitContainers,
				Volumes:                       trVolumes,
				SecurityContext:               tr.Spec.SecurityContext,
				Tolerations:                   tr.Spec.Tolerations,
				Affinity:                      tr.Spec.Affinity,
				TopologySpreadConstraints:     tr.Spec.TopologySpreadConstraints,
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
