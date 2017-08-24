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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
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

func TestStatefulSetVolumeInitial(t *testing.T) {
	expected := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/config",
									SubPath:   "",
								}, {
									Name:      "rules",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/rules",
									SubPath:   "",
								}, {
									Name:      "prometheus--db",
									ReadOnly:  false,
									MountPath: "/var/prometheus/data",
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
							Name: "rules",
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

	if !reflect.DeepEqual(expected.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes) || !reflect.DeepEqual(expected.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts) {
		t.Fatal("Volumes mounted in a Pod are not created correctly initially.")
	}
}

func TestStatefulSetVolumeSkip(t *testing.T) {
	old := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/config",
									SubPath:   "",
								}, {
									Name:      "rules",
									ReadOnly:  true,
									MountPath: "/etc/prometheus/rules",
									SubPath:   "",
								}, {
									Name:      "prometheus--db",
									ReadOnly:  false,
									MountPath: "/var/prometheus/data",
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
							Name: "rules",
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
				"test-secret2",
			},
		},
	}, old, defaultTestConfig, []*v1.ConfigMap{})

	require.NoError(t, err)

	if !reflect.DeepEqual(old.Spec.Template.Spec.Volumes, sset.Spec.Template.Spec.Volumes) || !reflect.DeepEqual(old.Spec.Template.Spec.Containers[0].VolumeMounts, sset.Spec.Template.Spec.Containers[0].VolumeMounts) {
		t.Fatal("Volumes mounted in a Pod should not be reconciled.")
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
