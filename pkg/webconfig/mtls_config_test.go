package webconfig_test

import (
	"reflect"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var trueVal = true

func TestGenerateMTLSConfigFileContents(t *testing.T) {
	tc := []struct {
		name         string
		mtlsConfig   *monitoringv1.AMClusterTLSConfig
		expectedData string
	}{
		{
			name:         "mTLS config not defined",
			mtlsConfig:   nil,
			expectedData: "",
		},
		{
			name: "minimal mTLS config",
			mtlsConfig: &monitoringv1.AMClusterTLSConfig{
				TLSServerConfig: &monitoringv1.ServerTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-server.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-server.key",
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-server.client_ca",
						},
					},
				},
				TLSClientConfig: &monitoringv1.ClientTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-client.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-client.key",
					},
					ServerCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-client.server_ca",
						},
					},
				},
			},
			expectedData: `tls_server_config:
  cert_file: /mtlsconfig_path_prefix/secret_test-secret_tls-server.crt
  key_file: /mtlsconfig_path_prefix/secret_test-secret_tls-server.key
  client_ca_file: /mtlsconfig_path_prefix/configmap_test-configmap_tls-server.client_ca
tls_client_config:
  cert_file: /mtlsconfig_path_prefix/secret_test-secret_tls-client.crt
  key_file: /mtlsconfig_path_prefix/secret_test-secret_tls-client.key
  ca_file: /mtlsconfig_path_prefix/configmap_test-configmap_tls-client.server_ca
`,
		},
		{
			name: "complete mTLS config",
			mtlsConfig: &monitoringv1.AMClusterTLSConfig{
				TLSServerConfig: &monitoringv1.ServerTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-server.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-server.key",
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-server.client_ca",
						},
					},
					ClientAuthType:           "VerifyClientCertIfGiven",
					CipherSuites:             []string{"cipher-1", "cipher-2"},
					CurvePreferences:         []string{"curve-1", "curve-2"},
					MinVersion:               "TLS11",
					MaxVersion:               "TLS13",
					PreferServerCipherSuites: &trueVal,
				},
				TLSClientConfig: &monitoringv1.ClientTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-client.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-client.key",
					},
					ServerCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-client.server_ca",
						},
					},
					ServerName:         "test.server.name",
					InsecureSkipVerify: &trueVal,
				},
			},
			expectedData: `tls_server_config:
  cert_file: /mtlsconfig_path_prefix/secret_test-secret_tls-server.crt
  key_file: /mtlsconfig_path_prefix/secret_test-secret_tls-server.key
  client_auth_type: VerifyClientCertIfGiven
  client_ca_file: /mtlsconfig_path_prefix/configmap_test-configmap_tls-server.client_ca
  min_version: TLS11
  max_version: TLS13
  cipher_suites:
  - cipher-1
  - cipher-2
  prefer_server_cipher_suites: true
  curve_preferences:
  - curve-1
  - curve-2
tls_client_config:
  cert_file: /mtlsconfig_path_prefix/secret_test-secret_tls-client.crt
  key_file: /mtlsconfig_path_prefix/secret_test-secret_tls-client.key
  ca_file: /mtlsconfig_path_prefix/configmap_test-configmap_tls-client.server_ca
  server_name: test.server.name
  insecure_skip_verify: true
`,
		},
	}

	for _, tt := range tc {
		config, err := webconfig.NewMTLSConfig("/mtlsconfig_path_prefix", "test-mtlsconfig-secret", tt.mtlsConfig)
		if err != nil {
			t.Fatal(err)
		}

		secret, err := config.MakeMTLSConfigFileSecret(nil, metav1.OwnerReference{})
		if err != nil {
			t.Fatal(err)
		}

		if tt.expectedData != string(secret.Data["mtls-config.yaml"]) {
			t.Fatalf("%s failed.\n\nGot %s\nwant %s\n", tt.name, secret.Data["mtls-config.yaml"], tt.expectedData)
		}
	}

}

func TestMTLSGetMountParameters(t *testing.T) {
	ts := []struct {
		mtlsConfig      *monitoringv1.AMClusterTLSConfig
		expectedVolumes []v1.Volume
		expectedMounts  []v1.VolumeMount
	}{
		{
			mtlsConfig: nil,
			expectedVolumes: []v1.Volume{
				{
					Name: "mtls-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "mtls-config",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "mtls-config",
					ReadOnly:         true,
					MountPath:        "/mtlsconfig_path_prefix/mtls-config.yaml",
					SubPath:          "mtls-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
		{
			mtlsConfig: &monitoringv1.AMClusterTLSConfig{
				TLSServerConfig: &monitoringv1.ServerTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-server.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-server.key",
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-server.client_ca",
						},
					},
					ClientAuthType:           "VerifyClientCertIfGiven",
					CipherSuites:             []string{"cipher-1", "cipher-2"},
					CurvePreferences:         []string{"curve-1", "curve-2"},
					MinVersion:               "TLS11",
					MaxVersion:               "TLS13",
					PreferServerCipherSuites: &trueVal,
				},
				TLSClientConfig: &monitoringv1.ClientTLSConfig{
					TLSCert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "tls-client.crt",
						},
					},
					TLSKeySecret: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls-client.key",
					},
					ServerCA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
							Key: "tls-client.server_ca",
						},
					},
					ServerName:         "test.server.name",
					InsecureSkipVerify: &trueVal,
				},
			},
			expectedVolumes: []v1.Volume{
				{
					Name: "mtls-config",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "mtls-config",
						},
					},
				},
				{
					Name: "mtls-config-secret-key-test-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "test-secret",
						},
					},
				},
				{
					Name: "mtls-config-secret-cert-test-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "test-secret",
						},
					},
				},
				{
					Name: "mtls-config-configmap-client-ca-test-configmap",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "test-configmap",
							},
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "mtls-config",
					ReadOnly:         true,
					MountPath:        "/mtlsconfig_path_prefix/mtls-config.yaml",
					SubPath:          "mtls-config.yaml",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "mtls-config-secret-key-test-secret",
					ReadOnly:         true,
					MountPath:        "/mtlsconfig_path_prefix/secret_test-secret_tls-client.key",
					SubPath:          "tls-client.key",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "mtls-config-secret-cert-test-secret",
					ReadOnly:         true,
					MountPath:        "/mtlsconfig_path_prefix/secret_test-secret_tls-client.crt",
					SubPath:          "tls-client.crt",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "mtls-config-configmap-client-ca-test-configmap",
					ReadOnly:         true,
					MountPath:        "/mtlsconfig_path_prefix/configmap_test-configmap_tls-client.server_ca",
					SubPath:          "tls-client.server_ca",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	for _, tt := range ts {
		tlsAssets, err := webconfig.NewMTLSConfig("/mtlsconfig_path_prefix", "mtls-config", tt.mtlsConfig)
		if err != nil {
			t.Fatal(err)
		}

		_, volumes, mounts := tlsAssets.GetMountParameters()

		if !reflect.DeepEqual(volumes, tt.expectedVolumes) {
			t.Errorf("invalid volumes,\ngot  %v,\nwant %v", volumes, tt.expectedVolumes)
		}

		if !reflect.DeepEqual(mounts, tt.expectedMounts) {
			t.Errorf("invalid mounts,\ngot  %v,\nwant %v", mounts, tt.expectedMounts)
		}
	}
}
