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
	"context"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

const (
	ThanosConfigDir                                  = "/etc/thanos/config"
	ThanosPrometheusHTTPClientConfigFileName         = "prometheus.http-client-file.yaml"
	ThanosPrometheusHTTPClientConfigSecretNameSuffix = "thanos-prometheus-http-client-file"
)

type ThanosPrometheusHTTPClientConfig struct {
	TLSConfig  *monitoringv1.TLSConfig
	secretName string
}

func NewPrometheusHTTPClientConfig(secretName string) *ThanosPrometheusHTTPClientConfig {
	return &ThanosPrometheusHTTPClientConfig{
		TLSConfig: &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				// sidecar is listen to prometheus server on localhost in pod, no need to use tls.
				InsecureSkipVerify: ptr.To(true),
			},
		},
		secretName: secretName,
	}
}

// CreateOrUpdatePrometheusHTTPClientConfigSecret create or update a kubernetes secret with the data for thanos sidecar
// communicated with prometheus server.
// https://thanos.io/tip/components/sidecar.md/#prometheus-http-client
func (c *ThanosPrometheusHTTPClientConfig) CreateOrUpdatePrometheusHTTPClientConfigSecret(ctx context.Context, secretClient clientv1.SecretInterface, s *v1.Secret) error {
	cfg := yaml.MapSlice{}
	cfg = append(cfg, yaml.MapItem{
		Key: "tls_config",
		Value: yaml.MapSlice{
			{
				Key:   "insecure_skip_verify",
				Value: c.TLSConfig.InsecureSkipVerify,
			},
		},
	})

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	s.Name = c.secretName
	s.Data = map[string][]byte{
		ThanosPrometheusHTTPClientConfigFileName: data,
	}

	return k8sutil.CreateOrUpdateSecret(ctx, secretClient, s)
}

func ThanosPrometheusHTTPClientConfigSecretName(p monitoringv1.PrometheusInterface) string {
	return fmt.Sprintf("%s-%s", prompkg.PrefixedName(p), ThanosPrometheusHTTPClientConfigSecretNameSuffix)
}

func ThanosConfigFilePath(thanosConfigPath string) string {
	return filepath.Join(ThanosConfigDir, thanosConfigPath)
}
