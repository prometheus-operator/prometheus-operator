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
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
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
		"prometheus-operator-input-hash": "abc",
		"testannotation":                 "testannotationvalue",
	}

	expectedStatefulSetLabels := map[string]string{
		"testlabel":                    "testlabelvalue",
		"managed-by":                   "prometheus-operator",
		"alertmanager":                 "test",
		"app.kubernetes.io/instance":   "test",
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/name":       "alertmanager",
	}

	expectedPodLabels := map[string]string{
		"alertmanager":                 "test",
		"app.kubernetes.io/name":       "alertmanager",
		"app.kubernetes.io/version":    strings.TrimPrefix(operator.DefaultAlertmanagerVersion, "v"),
		"app.kubernetes.io/managed-by": "prometheus-operator",
		"app.kubernetes.io/instance":   "test",
	}

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "ns",
			Labels:      labels,
			Annotations: annotations,
		},
	}, defaultTestConfig, "abc", &operator.ShardedSecret{})

	require.NoError(t, err)

	compareLabels := pretty.Compare(expectedStatefulSetLabels, sset.Labels)
	require.Equalf(t, expectedStatefulSetLabels, sset.Labels, "Labels are not properly being propagated to the StatefulSet:\n%s", compareLabels)

	compareAnnotations := pretty.Compare(expectedStatefulSetAnnotations, sset.Annotations)
	require.Equalf(t, expectedStatefulSetAnnotations, sset.Annotations, "Annotations are not properly being propagated to the StatefulSet:\n%s", compareAnnotations)

	comparePodLabels := pretty.Compare(expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels)
	require.Equalf(t, expectedPodLabels, sset.Spec.Template.ObjectMeta.Labels, "Labels are not properly being propagated to the Pod:\n%s", comparePodLabels)
}

func TestStatefulSetStoragePath(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)

	reg := strings.Join(sset.Spec.Template.Spec.Containers[0].Args, " ")
	for _, k := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if k.Name == "config-volume" {
			require.Contains(t, reg, k.MountPath, "config-volume Path not configured correctly")
			return
		}
	}
	require.Fail(t, "config-volume not set")
}

func TestPodLabelsAnnotations(t *testing.T) {
	annotations := map[string]string{
		"testannotation": "testvalue",
	}
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Annotations: annotations,
				Labels:      labels,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	valLabels, ok := sset.Spec.Template.ObjectMeta.Labels["testlabel"]
	require.True(t, ok && valLabels == "testvalue", "Pod labels are not properly propagated")

	valAnnotations, ok := sset.Spec.Template.ObjectMeta.Annotations["testannotation"]
	require.True(t, ok && valAnnotations == "testvalue", "Pod annotations are not properly propagated")
}

func TestPodLabelsShouldNotBeSelectorLabels(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testvalue",
	}
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
				Labels: labels,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)

	require.NotEqual(t, "testvalue", sset.Spec.Selector.MatchLabels["testlabel"], "Pod Selector are not properly propagated")
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

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				VolumeClaimTemplate: pvc,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetPvc := sset.Spec.VolumeClaimTemplates[0]
	require.Equal(t, *pvc.Spec.StorageClassName, *ssetPvc.Spec.StorageClassName, "Error adding PVC Spec to StatefulSetSpec")
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

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				EmptyDir: &emptyDir,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir, "Error adding EmptyDir Spec to StatefulSetSpec")
	require.Equal(t, emptyDir.Medium, ssetVolumes[len(ssetVolumes)-1].VolumeSource.EmptyDir.Medium, "Error adding EmptyDir Spec to StatefulSetSpec")
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

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Storage: &monitoringv1.StorageSpec{
				Ephemeral: &ephemeral,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)
	ssetVolumes := sset.Spec.Template.Spec.Volumes
	require.NotNil(t, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral, "Error adding Ephemeral Spec to StatefulSetSpec")
	require.Equal(t, ephemeral.VolumeClaimTemplate.Spec.StorageClassName, ssetVolumes[len(ssetVolumes)-1].VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName, "Error adding Ephemeral Spec to StatefulSetSpec")
}

func TestListenLocal(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			ListenLocal: true,
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	found := false
	for _, flag := range sset.Spec.Template.Spec.Containers[0].Args {
		if flag == "--web.listen-address=127.0.0.1:9093" {
			found = true
		}
	}

	require.True(t, found, "Alertmanager not listening on loopback when it should.")

	require.Nil(t, sset.Spec.Template.Spec.Containers[0].ReadinessProbe, "Alertmanager readiness probe expected to be empty")

	require.Nil(t, sset.Spec.Template.Spec.Containers[0].LivenessProbe, "Alertmanager readiness probe expected to be empty")

	require.Len(t, sset.Spec.Template.Spec.Containers[0].Ports, 2, "Alertmanager container should only have one port defined")
	for _, port := range sset.Spec.Template.Spec.Containers[0].Ports {
		require.Condition(
			t,
			func() bool {
				return port.Name == alertmanagerMeshUDPPortName || port.Name == alertmanagerMeshTCPPortName
			},
		)
	}
}

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
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
	}, defaultTestConfig, "", &operator.ShardedSecret{})
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

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		FailureThreshold: 10,
	}
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe, "Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:        expectedProbeHandler("/-/ready"),
		InitialDelaySeconds: 3,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
		FailureThreshold:    10,
	}
	require.Equal(t, expectedLivenessProbe, actualLivenessProbe, "Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9093/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		fmt.Println(arg)

		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	require.True(t, reloadURLFound, "expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/alertmanager/web_config/web-config.yaml",
		"--reload-url=https://localhost:9093/-/reload",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
		"--watched-dir=/etc/alertmanager/config",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
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

		statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
		require.NoError(t, err)

		amArgs := statefulSet.Template.Spec.Containers[0].Args

		for _, arg := range amArgs {
			require.Equal(t, test.prefix, arg[:test.amount], "expected all args to start with %v but got %v", test.prefix, arg)
		}
	}
}

func TestMakeStatefulSetSpecWebRoutePrefix(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsWebRoutePrefix := false

	for _, arg := range amArgs {
		if strings.Contains(arg, "-web.route-prefix") {
			containsWebRoutePrefix = true
		}
	}

	require.True(t, containsWebRoutePrefix, "expected stateful set to contain arg '-web.route-prefix'")
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
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Web = ts.web

			ss, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			args := ss.Template.Spec.Containers[0].Args
			got := slices.ContainsFunc(args, containsString("-web.timeout"))
			require.Equal(t, ts.expectTimeoutArg, got, "expected alertmanager args %v web.timeout to be %v but is %v", args, ts.expectTimeoutArg, got)
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
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Web = ts.web

			ss, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			args := ss.Template.Spec.Containers[0].Args
			got := slices.ContainsFunc(args, containsString("-web.get-concurrency"))
			require.Equal(t, ts.expectGetConcurrencyArg, got, "expected alertmanager args %v web.get-concurrency to be %v but is %v", args, ts.expectGetConcurrencyArg, got)
		})
	}
}

func TestMakeStatefulSetSpecMaxSilences(t *testing.T) {
	tt := []struct {
		scenario             string
		version              string
		limits               *monitoringv1.AlertmanagerLimitsSpec
		expectMaxSilencesArg bool
	}{
		{
			scenario:             "no maxSilences by default",
			version:              operator.DefaultAlertmanagerVersion,
			limits:               nil,
			expectMaxSilencesArg: false,
		}, {
			scenario: "no maxSilencesfor old version",
			version:  "0.27.9",
			limits: &monitoringv1.AlertmanagerLimitsSpec{
				MaxSilences: toPtr(int32(50)),
			},
			expectMaxSilencesArg: false,
		}, {
			scenario: "maxSilencesfor arg set if specified",
			version:  operator.DefaultAlertmanagerVersion,
			limits: &monitoringv1.AlertmanagerLimitsSpec{
				MaxSilences: toPtr(int32(50)),
			},
			expectMaxSilencesArg: true,
		},
	}

	for _, ts := range tt {
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Limits = ts.limits

			ss, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			args := ss.Template.Spec.Containers[0].Args
			got := slices.ContainsFunc(args, containsString("--silences.max-silences"))
			require.Equal(t, ts.expectMaxSilencesArg, got, "expected alertmanager args %v silences.max-silences to be %v but is %v", args, ts.expectMaxSilencesArg, got)
		})
	}
}

func TestMakeStatefulSetSpecMaxPerSilenceBytes(t *testing.T) {
	tt := []struct {
		scenario                    string
		version                     string
		limits                      *monitoringv1.AlertmanagerLimitsSpec
		expectMaxPerSilenceBytesArg bool
	}{
		{
			scenario:                    "no maxPerSilenceBytes by default",
			version:                     operator.DefaultAlertmanagerVersion,
			limits:                      nil,
			expectMaxPerSilenceBytesArg: false,
		}, {
			scenario: "no maxPerSilenceBytes old version",
			version:  "0.27.9",
			limits: &monitoringv1.AlertmanagerLimitsSpec{
				MaxPerSilenceBytes: toPtr(monitoringv1.ByteSize("5MB")),
			},
			expectMaxPerSilenceBytesArg: false,
		}, {
			scenario: "maxPerSilenceBytes arg set if specified",
			version:  operator.DefaultAlertmanagerVersion,
			limits: &monitoringv1.AlertmanagerLimitsSpec{
				MaxPerSilenceBytes: toPtr(monitoringv1.ByteSize("5MB")),
			},
			expectMaxPerSilenceBytesArg: true,
		},
	}

	for _, ts := range tt {
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{}
			a.Spec.Replicas = toPtr(int32(1))

			a.Spec.Version = ts.version
			a.Spec.Limits = ts.limits

			ss, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			args := ss.Template.Spec.Containers[0].Args
			got := slices.ContainsFunc(args, containsString("--silences.max-per-silence-bytes=5242880"))
			require.Equal(t, ts.expectMaxPerSilenceBytesArg, got, "expected alertmanager args %v silences.max-silences to be %v but is %v", args, ts.expectMaxPerSilenceBytesArg, got)
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

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	found := false
	amArgs := statefulSet.Template.Spec.Containers[0].Args
	expectedArg := "--cluster.peer=alertmanager-alertmanager-0.alertmanager-operated:9094"
	for _, arg := range amArgs {
		if arg == expectedArg {
			found = true
		}
	}

	require.True(t, found, "Cluster peer argument %v was not found in %v.", expectedArg, amArgs)
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

	statefulSet, err := makeStatefulSetSpec(nil, &a, configWithClusterDomain, &operator.ShardedSecret{})
	require.NoError(t, err)

	amArgs := statefulSet.Template.Spec.Containers[0].Args
	// Expected: --cluster.peer=alertmanager-<name>-0.<serviceName>.<namespace>.svc.<clusterDomain>.:9094
	expectedArg := "--cluster.peer=alertmanager-alertmanager-0.alertmanager-operated.monitoring.svc.custom.cluster.:9094"
	require.True(t, slices.Contains(amArgs, expectedArg), "Cluster peer argument %v was not found in %v.", expectedArg, amArgs)
}

func TestMakeStatefulSetSpecWithCustomServiceName(t *testing.T) {
	replicas := int32(1)
	customServiceName := "my-custom-alertmanager-svc"
	am := &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alertmanager-custom-svc",
			Namespace: "custom-ns",
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Version:     "v0.25.0",
			Replicas:    &replicas,
			ServiceName: &customServiceName,
		},
	}

	cfg := defaultTestConfig
	cfg.ClusterDomain = "cluster.local"

	spec, err := makeStatefulSetSpec(nil, am, cfg, &operator.ShardedSecret{})
	require.NoError(t, err)

	// Check StatefulSet.Spec.ServiceName
	require.Equal(t, customServiceName, spec.ServiceName, "StatefulSet.Spec.ServiceName should be the custom service name")

	// Check cluster.peer arguments
	amArgs := spec.Template.Spec.Containers[0].Args
	expectedPeerArg := fmt.Sprintf("--cluster.peer=alertmanager-%s-0.%s.%s.svc.%s.:9094", am.Name, customServiceName, am.Namespace, cfg.ClusterDomain)
	require.True(t, slices.Contains(amArgs, expectedPeerArg), "expected cluster.peer argument %q not found in %v", expectedPeerArg, amArgs)
}

func TestMakeStatefulSetSpecWithDefaultServiceName(t *testing.T) {
	replicas := int32(1)
	am := &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alertmanager-default-svc",
			Namespace: "default-ns",
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Version:  "v0.25.0",
			Replicas: &replicas,
			// ServiceName is not set, so default should be used
		},
	}

	cfg := defaultTestConfig
	cfg.ClusterDomain = "cluster.local"

	spec, err := makeStatefulSetSpec(nil, am, cfg, &operator.ShardedSecret{})
	require.NoError(t, err)

	defaultServiceName := "alertmanager-operated"

	// 1. Check StatefulSet.Spec.ServiceName
	require.Equal(t, defaultServiceName, spec.ServiceName, "StatefulSet.Spec.ServiceName should be the default service name")

	// 2. Check cluster.peer arguments
	amArgs := spec.Template.Spec.Containers[0].Args
	expectedPeerArg := fmt.Sprintf("--cluster.peer=alertmanager-%s-0.%s.%s.svc.%s.:9094", am.Name, defaultServiceName, am.Namespace, cfg.ClusterDomain)
	require.True(t, slices.Contains(amArgs, expectedPeerArg), "expected cluster.peer argument %q not found in %v", expectedPeerArg, amArgs)
}

func TestMakeStatefulSetSpecAdditionalPeers(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	a.Spec.Version = "v0.15.3"
	replicas := int32(1)
	a.Spec.Replicas = &replicas
	a.Spec.AdditionalPeers = []string{"example.com"}

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	peerFound := false
	amArgs := statefulSet.Template.Spec.Containers[0].Args
	for _, arg := range amArgs {
		if strings.Contains(arg, "example.com") {
			peerFound = true
		}
	}

	require.True(t, peerFound, "Additional peers were not found.")
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
	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	var foundConfigReloaderVM, foundVM, foundV bool
	for _, vm := range statefulSet.Template.Spec.Containers[0].VolumeMounts {
		if vm.Name == "notification-templates" && vm.MountPath == alertmanagerTemplatesDir {
			foundVM = true
			break
		}
	}
	for _, vm := range statefulSet.Template.Spec.Containers[1].VolumeMounts {
		if vm.Name == "notification-templates" && vm.MountPath == alertmanagerTemplatesDir {
			foundConfigReloaderVM = true
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

	require.True(t, (foundVM && foundV), "Notification templates were not found.")
	require.True(t, foundConfigReloaderVM, "Config reloader notification templates volume mount was not found.")

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/alertmanager/web_config/web-config.yaml",
		"--reload-url=http://localhost:9093/-/reload",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
		"--watched-dir=/etc/alertmanager/config",
		"--watched-dir=/etc/alertmanager/templates",
	}

	for _, c := range statefulSet.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expectd container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
		}
	}
}

func TestAdditionalSecretsMounted(t *testing.T) {
	secrets := []string{"secret1", "secret2"}
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			Secrets: secrets,
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
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

	require.True(t, (secret1Found && secret2Found), "Additional secrets were not found.")

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

	require.True(t, (secret1Found && secret2Found), "Additional secrets were not found.")
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

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, alertManagerBaseImageConfig, "", &operator.ShardedSecret{})

	require.NoError(t, err)

	image := sset.Spec.Template.Spec.Containers[0].Image
	expected := "nondefaultuseflag/quay.io/prometheus/alertmanager" + ":" + operator.DefaultAlertmanagerVersion
	require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
}

func TestSHAAndTagAndVersion(t *testing.T) {
	{
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
			},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/alertmanager:my-unrelated-tag"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	{
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
			},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)

		image := sset.Spec.Template.Spec.Containers[0].Image
		expected := "quay.io/prometheus/alertmanager@sha256:7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324"
		require.Equal(t, expected, image, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, image)
	}
	{
		image := "my-registry/alertmanager:latest"
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				SHA:     "7384a79f4b4991bf8269e7452390249b7c70bcdd10509c8c1c6c6e30e32fb324",
				Tag:     "my-unrelated-tag",
				Version: "v0.15.3",
				Image:   &image,
			},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)

		resultImage := sset.Spec.Template.Spec.Containers[0].Image
		expected := "my-registry/alertmanager:latest"
		require.Equal(t, expected, resultImage, "Unexpected container image.\n\nExpected: %s\n\nGot: %s", expected, resultImage)
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
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				Retention: test.specRetention,
			},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)

		amArgs := sset.Spec.Template.Spec.Containers[0].Args
		expectedRetentionArg := fmt.Sprintf("--data.retention=%s", test.expectedRetention)

		require.True(t, slices.Contains(amArgs, expectedRetentionArg), "expected Alertmanager args to contain %v, but got %v", expectedRetentionArg, amArgs)
	}
}

func TestAdditionalConfigMap(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			ConfigMaps: []string{"test-cm1"},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	cmVolumeFound := false
	for _, v := range sset.Spec.Template.Spec.Volumes {
		if v.Name == "configmap-test-cm1" {
			cmVolumeFound = true
			break
		}
	}
	require.True(t, cmVolumeFound, "ConfigMap volume not found")

	cmMounted := false
	for _, v := range sset.Spec.Template.Spec.Containers[0].VolumeMounts {
		if v.Name == "configmap-test-cm1" && v.MountPath == "/etc/alertmanager/configmaps/test-cm1" {
			cmMounted = true
			break
		}
	}
	require.True(t, cmMounted, "ConfigMap volume not mounted")
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

		sset, err := makeStatefulSet(nil, am, testConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)
		return sset
	})
}

func TestTerminationPolicy(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	for _, c := range sset.Spec.Template.Spec.Containers {
		require.Equal(t, v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy, "Unexpected TermintationMessagePolicy. Expected %v got %v", v1.TerminationMessageFallbackToLogsOnError, c.TerminationMessagePolicy)
	}
}

func TestClusterListenAddressForSingleReplica(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	a.Spec.ForceEnableClusterMode = false

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	require.Contains(t, amArgs, "--cluster.listen-address=", "expected stateful set to contain '--cluster.listen-address='")
}

func TestClusterListenAddressForSingleReplicaWithForceEnableClusterMode(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(1)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas
	a.Spec.ForceEnableClusterMode = true

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsEmptyClusterListenAddress := false

	for _, arg := range amArgs {
		if arg == "--cluster.listen-address=" {
			containsEmptyClusterListenAddress = true
		}
	}

	require.False(t, containsEmptyClusterListenAddress, "expected stateful set to not contain arg '--cluster.listen-address='")
}

func TestClusterListenAddressForMultiReplica(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	replicas := int32(3)
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = &replicas

	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	amArgs := statefulSet.Template.Spec.Containers[0].Args

	containsClusterListenAddress := false

	for _, arg := range amArgs {
		if arg == "--cluster.listen-address=[$(POD_IP)]:9094" {
			containsClusterListenAddress = true
		}
	}

	require.True(t, containsClusterListenAddress, "expected stateful set to contain arg '--cluster.listen-address=[$(POD_IP)]:9094'")
}

func TestExpectStatefulSetMinReadySeconds(t *testing.T) {
	a := monitoringv1.Alertmanager{}
	a.Spec.Version = operator.DefaultAlertmanagerVersion
	a.Spec.Replicas = ptr.To(int32(3))

	// assert defaults to zero if nil
	statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, int32(0), statefulSet.MinReadySeconds)

	// assert set correctly if not nil
	a.Spec.MinReadySeconds = ptr.To(int32(5))
	statefulSet, err = makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)
	require.Equal(t, int32(5), statefulSet.MinReadySeconds)
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
	hostUsers := true
	hostNetwork := false

	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
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
			HostUsers:          ptr.To(true),
			HostNetwork:        hostNetwork,
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, sset.Spec.Template.Spec.NodeSelector, nodeSelector, "expected node selector to match, want %v, got %v", nodeSelector, sset.Spec.Template.Spec.NodeSelector)
	require.Equal(t, *sset.Spec.Template.Spec.Affinity, affinity, "expected affinity to match, want %v, got %v", affinity, *sset.Spec.Template.Spec.Affinity)
	require.Equal(t, sset.Spec.Template.Spec.Tolerations, tolerations, "expected tolerations to match, want %v, got %v", tolerations, sset.Spec.Template.Spec.Tolerations)
	require.Equal(t, *sset.Spec.Template.Spec.SecurityContext, securityContext, "expected security context  to match, want %v, got %v", securityContext, *sset.Spec.Template.Spec.SecurityContext)
	require.Equal(t, sset.Spec.Template.Spec.PriorityClassName, priorityClassName, "expected priority class name to match, want %s, got %s", priorityClassName, sset.Spec.Template.Spec.PriorityClassName)
	require.Equal(t, sset.Spec.Template.Spec.ServiceAccountName, serviceAccountName, "expected service account name to match, want %s, got %s", serviceAccountName, sset.Spec.Template.Spec.ServiceAccountName)
	require.Equal(t, len(sset.Spec.Template.Spec.HostAliases), len(hostAliases), "expected length of host aliases to match, want %d, got %d", len(hostAliases), len(sset.Spec.Template.Spec.HostAliases))
	require.Equal(t, sset.Spec.Template.Spec.ImagePullSecrets, imagePullSecrets, "expected image pull secrets to match, want %s, got %s", imagePullSecrets, sset.Spec.Template.Spec.ImagePullSecrets)
	require.Equal(t, *sset.Spec.Template.Spec.HostUsers, hostUsers, "expected host users to match, want %s, got %s", hostUsers, sset.Spec.Template.Spec.HostUsers)
	require.Equal(t, sset.Spec.Template.Spec.HostNetwork, hostNetwork, "expected host network to match, want %v, got %v", hostNetwork, sset.Spec.Template.Spec.HostNetwork)

	for _, initContainer := range sset.Spec.Template.Spec.InitContainers {
		require.Equal(t, initContainer.ImagePullPolicy, imagePullPolicy, "expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, sset.Spec.Template.Spec.Containers[0].ImagePullPolicy)
	}
	for _, container := range sset.Spec.Template.Spec.Containers {
		require.Equal(t, container.ImagePullPolicy, imagePullPolicy, "expected imagePullPolicy to match, want %s, got %s", imagePullPolicy, sset.Spec.Template.Spec.Containers[0].ImagePullPolicy)
	}
}

func TestPodHostNetworkConfig(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			HostNetwork: true,
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.True(t, sset.Spec.Template.Spec.HostNetwork, "expected hostNetwork to be true")
	require.Equal(t, v1.DNSClusterFirstWithHostNet, sset.Spec.Template.Spec.DNSPolicy,
		"expected DNSPolicy to be ClusterFirstWithHostNet when hostNetwork is enabled")
}

func TestConfigReloader(t *testing.T) {
	baseSet, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--web-config-file=/etc/alertmanager/web_config/web-config.yaml",
		"--reload-url=http://localhost:9093/-/reload",
		"--config-file=/etc/alertmanager/config/alertmanager.yaml.gz",
		"--config-envsubst-file=/etc/alertmanager/config_out/alertmanager.env.yaml",
		"--watched-dir=/etc/alertmanager/config",
	}

	for _, c := range baseSet.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expectd container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
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
			require.Equal(t, expectedArgsConfigReloader, c.Args, "expectd init container args are %s, but found %s", expectedArgsInitConfigReloader, c.Args)
		}
	}

}

func TestAutomountServiceAccountToken(t *testing.T) {
	for i := range []int{0, 1} {
		automountServiceAccountToken := (i == 0)
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			Spec: monitoringv1.AlertmanagerSpec{
				AutomountServiceAccountToken: &automountServiceAccountToken,
			},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)
		require.Equal(t, *sset.Spec.Template.Spec.AutomountServiceAccountToken, automountServiceAccountToken, "AutomountServiceAccountToken not found")
	}
}

func TestClusterLabel(t *testing.T) {
	tt := []struct {
		scenario                string
		version                 string
		expectedClusterLabelArg bool
		customClusterLabel      string
	}{{
		scenario:                "--cluster.label set by default for version >= v0.26.0",
		version:                 "0.26.0",
		expectedClusterLabelArg: true,
	}, {
		scenario:                "--cluster.label set if specified explicitly",
		version:                 "0.26.0",
		expectedClusterLabelArg: true,
		customClusterLabel:      "custom.cluster",
	}, {
		scenario:                "no --cluster.label set for older versions",
		version:                 "0.25.0",
		expectedClusterLabelArg: false,
	}}

	for _, ts := range tt {
		t.Run(ts.scenario, func(t *testing.T) {
			a := monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "alertmanager",
					Namespace: "monitoring",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					Replicas: toPtr(int32(1)),
					Version:  ts.version,
				},
			}

			if ts.customClusterLabel != "" {
				a.Spec.ClusterLabel = &ts.customClusterLabel
			}

			ss, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			args := ss.Template.Spec.Containers[0].Args
			if ts.expectedClusterLabelArg {
				if ts.customClusterLabel != "" {
					require.Contains(t, args, fmt.Sprintf("--cluster.label=%s", ts.customClusterLabel))
					return
				}
				require.Contains(t, args, "--cluster.label=monitoring/alertmanager")
				return
			}
			require.NotContains(t, args, "--cluster.label")
		})
	}
}

func TestMakeStatefulSetSpecTemplatesUniqueness(t *testing.T) {
	replicas := int32(1)
	tt := []struct {
		a                   monitoringv1.Alertmanager
		expectedSourceCount int
	}{
		{
			a: monitoringv1.Alertmanager{
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
							{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "template2",
									},
									Key: "template1.tmpl",
								},
							},
						},
					},
				},
			},
			expectedSourceCount: 1,
		},
		{
			a: monitoringv1.Alertmanager{
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
							{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "template2",
									},
									Key: "template2.tmpl",
								},
							},
						},
					},
				},
			},
			expectedSourceCount: 2,
		},
		{
			a: monitoringv1.Alertmanager{
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
			},
			expectedSourceCount: 1,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, test := range tt {
		statefulSpec, err := makeStatefulSetSpec(logger, &test.a, defaultTestConfig, &operator.ShardedSecret{})
		require.NoError(t, err)
		volumes := statefulSpec.Template.Spec.Volumes
		for _, volume := range volumes {
			if volume.Name == "notification-templates" {
				require.Len(t, volume.VolumeSource.Projected.Sources, test.expectedSourceCount, "expected %d sources, got %d", test.expectedSourceCount, len(volume.VolumeSource.Projected.Sources))
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

func TestEnableFeatures(t *testing.T) {
	tt := []struct {
		name             string
		version          string
		features         []string
		expectedFeatures []string
	}{
		{
			name:             "EnableFeaturesUnsupportedVersion",
			version:          "v0.26.0",
			features:         []string{"classic-mode"},
			expectedFeatures: []string{},
		},
		{
			name:             "EnableFeaturesWithOneFeature",
			version:          "v0.27.0",
			features:         []string{"classic-mode"},
			expectedFeatures: []string{"classic-mode"},
		},
		{
			name:             "EnableFeaturesWithMultipleFeatures",
			version:          "v0.27.0",
			features:         []string{"classic-mode", "receiver-name-in-metrics"},
			expectedFeatures: []string{"classic-mode", "receiver-name-in-metrics"},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			statefulSpec, err := makeStatefulSetSpec(nil, &monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version:        test.version,
					Replicas:       toPtr(int32(1)),
					EnableFeatures: test.features,
				},
			}, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			expectedFeatures := make([]string, 0)
			for _, flag := range statefulSpec.Template.Spec.Containers[0].Args {
				if after, ok := strings.CutPrefix(flag, "--enable-feature="); ok {
					expectedFeatures = append(expectedFeatures, strings.Split(after, ",")...)
				}
			}
			require.ElementsMatch(t, test.expectedFeatures, expectedFeatures)
		})
	}
}

func TestValidateAdditionalArgs(t *testing.T) {
	additionalArgs := []monitoringv1.Argument{
		{Name: "auto-gomemlimit.ratio", Value: "0.7"},
	}
	expectedArgs := []string{"--auto-gomemlimit.ratio=0.7"}

	statefulSpec, err := makeStatefulSetSpec(nil, &monitoringv1.Alertmanager{
		Spec: monitoringv1.AlertmanagerSpec{
			Replicas:       toPtr(int32(1)),
			AdditionalArgs: additionalArgs,
		},
	}, defaultTestConfig, &operator.ShardedSecret{})
	require.NoError(t, err)

	actualArgs := statefulSpec.Template.Spec.Containers[0].Args

	for _, expectedArg := range expectedArgs {
		require.Contains(t, actualArgs, expectedArg, "Expected additional argument not found")
	}
}

func TestStatefulSetDNSPolicyAndDNSConfig(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			DNSPolicy: ptr.To(monitoringv1.DNSClusterFirst),
			DNSConfig: &monitoringv1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8"},
				Searches:    []string{"custom.search"},
				Options: []monitoringv1.PodDNSConfigOption{
					{
						Name:  "ndots",
						Value: ptr.To("5"),
					},
				},
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	require.Equal(t, v1.DNSClusterFirst, sset.Spec.Template.Spec.DNSPolicy, "expected dns policy to match")
	require.Equal(t,
		&v1.PodDNSConfig{
			Nameservers: []string{"8.8.8.8"},
			Searches:    []string{"custom.search"},
			Options: []v1.PodDNSConfigOption{
				{
					Name:  "ndots",
					Value: ptr.To("5"),
				},
			},
		}, sset.Spec.Template.Spec.DNSConfig, "expected dns configuration to match")
}

func TestPersistentVolumeClaimRetentionPolicy(t *testing.T) {
	sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: monitoringv1.AlertmanagerSpec{
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}, defaultTestConfig, "", &operator.ShardedSecret{})
	require.NoError(t, err)

	if sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted != appsv1.DeletePersistentVolumeClaimRetentionPolicyType {
		t.Fatalf("expected persistentVolumeClaimDeletePolicy.WhenDeleted to be %s but got %s", appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted)
	}

	if sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled != appsv1.DeletePersistentVolumeClaimRetentionPolicyType {
		t.Fatalf("expected persistentVolumeClaimDeletePolicy.WhenScaled to be %s but got %s", appsv1.DeletePersistentVolumeClaimRetentionPolicyType, sset.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled)
	}
}

func TestStatefulSetEnableServiceLinks(t *testing.T) {
	tests := []struct {
		enableServiceLinks    *bool
		expectedEnableService *bool
	}{
		{enableServiceLinks: ptr.To(false), expectedEnableService: ptr.To(false)},
		{enableServiceLinks: ptr.To(true), expectedEnableService: ptr.To(true)},
		{enableServiceLinks: nil, expectedEnableService: nil},
	}

	for _, test := range tests {
		sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       monitoringv1.AlertmanagerSpec{EnableServiceLinks: test.expectedEnableService},
		}, defaultTestConfig, "", &operator.ShardedSecret{})
		require.NoError(t, err)

		if test.expectedEnableService != nil {
			require.NotNil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks not nil")
			require.Equal(t, *test.expectedEnableService, *sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks to match")
		} else {
			require.Nil(t, sset.Spec.Template.Spec.EnableServiceLinks, "expected enableServiceLinks is nil")
		}
	}
}

func TestStatefulSetPodManagementPolicy(t *testing.T) {
	for _, tc := range []struct {
		podManagementPolicy *monitoringv1.PodManagementPolicyType
		exp                 appsv1.PodManagementPolicyType
	}{
		{
			podManagementPolicy: nil,
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.ParallelPodManagement),
			exp:                 appsv1.ParallelPodManagement,
		},
		{
			podManagementPolicy: ptr.To(monitoringv1.OrderedReadyPodManagement),
			exp:                 appsv1.OrderedReadyPodManagement,
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					PodManagementPolicy: tc.podManagementPolicy,
				},
			}, defaultTestConfig, "", &operator.ShardedSecret{})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.PodManagementPolicy)
		})
	}
}

func TestStatefulSetUpdateStrategy(t *testing.T) {
	for _, tc := range []struct {
		updateStrategy *monitoringv1.StatefulSetUpdateStrategy
		exp            appsv1.StatefulSetUpdateStrategy
	}{
		{
			updateStrategy: nil,
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &monitoringv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
				},
			},
		},
		{
			updateStrategy: &monitoringv1.StatefulSetUpdateStrategy{
				Type: monitoringv1.OnDeleteStatefulSetStrategyType,
			},
			exp: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			sset, err := makeStatefulSet(nil, &monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					UpdateStrategy: tc.updateStrategy,
				},
			}, defaultTestConfig, "", &operator.ShardedSecret{})

			require.NoError(t, err)
			require.Equal(t, tc.exp, sset.Spec.UpdateStrategy)
		})
	}
}

func TestMakeStatefulSetSpecDispatchStartDelay(t *testing.T) {
	for _, tc := range []struct {
		version         string
		minReadySeconds *int32
		additionalArgs  []monitoringv1.Argument

		expContains    string
		expNotContains string
	}{
		{
			version:        "v0.30.0",
			expNotContains: "dispatch.start-delay",
		},
		{
			version:        "v0.30.0",
			additionalArgs: []monitoringv1.Argument{{Name: "dispatch.start-delay", Value: "1m"}},
			expContains:    "--dispatch.start-delay=1m",
		},
		{
			version:         "v0.29.0",
			minReadySeconds: ptr.To(int32(60)),
			expNotContains:  "dispatch.start-delay",
		},
		{
			version:         "v0.30.0",
			minReadySeconds: ptr.To(int32(60)),
			expContains:     "--dispatch.start-delay=60s",
		},
		{
			version:         "v0.30.0",
			minReadySeconds: ptr.To(int32(60)),
			additionalArgs:  []monitoringv1.Argument{{Name: "dispatch.start-delay", Value: "10s"}},
			expContains:     "--dispatch.start-delay=10s",
		},
	} {
		t.Run("", func(t *testing.T) {
			a := monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Replicas:        ptr.To(int32(1)),
					Version:         tc.version,
					MinReadySeconds: tc.minReadySeconds,
					AdditionalArgs:  tc.additionalArgs,
				},
			}

			statefulSet, err := makeStatefulSetSpec(nil, &a, defaultTestConfig, &operator.ShardedSecret{})
			require.NoError(t, err)

			if tc.expContains != "" {
				require.Contains(t, statefulSet.Template.Spec.Containers[0].Args, tc.expContains)
			}
			if tc.expNotContains != "" {
				require.NotContains(t, statefulSet.Template.Spec.Containers[0].Args, tc.expNotContains)
			}
		})
	}
}
