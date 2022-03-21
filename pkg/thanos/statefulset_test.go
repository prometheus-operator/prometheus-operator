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
	"sort"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/kylelemons/godebug/pretty"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultTestConfig = Config{
		ReloaderConfig: operator.ReloaderConfig{
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos",
	}
	emptyQueryEndpoints = []string{""}
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
	}, defaultTestConfig, nil, "")

	require.NoError(t, err)

	if !reflect.DeepEqual(labels, sset.Labels) {
		t.Log(pretty.Compare(labels, sset.Labels))
		t.Fatal("Labels are not properly being propagated to the StatefulSet")
	}

	if !reflect.DeepEqual(expectedAnnotations, sset.Annotations) {
		t.Log(pretty.Compare(expectedAnnotations, sset.Annotations))
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
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, nil, "")
	require.NoError(t, err)
	if val, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]; !ok || val != "testvalue" {
		t.Fatal("Pod labels are not properly propagated")
	}
	if val, ok := sset.Spec.Template.ObjectMeta.Annotations["testannotation"]; !ok || val != "testvalue" {
		t.Fatal("Pod annotations are not properly propagated")
	}
}

func TestThanosDefaultBaseImageFlag(t *testing.T) {
	thanosBaseImageConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
		},
		ThanosDefaultBaseImage: "nondefaultuseflag/quay.io/thanos/thanos",
	}

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, thanosBaseImageConfig, nil, "")
	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/thanos/thanos" + ":" + operator.DefaultThanosVersion
	if image != expected {
		t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
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
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
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

func TestTracing(t *testing.T) {
	testKey := "thanos-tracing-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			TracingConfig: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	if sset.Spec.Template.Spec.Containers[0].Name != "thanos-ruler" {
		t.Fatalf("expected 1st containers to be thanos-ruler, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "TRACING_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos ruler is missing expected TRACING_CONFIG env var with correct value")
	}

	{
		var containsArg bool
		const expectedArg = "--tracing.config=$(TRACING_CONFIG)"
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArg)
		}
	}
}

func TestObjectStorage(t *testing.T) {
	testKey := "thanos-objstore-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			ObjectStorageConfig: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	if sset.Spec.Template.Spec.Containers[0].Name != "thanos-ruler" {
		t.Fatalf("expected 1st containers to be thanos-ruler, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "OBJSTORE_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos ruler is missing expected OBJSTORE_CONFIG env var with correct value")
	}

	{
		var containsArg bool
		const expectedArg = "--objstore.config=$(OBJSTORE_CONFIG)"
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArg)
		}
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
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
		if !containsArgConfigFile {
			t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArgConfigFile)
		}
		if containsArgConfig {
			t.Fatalf("Thanos ruler should not contain argument: %s", expectedArgConfig)
		}
	}
}

func TestAlertRelabel(t *testing.T) {
	testKey := "thanos-alertrelabel-config-secret"

	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.ThanosRulerSpec{
			QueryEndpoints: emptyQueryEndpoints,
			AlertRelabelConfigs: &v1.SecretKeySelector{
				Key: testKey,
			},
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	if sset.Spec.Template.Spec.Containers[0].Name != "thanos-ruler" {
		t.Fatalf("expected 1st containers to be thanos-ruler, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	var containsEnvVar bool
	for _, env := range sset.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "ALERT_RELABEL_CONFIG" {
			if env.ValueFrom.SecretKeyRef.Key == testKey {
				containsEnvVar = true
				break
			}
		}
	}
	if !containsEnvVar {
		t.Fatalf("Thanos ruler is missing expected ALERT_RELABEL_CONFIG env var with correct value")
	}

	{
		var containsArg bool
		const expectedArg = "--alert.relabel-config=$(ALERT_RELABEL_CONFIG)"
		for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
			if arg == expectedArg {
				containsArg = true
				break
			}
		}
		if !containsArg {
			t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArg)
		}
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
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
		if !containsArgConfigFile {
			t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArgConfigFile)
		}
		if containsArgConfigs {
			t.Fatalf("Thanos ruler should not contain argument: %s", expectedArgConfigs)
		}
	}
}

func TestLabelsAndAlertDropLabels(t *testing.T) {
	labelPrefix := "--label="
	alertDropLabelPrefix := "--alert.label-drop="

	tests := []struct {
		Labels                  map[string]string
		AlertDropLabels         []string
		ExpectedLabels          []string
		ExpectedAlertDropLabels []string
	}{
		{
			Labels:                  nil,
			AlertDropLabels:         nil,
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica"},
		},
		{
			Labels:                  nil,
			AlertDropLabels:         []string{"test"},
			ExpectedLabels:          []string{`thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test"},
		},
		{
			Labels: map[string]string{
				"test": "test",
			},
			AlertDropLabels:         nil,
			ExpectedLabels:          []string{`test="test"`, `thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica"},
		},
		{
			Labels: map[string]string{
				"test": "test",
			},
			AlertDropLabels:         []string{"test"},
			ExpectedLabels:          []string{`test="test"`, `thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test"},
		},
		{
			Labels: map[string]string{
				"thanos_ruler_replica": "$(POD_NAME)",
				"test":                 "test",
			},
			AlertDropLabels:         []string{"test", "aaa"},
			ExpectedLabels:          []string{`test="test"`, `thanos_ruler_replica="$(POD_NAME)"`, `thanos_ruler_replica="$(POD_NAME)"`},
			ExpectedAlertDropLabels: []string{"thanos_ruler_replica", "test", "aaa"},
		},
	}
	for _, tc := range tests {
		actualLabels := []string{}
		actualDropLabels := []string{}
		sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: monitoringv1.ThanosRulerSpec{
				QueryEndpoints:  emptyQueryEndpoints,
				Labels:          tc.Labels,
				AlertDropLabels: tc.AlertDropLabels,
			},
		}, defaultTestConfig, nil, "")
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		ruler := sset.Spec.Template.Spec.Containers[0]
		if ruler.Name != "thanos-ruler" {
			t.Fatalf("Expected 1st containers to be thanos-ruler, got %s", ruler.Name)
		}

		for _, arg := range ruler.Args {
			if strings.HasPrefix(arg, labelPrefix) {
				actualLabels = append(actualLabels, strings.TrimPrefix(arg, labelPrefix))
			} else if strings.HasPrefix(arg, alertDropLabelPrefix) {
				actualDropLabels = append(actualDropLabels, strings.TrimPrefix(arg, alertDropLabelPrefix))
			}
		}
		sort.Slice(actualLabels, func(i, j int) bool {
			return actualLabels[i] < actualLabels[j]
		})
		if !reflect.DeepEqual(actualLabels, tc.ExpectedLabels) {
			t.Fatal("label sets mismatch")
		}

		if !reflect.DeepEqual(actualDropLabels, tc.ExpectedAlertDropLabels) {
			t.Fatal("alert drop label sets mismatch")
		}
	}
}

func TestAdditionalContainers(t *testing.T) {
	// The base to compare everything against
	baseSet, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, defaultTestConfig, nil, "")
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
	}, defaultTestConfig, nil, "")
	require.NoError(t, err)

	if len(baseSet.Spec.Template.Spec.Containers)+1 != len(addSset.Spec.Template.Spec.Containers) {
		t.Fatalf("container count mismatch")
	}

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

func TestRetention(t *testing.T) {
	for _, tc := range []struct {
		specRetention     string
		expectedRetention string
		ok                bool
	}{
		{"", "24h", true},
		{"1d", "1d", true},
		{"1k", "", false},
		{"somevalue", "", false},
	} {
		t.Run(tc.specRetention, func(t *testing.T) {
			sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
				Spec: monitoringv1.ThanosRulerSpec{
					Retention:      tc.specRetention,
					QueryEndpoints: emptyQueryEndpoints,
				},
			}, defaultTestConfig, nil, "")

			if !tc.ok {
				if err == nil {
					t.Fatal("expecting error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error but got %q", err)
			}

			trArgs := sset.Spec.Template.Spec.Containers[0].Args
			expectedRetentionArg := fmt.Sprintf("--tsdb.retention=%s", tc.expectedRetention)
			found := false
			for _, flag := range trArgs {
				if flag == expectedRetentionArg {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("expected ThanosRuler args to contain %v, but got %v", expectedRetentionArg, trArgs)
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
	serviceAccountName := "thanos-ruler-sa"

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
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

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
}

func TestExternalQueryURL(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			AlertQueryURL:  "https://example.com/",
			QueryEndpoints: emptyQueryEndpoints,
		},
	}, defaultTestConfig, nil, "")
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	if sset.Spec.Template.Spec.Containers[0].Name != "thanos-ruler" {
		t.Fatalf("expected 1st containers to be thanos-ruler, got %s", sset.Spec.Template.Spec.Containers[0].Name)
	}

	const expectedArg = "--alert.query-url=https://example.com/"
	for _, arg := range sset.Spec.Template.Spec.Containers[0].Args {
		if arg == expectedArg {
			return
		}
	}
	t.Fatalf("Thanos ruler is missing expected argument: %s", expectedArg)
}

func TestSidecarsNoResources(t *testing.T) {
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "0",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "0",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "0",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "0",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "0",
			MemoryLimit:   "50Mi",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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
	testConfig := Config{
		ReloaderConfig: operator.ReloaderConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "0",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
		ThanosDefaultBaseImage: "quay.io/thanos/thanos:v0.7.0",
	}
	sset, err := makeStatefulSet(&monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{QueryEndpoints: emptyQueryEndpoints},
	}, testConfig, nil, "")
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

func TestStatefulSetMinReadySeconds(t *testing.T) {
	tr := monitoringv1.ThanosRuler{
		Spec: monitoringv1.ThanosRulerSpec{
			MinReadySeconds: nil,
			QueryEndpoints:  emptyQueryEndpoints,
		},
	}

	statefulSet, err := makeStatefulSetSpec(&tr, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	if statefulSet.MinReadySeconds != 0 {
		t.Fatalf("expected MinReadySeconds to be zero but got %d", statefulSet.MinReadySeconds)
	}

	// assert set correctly if not nil
	var expect uint32 = 5
	tr.Spec.MinReadySeconds = &expect
	statefulSet, err = makeStatefulSetSpec(&tr, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	if statefulSet.MinReadySeconds != int32(expect) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", expect, statefulSet.MinReadySeconds)
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
	}, defaultTestConfig, nil, "")

	require.NoError(t, err)
	ssetPvc := sset.Spec.VolumeClaimTemplates[0]
	if !reflect.DeepEqual(*pvc.Spec.StorageClassName, *ssetPvc.Spec.StorageClassName) {
		t.Fatal("Error adding PVC Spec to StatefulSetSpec")
	}
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
	}, defaultTestConfig, nil, "")

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
	}, defaultTestConfig, nil, "")

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral == nil ||
		!reflect.DeepEqual(ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName) {
		t.Fatal("Error adding Ephemeral Spec to StatefulSetSpec")
	}
}
