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
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/kylelemons/godebug/pretty"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

var (
	defaultTestConfig = &operator.Config{
		LocalHost:                  "localhost",
		ReloaderConfig:             operator.DefaultReloaderTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: operator.DefaultPrometheusBaseImage,
		ThanosDefaultBaseImage:     operator.DefaultThanosBaseImage,
	}
)

func newLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stderr), level.AllowWarn())
}

func makeStatefulSetFromPrometheus(p monitoringv1.Prometheus) (*appsv1.StatefulSet, error) {
	logger := newLogger()

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	if err != nil {
		return nil, err
	}

	return makeStatefulSet(
		logger,
		"test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		defaultTestConfig,
		cg,
		nil,
		"",
		0,
		nil)
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
		"prometheus-operator-input-hash": "",
		"testannotation":                 "testannotationvalue",
	}

	expectedStatefulSetLabels := map[string]string{
		"testlabel":                    "testlabelvalue",
		"operator.prometheus.io/name":  "",
		"operator.prometheus.io/shard": "0",
		"operator.prometheus.io/mode":  "server",
	}

	expectedPodLabels := map[string]string{
		"prometheus":                   "",
		"app.kubernetes.io/name":       "prometheus",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultPrometheusVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "",
		"operator.prometheus.io/name":  "",
		"operator.prometheus.io/shard": "0",
	}

	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	})
	require.NoError(t, err)

	if !reflect.DeepEqual(expectedStatefulSetLabels, sset.Labels) {
		t.Log(pretty.Compare(expectedStatefulSetLabels, sset.Labels))
		t.Fatal("Labels are not properly being propagated to the StatefulSet")
	}

	if !reflect.DeepEqual(expectedStatefulSetAnnotations, sset.Annotations) {
		t.Log(pretty.Compare(expectedStatefulSetAnnotations, sset.Annotations))
		t.Fatal("Annotations are not properly being propagated to the StatefulSet")
	}

	if !reflect.DeepEqual(expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels) {
		t.Log(pretty.Compare(expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels))
		t.Fatal("Labels are not properly being propagated to the Pod")
	}
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

	if val, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]; !ok || val != "testvalue" {
		t.Fatal("Pod labels are not properly propagated")
	}
	if val, ok := sset.Spec.Template.ObjectMeta.Annotations["testannotation"]; !ok || val != "testvalue" {
		t.Fatal("Pod annotations are not properly propagated")
	}
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

	if sset.Spec.Selector.MatchLabels["testlabel"] == "testvalue" {
		t.Fatal("Pod Selector are not properly propagated")
	}
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
	if !reflect.DeepEqual(*pvc.Spec.StorageClassName, *ssetPvc.Spec.StorageClassName) {
		t.Fatal("Error adding PVC Spec to StatefulSetSpec")
	}

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
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir == nil || !reflect.DeepEqual(emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium) {
		t.Fatal("Error adding EmptyDir Spec to StatefulSetSpec")
	}
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
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral == nil ||
		!reflect.DeepEqual(ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName) {
		t.Fatal("Error adding Ephemeral Spec to StatefulSetSpec")
	}
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
									SubPath:   "",
								},
								{
									Name:             "tls-assets",
									ReadOnly:         true,
									MountPath:        "/etc/prometheus/certs",
									SubPath:          "",
									MountPropagation: nil,
									SubPathExpr:      "",
								},
								{
									Name:      "prometheus-volume-init-test-db",
									ReadOnly:  false,
									MountPath: "/prometheus",
									SubPath:   "",
								},
								{
									Name:      "secret-test-secret1",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/secrets/test-secret1",
									SubPath:   "",
								},
								{
									Name:      "rules-configmap-one",
									ReadOnly:  false,
									MountPath: "/etc/prometheus/rules/rules-configmap-one",
									SubPath:   "",
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

	logger := newLogger()

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		logger,
		"volume-init-test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		defaultTestConfig,
		cg,
		[]string{"rules-configmap-one"},
		"",
		0,
		[]string{prompkg.TLSAssetsSecretName(&p) + "-0"})
	require.NoError(t, err)

	if !reflect.DeepEqual(expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes) {
		fmt.Println(pretty.Compare(expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes))
		t.Fatal("expected volumes to match")
	}

	if !reflect.DeepEqual(expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts) {
		fmt.Println(pretty.Compare(expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts))
		t.Fatal("expected volume mounts to match")
	}
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
	if !cmVolumeFound {
		t.Fatal("ConfigMap volume not found")
	}

	cmMounted := false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if strings.HasPrefix(v.Name, "configmap-test-cm1") && v.MountPath == "/etc/prometheus/configmaps/test-cm1" {
			cmMounted = true
		}
	}
	if !cmMounted {
		t.Fatal("ConfigMap volume not mounted")
	}
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

	if !found {
		t.Fatal("Prometheus not listening on loopback when it should.")
	}

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
	if !reflect.DeepEqual(actualStartupProbe, expectedStartupProbe) {
		t.Fatalf("Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)
	}

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 0 {
		t.Fatal("Prometheus container should have 0 ports defined")
	}
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
	if !reflect.DeepEqual(actualStartupProbe, expectedStartupProbe) {
		t.Fatalf("Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)
	}

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9090/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	if !reloadURLFound {
		t.Fatalf("expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)
	}

	expectedThanosSidecarPrometheusURL := "--prometheus.url=https://localhost:9090/"
	prometheusURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
		if arg == expectedThanosSidecarPrometheusURL {
			prometheusURLFound = true
		}
	}
	if !prometheusURLFound {
		t.Fatalf("expected to find arg %s in thanos sidecar", expectedThanosSidecarPrometheusURL)
	}

	fmt.Println(sset.Spec.Template.Spec.Containers[2].Args)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=https://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
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
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
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
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Explicit image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
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
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
}

func TestPrometheusDefaultBaseImageFlag(t *testing.T) {
	operatorConfig := &operator.Config{
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

	logger := newLogger()
	p := monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		logger,
		"test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		operatorConfig,
		cg,
		nil,
		"",
		0,
		nil)
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/prometheus/prometheus" + ":" + operator.DefaultPrometheusVersion
	if image != expected {
		t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
}

func TestThanosDefaultBaseImageFlag(t *testing.T) {
	thanosBaseImageConfig := &operator.Config{
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
	logger := newLogger()
	p := monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		logger,
		"test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		thanosBaseImageConfig,
		cg,
		nil,
		"",
		0,
		nil)
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[2].Image
	expected := "nondefaultuseflag/quay.io/thanos/thanos" + ":" + operator.DefaultThanosVersion
	if image != expected {
		t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
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
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
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
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
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
		if image != expected {
			t.Fatalf("Explicit Thanos image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
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
	if res.Limits != nil || res.Requests != nil {
		t.Fatalf("Unexpected resources defined. \n\nExpected: nil\n\nGot: %v, %v", res.Limits, res.Requests)
	}
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
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected resources defined. \n\nExpected: %v\n\nGot: %v", expected, actual)
	}
}

func TestThanosNoObjectStorage(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	})
	require.NoError(t, err)

	if sset.Spec.Template.Spec.Containers[0].Name != "prometheus" {
		t.Fatalf("expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	if sset.Spec.Template.Spec.Containers[2].Name != "thanos-sidecar" {
		t.Fatalf("expected 3rd container to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)
	}

	for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
		if strings.HasPrefix(arg, "--storage.tsdb.max-block-duration=2h") {
			t.Fatal("Prometheus compaction should be disabled")
		}
	}

	for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
		if strings.HasPrefix(arg, "--tsdb.path=") {
			t.Fatal("--tsdb.path argument should not be given to the Thanos sidecar")
		}
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

	if sset.Spec.Template.Spec.Containers[0].Name != "prometheus" {
		t.Fatalf("expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	if sset.Spec.Template.Spec.Containers[2].Name != "thanos-sidecar" {
		t.Fatalf("expected 3rd containers to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[2].Env {
		if env.Name == "OBJSTORE_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos sidecar is missing expected OBJSTORE_CONFIG env var with correct value")
	}

	{
		var containsArg bool
		const expectedArg = "--objstore.config=$(OBJSTORE_CONFIG)"
		for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Thanos sidecar is missing expected argument: %s", expectedArg)
		}
	}
	{
		var containsArg bool
		const expectedArg = "--storage.tsdb.max-block-duration=2h"
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Prometheus is missing expected argument: %s", expectedArg)
		}
	}

	{
		var found bool
		for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
			if strings.HasPrefix(arg, "--tsdb.path=") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("--tsdb.path argument should be given to the Thanos sidecar, got %q", strings.Join(sset.Spec.Template.Spec.Containers[3].Args, " "))
		}
	}

	{
		var found bool
		for _, vol := range sset.Spec.Template.Spec.Containers[2].VolumeMounts {
			if vol.MountPath == prompkg.StorageDir {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Prometheus data volume should be mounted in the Thanos sidecar")
		}
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
		var containsArg bool
		expectedArg := "--objstore.config-file=" + testPath
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "thanos-sidecar" {
				for _, arg := range container.Args {
					if arg == expectedArg {
						containsArg = true
						break
					}
				}
			}
		}
		if !containsArg {
			t.Fatalf("Thanos sidecar is missing expected argument: %s", expectedArg)
		}
	}

	{
		var containsArg bool
		const expectedArg = "--storage.tsdb.max-block-duration=2h"
		for _, container := range sset.Spec.Template.Spec.Containers {
			if container.Name == "prometheus" {
				for _, arg := range container.Args {
					if arg == expectedArg {
						containsArg = true
						break
					}
				}
			}

		}
		if !containsArg {
			t.Fatalf("Prometheus is missing expected argument: %s", expectedArg)
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
		if !found {
			t.Fatalf("--tsdb.path argument should be given to the Thanos sidecar, got %q", strings.Join(sset.Spec.Template.Spec.Containers[3].Args, " "))
		}
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
		if !found {
			t.Fatal("Prometheus data volume should be mounted in the Thanos sidecar")
		}
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
	if !found {
		t.Fatal("Thanos BlockDuration arg change not found")
	}
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

	if sset.Spec.Template.Spec.Containers[0].Name != "prometheus" {
		t.Fatalf("expected 1st containers to be prometheus, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	if sset.Spec.Template.Spec.Containers[2].Name != "thanos-sidecar" {
		t.Fatalf("expected 3rd containers to be thanos-sidecar, got %s", sset.Spec.Template.Spec.Containers[2].Name)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[2].Env {
		if env.Name == "TRACING_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos sidecar is missing expected TRACING_CONFIG env var with correct value")
	}

	{
		var containsArg bool
		const expectedArg = "--tracing.config=$(TRACING_CONFIG)"
		for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Thanos sidecar is missing expected argument: %s", expectedArg)
		}
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
	if !containsVolume {
		t.Fatalf("Thanos sidecar volume is missing expected volume: %s", testVolume)
	}

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

	if !containsVolumeMount {
		t.Fatal("expected thanos sidecar volume mounts to match")
	}
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

	for i, test := range tests {
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

		if foundRetention != test.shouldContainRetention || foundRetentionFlag != test.shouldContainRetention {
			if test.shouldContainRetention {
				t.Fatalf("test %d, expected Prometheus args to contain %v, but got %v", i, test.expectedRetentionArg, promArgs)
			} else {
				t.Fatalf("test %d, expected Prometheus args to NOT contain %v, but got %v", i, test.expectedRetentionArg, promArgs)
			}
		}

		if foundRetentionSize != test.shouldContainRetentionSize || foundRetentionSizeFlag != test.shouldContainRetentionSize {
			if test.shouldContainRetentionSize {
				t.Fatalf("test %d, expected Prometheus args to contain %v, but got %v", i, test.expectedRetentionSizeArg, promArgs)
			} else {
				t.Fatalf("test %d, expected Prometheus args to NOT contain %v, but got %v", i, test.expectedRetentionSizeArg, promArgs)
			}
		}
	}
}

func TestReplicasConfigurationWithSharding(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig:             defaultTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	replicas := int32(2)
	shards := int32(3)
	logger := newLogger()
	p := monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: &replicas,
				Shards:   &shards,
			},
		},
	}

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		logger,
		"test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		testConfig,
		cg,
		nil,
		"",
		1,
		nil)
	require.NoError(t, err)

	if *sset.Spec.Replicas != int32(2) {
		t.Fatal("Unexpected replicas configuration.")
	}

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
	if !found {
		t.Fatal("Shard.")
	}
}

func TestSidecarResources(t *testing.T) {
	operator.TestSidecarsResources(t, func(reloaderConfig operator.ContainerConfig) *appsv1.StatefulSet {
		testConfig := &operator.Config{
			ReloaderConfig:             reloaderConfig,
			PrometheusDefaultBaseImage: defaultTestConfig.PrometheusDefaultBaseImage,
			ThanosDefaultBaseImage:     defaultTestConfig.ThanosDefaultBaseImage,
		}
		logger := newLogger()
		p := monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{},
		}

		cg, err := prompkg.NewConfigGenerator(logger, &p, false)
		require.NoError(t, err)

		sset, err := makeStatefulSet(
			logger,
			"test",
			&p,
			p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
			p.Spec.Retention,
			p.Spec.RetentionSize,
			p.Spec.Rules,
			p.Spec.Query,
			p.Spec.AllowOverlappingBlocks,
			p.Spec.EnableAdminAPI,
			p.Spec.QueryLogFile,
			p.Spec.Thanos,
			p.Spec.DisableCompaction,
			testConfig,
			cg,
			nil,
			"",
			0,
			nil)
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

	if len(baseSet.Spec.Template.Spec.Containers)+1 != len(addSset.Spec.Template.Spec.Containers) {
		t.Fatalf("container count mismatch")
	}

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

	if len(baseSet.Spec.Template.Spec.Containers) != len(modSset.Spec.Template.Spec.Containers) {
		t.Fatalf("container count mismatch. container %s was added instead of merged", existingContainerName)
	}

	// Check that adding a container with an existing name results in a single patched container.
	for _, c := range modSset.Spec.Template.Spec.Containers {
		if c.Name == existingContainerName && c.Image != containerImage {
			t.Fatalf("expected container %s to have the image %s but got %s", existingContainerName, containerImage, c.Image)
		}
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
		found := false
		for _, flag := range promArgs {
			if flag == test.expectedArg {
				found = true
				break
			}
		}

		if found != test.shouldContain {
			if test.shouldContain {
				t.Fatalf("expected Prometheus args to contain %v, but got %v", test.expectedArg, promArgs)
			} else {
				t.Fatalf("expected Prometheus args to NOT contain %v, but got %v", test.expectedArg, promArgs)
			}
		}
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
		found := false
		for _, flag := range promArgs {
			if flag == expectedArg {
				found = true
				break
			}
		}

		if found != test.shouldContain {
			if test.shouldContain {
				t.Fatalf("expected Prometheus args to contain %v, but got %v", expectedArg, promArgs)
			} else {
				t.Fatalf("expected Prometheus args to NOT contain %v, but got %v", expectedArg, promArgs)
			}
		}
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
				for _, flag := range sset.Spec.Template.Spec.Containers[2].Args {
					if flag == exp {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("Expecting argument %q but not found in %v", exp, sset.Spec.Template.Spec.Containers[2].Args)
				}
			}
		})
	}
}

func TestTerminationPolicy(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{Spec: monitoringv1.PrometheusSpec{}})
	require.NoError(t, err)

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.TerminationMessagePolicy != v1.TerminationMessageFallbackToLogsOnError {
			t.Fatalf("Unexpected TermintationMessagePolicy. Expected %v got %v", v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy)
		}
	}
}

func TestEnableFeaturesWithOneFeature(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnableFeatures: []string{"exemplar-storage"},
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

	if !found {
		t.Fatal("Prometheus enabled feature is not correctly set.")
	}
}

func TestEnableFeaturesWithMultipleFeature(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				EnableFeatures: []string{"exemplar-storage1", "exemplar-storage2"},
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

	if !found {
		t.Fatal("Prometheus enabled features are not correctly set.")
	}
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

	if !found {
		t.Fatal("Prometheus web page title is not correctly set.")
	}
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

	if !found {
		t.Fatal("Prometheus web max connections is not correctly set.")
	}
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
		if res[i] != name {
			t.Fatal("Unexpected StatefulSet shard name")
		}
	}
}

func TestExpectStatefulSetMinReadySeconds(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	})
	require.NoError(t, err)

	// assert defaults to zero if nil
	if sset.Spec.MinReadySeconds != 0 {
		t.Fatalf("expected MinReadySeconds to be zero but got %d", sset.Spec.MinReadySeconds)
	}

	var expect uint32 = 5
	sset, err = makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				MinReadySeconds: &expect,
			},
		},
	})
	require.NoError(t, err)

	if sset.Spec.MinReadySeconds != int32(expect) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", expect, sset.Spec.MinReadySeconds)
	}
}

func TestConfigReloader(t *testing.T) {
	expectedShardNum := 0
	logger := newLogger()
	p := monitoringv1.Prometheus{}

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	require.NoError(t, err)

	sset, err := makeStatefulSet(
		logger,
		"test",
		&p,
		p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
		p.Spec.Retention,
		p.Spec.RetentionSize,
		p.Spec.Rules,
		p.Spec.Query,
		p.Spec.AllowOverlappingBlocks,
		p.Spec.EnableAdminAPI,
		p.Spec.QueryLogFile,
		p.Spec.Thanos,
		p.Spec.DisableCompaction,
		defaultTestConfig,
		cg,
		nil,
		"",
		int32(expectedShardNum),
		nil)
	require.NoError(t, err)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=http://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expectd container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
			for _, env := range c.Env {
				if env.Name == "SHARD" && !reflect.DeepEqual(env.Value, strconv.Itoa(expectedShardNum)) {
					t.Fatalf("expectd shard value is %s, but found %s", strconv.Itoa(expectedShardNum), env.Value)
				}
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
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expectd init container args are %s, but found %s", expectedArgsInitConfigReloader, c.Args)
			}
			for _, env := range c.Env {
				if env.Name == "SHARD" && !reflect.DeepEqual(env.Value, strconv.Itoa(expectedShardNum)) {
					t.Fatalf("expectd shard value is %s, but found %s", strconv.Itoa(expectedShardNum), env.Value)
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

	if !found {
		t.Fatal("Sidecar get_config_interval is not set when it should.")
	}
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

	if !found {
		t.Fatal("Sidecar get_config_timeout is not set when it should.")
	}
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

	if !found {
		t.Fatal("Sidecar ready timeout not set when it should.")
	}
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
		if volume.Name == "query-log-file" {
			found = true
		}
	}

	if !found {
		t.Fatal("Volume for query log file not found.")
	}

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == "query-log-file" {
					found = true
				}
			}
		}
	}

	if !found {
		t.Fatal("Query log file not mounted.")
	}
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
		if volume.Name == "query-log-file" {
			found = true
		}
	}

	if found {
		t.Fatal("Volume for query log file found, when it shouldn't be.")
	}

	found = false
	for _, container := range sset.Spec.Template.Spec.Containers {
		if container.Name == "prometheus" {
			for _, vm := range container.VolumeMounts {
				if vm.Name == "query-log-file" {
					found = true
				}
			}
		}
	}

	if found {
		t.Fatal("Query log file mounted, when it shouldn't be.")
	}
}

func TestEnableRemoteWriteReceiver(t *testing.T) {
	for _, tc := range []struct {
		version                         string
		enableRemoteWriteReceiver       bool
		expectedRemoteWriteReceiverFlag bool
	}{
		// Test lower version where feature not available
		{
			version:                   "2.32.0",
			enableRemoteWriteReceiver: true,
		},
		// Test correct version from which feature available
		{
			version:                         "2.33.0",
			enableRemoteWriteReceiver:       true,
			expectedRemoteWriteReceiverFlag: true,
		},
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
	} {
		t.Run(fmt.Sprintf("case %s", tc.version), func(t *testing.T) {
			sset, err := makeStatefulSetFromPrometheus(monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version:                   tc.version,
						EnableRemoteWriteReceiver: tc.enableRemoteWriteReceiver,
					},
				},
			})
			require.NoError(t, err)

			found := false
			for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
				if flag == "--web.enable-remote-write-receiver" {
					found = true
					break
				}
			}

			if found != tc.expectedRemoteWriteReceiverFlag {
				t.Fatalf("Expecting Prometheus remote write receiver to be %t, got %t", tc.expectedRemoteWriteReceiverFlag, found)
			}
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
			},
		},
	})
	require.NoError(t, err)

	if !reflect.DeepEqual(sset.Spec.Template.Spec.NodeSelector, nodeSelector) {
		t.Fatalf("expected node selector to match, want %v, got %v", nodeSelector, sset.Spec.Template.Spec.NodeSelector)
	}
	if !reflect.DeepEqual(*sset.Spec.Template.Spec.Affinity, affinity) {
		t.Fatalf("expected affinity to match, want %v, got %v", affinity, *sset.Spec.Template.Spec.Affinity)
	}
	if !reflect.DeepEqual(sset.Spec.Template.Spec.Tolerations, tolerations) {
		t.Fatalf("expected tolerations to match, want %v, got %v", tolerations, sset.Spec.Template.Spec.Tolerations)
	}
	if !reflect.DeepEqual(*sset.Spec.Template.Spec.SecurityContext, securityContext) {
		t.Fatalf("expected security context  to match, want %v, got %v", securityContext, *sset.Spec.Template.Spec.SecurityContext)
	}
	if sset.Spec.Template.Spec.PriorityClassName != priorityClassName {
		t.Fatalf("expected priority class name to match, want %s, got %s", priorityClassName, sset.Spec.Template.Spec.PriorityClassName)
	}
	if sset.Spec.Template.Spec.ServiceAccountName != serviceAccountName {
		t.Fatalf("expected service account name to match, want %s, got %s", serviceAccountName, sset.Spec.Template.Spec.ServiceAccountName)
	}
	if len(sset.Spec.Template.Spec.HostAliases) != len(hostAliases) {
		t.Fatalf("expected length of host aliases to match, want %d, got %d", len(hostAliases), len(sset.Spec.Template.Spec.HostAliases))
	}
	for _, initContainer := range sset.Spec.Template.Spec.InitContainers {
		if !reflect.DeepEqual(initContainer.ImagePullPolicy, imagePullPolicy) {
			t.Fatalf("expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, sset.Spec.Template.Spec.Containers[0].ImagePullPolicy)
		}
	}
	for _, container := range sset.Spec.Template.Spec.Containers {
		if !reflect.DeepEqual(container.ImagePullPolicy, imagePullPolicy) {
			t.Fatalf("expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, sset.Spec.Template.Spec.Containers[0].ImagePullPolicy)
		}
	}
	if !reflect.DeepEqual(sset.Spec.Template.Spec.ImagePullSecrets, imagePullSecrets) {
		t.Fatalf("expected image pull secrets to match, want %s, got %s", imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets)
	}
	if sset.Spec.Template.Spec.HostNetwork != hostNetwork {
		t.Fatalf("expected hostNetwork configuration to match but failed")
	}
}

func TestPrometheusAdditionalArgsNoError(t *testing.T) {
	expectedPrometheusArgs := []string{
		"--web.console.templates=/etc/prometheus/consoles",
		"--web.console.libraries=/etc/prometheus/console_libraries",
		"--config.file=/etc/prometheus/config_out/prometheus.env.yaml",
		"--web.enable-lifecycle",
		"--web.route-prefix=/",
		"--storage.tsdb.retention.time=24h",
		"--storage.tsdb.path=/prometheus",
		"--web.config.file=/etc/prometheus/web_config/web-config.yaml",
		"--scrape.discovery-reload-interval=30s",
		"--storage.tsdb.no-lockfile",
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
			},
		},
	})
	require.NoError(t, err)

	ssetContainerArgs := sset.Spec.Template.Spec.Containers[0].Args
	if !reflect.DeepEqual(ssetContainerArgs, expectedPrometheusArgs) {
		t.Fatalf("expected Prometheus container args to match, want %s, got %s", expectedPrometheusArgs, ssetContainerArgs)
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
	require.NotNil(t, err)

	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("expected the following text to be present in the error msg: %s", expectedErrorMsg)
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
	require.NotNil(t, err)

	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("expected the following text to be present in the error msg: %s", expectedErrorMsg)
	}
}

func TestPrometheusAdditionalNoPrefixArgsDuplicate(t *testing.T) {
	expectedErrorMsg := "can't set arguments which are already managed by the operator: no-storage.tsdb.wal-compression"
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
	require.NotNil(t, err)

	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("expected the following text to be present in the error msg: %s", expectedErrorMsg)
	}
}

func TestThanosAdditionalArgsNoError(t *testing.T) {
	expectedThanosArgs := []string{
		"sidecar",
		"--prometheus.url=http://localhost:9090/",
		`--prometheus.http-client={"tls_config": {"insecure_skip_verify":true}}`,
		"--grpc-address=:10901",
		"--http-address=:10902",
		"--log.level=info",
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
	if !reflect.DeepEqual(ssetContainerArgs, expectedThanosArgs) {
		t.Fatalf("expected Thanos container args to match, want %s, got %s", expectedThanosArgs, ssetContainerArgs)
	}
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
	require.NotNil(t, err)

	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("expected the following text to be present in the error msg: %s", expectedErrorMsg)
	}
}

func TestPrometheusQuerySpec(t *testing.T) {
	durationPtr := func(s string) *monitoringv1.Duration { d := monitoringv1.Duration(s); return &d }

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
			lookbackDelta:  pointer.String("2m"),
			maxConcurrency: pointer.Int32(10),
			maxSamples:     pointer.Int32(10000),
			timeout:        durationPtr("1m"),

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.max-concurrency=10",
				"--query.max-samples=10000",
				"--query.timeout=1m",
			},
		},
		{
			name:           "zero values are skipped",
			lookbackDelta:  pointer.String("2m"),
			maxConcurrency: pointer.Int32(0),
			maxSamples:     pointer.Int32(0),
			timeout:        durationPtr("1m"),

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.timeout=1m",
			},
		},
		{
			name:           "max samples skipped if version < 2.5",
			lookbackDelta:  pointer.String("2m"),
			maxConcurrency: pointer.Int32(10),
			maxSamples:     pointer.Int32(10000),
			timeout:        durationPtr("1m"),
			version:        "v2.4.0",

			expected: []string{
				"--query.lookback-delta=2m",
				"--query.max-concurrency=10",
				"--query.timeout=1m",
			},
		},
		{
			name:           "max samples not skipped if version > 2.5",
			lookbackDelta:  pointer.String("2m"),
			maxConcurrency: pointer.Int32(10),
			maxSamples:     pointer.Int32(10000),
			timeout:        durationPtr("1m"),
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
					if containerArg != "" {
						t.Fatalf("found %q while not expected", containerArg)
					}
					continue
				}

				if containerArg != expected {
					t.Fatalf("expected %q to be found but got %q", expected, containerArg)
				}
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
					ObjectStorageConfigFile: func(s string) *string { return &s }("/etc/thanos.cfg"),
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
			if len(sset.Spec.Template.Spec.Containers) != exp {
				t.Fatalf("Expecting %d containers, got %d", exp, len(sset.Spec.Template.Spec.Containers))
			}

			for _, c := range sset.Spec.Template.Spec.Containers {
				if len(c.SecurityContext.Capabilities.Add) != 0 {
					t.Fatalf("Expecting 0 added capabilities, got %d", len(c.SecurityContext.Capabilities.Add))
				}

				if len(c.SecurityContext.Capabilities.Drop) != 1 {
					t.Fatalf("Expecting 1 dropped capabilities, got %d", len(c.SecurityContext.Capabilities.Drop))
				}

				if string(c.SecurityContext.Capabilities.Drop[0]) != "ALL" {
					t.Fatalf("Expecting ALL dropped capability, got %s", c.SecurityContext.Capabilities.Drop[0])
				}
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

	if sset.Spec.Template.Spec.HostNetwork != hostNetwork {
		t.Fatalf("expected hostNetwork configuration to match but failed")
	}

	if sset.Spec.Template.Spec.DNSPolicy != v1.DNSClusterFirstWithHostNet {
		t.Fatalf("expected DNSPolicy configuration to match due to hostNetwork but failed")
	}
}
