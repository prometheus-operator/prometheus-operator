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
				"key1": "val1",
				"key2": "val2",
				"key3": "val3",
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key4": []byte("val4"),
				"key5": []byte("val5"),
				"key6": []byte("val6"),
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
				CA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key4",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key5",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			expectedCA:   "val4",
			expectedCert: "val5",
			expectedKey:  "val6",
		},
		{
			// CA in configmap, cert and key in secret.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key5",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			expectedCA:   "val1",
			expectedCert: "val5",
			expectedKey:  "val6",
		},
		{
			// CA and cert in configmap, key in secret.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key2",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			expectedCA:   "val1",
			expectedCert: "val2",
			expectedKey:  "val6",
		},
		{
			// Wrong namespace.
			ns: "ns2",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key2",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			err: true,
		},
		{
			// Wrong configmap selector for CA.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key4",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key2",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			err: true,
		},
		{
			// Wrong secret selector for CA.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key2",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			err: true,
		},
		{
			// Wrong configmap selector for cert.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key4",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			err: true,
		},
		{
			// Wrong secret selector for cert.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm",
						},
						Key: "key1",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key2",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key6",
				},
			},

			err: true,
		},
		{
			// Wrong key selector.
			ns: "ns1",
			tlsConfig: &monitoringv1.TLSConfig{
				CA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key4",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "key5",
					},
				},
				KeySecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key: "key7",
				},
			},

			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			store := newAssetStore(c.CoreV1(), c.CoreV1())

			err := store.addTLSConfig(context.Background(), tc.ns, tc.tlsConfig)

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
				t.Fatalf("expecting cert %q, got %q", tc.expectedCert, ca)
			}

			key = tlsAssetKeyFromSecretSelector(tc.ns, tc.tlsConfig.KeySecret)

			k, found := store.tlsAssets[key]
			if !found {
				t.Fatalf("expecting to find key %q but got nothing", key)
			}
			if string(k) != tc.expectedKey {
				t.Fatalf("expecting cert key %q, got %q", tc.expectedCert, ca)
			}
		})
	}
}
