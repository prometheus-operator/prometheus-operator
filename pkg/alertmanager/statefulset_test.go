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

package alertmanager

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultTestConfig = Config{
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

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, nil, defaultTestConfig)

	require.NoError(t, err)

	if !reflect.DeepEqual(labels, sset.Labels) || !reflect.DeepEqual(annotations, sset.Annotations) {
		t.Fatal("Labels or Annotations are not properly being propagated to the StatefulSet")
	}
}

func TestStatefulSetStoragePath(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, nil, defaultTestConfig)

	require.NoError(t, err)

	reg := strings.Join(sset.Spec.Template.Spec.Containers[0].Args, " ")
	for _, k := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if k.Name == "config-volume" {
			if !strings.Contains(reg, k.MountPath) {
				t.Fatal("config-volume Path not configured correctly")
			} else {
				return
			}

		}
	}
	t.Fatal("config-volume not set")
}

func TestPodLabelsAnnotations(t *testing.T) {
	annotations := map[string]string{
		"testannotation": "testvalue",
	}
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			PodMetadata: &metav1.ObjectMeta{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, nil, defaultTestConfig)
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

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: pvc,
			},
		},
	}, nil, defaultTestConfig)

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

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				EmptyDir: &emptyDir,
			},
		},
	}, nil, defaultTestConfig)

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir != nil && !reflect.DeepEqual(emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium) {
		t.Fatal("Error adding EmptyDir Spec to StatefulSetSpec")
	}
}
func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			ListenLocal: true,
		},
	}, nil, defaultTestConfig)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.listen-address=127.0.0.1:9093" {
			found = true
		}
	}

	if !found {
		t.Fatal("Alertmanager not listening on loopback when it should.")
	}

	if sset.Spec.Template.Spec.Containers[0].ReadinessProbe != nil {
		t.Fatal("Alertmanager readiness probe expected to be empty")
	}

	if sset.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
		t.Fatal("Alertmanager readiness probe expected to be empty")
	}

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 1 {
		t.Fatal("Alertmanager container should only have one port defined")
	}
}

// below Alertmanager v0.13.0 all flags are with single dash.
func TestMakeStatefulSetSpecSingleDoubleDashedArgs(t *testing.T) {
	tests := []struct {
		version string
		prefix  string
		amount  int
	}{
		{"v0.12.0", "-", 1},
		{"v0.13.0", "--", 2},
	}

	for _, test := range tests {
		a := monitoringv1.Alertmanager{}
		a.Spec.Version = test.version
		replicas := int32(3)
		a.Spec.Replicas = &replicas

		statefulSet, err := makeStatefulSetSpec(&a, Config{})
		if err != nil {
			t.Fatal(err)
		}

		amArgs := statefulSet.Template.Spec.Containers[0].Args

		for _, arg := range amArgs {
			if arg[:test.amount] != test.prefix {
				t.Fatalf("expected all args to start with %v but got %v", test.prefix, arg)
			}
		}
	}
}

// below Alertmanager v0.7.0 the flag 'web.route-prefix' does not exist
func TestMakeStatefulSetSpecWebRoutePrefix(t *testing.T) {
	tests := []struct {
		version              string
		expectWebRoutePrefix bool
	}{
		{"v0.6.0", false},
		{"v0.7.0", true},
	}

	for _, test := range tests {
		a := monitoringv1.Alertmanager{}
		a.Spec.Version = test.version
		replicas := int32(1)
		a.Spec.Replicas = &replicas

		statefulSet, err := makeStatefulSetSpec(&a, Config{})
		if err != nil {
			t.Fatal(err)
		}

		amArgs := statefulSet.Template.Spec.Containers[0].Args

		containsWebRoutePrefix := false

		for _, arg := range amArgs {
			if strings.Contains(arg, "-web.route-prefix") {
				containsWebRoutePrefix = true
			}
		}

		if containsWebRoutePrefix != test.expectWebRoutePrefix {
			t.Fatalf("expected stateful set containing arg '-web.route-prefix' to be: %v", test.expectWebRoutePrefix)
		}
	}
}

// below Alertmanager v0.15.0 high availability flags are prefixed with 'mesh' instead of 'cluster'
func TestMakeStatefulSetSpecMeshClusterFlags(t *testing.T) {
	tests := []struct {
		version       string
		rightHAPrefix string
		wrongHAPrefix string
	}{
		{"v0.14.0", "mesh", "cluster"},
		{"v0.15.0", "cluster", "mesh"},
	}

	for _, test := range tests {
		a := monitoringv1.Alertmanager{}
		a.Spec.Version = test.version
		replicas := int32(3)
		a.Spec.Replicas = &replicas

		statefulSet, err := makeStatefulSetSpec(&a, Config{})
		if err != nil {
			t.Fatal(err)
		}

		haFlags := []string{"--%v.listen-address", "--%v.peer="}

		amArgs := statefulSet.Template.Spec.Containers[0].Args

		for _, flag := range haFlags {
			if sliceContains(amArgs, fmt.Sprintf(flag, test.wrongHAPrefix)) {
				t.Fatalf("expected Alertmanager args not to contain %v, but got %v", test.wrongHAPrefix, amArgs)
			}
			if !sliceContains(amArgs, fmt.Sprintf(flag, test.rightHAPrefix)) {
				t.Fatalf("expected Alertmanager args to contain %v, but got %v", test.rightHAPrefix, amArgs)
			}
		}
	}
}

// below Alertmanager v0.15.0 peer address port specification is not necessary
func TestMakeStatefulSetSpecPeerFlagPort(t *testing.T) {
	tests := []struct {
		version    string
		portNeeded bool
	}{
		{"v0.14.0", false},
		{"v0.15.0", true},
	}

	for _, test := range tests {
		a := monitoringv1.Alertmanager{}
		a.Spec.Version = test.version
		replicas := int32(3)
		a.Spec.Replicas = &replicas

		statefulSet, err := makeStatefulSetSpec(&a, Config{})
		if err != nil {
			t.Fatal(err)
		}

		amArgs := statefulSet.Template.Spec.Containers[0].Args

		for _, arg := range amArgs {
			if strings.Contains(arg, ".peer") {
				if strings.Contains(arg, ":6783") != test.portNeeded {
					t.Fatalf("expected arg '%v' containing port specification to be: %v", arg, test.portNeeded)
				}
			}
		}
	}
}

func TestAdditionalSecretsMounted(t *testing.T) {
	secrets := []string{"secret1", "secret2"}
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			Secrets: secrets,
		},
	}, nil, defaultTestConfig)
	require.NoError(t, err)

	secret1Found := false
	secret2Found := false
	for _, v := range sset.Spec.Template.Spec.Volumes {
		if v.Secret != nil {
			if v.Secret.SecretName == "secret1" {
				secret1Found = true
			}
			if v.Secret.SecretName == "secret2" {
				secret2Found = true
			}
		}
	}

	if !(secret1Found && secret2Found) {
		t.Fatal("Additional secrets were not found.")
	}

	secret1Found = false
	secret2Found = false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if v.Name == "secret-secret1" && v.MountPath == "/etc/alertmanager/secrets/secret1" {
			secret1Found = true
		}
		if v.Name == "secret-secret2" && v.MountPath == "/etc/alertmanager/secrets/secret2" {
			secret2Found = true
		}
	}

	if !(secret1Found && secret2Found) {
		t.Fatal("Additional secrets were not found.")
	}
}

func sliceContains(slice []string, match string) bool {
	contains := false
	for _, s := range slice {
		if strings.Contains(s, match) {
			contains = true
		}
	}
	return contains
}
