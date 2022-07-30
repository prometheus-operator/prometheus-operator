// Copyright 2021 The prometheus-operator Authors
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

package webconfig_test

import (
	"context"
	"reflect"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var falseVal = false

func TestCreateOrUpdateWebConfigSecret(t *testing.T) {
	tc := []struct {
		name                 string
		webConfigFileFields  monitoringv1.WebConfigFileFields
		basicAuthCredentials map[string]assets.BasicAuthCredentials
		expectedData         string
	}{
		{
			name:                "web config file fields not defined",
			webConfigFileFields: monitoringv1.WebConfigFileFields{},
			expectedData:        "",
		},
		{
			name: "minimal TLS config with certificate from secret",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
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
			},
			expectedData: `tls_server_config:
  cert_file: /web_certs_path_prefix/secret_test-secret_tls.crt
  key_file: /web_certs_path_prefix/secret_test-secret_tls.key
`,
		},
		{
			name: "minimal TLS config with certificate from configmap",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
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
			},
			expectedData: `tls_server_config:
  cert_file: /web_certs_path_prefix/configmap_test-configmap_tls.crt
  key_file: /web_certs_path_prefix/secret_test-secret_tls.key
`,
		},
		{
			name: "minimal TLS config with client CA from configmap",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
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
			},
			expectedData: `tls_server_config:
  cert_file: /web_certs_path_prefix/configmap_test-configmap_tls.crt
  key_file: /web_certs_path_prefix/secret_test-secret_tls.key
  client_ca_file: /web_certs_path_prefix/configmap_test-configmap_tls.client_ca
`,
		},
		{
			name: "TLS config with all parameters from secrets",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
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
					PreferServerCipherSuites: &falseVal,
					CurvePreferences:         []string{"curve-1", "curve-2"},
				},
			},
			expectedData: `tls_server_config:
  cert_file: /web_certs_path_prefix/secret_test-secret_tls.crt
  key_file: /web_certs_path_prefix/secret_test-secret_tls.keySecret
  client_auth_type: RequireAnyClientCert
  client_ca_file: /web_certs_path_prefix/secret_test-secret_tls.ca
  min_version: TLS11
  max_version: TLS13
  cipher_suites:
  - cipher-1
  - cipher-2
  prefer_server_cipher_suites: false
  curve_preferences:
  - curve-1
  - curve-2
`,
		},
		{
			name: "HTTP config with all parameters",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				HTTPConfig: &monitoringv1.WebHTTPConfig{
					HTTP2: &falseVal,
					Headers: &monitoringv1.WebHTTPHeaders{
						ContentSecurityPolicy:   "test",
						StrictTransportSecurity: "test",
						XContentTypeOptions:     "NoSniff",
						XFrameOptions:           "SameOrigin",
						XXSSProtection:          "test",
					},
				},
			},
			expectedData: `http_server_config:
  http2: false
  headers:
    Content-Security-Policy: test
    Strict-Transport-Security: test
    X-Content-Type-Options: nosniff
    X-Frame-Options: sameorigin
    X-XSS-Protection: test
`,
		},
		{
			name: "basic auth users config",
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				BasicAuthUsers: []monitoringv1.BasicAuth{
					{
						Username: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "foo",
							},
							Key: "username",
						},
						Password: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "bar",
							},
							Key: "password",
						},
					},
				},
			},
			basicAuthCredentials: map[string]assets.BasicAuthCredentials{
				"webconfig/managed/prometheus-operator-managed-user": {
					Username: "prometheus-operator-managed-user",
					Password: "abc",
				},
				"webconfig/userDefine/foo": {
					Username: "foo",
					Password: "bar",
				},
			},
			expectedData: `basic_auth_users:
  prometheus-operator-managed-user: abc
  foo: bar
`,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			secretName := "test-secret"
			ctx := context.TODO()
			secretClient := fake.NewSimpleClientset().CoreV1().Secrets("default")

			config, err := webconfig.New("/web_certs_path_prefix", secretName, "default", tt.webConfigFileFields)
			if err != nil {
				t.Fatal(err)
			}

			if err := config.CreateOrUpdateWebConfigSecret(ctx, secretClient, nil, metav1.OwnerReference{}, &assets.Store{BasicAuthAssets: tt.basicAuthCredentials}); err != nil {
				t.Fatal(err)
			}

			secret, err := secretClient.Get(ctx, secretName, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if tt.expectedData != string(secret.Data["web-config.yaml"]) {
				t.Fatalf("Got %s\nwant %s\n", secret.Data["web-config.yaml"], tt.expectedData)
			}
		})
	}
}

func TestGetMountParameters(t *testing.T) {
	ts := []struct {
		webConfigFileFields monitoringv1.WebConfigFileFields
		expectedVolumes     []v1.Volume
		expectedMounts      []v1.VolumeMount
	}{
		{
			webConfigFileFields: monitoringv1.WebConfigFileFields{},
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
			webConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
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
					Name: "web-config-tls-secret-key-some-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-cert-some-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-client-ca-some-secret",
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
					Name:             "web-config-tls-secret-key-some-secret",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret_some-secret_tls.key",
					SubPath:          "tls.key",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-cert-some-secret",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret_some-secret_tls.crt",
					SubPath:          "tls.crt",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-client-ca-some-secret",
					ReadOnly:         true,
					MountPath:        "/etc/prometheus/web_config/secret_some-secret_tls.client_ca",
					SubPath:          "tls.client_ca",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	for _, tt := range ts {
		tlsAssets, err := webconfig.New("/etc/prometheus/web_config", "web-config", "default", tt.webConfigFileFields)
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
