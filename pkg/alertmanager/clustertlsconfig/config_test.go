// Copyright 2024 The prometheus-operator Authors
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

package clustertlsconfig_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/clustertlsconfig"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestCreateOrUpdateWebConfigSecret(t *testing.T) {
	tc := []struct {
		name             string
		clusterTLSConfig monitoringv1.ClusterTLSConfig
		golden           string
	}{
		{
			name:   "tls config not defined",
			golden: "tls_config_not_defined.golden",
		},
		{
			name:   "minimal TLS config with certificate from secret",
			golden: "minimal_TLS_config_with_certificate_from_secret.golden",
		},
		{
			name:   "minimal TLS config with certificate from configmap",
			golden: "minimal_TLS_config_with_certificate_from_configmap.golden",
		},
		{
			name:   "minimal TLS config with client CA from configmap",
			golden: "minimal_TLS_config_with_client_CA_from_configmap.golden",
		},
		{
			name:   "TLS config with all parameters from secrets",
			golden: "TLS_config_with_all_parameters_from_secrets.golden",
		},
		{
			name:   "TLS config with client CA, cert and key files",
			golden: "TLS_config_with_client_CA_cert_and_key_files.golden",
		},
		{
			name:   "HTTP config with all parameters",
			golden: "HTTP_config_with_all_parameters.golden",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			secretName := "test-secret"
			config, err := clustertlsconfig.New("/web_certs_path_prefix", secretName, &tt.clusterTLSConfig)
			require.NoError(t, err)

			var (
				s            = v1.Secret{}
				secretClient = fake.NewSimpleClientset().CoreV1().Secrets("default")
			)
			err = config.CreateOrUpdateClusterTLSConfigSecret(context.Background(), secretClient, &s)
			require.NoError(t, err)

			secret, err := secretClient.Get(context.Background(), secretName, metav1.GetOptions{})
			require.NoError(t, err)

			golden.Assert(t, string(secret.Data["web-config.yaml"]), tt.golden)
		})
	}
}

func TestGetMountParameters(t *testing.T) {
	ts := []struct {
		clusterTLSConfig monitoringv1.ClusterTLSConfig
		expectedVolumes  []v1.Volume
		expectedMounts   []v1.VolumeMount
	}{
		{
			expectedVolumes: []v1.Volume{
				{
					Name: "web-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "web-config",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "web-config",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/web-config.yaml",
					SubPath:          "web-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
		{
			clusterTLSConfig: monitoringv1.ClusterTLSConfig{
				ServerTLS: &monitoringv1.WebTLSConfig{
					KeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "some-secret",
						},
						Key: "tls.key",
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "some-secret",
							},
							Key: "tls.crt",
						},
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "some-secret",
							},
							Key: "tls.client_ca",
						},
					},
				},
			},
			expectedVolumes: []v1.Volume{
				{
					Name: "web-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "web-config",
						},
					},
				},
				{
					Name: "web-config-tls-secret-key-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-cert-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-client-ca-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "web-config",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/web-config.yaml",
					SubPath:          "web-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-key-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret/some-secret-key",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-cert-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret/some-secret-cert",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-client-ca-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret/some-secret-ca",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	for _, tt := range ts {
		t.Run("", func(t *testing.T) {
			tlsAssets, err := clustertlsconfig.New("/etc/prometheus/web_config", "web-config", &tt.clusterTLSConfig)
			require.NoError(t, err)

			_, volumes, mounts, err := tlsAssets.GetMountParameters()
			require.NoError(t, err)

			require.Equal(t, tt.expectedVolumes, volumes)
			require.Equal(t, tt.expectedMounts, mounts)
		})
	}
}
