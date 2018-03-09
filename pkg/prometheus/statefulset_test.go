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
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"reflect"
	"testing"
	"path"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
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

func TestPrometheusArgs(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}
	promArgs := []string{"-config.file=/etc/prometheus/custom/prometheus.yaml"}

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1alpha1.PrometheusSpec{
			PrometheusArgs: promArgs,
		},
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

	require.NoError(t, err)
    args := sset.Spec.Template.Spec.Containers[0].Args

    // We expect defaults and overridden values to be present
    expected := promArgs
    expected = append(expected, "-storage.local.path=/var/prometheus/data")
Outer:
    for _,expect := range expected {
		for _,a := range args {
			if a == expect {
				continue Outer
			}
		}
		t.Fatalf("PrometheusArgs are not properly being propagated to the StatefulSet. Failed to find '%s' in %v", expect, args)
	}
}

func TestPrometheusProbes(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	one := int32(1)
	two := int32(2)
	three := int32(3)
	four := int32(4)
	five := int32(5)

	livenessProbe := &v1alpha1.Probe{
		InitialDelaySeconds: &one,
		TimeoutSeconds: &two,
		PeriodSeconds: &three,
		SuccessThreshold: &four,
		FailureThreshold: &five,
	}

	readinessProbe := &v1alpha1.Probe{
		InitialDelaySeconds: &five,
		TimeoutSeconds: &four,
		PeriodSeconds: &three,
		SuccessThreshold: &two,
		FailureThreshold: &one,
	}

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1alpha1.PrometheusSpec{
			LivenessProbe: livenessProbe,
			ReadinessProbe: readinessProbe,
		},
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

	require.NoError(t, err)

	probeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: path.Clean("//status"),
			Port: intstr.FromString("web"),
		},
	}

	expectedLivenessProbe := &v1.Probe{
		Handler: probeHandler,
		InitialDelaySeconds: *livenessProbe.InitialDelaySeconds,
		TimeoutSeconds: *livenessProbe.TimeoutSeconds,
		PeriodSeconds: *livenessProbe.PeriodSeconds,
		SuccessThreshold: *livenessProbe.SuccessThreshold,
		FailureThreshold: *livenessProbe.FailureThreshold,
	}

	expectedReadinessProbe := &v1.Probe{
		Handler: probeHandler,
		InitialDelaySeconds: *readinessProbe.InitialDelaySeconds,
		TimeoutSeconds: *readinessProbe.TimeoutSeconds,
		PeriodSeconds: *readinessProbe.PeriodSeconds,
		SuccessThreshold: *readinessProbe.SuccessThreshold,
		FailureThreshold: *readinessProbe.FailureThreshold,
	}

	if !reflect.DeepEqual(expectedLivenessProbe, sset.Spec.Template.Spec.Containers[0].LivenessProbe) || !reflect.DeepEqual(expectedReadinessProbe, sset.Spec.Template.Spec.Containers[0].ReadinessProbe) {
		t.Fatal("Probes are not properly being propagated to the StatefulSet")
	}
}

func TestOverrideProbeDefaults(t *testing.T) {
	one := int32(1)
	two := int32(2)
	three := int32(3)

	src := &v1alpha1.Probe{
		InitialDelaySeconds: &one,
		PeriodSeconds: &two,
		FailureThreshold: &three,
	}

	dst := &v1.Probe{
		InitialDelaySeconds: *src.InitialDelaySeconds,
		TimeoutSeconds: 10,
		PeriodSeconds: *src.PeriodSeconds,
		SuccessThreshold: 20,
		FailureThreshold: *src.FailureThreshold,
	}

	overrideProbeDefaults(src, dst)

	if !(dst.TimeoutSeconds == 10 || dst.SuccessThreshold == 20) {
		t.Fatal("Defaults are not respected by probe overrides")
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

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1alpha1.PrometheusSpec{
			Storage: &v1alpha1.StorageSpec{
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

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
		Spec: v1alpha1.PrometheusSpec{
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

	sset, err := makeStatefulSet(v1alpha1.Prometheus{
		Spec: v1alpha1.PrometheusSpec{
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
