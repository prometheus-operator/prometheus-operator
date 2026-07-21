// Copyright The prometheus-operator Authors
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
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

func TestValidateTLSAssets(t *testing.T) {
	const namespace = "ns1"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"tls.crt": []byte("cert-data"),
			"tls.key": []byte("key-data"),
			"ca.crt":  []byte("ca-data"),
		},
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-configmap",
			Namespace: namespace,
		},
		Data: map[string]string{
			"tls.crt": "cert-data",
			"ca.crt":  "ca-data",
		},
	}

	validKeySecret := corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
		Key:                  "tls.key",
	}
	validCertSecret := monitoringv1.SecretOrConfigMap{
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
			Key:                  "tls.crt",
		},
	}
	validCertConfigMap := monitoringv1.SecretOrConfigMap{
		ConfigMap: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "tls-configmap"},
			Key:                  "tls.crt",
		},
	}
	validClientCA := monitoringv1.SecretOrConfigMap{
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
			Key:                  "ca.crt",
		},
	}

	for _, tc := range []struct {
		name    string
		tls     *monitoringv1.WebTLSConfig
		wantErr bool
	}{
		{
			name: "nil tls config",
			tls:  nil,
		},
		{
			name: "valid secret references",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: validKeySecret,
				Cert:      validCertSecret,
				ClientCA:  validClientCA,
			},
		},
		{
			name: "valid cert from configmap",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: validKeySecret,
				Cert:      validCertConfigMap,
			},
		},
		{
			name: "missing key in secret",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
					Key:                  "missing.key",
				},
				Cert: validCertSecret,
			},
			wantErr: true,
		},
		{
			name: "missing secret",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "missing-secret"},
					Key:                  "tls.key",
				},
				Cert: validCertSecret,
			},
			wantErr: true,
		},
		{
			name: "missing cert key",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: validKeySecret,
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
						Key:                  "missing.crt",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing client CA key",
			tls: &monitoringv1.WebTLSConfig{
				KeySecret: validKeySecret,
				Cert:      validCertSecret,
				ClientCA: monitoringv1.SecretOrConfigMap{
					Secret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "tls-secret"},
						Key:                  "missing-ca.crt",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing cert and key definition",
			tls:  &monitoringv1.WebTLSConfig{},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewClientset(secret, configMap)
			store := assets.NewStoreBuilder(client.CoreV1(), client.CoreV1())

			err := webconfig.ValidateTLSAssets(context.Background(), namespace, store, tc.tls)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
