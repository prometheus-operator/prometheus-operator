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

package prometheus

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

var defaultTestConfig = prompkg.Config{
	LocalHost:                  "localhost",
	ReloaderConfig:             operator.DefaultReloaderTestConfig.ReloaderConfig,
	PrometheusDefaultBaseImage: operator.DefaultPrometheusBaseImage,
	ThanosDefaultBaseImage:     operator.DefaultThanosBaseImage,
}

func makeStatefulSetFromPrometheus(p monitoringv1.Prometheus) (*appsv1.StatefulSet, error) {
	logger := prompkg.NewLogger()

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	if err != nil {
		return nil, err
	}

	return makeStatefulSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		nil,
		"abc",
		0,
		&operator.ShardedSecret{})
}

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
		"kubectl.kubernetes.io/last-applied-configuration": "something",
		"kubectl.kubernetes.io/something":                  "something",
	}

	// kubectl annotations must not be on the statefulset so kubectl does
	// not manage the generated object
	expectedStatefulSetAnnotations := map[string]string{
		"prometheus-operator-input-hash": "abc",
		"testannotation":                 "testannotationvalue",
	}

	expectedStatefulSetLabels := map[string]string{
		"testlabel":                    "testlabelvalue",
		"operator.prometheus.io/name":  "test",
		"operator.prometheus.io/shard": "0",
		"operator.prometheus.io/mode":  "server",
		"managed-by":                   "prometheus-operator",
		"prometheus":                   "test",
		"app.kubernetes.io/instance":   "test",
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/name":       "prometheus",
	}

	expectedPodLabels := map[string]string{
		"prometheus":                   "test",
		"app.kubernetes.io/name":       "prometheus",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultPrometheusVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "test",
		"operator.prometheus.io/name":  "test",
		"operator.prometheus.io/shard": "0",
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "ns",
			Labels:      labels,
			Annotations: annotations,
		},
	})
	require.NoError(t, err)

	require.Equalf(t, expectedStatefulSetLabels, sset.Labels, "Labels are not properly being propagated to the StatefulSet\n%s", pretty.Compare(expectedStatefulSetLabels, sset.Labels))
	require.Equalf(t, expectedStatefulSetAnnotations, sset.Annotations, "Annotations are not properly being propagated to the StatefulSet\n%s", pretty.Compare(expectedStatefulSetAnnotations, sset.Annotations))
	require.Equalf(t, expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels, "Labels are not properly being propagated to the Pod\n%s", pretty.Compare(expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels))
}

func TestPodLabelsAnnotations(t *testing.T) {
	annotations := map[string]string{
		"testannotation": "testvalue",
	}
	labels := map[string]string{
		"testlabel": "testvalue",
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
					Annotations: annotations,
					Labels:      labels,
				},
			},
		},
	})
	require.NoError(t, err)

	valLabel := sset.Spec.Template.ObjectMeta.Labels["testlabel"]
	require.Equal(t, "testvalue", valLabel, "Pod labels are not properly propagated")

	valAnnotation := sset.Spec.Template.ObjectMeta.Annotations["testannotation"]
	require.Equal(t, "testvalue", valAnnotation, "Pod annotations are not properly propagated")
}

func TestPodLabelsShouldNotBeSelectorLabels(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
					Labels: labels,
				},
			},
		},
	})
	require.NoError(t, err)

	require.NotEqual(t, "testvalue", sset.Spec.Selector.MatchLabels["testlabel"], "Pod Selector are not properly propagated")
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

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: pvc,
				},
			},
		},
	})
	require.NoError(t, err)

	ssetPvc := sset.Spec.VolumeClaimTemplates[0]
	require.Equal(t, *pvc.Spec.StorageClassName, *ssetPvc.Spec.StorageClassName, "Error adding PVC Spec to StatefulSetSpec")
}

func TestStatefulSetEmptyDir(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	emptyDir := v1.EmptyDirVolumeSource{
		Medium: v1.StorageMediumMemory,
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					EmptyDir: &emptyDir,
				},
			},
		},
	})
	require.NoError(t, err)

	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir, "Error adding EmptyDir Spec to StatefulSetSpec")
	require.Equal(t, emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium, "Error adding EmptyDir Spec to StatefulSetSpec")
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

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					Ephemeral: &ephemeral,
				},
			},
		},
	})
	require.NoError(t, err)

	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral, "Error adding Ephemeral Spec to StatefulSetSpec")
	require.Equal(t, ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName, "Error adding Ephemeral Spec to StatefulSetSpec")
}

func TestStatefulSetVolumeInitial(t *testing.T) {
	p := monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "volume-init-test",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Secrets: []string{
					"test-secret1",
				},
			},
		},
	}

	expected := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config-out",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/config_out",
								},
								{
									Name:      "tls-assets",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/certs",
								},
								{
									Name:      "prometheus-volume-init-test-db",
									ReadOnly:  false,
									MountPath: "/prometheus",
								},
								{
									Name:      "secret-test-secret1",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/secrets/test-secret1",
								},
								{
									Name:      "rules-configmap-one",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/rules/rules-configmap-one",
								},
								{
									Name:      "web-config",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/web_config/web-config.yaml",
									SubPath:   "web-config.yaml",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: prompkg.ConfigSecretName(&p),
								},
							},
						},
						{
							Name: "tls-assets",
							VolumeSource: v1.VolumeSource{
								Projected: &v1.ProjectedVolumeSource{
									Sources: []v1.VolumeProjection{
										{
											Secret: &v1.SecretProjection{
												LocalObjectReference: v1.LocalObjectReference{
													Name: prompkg.TLSAssetsSecretName(&p) + "-0",
												},
											},
										},
									},
								},
							},
						},
						{
							Name: "config-out",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: v1.StorageMediumMemory,
								},
							},
						},
						{
							Name: "secret-test-secret1",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "test-secret1",
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
									Optional: ptr.To(true),
								},
							},
						},
						{
							Name: "web-config",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "prometheus-volume-init-test-web-config",
								},
							},
						},
						{
							Name: "prometheus-volume-init-test-db",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	logger := prompkg.NewLogger()

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	shardedSecret, err := operator.ReconcileShardedSecret(
		context.Background(),
		map[string][]byte{},
		fake.NewSimpleClientset(),
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      prompkg.TLSAssetsSecretName(&p),
				Namespace: "test",
			},
		},
	)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"volume-init-test",
		&p,
		defaultTestConfig,
		cg,
		[]string{"rules-configmap-one"},
		"",
		0,
		shardedSecret)
	require.NoError(t, err)

	require.Equalf(t, expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes, "expected volumes to match \n%s", pretty.Compare(expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes))
	require.Equalf(t, expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts, "expected volume mounts to match \n%s", pretty.Compare(expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts))
}

func TestAdditionalConfigMap(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ConfigMaps: []string{"test-cm1"},
			},
		},
	})
	require.NoError(t, err)

	cmVolumeFound := false
	for _, v := range sset.Spec.Template.Spec.Volumes {
		if strings.HasPrefix(v.Name, "configmap-test-cm1") {
			cmVolumeFound = true
		}
	}
	require.True(t, cmVolumeFound, "ConfigMap volume not found")

	cmMounted := false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if strings.HasPrefix(v.Name, "configmap-test-cm1") && v.MountPath == "/etc/prometheus/configmaps/test-cm1" {
			cmMounted = true
		}
	}
	require.True(t, cmMounted, "ConfigMap volume not mounted")
}

func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ListenLocal: true,
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.listen-address=127.0.0.1:9090" {
			found = true
		}
	}

	require.True(t, found, "Prometheus not listening on loopback when it should.")

	expectedProbeHandler := func(probePath string) v1.ProbeHandler {
		return v1.ProbeHandler{
			Exec: &v1.ExecAction{
				Command: []string{
					`sh`,
					`-c`,
					fmt.Sprintf(`if [ -x "$(command -v curl)" ]; then exec curl --fail %[1]s; elif [ -x "$(command -v wget)" ]; then exec wget -q -O /dev/null %[1]s; else exit 1; fi`, fmt.Sprintf("http://localhost:9090%s", probePath)),
				},
			},
		}
	}

	actualStartupProbe := sset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
	require.Equal(t, expectedStartupProbe, actualStartupProbe, "Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe, "Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	require.Equal(t, expectedReadinessProbe, actualReadinessProbe, "Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)

	require.Empty(t, sset.Spec.Template.Spec.Containers[0].Ports, "Prometheus container should have 0 ports defined")
}

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Web: &monitoringv1.PrometheusWebSpec{
					WebConfigFileFields: monitoringv1.WebConfigFileFields{
						TLSConfig: &monitoringv1.WebTLSConfig{
							KeySecret: v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "some-secret",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								ConfigMap: &v1.ConfigMapKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "some-configmap",
									},
								},
							},
						},
					},
				},
			},
			Thanos: &monitoringv1.ThanosSpec{},
		},
	})
	require.NoError(t, err)

	expectedProbeHandler := func(probePath string) v1.ProbeHandler {
		return v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   probePath,
				Port:   intstr.FromString("web"),
				Scheme: "HTTPS",
			},
		}
	}

	actualStartupProbe := sset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
	require.Equal(t, expectedStartupProbe, actualStartupProbe, "Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe, "Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	require.Equal(t, expectedReadinessProbe, actualReadinessProbe, "Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9090/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	require.True(t, reloadURLFound, "expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)

	expectedThanosSidecarPrometheusURL := "--prometheus.url=https://localhost:9090/"
	prometheusURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
		if arg == expectedThanosSidecarPrometheusURL {
			prometheusURLFound = true
		}
	}
	require.True(t, prometheusURLFound, "expected to find arg %s in thanos sidecar", expectedThanosSidecarPrometheusURL)

	fmt.Println(sset.Spec.Template.Spec.Containers[2].Args)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/prometheus/web_config/web-config.yaml",
		"--reload-url=https://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
		}
	}
}

func TestTagAndShaAndVersion(t *testing.T) {
	{
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
				},
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus:my-unrelated-tag"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	{
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	// For tests which set monitoringv1.PrometheusSpec.Image, the result will be Image only. SHA, Tag, Version are not considered.
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
					Image:   &image,
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus:latest"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
					Image:   &image,
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Explicit image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
					Image:   &image,
				},
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: "v2.3.2",
					Image:   &image,
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: &image,
				},
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: &image,
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := ""
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: &image,
				},
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus:my-unrelated-tag"
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
	{
		image := "my-reg/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb325"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Image: &image,
				},
				SHA: "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag: "my-unrelated-tag",
			},
		})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "my-reg/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb325"
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
	}
}

func TestPrometheusDefaultBaseImageFlag(t *testing.T) {
	operatorConfig := prompkg.Config{
		ReloaderConfig:             defaultTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: "nondefaultuseflag/quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "nondefaultuseflag/quay.io/thanos/thanos",
	}
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	logger := prompkg.NewLogger()
	p := monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"test",
		&p,
		operatorConfig,
		cg,
		nil,
		"",
		0,
		&operator.ShardedSecret{})
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/prometheus/prometheus" + ":" + operator.DefaultPrometheusVersion
	require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
}

func TestThanosDefaultBaseImageFlag(t *testing.T) {
	thanosBaseImageConfig := prompkg.Config{
		ReloaderConfig:             defaultTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: "nondefaultuseflag/quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "nondefaultuseflag/quay.io/thanos/thanos",
	}
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}
	logger := prompkg.NewLogger()
	p := monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"test",
		&p,
		thanosBaseImageConfig,
		cg,
		nil,
		"",
		0,
		&operator.ShardedSecret{})
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[2].Image
	expected := "nondefaultuseflag/quay.io/thanos/thanos" + ":" + operator.DefaultThanosVersion
	require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
}

func TestThanosTagAndShaAndVersion(t *testing.T) {
	{
		thanosTag := "my-unrelated-tag"
		thanosVersion := "v0.1.0"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[2].Image
		expected := "quay.io/thanos/thanos:my-unrelated-tag"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	{
		thanosSHA := "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		thanosTag := "my-unrelated-tag"
		thanosVersion := "v0.1.0-rc.2"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[2].Image
		expected := "quay.io/thanos/thanos@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	{
		thanosSHA := "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		thanosTag := "my-unrelated-tag"
		thanosVersion := "v0.1.0-rc.2"
		thanosImage := "my-registry/thanos:latest"
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
					Image:   &thanosImage,
				},
			},
		})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[2].Image
		expected := "my-registry/thanos:latest"
		require.Equal(t, expected, image, "Explicit Thanos image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
}

func TestThanosResourcesNotSet(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	})
	require.NoError(t, err)

	res := sset.Spec.Template.Spec.Containers[2].Resources
	require.False(t, (res.Limits != nil || res.Requests != nil), "Unexpected resources defined. \n\nExpected: nil\n\nGot: %v, %v", res.Limits, res.Requests)
}

func TestThanosResourcesSet(t *testing.T) {
	expected := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("125m"),
			v1.ResourceMemory: resource.MustParse("75Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				Resources: expected,
			},
		},
	})
	require.NoError(t, err)

	actual := sset.Spec.Template.Spec.Containers[2].Resources
	require.Equal(t, expected, actual, "Unexpected resources defined. \n\nExpected: %v\n\nGot: %v", expected, actual)
}

func TestThanosNoObjectStorage(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "prometheus", sset.Spec.Template.Spec.Containers[0].Name, "expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	require.Equal(t, "thanos-sidecar", sset.Spec.Template.Spec.Containers[2].Name, "expected 3rd container to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)

	for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
		require.False(t, strings.HasPrefix(arg, "--storage.tsdb.max-block-duration=2h"), "Prometheus compaction should be disabled")
	}

	for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
		require.False(t, strings.HasPrefix(arg, "--tsdb.path="), "--tsdb.path argument should not be given to the Thanos sidecar")
	}
}

func TestThanosObjectStorage(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ObjectStorageConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
				BlockDuration: "2h",
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "prometheus", sset.Spec.Template.Spec.Containers[0].Name, "expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	require.Equal(t, "thanos-sidecar", sset.Spec.Template.Spec.Containers[2].Name, "expected 3rd containers to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[2].Env {
		if env.Name == "OBJSTORE_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	require.True(t, containsEnvVar, "Thanos sidecar is missing expected OBJSTORE_CONFIG env var with correct value")

	{
		const expectedArg = "--objstore.config=$(OBJSTORE_CONFIG)"
		require.True(t, slices.Contains(sset.Spec.Template.Spec.Containers[2].Args, expectedArg), "Thanos sidecar is missing expected argument: %s", expectedArg)
	}
	{
		const expectedArg = "--storage.tsdb.max-block-duration=2h"
		require.True(t, slices.Contains(sset.Spec.Template.Spec.Containers[0].Args, expectedArg), "Prometheus is missing expected argument: %s", expectedArg)
	}

	{
		var found bool
		for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
			if strings.HasPrefix(arg, "--tsdb.path=") {
				found = true
				break
			}
		}
		require.True(t, found, "--tsdb.path argument should be given to the Thanos sidecar, got %q", strings.Join(sset.Spec.Template.Spec.Containers[2].Args, " "))
	}

	{
		var found bool
		for _, vol := range sset.Spec.Template.Spec.Containers[2].VolumeMounts {
			if vol.MountPath == prompkg.StorageDir {
				found = true
				break
			}
		}
		require.True(t, found, "Prometheus data volume should be mounted in the Thanos sidecar")
	}
}

func TestThanosObjectStorageFile(t *testing.T) {
	testPath := "/vault/secret/config.yaml"
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ObjectStorageConfigFile: &testPath,
				BlockDuration:           "2h",
			},
		},
	})
	require.NoError(t, err)

	{
		expectedArg := "--objstore.config-file=" + testPath
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-sidecar" {
				require.True(t, slices.Contains(container.Args, expectedArg),
					"Thanos sidecar is missing expected argument: %s", expectedArg)
				break
			}
		}
	}

	{
		const expectedArg = "--storage.tsdb.max-block-duration=2h"
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "prometheus" {
				require.True(t, slices.Contains(container.Args, expectedArg),
					"Prometheus is missing expected argument: %s", expectedArg)
				break
			}
		}
	}

	{
		var found bool
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-sidecar" {
				for _, arg := range container.Args {
					if strings.HasPrefix(arg, "--tsdb.path=") {
						found = true
						break
					}
				}
			}
		}
		require.True(t, found, "--tsdb.path argument should be given to the Thanos sidecar, got %q", strings.Join(sset.Spec.Template.Spec.Containers[2].Args, " "))
	}

	{
		var found bool
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-sidecar" {
				for _, vol := range container.VolumeMounts {
					if vol.MountPath == prompkg.StorageDir {
						found = true
						break
					}
				}
			}
		}
		require.True(t, found, "Prometheus data volume should be mounted in the Thanos sidecar")
	}
}

func TestThanosBlockDuration(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				BlockDuration: "1h",
				ObjectStorageConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
		if arg == "--storage.tsdb.max-block-duration=1h" {
			found = true
		}
	}
	require.True(t, found, "Thanos BlockDuration arg change not found")
}

func TestThanosWithNamedPVC(t *testing.T) {
	testKey := "named-pvc"
	storageClass := "storageclass"

	pvc := monitoringv1.EmbeddedPersistentVolumeClaim{
		EmbeddedObjectMetadata: monitoringv1.EmbeddedObjectMetadata{
			Name: testKey,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			StorageClassName: &storageClass,
		},
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: pvc,
				},
			},
			Thanos: &monitoringv1.ThanosSpec{
				BlockDuration: "1h",
				ObjectStorageConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "thanos-sidecar" {
			for _, vol := range container.VolumeMounts {
				if vol.Name == testKey {
					found = true
				}
			}
		}
	}
	require.True(t, found, "VolumeClaimTemplate name not found on thanos-sidecar volumeMounts")
}

func TestThanosTracing(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				TracingConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "prometheus", sset.Spec.Template.Spec.Containers[0].Name, "expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	require.Equal(t, "thanos-sidecar", sset.Spec.Template.Spec.Containers[2].Name, "expected 3rd containers to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[2].Env {
		if env.Name == "TRACING_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	require.True(t, containsEnvVar, "Thanos sidecar is missing expected TRACING_CONFIG env var with correct value")

	{
		const expectedArg = "--tracing.config=$(TRACING_CONFIG)"
		require.True(t, slices.Contains(sset.Spec.Template.Spec.Containers[2].Args, expectedArg), "Thanos sidecar is missing expected argument: %s", expectedArg)
	}
}

func TestThanosSideCarVolumes(t *testing.T) {
	testVolume := "test-volume"
	testVolumeMountPath := "/prometheus/thanos-sidecar"
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Volumes: []v1.Volume{
					{
						Name: testVolume,
						VolumeSource: v1.VolumeSource{
							EmptyDir: &v1.EmptyDirVolumeSource{},
						},
					},
				},
			},
			Thanos: &monitoringv1.ThanosSpec{
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      testVolume,
						MountPath: testVolumeMountPath,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	var containsVolume bool
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == testVolume {
			containsVolume = true
			break
		}
	}
	require.True(t, containsVolume, "Thanos sidecar volume is missing expected volume: %s", testVolume)

	var containsVolumeMount bool
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "thanos-sidecar" {
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == testVolume && volumeMount.MountPath == testVolumeMountPath {
					containsVolumeMount = true
					break
				}
			}
		}
	}

	require.True(t, containsVolumeMount, "expected thanos sidecar volume mounts to match")
}

func TestRetentionAndRetentionSize(t *testing.T) {
	tests := []struct {
		version                    string
		specRetention              monitoringv1.Duration
		specRetentionSize          monitoringv1.ByteSize
		expectedRetentionArg       string
		expectedRetentionSizeArg   string
		shouldContainRetention     bool
		shouldContainRetentionSize bool
	}{
		{"v2.5.0", "", "", "--storage.tsdb.retention=24h", "--storage.tsdb.retention.size=", true, false},
		{"v2.5.0", "1d", "", "--storage.tsdb.retention=1d", "--storage.tsdb.retention.size=", true, false},
		{"v2.5.0", "", "512MB", "--storage.tsdb.retention=24h", "--storage.tsdb.retention.size=512MB", true, false},
		{"v2.5.0", "1d", "512MB", "--storage.tsdb.retention=1d", "--storage.tsdb.retention.size=512MB", true, false},
		{"v2.7.0", "", "", "--storage.tsdb.retention.time=24h", "--storage.tsdb.retention.size=", true, false},
		{"v2.7.0", "1d", "", "--storage.tsdb.retention.time=1d", "--storage.tsdb.retention.size=", true, false},
		{"v2.7.0", "", "512MB", "--storage.tsdb.retention.time=24h", "--storage.tsdb.retention.size=512MB", false, true},
		{"v2.7.0", "1d", "512MB", "--storage.tsdb.retention.time=1d", "--storage.tsdb.retention.size=512MB", true, true},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: test.version,
				},
				Retention:     test.specRetention,
				RetentionSize: test.specRetentionSize,
			},
		})
		require.NoError(t, err)

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		retentionFlag := strings.Split(test.expectedRetentionArg, "=")[0]
		foundRetentionFlag := false
		foundRetentionSizeFlag := false
		foundRetention := false
		foundRetentionSize := false
		for _, flag := range promArgs {
			if flag == test.expectedRetentionArg {
				foundRetention = true
			} else if flag == test.expectedRetentionSizeArg {
				foundRetentionSize = true
			}

			if strings.HasPrefix(flag, retentionFlag) {
				foundRetentionFlag = true
			} else if strings.HasPrefix(flag, "--storage.tsdb.retention.size") {
				foundRetentionSizeFlag = true
			}
		}

		if test.shouldContainRetention {
			require.True(t, (foundRetention && foundRetentionFlag))
		}

		if test.shouldContainRetentionSize {
			require.True(t, (foundRetentionSize && foundRetentionSizeFlag))
		}
	}
}

func TestReplicasConfigurationWithSharding(t *testing.T) {
	testConfig := prompkg.Config{
		ReloaderConfig:             defaultTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	replicas := int32(2)
	shards := int32(3)
	logger := prompkg.NewLogger()
	p := monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: &replicas,
				Shards:   &shards,
			},
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"test",
		&p,
		testConfig,
		cg,
		nil,
		"",
		1,
		&operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, int32(2), *sset.Spec.Replicas, "Unexpected replicas configuration.")

	found := false
	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			for _, env := range c.Env {
				if env.Name == "SHARD" && env.Value == "1" {
					found = true
				}
			}
		}
	}
	require.True(t, found, "Shard.")
}

func TestSidecarResources(t *testing.T) {
	operator.TestSidecarsResources(t, func(reloaderConfig operator.ContainerConfig) *appsv1.StatefulSet {
		testConfig := prompkg.Config{
			ReloaderConfig:             reloaderConfig,
			PrometheusDefaultBaseImage: defaultTestConfig.PrometheusDefaultBaseImage,
			ThanosDefaultBaseImage:     defaultTestConfig.ThanosDefaultBaseImage,
		}
		logger := prompkg.NewLogger()
		p := monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{},
		}

		cg, err := prompkg.NewConfigGenerator(logger, &p)
		require.NoError(t, err)

		sset, err := makeStatefulSet(
			"test",
			&p,
			testConfig,
			cg,
			nil,
			"",
			0,
			&operator.ShardedSecret{})
		require.NoError(t, err)
		return sset
	})
}

func TestAdditionalContainers(t *testing.T) {
	// The base to compare everything against
	baseSet, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{})
	require.NoError(t, err)

	// Add an extra container
	addSset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Containers: []v1.Container{
					{
						Name: "extra-container",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	require.Len(t, addSset.Spec.Template.Spec.Containers, len(baseSet.Spec.Template.Spec.Containers)+1, "container count mismatch")

	// Adding a new container with the same name results in a merge and just one container
	const existingContainerName = "prometheus"
	const containerImage = "madeUpContainerImage"
	modSset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Containers: []v1.Container{
					{
						Name:  existingContainerName,
						Image: containerImage,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, len(baseSet.Spec.Template.Spec.Containers), len(modSset.Spec.Template.Spec.Containers), "container count mismatch. container %s was added instead of merged", existingContainerName)

	// Check that adding a container with an existing name results in a single patched container.
	for _, c := range modSset.Spec.Template.Spec.Containers {
		require.False(t, (c.Name == existingContainerName && c.Image != containerImage), "expected container %s to have the image %s but got %s", existingContainerName, containerImage, c.Image)
	}
}

func TestWALCompression(t *testing.T) {
	var (
		tr = true
		fa = false
	)
	tests := []struct {
		version       string
		enabled       *bool
		expectedArg   string
		shouldContain bool
	}{
		// Nil should not have either flag.
		{"v2.10.0", nil, "--no-storage.tsdb.wal-compression", false},
		{"v2.10.0", nil, "--storage.tsdb.wal-compression", false},
		{"v2.10.0", &fa, "--no-storage.tsdb.wal-compression", false},
		{"v2.10.0", &tr, "--storage.tsdb.wal-compression", false},
		{"v2.11.0", nil, "--no-storage.tsdb.wal-compression", false},
		{"v2.11.0", nil, "--storage.tsdb.wal-compression", false},
		{"v2.11.0", &fa, "--no-storage.tsdb.wal-compression", true},
		{"v2.11.0", &tr, "--storage.tsdb.wal-compression", true},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version:        test.version,
					WALCompression: test.enabled,
				},
			},
		})
		require.NoError(t, err)

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		require.Equal(t, test.shouldContain, slices.Contains(promArgs, test.expectedArg))
	}
}

func TestTSDBAllowOverlappingBlocks(t *testing.T) {
	expectedArg := "--storage.tsdb.allow-overlapping-blocks"
	tests := []struct {
		version       string
		enabled       bool
		shouldContain bool
	}{
		{"v2.10.0", true, false},
		{"v2.11.0", true, true},
		{"v2.38.0", true, true},
		{"v2.39.0", true, false},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				AllowOverlappingBlocks: test.enabled,
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: test.version,
				},
			},
		})
		require.NoError(t, err)

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		require.Equal(t, test.shouldContain, slices.Contains(promArgs, expectedArg))
	}
}

func TestTSDBAllowOverlappingCompaction(t *testing.T) {
	expectedArg := "--no-storage.tsdb.allow-overlapping-compaction"
	tests := []struct {
		name                    string
		version                 string
		outOfOrderTimeWindow    monitoringv1.Duration
		objectStorageConfigFile *string
		shouldContain           bool
	}{
		{
			name:          "Prometheus version less than or equal to v2.55.0",
			version:       "v2.54.0",
			shouldContain: false,
		},
		{
			name:          "outOfOrderTimeWindow equal to 0s",
			version:       "v2.55.0",
			shouldContain: false,
		},
		{
			name:                    "Thanos is not object storage",
			version:                 "v2.55.0",
			outOfOrderTimeWindow:    "1s",
			objectStorageConfigFile: nil,
			shouldContain:           false,
		},
		{
			name:                    "Verify AllowOverlappingCompaction",
			version:                 "v2.55.0",
			outOfOrderTimeWindow:    "1s",
			objectStorageConfigFile: ptr.To("/etc/thanos.cfg"),
			shouldContain:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: test.version,
						TSDB: &monitoringv1.TSDBSpec{
							OutOfOrderTimeWindow: ptr.To(test.outOfOrderTimeWindow),
						},
					},
					Thanos: &monitoringv1.ThanosSpec{
						ListenLocal:             true,
						ObjectStorageConfigFile: test.objectStorageConfigFile,
					},
				},
			})
			require.NoError(t, err)

			promArgs := sset.Spec.Template.Spec.Containers[0].Args
			require.Equal(t, test.shouldContain, slices.Contains(promArgs, expectedArg))
		})
	}
}

func TestThanosListenLocal(t *testing.T) {
	for _, tc := range []struct {
		spec     monitoringv1.ThanosSpec
		expected []string
	}{
		{
			spec: monitoringv1.ThanosSpec{
				ListenLocal: true,
			},
			expected: []string{
				"--grpc-address=127.0.0.1:10901",
				"--http-address=127.0.0.1:10902",
			},
		},
		{
			spec: monitoringv1.ThanosSpec{
				GRPCListenLocal: true,
			},
			expected: []string{
				"--grpc-address=127.0.0.1:10901",
				"--http-address=:10902",
			},
		},
		{
			spec: monitoringv1.ThanosSpec{
				HTTPListenLocal: true,
			},
			expected: []string{
				"--grpc-address=:10901",
				"--http-address=127.0.0.1:10902",
			},
		},
		{
			spec: monitoringv1.ThanosSpec{},
			expected: []string{
				"--grpc-address=:10901",
				"--http-address=:10902",
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					Thanos: &tc.spec,
				},
			})
			require.NoError(t, err)

			for _, exp := range tc.expected {
				var found bool
				if slices.Contains(sset.Spec.Template.Spec.Containers[2].Args, exp) {
					found = true
				}

				require.True(t, found, "Expecting argument %q but not found in %v", exp, sset.Spec.Template.Spec.Containers[2].Args)
			}
		})
	}
}

func TestTerminationPolicy(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{Spec: monitoringv1.PrometheusSpec{}})
	require.NoError(t, err)

	for _, c := range sset.Spec.Template.Spec.Containers {
		require.Equal(t, v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy, "Unexpected TermintationMessagePolicy. Expected %v got %v", v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy)
	}
}

func TestEnableFeaturesWithOneFeature(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnableFeatures: []monitoringv1.EnableFeature{"exemplar-storage"},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--enable-feature=exemplar-storage" {
			found = true
		}
	}

	require.True(t, found, "Prometheus enabled feature is not correctly set.")
}

func TestEnableFeaturesWithMultipleFeature(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnableFeatures: []monitoringv1.EnableFeature{"exemplar-storage1", "exemplar-storage2"},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--enable-feature=exemplar-storage1,exemplar-storage2" {
			found = true
		}
	}

	require.True(t, found, "Prometheus enabled features are not correctly set.")
}

func TestWebPageTitle(t *testing.T) {
	pageTitle := "my-page-title"
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Web: &monitoringv1.PrometheusWebSpec{
					PageTitle: &pageTitle,
				},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.page-title=my-page-title" {
			found = true
		}
	}

	require.True(t, found, "Prometheus web page title is not correctly set.")
}

func TestMaxConnections(t *testing.T) {
	maxConnections := int32(600)
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Web: &monitoringv1.PrometheusWebSpec{
					MaxConnections: &maxConnections,
				},
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.max-connections=600" {
			found = true
		}
	}

	require.True(t, found, "Prometheus web max connections is not correctly set.")
}

func TestExpectedStatefulSetShardNames(t *testing.T) {
	replicas := int32(2)
	shards := int32(3)
	res := prompkg.ExpectedStatefulSetShardNames(&monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Shards:   &shards,
				Replicas: &replicas,
			},
		},
	})

	expected := []string{
		"prometheus-test",
		"prometheus-test-shard-1",
		"prometheus-test-shard-2",
	}

	for i, name := range expected {
		require.Equal(t, name, res[i], "Unexpected StatefulSet shard name")
	}
}

func TestExpectStatefulSetMinReadySeconds(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	})
	require.NoError(t, err)

	// assert defaults to zero if nil
	require.Equal(t, int32(0), sset.Spec.MinReadySeconds)

	sset, err = makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				MinReadySeconds: ptr.To(int32(5)),
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, int32(5), sset.Spec.MinReadySeconds)
}

func TestConfigReloader(t *testing.T) {
	expectedShardNum := 0
	logger := prompkg.NewLogger()
	p := monitoringv1.Prometheus{}

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		nil,
		"",
		int32(expectedShardNum),
		&operator.ShardedSecret{})
	require.NoError(t, err)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=http://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expectd container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			for _, env := range c.Env {
				require.False(t, (env.Name == "SHARD" && !reflect.DeepEqual(env.Value, strconv.Itoa(expectedShardNum))), "expectd shard value is %s, but found %s", strconv.Itoa(expectedShardNum), env.Value)
			}
		}
	}

	expectedArgsInitConfigReloader := []string{
		"--watch-interval=0",
		"--listen-address=:8080",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "init-config-reloader" {
			require.Equal(t, expectedArgsInitConfigReloader, c.Args, "expectd init container args are %s, but found %s", expectedArgsInitConfigReloader, c.Args)
			for _, env := range c.Env {
				require.False(t, (env.Name == "SHARD" && !reflect.DeepEqual(env.Value, strconv.Itoa(expectedShardNum))), "expectd shard value is %s, but found %s", strconv.Itoa(expectedShardNum), env.Value)
			}
		}
	}
}

func TestConfigReloaderWithSignal(t *testing.T) {
	expectedShardNum := 0
	logger := prompkg.NewLogger()
	p := monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ReloadStrategy: func(r monitoringv1.ReloadStrategyType) *monitoringv1.ReloadStrategyType { return &r }(monitoringv1.ProcessSignalReloadStrategyType),
			},
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		"test",
		&p,
		defaultTestConfig,
		cg,
		nil,
		"",
		int32(expectedShardNum),
		&operator.ShardedSecret{})
	require.NoError(t, err)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-method=signal",
		"--runtimeinfo-url=http://localhost:9090/api/v1/status/runtimeinfo",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		switch c.Name {
		case "config-reloader":
			require.Equal(t, expectedArgsConfigReloader, c.Args)
			for _, env := range c.Env {
				if env.Name == "SHARD" {
					require.Equal(t, strconv.Itoa(expectedShardNum), env.Value)
				}
			}

		case "prometheus":
			require.NotContains(t, c.Args, "--web.enable-lifecycle")
		}
	}

	expectedArgsInitConfigReloader := []string{
		"--watch-interval=0",
		"--listen-address=:8081",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.InitContainers {
		if c.Name == "init-config-reloader" {
			require.Equal(t, expectedArgsInitConfigReloader, c.Args)
			for _, env := range c.Env {
				if env.Name == "SHARD" {
					require.Equal(t, strconv.Itoa(expectedShardNum), env.Value)
				}
			}
		}
	}
}

func TestThanosGetConfigInterval(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				GetConfigInterval: "1m",
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "thanos-sidecar" {
			for _, flag := range container.Args {
				if flag == "--prometheus.get_config_interval=1m" {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Sidecar get_config_interval is not set when it should.")
}

func TestThanosGetConfigTimeout(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				GetConfigTimeout: "30s",
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "thanos-sidecar" {
			for _, flag := range container.Args {
				if flag == "--prometheus.get_config_timeout=30s" {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Sidecar get_config_timeout is not set when it should.")
}

func TestThanosReadyTimeout(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ReadyTimeout: "20m",
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "thanos-sidecar" {
			for _, flag := range container.Args {
				if flag == "--prometheus.ready_timeout=20m" {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Sidecar ready timeout not set when it should.")
}

func TestQueryLogFileVolumeMountPresent(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			QueryLogFile: "test.log",
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.True(t, found, "Volume for query log file not found.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Query log file not mounted.")
}

func TestQueryLogFileVolumeMountNotPresent(t *testing.T) {
	// An emptyDir is only mounted by the Operator if the given
	// path is only a base filename.
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			QueryLogFile: "/tmp/test.log",
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.False(t, found, "Volume for query log file found, when it shouldn't be.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.False(t, found, "Query log file mounted, when it shouldn't be.")
}

func TestScrapeFailureLogFileVolumeMountPresent(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeFailureLogFile: ptr.To("file.log"),
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.True(t, found, "Volume for scrape failure log file not found.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.True(t, found, "Scrape failure log file not mounted.")
}

func TestScrapeFailureLogFileVolumeMountNotPresent(t *testing.T) {
	// An emptyDir is only mounted by the Operator if the given
	// path is only a base filename.
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				ScrapeFailureLogFile: ptr.To("/tmp/file.log"),
			},
		},
	})
	require.NoError(t, err)

	found := false
	for _, volume := range sset.Spec.Template.Spec.Volumes {
		if volume.Name == prompkg.DefaultLogFileVolume {
			found = true
		}
	}

	require.False(t, found, "Volume for scrape failure file found, when it shouldn't be.")

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == prompkg.DefaultLogFileVolume {
					found = true
				}
			}
		}
	}

	require.False(t, found, "Scrape failure log file mounted, when it shouldn't be.")
}

func TestRemoteWriteReceiver(t *testing.T) {
	for _, tc := range []struct {
		version                   string
		enableRemoteWriteReceiver bool
		messageVersions           []monitoringv1.RemoteWriteMessageVersion

		expectedRemoteWriteReceiverFlag bool
		expectedMessageVersions         string
	}{
		// Remote write receiver not supported.
		{
			version:                   "2.32.0",
			enableRemoteWriteReceiver: true,
		},
		// Remote write receiver supported starting with v2.33.0.
		{
			version:                         "2.33.0",
			enableRemoteWriteReceiver:       true,
			expectedRemoteWriteReceiverFlag: true,
		},
		// Remote write receiver supported but not enabled.
		{
			version:                         "2.33.0",
			enableRemoteWriteReceiver:       false,
			expectedRemoteWriteReceiverFlag: false,
		},
		// Test higher version from which feature available
		{
			version:                         "2.33.5",
			enableRemoteWriteReceiver:       true,
			expectedRemoteWriteReceiverFlag: true,
		},
		// RemoteWriteMessageVersions not supported.
		{
			version:                         "2.53.0",
			enableRemoteWriteReceiver:       true,
			expectedRemoteWriteReceiverFlag: true,
			messageVersions: []monitoringv1.RemoteWriteMessageVersion{
				monitoringv1.RemoteWriteMessageVersion2_0,
			},
		},
		// RemoteWriteMessageVersions supported and set to one value.
		{
			version:                   "2.54.0",
			enableRemoteWriteReceiver: true,
			messageVersions: []monitoringv1.RemoteWriteMessageVersion{
				monitoringv1.RemoteWriteMessageVersion2_0,
			},
			expectedRemoteWriteReceiverFlag: true,
			expectedMessageVersions:         "io.prometheus.write.v2.Request",
		},
		// RemoteWriteMessageVersions supported and set to 2 values.
		{
			version:                   "2.54.0",
			enableRemoteWriteReceiver: true,
			messageVersions: []monitoringv1.RemoteWriteMessageVersion{
				monitoringv1.RemoteWriteMessageVersion1_0,
				monitoringv1.RemoteWriteMessageVersion2_0,
			},
			expectedRemoteWriteReceiverFlag: true,
			expectedMessageVersions:         "prometheus.WriteRequest,io.prometheus.write.v2.Request",
		},
	} {
		t.Run(tc.version, func(t *testing.T) {
			p := monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:                            tc.version,
						EnableRemoteWriteReceiver:          tc.enableRemoteWriteReceiver,
						RemoteWriteReceiverMessageVersions: tc.messageVersions,
					},
				},
			}
			sset, err := makeStatefulSetFromPrometheus(p)
			require.NoError(t, err)

			var (
				enabled         bool
				messageVersions string
			)
			for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
				flag = strings.TrimPrefix(flag, "--")
				values := strings.Split(flag, "=")
				switch values[0] {
				case "web.enable-remote-write-receiver":
					enabled = true
				case "web.remote-write-receiver.accepted-protobuf-messages":
					messageVersions = values[1]
				}
			}

			require.Equal(t, tc.expectedRemoteWriteReceiverFlag, enabled)
			require.Equal(t, tc.expectedMessageVersions, messageVersions)
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
	serviceAccountName := "prometheus-sa"
	hostAliases := []monitoringv1.HostAlias{
		{
			Hostnames: []string{"foo.com"},
			IP:        "1.1.1.1",
		},
	}
	imagePullPolicy := v1.PullAlways
	imagePullSecrets := []v1.LocalObjectReference{
		{
			Name: "registry-secret",
		},
	}

	hostNetwork := false
	hostUsers := true

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				NodeSelector:       nodeSelector,
				Affinity:           &affinity,
				Tolerations:        tolerations,
				SecurityContext:    &securityContext,
				PriorityClassName:  priorityClassName,
				ServiceAccountName: serviceAccountName,
				HostAliases:        hostAliases,
				ImagePullPolicy:    imagePullPolicy,
				ImagePullSecrets:   imagePullSecrets,
				HostNetwork:        hostNetwork,
				HostUsers:          ptr.To(true),
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, nodeSelector, sset.Spec.Template.Spec.NodeSelector, "expected node selector to match, want %v, got %v", nodeSelector, sset.Spec.Template.Spec.NodeSelector)
	require.Equal(t, affinity, *sset.Spec.Template.Spec.Affinity, "expected affinity to match, want %v, got %v", affinity, *sset.Spec.Template.Spec.Affinity)
	require.Equal(t, tolerations, sset.Spec.Template.Spec.Tolerations, "expected tolerations to match, want %v, got %v", tolerations, sset.Spec.Template.Spec.Tolerations)
	require.Equal(t, securityContext, *sset.Spec.Template.Spec.SecurityContext, "expected security context  to match, want %v, got %v", securityContext, *sset.Spec.Template.Spec.SecurityContext)
	require.Equal(t, priorityClassName, sset.Spec.Template.Spec.PriorityClassName, "expected priority class name to match, want %s, got %s", priorityClassName, sset.Spec.Template.Spec.PriorityClassName)
	require.Equal(t, serviceAccountName, sset.Spec.Template.Spec.ServiceAccountName, "expected service account name to match, want %s, got %s", serviceAccountName, sset.Spec.Template.Spec.ServiceAccountName)
	require.Len(t, sset.Spec.Template.Spec.HostAliases, len(hostAliases), "expected length of host aliases to match, want %d, got %d", len(hostAliases), len(sset.Spec.Template.Spec.HostAliases))
	require.Equal(t, hostUsers, *sset.Spec.Template.Spec.HostUsers, "expected host users to match, want %s, got %s", hostUsers, sset.Spec.Template.Spec.HostUsers)
	for _, initContainer := range sset.Spec.Template.Spec.InitContainers {
		require.Equal(t, imagePullPolicy, initContainer.ImagePullPolicy, "expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, initContainer.ImagePullPolicy)
	}
	for _, container := range sset.Spec.Template.Spec.Containers {
		require.Equal(t, imagePullPolicy, container.ImagePullPolicy, "expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, container.ImagePullPolicy)
	}
	require.Equal(t, imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets, "expected image pull secrets to match, want %s, got %s", imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets)
	require.Equal(t, hostNetwork, sset.Spec.Template.Spec.HostNetwork, "expected hostNetwork configuration to match but failed")
}

func TestPrometheusAdditionalArgsNoError(t *testing.T) {
	expectedPrometheusArgsV3 := []string{
		"--config.file=/etc/prometheus/config_out/prometheus.env.yaml",
		"--web.enable-lifecycle",
		"--web.route-prefix=/",
		"--storage.tsdb.retention.time=24h",
		"--storage.tsdb.path=/prometheus",
		"--web.config.file=/etc/prometheus/web_config/web-config.yaml",
		"--scrape.discovery-reload-interval=30s",
		"--storage.tsdb.no-lockfile",
	}

	expectedPrometheusArgsV2 := []string{
		"--config.file=/etc/prometheus/config_out/prometheus.env.yaml",
		"--web.console.templates=/etc/prometheus/consoles",
		"--web.console.libraries=/etc/prometheus/console_libraries",
		"--web.enable-lifecycle",
		"--web.route-prefix=/",
		"--storage.tsdb.retention.time=24h",
		"--storage.tsdb.path=/prometheus",
		"--web.config.file=/etc/prometheus/web_config/web-config.yaml",
		"--scrape.discovery-reload-interval=30s",
		"--storage.tsdb.no-lockfile",
	}

	argTests := []struct {
		version      string
		expectedArgs []string
	}{
		{
			version:      "v2.54.0",
			expectedArgs: expectedPrometheusArgsV2,
		},
		{
			version:      "v3.0.0",
			expectedArgs: expectedPrometheusArgsV3,
		},
	}

	for _, argTest := range argTests {
		t.Run(argTest.version, func(t *testing.T) {

			labels := map[string]string{
				"testlabel": "testlabelvalue",
			}
			annotations := map[string]string{
				"testannotation": "testannotationvalue",
			}

			p := monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						AdditionalArgs: []monitoringv1.Argument{
							{
								Name:  "scrape.discovery-reload-interval",
								Value: "30s",
							},
							{
								Name: "storage.tsdb.no-lockfile",
							},
						},
						Version: argTest.version,
					},
				},
			}
			sset, err := makeStatefulSetFromPrometheus(p)
			require.NoError(t, err)
			ssetContainerArgs := sset.Spec.Template.Spec.Containers[0].Args
			// web.console.templates and web.console.libraries should be present in prometheus versisons < 3
			require.Equal(t, argTest.expectedArgs, ssetContainerArgs, "expected Prometheus container args to match, want %s, got %s", argTest.expectedArgs, ssetContainerArgs)
		})

	}
}

func TestPrometheusAdditionalArgsDuplicate(t *testing.T) {
	expectedErrorMsg := "can't set arguments which are already managed by the operator: config.file"

	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	_, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				AdditionalArgs: []monitoringv1.Argument{
					{
						Name:  "config.file",
						Value: "/foo/bar.yaml",
					},
				},
			},
		},
	})
	require.Error(t, err)

	require.Contains(t, err.Error(), expectedErrorMsg, "expected the following text to be present in the error msg: %s", expectedErrorMsg)
}

func TestRuntimeGOGCEnvVar(t *testing.T) {
	for _, tc := range []struct {
		scenario       string
		version        string
		gogc           *int32
		expectedEnvVar bool
	}{
		{
			scenario:       "Prometheus < 2.53.0",
			version:        "v2.51.2",
			gogc:           ptr.To(int32(50)),
			expectedEnvVar: true,
		},
		{
			scenario:       "Prometheus > 2.53.0",
			version:        "v2.54.0",
			gogc:           ptr.To(int32(50)),
			expectedEnvVar: false,
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.scenario), func(t *testing.T) {
			ss, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.version,
						Runtime: &monitoringv1.RuntimeConfig{
							GoGC: tc.gogc,
						},
					},
				},
			})

			var containsEnvVar bool
			for _, env := range ss.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "GOGC" {
					if env.Value == fmt.Sprintf("%d", *tc.gogc) {
						containsEnvVar = true
						break
					}
				}
			}

			require.NoError(t, err)
			if tc.expectedEnvVar {
				require.True(t, containsEnvVar, "Prometheus is missing expected GOGC env var with correct value")
			}

			if !tc.expectedEnvVar {
				require.False(t, containsEnvVar, "Prometheus didn't expect GOGC env var for this version of Prometheus")
			}
		})
	}
}

func TestPrometheusAdditionalBinaryArgsDuplicate(t *testing.T) {
	expectedErrorMsg := "can't set arguments which are already managed by the operator: web.enable-lifecycle"

	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	_, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				AdditionalArgs: []monitoringv1.Argument{
					{
						Name: "web.enable-lifecycle",
					},
				},
			},
		},
	})
	require.Error(t, err)

	require.Contains(t, err.Error(), expectedErrorMsg, "expected the following text to be present in the error msg: %s", expectedErrorMsg)
}

func TestPrometheusAdditionalNoPrefixArgsDuplicate(t *testing.T) {
	expectedErrorMsg := "can't set arguments which are already managed by the operator: storage.tsdb.wal-compression"
	walCompression := new(bool)
	*walCompression = true

	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	_, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				WALCompression: walCompression,
				AdditionalArgs: []monitoringv1.Argument{
					{
						Name: "no-storage.tsdb.wal-compression",
					},
				},
			},
		},
	})
	require.Error(t, err)

	require.Contains(t, err.Error(), expectedErrorMsg, "expected the following text to be present in the error msg: %s", expectedErrorMsg)
}

func TestThanosAdditionalArgsNoError(t *testing.T) {
	expectedThanosArgs := []string{
		"sidecar",
		"--prometheus.url=http://localhost:9090/",
		"--grpc-address=:10901",
		"--http-address=:10902",
		"--log.level=info",
		"--prometheus.http-client-file=/etc/thanos/config/prometheus.http-client-file.yaml",
		"--reloader.watch-interval=5m",
	}

	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				LogLevel: "info",
				AdditionalArgs: []monitoringv1.Argument{
					{
						Name:  "reloader.watch-interval",
						Value: "5m",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	ssetContainerArgs := sset.Spec.Template.Spec.Containers[2].Args
	require.Equal(t, expectedThanosArgs, ssetContainerArgs, "expected Thanos container args to match, want %s, got %s", expectedThanosArgs, ssetContainerArgs)
}

func TestThanosAdditionalArgsDuplicate(t *testing.T) {
	expectedErrorMsg := "can't set arguments which are already managed by the operator: log.level"

	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	_, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				LogLevel: "info",
				AdditionalArgs: []monitoringv1.Argument{
					{
						Name:  "log.level",
						Value: "error",
					},
				},
			},
		},
	})
	require.Error(t, err)

	require.Contains(t, err.Error(), expectedErrorMsg, "expected the following text to be present in the error msg: %s", expectedErrorMsg)
}

func TestPrometheusQuerySpec(t *testing.T) {
	for _, tc := range []struct {
		name string

		lookbackDelta  *string
		maxConcurrency *int32
		maxSamples     *int32
		timeout        *monitoringv1.Duration
		version        string

		expected []string
	}{
		{
			name:     "default",
			expected: []string{},
		},
		{
			name:           "all values provided",
			lookbackDelta:  ptr.To("2m"),
			maxConcurrency: ptr.To(int32(10)),
			maxSamples:     ptr.To(int32(10000)),
			timeout:        ptr.To(monitoringv1.Duration("1m")),

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.max-concurrency=10",
				"--query.max-samples=10000",
				"--query.timeout=1m",
			},
		},
		{
			name:           "zero values are skipped",
			lookbackDelta:  ptr.To("2m"),
			maxConcurrency: ptr.To(int32(0)),
			maxSamples:     ptr.To(int32(0)),
			timeout:        ptr.To(monitoringv1.Duration("1m")),

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.timeout=1m",
			},
		},
		{
			name:           "maxConcurrency set to 1",
			maxConcurrency: ptr.To(int32(1)),

			expected: []string{
				"--query.max-concurrency=1",
			},
		},
		{
			name:           "max samples skipped if version < 2.5",
			lookbackDelta:  ptr.To("2m"),
			maxConcurrency: ptr.To(int32(10)),
			maxSamples:     ptr.To(int32(10000)),
			timeout:        ptr.To(monitoringv1.Duration("1m")),
			version:        "v2.4.0",

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.max-concurrency=10",
				"--query.timeout=1m",
			},
		},
		{
			name:           "max samples not skipped if version > 2.5",
			lookbackDelta:  ptr.To("2m"),
			maxConcurrency: ptr.To(int32(10)),
			maxSamples:     ptr.To(int32(10000)),
			timeout:        ptr.To(monitoringv1.Duration("1m")),
			version:        "v2.5.0",

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.max-concurrency=10",
				"--query.max-samples=10000",
				"--query.timeout=1m",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.version,
					},
					Query: &monitoringv1.QuerySpec{
						LookbackDelta:  tc.lookbackDelta,
						MaxConcurrency: tc.maxConcurrency,
						MaxSamples:     tc.maxSamples,
						Timeout:        tc.timeout,
					},
				},
			})
			require.NoError(t, err)

			for _, arg := range []string{
				"--query.lookback-delta",
				"--query.max-concurrency",
				"--query.max-samples",
				"--query.timeout",
			} {
				var containerArg string
				for _, a := range sset.Spec.Template.Spec.Containers[0].Args {
					if strings.HasPrefix(a, arg) {
						containerArg = a
						break
					}
				}

				var expected string
				for _, exp := range tc.expected {
					if strings.HasPrefix(exp, arg) {
						expected = exp
						break
					}
				}

				if expected == "" {
					require.Equal(t, "", containerArg, "found %q while not expected", containerArg)
					continue
				}

				require.Equal(t, expected, containerArg, "expected %q to be found but got %q", expected, containerArg)
			}
		})
	}
}

func TestSecurityContextCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec monitoringv1.PrometheusSpec
	}{
		{
			name: "default",
			spec: monitoringv1.PrometheusSpec{},
		},
		{
			name: "Thanos sidecar",
			spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{},
			},
		},
		{
			name: "Thanos sidecar with object storage",
			spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					ObjectStorageConfigFile: ptr.To("/etc/thanos.cfg"),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{Spec: tc.spec})
			require.NoError(t, err)

			exp := 2
			if tc.spec.Thanos != nil {
				exp++
			}
			require.Len(t, sset.Spec.Template.Spec.Containers, exp, "Expecting %d containers, got %d", exp, len(sset.Spec.Template.Spec.Containers))

			for _, c := range sset.Spec.Template.Spec.Containers {
				require.Empty(t, c.SecurityContext.Capabilities.Add, "Expecting 0 added capabilities, got %d", len(c.SecurityContext.Capabilities.Add))
				require.Len(t, c.SecurityContext.Capabilities.Drop, 1, "Expecting 1 dropped capabilities, got %d", len(c.SecurityContext.Capabilities.Drop))

				require.Equal(t, "ALL", string(c.SecurityContext.Capabilities.Drop[0]), "Expecting ALL dropped capability, got %s", c.SecurityContext.Capabilities.Drop[0])
			}
		})
	}
}

func TestPodHostNetworkConfig(t *testing.T) {
	hostNetwork := true
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				HostNetwork: hostNetwork,
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, hostNetwork, sset.Spec.Template.Spec.HostNetwork, "expected hostNetwork configuration to match but failed")
	require.Equal(t, v1.DNSClusterFirstWithHostNet, sset.Spec.Template.Spec.DNSPolicy, "expected DNSPolicy configuration to match due to hostNetwork but failed")
}

func TestPersistentVolumeClaimRetentionPolicy(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
					WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
					WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted, "expected persistentVolumeClaimDeletePolicy.WhenDeleted to be %s but got %s", appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted)
	require.Equal(t, appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled, "expected persistentVolumeClaimDeletePolicy.WhenScaled to be %s but got %s", appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled)
}

func TestPodTopologySpreadConstraintWithAdditionalLabels(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec monitoringv1.PrometheusSpec
		tsc  v1.TopologySpreadConstraint
	}{
		{
			name: "without labelSelector and additionalLabels",
			spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
			},
		},
		{
			name: "with labelSelector and without additionalLabels",
			spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "prometheus",
					},
				},
			},
		},
		{
			name: "with labelSelector and additionalLabels as ShardAndNameResource",
			spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							AdditionalLabelSelectors: ptr.To(monitoringv1.ShardAndResourceNameLabelSelector),
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":                            "prometheus",
						"app.kubernetes.io/instance":     "test",
						"app.kubernetes.io/managed-by":   "prometheus-operator",
						"prometheus":                     "test",
						prompkg.ShardLabelName:           "0",
						prompkg.PrometheusNameLabelName:  "test",
						operator.ApplicationNameLabelKey: "prometheus",
					},
				},
			},
		},
		{
			name: "with labelSelector and additionalLabels as ResourceName",
			spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					TopologySpreadConstraints: []monitoringv1.TopologySpreadConstraint{
						{
							AdditionalLabelSelectors: ptr.To(monitoringv1.ResourceNameLabelSelector),
							CoreV1TopologySpreadConstraint: monitoringv1.CoreV1TopologySpreadConstraint{
								MaxSkew:           1,
								TopologyKey:       "kubernetes.io/hostname",
								WhenUnsatisfiable: v1.DoNotSchedule,
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "prometheus",
									},
								},
							},
						},
					},
				},
			},
			tsc: v1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: v1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":                            "prometheus",
						"app.kubernetes.io/instance":     "test",
						"app.kubernetes.io/managed-by":   "prometheus-operator",
						"prometheus":                     "test",
						prompkg.PrometheusNameLabelName:  "test",
						operator.ApplicationNameLabelKey: "prometheus",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sts, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns-test",
				},
				Spec: tc.spec,
			})

			require.NoError(t, err)

			assert.NotEmpty(t, sts.Spec.Template.Spec.TopologySpreadConstraints)
			assert.Equal(t, tc.tsc, sts.Spec.Template.Spec.TopologySpreadConstraints[0])
		})
	}
}

func TestStartupProbeTimeoutSeconds(t *testing.T) {
	tests := []struct {
		maximumStartupDurationSeconds   *int32
		expectedStartupPeriodSeconds    int32
		expectedStartupFailureThreshold int32
	}{
		{
			maximumStartupDurationSeconds:   nil,
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(600)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 10,
		},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					MaximumStartupDurationSeconds: test.maximumStartupDurationSeconds,
				},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, sset.Spec.Template.Spec.Containers[0].StartupProbe)
		require.Equal(t, test.expectedStartupPeriodSeconds, sset.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, sset.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold)
	}
}

func TestIfThanosVersionDontHaveHttpClientFlag(t *testing.T) {
	version := "v0.23.0"

	for _, tc := range []struct {
		name string
		spec monitoringv1.PrometheusSpec
	}{
		{
			name: "default",
			spec: monitoringv1.PrometheusSpec{},
		},
		{
			name: "Thanos sidecar",
			spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &version,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{Spec: tc.spec})
			require.NoError(t, err)
			for _, c := range sset.Spec.Template.Spec.Containers {
				for _, arg := range c.Args {
					require.NotContains(t, arg, "http-client", "Expecting http-client flag to not be present in Thanos sidecar")
				}
			}
		})
	}
}

func TestThanosWithPrometheusHTTPClientConfigFile(t *testing.T) {
	version := "0.24.0"

	for _, tc := range []struct {
		name string
		spec monitoringv1.PrometheusSpec
	}{
		{
			name: "thanos sidecar with prometheus.http-client-file",
			spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &version,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := monitoringv1.Prometheus{Spec: tc.spec}
			sset, err := makeStatefulSetFromPrometheus(p)
			require.NoError(t, err)
			for _, v := range sset.Spec.Template.Spec.Volumes {
				if v.Name == thanosPrometheusHTTPClientConfigSecretNameSuffix {
					require.Equal(t, v.VolumeSource.Secret.SecretName, thanosPrometheusHTTPClientConfigSecretName(&p))
				}
			}
			for _, c := range sset.Spec.Template.Spec.Containers {
				if c.Name == "thanos-sidecar" {
					require.NotEmpty(t, c.VolumeMounts)
					require.Equal(t, thanosPrometheusHTTPClientConfigSecretNameSuffix, c.VolumeMounts[0].Name)
				}
			}
		})
	}
}

func TestAutomountServiceAccountToken(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		automountServiceAccountToken *bool
		expectedValue                bool
	}{
		{
			name:                         "automountServiceAccountToken not set",
			automountServiceAccountToken: nil,
			expectedValue:                true,
		},
		{
			name:                         "automountServiceAccountToken set to true",
			automountServiceAccountToken: ptr.To(true),
			expectedValue:                true,
		},
		{
			name:                         "automountServiceAccountToken set to false",
			automountServiceAccountToken: ptr.To(false),
			expectedValue:                false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						AutomountServiceAccountToken: tc.automountServiceAccountToken,
					},
				},
			})
			require.NoError(t, err)

			require.NotNil(t, sset.Spec.Template.Spec.AutomountServiceAccountToken, "expected automountServiceAccountToken to be set")

			require.Equal(t, tc.expectedValue, *sset.Spec.Template.Spec.AutomountServiceAccountToken, "expected automountServiceAccountToken to be %v", tc.expectedValue)
		})
	}
}

func TestDNSPolicyAndDNSConfig(t *testing.T) {
	tests := []struct {
		name              string
		dnsPolicy         v1.DNSPolicy
		dnsConfig         *v1.PodDNSConfig
		expectedDNSPolicy v1.DNSPolicy
		expectedDNSConfig *v1.PodDNSConfig
	}{
		{
			name:              "Default DNSPolicy and DNSConfig",
			dnsPolicy:         v1.DNSClusterFirst,
			dnsConfig:         nil,
			expectedDNSPolicy: v1.DNSClusterFirst,
			expectedDNSConfig: nil,
		},
		{
			name:              "Custom DNSPolicy",
			dnsPolicy:         v1.DNSDefault,
			dnsConfig:         nil,
			expectedDNSPolicy: v1.DNSDefault,
			expectedDNSConfig: nil,
		},
		{
			name:      "Custom DNSConfig",
			dnsPolicy: v1.DNSClusterFirst,
			dnsConfig: &v1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
				Searches:    []string{"custom.svc.cluster.local"},
			},
			expectedDNSPolicy: v1.DNSClusterFirst,
			expectedDNSConfig: &v1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
				Searches:    []string{"custom.svc.cluster.local"},
			},
		},
		{
			name:      "Custom DNS Policy with Search Domains",
			dnsPolicy: v1.DNSDefault,
			dnsConfig: &v1.PodDNSConfig{
				Searches: []string{"kitsos.com", "kitsos.org"},
			},
			expectedDNSPolicy: v1.DNSDefault,
			expectedDNSConfig: &v1.PodDNSConfig{
				Searches: []string{"kitsos.com", "kitsos.org"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			monitoringDNSPolicyPtr := ptr.To(monitoringv1.DNSPolicy(test.dnsPolicy))

			var monitoringDNSConfig *monitoringv1.PodDNSConfig
			if test.dnsConfig != nil {
				monitoringDNSConfig = &monitoringv1.PodDNSConfig{
					Nameservers: test.dnsConfig.Nameservers,
					Searches:    test.dnsConfig.Searches,
				}
			}

			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						DNSPolicy: monitoringDNSPolicyPtr,
						DNSConfig: monitoringDNSConfig,
					},
				},
			})
			require.NoError(t, err)

			require.Equal(t, test.expectedDNSPolicy, sset.Spec.Template.Spec.DNSPolicy, "expected DNSPolicy to match, want %v, got %v", test.expectedDNSPolicy, sset.Spec.Template.Spec.DNSPolicy)
			if test.expectedDNSConfig != nil {
				require.NotNil(t, sset.Spec.Template.Spec.DNSConfig, "expected DNSConfig to be set")
				require.Equal(t, test.expectedDNSConfig.Nameservers, sset.Spec.Template.Spec.DNSConfig.Nameservers, "expected DNSConfig Nameservers to match, want %v, got %v", test.expectedDNSConfig.Nameservers, sset.Spec.Template.Spec.DNSConfig.Nameservers)
				require.Equal(t, test.expectedDNSConfig.Searches, sset.Spec.Template.Spec.DNSConfig.Searches, "expected DNSConfig Searches to match, want %v, got %v", test.expectedDNSConfig.Searches, sset.Spec.Template.Spec.DNSConfig.Searches)
			} else {
				require.Nil(t, sset.Spec.Template.Spec.DNSConfig, "expected DNSConfig to be nil")
			}
		})
	}
}

func TestStatefulSetenableServiceLinks(t *testing.T) {
	tests := []struct {
		enableServiceLinks         *bool
		expectedEnableServiceLinks *bool
	}{
		{enableServiceLinks: ptr.To(false), expectedEnableServiceLinks: ptr.To(false)},
		{enableServiceLinks: ptr.To(true), expectedEnableServiceLinks: ptr.To(true)},
		{enableServiceLinks: nil, expectedEnableServiceLinks: nil},
	}

	for _, test := range tests {
		sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					EnableServiceLinks: test.enableServiceLinks,
				},
			},
		})
		require.NoError(t, err)

		if test.expectedEnableServiceLinks != nil {
			require.NotNil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to be non-nil")
			require.Equal(t, *test.expectedEnableServiceLinks, *sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to match")
		} else {
			require.Nil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to be nil")
		}
	}
}

func TestStatefulPodManagementPolicy(t *testing.T) {
	for _, tc := range []struct {
		podManagementPolicy *monitoringv1.PodManagementPolicyType
		exp                 appsv1.PodManagementPolicyType
	}{
		{
			podManagementPolicy: nil,
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.ParallelPodManagement),
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.OrderedReadyPodManagement),
			exp:                 appsv1.OrderedReadyPodManagement,
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						PodManagementPolicy: tc.podManagementPolicy,
					},
				},
			})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.PodManagementPolicy)
		})
	}
}

func TestStatefulSetUpdateStrategy(t *testing.T) {
	for _, tc := range []struct {
		updateStrategy *monitoringv1.StatefulSetUpdateStrategy
		exp            appsv1.StatefulSetUpdateStrategy
	}{
		{
			updateStrategy: nil,
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.OnDeleteStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						UpdateStrategy: tc.updateStrategy,
					},
				},
			})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.UpdateStrategy)
		})
	}
}
