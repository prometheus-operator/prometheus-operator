// Copyright 2023 The prometheus-operator Authors
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

package prometheusagent

import (
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

var (
	defaultTestConfig = &operator.Config{
		LocalHost:                  "localhost",
		ReloaderConfig:             operator.DefaultReloaderTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: operator.DefaultPrometheusBaseImage,
		ThanosDefaultBaseImage:     operator.DefaultThanosBaseImage,
	}
)

func TestListenTLS(t *testing.T) {
	sset, err := makeStatefulSetFromPrometheus(monitoringv1alpha1.PrometheusAgent{
		Spec: monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Web: &monitoringv1.PrometheusWebSpec{
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
		},
	})
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

	actualStartupProbe := sset.Spec.Template.Spec.Containers[0].StartupProbe
	expectedStartupProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
	if !reflect.DeepEqual(actualStartupProbe, expectedStartupProbe) {
		t.Fatalf("Startup probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedStartupProbe, actualStartupProbe)
	}

	actualLivenessProbe := sset.Spec.Template.Spec.Containers[0].LivenessProbe
	expectedLivenessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
	if !reflect.DeepEqual(actualLivenessProbe, expectedLivenessProbe) {
		t.Fatalf("Liveness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedLivenessProbe, actualLivenessProbe)
	}

	actualReadinessProbe := sset.Spec.Template.Spec.Containers[0].ReadinessProbe
	expectedReadinessProbe := &v1.Probe{
		ProbeHandler:     expectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
	if !reflect.DeepEqual(actualReadinessProbe, expectedReadinessProbe) {
		t.Fatalf("Readiness probe doesn't match expected. \n\nExpected: %+v\n\nGot: %+v", expectedReadinessProbe, actualReadinessProbe)
	}

	expectedConfigReloaderReloadURL := "--reload-url=https://localhost:9090/-/reload"
	reloadURLFound := false
	for _, arg := range sset.Spec.Template.Spec.Containers[1].Args {
		if arg == expectedConfigReloaderReloadURL {
			reloadURLFound = true
		}
	}
	if !reloadURLFound {
		t.Fatalf("expected to find arg %s in config reloader", expectedConfigReloaderReloadURL)
	}

	expectedArgsConfigReloader := []string{
		"--listen-address=:8080",
		"--reload-url=https://localhost:9090/-/reload",
		"--config-file=/etc/prometheus/config/prometheus.yaml.gz",
		"--config-envsubst-file=/etc/prometheus/config_out/prometheus.env.yaml",
	}

	for _, c := range sset.Spec.Template.Spec.Containers {
		if c.Name == "config-reloader" {
			if !reflect.DeepEqual(c.Args, expectedArgsConfigReloader) {
				t.Fatalf("expected container args are %s, but found %s", expectedArgsConfigReloader, c.Args)
			}
		}
	}
}

func newLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stderr), level.AllowWarn())
}

func makeStatefulSetFromPrometheus(p monitoringv1alpha1.PrometheusAgent) (*appsv1.StatefulSet, error) {
	logger := newLogger()

	cg, err := prompkg.NewConfigGenerator(logger, &p, false)
	if err != nil {
		return nil, err
	}

	return makeStatefulSet(
		logger,
		"test",
		&p,
		defaultTestConfig,
		cg,
		"",
		0,
		nil)
}
