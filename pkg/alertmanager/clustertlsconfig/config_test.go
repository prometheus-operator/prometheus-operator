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
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/clustertlsconfig"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestCreateOrUpdateClusterTLSConfigSecret(t *testing.T) {

	tc := []struct {
		name             string
		clusterTLSConfig monitoringv1.ClusterTLSConfigFields
		golden           string
	}{
		{
			name:             "cluster tls config not defined",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{},
			golden:           "clusterTLS_config_not_defined.golden",
		},
		{
			name: "minimal cluster TLS config with server certificate from secret",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
				ServerTLS: &monitoringv1.WebTLSConfig{
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
				},
				ClientTLS: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.KeySecret",
					},
				},
			},
			golden: "minimal_clusterTLS_config_with_certificate_from_secret.golden",
		},
		{
			name: "minimal cluster TLS config with server and client certificates from configmap",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
				ServerTLS: &monitoringv1.WebTLSConfig{
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
				},
				ClientTLS: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
				},
			},
			golden: "minimal_clusterTLS_config_with_certificate_from_configmap.golden",
		},
		{
			name: "minimal cluster TLS config with server TLS cert and clientCA from configmap",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
				ServerTLS: &monitoringv1.WebTLSConfig{
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls.client_ca",
						},
					},
				},
				ClientTLS: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
				},
			},
			golden: "minimal_clusterTLS_config_with_client_CA_from_configmap.golden",
		},
		{
			name: "cluster tls config with all parameters from secrets",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
				ServerTLS: &monitoringv1.WebTLSConfig{
					ClientCA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.keySecret",
					},
					ClientAuthType:           "RequireAnyClientCert",
					MinVersion:               "TLS11",
					MaxVersion:               "TLS13",
					CipherSuites:             []string{"cipher-1", "cipher-2"},
					PreferServerCipherSuites: ptr.To(false),
					CurvePreferences:         []string{"curve-1", "curve-2"},
				},
				ClientTLS: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.ca",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.KeySecret",
					},
				},
			},
			golden: "clusterTLS_config_with_all_parameters_from_secrets.golden",
		},
		{
			name: "cluster tls config with server client CA, cert and key files",
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
				ServerTLS: &monitoringv1.WebTLSConfig{
					ClientCAFile: "/etc/ssl/certs/tls.client_ca",
					CertFile:     "/etc/ssl/certs/tls.crt",
					KeyFile:      "/etc/ssl/secrets/tls.key",
				},
				ClientTLS: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.key",
					},
				},
			},
			golden: "clusterTLS_config_with_client_CA_cert_and_key_files.golden",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			secretName := "test-secret"
			config, err := clustertlsconfig.New("/cluster_tls_certs_path_prefix", secretName, tt.clusterTLSConfig)
			require.NoError(t, err)

			var (
				s            = v1.Secret{}
				secretClient = fake.NewSimpleClientset().CoreV1().Secrets("default")
			)
			err = config.CreateOrUpdateClusterTLSConfigSecret(context.Background(), secretClient, &s)
			require.NoError(t, err)

			secret, err := secretClient.Get(context.Background(), secretName, metav1.GetOptions{})
			require.NoError(t, err)

			golden.Assert(t, string(secret.Data["cluster-tls-config.yaml"]), tt.golden)
		})
	}

}

func TestGetMountParameters(t *testing.T) {
	ts := []struct {
		clusterTLSConfig monitoringv1.ClusterTLSConfigFields
		expectedVolumes  []v1.Volume
		expectedMounts   []v1.VolumeMount
	}{
		{
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{},
			expectedVolumes: []v1.Volume{
				{
					Name: "cluster-tls-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "cluster-tls-config",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "cluster-tls-config",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/cluster_tls_config/cluster-tls-config.yaml",
					SubPath:          "cluster-tls-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
		{
			clusterTLSConfig: monitoringv1.ClusterTLSConfigFields{
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
					Name: "cluster-tls-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "cluster-tls-config",
						},
					},
				},
				{
					Name: "cluster-tls-server-config-secret-key-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "cluster-tls-server-config-secret-cert-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "cluster-tls-server-config-secret-client-ca-some-secret-3556f148",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "cluster-tls-config",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/cluster_tls_config/cluster-tls-config.yaml",
					SubPath:          "cluster-tls-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "cluster-tls-server-config-secret-key-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/cluster_tls_config/server_tls/secret/some-secret-key",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "cluster-tls-server-config-secret-cert-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/cluster_tls_config/server_tls/secret/some-secret-cert",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "cluster-tls-server-config-secret-client-ca-some-secret-3556f148",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/cluster_tls_config/server_tls/secret/some-secret-ca",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	for _, tt := range ts {
		t.Run("", func(t *testing.T) {
			tlsAssets, err := clustertlsconfig.New("/etc/prometheus/cluster_tls_config", "cluster-tls-config", tt.clusterTLSConfig)
			require.NoError(t, err)

			_, volumes, mounts, err := tlsAssets.GetMountParameters()
			require.NoError(t, err)

			require.Equal(t, tt.expectedVolumes, volumes)
			require.Equal(t, tt.expectedMounts, mounts)
		})
	}
}
