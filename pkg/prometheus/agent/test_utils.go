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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

var (
	defaultTestConfig = &prompkg.Config{
		LocalHost:                  "localhost",
		ReloaderConfig:             operator.DefaultReloaderTestConfig.ReloaderConfig,
		PrometheusDefaultBaseImage: operator.DefaultPrometheusBaseImage,
		ThanosDefaultBaseImage:     operator.DefaultThanosBaseImage,
	}
)

func makeSpecForTestListenTLS() monitoringv1alpha1.PrometheusAgentSpec {
	return monitoringv1alpha1.PrometheusAgentSpec{
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
	}
}

func makeExpectedProbeHandler(probePath string) v1.ProbeHandler {
	return v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path:   probePath,
			Port:   intstr.FromString("web"),
			Scheme: "HTTPS",
		},
	}
}

func makeExpectedStartupProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
}

func makeExpectedLivenessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
}
