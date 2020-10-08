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
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	caPEM = `-----BEGIN CERTIFICATE-----
MIIB4TCCAYugAwIBAgIUEP77V2fBpMyuLHrrltbhINcQmUAwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMDEwMDgxMjUyMDBaFw0yMTEw
MDgxMjUyMDBaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwXDANBgkqhkiG9w0BAQEF
AANLADBIAkEA6Ms44U9YF87b8H4Xeg0WG9HdjG+8YmVtn7ADI07JuUnRiC+s/r0D
6Y9YoUZS4SYCF9EL1KScuqI886F1pF29rwIDAQABo1MwUTAdBgNVHQ4EFgQU8Aj6
J+hDw0JYZCCpy354WcczbD8wHwYDVR0jBBgwFoAU8Aj6J+hDw0JYZCCpy354Wccz
bD8wDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAANBALC/hvicZRBdwzNY
P00PA2vNN40XYRLB+MhuN+IyaOU3sR9ckLJqDSLtlXlO0BM8985MGOgMuRHTtOAv
mB8zmAA=
-----END CERTIFICATE-----`

	certPEM = `-----BEGIN CERTIFICATE-----
MIIBhzCCATECFCKCReYDqoFeT3PU2/WOLWZVURtyMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMjAxMDA4MTMxNjQ0WhcNMjExMDA4MTMx
NjQ0WjBFMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAw
SAJBAKGlll7kmzMUvu9F6KiyUlDpJYTSaY1bcyyrvZ8p8es/XuxIUSuANkOR9HKM
pieiJ8uDE8fmgdnd5kQ2mhkvlRECAwEAATANBgkqhkiG9w0BAQsFAANBAEUzjBRc
lCOfcotJnliMu2e8RWYtlgliIynN+N8Pf/QujwtgGRlN8rdXoWOQJOl6+JdJWX5E
hmQbLIusBcbFKHI=
-----END CERTIFICATE-----`

	keyPEM = `-----BEGIN PRIVATE KEY-----
MIIBUwIBADANBgkqhkiG9w0BAQEFAASCAT0wggE5AgEAAkEAoaWWXuSbMxS+70Xo
qLJSUOklhNJpjVtzLKu9nynx6z9e7EhRK4A2Q5H0coymJ6Iny4MTx+aB2d3mRDaa
GS+VEQIDAQABAkB2JnAYgAOofHtqrLB3zY85MJCJ2rnn5nXyqrz4v1Hh3dBzc+vU
j8Zx/YXhDLgWphiuYd1XPglL782mp/tKQ10ZAiEAz+QvXnZjZgQtigtXoezdwIr2
Crim23maU0mgaB/M4EMCIQDHDc2ZGnEIi47/0IHDtN8bVEol90yPCFLqEPLwhDx6
GwIgHWPD8pXIDZcPnRFnbSPgYaUDjZZ3OFXjpFynSbEdNKMCIBHWgdMzlGeQohr4
o3hXUBsR3aczVzAGLe/93teA8i57AiAj/mocpjwcIVQOgYQ2E9PQCHJQdghS3brA
Qv34Vym87w==
-----END PRIVATE KEY-----`
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
			store := newAssetStore(c.CoreV1(), c.CoreV1())

			sel := v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: tc.selectedName,
				},
				Key: tc.selectedKey,
			}

			key := fmt.Sprintf("basicauth/%d", i)
			err := store.addBearerToken(context.Background(), tc.ns, sel, key)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			s, found := store.bearerTokenAssets[key]

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
			store := newAssetStore(c.CoreV1(), c.CoreV1())

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
			err := store.addBasicAuth(context.Background(), tc.ns, basicAuth, key)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			s, found := store.basicAuthAssets[key]

			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}

			if s.username != tc.expectedUser {
				t.Fatalf("expecting username %q, got %q", tc.expectedUser, s)
			}
			if s.password != tc.expectedPassword {
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
			store := newAssetStore(c.CoreV1(), c.CoreV1())

			err := store.addSafeTLSConfig(context.Background(), tc.ns, tc.tlsConfig.SafeTLSConfig)

			if tc.err {
				if err == nil {
					t.Fatal("expecting error, got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}

			key := tlsAssetKeyFromSelector(tc.ns, tc.tlsConfig.CA)

			ca, found := store.tlsAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(ca) != tc.expectedCA {
				t.Fatalf("expecting CA %q, got %q", tc.expectedCA, ca)
			}

			key = tlsAssetKeyFromSelector(tc.ns, tc.tlsConfig.Cert)

			cert, found := store.tlsAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(cert) != tc.expectedCert {
				t.Fatalf("expecting cert %q, got %q", tc.expectedCert, cert)
			}

			key = tlsAssetKeyFromSecretSelector(tc.ns, tc.tlsConfig.KeySecret)

			k, found := store.tlsAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(k) != tc.expectedKey {
				t.Fatalf("expecting cert key %q, got %q", tc.expectedCert, k)
			}
		})
	}
}
