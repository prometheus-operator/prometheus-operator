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
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kylelemons/godebug/pretty"
)

var (
	defaultTestConfig = &Config{
		ConfigReloaderImage:           "quay.io/coreos/configmap-reload:latest",
		ConfigReloaderCPU:             "100m",
		ConfigReloaderMemory:          "25Mi",
		PrometheusConfigReloaderImage: "quay.io/coreos/prometheus-config-reloader:latest",
		PrometheusDefaultBaseImage:    "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:        "quay.io/thanos/thanos",
	}
)

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, defaultTestConfig, nil, "")

	require.NoError(t, err)

	if !reflect.DeepEqual(labels, sset.Labels) {
		fmt.Println(pretty.Compare(labels, sset.Labels))
		t.Fatal("Labels are not properly being propagated to the StatefulSet")
	}

	expectedAnnotations := map[string]string{
		"prometheus-operator-input-hash": "",
	}
	for k, v := range annotations {
		expectedAnnotations[k] = v
	}

	if !reflect.DeepEqual(expectedAnnotations, sset.Annotations) {
		fmt.Println(pretty.Compare(expectedAnnotations, sset.Annotations))
		t.Fatal("Annotations are not properly being propagated to the StatefulSet")
	}
}

func TestPodLabelsAnnotations(t *testing.T) {
	annotations := map[string]string{
		"testannotation": "testvalue",
	}
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.PrometheusSpec{
			PodMetadata: &metav1.ObjectMeta{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, nil, "")
	require.NoError(t, err)
	if _, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]; !ok {
		t.Fatal("Pod labes are not properly propagated")
	}
	if !reflect.DeepEqual(annotations, sset.Spec.Template.ObjectMeta.Annotations) {
		t.Fatal("Pod annotaitons are not properly propagated")
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

	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			StorageClassName: &storageClass,
		},
	}

	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: pvc,
			},
		},
	}, defaultTestConfig, nil, "")

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

	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PrometheusSpec{
			Storage: &monitoringv1.StorageSpec{
				EmptyDir: &emptyDir,
			},
		},
	}, defaultTestConfig, nil, "")

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

	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "volume-init-test",
		},
		Spec: monitoringv1.PrometheusSpec{
			Secrets: []string{
				"test-secret1",
			},
		},
	}, defaultTestConfig, []string{"rules-configmap-one"}, "")

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

func TestMemoryRequestNotAdjustedWhenLimitLarger2Gi(t *testing.T) {
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Version: "v1.8.2",
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("3Gi"),
				},
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	resourceRequest := sset.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceMemory]
	requestString := resourceRequest.String()
	resourceLimit := sset.Spec.Template.Spec.Containers[0].Resources.Limits[v1.ResourceMemory]
	limitString := resourceLimit.String()
	if requestString != "2Gi" {
		t.Fatalf("Resource request is expected to be 1Gi, instead found %s", requestString)
	}
	if limitString != "3Gi" {
		t.Fatalf("Resource limit is expected to be 1Gi, instead found %s", limitString)
	}
}

func TestAdditionalConfigMap(t *testing.T) {
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			ConfigMaps: []string{"test-cm1"},
		},
	}, defaultTestConfig, nil, "")
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

func TestMemoryRequestAdjustedWhenOnlyLimitGiven(t *testing.T) {
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Version: "v1.8.2",
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	resourceRequest := sset.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceMemory]
	requestString := resourceRequest.String()
	resourceLimit := sset.Spec.Template.Spec.Containers[0].Resources.Limits[v1.ResourceMemory]
	limitString := resourceLimit.String()
	if requestString != "1Gi" {
		t.Fatalf("Resource request is expected to be 1Gi, instead found %s", requestString)
	}
	if limitString != "1Gi" {
		t.Fatalf("Resource limit is expected to be 1Gi, instead found %s", limitString)
	}
}

func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			ListenLocal: true,
		},
	}, defaultTestConfig, nil, "")
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
					`if [ -x "$(command -v curl)" ]; then curl http://localhost:9090/-/ready; elif [ -x "$(command -v wget)" ]; then wget -q http://localhost:9090/-/ready; else exit 1; fi`,
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

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{
					`sh`,
					`-c`,
					`if [ -x "$(command -v curl)" ]; then curl http://localhost:9090/-/healthy; elif [ -x "$(command -v wget)" ]; then wget -q http://localhost:9090/-/healthy; else exit 1; fi`,
				},
			},
		},
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %v\n\nGot: %v", expectedLivenessProbe, actualLivenessProbe)
	}

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 0 {
		t.Fatal("Prometheus container should have 0 ports defined")
	}
}

func TestTagAndShaAndVersion(t *testing.T) {
	{
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
			},
		}, defaultTestConfig, nil, "")
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
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
			},
		}, defaultTestConfig, nil, "")
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/prometheus@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
	{
		image := "my-reg/prometheus:latest"
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v2.3.2",
				Image:   &image,
			},
		}, defaultTestConfig, nil, "")
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "my-reg/prometheus:latest"
		if resultImage != expected {
			t.Fatalf("Explicit image should have precedence. Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
}

func TestThanosTagAndShaAndVersion(t *testing.T) {
	{
		thanosTag := "my-unrelated-tag"
		thanosVersion := "v0.1.0"
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		}, defaultTestConfig, nil, "")
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
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
				},
			},
		}, defaultTestConfig, nil, "")
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
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Thanos: &monitoringv1.ThanosSpec{
					SHA:     &thanosSHA,
					Version: &thanosVersion,
					Tag:     &thanosTag,
					Image:   &thanosImage,
				},
			},
		}, defaultTestConfig, nil, "")
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
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{},
		},
	}, defaultTestConfig, nil, "")
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
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				Resources: expected,
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	actual := sset.Spec.Template.Spec.Containers[2].Resources
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected resources defined. \n\nExpected: %v\n\nGot: %v", expected, actual)
	}
}

func TestThanosObjectStorage(t *testing.T) {
	testKey := "thanos-config-secret-test"

	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ObjectStorageConfig: &v1.SecretKeySelector{
					Key: testKey,
				},
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[2].Env {
		if env.Name == "OBJSTORE_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos sidecar is missing expected OBJSTORE_CONFIG env var with correct value")
	}

	var containsArg bool
	const expectedArg = "--objstore.config=$(OBJSTORE_CONFIG)"
	for _, arg := range sset.Spec.Template.Spec.Containers[2].Args {
		if arg == expectedArg {
			containsArg = true
		}
	}
	if !containsArg {
		t.Fatalf("Thanos sidecar is missing expected argument: %s", expectedArg)
	}
}

func TestRetentionSize(t *testing.T) {
	tests := []struct {
		version              string
		specRetentionSize    string
		expectedRetentionArg string
		shouldContain        bool
	}{
		{"v1.8.2", "2M", "--storage.tsdb.retention.size=2M", false},
		{"v1.8.2", "1Gi", "--storage.tsdb.retention.size=1Gi", false},
		{"v2.5.0", "2M", "--storage.tsdb.retention.size=2M", false},
		{"v2.5.0", "1Gi", "--storage.tsdb.retention.size=1Gi", false},
		{"v2.7.0", "2M", "--storage.tsdb.retention.size=2M", true},
		{"v2.7.0", "1Gi", "--storage.tsdb.retention.size=1Gi", true},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:       test.version,
				RetentionSize: test.specRetentionSize,
			},
		}, defaultTestConfig, nil, "")
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
		{"v1.8.2", "", "-storage.local.retention=24h"},
		{"v1.8.2", "1d", "-storage.local.retention=1d"},
		{"v2.5.0", "", "--storage.tsdb.retention=24h"},
		{"v2.5.0", "1d", "--storage.tsdb.retention=1d"},
		{"v2.7.0", "", "--storage.tsdb.retention.time=24h"},
		{"v2.7.0", "1d", "--storage.tsdb.retention.time=1d"},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:   test.version,
				Retention: test.specRetention,
			},
		}, defaultTestConfig, nil, "")
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

func TestSidecarsNoCPULimits(t *testing.T) {
	testConfig := &Config{
		ConfigReloaderImage:           "quay.io/coreos/configmap-reload:latest",
		ConfigReloaderCPU:             "0",
		ConfigReloaderMemory:          "50Mi",
		PrometheusConfigReloaderImage: "quay.io/coreos/prometheus-config-reloader:latest",
		PrometheusDefaultBaseImage:    "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:        "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{Limits: v1.ResourceList{
		v1.ResourceMemory: resource.MustParse("50Mi"),
	}}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatal("Unexpected resource requests/limits set, when none should be set.")
		}
	}
}

func TestSidecarsNoMemoryLimits(t *testing.T) {
	testConfig := &Config{
		ConfigReloaderImage:           "quay.io/coreos/configmap-reload:latest",
		ConfigReloaderCPU:             "100m",
		ConfigReloaderMemory:          "0",
		PrometheusConfigReloaderImage: "quay.io/coreos/prometheus-config-reloader:latest",
		PrometheusDefaultBaseImage:    "quay.io/prometheus/prometheus",
		ThanosDefaultBaseImage:        "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{},
	}, testConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedResources := v1.ResourceRequirements{Limits: v1.ResourceList{
		v1.ResourceCPU: resource.MustParse("100m"),
	}}
	for _, c := range sset.Spec.Template.Spec.Containers {
		if (c.Name == "prometheus-config-reloader" || c.Name == "rules-configmap-reloader") && !reflect.DeepEqual(c.Resources, expectedResources) {
			t.Fatal("Unexpected resource requests/limits set, when none should be set.")
		}
	}
}

func TestAdditionalContainers(t *testing.T) {
	// The base to compare everything against
	baseSet, err := makeStatefulSet(monitoringv1.Prometheus{}, defaultTestConfig, nil, "")

	// Add an extra container
	addSset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Containers: []v1.Container{
				{
					Name: "extra-container",
				},
			},
		},
	}, defaultTestConfig, nil, "")
	require.NoError(t, err)

	if len(baseSet.Spec.Template.Spec.Containers)+1 != len(addSset.Spec.Template.Spec.Containers) {
		t.Fatalf("container count mismatch")
	}

	// Adding a new container with the same name results in a merge and just one container
	const existingContainerName = "prometheus"
	const containerImage = "madeUpContainerImage"
	modSset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Containers: []v1.Container{
				{
					Name:  existingContainerName,
					Image: containerImage,
				},
			},
		},
	}, defaultTestConfig, nil, "")
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
		{"v1.8.2", nil, "-no-storage.tsdb.wal-compression", false},
		{"v1.8.2", nil, "-storage.tsdb.wal-compression", false},
		{"v1.8.2", &fa, "-no-storage.tsdb.wal-compression", false},
		{"v1.8.2", &tr, "-storage.tsdb.wal-compression", false},
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
		sset, err := makeStatefulSet(monitoringv1.Prometheus{
			Spec: monitoringv1.PrometheusSpec{
				Version:        test.version,
				WALCompression: test.enabled,
			},
		}, defaultTestConfig, nil, "")
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
	sset, err := makeStatefulSet(monitoringv1.Prometheus{
		Spec: monitoringv1.PrometheusSpec{
			Thanos: &monitoringv1.ThanosSpec{
				ListenLocal: true,
			},
		},
	}, defaultTestConfig, nil, "")
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
