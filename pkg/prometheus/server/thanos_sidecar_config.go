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

package prometheus

import (
	"fmt"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

const (
	thanosConfigDir                                  = "/etc/thanos/config"
	thanosPrometheusHTTPClientConfigFileName         = "prometheus.http-client-file.yaml"
	thanosPrometheusHTTPClientConfigSecretNameSuffix = "thanos-prometheus-http-client-file"
)

// buildPrometheusHTTPClientConfigSecret returns a kubernetes secret with the HTTP configuration for the Thanos sidecar
// to communicated with prometheus server.
// https://thanos.io/tip/components/sidecar.md/#prometheus-http-client
func buildPrometheusHTTPClientConfigSecret(p *monitoringv1.Prometheus) (*v1.Secret, error) {
	dataYaml := yaml.MapSlice{}
	dataYaml = append(dataYaml, yaml.MapItem{
		Key: "tls_config",
		Value: yaml.MapSlice{
			{
				Key:   "insecure_skip_verify",
				Value: true,
			},
		},
	})

	data, err := yaml.Marshal(dataYaml)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      thanosPrometheusHTTPClientConfigSecretName(p),
			Namespace: p.Namespace,
		},
		Data: map[string][]byte{
			thanosPrometheusHTTPClientConfigFileName: data,
		},
	}, nil
}

func thanosPrometheusHTTPClientConfigSecretName(p monitoringv1.PrometheusInterface) string {
	return fmt.Sprintf("%s-%s", prompkg.PrefixedName(p), thanosPrometheusHTTPClientConfigSecretNameSuffix)
}
