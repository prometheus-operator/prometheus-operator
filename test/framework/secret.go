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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeSecretWithCert(ns, name string, keyList []string,
	dataList [][]byte) *corev1.Secret {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Type:       corev1.SecretType("Opaque"),
		Data:       map[string][]byte{},
	}

	for i := range keyList {
		secret.Data[keyList[i]] = dataList[i]
	}

	return secret
}

func (f *Framework) CreateSecretWithCert(certBytes, keyBytes []byte, ns, name string) error {

	secret := MakeSecretWithCert(ns, name, []string{"tls.key", "tls.crt"}, [][]byte{keyBytes, certBytes})
	_, err := f.KubeClient.CoreV1().Secrets(ns).Create(f.Ctx, secret, metav1.CreateOptions{})

	return err
}
