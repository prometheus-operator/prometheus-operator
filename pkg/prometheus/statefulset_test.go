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
	"reflect"
	"strings"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kylelemons/godebug/pretty"
)

var (
	defaultTestConfig = &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos",
	}
)

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
	}

	expectedPodLabels := map[string]string{
		"prometheus":                   "",
		"app":                          "prometheus",
		"app.kubernetes.io/name":       "prometheus",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultPrometheusVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "",
		"operator.prometheus.io/name":  "",
		"operator.prometheus.io/shard": "0",
	}

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, defaultTestConfig, nil, "", 0)

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
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, nil, "", 0)
	require.NoError(t, err)
	if _, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]; !ok {
		t.Fatal("Pod labes are not properly propagated")
	}
	if !reflect.DeepEqual(annotations, sset.Spec.Template.ObjectMeta.Annotations) {
		t.Fatal("Pod annotaitons are not properly propagated")
	}
}
func TestPodLabelsShouldNotBeSelectorLabels(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Labels: labels,
			},
		},
	}, defaultTestConfig, nil, "", 0)

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

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: pvc,
			},
		},
	}, defaultTestConfig, nil, "", 0)

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

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Storage: &monitoringv1.StorageSpec{
				EmptyDir: &emptyDir,
			},
		},
	}, defaultTestConfig, nil, "", 0)

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir != nil && !reflect.DeepEqual(emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium) {
		t.Fatal("Error adding EmptyDir Spec to StatefulSetSpec")
	}
}

func TestStatefulSetVolumeInitial(t *testing.T) {
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
									Name:      "rules-configmap-one",
									ReadOnly:  false,
									MountPath: "/etc/prometheus/rules/rules-configmap-one",
									SubPath:   "",
								},
								{
									Name:      "secret-test-secret1",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/secrets/test-secret1",
									SubPath:   "",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: configSecretName("volume-init-test"),
								},
							},
						},
						{
							Name: "tls-assets",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: tlsAssetsSecretName("volume-init-test"),
								},
							},
						},
						{
							Name: "config-out",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
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
							Name: "secret-test-secret1",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "test-secret1",
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

	sset, err := makeStatefulSet("volume-init-test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "volume-init-test",
		},
		Spec: monitoringv1.PrometheusSpec{
			Secrets: []string{
				"test-secret1",
			},
		},
	}, defaultTestConfig, []string{"rules-configmap-one"}, "", 0)

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
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			ConfigMaps: []string{"test-cm1"},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	cmVolumeFound := false
	for _, v := range sset.Spec.Template.Spec.Volumes {
		if v.Name == "configmap-test-cm1" {
			cmVolumeFound = true
		}
	}
	if !cmVolumeFound {
		t.Fatal("ConfigMap volume not found")
	}

	cmMounted := false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if v.Name == "configmap-test-cm1" && v.MountPath == "/etc/prometheus/configmaps/test-cm1" {
			cmMounted = true
		}
	}
	if !cmMounted {
		t.Fatal("ConfigMap volume not mounted")
	}
}

func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			ListenLocal: true,
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.listen-address=127.0.0.1:9090" {
			found = true
		}
	}

	if !found {
		t.Fatal("Prometheus not listening on loopback when it should.")
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{
					`sh`,
					`-c`,
					`if [ -x "$(command -v curl)" ]; then exec curl http://localhost:9090/-/ready; elif [ -x "$(command -v wget)" ]; then exec wget -q -O /dev/null http://localhost:9090/-/ready; else exit 1; fi`,
				},
			},
		},
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 120,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 0 {
		t.Fatal("Prometheus container should have 0 ports defined")
	}
}

func TestTagAndShaAndVersion(t *testing.T) {
	{
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus:my-unrelated-tag"
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
	{
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
	// For tests which set monitoringv1.PrometheusSpec.Image, the result will be Image only. SHA, Tag, Version are not considered.
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
				Image:   &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus:latest"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
				Image:   &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Explicit image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version: "v2.3.2",
				Image:   &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Version: "v2.3.2",
				Image:   &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Image: &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:   "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Image: &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := image
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := ""
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Tag:   "my-unrelated-tag",
				Image: &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus:my-unrelated-tag"
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
	{
		image := "my-reg/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb325"
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:   "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:   "my-unrelated-tag",
				Image: &image,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "my-reg/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb325"
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
}

func TestPrometheusDefaultBaseImageFlag(t *testing.T) {
	prometheusBaseImageConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "nondefaultuseflag/quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "nondefaultuseflag/quay.io/thanos/thanos",
	}
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, prometheusBaseImageConfig, nil, "", 0)

	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/prometheus/prometheus"
	if image != expected {
		t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
}

func TestThanosDefaultBaseImageFlag(t *testing.T) {
	thanosBaseImageConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "nondefaultuseflag/quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "nondefaultuseflag/quay.io/thanos/thanos",
	}
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}, thanosBaseImageConfig, nil, "", 0)

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
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

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
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

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
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
					Image:   &thanosImage,
				},
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[2].Image
		expected := "my-registry/thanos:latest"
		if image != expected {
			t.Fatalf("Explicit Thanos image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
}

func TestThanosResourcesNotSet(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				Resources: expected,
			},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	actual := sset.Spec.Template.Spec.Containers[2].Resources
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected resources defined. \n\nExpected: %v\n\nGot: %v", expected, actual)
	}
}

func TestThanosNoObjectStorage(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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

	for _, vol := range sset.Spec.Template.Spec.Containers[2].VolumeMounts {
		if vol.MountPath == storageDir {
			t.Fatal("Prometheus data volume should not be mounted in the Thanos sidecar")
		}
	}
}

func TestThanosObjectStorage(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ObjectStorageConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
			if vol.MountPath == storageDir {
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
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ObjectStorageConfigFile: &testPath,
			},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
					if vol.MountPath == storageDir {
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

func TestThanosTracing(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				TracingConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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

func TestRetentionSize(t *testing.T) {
	tests := []struct {
		version              string
		specRetentionSize    string
		expectedRetentionArg string
		shouldContain        bool
	}{
		{"v2.5.0", "2M", "--storage.tsdb.retention.size=2M", false},
		{"v2.5.0", "1Gi", "--storage.tsdb.retention.size=1Gi", false},
		{"v2.7.0", "2M", "--storage.tsdb.retention.size=2M", true},
		{"v2.7.0", "1Gi", "--storage.tsdb.retention.size=1Gi", true},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:       test.version,
				RetentionSize: test.specRetentionSize,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatal(err)
		}

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		found := false
		for _, flag := range promArgs {
			if flag == test.expectedRetentionArg {
				found = true
				break
			}
		}

		if found != test.shouldContain {
			if test.shouldContain {
				t.Fatalf("expected Prometheus args to contain %v, but got %v", test.expectedRetentionArg, promArgs)
			} else {
				t.Fatalf("expected Prometheus args to NOT contain %v, but got %v", test.expectedRetentionArg, promArgs)
			}
		}
	}
}

func TestRetention(t *testing.T) {
	tests := []struct {
		version              string
		specRetention        string
		expectedRetentionArg string
	}{
		{"v2.5.0", "", "--storage.tsdb.retention=24h"},
		{"v2.5.0", "1d", "--storage.tsdb.retention=1d"},
		{"v2.7.0", "", "--storage.tsdb.retention.time=24h"},
		{"v2.7.0", "1d", "--storage.tsdb.retention.time=1d"},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:   test.version,
				Retention: test.specRetention,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatal(err)
		}

		promArgs := sset.Spec.Template.Spec.Containers[0].Args
		found := false
		for _, flag := range promArgs {
			if flag == test.expectedRetentionArg {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected Prometheus args to contain %v, but got %v", test.expectedRetentionArg, promArgs)
		}
	}
}

func TestReplicasConfigurationWithSharding(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	replicas := int32(2)
	shards := int32(3)
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Replicas: &replicas,
			Shards:   &shards,
		},
	}, testConfig, nil, "", 1)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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

func TestSidecarsNoResources(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "0",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoRequests(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Requests: v1.ResourceList{},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoLimits(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoCPUResources(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoCPURequests(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoCPULimits(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoMemoryResources(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("100m"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("100m"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoMemoryRequests(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("100m"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestSidecarsNoMemoryLimits(t *testing.T) {
	testConfig := &operator.Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		PrometheusDefaultBaseImage: "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:     "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("100m"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", expectedResources.String(), c.Resources.String())
		}
	}
}

func TestAdditionalContainers(t *testing.T) {
	// The base to compare everything against
	baseSet, err := makeStatefulSet("test", monitoringv1.Prometheus{}, defaultTestConfig, nil, "", 0)
	require.NoError(t, err)

	// Add an extra container
	addSset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Containers: []v1.Container{
				{
					Name: "extra-container",
				},
			},
		},
	}, defaultTestConfig, nil, "", 0)
	require.NoError(t, err)

	if len(baseSet.Spec.Template.Spec.Containers)+1 != len(addSset.Spec.Template.Spec.Containers) {
		t.Fatalf("container count mismatch")
	}

	// Adding a new container with the same name results in a merge and just one container
	const existingContainerName = "prometheus"
	const containerImage = "madeUpContainerImage"
	modSset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Containers: []v1.Container{
				{
					Name:  existingContainerName,
					Image: containerImage,
				},
			},
		},
	}, defaultTestConfig, nil, "", 0)
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
		sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:        test.version,
				WALCompression: test.enabled,
			},
		}, defaultTestConfig, nil, "", 0)
		if err != nil {
			t.Fatal(err)
		}

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

func TestThanosListenLocal(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ListenLocal: true,
			},
		},
	}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}
	foundGrpcFlag := false
	foundHTTPFlag := false
	for _, flag := range sset.Spec.Template.Spec.Containers[2].Args {
		if flag == "--grpc-address=127.0.0.1:10901" {
			foundGrpcFlag = true
		}
		if flag == "--http-address=127.0.0.1:10902" {
			foundHTTPFlag = true
		}
	}

	if !foundGrpcFlag || !foundHTTPFlag {
		t.Fatal("Thanos not listening on loopback when it should.")
	}
}

func TestTerminationPolicy(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{Spec: monitoringv1.PrometheusSpec{}}, defaultTestConfig, nil, "", 0)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.TerminationMessagePolicy != v1.TerminationMessageFallbackToLogsOnError {
			t.Fatalf("Unexpected TermintationMessagePolicy. Expected %v got %v", v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy)
		}
	}
}

func TestEnableFeaturesWithOneFeature(t *testing.T) {
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			EnableFeatures: []string{"exemplar-storage"},
		},
	}, defaultTestConfig, nil, "", 0)

	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			EnableFeatures: []string{"exemplar-storage1", "exemplar-storage2"},
		},
	}, defaultTestConfig, nil, "", 0)

	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
	var pageTitle string = "my-page-title"
	sset, err := makeStatefulSet("test", monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Web: &monitoringv1.WebSpec{
				PageTitle: &pageTitle,
			},
		},
	}, defaultTestConfig, nil, "", 0)

	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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

func TestExpectedStatefulSetShardNames(t *testing.T) {
	replicas := int32(2)
	shards := int32(3)
	res := expectedStatefulSetShardNames(&monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: monitoringv1.PrometheusSpec{
			Shards:   &shards,
			Replicas: &replicas,
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
