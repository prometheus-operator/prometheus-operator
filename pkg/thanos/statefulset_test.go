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
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	containerName = "thanos-ruler"
)

var (
	defaultTestConfig = Config{
		ReloaderConfig:         operator.DefaultReloaderTestConfig.ReloaderConfig,
		ThanosDefaultBaseImage: operator.DefaultThanosBaseImage,
	}
	emptyQueryEndpoints = []string{""}
)

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel":  "testlabelvalue",
		"managed-by": "prometheus-operator",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
		"kubectl.kubernetes.io/last-applied-configuration": "something",
		"kubectl.kubernetes.io/something":                  "something",
	}
	// kubectl annotations must not be on the statefulset so kubectl does
	// not manage the generated object
	expectedAnnotations := map[string]string{
		"prometheus-operator-input-hash": "",
		"testannotation":                 "testannotationvalue",
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

	require.NoError(t, err)

	require.Equal(t, labels, sset.Labels, pretty.Compare(labels, sset.Labels))

	require.Equal(t, expectedAnnotations, sset.Annotations, pretty.Compare(expectedAnnotations, sset.Annotations))
}

func TestPodLabelsAnnotations(t *testing.T) {
	annotations := map[string]string{
		"testannotation": "testvalue",
	}
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	valLabel := sset.Spec.Template.ObjectMeta.Labels["testlabel"]
	require.Equal(t, "testvalue", valLabel)

	valAnnotations := sset.Spec.Template.ObjectMeta.Annotations["testannotation"]
	require.Equal(t, "testvalue", valAnnotations)
}

func TestThanosDefaultBaseImageFlag(t *testing.T) {
	thanosBaseImageConfig := Config{
		ReloaderConfig:         defaultTestConfig.ReloaderConfig,
		ThanosDefaultBaseImage: "nondefaultuseflag/quay.io/thanos/thanos",
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, thanosBaseImageConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/thanos/thanos" + ":" + operator.DefaultThanosVersion
	require.Equal(t, expected, image)
}

func TestStatefulSetVolumes(t *testing.T) {
	expected := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "tls-assets",
									ReadOnly:  true,
									MountPath: "/etc/thanos/certs",
								},
								{
									Name:      "web-config",
									ReadOnly:  true,
									MountPath: "/etc/thanos/web_config/web-config.yaml",
									SubPath:   "web-config.yaml",
								},
								{
									Name:      "thanos-ruler-foo-data",
									ReadOnly:  false,
									MountPath: "/thanos/data",
									SubPath:   "",
								},
								{
									Name:      "rules-configmap-one",
									ReadOnly:  false,
									MountPath: "/etc/thanos/rules/rules-configmap-one",
									SubPath:   "",
								},
								{
									Name:      "additional-volume",
									ReadOnly:  false,
									MountPath: "/thanos/additional-volume",
									SubPath:   "",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "tls-assets",
							VolumeSource: v1.VolumeSource{
								Projected: &v1.ProjectedVolumeSource{
									Sources: []v1.VolumeProjection{},
								},
							},
						},
						{
							Name: "web-config",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "thanos-ruler-foo-web-config",
								},
							},
						},
						{
							Name: "rules-configmap-one",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "rules-configmap-one",
									},
								},
							},
						},
						{
							Name: "thanos-ruler-foo-data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: "",
								},
							},
						},
						{
							Name: "additional-volume",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: "",
								},
							},
						},
					},
				},
			},
		},
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Volumes: []v1.Volume{
				{
					Name: "additional-volume",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium: "",
						},
					},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "additional-volume",
					ReadOnly:  false,
					MountPath: "/thanos/additional-volume",
					SubPath:   "",
				},
			},
		},
	}, defaultTestConfig, []string{"rules-configmap-one"}, "", &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes)
	require.Equal(t, expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts)
}

func TestTracing(t *testing.T) {
	const (
		secretName = "thanos-tracing-config-secret"
		secretKey  = "config.yaml"
		volumeName = "tracing-config"
		mountPath  = "/etc/thanos/config/tracing-config"
		fullPath   = "/etc/thanos/config/tracing-config/config.yaml"
	)

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			TracingConfig: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, containerName, sset.Spec.Template.Spec.Containers[0].Name)
	{
		var containsVolume bool
		for _, volume := range sset.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				if volume.Secret.SecretName == secretName && volume.Secret.Items[0].Key == secretKey && volume.Secret.Items[0].Path == secretKey {
					containsVolume = true
					break
				}
			}
		}
		require.True(t, containsVolume)
	}
	{
		var containsVolumeMount bool
		for _, volumeMount := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
			if volumeMount.Name == volumeName && volumeMount.MountPath == mountPath {
				containsVolumeMount = true
			}
		}
		require.True(t, containsVolumeMount)
	}
	{
		const expectedArg = "--tracing.config-file=" + fullPath
		var containsArg bool
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		require.True(t, containsArg)
	}
}

func TestTracingFile(t *testing.T) {
	testPath := "/vault/secret/config.yaml"
	testKey := "thanos-tracing-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints:    emptyQueryEndpoints,
			TracingConfigFile: testPath,
			TracingConfig: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	{
		var containsArgConfigFile, containsArgConfig bool
		expectedArgConfigFile := "--tracing.config-file=" + testPath
		expectedArgConfig := "--tracing.config=$(TRACING_CONFIG)"
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-ruler" {
				for _, arg := range container.Args {
					if arg == expectedArgConfigFile {
						containsArgConfigFile = true
					}
					if arg == expectedArgConfig {
						containsArgConfig = true
					}
				}
			}
		}
		require.True(t, containsArgConfigFile)
		require.False(t, containsArgConfig)
	}
}

func TestObjectStorage(t *testing.T) {
	const (
		secretName = "thanos-objstore-config-secret"
		secretKey  = "config.yaml"
		volumeName = "objstorage-config"
		mountPath  = "/etc/thanos/config/objstorage-config"
		fullPath   = "/etc/thanos/config/objstorage-config/config.yaml"
	)

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			ObjectStorageConfig: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, containerName, sset.Spec.Template.Spec.Containers[0].Name)
	{
		var containsVolume bool
		for _, volume := range sset.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				if volume.Secret.SecretName == secretName && volume.Secret.Items[0].Key == secretKey && volume.Secret.Items[0].Path == secretKey {
					containsVolume = true
					break
				}
			}
		}
		require.True(t, containsVolume)
	}
	{
		var containsVolumeMount bool
		for _, volumeMount := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
			if volumeMount.Name == volumeName && volumeMount.MountPath == mountPath {
				containsVolumeMount = true
			}
		}
		require.True(t, containsVolumeMount)
	}
	{
		const expectedArg = "--objstore.config-file=" + fullPath
		var containsArg bool
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		require.True(t, containsArg)
	}
}

func TestObjectStorageFile(t *testing.T) {
	testPath := "/vault/secret/config.yaml"
	testKey := "thanos-objstore-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints:          emptyQueryEndpoints,
			ObjectStorageConfigFile: &testPath,
			ObjectStorageConfig: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	{
		var containsArgConfigFile, containsArgConfig bool
		expectedArgConfigFile := "--objstore.config-file=" + testPath
		expectedArgConfig := "--objstore.config=$(OBJSTORE_CONFIG)"
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-ruler" {
				for _, arg := range container.Args {
					if arg == expectedArgConfigFile {
						containsArgConfigFile = true
					}
					if arg == expectedArgConfig {
						containsArgConfig = true
					}
				}
			}
		}
		require.True(t, containsArgConfigFile)
		require.False(t, containsArgConfig)
	}
}

func TestAlertRelabel(t *testing.T) {
	const (
		secretName = "thanos-alertrelabel-config-secret"
		secretKey  = "config.yaml"
		volumeName = "alertrelabel-config"
		mountPath  = "/etc/thanos/config/alertrelabel-config"
		fullPath   = "/etc/thanos/config/alertrelabel-config/config.yaml"
	)

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			AlertRelabelConfigs: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, containerName, sset.Spec.Template.Spec.Containers[0].Name)
	{
		var containsVolume bool
		for _, volume := range sset.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				if volume.Secret.SecretName == secretName && volume.Secret.Items[0].Key == secretKey && volume.Secret.Items[0].Path == secretKey {
					containsVolume = true
					break
				}
			}
		}
		require.True(t, containsVolume)
	}
	{
		var containsVolumeMount bool
		for _, volumeMount := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
			if volumeMount.Name == volumeName && volumeMount.MountPath == mountPath {
				containsVolumeMount = true
			}
		}
		require.True(t, containsVolumeMount)
	}
	{
		const expectedArg = "--alert.relabel-config-file=" + fullPath
		var containsArg bool
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		require.True(t, containsArg)
	}
}

func TestAlertRelabelFile(t *testing.T) {
	testPath := "/vault/secret/config.yaml"
	testKey := "thanos-alertrelabel-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints:         emptyQueryEndpoints,
			AlertRelabelConfigFile: &testPath,
			AlertRelabelConfigs: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	{
		var containsArgConfigFile, containsArgConfigs bool
		expectedArgConfigFile := "--alert.relabel-config-file=" + testPath
		expectedArgConfigs := "--alert.relabel-config=$(ALERT_RELABEL_CONFIG)"
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-ruler" {
				for _, arg := range container.Args {
					if arg == expectedArgConfigFile {
						containsArgConfigFile = true
					}
					if arg == expectedArgConfigs {
						containsArgConfigs = true
					}
				}
			}
		}
		require.True(t, containsArgConfigFile)
		require.False(t, containsArgConfigs)
	}
}

func TestLabelsAndAlertDropLabels(t *testing.T) {
	labelPrefix := "--label="
	alertDropLabelPrefix := "--alert.label-drop="

	tests := []struct {
		Name                    string
		Labels                  map[string]string
		AlertDropLabels         []string
		ExpectedLabels          []string
		ExpectedAlertDropLabels []string
	}{
		{
			Name:                    "thanos_ruler_replica-is-set",
			Labels:                  nil,
			AlertDropLabels:         nil,
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica"},
		},
		{
			Name:                    "alert-drop-labels-are-set",
			Labels:                  nil,
			AlertDropLabels:         []string{"test"},
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test"},
		},
		{
			Name: "labels-are-set",
			Labels: map[string]string{
				"test": "test",
			},
			AlertDropLabels:         nil,
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`, `test="test"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica"},
		},
		{
			Name: "both-alert-drop-labels-and-labels-are-set",
			Labels: map[string]string{
				"test": "test",
			},
			AlertDropLabels:         []string{"test"},
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`, `test="test"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test"},
		},
		{
			Name: "assert-labels-order-is-constant",
			Labels: map[string]string{
				"thanos_ruler_replica": "$(POD_NAME)",
				"test":                 "test",
				"test4":                "test4",
				"test1":                "test1",
				"test2":                "test2",
				"test3":                "test3",
				"foo":                  "bar",
				"bob":                  "alice",
			},
			AlertDropLabels:         []string{"test", "aaa", "foo", "bar", "foo1", "foo2", "foo3"},
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`, `bob="alice"`, `foo="bar"`, `test="test"`, `test1="test1"`, `test2="test2"`, `test3="test3"`, `test4="test4"`, `thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test", "aaa", "foo", "bar", "foo1", "foo2", "foo3"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			actualLabels := []string{}
			actualDropLabels := []string{}
			sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: monitoringv1.ThanosRulerSpec{
					QueryEndpoints:  emptyQueryEndpoints,
					Labels:          tc.Labels,
					AlertDropLabels: tc.AlertDropLabels,
				},
			}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
			require.NoError(t, err)

			ruler := sset.Spec.Template.Spec.Containers[0]
			require.Equal(t, "thanos-ruler", ruler.Name)

			for _, arg := range ruler.Args {
				if strings.HasPrefix(arg, labelPrefix) {
					actualLabels = append(actualLabels, strings.TrimPrefix(arg, labelPrefix))
				} else if strings.HasPrefix(arg, alertDropLabelPrefix) {
					actualDropLabels = append(actualDropLabels, strings.TrimPrefix(arg, alertDropLabelPrefix))
				}
			}
			require.Equal(t, tc.ExpectedLabels, actualLabels)
			require.Equal(t, tc.ExpectedAlertDropLabels, actualDropLabels)
		})
	}
}

func TestAdditionalContainers(t *testing.T) {
	// The base to compare everything against
	baseSet, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	// Add an extra container
	addSset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Containers: []v1.Container{
				{
					Name: "extra-container",
				},
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Len(t, addSset.Spec.Template.Spec.Containers, len(baseSet.Spec.Template.Spec.Containers)+1)

	// Adding a new container with the same name results in a merge and just one container
	const existingContainerName = "thanos-ruler"
	const containerImage = "madeUpContainerImage"
	modSset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Containers: []v1.Container{
				{
					Name:  existingContainerName,
					Image: containerImage,
				},
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, len(baseSet.Spec.Template.Spec.Containers), len(modSset.Spec.Template.Spec.Containers))

	// Check that adding a container with an existing name results in a single patched container.
	for _, c := range modSset.Spec.Template.Spec.Containers {
		require.False(t, c.Name == existingContainerName && c.Image != containerImage)
	}
}

func TestRetention(t *testing.T) {
	for _, tc := range []struct {
		specRetention     monitoringv1.Duration
		expectedRetention monitoringv1.Duration
	}{
		{"1d", "1d"},
	} {
		t.Run(string(tc.specRetention), func(t *testing.T) {
			sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
				Spec: monitoringv1.ThanosRulerSpec{
					Retention:      tc.specRetention,
					QueryEndpoints: emptyQueryEndpoints,
				},
			}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

			require.NoError(t, err)

			trArgs := sset.Spec.Template.Spec.Containers[0].Args
			expectedRetentionArg := fmt.Sprintf("--tsdb.retention=%s", tc.expectedRetention)
			found := false
			for _, flag := range trArgs {
				if flag == expectedRetentionArg {
					found = true
					break
				}
			}

			require.True(t, found)
		})
	}
}

func TestPodTemplateConfig(t *testing.T) {
	nodeSelector := map[string]string{
		"foo": "bar",
	}
	affinity := v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{},
		PodAffinity: &v1.PodAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					PodAffinityTerm: v1.PodAffinityTerm{
						Namespaces: []string{"foo"},
					},
					Weight: 100,
				},
			},
		},
		PodAntiAffinity: &v1.PodAntiAffinity{},
	}

	tolerations := []v1.Toleration{
		{
			Key: "key",
		},
	}
	userid := int64(1234)
	securityContext := v1.PodSecurityContext{
		RunAsUser: &userid,
	}
	priorityClassName := "foo"
	serviceAccountName := "thanos-ruler-sa"
	hostAliases := []monitoringv1.HostAlias{
		{
			Hostnames: []string{"foo.com"},
			IP:        "1.1.1.1",
		},
	}
	imagePullSecrets := []v1.LocalObjectReference{
		{
			Name: "registry-secret",
		},
	}
	imagePullPolicy := v1.PullAlways

	additionalArgs := []monitoringv1.Argument{
		{Name: "additional.arg", Value: "additional-arg-value"},
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints:     emptyQueryEndpoints,
			NodeSelector:       nodeSelector,
			Affinity:           &affinity,
			Tolerations:        tolerations,
			SecurityContext:    &securityContext,
			PriorityClassName:  priorityClassName,
			ServiceAccountName: serviceAccountName,
			HostAliases:        hostAliases,
			ImagePullSecrets:   imagePullSecrets,
			ImagePullPolicy:    imagePullPolicy,
			AdditionalArgs:     additionalArgs,
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, nodeSelector, sset.Spec.Template.Spec.NodeSelector)
	require.Equal(t, affinity, *sset.Spec.Template.Spec.Affinity)
	require.Equal(t, tolerations, sset.Spec.Template.Spec.Tolerations)
	require.Equal(t, securityContext, *sset.Spec.Template.Spec.SecurityContext)
	require.Equal(t, priorityClassName, sset.Spec.Template.Spec.PriorityClassName)
	require.Equal(t, serviceAccountName, sset.Spec.Template.Spec.ServiceAccountName)
	require.Equal(t, len(hostAliases), len(sset.Spec.Template.Spec.HostAliases))
	require.Equal(t, imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets)
	for _, initContainer := range sset.Spec.Template.Spec.InitContainers {
		require.Equal(t, imagePullPolicy, initContainer.ImagePullPolicy)
	}
	for _, container := range sset.Spec.Template.Spec.Containers {
		require.Equal(t, imagePullPolicy, container.ImagePullPolicy)
	}
	require.Contains(t, sset.Spec.Template.Spec.Containers[0].Args[len(sset.Spec.Template.Spec.Containers[0].Args)-1], "--additional.arg=additional-arg-value")
	require.Equal(t, "rule", sset.Spec.Template.Spec.Containers[0].Args[0])
}

func TestExternalQueryURL(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			AlertQueryURL:  "https://example.com/",
			QueryEndpoints: emptyQueryEndpoints,
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, containerName, sset.Spec.Template.Spec.Containers[0].Name)

	const expectedArg = "--alert.query-url=https://example.com/"
	for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
		if arg == expectedArg {
			return
		}
	}
	require.FailNow(t, "Thanos ruler is missing expected argument: %s", expectedArg)
}

func TestSidecarResources(t *testing.T) {
	operator.TestSidecarsResources(t, func(reloaderConfig operator.ContainerConfig) *appsv1.StatefulSet {
		testConfig := Config{
			ReloaderConfig:         reloaderConfig,
			ThanosDefaultBaseImage: operator.DefaultThanosBaseImage,
		}
		tr := &monitoringv1.ThanosRuler{
			Spec: monitoringv1.ThanosRulerSpec{
				QueryEndpoints: emptyQueryEndpoints,
			},
		}
		// thanos-ruler sset will only have a configReloader side car
		// if it has to mount a ConfigMap
		sset, err := makeStatefulSet(tr, testConfig, []string{"my-configmap"}, "", &operator.ShardedSecret{})
		require.NoError(t, err)
		return sset
	})
}

func TestStatefulSetMinReadySeconds(t *testing.T) {
	tr := monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			MinReadySeconds: nil,
			QueryEndpoints:  emptyQueryEndpoints,
		},
	}

	statefulSet, err := makeStatefulSetSpec(&tr, defaultTestConfig, nil, &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, int32(0), statefulSet.MinReadySeconds)

	// assert set correctly if not nil
	var expect uint32 = 5
	tr.Spec.MinReadySeconds = &expect
	statefulSet, err = makeStatefulSetSpec(&tr, defaultTestConfig, nil, &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, int32(expect), statefulSet.MinReadySeconds)
}

func TestStatefulSetServiceName(t *testing.T) {
	tr := monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
		},
	}

	// assert set correctly
	expect := governingServiceName
	spec, err := makeStatefulSetSpec(&tr, defaultTestConfig, nil, &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, expect, spec.ServiceName)
}

func TestStatefulSetPVC(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	storageClass := "storageclass"

	pvc := monitoringv1.EmbeddedPersistentVolumeClaim{
		EmbeddedObjectMetadata: monitoringv1.EmbeddedObjectMetadata{
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			StorageClassName: &storageClass,
		},
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: pvc,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetPvc := sset.Spec.VolumeClaimTemplates[0]
	require.Equal(t, *pvc.Spec.StorageClassName, *ssetPvc.Spec.StorageClassName)
}

func TestStatefulEmptyDir(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	emptyDir := v1.EmptyDirVolumeSource{
		Medium: v1.StorageMediumMemory,
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Storage: &monitoringv1.StorageSpec{
				EmptyDir: &emptyDir,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir)
	require.Equal(t, emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium)
}

func TestStatefulSetEphemeral(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	storageClass := "storageclass"

	ephemeral := v1.EphemeralVolumeSource{
		VolumeClaimTemplate: &v1.PersistentVolumeClaimTemplate{
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				StorageClassName: &storageClass,
			},
		},
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			Storage: &monitoringv1.StorageSpec{
				Ephemeral: &ephemeral,
			},
		},
	}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral)
	require.Equal(t, ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName)
}

func TestThanosVersion(t *testing.T) {
	thanosBaseImage := defaultTestConfig.ThanosDefaultBaseImage
	for _, tc := range []struct {
		version       string
		expectedImage string
		expectedError bool
	}{
		{"v0.29.0", thanosBaseImage + ":" + "v0.29.0", false},
		{"0.29.0", thanosBaseImage + ":" + "0.29.0", false},
		{"", thanosBaseImage + ":" + operator.DefaultThanosVersion, false},
		{"0.29.0-0123", "", true},
		{"0.29.0.DEV", "", true},
	} {
		t.Run(tc.version, func(t *testing.T) {
			sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
				Spec: monitoringv1.ThanosRulerSpec{
					QueryEndpoints: emptyQueryEndpoints,
					Version:        tc.version,
				},
			}, defaultTestConfig, nil, "", &operator.ShardedSecret{})

			if tc.expectedError {
				require.Error(t, err)
			}

			if !tc.expectedError {
				image := sset.Spec.Template.Spec.Containers[0].Image
				require.Equal(t, tc.expectedImage, image)
			}
		})
	}
}
