// Copyright 2019 The prometheus-operator Authors
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

package framework

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	certutil "k8s.io/client-go/util/cert"
)

// MakeSecretWithCert returns a TLS Secret object from key and data slices.
func MakeSecretWithCert(ns, name string, keys []string, data [][]byte) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}

	for i := range keys {
		secret.Data[keys[i]] = data[i]
	}

	return secret
}

// GenerateServerCertificateSecret creates a self-signed certificate for the
// server and stores it in a Secret (if the secret already exists, it is updated). It returns the generated certificate.
func (f *Framework) GenerateServerCertificateSecret(ctx context.Context, n types.NamespacedName, server string) ([]byte, error) {
	cert, key, err := certutil.GenerateSelfSignedCertKey(server, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate/key for %q: %w", server, err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			PrivateKey: key,
			CertKey:    cert,
		},
	}

	_, err = f.KubeClient.CoreV1().Secrets(n.Namespace).Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get secret for %q: %w", server, err)
	}

	if apierrors.IsNotFound(err) {
		_, err = f.KubeClient.CoreV1().Secrets(n.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create secret for %q: %w", server, err)
		}

		return cert, nil
	}

	_, err = f.KubeClient.CoreV1().Secrets(n.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update secret for %q: %w", server, err)
	}

	return cert, nil
}
