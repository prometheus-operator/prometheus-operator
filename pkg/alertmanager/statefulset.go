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
	"net/url"
	"path"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/blang/semver"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/operator"
	"github.com/pkg/errors"
)

const (
	governingServiceName   = "alertmanager-operated"
	defaultRetention       = "120h"
	secretsDir             = "/etc/alertmanager/secrets/"
	configmapsDir          = "/etc/alertmanager/configmaps/"
	alertmanagerConfDir    = "/etc/alertmanager/config"
	alertmanagerConfFile   = alertmanagerConfDir + "/alertmanager.yaml"
	alertmanagerStorageDir = "/alertmanager"
	defaultPortName        = "web"
)

var (
	minReplicas         int32 = 1
	probeTimeoutSeconds int32 = 3
)

func makeStatefulSet(am *monitoringv1.Alertmanager, old *appsv1.StatefulSet, config Config) (*appsv1.StatefulSet, error) {
	// TODO(fabxc): is this the right point to inject defaults?
	// Ideally we would do it before storing but that's currently not possible.
	// Potentially an update handler on first insertion.

	if am.Spec.BaseImage == "" {
		am.Spec.BaseImage = config.AlertmanagerDefaultBaseImage
	}
	if am.Spec.PortName == "" {
		am.Spec.PortName = defaultPortName
	}
	if am.Spec.Version == "" {
		am.Spec.Version = operator.DefaultAlertmanagerVersion
	}
	if am.Spec.Replicas == nil {
		am.Spec.Replicas = &minReplicas
	}
	intZero := int32(0)
	if am.Spec.Replicas != nil && *am.Spec.Replicas < 0 {
		am.Spec.Replicas = &intZero
	}
	if am.Spec.Retention == "" {
		am.Spec.Retention = defaultRetention
	}
	if am.Spec.Resources.Requests == nil {
		am.Spec.Resources.Requests = v1.ResourceList{}
	}
	if _, ok := am.Spec.Resources.Requests[v1.ResourceMemory]; !ok {
		am.Spec.Resources.Requests[v1.ResourceMemory] = resource.MustParse("200Mi")
	}
	if am.Spec.ConfigSecret == "" {
		am.Spec.ConfigSecret = configSecretName(am.Name)
	}

	spec, err := makeStatefulSetSpec(am, config)
	if err != nil {
		return nil, err
	}

	boolTrue := true
	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	annotations := make(map[string]string)
	for key, value := range am.ObjectMeta.Annotations {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(am.Name),
			Labels:      config.Labels.Merge(am.ObjectMeta.Labels),
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         am.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               am.Kind,
					Name:               am.Name,
					UID:                am.UID,
				},
			},
		},
		Spec: *spec,
	}

	if am.Spec.ImagePullSecrets != nil && len(am.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = am.Spec.ImagePullSecrets
	}

	storageSpec := am.Spec.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(am.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else {
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

	if old != nil {
		statefulset.Annotations = old.Annotations
	}

	for _, volume := range am.Spec.Volumes {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, volume)
	}

	return statefulset, nil
}

func makeStatefulSetService(p *monitoringv1.Alertmanager, config Config) *v1.Service {

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			Labels: config.Labels.Merge(map[string]string{
				"operated-alertmanager": "true",
			}),
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					Name:       p.GetName(),
					Kind:       p.Kind,
					APIVersion: p.APIVersion,
					UID:        p.GetUID(),
				},
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       p.Spec.PortName,
					Port:       9093,
					TargetPort: intstr.FromString(p.Spec.PortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "tcp-mesh",
					Port:       9094,
					TargetPort: intstr.FromInt(9094),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "udp-mesh",
					Port:       9094,
					TargetPort: intstr.FromInt(9094),
					Protocol:   v1.ProtocolUDP,
				},
			},
			Selector: map[string]string{
				"app": "alertmanager",
			},
		},
	}
	return svc
}

func makeStatefulSetSpec(a *monitoringv1.Alertmanager, config Config) (*appsv1.StatefulSetSpec, error) {
	// Before editing 'a' create deep copy, to prevent side effects. For more
	// details see https://github.com/coreos/prometheus-operator/issues/1659
	a = a.DeepCopy()

	// Version is used by default.
	// If the tag is specified, we use the tag to identify the container image.
	// If the sha is specified, we use the sha to identify the container image,
	// as it has even stronger immutable guarantees to identify the image.
	image := fmt.Sprintf("%s:%s", a.Spec.BaseImage, a.Spec.Version)
	if a.Spec.Tag != "" {
		image = fmt.Sprintf("%s:%s", a.Spec.BaseImage, a.Spec.Tag)
	}
	if a.Spec.SHA != "" {
		image = fmt.Sprintf("%s@sha256:%s", a.Spec.BaseImage, a.Spec.SHA)
	}
	if a.Spec.Image != nil && *a.Spec.Image != "" {
		image = *a.Spec.Image
	}

	version, err := semver.ParseTolerant(a.Spec.Version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse alertmanager version")
	}

	amArgs := []string{
		fmt.Sprintf("--config.file=%s", alertmanagerConfFile),
		fmt.Sprintf("--storage.path=%s", alertmanagerStorageDir),
		fmt.Sprintf("--data.retention=%s", a.Spec.Retention),
	}

	if *a.Spec.Replicas == 1 {
		amArgs = append(amArgs, "--cluster.listen-address=")
	} else {
		amArgs = append(amArgs, "--cluster.listen-address=[$(POD_IP)]:9094")
	}

	if a.Spec.ListenLocal {
		amArgs = append(amArgs, "--web.listen-address=127.0.0.1:9093")
	} else {
		amArgs = append(amArgs, "--web.listen-address=:9093")
	}

	if a.Spec.ExternalURL != "" {
		amArgs = append(amArgs, "--web.external-url="+a.Spec.ExternalURL)
	}

	webRoutePrefix := "/"
	if a.Spec.RoutePrefix != "" {
		webRoutePrefix = a.Spec.RoutePrefix
	}
	amArgs = append(amArgs, fmt.Sprintf("--web.route-prefix=%v", webRoutePrefix))

	if a.Spec.LogLevel != "" && a.Spec.LogLevel != "info" {
		amArgs = append(amArgs, fmt.Sprintf("--log.level=%s", a.Spec.LogLevel))
	}

	if version.GTE(semver.MustParse("0.16.0")) {
		if a.Spec.LogFormat != "" && a.Spec.LogFormat != "logfmt" {
			amArgs = append(amArgs, fmt.Sprintf("--log.format=%s", a.Spec.LogFormat))
		}
	}

	if a.Spec.ClusterAdvertiseAddress != "" {
		amArgs = append(amArgs, fmt.Sprintf("--cluster.advertise-address=%s", a.Spec.ClusterAdvertiseAddress))
	}

	localReloadURL := &url.URL{
		Scheme: "http",
		Host:   config.LocalHost + ":9093",
		Path:   path.Clean(webRoutePrefix + "/-/reload"),
	}

	livenessProbeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/-/healthy"),
			Port: intstr.FromString(a.Spec.PortName),
		},
	}

	readinessProbeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean(webRoutePrefix + "/-/ready"),
			Port: intstr.FromString(a.Spec.PortName),
		},
	}

	var livenessProbe *v1.Probe
	var readinessProbe *v1.Probe
	if !a.Spec.ListenLocal {
		livenessProbe = &v1.Probe{
			Handler:          livenessProbeHandler,
			TimeoutSeconds:   probeTimeoutSeconds,
			FailureThreshold: 10,
		}

		readinessProbe = &v1.Probe{
			Handler:             readinessProbeHandler,
			InitialDelaySeconds: 3,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
			FailureThreshold:    10,
		}
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	if a.Spec.PodMetadata != nil {
		if a.Spec.PodMetadata.Labels != nil {
			for k, v := range a.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if a.Spec.PodMetadata.Annotations != nil {
			for k, v := range a.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}
	podLabels["app"] = "alertmanager"
	podLabels["alertmanager"] = a.Name

	var clusterPeerDomain string
	if config.ClusterDomain != "" {
		clusterPeerDomain = fmt.Sprintf("%s.%s.svc.%s.", governingServiceName, a.Namespace, config.ClusterDomain)
	} else {
		// The default DNS search path is .svc.<cluster domain>
		clusterPeerDomain = governingServiceName
	}
	for i := int32(0); i < *a.Spec.Replicas; i++ {
		amArgs = append(amArgs, fmt.Sprintf("--cluster.peer=%s-%d.%s:9094", prefixedName(a.Name), i, clusterPeerDomain))
	}

	for _, peer := range a.Spec.AdditionalPeers {
		amArgs = append(amArgs, fmt.Sprintf("--cluster.peer=%s", peer))
	}

	ports := []v1.ContainerPort{
		{
			Name:          "mesh-tcp",
			ContainerPort: 9094,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "mesh-udp",
			ContainerPort: 9094,
			Protocol:      v1.ProtocolUDP,
		},
	}
	if !a.Spec.ListenLocal {
		ports = append([]v1.ContainerPort{
			{
				Name:          a.Spec.PortName,
				ContainerPort: 9093,
				Protocol:      v1.ProtocolTCP,
			},
		}, ports...)
	}

	// Adjust Alertmanager command line args to specified AM version
	//
	// Alertmanager versions < v0.15.0 are only supported on a best effort basis
	// starting with Prometheus Operator v0.30.0.
	switch version.Major {
	case 0:
		if version.Minor < 15 {
			for i := range amArgs {
				// below Alertmanager v0.15.0 peer address port specification is not necessary
				if strings.Contains(amArgs[i], "--cluster.peer") {
					amArgs[i] = strings.TrimSuffix(amArgs[i], ":9094")
				}

				// below Alertmanager v0.15.0 high availability flags are prefixed with 'mesh' instead of 'cluster'
				amArgs[i] = strings.Replace(amArgs[i], "--cluster.", "--mesh.", 1)
			}
		}
		if version.Minor < 13 {
			for i := range amArgs {
				// below Alertmanager v0.13.0 all flags are with single dash.
				amArgs[i] = strings.Replace(amArgs[i], "--", "-", 1)
			}
		}
		if version.Minor < 7 {
			// below Alertmanager v0.7.0 the flag 'web.route-prefix' does not exist
			amArgs = filter(amArgs, func(s string) bool {
				return !strings.Contains(s, "web.route-prefix")
			})
		}
	default:
		return nil, errors.Errorf("unsupported Alertmanager major version %s", version)
	}

	volumes := []v1.Volume{
		{
			Name: "config-volume",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.Spec.ConfigSecret,
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
			Name:      "config-volume",
			MountPath: alertmanagerConfDir,
		},
		{
			Name:      volName,
			MountPath: alertmanagerStorageDir,
			SubPath:   subPathForStorage(a.Spec.Storage),
		},
	}

	for _, s := range a.Spec.Secrets {
		volumes = append(volumes, v1.Volume{
			Name: k8sutil.SanitizeVolumeName("secret-" + s),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		amVolumeMounts = append(amVolumeMounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("secret-" + s),
			ReadOnly:  true,
			MountPath: secretsDir + s,
		})
	}

	for _, c := range a.Spec.ConfigMaps {
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
		amVolumeMounts = append(amVolumeMounts, v1.VolumeMount{
			Name:      k8sutil.SanitizeVolumeName("configmap-" + c),
			ReadOnly:  true,
			MountPath: configmapsDir + c,
		})
	}

	amVolumeMounts = append(amVolumeMounts, a.Spec.VolumeMounts...)

	resources := v1.ResourceRequirements{Limits: v1.ResourceList{}}
	if config.ConfigReloaderCPU != "0" {
		resources.Limits[v1.ResourceCPU] = resource.MustParse(config.ConfigReloaderCPU)
	}
	if config.ConfigReloaderMemory != "0" {
		resources.Limits[v1.ResourceMemory] = resource.MustParse(config.ConfigReloaderMemory)
	}

	terminationGracePeriod := int64(120)
	finalLabels := config.Labels.Merge(podLabels)

	defaultContainers := []v1.Container{
		{
			Args:           amArgs,
			Name:           "alertmanager",
			Image:          image,
			Ports:          ports,
			VolumeMounts:   amVolumeMounts,
			LivenessProbe:  livenessProbe,
			ReadinessProbe: readinessProbe,
			Resources:      a.Spec.Resources,
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
		}, {
			Name:  "config-reloader",
			Image: config.ConfigReloaderImage,
			Args: []string{
				fmt.Sprintf("-webhook-url=%s", localReloadURL),
				fmt.Sprintf("-volume-dir=%s", alertmanagerConfDir),
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "config-volume",
					ReadOnly:  true,
					MountPath: alertmanagerConfDir,
				},
			},
			Resources:                resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}

	containers, err := k8sutil.MergePatchContainers(defaultContainers, a.Spec.Containers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge containers spec")
	}

	// PodManagementPolicy is set to Parallel to mitigate issues in kubernetes: https://github.com/kubernetes/kubernetes/issues/60164
	// This is also mentioned as one of limitations of StatefulSets: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations
	return &appsv1.StatefulSetSpec{
		ServiceName:         governingServiceName,
		Replicas:            a.Spec.Replicas,
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
				NodeSelector:                  a.Spec.NodeSelector,
				PriorityClassName:             a.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				InitContainers:                a.Spec.InitContainers,
				Containers:                    containers,
				Volumes:                       volumes,
				ServiceAccountName:            a.Spec.ServiceAccountName,
				SecurityContext:               a.Spec.SecurityContext,
				Tolerations:                   a.Spec.Tolerations,
				Affinity:                      a.Spec.Affinity,
			},
		},
	}, nil
}

func configSecretName(name string) string {
	return prefixedName(name)
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

	if s.DisableMountSubPath {
		return ""
	}

	return "alertmanager-db"
}

func filter(strings []string, f func(string) bool) []string {
	filteredStrings := make([]string, 0)
	for _, s := range strings {
		if f(s) {
			filteredStrings = append(filteredStrings, s)
		}
	}
	return filteredStrings
}
