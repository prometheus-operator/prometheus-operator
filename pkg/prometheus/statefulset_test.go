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
	"reflect"
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultTestConfig = &Config{
		ConfigReloaderImage: "quay.io/coreos/configmap-reload:latest",
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

	require.NoError(t, err)

	if !reflect.DeepEqual(labels, sset.Labels) || !reflect.DeepEqual(annotations, sset.Annotations) {
		t.Fatal("Labels or Annotations are not properly being propagated to the StatefulSet")
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

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
								}, {
									Name:      "prometheus--db",
									ReadOnly:  false,
									MountPath: "/prometheus",
									SubPath:   "",
								}, {
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
									SecretName: configSecretName(""),
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
							Name: "secret-test-secret1",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "test-secret1",
								},
							},
						},
						{
							Name: "prometheus--db",
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
		Spec: monitoringv1.PrometheusSpec{
			Secrets: []string{
				"test-secret1",
			},
		},
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

	require.NoError(t, err)

	if !reflect.DeepEqual(expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes) {
		t.Fatalf("Unexpected volumes: want %v, got %v",
			expected.Spec.Template.Spec.Volumes,
			sset.Spec.Template.Spec.Volumes)
	}
	if !reflect.DeepEqual(expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts) {
		t.Fatalf("Unexpected volume mounts: want %v, got %v",
			expected.Spec.Template.Spec.Containers[0].VolumeMounts,
			sset.Spec.Template.Spec.Containers[0].VolumeMounts)
	}
}

func TestDeterministicRuleFileHashing(t *testing.T) {
	cmr, err := makeRuleConfigMap(makeConfigMap())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		testcmr, err := makeRuleConfigMap(makeConfigMap())
		if err != nil {
			t.Fatal(err)
		}

		if cmr.Checksum != testcmr.Checksum {
			t.Fatalf("Non-deterministic rule file hash generation")
		}
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})
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
	}, nil, defaultTestConfig, []*v1.ConfigMap{})
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

	if sset.Spec.Template.Spec.Containers[0].ReadinessProbe != nil {
		t.Fatal("Prometheus readiness probe expected to be empty")
	}

	if sset.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
		t.Fatal("Prometheus readiness probe expected to be empty")
	}

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 0 {
		t.Fatal("Prometheus container should have 0 ports defined")
	}
}

func makeConfigMap() *v1.ConfigMap {
	res := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testcm",
			Namespace: "default",
		},
		Data: map[string]string{},
	}

	res.Data["test1"] = "value 1"
	res.Data["test2"] = "value 2"
	res.Data["test3"] = "value 3"

	return res
}
