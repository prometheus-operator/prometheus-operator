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

	"github.com/kylelemons/godebug/pretty"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultTestConfig = Config{
		LocalHost:                    "localhost",
		ReloaderConfig:               operator.DefaultReloaderTestConfig.ReloaderConfig,
		AlertmanagerDefaultBaseImage: operator.DefaultAlertmanagerBaseImage,
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
		"testlabel": "testlabelvalue",
	}

	expectedPodLabels := map[string]string{
		"alertmanager":                 "",
		"app.kubernetes.io/name":       "alertmanager",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultAlertmanagerVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "",
	}

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, defaultTestConfig, "", nil)

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
	}, defaultTestConfig, "", nil)

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
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, "", nil)
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
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Labels: labels,
			},
		},
	}, defaultTestConfig, "", nil)

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
	}, defaultTestConfig, "", nil)

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
	}, defaultTestConfig, "", nil)

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

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				Ephemeral: &ephemeral,
			},
		},
	}, defaultTestConfig, "", nil)

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	if ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral == nil ||
		!reflect.DeepEqual(ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName) {
		t.Fatal("Error adding Ephemeral Spec to StatefulSetSpec")
	}
}

func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			ListenLocal: true,
		},
	}, defaultTestConfig, "", nil)
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

	if len(sset.Spec.Template.Spec.Containers[0].Ports) != 2 {
		t.Fatal("Alertmanager container should only have one port defined")
	}
}

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			Web: &monitoringv1.AlertmanagerWebSpec{
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
	}, defaultTestConfig, "", nil)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	expectedProbeHandler := func(probePath string) v1.ProbeHandler {
		return v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   probePath,
				Port:   intstr.FromString("web"),
				Scheme: "HTTPS",
			},
		}
	}

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		FailureThreshold: 10,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:        expectedProbeHandler("/-/ready"),
		InitialDelaySeconds: 3,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
		FailureThreshold:    10,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9093/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		fmt.Println(arg)

		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	if !reloadURLFound {
		t.Fatalf("expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)
	}

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=https://localhost:9093/-/reload",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
		}
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

		statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
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

func TestMakeStatefulSetSpecWebRoutePrefix(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
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

	if !containsWebRoutePrefix {
		t.Fatal("expected stateful set to contain arg '-web.route-prefix'")
	}
}

func TestMakeStatefulSetSpecWebTimeout(t *testing.T) {

	tt := []struct {
		scenario         string
		version          string
		web              *monitoringv1.AlertmanagerWebSpec
		expectTimeoutArg bool
	}{{
		scenario:         "no timeout by default",
		version:          operator.DefaultAlertmanagerVersion,
		web:              nil,
		expectTimeoutArg: false,
	}, {
		scenario: "no timeout for old version",
		version:  "0.16.9",
		web: &monitoringv1.AlertmanagerWebSpec{
			Timeout: toPtr(uint32(50)),
		},
		expectTimeoutArg: false,
	}, {
		scenario: "timeout arg set if specified",
		version:  operator.DefaultAlertmanagerVersion,
		web: &monitoringv1.AlertmanagerWebSpec{
			Timeout: toPtr(uint32(50)),
		},
		expectTimeoutArg: true,
	}}

	for _, ts := range tt {
		ts := ts
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Web = ts.web

			ss, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
			if err != nil {
				t.Fatal(err)
			}

			args := ss.Template.Spec.Containers[0].Args
			if got := slices.ContainsFunc(args, containsString("-web.timeout")); got != ts.expectTimeoutArg {
				t.Fatalf("expected alertmanager args %v web.timeout to be %v but is %v", args, ts.expectTimeoutArg, got)
			}

		})
	}
}

func TestMakeStatefulSetSpecWebConcurrency(t *testing.T) {

	tt := []struct {
		scenario                string
		version                 string
		web                     *monitoringv1.AlertmanagerWebSpec
		expectGetConcurrencyArg bool
	}{{
		scenario:                "no get-concurrency by default",
		version:                 operator.DefaultAlertmanagerVersion,
		web:                     nil,
		expectGetConcurrencyArg: false,
	}, {
		scenario: "no get-concurrency for old version",
		version:  "0.16.9",
		web: &monitoringv1.AlertmanagerWebSpec{
			GetConcurrency: toPtr(uint32(50)),
		},
		expectGetConcurrencyArg: false,
	}, {
		scenario: "get-concurrency arg set if specified",
		version:  operator.DefaultAlertmanagerVersion,

		web: &monitoringv1.AlertmanagerWebSpec{
			GetConcurrency: toPtr(uint32(50)),
		},
		expectGetConcurrencyArg: true,
	}}

	for _, ts := range tt {
		ts := ts
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Web = ts.web

			ss, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
			if err != nil {
				t.Fatal(err)
			}

			args := ss.Template.Spec.Containers[0].Args
			if got := slices.ContainsFunc(args, containsString("-web.get-concurrency")); got != ts.expectGetConcurrencyArg {
				t.Fatalf("expected alertmanager args %v web.get-concurrency to be %v but is %v", args, ts.expectGetConcurrencyArg, got)
			}

		})
	}
}

func TestMakeStatefulSetSpecPeersWithoutClusterDomain(t *testing.T) {
	replicas := int32(1)
	a := monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alertmanager",
			Namespace: "monitoring",
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Version:  "v0.15.3",
			Replicas: &replicas,
		},
	}

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	amArgs := statefulSet.Template.Spec.Containers[0].Args
	expectedArg := "--cluster.peer=alertmanager-alertmanager-0.alertmanager-operated:9094"
	for _, arg := range amArgs {
		if arg == expectedArg {
			found = true
		}
	}

	if !found {
		t.Fatalf("Cluster peer argument %v was not found in %v.", expectedArg, amArgs)
	}
}

func TestMakeStatefulSetSpecPeersWithClusterDomain(t *testing.T) {
	replicas := int32(1)
	a := monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alertmanager",
			Namespace: "monitoring",
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Version:  "v0.15.3",
			Replicas: &replicas,
		},
	}

	configWithClusterDomain := defaultTestConfig
	configWithClusterDomain.ClusterDomain = "custom.cluster"

	statefulSet, err := makeStatefulSetSpec(&a, configWithClusterDomain, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	amArgs := statefulSet.Template.Spec.Containers[0].Args
	expectedArg := "--cluster.peer=alertmanager-alertmanager-0.alertmanager-operated.monitoring.svc.custom.cluster.:9094"
	for _, arg := range amArgs {
		if arg == expectedArg {
			found = true
		}
	}

	if !found {
		t.Fatalf("Cluster peer argument %v was not found in %v.", expectedArg, amArgs)
	}
}

func TestMakeStatefulSetSpecAdditionalPeers(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	a.Spec.Version = "v0.15.3"
	replicas := int32(1)
	a.Spec.Replicas = &replicas
	a.Spec.AdditionalPeers = []string{"example.com"}

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	peerFound := false
	amArgs := statefulSet.Template.Spec.Containers[0].Args
	for _, arg := range amArgs {
		if strings.Contains(arg, "example.com") {
			peerFound = true
		}
	}

	if !peerFound {
		t.Fatal("Additional peers were not found.")
	}
}

func TestMakeStatefulSetSpecNotificationTemplates(t *testing.T) {
	replicas := int32(1)
	a := monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			Replicas: &replicas,
			AlertmanagerConfiguration: &monitoringv1.AlertmanagerConfiguration{
				Templates: []monitoringv1.SecretOrConfigMap{
					{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "template1",
							},
							Key: "template1.tmpl",
						},
					},
				},
			},
		},
	}
	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	var foundVM, foundV bool
	for _, vm := range statefulSet.Template.Spec.Containers[0].VolumeMounts {
		if vm.Name == "notification-templates" && vm.MountPath == alertmanagerTemplatesDir {
			foundVM = true
			break
		}
	}
	for _, v := range statefulSet.Template.Spec.Volumes {
		if v.Name == "notification-templates" && v.Projected != nil {
			for _, s := range v.Projected.Sources {
				if s.Secret != nil && s.Secret.Name == "template1" {
					foundV = true
					break
				}
			}
		}
	}

	if !(foundVM && foundV) {
		t.Fatal("Notification templates were not found.")
	}
}

func TestAdditionalSecretsMounted(t *testing.T) {
	secrets := []string{"secret1", "secret2"}
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			Secrets: secrets,
		},
	}, defaultTestConfig, "", nil)
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

func TestAlertManagerDefaultBaseImageFlag(t *testing.T) {
	alertManagerBaseImageConfig := Config{
		ReloaderConfig:               defaultTestConfig.ReloaderConfig,
		AlertmanagerDefaultBaseImage: "nondefaultuseflag/quay.io/prometheus/alertmanager",
	}

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
	}, alertManagerBaseImageConfig, "", nil)

	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/prometheus/alertmanager" + ":" + operator.DefaultAlertmanagerVersion
	if image != expected {
		t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
}

func TestSHAAndTagAndVersion(t *testing.T) {
	{
		sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
			},
		}, defaultTestConfig, "", nil)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/alertmanager:my-unrelated-tag"
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
	{
		sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
			},
		}, defaultTestConfig, "", nil)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/alertmanager@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		if image != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
		}
	}
	{
		image := "my-registry/alertmanager:latest"
		sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
				Image:   &image,
			},
		}, defaultTestConfig, "", nil)
		if err != nil {
			t.Fatalf("Unexpected error while making StatefulSet: %v", err)
		}

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "my-registry/alertmanager:latest"
		if resultImage != expected {
			t.Fatalf("Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
		}
	}
}

func TestRetention(t *testing.T) {
	tests := []struct {
		specRetention     monitoringv1.GoDuration
		expectedRetention monitoringv1.GoDuration
	}{
		{"", "120h"},
		{"1d", "1d"},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				Retention: test.specRetention,
			},
		}, defaultTestConfig, "", nil)
		if err != nil {
			t.Fatal(err)
		}

		amArgs := sset.Spec.Template.Spec.Containers[0].Args
		expectedRetentionArg := fmt.Sprintf("--data.retention=%s", test.expectedRetention)
		found := false
		for _, flag := range amArgs {
			if flag == expectedRetentionArg {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected Alertmanager args to contain %v, but got %v", expectedRetentionArg, amArgs)
		}
	}
}

func TestAdditionalConfigMap(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			ConfigMaps: []string{"test-cm1"},
		},
	}, defaultTestConfig, "", nil)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	cmVolumeFound := false
	for _, v := range sset.Spec.Template.Spec.Volumes {
		if v.Name == "configmap-test-cm1" {
			cmVolumeFound = true
			break
		}
	}
	if !cmVolumeFound {
		t.Fatal("ConfigMap volume not found")
	}

	cmMounted := false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if v.Name == "configmap-test-cm1" && v.MountPath == "/etc/alertmanager/configmaps/test-cm1" {
			cmMounted = true
			break
		}
	}
	if !cmMounted {
		t.Fatal("ConfigMap volume not mounted")
	}
}

func TestSidecarResources(t *testing.T) {
	operator.TestSidecarsResources(t, func(reloaderConfig operator.ContainerConfig) *appsv1.StatefulSet {
		testConfig := Config{
			ReloaderConfig:               reloaderConfig,
			AlertmanagerDefaultBaseImage: operator.DefaultAlertmanagerBaseImage,
		}
		am := &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{},
		}

		sset, err := makeStatefulSet(am, testConfig, "", nil)
		require.NoError(t, err)
		return sset
	})
}

func TestTerminationPolicy(t *testing.T) {
	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{},
	}, defaultTestConfig, "", nil)
	if err != nil {
		t.Fatalf("Unexpected error while making StatefulSet: %v", err)
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.TerminationMessagePolicy != v1.TerminationMessageFallbackToLogsOnError {
			t.Fatalf("Unexpected TermintationMessagePolicy. Expected %v got %v", v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy)
		}
	}
}

func TestClusterListenAddressForSingleReplica(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsEmptyClusterListenAddress := false

	for _, arg := range amArgs {
		if arg == "--cluster.listen-address=" {
			containsEmptyClusterListenAddress = true
		}
	}

	if !containsEmptyClusterListenAddress {
		t.Fatal("expected stateful set to contain arg '--cluster.listen-address='")
	}
}

func TestClusterListenAddressForSingleReplicaWithForceEnableClusterMode(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas
	a.Spec.ForceEnableClusterMode = true

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsEmptyClusterListenAddress := false

	for _, arg := range amArgs {
		if arg == "--cluster.listen-address=" {
			containsEmptyClusterListenAddress = true
		}
	}

	if containsEmptyClusterListenAddress {
		t.Fatal("expected stateful set to not contain arg '--cluster.listen-address='")
	}
}

func TestClusterListenAddressForMultiReplica(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(3)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsClusterListenAddress := false

	for _, arg := range amArgs {
		if arg == "--cluster.listen-address=[$(POD_IP)]:9094" {
			containsClusterListenAddress = true
		}
	}

	if !containsClusterListenAddress {
		t.Fatal("expected stateful set to contain arg '--cluster.listen-address=[$(POD_IP)]:9094'")
	}
}

func TestExpectStatefulSetMinReadySeconds(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(3)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	// assert defaults to zero if nil
	statefulSet, err := makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	if statefulSet.MinReadySeconds != 0 {
		t.Fatalf("expected MinReadySeconds to be zero but got %d", statefulSet.MinReadySeconds)
	}

	// assert set correctly if not nil
	var expect uint32 = 5
	a.Spec.MinReadySeconds = &expect
	statefulSet, err = makeStatefulSetSpec(&a, defaultTestConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	if statefulSet.MinReadySeconds != int32(expect) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", expect, statefulSet.MinReadySeconds)
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
	serviceAccountName := "alertmanager-sa"
	hostAliases := []monitoringv1.HostAlias{
		{
			Hostnames: []string{"foo.com"},
			IP:        "1.1.1.1",
		},
	}
	imagePullSecrets := []v1.LocalObjectReference{
		{
			Name: "registry-secret",
		},
	}
	imagePullPolicy := v1.PullAlways

	sset, err := makeStatefulSet(&monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			NodeSelector:       nodeSelector,
			Affinity:           &affinity,
			Tolerations:        tolerations,
			SecurityContext:    &securityContext,
			PriorityClassName:  priorityClassName,
			ServiceAccountName: serviceAccountName,
			HostAliases:        hostAliases,
			ImagePullSecrets:   imagePullSecrets,
			ImagePullPolicy:    imagePullPolicy,
		},
	}, defaultTestConfig, "", nil)
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
	if len(sset.Spec.Template.Spec.HostAliases) != len(hostAliases) {
		t.Fatalf("expected length of host aliases to match, want %d, got %d", len(hostAliases), len(sset.Spec.Template.Spec.HostAliases))
	}
	if !reflect.DeepEqual(sset.Spec.Template.Spec.ImagePullSecrets, imagePullSecrets) {
		t.Fatalf("expected image pull secrets to match, want %s, got %s", imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets)
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
}

func TestConfigReloader(t *testing.T) {
	baseSet, err := makeStatefulSet(&monitoringv1.Alertmanager{}, defaultTestConfig, "", nil)
	require.NoError(t, err)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=http://localhost:9093/-/reload",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
	}

	for _, c := range baseSet.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expectd container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
		}
	}

	expectedArgsInitConfigReloader := []string{
		"--watch-interval=0",
		"--listen-address=:8080",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
	}

	for _, c := range baseSet.Spec.Template.Spec.Containers {
		if c.Name == "init-config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expectd init container args are %s, but found %s", expectedArgsInitConfigReloader, c.Args)
			}
		}
	}

}

func containsString(sub string) func(string) bool {
	return func(x string) bool {
		return strings.Contains(x, sub)
	}
}

func toPtr[T any](t T) *T {
	return &t
}
