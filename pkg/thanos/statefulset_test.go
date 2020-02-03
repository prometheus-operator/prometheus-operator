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
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultTestConfig = Config{
		ConfigReloaderImage:    "jimmidyson/configmap-reload:latest",
		ConfigReloaderCPU:      "100m",
		ConfigReloaderMemory:   "25Mi",
		ThanosDefaultBaseImage: "quay.io/thanos/thanos",
	}
)

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: []string{""},
		},
	}, nil, defaultTestConfig, nil)

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
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: []string{""},
			PodMetadata: &metav1.ObjectMeta{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, nil, defaultTestConfig, nil)
	require.NoError(t, err)
	if _, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]; !ok {
		t.Fatal("Pod labes are not properly propagated")
	}
	if !reflect.DeepEqual(annotations, sset.Spec.Template.ObjectMeta.Annotations) {
		t.Fatal("Pod annotaitons are not properly propagated")
	}
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
							},
						},
					},
					Volumes: []v1.Volume{
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
			QueryEndpoints: []string{""},
		},
	}, nil, defaultTestConfig, []string{"rules-configmap-one"})
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
