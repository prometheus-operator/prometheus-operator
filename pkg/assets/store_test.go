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

package assets

import (
	"context"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	caPEM = `-----BEGIN CERTIFICATE-----
MIIB4zCCAY2gAwIBAgIUf+9T+SQuY7RzRfLrT/m3ZLZa/nswDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yMDEwMTkxMzA1MDlaGA8yMTIw
MDkyNTEzMDUwOVowRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0GCSqGSIb3DQEB
AQUAA0sAMEgCQQDbXwmz6fkHnfs3p5dirgW/m5G1eOSddS8atIwhOzaYSNG03/Z4
P6HWCGDCgUg77fOsX+tzYWkXy0T+GwQrTLDdAgMBAAGjUzBRMB0GA1UdDgQWBBTC
CNvaPTFE1Xt5WUREDoF/mTOg7DAfBgNVHSMEGDAWgBTCCNvaPTFE1Xt5WUREDoF/
mTOg7DAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA0EAzhzA2n5nSnka
k9iw9ZHayRBSgnGAYKFdiGyvceKPzR3LJ8vMdGeYh/TSHHgZ4QSam/J7vHWCkJmc
7c98vpkIaw==
-----END CERTIFICATE-----`

	certPEM = `-----BEGIN CERTIFICATE-----
MIIBiTCCATMCFCgn66sq14Tsx6iP8nRdP4/uiguXMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwIBcNMjAxMDE5MTMwNTI5WhgPMjEyMDA5MjUx
MzA1MjlaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwXDANBgkqhkiG9w0BAQEFAANL
ADBIAkEA1wNGN6zrF4eBpW/LcNt3Qxy9bZZss6c/pUy5V4n2O+tZZuvKXF3Q6g4+
fOZ5xgqzqPgg2UzrG1Mmt/Ol4UikZQIDAQABMA0GCSqGSIb3DQEBCwUAA0EAGsWD
5UlmIIbFOi50jqNE3KitIwbPuY8nYR8pS2HYSE+eVKpGFmmzIRXkb4ZmdVymI+vG
B9nfCt+guZqCLxZMDQ==
-----END CERTIFICATE-----`

	keyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBANcDRjes6xeHgaVvy3Dbd0McvW2WbLOnP6VMuVeJ9jvrWWbrylxd
0OoOPnzmecYKs6j4INlM6xtTJrfzpeFIpGUCAwEAAQJAMhPxJsZ/ett0trNzDrYO
8PKgrAV9C9rIWBemk1zunMWmmtBt295sEK555iedWanANhTYKlaezUXMBZaoHIhc
AQIhAPB6QM5fGEsH1VSXEgaSb/EewQLFGjkWj9DtFtwOtmWpAiEA5OQ7NTVq9ULq
6qAI/JJ6qVGCjS/bmUQD2aBrUUhdxl0CIQDrOvsno/fUdS4ll70nNplPqICu3/Ud
wMcfXLwOuEmNOQIhAMSYi4o+IWobWe7AGjfmEFkR25ItAu73jl8D/GlKQNE5AiEA
hvBlhCknnq89u57O41ID6Mqxz3bRxNxpkqhfMyVWcVU=
-----END RSA PRIVATE KEY-----`
)

func TestAddBearerToken(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1": []byte("val1"),
			},
		},
	)

	for i, tc := range []struct {
		ns           string
		selectedName string
		selectedKey  string

		err      bool
		expected string
	}{
		{
			ns:           "ns1",
			selectedName: "secret",
			selectedKey:  "key1",

			expected: "val1",
		},
		// Wrong namespace.
		{
			ns:           "ns2",
			selectedName: "secret",
			selectedKey:  "key1",

			err: true,
		},
		// Wrong name.
		{
			ns:           "ns1",
			selectedName: "secreet",
			selectedKey:  "key1",

			err: true,
		},
		// Wrong key.
		{
			ns:           "ns1",
			selectedName: "secret",
			selectedKey:  "key2",

			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			store := NewStore(c.CoreV1(), c.CoreV1())

			sel := v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: tc.selectedName,
				},
				Key: tc.selectedKey,
			}

			key := fmt.Sprintf("basicauth/%d", i)
			err := store.AddBearerToken(context.Background(), tc.ns, sel, key)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			s, found := store.BearerTokenAssets[key]

			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}

			if string(s) != tc.expected {
				t.Fatalf("expecting %q, got %q", tc.expected, s)
			}
		})
	}
}

func TestAddBasicAuth(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1": []byte("val1"),
				"key2": []byte("val2"),
			},
		},
	)

	for i, tc := range []struct {
		ns                   string
		selectedUserName     string
		selectedUserKey      string
		selectedPasswordName string
		selectedPasswordKey  string

		err              bool
		expectedUser     string
		expectedPassword string
	}{
		{
			ns:                   "ns1",
			selectedUserName:     "secret",
			selectedUserKey:      "key1",
			selectedPasswordName: "secret",
			selectedPasswordKey:  "key2",

			expectedUser:     "val1",
			expectedPassword: "val2",
		},
		// Wrong namespace.
		{
			ns:                   "ns2",
			selectedUserName:     "secret",
			selectedUserKey:      "key1",
			selectedPasswordName: "secret",
			selectedPasswordKey:  "key2",

			err: true,
		},
		// Wrong name for username selector.
		{
			ns:                   "ns1",
			selectedUserName:     "secreet",
			selectedUserKey:      "key1",
			selectedPasswordName: "secret",
			selectedPasswordKey:  "key2",

			err: true,
		},
		// Wrong key for username selector.
		{
			ns:                   "ns1",
			selectedUserName:     "secret",
			selectedUserKey:      "key3",
			selectedPasswordName: "secret",
			selectedPasswordKey:  "key2",

			err: true,
		},
		// Wrong name for password selector.
		{
			ns:                   "ns1",
			selectedUserName:     "secret",
			selectedUserKey:      "key1",
			selectedPasswordName: "secreet",
			selectedPasswordKey:  "key2",

			err: true,
		},
		// Wrong key for password selector.
		{
			ns:                   "ns1",
			selectedUserName:     "secret",
			selectedUserKey:      "key1",
			selectedPasswordName: "secret",
			selectedPasswordKey:  "key3",

			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			store := NewStore(c.CoreV1(), c.CoreV1())

			basicAuth := &monitoringv1.BasicAuth{
				Username: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: tc.selectedUserName,
					},
					Key: tc.selectedUserKey,
				},
				Password: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: tc.selectedPasswordName,
					},
					Key: tc.selectedPasswordKey,
				},
			}

			key := fmt.Sprintf("basicauth/%d", i)
			err := store.AddBasicAuth(context.Background(), tc.ns, basicAuth, key)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			s, found := store.BasicAuthAssets[key]

			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}

			if s.Username != tc.expectedUser {
				t.Fatalf("expecting username %q, got %q", tc.expectedUser, s)
			}
			if s.Password != tc.expectedPassword {
				t.Fatalf("expecting password %q, got %q", tc.expectedPassword, s)
			}
		})
	}
}

func TestAddTLSConfig(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm",
				Namespace: "ns1",
			},
			Data: map[string]string{
				"cmCA":   caPEM,
				"cmCert": certPEM,
				"cmKey":  keyPEM,
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"secretCA":   []byte(caPEM),
				"secretCert": []byte(certPEM),
				"secretKey":  []byte(keyPEM),

				"invalidCA": []byte("invalidCA"),
				"wrongKey":  []byte("wrongKey"),
			},
		},
	)

	for _, tc := range []struct {
		ns        string
		tlsConfig *monitoringv1.TLSConfig

		err          bool
		expectedCA   string
		expectedCert string
		expectedKey  string
	}{
		{
			// CA, cert and key in secret.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			expectedCA:   caPEM,
			expectedCert: certPEM,
			expectedKey:  keyPEM,
		},
		{
			// CA in configmap, cert and key in secret.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			expectedCA:   caPEM,
			expectedCert: certPEM,
			expectedKey:  keyPEM,
		},
		{
			// CA and cert in configmap, key in secret.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			expectedCA:   caPEM,
			expectedCert: certPEM,
			expectedKey:  keyPEM,
		},
		{
			// Wrong namespace.
			ns: "ns2",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Wrong configmap selector for CA.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "secretCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Wrong secret selector for CA.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Wrong configmap selector for cert.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Wrong secret selector for cert.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm",
							},
							Key: "cmCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "cmCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Wrong key selector.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "cmKey",
					},
				},
			},

			err: true,
		},
		{
			// Cert without key.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
				},
			},

			err: true,
		},
		{
			// Key without cert.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCA",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
		{
			// Cert with wrong key.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "wrongKey",
					},
				},
			},

			err: true,
		},
		{
			// Invalid CA certificate.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "invalidCA",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "secretCert",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "secretKey",
					},
				},
			},

			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			store := NewStore(c.CoreV1(), c.CoreV1())

			err := store.AddSafeTLSConfig(context.Background(), tc.ns, &tc.tlsConfig.SafeTLSConfig)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			key := TLSAssetKeyFromSelector(tc.ns, tc.tlsConfig.CA)

			ca, found := store.TLSAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(ca) != tc.expectedCA {
				t.Fatalf("expecting CA %q, got %q", tc.expectedCA, ca)
			}

			key = TLSAssetKeyFromSelector(tc.ns, tc.tlsConfig.Cert)

			cert, found := store.TLSAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(cert) != tc.expectedCert {
				t.Fatalf("expecting cert %q, got %q", tc.expectedCert, cert)
			}

			key = TLSAssetKeyFromSecretSelector(tc.ns, tc.tlsConfig.KeySecret)

			k, found := store.TLSAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(k) != tc.expectedKey {
				t.Fatalf("expecting cert key %q, got %q", tc.expectedCert, k)
			}
		})
	}
}
